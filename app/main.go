package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
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

type Config struct {
	BrokerHost string `json:"broker_host"`
	BrokerPort int    `json:"broker_port"`
	TestName   string `json:"test_name"`
	OutDir     string `json:"out_dir"`

	// Load profile
	RampIntervalSec int `json:"ramp_interval_sec"`
	RampStepSubs    int `json:"ramp_step_subscribers"`
	RampStepPubs    int `json:"ramp_step_publishers"`
	MaxSubs         int `json:"max_subscribers"`
	MaxPubs         int `json:"max_publishers"`

	// Publish profile
	PubRatePerSec float64 `json:"pub_rate_per_sec_per_publisher"`
	TopicPrefix   string  `json:"topic_prefix"`
	TopicCount    int     `json:"topic_count"` // publishers distribute across topics

	// Payload
	PayloadKB int `json:"payload_kb"`

	// Stop conditions
	MaxConnFailPct float64 `json:"max_conn_fail_pct"` // per ramp window
	MaxDiscPerMin  float64 `json:"max_disconnects_per_min"`
	MinHoldSteps   int     `json:"min_hold_steps"` // require at least N steps before stop triggers
}

type Summary struct {
	Config      Config    `json:"config"`
	StartedAt   time.Time `json:"started_at"`
	FinishedAt  time.Time `json:"finished_at"`
	DurationSec float64   `json:"duration_sec"`

	PeakConnectedSubs int `json:"peak_connected_subscribers"`
	PeakConnectedPubs int `json:"peak_connected_publishers"`

	StopReason string `json:"stop_reason"`

	TotalConnOK   uint64 `json:"total_conn_ok"`
	TotalConnFail uint64 `json:"total_conn_fail"`
	TotalDisc     uint64 `json:"total_disconnects"`

	TotalPubsSent uint64 `json:"total_pubs_sent"`
	TotalPubsErr  uint64 `json:"total_pubs_err"`
	TotalMsgsRecv uint64 `json:"total_msgs_received"`

	LatencyP50Ms float64 `json:"latency_p50_ms"`
	LatencyP95Ms float64 `json:"latency_p95_ms"`
	LatencyP99Ms float64 `json:"latency_p99_ms"`
}

type Metrics struct {
	ConnOK   uint64
	ConnFail uint64
	Disc     uint64

	SubsConnected int64
	PubsConnected int64

	PubsSent uint64
	PubsErr  uint64
	MsgsRecv uint64

	// latency samples aggregated via histogram
	LatHist LatencyHist
}

type LatencyHist struct {
	// bucket upper bounds in ms
	edges []float64
	// counts per bucket (len = len(edges)+1)
	counts []uint64
	total  uint64
}

func NewLatencyHist() LatencyHist {
	edges := []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000}
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
	for i, c := range h.counts {
		cum += c
		if cum >= target {
			if i < len(h.edges) {
				return h.edges[i]
			}
			// overflow bucket: approximate as last edge * 2
			return h.edges[len(h.edges)-1] * 2
		}
	}
	return h.edges[len(h.edges)-1] * 2
}

type CSVRow struct {
	TS string

	Step int

	TargetSubs int
	TargetPubs int

	ConnectedSubs int64
	ConnectedPubs int64

	ConnOK   uint64
	ConnFail uint64
	Disc     uint64

	PubsSent uint64
	PubsErr  uint64
	MsgsRecv uint64

	RecvRate float64

	P50 float64
	P95 float64
	P99 float64
}

type CSVWriter struct {
	f *os.File
}

func NewCSVWriter(path string) (*CSVWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	// header
	_, _ = f.WriteString(strings.Join([]string{
		"ts", "step",
		"target_subs", "target_pubs",
		"connected_subs", "connected_pubs",
		"conn_ok", "conn_fail", "disconnects",
		"pubs_sent", "pubs_err", "msgs_recv",
		"recv_rate_msgs_per_s",
		"p50_ms", "p95_ms", "p99_ms",
	}, ",") + "\n")
	return &CSVWriter{f: f}, nil
}

func (w *CSVWriter) WriteRow(r CSVRow) {
	line := fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%.3f,%.3f,%.3f,%.3f\n",
		r.TS, r.Step, r.TargetSubs, r.TargetPubs,
		r.ConnectedSubs, r.ConnectedPubs,
		r.ConnOK, r.ConnFail, r.Disc,
		r.PubsSent, r.PubsErr, r.MsgsRecv,
		r.RecvRate,
		r.P50, r.P95, r.P99,
	)
	_, _ = w.f.WriteString(line)
}

func (w *CSVWriter) Close() error { return w.f.Close() }

func main() {
	var cfg Config

	flag.StringVar(&cfg.BrokerHost, "broker-host", getenv("BROKER_HOST", "127.0.0.1"), "Broker IP/hostname")
	flag.IntVar(&cfg.BrokerPort, "broker-port", getenvInt("BROKER_PORT", 1883), "Broker port")
	flag.StringVar(&cfg.TestName, "test-name", getenv("TEST_NAME", "baseline"), "Test name (used for output filenames)")
	flag.StringVar(&cfg.OutDir, "out-dir", getenv("OUT_DIR", "/app/results"), "Output directory for JSON/CSV")

	flag.IntVar(&cfg.RampIntervalSec, "ramp-interval", 10, "Ramp interval in seconds")
	flag.IntVar(&cfg.RampStepSubs, "ramp-step-subs", 200, "Add this many subscribers per ramp interval")
	flag.IntVar(&cfg.RampStepPubs, "ramp-step-pubs", 5, "Add this many publishers per ramp interval")

	flag.IntVar(&cfg.MaxSubs, "max-subs", 5000, "Max subscribers to attempt")
	flag.IntVar(&cfg.MaxPubs, "max-pubs", 200, "Max publishers to attempt")

	flag.Float64Var(&cfg.PubRatePerSec, "pub-rate", 1.0, "Publish rate per second per publisher (QoS0)")
	flag.StringVar(&cfg.TopicPrefix, "topic-prefix", "bench/topic", "Topic prefix")
	flag.IntVar(&cfg.TopicCount, "topic-count", 10, "Number of topics to spread publishers across")

	flag.IntVar(&cfg.PayloadKB, "payload-kb", 10, "Payload size in KB (default 10). Example: --payload-kb=100")
	flag.Float64Var(&cfg.MaxConnFailPct, "max-conn-fail-pct", 2.0, "Stop if connection fail %% exceeds this within a ramp window")
	flag.Float64Var(&cfg.MaxDiscPerMin, "max-disc-per-min", 50.0, "Stop if disconnects per minute exceed this")
	flag.IntVar(&cfg.MinHoldSteps, "min-hold-steps", 3, "Do not stop before this many ramp steps")
	flag.Parse()

	if cfg.PayloadKB <= 0 {
		log.Fatalf("payload-kb must be > 0")
	}

	if err := os.MkdirAll(cfg.OutDir, 0o755); err != nil {
		log.Fatalf("failed to create out dir: %v", err)
	}

	csvPath := filepath.Join(cfg.OutDir, cfg.TestName+".csv")
	jsonPath := filepath.Join(cfg.OutDir, cfg.TestName+".json")

	csvw, err := NewCSVWriter(csvPath)
	if err != nil {
		log.Fatalf("failed to open csv: %v", err)
	}
	defer func() { _ = csvw.Close() }()

	log.Printf("loadtest: starting test=%s broker=%s:%d payload=%dKB qos=0",
		cfg.TestName, cfg.BrokerHost, cfg.BrokerPort, cfg.PayloadKB)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var m Metrics
	m.LatHist = NewLatencyHist()

	latCh := make(chan float64, 100000) // ms
	go func() {
		for ms := range latCh {
			m.LatHist.Add(ms)
		}
	}()

	// Keep references so we can stop them
	var subsMu sync.Mutex
	var pubsMu sync.Mutex
	subs := make([]*ClientRunner, 0, cfg.MaxSubs)
	pubs := make([]*ClientRunner, 0, cfg.MaxPubs)

	start := time.Now()
	peakSubs := int64(0)
	peakPubs := int64(0)

	// per-window counters
	var prevMsgsRecv uint64
	var stopReason string

	step := 0
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// ramp timer
	ramp := time.NewTicker(time.Duration(cfg.RampIntervalSec) * time.Second)
	defer ramp.Stop()

	// snapshot state for per-second CSV
	targetSubs := 0
	targetPubs := 0

	// track per-window increments
	var windowConnOK, windowConnFail, windowDisc uint64
	windowStart := time.Now()

	// drain metrics every second & write CSV
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ok := atomic.LoadUint64(&m.ConnOK)
				fail := atomic.LoadUint64(&m.ConnFail)
				disc := atomic.LoadUint64(&m.Disc)
				pubsSent := atomic.LoadUint64(&m.PubsSent)
				pubsErr := atomic.LoadUint64(&m.PubsErr)
				msgsRecv := atomic.LoadUint64(&m.MsgsRecv)

				connSubs := atomic.LoadInt64(&m.SubsConnected)
				connPubs := atomic.LoadInt64(&m.PubsConnected)

				if connSubs > peakSubs {
					peakSubs = connSubs
				}
				if connPubs > peakPubs {
					peakPubs = connPubs
				}

				recvRate := float64(msgsRecv-prevMsgsRecv) / 1.0
				prevMsgsRecv = msgsRecv

				p50 := m.LatHist.Quantile(0.50)
				p95 := m.LatHist.Quantile(0.95)
				p99 := m.LatHist.Quantile(0.99)

				csvw.WriteRow(CSVRow{
					TS: time.Now().UTC().Format(time.RFC3339),
					Step: step,

					TargetSubs: targetSubs,
					TargetPubs: targetPubs,

					ConnectedSubs: connSubs,
					ConnectedPubs: connPubs,

					ConnOK:   ok,
					ConnFail: fail,
					Disc:     disc,

					PubsSent: pubsSent,
					PubsErr:  pubsErr,
					MsgsRecv: msgsRecv,

					RecvRate: recvRate,

					P50: p50,
					P95: p95,
					P99: p99,
				})
			}
		}
	}()

	// main loop: ramp
	for {
		select {
		case <-ctx.Done():
			goto done
		case <-ramp.C:
			step++

			// update window counters since last step
			curOK := atomic.LoadUint64(&m.ConnOK)
			curFail := atomic.LoadUint64(&m.ConnFail)
			curDisc := atomic.LoadUint64(&m.Disc)

			// delta since last window start snapshot
			dOK := curOK - windowConnOK
			dFail := curFail - windowConnFail
			dDisc := curDisc - windowDisc

			// Update baseline markers for next window
			windowConnOK = curOK
			windowConnFail = curFail
			windowDisc = curDisc

			windowDur := time.Since(windowStart)
			windowStart = time.Now()

			// Evaluate stop conditions (after a few steps)
			if step >= cfg.MinHoldSteps {
				totalAttempts := dOK + dFail
				failPct := 0.0
				if totalAttempts > 0 {
					failPct = (float64(dFail) / float64(totalAttempts)) * 100.0
				}
				discPerMin := 0.0
				if windowDur > 0 {
					discPerMin = float64(dDisc) / windowDur.Minutes()
				}

				if failPct > cfg.MaxConnFailPct {
					stopReason = fmt.Sprintf("stop: conn_fail_pct=%.2f%% > %.2f%% in last ramp window",
						failPct, cfg.MaxConnFailPct)
					cancel()
					continue
				}
				if discPerMin > cfg.MaxDiscPerMin {
					stopReason = fmt.Sprintf("stop: disconnects_per_min=%.2f > %.2f in last ramp window",
						discPerMin, cfg.MaxDiscPerMin)
					cancel()
					continue
				}
			}

			// Ramp targets up
			if targetSubs < cfg.MaxSubs {
				targetSubs = min(cfg.MaxSubs, targetSubs+cfg.RampStepSubs)
			}
			if targetPubs < cfg.MaxPubs {
				targetPubs = min(cfg.MaxPubs, targetPubs+cfg.RampStepPubs)
			}

			// Ensure we have target number of subscribers
			subsMu.Lock()
			for len(subs) < targetSubs {
				id := len(subs)
				r := NewSubscriber(cfg, id, &m, latCh)
				subs = append(subs, r)
				go r.Run(ctx)
			}
			subsMu.Unlock()

			// Ensure we have target number of publishers
			pubsMu.Lock()
			for len(pubs) < targetPubs {
				id := len(pubs)
				r := NewPublisher(cfg, id, &m)
				pubs = append(pubs, r)
				go r.Run(ctx)
			}
			pubsMu.Unlock()

			log.Printf("loadtest: step=%d target_subs=%d target_pubs=%d connected_subs=%d connected_pubs=%d",
				step, targetSubs, targetPubs,
				atomic.LoadInt64(&m.SubsConnected), atomic.LoadInt64(&m.PubsConnected),
			)
		}
	}

done:
	// stop everything
	cancel()
	time.Sleep(500 * time.Millisecond) // give goroutines a beat
	close(latCh)

	end := time.Now()

	if stopReason == "" {
		stopReason = "stopped: context cancelled"
	}

	s := Summary{
		Config:      cfg,
		StartedAt:   start,
		FinishedAt:  end,
		DurationSec: end.Sub(start).Seconds(),

		PeakConnectedSubs: int(peakSubs),
		PeakConnectedPubs: int(peakPubs),

		StopReason: stopReason,

		TotalConnOK:   atomic.LoadUint64(&m.ConnOK),
		TotalConnFail: atomic.LoadUint64(&m.ConnFail),
		TotalDisc:     atomic.LoadUint64(&m.Disc),

		TotalPubsSent: atomic.LoadUint64(&m.PubsSent),
		TotalPubsErr:  atomic.LoadUint64(&m.PubsErr),
		TotalMsgsRecv: atomic.LoadUint64(&m.MsgsRecv),

		LatencyP50Ms: m.LatHist.Quantile(0.50),
		LatencyP95Ms: m.LatHist.Quantile(0.95),
		LatencyP99Ms: m.LatHist.Quantile(0.99),
	}

	b, _ := json.MarshalIndent(s, "", "  ")
	if err := os.WriteFile(jsonPath, b, 0o644); err != nil {
		log.Printf("loadtest: failed to write json: %v", err)
	} else {
		log.Printf("loadtest: wrote %s", jsonPath)
	}
	log.Printf("loadtest: wrote %s", csvPath)
	log.Printf("loadtest: %s", stopReason)
}

// -------------------- Client runners --------------------

type ClientRunner struct {
	name string
	run  func(ctx context.Context)
}

func (c *ClientRunner) Run(ctx context.Context) { c.run(ctx) }

func mqttOpts(cfg Config, clientID string, onConn func(), onLost func(err error)) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.BrokerHost, cfg.BrokerPort))
	opts.SetClientID(clientID)

	// Keep these minimal & stable for benchmarking
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

func NewSubscriber(cfg Config, id int, m *Metrics, latCh chan<- float64) *ClientRunner {
	clientID := fmt.Sprintf("sub-%s-%d-%d", cfg.TestName, os.Getpid(), id)
	topics := make([]string, cfg.TopicCount)
	for i := 0; i < cfg.TopicCount; i++ {
		topics[i] = fmt.Sprintf("%s/%d", cfg.TopicPrefix, i)
	}

	return &ClientRunner{
		name: clientID,
		run: func(ctx context.Context) {
			var connected int32

			opts := mqttOpts(cfg, clientID,
				func() {
					if atomic.CompareAndSwapInt32(&connected, 0, 1) {
						atomic.AddInt64(&m.SubsConnected, 1)
					}
				},
				func(err error) {
					atomic.AddUint64(&m.Disc, 1)
					if atomic.CompareAndSwapInt32(&connected, 1, 0) {
						atomic.AddInt64(&m.SubsConnected, -1)
					}
				},
			)

			// message handler
			opts.SetDefaultPublishHandler(func(_ mqtt.Client, msg mqtt.Message) {
				atomic.AddUint64(&m.MsgsRecv, 1)
				payload := msg.Payload()
				if len(payload) >= 16 {
					tsn := int64(binary.BigEndian.Uint64(payload[0:8]))
					// seq := binary.BigEndian.Uint64(payload[8:16]) // optional
					now := time.Now().UnixNano()
					latMs := float64(now-tsn) / 1e6
					select {
					case latCh <- latMs:
					default:
						// drop latency sample if channel congested
					}
				}
			})

			client := mqtt.NewClient(opts)

			// connect
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				atomic.AddUint64(&m.ConnFail, 1)
				return
			}
			atomic.AddUint64(&m.ConnOK, 1)

			// subscribe to all topics (QoS0)
			for _, t := range topics {
				if token := client.Subscribe(t, 0, nil); token.Wait() && token.Error() != nil {
					// treat as failure, disconnect
					atomic.AddUint64(&m.Disc, 1)
					client.Disconnect(100)
					if atomic.CompareAndSwapInt32(&connected, 1, 0) {
						atomic.AddInt64(&m.SubsConnected, -1)
					}
					return
				}
			}

			// hold
			<-ctx.Done()
			client.Disconnect(100)
			if atomic.CompareAndSwapInt32(&connected, 1, 0) {
				atomic.AddInt64(&m.SubsConnected, -1)
			}
		},
	}
}

func NewPublisher(cfg Config, id int, m *Metrics) *ClientRunner {
	clientID := fmt.Sprintf("pub-%s-%d-%d", cfg.TestName, os.Getpid(), id)

	// payload bytes
	payloadBytes := cfg.PayloadKB * 1024
	if payloadBytes < 16 {
		payloadBytes = 16
	}

	// Choose a stable topic per publisher (spreads load)
	topicIdx := id % max(1, cfg.TopicCount)
	topic := fmt.Sprintf("%s/%d", cfg.TopicPrefix, topicIdx)

	return &ClientRunner{
		name: clientID,
		run: func(ctx context.Context) {
			var connected int32

			opts := mqttOpts(cfg, clientID,
				func() {
					if atomic.CompareAndSwapInt32(&connected, 0, 1) {
						atomic.AddInt64(&m.PubsConnected, 1)
					}
				},
				func(err error) {
					atomic.AddUint64(&m.Disc, 1)
					if atomic.CompareAndSwapInt32(&connected, 1, 0) {
						atomic.AddInt64(&m.PubsConnected, -1)
					}
				},
			)

			client := mqtt.NewClient(opts)
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				atomic.AddUint64(&m.ConnFail, 1)
				return
			}
			atomic.AddUint64(&m.ConnOK, 1)

			// Rate limiter
			interval := time.Duration(float64(time.Second) / math.Max(cfg.PubRatePerSec, 0.000001))
			t := time.NewTicker(interval)
			defer t.Stop()

			// pre-allocate payload
			buf := make([]byte, payloadBytes)
			// fill with deterministic bytes for stable compression behavior (if any)
			for i := 16; i < len(buf); i++ {
				buf[i] = byte('a' + (i % 26))
			}

			var seq uint64 = uint64(rand.New(rand.NewSource(time.Now().UnixNano() + int64(id))).Int63())

			for {
				select {
				case <-ctx.Done():
					client.Disconnect(100)
					if atomic.CompareAndSwapInt32(&connected, 1, 0) {
						atomic.AddInt64(&m.PubsConnected, -1)
					}
					return
				case <-t.C:
					tsn := uint64(time.Now().UnixNano())
					seq++
					binary.BigEndian.PutUint64(buf[0:8], tsn)
					binary.BigEndian.PutUint64(buf[8:16], seq)

					token := client.Publish(topic, 0, false, buf)
					// QoS0 still returns a token; waiting ensures we detect client-side errors
					token.Wait()
					if token.Error() != nil {
						atomic.AddUint64(&m.PubsErr, 1)
					} else {
						atomic.AddUint64(&m.PubsSent, 1)
					}
				}
			}
		},
	}
}

// -------------------- helpers --------------------

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
