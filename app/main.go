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
   Config & Report Models
   ======================= */

type Config struct {
	// Single-broker mode
	BrokerHost string `json:"broker_host"`
	BrokerPort int    `json:"broker_port"`

	// Cluster mode (optional): if set, overrides broker-host/broker-port
	BrokersJSON string `json:"brokers_json,omitempty"`

	TestName string `json:"test_name"`
	OutDir   string `json:"out_dir"`

	// Load profile (optional)
	RampIntervalSec int `json:"ramp_interval_sec"`
	RampStepSubs    int `json:"ramp_step_subscribers"`
	RampStepPubs    int `json:"ramp_step_publishers"`
	MaxSubs         int `json:"max_subscribers"`
	MaxPubs         int `json:"max_publishers"`

	// Fixed load mode (paper-friendly)
	FixedSubs int `json:"fixed_subscribers"`
	FixedPubs int `json:"fixed_publishers"`
	WarmupSec int `json:"warmup_sec"`

	// Publish profile
	PubRatePerSec float64 `json:"pub_rate_per_sec_per_publisher"`
	TopicPrefix   string  `json:"topic_prefix"`
	TopicCount    int     `json:"topic_count"`

	// Payload
	PayloadKB int `json:"payload_kb"`

	// Stop conditions (SLA-ish)
	MinHoldSteps         int     `json:"min_hold_steps"`
	MaxConnFailPct       float64 `json:"max_conn_fail_pct"`
	MaxDiscPerMin        float64 `json:"max_disconnects_per_min"`
	MinConnectedSubPct   float64 `json:"min_connected_sub_pct"`    // % of target subs required
	MinDeliveryRatioPct  float64 `json:"min_delivery_ratio_pct"`   // delivered/expected * 100
	MaxP95LatencyMs      float64 `json:"max_p95_latency_ms"`       // SLA p95
	ConsecutiveSLABreaches int   `json:"consecutive_sla_breaches"` // consecutive windows before stop

	// Hard cap on duration
	MaxDurationSec int `json:"max_duration_sec"`
}

type Summary struct {
	Config      Config    `json:"config"`
	StartedAt   time.Time `json:"started_at"`
	FinishedAt  time.Time `json:"finished_at"`
	DurationSec float64   `json:"duration_sec"`

	PeakConnectedSubs int `json:"peak_connected_subscribers"`
	PeakConnectedPubs int `json:"peak_connected_publishers"`

	StopReason string `json:"stop_reason"`

	// Control-plane totals
	TotalSubConnOK   uint64 `json:"total_sub_conn_ok"`
	TotalSubConnFail uint64 `json:"total_sub_conn_fail"`
	TotalSubDisc     uint64 `json:"total_sub_disconnects"`
	TotalSubSubErr   uint64 `json:"total_sub_subscribe_errors"`

	TotalPubConnOK   uint64 `json:"total_pub_conn_ok"`
	TotalPubConnFail uint64 `json:"total_pub_conn_fail"`
	TotalPubDisc     uint64 `json:"total_pub_disconnects"`

	// Data-plane totals
	TotalPubsSent  uint64 `json:"total_pubs_sent"`
	TotalPubsErr   uint64 `json:"total_pubs_err"`
	TotalMsgsRecv  uint64 `json:"total_msgs_received"`
	TotalExpected  uint64 `json:"total_expected_deliveries"`
	TotalLost      uint64 `json:"total_lost_deliveries"`
	TotalDup       uint64 `json:"total_duplicate_deliveries"`
	TotalReorder   uint64 `json:"total_reordered_deliveries"`

	OverallDeliveryRatio float64 `json:"overall_delivery_ratio"` // TotalMsgsRecv / TotalExpected

	LatencyP50Ms float64 `json:"latency_p50_ms"`
	LatencyP95Ms float64 `json:"latency_p95_ms"`
	LatencyP99Ms float64 `json:"latency_p99_ms"`
	LatencyMeanMs float64 `json:"latency_mean_ms"`
	LatencyStdMs  float64 `json:"latency_std_ms"`
	LatencyMinMs  float64 `json:"latency_min_ms"`
	LatencyMaxMs  float64 `json:"latency_max_ms"`
}

/* =======================
   Cluster broker pool
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

/* =======================
   Metrics & Latency Histogram
   ======================= */

type Metrics struct {
	// Control-plane split
	SubConnOK   uint64
	SubConnFail uint64
	SubDisc     uint64
	SubSubErr   uint64

	PubConnOK   uint64
	PubConnFail uint64
	PubDisc     uint64

	SubsConnected int64
	PubsConnected int64

	// Data-plane
	PubsSent uint64
	PubsErr  uint64
	MsgsRecv uint64

	ExpectedDeliveries uint64
	LostDeliveries     uint64
	DupDeliveries      uint64
	ReorderDeliveries  uint64

	LatHist  LatencyHist
	LatStats OnlineStats
}

type LatencyHist struct {
	mu     sync.Mutex
	edges  []float64
	counts []uint64
	total  uint64
}

func NewLatencyHist() LatencyHist {
	// More granular, log-ish-ish edges in ms for better paper plots.
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
	h.mu.Lock()
	defer h.mu.Unlock()

	i := 0
	for i < len(h.edges) && ms > h.edges[i] {
		i++
	}
	h.counts[i]++
	h.total++
}

// Returns bucket-midpoint approximation for better stability than returning upper edge.
func (h *LatencyHist) Quantile(q float64) float64 {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.total == 0 {
		return 0
	}
	target := uint64(math.Ceil(float64(h.total) * q))
	var cum uint64

	var lower float64 = 0
	for i, c := range h.counts {
		cum += c
		if cum >= target {
			if i < len(h.edges) {
				upper := h.edges[i]
				return (lower + upper) / 2.0
			}
			// overflow bucket
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
   Online latency stats
   ======================= */

type OnlineStats struct {
	mu    sync.Mutex
	n     uint64
	mean  float64
	m2    float64
	min   float64
	max   float64
	inited bool
}

func (s *OnlineStats) Add(x float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.n++
	if !s.inited {
		s.min = x
		s.max = x
		s.mean = x
		s.m2 = 0
		s.inited = true
		return
	}
	if x < s.min {
		s.min = x
	}
	if x > s.max {
		s.max = x
	}
	delta := x - s.mean
	s.mean += delta / float64(s.n)
	delta2 := x - s.mean
	s.m2 += delta * delta2
}

func (s *OnlineStats) Snapshot() (n uint64, mean, std, min, max float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.n < 2 {
		var st float64
		if s.n == 1 {
			st = 0
		}
		return s.n, s.mean, st, s.min, s.max
	}
	variance := s.m2 / float64(s.n-1)
	return s.n, s.mean, math.Sqrt(variance), s.min, s.max
}

/* =======================
   CSV Writers
   ======================= */

type TrafficRow struct {
	TS string

	Step int

	TargetSubs int
	TargetPubs int

	ConnectedSubs int64
	ConnectedPubs int64

	// per-second deltas
	SubConnOKDelta   uint64
	SubConnFailDelta uint64
	SubDiscDelta     uint64
	SubSubErrDelta   uint64

	PubConnOKDelta   uint64
	PubConnFailDelta uint64
	PubDiscDelta     uint64

	PubsSentDelta uint64
	PubsErrDelta  uint64
	MsgsRecvDelta uint64

	ExpectedDelta uint64
	LostDelta     uint64
	DupDelta      uint64
	ReorderDelta  uint64

	DeliveryRatio float64 // MsgsRecvDelta / ExpectedDelta

	P50 float64
	P95 float64
	P99 float64
}

type TrafficCSVWriter struct {
	f  *os.File
	w  *csv.Writer
	mu sync.Mutex
}

func NewTrafficCSVWriter(path string) (*TrafficCSVWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	w := csv.NewWriter(f)
	header := []string{
		"ts", "step",
		"target_subs", "target_pubs",
		"connected_subs", "connected_pubs",
		"sub_conn_ok_delta", "sub_conn_fail_delta", "sub_disc_delta", "sub_subscribe_err_delta",
		"pub_conn_ok_delta", "pub_conn_fail_delta", "pub_disc_delta",
		"pubs_sent_delta", "pubs_err_delta", "msgs_recv_delta",
		"expected_deliveries_delta", "lost_delta", "dup_delta", "reorder_delta",
		"delivery_ratio",
		"p50_ms", "p95_ms", "p99_ms",
	}
	_ = w.Write(header)
	w.Flush()
	return &TrafficCSVWriter{f: f, w: w}, nil
}

func (w *TrafficCSVWriter) WriteRow(r TrafficRow) {
	w.mu.Lock()
	defer w.mu.Unlock()

	_ = w.w.Write([]string{
		r.TS,
		strconv.Itoa(r.Step),

		strconv.Itoa(r.TargetSubs),
		strconv.Itoa(r.TargetPubs),

		strconv.FormatInt(r.ConnectedSubs, 10),
		strconv.FormatInt(r.ConnectedPubs, 10),

		strconv.FormatUint(r.SubConnOKDelta, 10),
		strconv.FormatUint(r.SubConnFailDelta, 10),
		strconv.FormatUint(r.SubDiscDelta, 10),
		strconv.FormatUint(r.SubSubErrDelta, 10),

		strconv.FormatUint(r.PubConnOKDelta, 10),
		strconv.FormatUint(r.PubConnFailDelta, 10),
		strconv.FormatUint(r.PubDiscDelta, 10),

		strconv.FormatUint(r.PubsSentDelta, 10),
		strconv.FormatUint(r.PubsErrDelta, 10),
		strconv.FormatUint(r.MsgsRecvDelta, 10),

		strconv.FormatUint(r.ExpectedDelta, 10),
		strconv.FormatUint(r.LostDelta, 10),
		strconv.FormatUint(r.DupDelta, 10),
		strconv.FormatUint(r.ReorderDelta, 10),

		fmt.Sprintf("%.6f", r.DeliveryRatio),

		fmt.Sprintf("%.3f", r.P50),
		fmt.Sprintf("%.3f", r.P95),
		fmt.Sprintf("%.3f", r.P99),
	})
	w.w.Flush()
}

func (w *TrafficCSVWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.w.Flush()
	return w.f.Close()
}

type LatencySample struct {
	TSUTC   string
	RecvNS  int64
	PubNS   int64
	PubID   uint32
	Seq     uint64
	Latency float64
	Topic   string
}

type LatencyCSVWriter struct {
	f  *os.File
	w  *csv.Writer
	mu sync.Mutex
}

func NewLatencyCSVWriter(path string) (*LatencyCSVWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	w := csv.NewWriter(f)
	_ = w.Write([]string{
		"ts_utc",
		"recv_ts_unix_nano",
		"pub_ts_unix_nano",
		"pub_id",
		"seq",
		"latency_ms",
		"topic",
	})
	w.Flush()
	return &LatencyCSVWriter{f: f, w: w}, nil
}

func (w *LatencyCSVWriter) WriteSample(s LatencySample) {
	w.mu.Lock()
	defer w.mu.Unlock()

	_ = w.w.Write([]string{
		s.TSUTC,
		strconv.FormatInt(s.RecvNS, 10),
		strconv.FormatInt(s.PubNS, 10),
		strconv.FormatUint(uint64(s.PubID), 10),
		strconv.FormatUint(s.Seq, 10),
		fmt.Sprintf("%.3f", s.Latency),
		s.Topic,
	})
	w.w.Flush()
}

func (w *LatencyCSVWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.w.Flush()
	return w.f.Close()
}

/* =======================
   Main
   ======================= */

func main() {
	var cfg Config

	flag.StringVar(&cfg.BrokerHost, "broker-host", getenv("BROKER_HOST", "127.0.0.1"), "Broker IP/hostname (single mode)")
	flag.IntVar(&cfg.BrokerPort, "broker-port", getenvInt("BROKER_PORT", 1883), "Broker port (single mode)")
	flag.StringVar(&cfg.BrokersJSON, "brokers-json", getenv("BROKERS_JSON", ""), "Path to brokers.json for cluster mode (overrides broker-host/broker-port)")

	flag.StringVar(&cfg.TestName, "test-name", getenv("TEST_NAME", "baseline"), "Test name (used for output filenames)")
	flag.StringVar(&cfg.OutDir, "out-dir", getenv("OUT_DIR", "./results"), "Output directory for JSON/CSV")

	// Ramp (optional)
	flag.IntVar(&cfg.RampIntervalSec, "ramp-interval", 10, "Ramp interval in seconds")
	flag.IntVar(&cfg.RampStepSubs, "ramp-step-subs", 200, "Add this many subscribers per ramp interval")
	flag.IntVar(&cfg.RampStepPubs, "ramp-step-pubs", 5, "Add this many publishers per ramp interval")
	flag.IntVar(&cfg.MaxSubs, "max-subs", 5000, "Max subscribers to attempt")
	flag.IntVar(&cfg.MaxPubs, "max-pubs", 200, "Max publishers to attempt")

	// Fixed mode (paper-friendly)
	flag.IntVar(&cfg.FixedSubs, "fixed-subs", 0, "If >0, start exactly this many subscribers (disables ramp)")
	flag.IntVar(&cfg.FixedPubs, "fixed-pubs", 0, "If >0, start exactly this many publishers (disables ramp)")
	flag.IntVar(&cfg.WarmupSec, "warmup-sec", 10, "Warmup seconds before SLA evaluation (fixed mode)")

	flag.Float64Var(&cfg.PubRatePerSec, "pub-rate", 1.0, "Publish rate per second per publisher (QoS0)")
	flag.StringVar(&cfg.TopicPrefix, "topic-prefix", "bench/topic", "Topic prefix")
	flag.IntVar(&cfg.TopicCount, "topic-count", 10, "Number of topics (subscribers subscribe to all topics)")

	flag.IntVar(&cfg.PayloadKB, "payload-kb", 10, "Payload size in KB (default 10). Example: --payload-kb=100")

	// Stop conditions
	flag.Float64Var(&cfg.MaxConnFailPct, "max-conn-fail-pct", 2.0, "Stop if connection fail %% exceeds this within a ramp window")
	flag.Float64Var(&cfg.MaxDiscPerMin, "max-disc-per-min", 50.0, "Stop if disconnects per minute exceed this")
	flag.IntVar(&cfg.MinHoldSteps, "min-hold-steps", 3, "Do not stop before this many ramp steps")

	// SLA-style (applied after warmup in fixed mode; also applied during ramp steps once MinHoldSteps reached)
	flag.Float64Var(&cfg.MinConnectedSubPct, "min-connected-sub-pct", 99.0, "Stop if connected_subs < this %% of target for consecutive windows")
	flag.Float64Var(&cfg.MinDeliveryRatioPct, "min-delivery-ratio-pct", 99.0, "Stop if delivery_ratio < this %% for consecutive windows")
	flag.Float64Var(&cfg.MaxP95LatencyMs, "max-p95-latency-ms", 500.0, "Stop if p95 latency exceeds this for consecutive windows")
	flag.IntVar(&cfg.ConsecutiveSLABreaches, "consecutive-sla-breaches", 3, "Stop after this many consecutive SLA-breaching windows")

	flag.IntVar(&cfg.MaxDurationSec, "max-duration-sec", 600, "Maximum test duration in seconds (hard stop). Default 600 = 10 minutes")

	flag.Parse()

	if cfg.PayloadKB <= 0 {
		log.Fatalf("payload-kb must be > 0")
	}
	if cfg.TopicCount <= 0 {
		log.Fatalf("topic-count must be > 0")
	}
	if cfg.MaxDurationSec <= 0 {
		log.Fatalf("max-duration-sec must be > 0")
	}
	if cfg.FixedSubs < 0 || cfg.FixedPubs < 0 {
		log.Fatalf("fixed-subs/fixed-pubs must be >= 0")
	}
	if cfg.ConsecutiveSLABreaches <= 0 {
		cfg.ConsecutiveSLABreaches = 1
	}

	if err := os.MkdirAll(cfg.OutDir, 0o755); err != nil {
		log.Fatalf("failed to create out dir: %v", err)
	}

	// Load cluster broker pool (optional)
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

	trafficCSVPath := filepath.Join(cfg.OutDir, cfg.TestName+".traffic.csv")
	latCSVPath := filepath.Join(cfg.OutDir, cfg.TestName+".latency.csv")
	jsonPath := filepath.Join(cfg.OutDir, cfg.TestName+".json")

	trafficW, err := NewTrafficCSVWriter(trafficCSVPath)
	if err != nil {
		log.Fatalf("failed to open traffic csv: %v", err)
	}
	defer func() { _ = trafficW.Close() }()

	latW, err := NewLatencyCSVWriter(latCSVPath)
	if err != nil {
		log.Fatalf("failed to open latency csv: %v", err)
	}
	defer func() { _ = latW.Close() }()

	if len(brokerPool) > 0 {
		log.Printf("loadtest: starting test=%s mode=cluster brokers=%d payload=%dKB qos=0 max_duration=%ds",
			cfg.TestName, len(brokerPool), cfg.PayloadKB, cfg.MaxDurationSec)
	} else {
		log.Printf("loadtest: starting test=%s mode=single broker=%s:%d payload=%dKB qos=0 max_duration=%ds",
			cfg.TestName, cfg.BrokerHost, cfg.BrokerPort, cfg.PayloadKB, cfg.MaxDurationSec)
	}

	baseCtx := context.Background()
	ctx, cancel := context.WithTimeout(baseCtx, time.Duration(cfg.MaxDurationSec)*time.Second)
	defer cancel()

	var m Metrics
	m.LatHist = NewLatencyHist()

	// Channel for latency samples written to latency CSV
	sampleCh := make(chan LatencySample, 200000)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-sampleCh:
				// Note: if ctx.Done happens, we may drop remaining buffered samples—acceptable for load tests.
				latW.WriteSample(s)
				m.LatHist.Add(s.Latency)
				m.LatStats.Add(s.Latency)
			}
		}
	}()

	var subsMu sync.Mutex
	var pubsMu sync.Mutex
	subs := make([]*ClientRunner, 0, max(1, cfg.MaxSubs))
	pubs := make([]*ClientRunner, 0, max(1, cfg.MaxPubs))

	start := time.Now()
	peakSubs := int64(0)
	peakPubs := int64(0)

	var stopReason string

	step := 0
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// If fixed mode, we don't ramp; otherwise we ramp
	useFixed := cfg.FixedSubs > 0 || cfg.FixedPubs > 0
	var ramp *time.Ticker
	if !useFixed {
		if cfg.RampIntervalSec <= 0 {
			log.Fatalf("ramp-interval must be > 0")
		}
		ramp = time.NewTicker(time.Duration(cfg.RampIntervalSec) * time.Second)
		defer ramp.Stop()
	}

	targetSubs := 0
	targetPubs := 0
	if useFixed {
		targetSubs = cfg.FixedSubs
		targetPubs = cfg.FixedPubs
	} else {
		targetSubs = 0
		targetPubs = 0
	}

	// Previous cumulative values for per-second deltas
	var prevSubConnOK, prevSubConnFail, prevSubDisc, prevSubSubErr uint64
	var prevPubConnOK, prevPubConnFail, prevPubDisc uint64
	var prevPubsSent, prevPubsErr, prevMsgsRecv uint64
	var prevExpected, prevLost, prevDup, prevReorder uint64

	// SLA evaluation window (we'll reuse ramp interval for windowing; in fixed mode it's 1s windows)
	slaBreaches := 0
	warmupUntil := start.Add(time.Duration(cfg.WarmupSec) * time.Second)

	// Start fixed clients immediately (if fixed)
	if useFixed {
		if targetSubs > 0 {
			subsMu.Lock()
			for len(subs) < targetSubs {
				id := len(subs)
				r := NewSubscriber(cfg, brokerPool, id, &m, sampleCh)
				subs = append(subs, r)
				go r.Run(ctx)
			}
			subsMu.Unlock()
		}
		if targetPubs > 0 {
			pubsMu.Lock()
			for len(pubs) < targetPubs {
				id := len(pubs)
				r := NewPublisher(cfg, brokerPool, id, &m)
				pubs = append(pubs, r)
				go r.Run(ctx)
			}
			pubsMu.Unlock()
		}
	}

	// Per-second: traffic CSV + expected delivery accounting
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				step++

				connSubs := atomic.LoadInt64(&m.SubsConnected)
				connPubs := atomic.LoadInt64(&m.PubsConnected)

				if connSubs > peakSubs {
					peakSubs = connSubs
				}
				if connPubs > peakPubs {
					peakPubs = connPubs
				}

				// Load cumulative
				curSubConnOK := atomic.LoadUint64(&m.SubConnOK)
				curSubConnFail := atomic.LoadUint64(&m.SubConnFail)
				curSubDisc := atomic.LoadUint64(&m.SubDisc)
				curSubSubErr := atomic.LoadUint64(&m.SubSubErr)

				curPubConnOK := atomic.LoadUint64(&m.PubConnOK)
				curPubConnFail := atomic.LoadUint64(&m.PubConnFail)
				curPubDisc := atomic.LoadUint64(&m.PubDisc)

				curPubsSent := atomic.LoadUint64(&m.PubsSent)
				curPubsErr := atomic.LoadUint64(&m.PubsErr)
				curMsgsRecv := atomic.LoadUint64(&m.MsgsRecv)

				// Deltas
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

				// Expected deliveries this second:
				// Since subscribers subscribe to ALL topics, each published message should reach all connected subs.
				expectedDelta := uint64(0)
				if connSubs > 0 && dPubsSent > 0 {
					expectedDelta = dPubsSent * uint64(connSubs)
				}
				atomic.AddUint64(&m.ExpectedDeliveries, expectedDelta)

				curExpected := atomic.LoadUint64(&m.ExpectedDeliveries)
				curLost := atomic.LoadUint64(&m.LostDeliveries)
				curDup := atomic.LoadUint64(&m.DupDeliveries)
				curReorder := atomic.LoadUint64(&m.ReorderDeliveries)

				dExpected := curExpected - prevExpected
				dLost := curLost - prevLost
				dDup := curDup - prevDup
				dReorder := curReorder - prevReorder

				// Delivery ratio for this second
				deliveryRatio := 0.0
				if dExpected > 0 {
					deliveryRatio = float64(dMsgsRecv) / float64(dExpected)
				}

				// Latency quantiles (global so far; acceptable for quick plotting; for strict windowed quantiles you'd need window histograms)
				p50 := m.LatHist.Quantile(0.50)
				p95 := m.LatHist.Quantile(0.95)
				p99 := m.LatHist.Quantile(0.99)

				trafficW.WriteRow(TrafficRow{
					TS:   time.Now().UTC().Format(time.RFC3339),
					Step: step,

					TargetSubs: targetSubs,
					TargetPubs: targetPubs,

					ConnectedSubs: connSubs,
					ConnectedPubs: connPubs,

					SubConnOKDelta:   dSubConnOK,
					SubConnFailDelta: dSubConnFail,
					SubDiscDelta:     dSubDisc,
					SubSubErrDelta:   dSubSubErr,

					PubConnOKDelta:   dPubConnOK,
					PubConnFailDelta: dPubConnFail,
					PubDiscDelta:     dPubDisc,

					PubsSentDelta: dPubsSent,
					PubsErrDelta:  dPubsErr,
					MsgsRecvDelta: dMsgsRecv,

					ExpectedDelta: dExpected,
					LostDelta:     dLost,
					DupDelta:      dDup,
					ReorderDelta:  dReorder,

					DeliveryRatio: deliveryRatio,

					P50: p50,
					P95: p95,
					P99: p99,
				})

				// Update prev
				prevSubConnOK, prevSubConnFail, prevSubDisc, prevSubSubErr = curSubConnOK, curSubConnFail, curSubDisc, curSubSubErr
				prevPubConnOK, prevPubConnFail, prevPubDisc = curPubConnOK, curPubConnFail, curPubDisc
				prevPubsSent, prevPubsErr, prevMsgsRecv = curPubsSent, curPubsErr, curMsgsRecv
				prevExpected, prevLost, prevDup, prevReorder = curExpected, curLost, curDup, curReorder

				// SLA evaluation (after warmup)
				now := time.Now()
				if now.After(warmupUntil) && (useFixed || step >= cfg.MinHoldSteps) && stopReason == "" {
					// connected subs %
					connectedPct := 100.0
					if targetSubs > 0 {
						connectedPct = (float64(connSubs) / float64(targetSubs)) * 100.0
					}

					// delivery ratio %
					deliveryPct := 100.0 * deliveryRatio

					// p95 SLA
					breach := false
					if targetSubs > 0 && connectedPct < cfg.MinConnectedSubPct {
						breach = true
					}
					if dExpected > 0 && deliveryPct < cfg.MinDeliveryRatioPct {
						breach = true
					}
					if cfg.MaxP95LatencyMs > 0 && p95 > cfg.MaxP95LatencyMs {
						breach = true
					}

					if breach {
						slaBreaches++
					} else {
						slaBreaches = 0
					}

					if slaBreaches >= cfg.ConsecutiveSLABreaches {
						stopReason = fmt.Sprintf("stop: SLA breach for %d consecutive seconds (connected_pct>=%.2f, delivery_pct>=%.2f, p95<=%.2f)",
							cfg.ConsecutiveSLABreaches, cfg.MinConnectedSubPct, cfg.MinDeliveryRatioPct, cfg.MaxP95LatencyMs)
						cancel()
						return
					}
				}
			}
		}
	}()

	// Ramp loop (only if not fixed)
	if !useFixed {
		var windowConnOK, windowConnFail, windowDisc uint64
		windowStart := time.Now()

		for {
			select {
			case <-ctx.Done():
				goto done
			case <-ramp.C:
				// Use publisher+subscriber totals for legacy stop checks (connection fail %, disconnects/min)
				curOK := atomic.LoadUint64(&m.SubConnOK) + atomic.LoadUint64(&m.PubConnOK)
				curFail := atomic.LoadUint64(&m.SubConnFail) + atomic.LoadUint64(&m.PubConnFail)
				curDisc := atomic.LoadUint64(&m.SubDisc) + atomic.LoadUint64(&m.PubDisc)

				dOK := curOK - windowConnOK
				dFail := curFail - windowConnFail
				dDisc := curDisc - windowDisc

				windowConnOK = curOK
				windowConnFail = curFail
				windowDisc = curDisc

				windowDur := time.Since(windowStart)
				windowStart = time.Now()

				if step >= cfg.MinHoldSteps && stopReason == "" {
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

				// Ensure subscribers
				subsMu.Lock()
				for len(subs) < targetSubs {
					id := len(subs)
					r := NewSubscriber(cfg, brokerPool, id, &m, sampleCh)
					subs = append(subs, r)
					go r.Run(ctx)
				}
				subsMu.Unlock()

				// Ensure publishers
				pubsMu.Lock()
				for len(pubs) < targetPubs {
					id := len(pubs)
					r := NewPublisher(cfg, brokerPool, id, &m)
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
	}

	// Fixed mode just waits
	<-ctx.Done()

done:
	end := time.Now()

	if stopReason == "" {
		if ctx.Err() == context.DeadlineExceeded {
			stopReason = fmt.Sprintf("stop: max_duration_sec=%d reached", cfg.MaxDurationSec)
		} else {
			stopReason = "stopped: context cancelled"
		}
	}

	// Snapshot latency stats
	_, mean, std, lmin, lmax := m.LatStats.Snapshot()

	totalExpected := atomic.LoadUint64(&m.ExpectedDeliveries)
	totalRecv := atomic.LoadUint64(&m.MsgsRecv)
	overallDelivery := 0.0
	if totalExpected > 0 {
		overallDelivery = float64(totalRecv) / float64(totalExpected)
	}

	s := Summary{
		Config:      cfg,
		StartedAt:   start,
		FinishedAt:  end,
		DurationSec: end.Sub(start).Seconds(),

		PeakConnectedSubs: int(peakSubs),
		PeakConnectedPubs: int(peakPubs),

		StopReason: stopReason,

		TotalSubConnOK:   atomic.LoadUint64(&m.SubConnOK),
		TotalSubConnFail: atomic.LoadUint64(&m.SubConnFail),
		TotalSubDisc:     atomic.LoadUint64(&m.SubDisc),
		TotalSubSubErr:   atomic.LoadUint64(&m.SubSubErr),

		TotalPubConnOK:   atomic.LoadUint64(&m.PubConnOK),
		TotalPubConnFail: atomic.LoadUint64(&m.PubConnFail),
		TotalPubDisc:     atomic.LoadUint64(&m.PubDisc),

		TotalPubsSent: atomic.LoadUint64(&m.PubsSent),
		TotalPubsErr:  atomic.LoadUint64(&m.PubsErr),
		TotalMsgsRecv: totalRecv,
		TotalExpected: totalExpected,

		TotalLost:    atomic.LoadUint64(&m.LostDeliveries),
		TotalDup:     atomic.LoadUint64(&m.DupDeliveries),
		TotalReorder: atomic.LoadUint64(&m.ReorderDeliveries),

		OverallDeliveryRatio: overallDelivery,

		LatencyP50Ms: m.LatHist.Quantile(0.50),
		LatencyP95Ms: m.LatHist.Quantile(0.95),
		LatencyP99Ms: m.LatHist.Quantile(0.99),

		LatencyMeanMs: mean,
		LatencyStdMs:  std,
		LatencyMinMs:  lmin,
		LatencyMaxMs:  lmax,
	}

	b, _ := json.MarshalIndent(s, "", "  ")
	if err := os.WriteFile(jsonPath, b, 0o644); err != nil {
		log.Printf("loadtest: failed to write json: %v", err)
	} else {
		log.Printf("loadtest: wrote %s", jsonPath)
	}
	log.Printf("loadtest: wrote %s", trafficCSVPath)
	log.Printf("loadtest: wrote %s", latCSVPath)
	log.Printf("loadtest: %s", stopReason)
}

/* =======================
   Client runners
   ======================= */

type ClientRunner struct {
	name string
	run  func(ctx context.Context)
}

func (c *ClientRunner) Run(ctx context.Context) { c.run(ctx) }

func mqttOpts(cfg Config, brokerPool []Broker, clientID string, onConn func(), onLost func(err error)) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()

	host := cfg.BrokerHost
	port := cfg.BrokerPort

	if len(brokerPool) > 0 {
		b := pickBrokerDeterministic(brokerPool, clientID)
		host = b.Host
		port = b.Port
	}

	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", host, port))
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

// Payload header format (total 20 bytes):
// [0:8]   publish_ts_unix_nano (uint64)
// [8:12]  pub_id (uint32)
// [12:20] seq (uint64)
const payloadHeaderBytes = 20

func NewSubscriber(cfg Config, brokerPool []Broker, id int, m *Metrics, sampleCh chan<- LatencySample) *ClientRunner {
	clientID := fmt.Sprintf("sub-%s-%d-%d", cfg.TestName, os.Getpid(), id)
	topics := make([]string, cfg.TopicCount)
	for i := 0; i < cfg.TopicCount; i++ {
		topics[i] = fmt.Sprintf("%s/%d", cfg.TopicPrefix, i)
	}

	return &ClientRunner{
		name: clientID,
		run: func(ctx context.Context) {
			var connected int32

			// Track per-publisher last sequence seen (per subscriber) to estimate loss/dup/reorder deliveries.
			lastSeq := make(map[uint32]uint64, 128)

			opts := mqttOpts(cfg, brokerPool, clientID,
				func() {
					if atomic.CompareAndSwapInt32(&connected, 0, 1) {
						atomic.AddInt64(&m.SubsConnected, 1)
					}
				},
				func(err error) {
					atomic.AddUint64(&m.SubDisc, 1)
					if atomic.CompareAndSwapInt32(&connected, 1, 0) {
						atomic.AddInt64(&m.SubsConnected, -1)
					}
				},
			)

			opts.SetDefaultPublishHandler(func(_ mqtt.Client, msg mqtt.Message) {
				atomic.AddUint64(&m.MsgsRecv, 1)

				payload := msg.Payload()
				if len(payload) < payloadHeaderBytes {
					return
				}
				pubNS := int64(binary.BigEndian.Uint64(payload[0:8]))
				pubID := binary.BigEndian.Uint32(payload[8:12])
				seq := binary.BigEndian.Uint64(payload[12:20])

				recvNS := time.Now().UnixNano()
				latMs := float64(recvNS-pubNS) / 1e6

				// Gap/dup/reorder estimation for delivery quality (per subscriber).
				if prev, ok := lastSeq[pubID]; ok {
					if seq == prev {
						atomic.AddUint64(&m.DupDeliveries, 1)
					} else if seq > prev+1 {
						atomic.AddUint64(&m.LostDeliveries, uint64(seq-prev-1))
					} else if seq < prev {
						atomic.AddUint64(&m.ReorderDeliveries, 1)
					}
					if seq > prev {
						lastSeq[pubID] = seq
					}
				} else {
					lastSeq[pubID] = seq
				}

				// Timestamp sample CSV
				select {
				case sampleCh <- LatencySample{
					TSUTC:   time.Now().UTC().Format(time.RFC3339Nano),
					RecvNS:  recvNS,
					PubNS:   pubNS,
					PubID:   pubID,
					Seq:     seq,
					Latency: latMs,
					Topic:   msg.Topic(),
				}:
				default:
					// drop sample if congested
				}
			})

			client := mqtt.NewClient(opts)

			if token := client.Connect(); token.Wait() && token.Error() != nil {
				atomic.AddUint64(&m.SubConnFail, 1)
				return
			}
			atomic.AddUint64(&m.SubConnOK, 1)

			for _, t := range topics {
				if token := client.Subscribe(t, 0, nil); token.Wait() && token.Error() != nil {
					atomic.AddUint64(&m.SubSubErr, 1)
					atomic.AddUint64(&m.SubDisc, 1)
					client.Disconnect(100)
					if atomic.CompareAndSwapInt32(&connected, 1, 0) {
						atomic.AddInt64(&m.SubsConnected, -1)
					}
					return
				}
			}

			<-ctx.Done()
			client.Disconnect(100)
			if atomic.CompareAndSwapInt32(&connected, 1, 0) {
				atomic.AddInt64(&m.SubsConnected, -1)
			}
		},
	}
}

func NewPublisher(cfg Config, brokerPool []Broker, id int, m *Metrics) *ClientRunner {
	clientID := fmt.Sprintf("pub-%s-%d-%d", cfg.TestName, os.Getpid(), id)

	payloadBytes := cfg.PayloadKB * 1024
	if payloadBytes < payloadHeaderBytes {
		payloadBytes = payloadHeaderBytes
	}

	topicIdx := id % max(1, cfg.TopicCount)
	topic := fmt.Sprintf("%s/%d", cfg.TopicPrefix, topicIdx)

	pubID := uint32(id)

	return &ClientRunner{
		name: clientID,
		run: func(ctx context.Context) {
			var connected int32

			opts := mqttOpts(cfg, brokerPool, clientID,
				func() {
					if atomic.CompareAndSwapInt32(&connected, 0, 1) {
						atomic.AddInt64(&m.PubsConnected, 1)
					}
				},
				func(err error) {
					atomic.AddUint64(&m.PubDisc, 1)
					if atomic.CompareAndSwapInt32(&connected, 1, 0) {
						atomic.AddInt64(&m.PubsConnected, -1)
					}
				},
			)

			client := mqtt.NewClient(opts)
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				atomic.AddUint64(&m.PubConnFail, 1)
				return
			}
			atomic.AddUint64(&m.PubConnOK, 1)

			interval := time.Duration(float64(time.Second) / math.Max(cfg.PubRatePerSec, 0.000001))
			t := time.NewTicker(interval)
			defer t.Stop()

			buf := make([]byte, payloadBytes)
			for i := payloadHeaderBytes; i < len(buf); i++ {
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
					seq++
					tsn := uint64(time.Now().UnixNano())
					binary.BigEndian.PutUint64(buf[0:8], tsn)
					binary.BigEndian.PutUint32(buf[8:12], pubID)
					binary.BigEndian.PutUint64(buf[12:20], seq)

					token := client.Publish(topic, 0, false, buf)
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

/* =======================
   Helpers
   ======================= */

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
