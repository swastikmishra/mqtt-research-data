// main.go
//
// What it does:
// - Stress test single broker OR clustered brokers (brokers.json).
// - Fixed publishers for the whole test run (per broker-set epoch).
// - Subscriber ramp: start with N subscribers, then N+100, N+200... until SLA fails.
// - If brokers.json is present AND --cluster-incremental=true:
//     - Start with 1 broker active
//     - Run the subscriber ramp until SLA fails
//     - Stop the epoch, add the next broker, restart epoch from scratch
//     - Repeat until the last broker is used
//
// QoS: fixed QoS0
// Payload: choose 10, 100, 1000 bytes via --payload-bytes
//
// Output files (in --out-dir):
// - <test-name>.windows.csv : window/step metrics (best for plotting & paper)
// - <test-name>.summary.json: summary (max sustainable subscribers per broker count, failure reason, etc.)
//
// Notes:
// - Latency SLA is computed on a WINDOW histogram (not "since start").
// - Delivery ratio is WINDOW delivered / WINDOW expected (expected = pubs_sent_delta * fully_subscribed_connected_subs).
// - To avoid the tester becoming the bottleneck, latency sampling is supported via --lat-sample-every (default 50).
//
// Build:
//   go build -o main main.go
//
// Examples:
// Single broker baseline:
//   ./main --broker-host=127.0.0.1 --broker-port=1883 --publishers=50 --pub-rate=1 --payload-bytes=100
//
// Cluster incremental (brokers.json):
//   ./main --brokers-json=./brokers.json --cluster-incremental=true --publishers=50 --pub-rate=1 --payload-bytes=100
//
package main

import (
	"context"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"hash/fnv"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

/* =======================
   Config
   ======================= */

type Config struct {
	// Single broker (if brokers-json is empty)
	BrokerHost string
	BrokerPort int

	// Cluster brokers file (optional)
	BrokersJSON string

	// Existing behavior (restart epochs)
	ClusterIncremental bool

	// NEW (S3): hot-add; only new clients pinned to next broker (no rebalance, no restart)
	ClusterHotAddNewClients bool

	// Test naming / output
	TestName string
	OutDir   string

	// Load: subscribers
	InitialSubs int
	SubStep     int // +100
	MaxSubs     int // safety cap
	WindowSec   int // hold each step for this many seconds
	WarmupSec   int // warmup per step (ignored windows)
	MaxEpochSec int // safety cap per broker-count epoch (used by restart-epoch incremental / fixed)
	MaxTotalSec int // safety cap overall

	// Publishers fixed per epoch / run
	Publishers     int
	PubRatePerSec  float64 // per publisher
	TopicPrefix    string
	TopicCount     int
	PayloadBytes   int // 10/100/1000
	LatencySampleN int // sample every N messages per subscriber (approx). If 1, sample all.

	// SLA kill switches (window-based)
	MinConnectedSubPct  float64 // e.g. 99
	MinDeliveryRatioPct float64 // e.g. 99
	MaxP95LatencyMs     float64 // e.g. 500
	MaxDiscPerMin       float64 // e.g. 50

	// Consecutive failing windows to declare SLA failure (avoid flukes)
	ConsecutiveBreaches int
}

/* =======================
   Broker pool
   ======================= */

type Broker struct {
	ID   string `json:"id"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

type BrokersFile struct {
	Mode    string   `json:"mode,omitempty"`
	Brokers []Broker `json:"brokers"`
}

func loadBrokersFile(path string) (*BrokersFile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var bf BrokersFile
	if err := json.Unmarshal(b, &bf); err != nil {
		return nil, err
	}
	for i := range bf.Brokers {
		if strings.TrimSpace(bf.Brokers[i].Host) == "" {
			return nil, fmt.Errorf("broker[%d] host is empty", i)
		}
		if bf.Brokers[i].Port == 0 {
			bf.Brokers[i].Port = 1883
		}
		if strings.TrimSpace(bf.Brokers[i].ID) == "" {
			bf.Brokers[i].ID = fmt.Sprintf("broker-%d", i)
		}
	}
	return &bf, nil
}

func pickBrokerDeterministic(brokers []Broker, key string) Broker {
	if len(brokers) == 1 {
		return brokers[0]
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	idx := int(h.Sum32()) % len(brokers)
	return brokers[idx]
}

/* =======================
   Metrics and latency histogram (windowed)
   ======================= */

type LatencyHist struct {
	edges  []float64
	counts []uint64
	total  uint64
}

func NewLatencyHist() LatencyHist {
	edges := []float64{
		0.5, 1, 2, 3, 5, 7.5, 10, 15, 20, 30, 50, 75, 100,
		150, 200, 300, 500, 750, 1000, 1500, 2000, 3000, 5000,
	}
	return LatencyHist{
		edges:  edges,
		counts: make([]uint64, len(edges)+1),
	}
}

func (h *LatencyHist) Add(ms float64) {
	i := 0
	for i < len(h.edges) && ms > h.edges[i] {
		i++
	}
	h.counts[i]++
	h.total++
}

func (h *LatencyHist) Quantile(q float64) float64 {
	if h.total == 0 {
		return 0
	}
	target := uint64(math.Ceil(float64(h.total) * q))
	var cum uint64
	lower := 0.0
	for i, c := range h.counts {
		cum += c
		if cum >= target {
			if i < len(h.edges) {
				upper := h.edges[i]
				return (lower + upper) / 2.0
			}
			upper := h.edges[len(h.edges)-1] * 2
			return (h.edges[len(h.edges)-1] + upper) / 2
		}
		if i < len(h.edges) {
			lower = h.edges[i]
		}
	}
	return h.edges[len(h.edges)-1]
}

/* =======================
   Global counters (atomic)
   ======================= */

type Counters struct {
	// Control-plane totals
	SubConnOK   uint64
	SubConnFail uint64
	SubDisc     uint64
	SubSubErr   uint64

	PubConnOK   uint64
	PubConnFail uint64
	PubDisc     uint64

	// Connected state
	SubsConnected            int64
	SubsFullySubscribed      int64 // connected + subscribed to all topics successfully
	PubsConnected            int64
	SubsAttemptedTargetGauge int64 // last step target (for reporting)

	// Data-plane totals
	PubsSent uint64
	PubsErr  uint64
	MsgsRecv uint64
}

/* =======================
   Payload format
   ======================= */

// header is 20 bytes:
// [0:8]  publish_ts_unix_nano (uint64)
// [8:12] pub_id (uint32)
// [12:20] seq (uint64)
const payloadHeaderBytes = 20

/* =======================
   CSV writer (window rows)
   ======================= */

type WindowRow struct {
	TSUTC string

	EpochIndex  int
	BrokersUsed int

	StepIndex int

	TargetSubs int
	TargetPubs int

	ConnectedSubs       int64
	FullySubscribedSubs int64
	ConnectedPubs       int64

	// deltas (window)
	SubConnOK   uint64
	SubConnFail uint64
	SubDisc     uint64
	SubSubErr   uint64

	PubConnOK   uint64
	PubConnFail uint64
	PubDisc     uint64

	PubsSent uint64
	PubsErr  uint64
	MsgsRecv uint64
	Expected uint64

	DeliveryRatio float64

	P50 float64
	P95 float64
	P99 float64

	DiscPerMin float64

	SLA_Breached bool
	SLA_Reason   string
}

type WindowCSV struct {
	f  *os.File
	w  *csv.Writer
	mu sync.Mutex
}

func NewWindowCSV(path string) (*WindowCSV, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	w := csv.NewWriter(f)
	_ = w.Write([]string{
		"ts_utc",
		"epoch_index", "brokers_used",
		"step_index",
		"target_subs", "target_pubs",
		"connected_subs", "fully_subscribed_subs", "connected_pubs",
		"sub_conn_ok", "sub_conn_fail", "sub_disc", "sub_sub_err",
		"pub_conn_ok", "pub_conn_fail", "pub_disc",
		"pubs_sent", "pubs_err", "msgs_recv", "expected",
		"delivery_ratio",
		"p50_ms", "p95_ms", "p99_ms",
		"disconnects_per_min",
		"sla_breached", "sla_reason",
	})
	w.Flush()
	return &WindowCSV{f: f, w: w}, nil
}

func (c *WindowCSV) WriteRow(r WindowRow) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_ = c.w.Write([]string{
		r.TSUTC,
		strconv.Itoa(r.EpochIndex),
		strconv.Itoa(r.BrokersUsed),
		strconv.Itoa(r.StepIndex),
		strconv.Itoa(r.TargetSubs),
		strconv.Itoa(r.TargetPubs),
		strconv.FormatInt(r.ConnectedSubs, 10),
		strconv.FormatInt(r.FullySubscribedSubs, 10),
		strconv.FormatInt(r.ConnectedPubs, 10),

		strconv.FormatUint(r.SubConnOK, 10),
		strconv.FormatUint(r.SubConnFail, 10),
		strconv.FormatUint(r.SubDisc, 10),
		strconv.FormatUint(r.SubSubErr, 10),

		strconv.FormatUint(r.PubConnOK, 10),
		strconv.FormatUint(r.PubConnFail, 10),
		strconv.FormatUint(r.PubDisc, 10),

		strconv.FormatUint(r.PubsSent, 10),
		strconv.FormatUint(r.PubsErr, 10),
		strconv.FormatUint(r.MsgsRecv, 10),
		strconv.FormatUint(r.Expected, 10),

		fmt.Sprintf("%.6f", r.DeliveryRatio),

		fmt.Sprintf("%.3f", r.P50),
		fmt.Sprintf("%.3f", r.P95),
		fmt.Sprintf("%.3f", r.P99),

		fmt.Sprintf("%.3f", r.DiscPerMin),

		strconv.FormatBool(r.SLA_Breached),
		r.SLA_Reason,
	})
	c.w.Flush()
}

func (c *WindowCSV) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.w.Flush()
	return c.f.Close()
}

/* =======================
   Summary model
   ======================= */

type EpochSummary struct {
	EpochIndex  int       `json:"epoch_index"`
	BrokersUsed int       `json:"brokers_used"`
	StartedAt   time.Time `json:"started_at"`
	FinishedAt  time.Time `json:"finished_at"`

	MaxSustainableSubs int `json:"max_sustainable_subscribers"`

	FailAtSubs   int    `json:"failed_at_subscribers"`
	FailReason   string `json:"fail_reason"`
	StoppedBySLA bool   `json:"stopped_by_sla"`
}

type Summary struct {
	Config     Config         `json:"config"`
	StartedAt  time.Time      `json:"started_at"`
	FinishedAt time.Time      `json:"finished_at"`
	Epochs     []EpochSummary `json:"epochs"`

	StopReason string `json:"stop_reason"`
}

/* =======================
   MQTT client runners
   ======================= */

type ClientRunner struct {
	name string
	run  func(ctx context.Context)
}

func (c *ClientRunner) Run(ctx context.Context) { c.run(ctx) }

func mqttOpts(cfg Config, activeBrokers []Broker, clientID string, onConn func(), onLost func(err error)) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()

	host := cfg.BrokerHost
	port := cfg.BrokerPort
	if len(activeBrokers) > 0 {
		b := pickBrokerDeterministic(activeBrokers, clientID)
		host, port = b.Host, b.Port
	}
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", host, port))
	opts.SetClientID(clientID)

	// Benchmark-stable defaults
	opts.SetCleanSession(true)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetAutoReconnect(false)

	if onConn != nil {
		opts.OnConnect = func(_ mqtt.Client) { onConn() }
	}
	if onLost != nil {
		opts.OnConnectionLost = func(_ mqtt.Client, err error) { onLost(err) }
	}
	return opts
}

type LatSample struct {
	LatencyMs float64
}

func NewSubscriber(cfg Config, activeBrokers []Broker, id int, ctr *Counters, latCh chan<- LatSample) *ClientRunner {
	clientID := fmt.Sprintf("sub-%s-%d-%d", cfg.TestName, os.Getpid(), id)

	topics := make([]string, cfg.TopicCount)
	for i := 0; i < cfg.TopicCount; i++ {
		topics[i] = fmt.Sprintf("%s/%d", cfg.TopicPrefix, i)
	}

	// sampling: every N messages (per subscriber)
	sampleEvery := cfg.LatencySampleN
	if sampleEvery <= 0 {
		sampleEvery = 50
	}
	seed := time.Now().UnixNano() + int64(id)*1315423911
	rng := rand.New(rand.NewSource(seed))
	counter := 0

	return &ClientRunner{
		name: clientID,
		run: func(ctx context.Context) {
			var connected int32
			var fullySub int32

			opts := mqttOpts(cfg, activeBrokers, clientID,
				func() {
					if atomic.CompareAndSwapInt32(&connected, 0, 1) {
						atomic.AddInt64(&ctr.SubsConnected, 1)
					}
				},
				func(err error) {
					atomic.AddUint64(&ctr.SubDisc, 1)
					if atomic.CompareAndSwapInt32(&fullySub, 1, 0) {
						atomic.AddInt64(&ctr.SubsFullySubscribed, -1)
					}
					if atomic.CompareAndSwapInt32(&connected, 1, 0) {
						atomic.AddInt64(&ctr.SubsConnected, -1)
					}
				},
			)

			opts.SetDefaultPublishHandler(func(_ mqtt.Client, msg mqtt.Message) {
				atomic.AddUint64(&ctr.MsgsRecv, 1)

				payload := msg.Payload()
				if len(payload) < payloadHeaderBytes {
					return
				}
				pubNS := int64(binary.BigEndian.Uint64(payload[0:8]))
				recvNS := time.Now().UnixNano()
				latMs := float64(recvNS-pubNS) / 1e6

				// probabilistic + periodic sampling so all subs don't sync
				counter++
				doSample := false
				if sampleEvery <= 1 {
					doSample = true
				} else if counter%sampleEvery == 0 && rng.Intn(2) == 0 {
					doSample = true
				}
				if doSample {
					select {
					case latCh <- LatSample{LatencyMs: latMs}:
					default:
						// drop if congested
					}
				}
			})

			client := mqtt.NewClient(opts)
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				atomic.AddUint64(&ctr.SubConnFail, 1)
				return
			}
			atomic.AddUint64(&ctr.SubConnOK, 1)

			// Subscribe to all topics
			for _, t := range topics {
				if token := client.Subscribe(t, 0, nil); token.Wait() && token.Error() != nil {
					atomic.AddUint64(&ctr.SubSubErr, 1)
					atomic.AddUint64(&ctr.SubDisc, 1)
					client.Disconnect(100)
					if atomic.CompareAndSwapInt32(&connected, 1, 0) {
						atomic.AddInt64(&ctr.SubsConnected, -1)
					}
					return
				}
			}

			// Mark fully subscribed
			if atomic.CompareAndSwapInt32(&fullySub, 0, 1) {
				atomic.AddInt64(&ctr.SubsFullySubscribed, 1)
			}

			<-ctx.Done()
			client.Disconnect(200)
			if atomic.CompareAndSwapInt32(&fullySub, 1, 0) {
				atomic.AddInt64(&ctr.SubsFullySubscribed, -1)
			}
			if atomic.CompareAndSwapInt32(&connected, 1, 0) {
				atomic.AddInt64(&ctr.SubsConnected, -1)
			}
		},
	}
}

func NewPublisher(cfg Config, activeBrokers []Broker, id int, ctr *Counters) *ClientRunner {
	clientID := fmt.Sprintf("pub-%s-%d-%d", cfg.TestName, os.Getpid(), id)

	payloadBytes := cfg.PayloadBytes
	if payloadBytes < payloadHeaderBytes {
		payloadBytes = payloadHeaderBytes
	}

	// choose topic deterministically
	topicIdx := id % max(1, cfg.TopicCount)
	topic := fmt.Sprintf("%s/%d", cfg.TopicPrefix, topicIdx)
	pubID := uint32(id)

	return &ClientRunner{
		name: clientID,
		run: func(ctx context.Context) {
			var connected int32
			opts := mqttOpts(cfg, activeBrokers, clientID,
				func() {
					if atomic.CompareAndSwapInt32(&connected, 0, 1) {
						atomic.AddInt64(&ctr.PubsConnected, 1)
					}
				},
				func(err error) {
					atomic.AddUint64(&ctr.PubDisc, 1)
					if atomic.CompareAndSwapInt32(&connected, 1, 0) {
						atomic.AddInt64(&ctr.PubsConnected, -1)
					}
				},
			)

			client := mqtt.NewClient(opts)
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				atomic.AddUint64(&ctr.PubConnFail, 1)
				return
			}
			atomic.AddUint64(&ctr.PubConnOK, 1)

			interval := time.Duration(float64(time.Second) / math.Max(cfg.PubRatePerSec, 0.000001))
			t := time.NewTicker(interval)
			defer t.Stop()

			buf := make([]byte, payloadBytes)
			for i := payloadHeaderBytes; i < len(buf); i++ {
				buf[i] = byte('a' + (i % 26))
			}

			// random starting seq
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)*1337))
			seq := uint64(rng.Int63())

			for {
				select {
				case <-ctx.Done():
					client.Disconnect(200)
					if atomic.CompareAndSwapInt32(&connected, 1, 0) {
						atomic.AddInt64(&ctr.PubsConnected, -1)
					}
					return
				case <-t.C:
					seq++
					tsn := uint64(time.Now().UnixNano())
					binary.BigEndian.PutUint64(buf[0:8], tsn)
					binary.BigEndian.PutUint32(buf[8:12], pubID)
					binary.BigEndian.PutUint64(buf[12:20], seq)

					token := client.Publish(topic, 0, false, buf)
					ok := token.WaitTimeout(200 * time.Millisecond)
					if !ok || token.Error() != nil {
						atomic.AddUint64(&ctr.PubsErr, 1)
					} else {
						atomic.AddUint64(&ctr.PubsSent, 1)
					}
				}
			}
		},
	}
}

/* =======================
   Epoch runner (unchanged)
   ======================= */

type EpochResult struct {
	Summary EpochSummary
}

func runEpoch(parent context.Context, cfg Config, epochIndex int, activeBrokers []Broker, windowCSV *WindowCSV) (EpochResult, error) {
	start := time.Now()

	epochCtx, cancel := context.WithTimeout(parent, time.Duration(cfg.MaxEpochSec)*time.Second)
	defer cancel()

	var ctr Counters

	latCh := make(chan LatSample, 200000)

	// Windowed latency histogram accumulator for current window
	var windowLatMu sync.Mutex
	windowLat := NewLatencyHist()

	// Collect sampled latency
	go func() {
		for {
			select {
			case <-epochCtx.Done():
				return
			case s := <-latCh:
				windowLatMu.Lock()
				windowLat.Add(s.LatencyMs)
				windowLatMu.Unlock()
			}
		}
	}()

	// Start fixed publishers
	pubs := make([]*ClientRunner, 0, cfg.Publishers)
	for i := 0; i < cfg.Publishers; i++ {
		r := NewPublisher(cfg, activeBrokers, i, &ctr)
		pubs = append(pubs, r)
		go r.Run(epochCtx)
	}

	// Subscriber ramp
	targetSubs := cfg.InitialSubs
	if targetSubs < 0 {
		targetSubs = 0
	}
	if cfg.SubStep <= 0 {
		cfg.SubStep = 100
	}
	if cfg.WindowSec <= 0 {
		cfg.WindowSec = 60
	}
	if cfg.WarmupSec < 0 {
		cfg.WarmupSec = 0
	}

	subs := make([]*ClientRunner, 0, max(1, cfg.MaxSubs))
	var subsMu sync.Mutex

	// cumulative snapshots for deltas
	var prevSubConnOK, prevSubConnFail, prevSubDisc, prevSubSubErr uint64
	var prevPubConnOK, prevPubConnFail, prevPubDisc uint64
	var prevPubsSent, prevPubsErr, prevMsgsRecv uint64

	stepIdx := 0
	lastGoodSubs := 0
	failAt := 0
	failReason := ""
	breaches := 0

	for {
		select {
		case <-epochCtx.Done():
			// time cap or parent cancel
			goto done
		default:
		}

		// cap
		if cfg.MaxSubs > 0 && targetSubs > cfg.MaxSubs {
			targetSubs = cfg.MaxSubs
		}

		stepIdx++
		atomic.StoreInt64(&ctr.SubsAttemptedTargetGauge, int64(targetSubs))

		// Ensure subscribers up to targetSubs
		subsMu.Lock()
		for len(subs) < targetSubs {
			id := len(subs)
			r := NewSubscriber(cfg, activeBrokers, id, &ctr, latCh)
			subs = append(subs, r)
			go r.Run(epochCtx)
		}
		subsMu.Unlock()

		// Hold this step for WindowSec, evaluating SLA each second, but only deciding at end-of-window.
		// We compute ONE window row per step.
		windowStart := time.Now()
		warmupUntil := windowStart.Add(time.Duration(cfg.WarmupSec) * time.Second)

		// Reset window latency hist
		windowLatMu.Lock()
		windowLat = NewLatencyHist()
		windowLatMu.Unlock()

		// For disconnects/min calculation in this window, use window delta of sub+pub disc.
		discStartSub := atomic.LoadUint64(&ctr.SubDisc)
		discStartPub := atomic.LoadUint64(&ctr.PubDisc)

		// Wait for window duration
		select {
		case <-epochCtx.Done():
			goto done
		case <-time.After(time.Duration(cfg.WindowSec) * time.Second):
		}

		windowEnd := time.Now()
		windowDur := windowEnd.Sub(windowStart)

		// Connected gauges at end of window
		connSubs := atomic.LoadInt64(&ctr.SubsConnected)
		fullSubs := atomic.LoadInt64(&ctr.SubsFullySubscribed)
		connPubs := atomic.LoadInt64(&ctr.PubsConnected)

		// Load cumulative totals
		curSubConnOK := atomic.LoadUint64(&ctr.SubConnOK)
		curSubConnFail := atomic.LoadUint64(&ctr.SubConnFail)
		curSubDisc := atomic.LoadUint64(&ctr.SubDisc)
		curSubSubErr := atomic.LoadUint64(&ctr.SubSubErr)

		curPubConnOK := atomic.LoadUint64(&ctr.PubConnOK)
		curPubConnFail := atomic.LoadUint64(&ctr.PubConnFail)
		curPubDisc := atomic.LoadUint64(&ctr.PubDisc)

		curPubsSent := atomic.LoadUint64(&ctr.PubsSent)
		curPubsErr := atomic.LoadUint64(&ctr.PubsErr)
		curMsgsRecv := atomic.LoadUint64(&ctr.MsgsRecv)

		// Deltas for this step window
		dSubConnOK := curSubConnOK - prevSubConnOK
		dSubConnFail := curSubConnFail - prevSubConnFail
		dSubDisc := curSubDisc - prevSubDisc
		dSubSubErr := curSubSubErr - prevSubSubErr

		dPubConnOK := curPubConnOK - prevPubConnOK
		dPubConnFail := curPubConnFail - prevPubConnFail
		dPubDisc := curPubDisc - prevPubDisc

		dPubsSent := curPubsSent - prevPubsSent
		dPubsErr := curPubsErr - prevPubsErr
		dMsgsRecv := curMsgsRecv - prevMsgsRecv

		// Expected deliveries for this window:
		// subscribers subscribe to all topics => each publish should reach each FULLY subscribed connected sub.
		expected := uint64(0)
		if fullSubs > 0 && dPubsSent > 0 {
			expected = dPubsSent * uint64(fullSubs)
		}

		deliveryRatio := 0.0
		if expected > 0 {
			deliveryRatio = float64(dMsgsRecv) / float64(expected)
		}

		// Window latency quantiles (sampled)
		windowLatMu.Lock()
		p50 := windowLat.Quantile(0.50)
		p95 := windowLat.Quantile(0.95)
		p99 := windowLat.Quantile(0.99)
		windowLatMu.Unlock()

		// Window disconnects per minute (sub + pub)
		discEndSub := atomic.LoadUint64(&ctr.SubDisc)
		discEndPub := atomic.LoadUint64(&ctr.PubDisc)
		discDelta := (discEndSub - discStartSub) + (discEndPub - discStartPub)

		discPerMin := 0.0
		if windowDur > 0 {
			discPerMin = float64(discDelta) / windowDur.Minutes()
		}

		// SLA evaluation
		slaBreached := false
		slaReason := ""
		if cfg.WarmupSec >= cfg.WindowSec {
			// skip SLA evaluation entirely
		} else if time.Now().After(warmupUntil) {
			connectedPct := 100.0
			if targetSubs > 0 {
				connectedPct = (float64(connSubs) / float64(targetSubs)) * 100.0
			}
			deliveryPct := 100.0 * deliveryRatio

			// Prioritized reason (for paper "failure mode")
			if targetSubs > 0 && connectedPct < cfg.MinConnectedSubPct {
				slaBreached = true
				slaReason = fmt.Sprintf("connected_subs_pct=%.2f < %.2f", connectedPct, cfg.MinConnectedSubPct)
			} else if expected > 0 && deliveryPct < cfg.MinDeliveryRatioPct {
				slaBreached = true
				slaReason = fmt.Sprintf("delivery_pct=%.2f < %.2f", deliveryPct, cfg.MinDeliveryRatioPct)
			} else if cfg.MaxP95LatencyMs > 0 && p95 > cfg.MaxP95LatencyMs {
				slaBreached = true
				slaReason = fmt.Sprintf("p95_ms=%.2f > %.2f", p95, cfg.MaxP95LatencyMs)
			} else if cfg.MaxDiscPerMin > 0 && discPerMin > cfg.MaxDiscPerMin {
				slaBreached = true
				slaReason = fmt.Sprintf("disconnects_per_min=%.2f > %.2f", discPerMin, cfg.MaxDiscPerMin)
			}
		}

		// Write window row
		windowCSV.WriteRow(WindowRow{
			TSUTC: time.Now().UTC().Format(time.RFC3339),

			EpochIndex:  epochIndex,
			BrokersUsed: max(1, len(activeBrokers)),

			StepIndex: stepIdx,

			TargetSubs: targetSubs,
			TargetPubs: cfg.Publishers,

			ConnectedSubs:       connSubs,
			FullySubscribedSubs: fullSubs,
			ConnectedPubs:       connPubs,

			SubConnOK:   dSubConnOK,
			SubConnFail: dSubConnFail,
			SubDisc:     dSubDisc,
			SubSubErr:   dSubSubErr,

			PubConnOK:   dPubConnOK,
			PubConnFail: dPubConnFail,
			PubDisc:     dPubDisc,

			PubsSent: dPubsSent,
			PubsErr:  dPubsErr,
			MsgsRecv: dMsgsRecv,
			Expected: expected,

			DeliveryRatio: deliveryRatio,

			P50: p50,
			P95: p95,
			P99: p99,

			DiscPerMin: discPerMin,

			SLA_Breached: slaBreached,
			SLA_Reason:   slaReason,
		})

		// Update prev snapshots
		prevSubConnOK, prevSubConnFail, prevSubDisc, prevSubSubErr = curSubConnOK, curSubConnFail, curSubDisc, curSubSubErr
		prevPubConnOK, prevPubConnFail, prevPubDisc = curPubConnOK, curPubConnFail, curPubDisc
		prevPubsSent, prevPubsErr, prevMsgsRecv = curPubsSent, curPubsErr, curMsgsRecv

		// SLA breach logic with consecutive windows
		if slaBreached {
			breaches++
		} else {
			breaches = 0
			// Only mark last-good when the window is good
			lastGoodSubs = targetSubs
		}

		if breaches >= max(1, cfg.ConsecutiveBreaches) {
			failAt = targetSubs
			failReason = slaReason
			goto done
		}

		// Next step
		if cfg.MaxSubs > 0 && targetSubs >= cfg.MaxSubs {
			goto done
		}
		targetSubs += cfg.SubStep
	}

done:
	cancel() // stop clients

	finished := time.Now()
	es := EpochSummary{
		EpochIndex:  epochIndex,
		BrokersUsed: max(1, len(activeBrokers)),
		StartedAt:   start,
		FinishedAt:  finished,

		MaxSustainableSubs: lastGoodSubs,
		FailAtSubs:         failAt,
		FailReason:         failReason,
		StoppedBySLA:       failReason != "",
	}
	_ = pubs // silence (publishers run under ctx)
	return EpochResult{Summary: es}, nil
}

/* =======================
   NEW S3 runner: Hot-add (new clients only), no restart, no rebalance
   ======================= */

func runHotAddNewClientsOnly(parent context.Context, cfg Config, brokerPool []Broker, windowCSV *WindowCSV) ([]EpochSummary, string, error) {
	start := time.Now()

	runCtx, cancel := context.WithTimeout(parent, time.Duration(cfg.MaxTotalSec)*time.Second)
	defer cancel()

	var ctr Counters

	latCh := make(chan LatSample, 200000)

	var windowLatMu sync.Mutex
	windowLat := NewLatencyHist()

	go func() {
		for {
			select {
			case <-runCtx.Done():
				return
			case s := <-latCh:
				windowLatMu.Lock()
				windowLat.Add(s.LatencyMs)
				windowLatMu.Unlock()
			}
		}
	}()

	// Publishers fixed for entire run; pin to broker[0] for the vertical/horizontal story.
	pubBroker := []Broker{brokerPool[0]}
	pubs := make([]*ClientRunner, 0, cfg.Publishers)
	for i := 0; i < cfg.Publishers; i++ {
		r := NewPublisher(cfg, pubBroker, i, &ctr)
		pubs = append(pubs, r)
		go r.Run(runCtx)
	}

	// Subscriber ramp parameters (same defaults as runEpoch)
	targetSubs := cfg.InitialSubs
	if targetSubs < 0 {
		targetSubs = 0
	}
	if cfg.SubStep <= 0 {
		cfg.SubStep = 100
	}
	if cfg.WindowSec <= 0 {
		cfg.WindowSec = 60
	}
	if cfg.WarmupSec < 0 {
		cfg.WarmupSec = 0
	}

	subs := make([]*ClientRunner, 0, max(1, cfg.MaxSubs))
	var subsMu sync.Mutex

	// cumulative snapshots for deltas
	var prevSubConnOK, prevSubConnFail, prevSubDisc, prevSubSubErr uint64
	var prevPubConnOK, prevPubConnFail, prevPubDisc uint64
	var prevPubsSent, prevPubsErr, prevMsgsRecv uint64

	stageIdx := 0
	stageStartAt := start
	stageLastGoodSubs := 0
	stageFailAt := 0
	stageFailReason := ""
	breaches := 0

	stepIdx := 0
	var summaries []EpochSummary

	for {
		select {
		case <-runCtx.Done():
			goto done
		default:
		}

		if cfg.MaxSubs > 0 && targetSubs > cfg.MaxSubs {
			targetSubs = cfg.MaxSubs
		}

		stepIdx++
		atomic.StoreInt64(&ctr.SubsAttemptedTargetGauge, int64(targetSubs))

		// Ensure subscribers up to targetSubs.
		// ONLY newly created subscribers are pinned to brokerPool[stageIdx].
		stageBroker := []Broker{brokerPool[stageIdx]}
		subsMu.Lock()
		for len(subs) < targetSubs {
			id := len(subs)
			r := NewSubscriber(cfg, stageBroker, id, &ctr, latCh)
			subs = append(subs, r)
			go r.Run(runCtx)
		}
		subsMu.Unlock()

		windowStart := time.Now()
		warmupUntil := windowStart.Add(time.Duration(cfg.WarmupSec) * time.Second)

		windowLatMu.Lock()
		windowLat = NewLatencyHist()
		windowLatMu.Unlock()

		discStartSub := atomic.LoadUint64(&ctr.SubDisc)
		discStartPub := atomic.LoadUint64(&ctr.PubDisc)

		select {
		case <-runCtx.Done():
			goto done
		case <-time.After(time.Duration(cfg.WindowSec) * time.Second):
		}

		windowEnd := time.Now()
		windowDur := windowEnd.Sub(windowStart)

		connSubs := atomic.LoadInt64(&ctr.SubsConnected)
		fullSubs := atomic.LoadInt64(&ctr.SubsFullySubscribed)
		connPubs := atomic.LoadInt64(&ctr.PubsConnected)

		curSubConnOK := atomic.LoadUint64(&ctr.SubConnOK)
		curSubConnFail := atomic.LoadUint64(&ctr.SubConnFail)
		curSubDisc := atomic.LoadUint64(&ctr.SubDisc)
		curSubSubErr := atomic.LoadUint64(&ctr.SubSubErr)

		curPubConnOK := atomic.LoadUint64(&ctr.PubConnOK)
		curPubConnFail := atomic.LoadUint64(&ctr.PubConnFail)
		curPubDisc := atomic.LoadUint64(&ctr.PubDisc)

		curPubsSent := atomic.LoadUint64(&ctr.PubsSent)
		curPubsErr := atomic.LoadUint64(&ctr.PubsErr)
		curMsgsRecv := atomic.LoadUint64(&ctr.MsgsRecv)

		dSubConnOK := curSubConnOK - prevSubConnOK
		dSubConnFail := curSubConnFail - prevSubConnFail
		dSubDisc := curSubDisc - prevSubDisc
		dSubSubErr := curSubSubErr - prevSubSubErr

		dPubConnOK := curPubConnOK - prevPubConnOK
		dPubConnFail := curPubConnFail - prevPubConnFail
		dPubDisc := curPubDisc - prevPubDisc

		dPubsSent := curPubsSent - prevPubsSent
		dPubsErr := curPubsErr - prevPubsErr
		dMsgsRecv := curMsgsRecv - prevMsgsRecv

		expected := uint64(0)
		if fullSubs > 0 && dPubsSent > 0 {
			expected = dPubsSent * uint64(fullSubs)
		}
		deliveryRatio := 0.0
		if expected > 0 {
			deliveryRatio = float64(dMsgsRecv) / float64(expected)
		}

		windowLatMu.Lock()
		p50 := windowLat.Quantile(0.50)
		p95 := windowLat.Quantile(0.95)
		p99 := windowLat.Quantile(0.99)
		windowLatMu.Unlock()

		discEndSub := atomic.LoadUint64(&ctr.SubDisc)
		discEndPub := atomic.LoadUint64(&ctr.PubDisc)
		discDelta := (discEndSub - discStartSub) + (discEndPub - discStartPub)
		discPerMin := 0.0
		if windowDur > 0 {
			discPerMin = float64(discDelta) / windowDur.Minutes()
		}

		// SLA evaluation (same as runEpoch)
		slaBreached := false
		slaReason := ""
		if cfg.WarmupSec >= cfg.WindowSec {
			// skip
		} else if time.Now().After(warmupUntil) {
			connectedPct := 100.0
			if targetSubs > 0 {
				connectedPct = (float64(connSubs) / float64(targetSubs)) * 100.0
			}
			deliveryPct := 100.0 * deliveryRatio

			if targetSubs > 0 && connectedPct < cfg.MinConnectedSubPct {
				slaBreached = true
				slaReason = fmt.Sprintf("connected_subs_pct=%.2f < %.2f", connectedPct, cfg.MinConnectedSubPct)
			} else if expected > 0 && deliveryPct < cfg.MinDeliveryRatioPct {
				slaBreached = true
				slaReason = fmt.Sprintf("delivery_pct=%.2f < %.2f", deliveryPct, cfg.MinDeliveryRatioPct)
			} else if cfg.MaxP95LatencyMs > 0 && p95 > cfg.MaxP95LatencyMs {
				slaBreached = true
				slaReason = fmt.Sprintf("p95_ms=%.2f > %.2f", p95, cfg.MaxP95LatencyMs)
			} else if cfg.MaxDiscPerMin > 0 && discPerMin > cfg.MaxDiscPerMin {
				slaBreached = true
				slaReason = fmt.Sprintf("disconnects_per_min=%.2f > %.2f", discPerMin, cfg.MaxDiscPerMin)
			}
		}

		// Write row with epoch_index/brokers_used used to encode stage.
		windowCSV.WriteRow(WindowRow{
			TSUTC: time.Now().UTC().Format(time.RFC3339),

			EpochIndex:  stageIdx + 1,
			BrokersUsed: stageIdx + 1,

			StepIndex: stepIdx,

			TargetSubs: targetSubs,
			TargetPubs: cfg.Publishers,

			ConnectedSubs:       connSubs,
			FullySubscribedSubs: fullSubs,
			ConnectedPubs:       connPubs,

			SubConnOK:   dSubConnOK,
			SubConnFail: dSubConnFail,
			SubDisc:     dSubDisc,
			SubSubErr:   dSubSubErr,

			PubConnOK:   dPubConnOK,
			PubConnFail: dPubConnFail,
			PubDisc:     dPubDisc,

			PubsSent: dPubsSent,
			PubsErr:  dPubsErr,
			MsgsRecv: dMsgsRecv,
			Expected: expected,

			DeliveryRatio: deliveryRatio,

			P50: p50,
			P95: p95,
			P99: p99,

			DiscPerMin: discPerMin,

			SLA_Breached: slaBreached,
			SLA_Reason:   slaReason,
		})

		// update prev snapshots
		prevSubConnOK, prevSubConnFail, prevSubDisc, prevSubSubErr = curSubConnOK, curSubConnFail, curSubDisc, curSubSubErr
		prevPubConnOK, prevPubConnFail, prevPubDisc = curPubConnOK, curPubConnFail, curPubDisc
		prevPubsSent, prevPubsErr, prevMsgsRecv = curPubsSent, curPubsErr, curMsgsRecv

		if slaBreached {
			breaches++
		} else {
			breaches = 0
			stageLastGoodSubs = targetSubs
		}

		// On sustained breach: advance stage if possible, else stop
		if breaches >= max(1, cfg.ConsecutiveBreaches) {
			stageFailAt = targetSubs
			stageFailReason = slaReason

			if stageIdx < len(brokerPool)-1 {
				// record stage summary
				summaries = append(summaries, EpochSummary{
					EpochIndex:  stageIdx + 1,
					BrokersUsed: stageIdx + 1,
					StartedAt:   stageStartAt,
					FinishedAt:  time.Now(),

					MaxSustainableSubs: stageLastGoodSubs,
					FailAtSubs:         stageFailAt,
					FailReason:         stageFailReason,
					StoppedBySLA:       true,
				})

				// advance to next broker for NEW clients only
				stageIdx++
				stageStartAt = time.Now()
				stageLastGoodSubs = targetSubs
				stageFailAt = 0
				stageFailReason = ""
				breaches = 0

				// continue ramping
				if cfg.MaxSubs > 0 && targetSubs >= cfg.MaxSubs {
					goto done
				}
				targetSubs += cfg.SubStep
				continue
			}

			// last broker already in use -> stop
			goto done
		}

		if cfg.MaxSubs > 0 && targetSubs >= cfg.MaxSubs {
			goto done
		}
		targetSubs += cfg.SubStep
	}

done:
	cancel()
	_ = pubs // publishers run under ctx; silence

	// finalize last stage summary
	summaries = append(summaries, EpochSummary{
		EpochIndex:  stageIdx + 1,
		BrokersUsed: stageIdx + 1,
		StartedAt:   stageStartAt,
		FinishedAt:  time.Now(),

		MaxSustainableSubs: stageLastGoodSubs,
		FailAtSubs:         stageFailAt,
		FailReason:         stageFailReason,
		StoppedBySLA:       stageFailReason != "",
	})

	stopReason := "stopped: completed or time cap"
	if stageFailReason != "" && stageIdx == len(brokerPool)-1 {
		stopReason = "stopped: SLA failure on last broker (no more brokers to add)"
	}
	if runCtx.Err() == context.DeadlineExceeded || parent.Err() == context.DeadlineExceeded {
		stopReason = "stopped: max-total-sec reached"
	}
	return summaries, stopReason, nil
}

/* =======================
   Main
   ======================= */

func main() {
	var cfg Config

	// brokers
	flag.StringVar(&cfg.BrokerHost, "broker-host", getenv("BROKER_HOST", "127.0.0.1"), "Broker host (single mode)")
	flag.IntVar(&cfg.BrokerPort, "broker-port", getenvInt("BROKER_PORT", 1883), "Broker port (single mode)")
	flag.StringVar(&cfg.BrokersJSON, "brokers-json", getenv("BROKERS_JSON", ""), "Path to brokers.json for cluster mode")
	flag.BoolVar(&cfg.ClusterIncremental, "cluster-incremental", false, "If true and brokers-json set: start with 1 broker and add broker per epoch after SLA failure (RESTART EPOCHS; unchanged)")
	flag.BoolVar(&cfg.ClusterHotAddNewClients, "cluster-hot-add-new-clients", false, "NEW (S3): on SLA breach, advance broker stage and place ONLY new clients on next broker (no rebalance, no restart)")

	// output
	flag.StringVar(&cfg.TestName, "test-name", getenv("TEST_NAME", "mqtt-scale"), "Test name prefix for outputs")
	flag.StringVar(&cfg.OutDir, "out-dir", getenv("OUT_DIR", "./results"), "Output directory")

	// ramp
	flag.IntVar(&cfg.InitialSubs, "initial-subs", 500, "Initial subscribers N")
	flag.IntVar(&cfg.SubStep, "sub-step", 100, "Increase subscribers by this step (n+100)")
	flag.IntVar(&cfg.MaxSubs, "max-subs", 20000, "Max subscribers cap (safety)")
	flag.IntVar(&cfg.WindowSec, "window-sec", 60, "Hold window per step (seconds)")
	flag.IntVar(&cfg.WarmupSec, "warmup-sec", 10, "Warmup per step (seconds)")

	flag.IntVar(&cfg.MaxEpochSec, "max-epoch-sec", 1800, "Max seconds per broker-count epoch (safety)")
	flag.IntVar(&cfg.MaxTotalSec, "max-total-sec", 7200, "Max seconds total (safety)")

	// publishers
	flag.IntVar(&cfg.Publishers, "publishers", 50, "Fixed number of publishers")
	flag.Float64Var(&cfg.PubRatePerSec, "pub-rate", 1.0, "Publish rate per second per publisher (QoS0)")
	flag.StringVar(&cfg.TopicPrefix, "topic-prefix", "bench/topic", "Topic prefix")
	flag.IntVar(&cfg.TopicCount, "topic-count", 10, "Number of topics")
	flag.IntVar(&cfg.PayloadBytes, "payload-bytes", 100, "Payload size in bytes (10/100/1000 recommended)")
	flag.IntVar(&cfg.LatencySampleN, "lat-sample-every", 50, "Sample approx every N messages per subscriber for latency hist (1 = sample all)")

	// SLA thresholds
	flag.Float64Var(&cfg.MinConnectedSubPct, "sla-min-connected-sub-pct", 99.0, "SLA: connected subscribers >= this % of target")
	flag.Float64Var(&cfg.MinDeliveryRatioPct, "sla-min-delivery-pct", 99.0, "SLA: delivery ratio >= this % (window delivered/expected)")
	flag.Float64Var(&cfg.MaxP95LatencyMs, "sla-max-p95-ms", 500.0, "SLA: p95 latency <= this (ms)")
	flag.Float64Var(&cfg.MaxDiscPerMin, "sla-max-disc-per-min", 50.0, "SLA: disconnects per minute <= this")

	flag.IntVar(&cfg.ConsecutiveBreaches, "sla-consecutive-breaches", 2, "Stop after this many consecutive SLA-breaching windows")

	flag.Parse()

	// validate
	if cfg.TopicCount <= 0 {
		log.Fatalf("topic-count must be > 0")
	}
	if cfg.PayloadBytes <= 0 {
		log.Fatalf("payload-bytes must be > 0")
	}
	if cfg.WindowSec <= 0 {
		log.Fatalf("window-sec must be > 0")
	}
	if cfg.Publishers < 0 {
		log.Fatalf("publishers must be >= 0")
	}
	if cfg.InitialSubs < 0 || cfg.SubStep <= 0 {
		log.Fatalf("initial-subs must be >= 0 and sub-step > 0")
	}
	if cfg.MaxEpochSec <= 0 || cfg.MaxTotalSec <= 0 {
		log.Fatalf("max-epoch-sec and max-total-sec must be > 0")
	}
	if cfg.ConsecutiveBreaches <= 0 {
		cfg.ConsecutiveBreaches = 1
	}
	if cfg.ClusterHotAddNewClients && cfg.ClusterIncremental {
		log.Fatalf("choose only one of --cluster-hot-add-new-clients or --cluster-incremental")
	}
	if err := os.MkdirAll(cfg.OutDir, 0o755); err != nil {
		log.Fatalf("failed to create out dir: %v", err)
	}

	// load brokers.json if provided
	var brokerPool []Broker
	if strings.TrimSpace(cfg.BrokersJSON) != "" {
		bf, err := loadBrokersFile(cfg.BrokersJSON)
		if err != nil {
			log.Fatalf("failed to load brokers-json: %v", err)
		}
		if len(bf.Brokers) == 0 {
			log.Fatalf("brokers-json has no brokers")
		}
		brokerPool = bf.Brokers
	}

	// output writers
	windowsCSVPath := filepath.Join(cfg.OutDir, cfg.TestName+".windows.csv")
	summaryPath := filepath.Join(cfg.OutDir, cfg.TestName+".summary.json")

	windowCSV, err := NewWindowCSV(windowsCSVPath)
	if err != nil {
		log.Fatalf("failed to open windows csv: %v", err)
	}
	defer func() { _ = windowCSV.Close() }()

	start := time.Now()
	parentCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.MaxTotalSec)*time.Second)
	defer cancel()

	var epochs []EpochSummary
	stopReason := ""

	// Decide plan:
	// - If no brokers.json => single epoch (activeBrokers empty => uses cfg.BrokerHost:Port)
	// - If brokers.json and hot-add-new-clients => S3 new logic
	// - If brokers.json and cluster-incremental => restart epochs k=1..B (unchanged)
	// - If brokers.json and NOT incremental => one epoch with all brokers (unchanged)

	if len(brokerPool) == 0 {
		log.Printf("mode=single broker=%s:%d publishers=%d pub_rate=%.3f payload=%dB",
			cfg.BrokerHost, cfg.BrokerPort, cfg.Publishers, cfg.PubRatePerSec, cfg.PayloadBytes)

		res, _ := runEpoch(parentCtx, cfg, 1, nil, windowCSV)
		epochs = append(epochs, res.Summary)
		if res.Summary.StoppedBySLA {
			stopReason = "stopped: SLA failure"
		} else {
			stopReason = "stopped: completed or time cap"
		}
	} else {
		if cfg.ClusterHotAddNewClients {
			log.Printf("mode=cluster hot-add-new-clients-only total_brokers=%d publishers=%d pub_rate=%.3f payload=%dB",
				len(brokerPool), cfg.Publishers, cfg.PubRatePerSec, cfg.PayloadBytes)

			ep, reason, _ := runHotAddNewClientsOnly(parentCtx, cfg, brokerPool, windowCSV)
			epochs = append(epochs, ep...)
			stopReason = reason

		} else if cfg.ClusterIncremental {
			log.Printf("mode=cluster incremental total_brokers=%d publishers=%d pub_rate=%.3f payload=%dB",
				len(brokerPool), cfg.Publishers, cfg.PubRatePerSec, cfg.PayloadBytes)

			for k := 1; k <= len(brokerPool); k++ {
				select {
				case <-parentCtx.Done():
					stopReason = "stopped: max-total-sec reached"
					goto finish
				default:
				}

				active := brokerPool[:k]
				log.Printf("epoch=%d brokers_used=%d", k, k)

				res, _ := runEpoch(parentCtx, cfg, k, active, windowCSV)
				epochs = append(epochs, res.Summary)

				// Requirement for restart-epoch incremental mode: add broker AFTER SLA failure
				if !res.Summary.StoppedBySLA {
					stopReason = fmt.Sprintf("stopped: epoch brokers_used=%d did not breach SLA (hit cap/time)", k)
					goto finish
				}
				if k == len(brokerPool) {
					stopReason = "stopped: SLA failure even with all brokers active"
					goto finish
				}
			}

		} else {
			log.Printf("mode=cluster fixed brokers_used=%d publishers=%d pub_rate=%.3f payload=%dB",
				len(brokerPool), cfg.Publishers, cfg.PubRatePerSec, cfg.PayloadBytes)

			res, _ := runEpoch(parentCtx, cfg, 1, brokerPool, windowCSV)
			epochs = append(epochs, res.Summary)
			if res.Summary.StoppedBySLA {
				stopReason = "stopped: SLA failure"
			} else {
				stopReason = "stopped: completed or time cap"
			}
		}
	}

finish:
	end := time.Now()
	s := Summary{
		Config:     cfg,
		StartedAt:  start,
		FinishedAt: end,
		Epochs:     epochs,
		StopReason: stopReason,
	}

	b, _ := json.MarshalIndent(s, "", "  ")
	if err := os.WriteFile(summaryPath, b, 0o644); err != nil {
		log.Printf("failed to write summary json: %v", err)
	} else {
		log.Printf("wrote %s", summaryPath)
	}
	log.Printf("wrote %s", windowsCSVPath)
	log.Printf("done: %s", stopReason)
}

/* =======================
   Helpers
   ======================= */

func getenv(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

func getenvInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
