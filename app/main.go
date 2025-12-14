package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"hash/fnv"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"
)

/* =======================
   Config & Data Models
   ======================= */

type Config struct {
	BrokerHost string `json:"broker_host"`
	BrokerPort int    `json:"broker_port"`
	BrokersJSON string `json:"brokers_json,omitempty"`

	TestName string `json:"test_name"`
	OutDir   string `json:"out_dir"`

	RampIntervalSec int `json:"ramp_interval_sec"`
	RampStepSubs    int `json:"ramp_step_subscribers"`
	RampStepPubs    int `json:"ramp_step_publishers"`
	MaxSubs         int `json:"max_subscribers"`
	MaxPubs         int `json:"max_publishers"`

	PubRatePerSec float64 `json:"pub_rate_per_sec"`
	TopicPrefix   string  `json:"topic_prefix"`
	TopicCount    int     `json:"topic_count"`

	PayloadKB int `json:"payload_kb"`

	MaxConnFailPct float64 `json:"max_conn_fail_pct"`
	MaxDiscPerMin  float64 `json:"max_disconnects_per_min"`
	MinHoldSteps   int     `json:"min_hold_steps"`
}

type Broker struct {
	ID   string `json:"id"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

type BrokersFile struct {
	Brokers []Broker `json:"brokers"`
}

type Summary struct {
	Config Config `json:"config"`

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

/* =======================
   Metrics
   ======================= */

type Metrics struct {
	ConnOK   uint64
	ConnFail uint64
	Disc     uint64

	SubsConnected int64
	PubsConnected int64

	PubsSent uint64
	PubsErr  uint64
	MsgsRecv uint64

	LatHist LatencyHist
}

type LatencyHist struct {
	edges  []float64
	counts []uint64
	total  uint64
}

func NewLatencyHist() LatencyHist {
	edges := []float64{1,2,5,10,20,50,100,200,500,1000,2000,5000}
	return LatencyHist{edges: edges, counts: make([]uint64, len(edges)+1)}
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
			return h.edges[len(h.edges)-1] * 2
		}
	}
	return h.edges[len(h.edges)-1] * 2
}

/* =======================
   CSV Writer
   ======================= */

type CSVWriter struct{ f *os.File }

func NewCSVWriter(path string) (*CSVWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	header := "ts,step,target_subs,target_pubs,connected_subs,connected_pubs,conn_ok,conn_fail,disconnects,pubs_sent,pubs_err,msgs_recv,recv_rate,p50,p95,p99\n"
	_, _ = f.WriteString(header)
	return &CSVWriter{f: f}, nil
}

func (w *CSVWriter) Write(line string) {
	_, _ = w.f.WriteString(line)
}

func (w *CSVWriter) Close() error { return w.f.Close() }

/* =======================
   Main
   ======================= */

func main() {
	var cfg Config

	flag.StringVar(&cfg.BrokerHost, "broker-host", getenv("BROKER_HOST", "127.0.0.1"), "")
	flag.IntVar(&cfg.BrokerPort, "broker-port", getenvInt("BROKER_PORT", 1883), "")
	flag.StringVar(&cfg.BrokersJSON, "brokers-json", getenv("BROKERS_JSON", ""), "")

	flag.StringVar(&cfg.TestName, "test-name", getenv("TEST_NAME", "baseline"), "")
	flag.StringVar(&cfg.OutDir, "out-dir", getenv("OUT_DIR", "./results"), "")

	flag.IntVar(&cfg.PayloadKB, "payload-kb", 10, "")

	flag.IntVar(&cfg.RampIntervalSec, "ramp-interval", 10, "")
	flag.IntVar(&cfg.RampStepSubs, "ramp-step-subs", 200, "")
	flag.IntVar(&cfg.RampStepPubs, "ramp-step-pubs", 5, "")

	flag.IntVar(&cfg.MaxSubs, "max-subs", 5000, "")
	flag.IntVar(&cfg.MaxPubs, "max-pubs", 200, "")

	flag.Float64Var(&cfg.PubRatePerSec, "pub-rate", 1.0, "")
	flag.StringVar(&cfg.TopicPrefix, "topic-prefix", "bench/topic", "")
	flag.IntVar(&cfg.TopicCount, "topic-count", 10, "")

	flag.Float64Var(&cfg.MaxConnFailPct, "max-conn-fail-pct", 2.0, "")
	flag.Float64Var(&cfg.MaxDiscPerMin, "max-disc-per-min", 50.0, "")
	flag.IntVar(&cfg.MinHoldSteps, "min-hold-steps", 3, "")
	flag.Parse()

	var brokerPool []Broker
	if cfg.BrokersJSON != "" {
		bf := mustLoadBrokers(cfg.BrokersJSON)
		brokerPool = bf.Brokers
		log.Printf("cluster mode enabled with %d brokers", len(brokerPool))
	}

	_ = os.MkdirAll(cfg.OutDir, 0755)
	csvPath := filepath.Join(cfg.OutDir, cfg.TestName+".csv")
	jsonPath := filepath.Join(cfg.OutDir, cfg.TestName+".json")

	csvw, _ := NewCSVWriter(csvPath)
	defer csvw.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var m Metrics
	m.LatHist = NewLatencyHist()

	latCh := make(chan float64, 100000)
	go func() {
		for ms := range latCh {
			m.LatHist.Add(ms)
		}
	}()

	start := time.Now()
	step := 0
	targetSubs := 0
	targetPubs := 0

	ticker := time.NewTicker(time.Second)
	ramp := time.NewTicker(time.Duration(cfg.RampIntervalSec) * time.Second)

	var subs []*ClientRunner
	var pubs []*ClientRunner

	go func() {
		var prev uint64
		for range ticker.C {
			now := time.Now().UTC().Format(time.RFC3339)
			recv := atomic.LoadUint64(&m.MsgsRecv)
			rate := recv - prev
			prev = recv
			line := fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%.2f,%.2f,%.2f\n",
				now, step, targetSubs, targetPubs,
				atomic.LoadInt64(&m.SubsConnected),
				atomic.LoadInt64(&m.PubsConnected),
				m.ConnOK, m.ConnFail, m.Disc,
				m.PubsSent, m.PubsErr, recv, rate,
				m.LatHist.Quantile(0.50),
				m.LatHist.Quantile(0.95),
				m.LatHist.Quantile(0.99))
			csvw.Write(line)
		}
	}()

	for range ramp.C {
		step++
		if targetSubs < cfg.MaxSubs {
			targetSubs += cfg.RampStepSubs
		}
		if targetPubs < cfg.MaxPubs {
			targetPubs += cfg.RampStepPubs
		}

		for len(subs) < targetSubs {
			id := len(subs)
			s := NewSubscriber(cfg, brokerPool, id, &m, latCh)
			subs = append(subs, s)
			go s.Run(ctx)
		}

		for len(pubs) < targetPubs {
			id := len(pubs)
			p := NewPublisher(cfg, brokerPool, id, &m)
			pubs = append(pubs, p)
			go p.Run(ctx)
		}
	}

	end := time.Now()
	summary := Summary{
		Config: cfg,
		StartedAt: start,
		FinishedAt: end,
		DurationSec: end.Sub(start).Seconds(),
		PeakConnectedSubs: int(m.SubsConnected),
		PeakConnectedPubs: int(m.PubsConnected),
		TotalConnOK: m.ConnOK,
		TotalConnFail: m.ConnFail,
		TotalDisc: m.Disc,
		TotalPubsSent: m.PubsSent,
		TotalPubsErr: m.PubsErr,
		TotalMsgsRecv: m.MsgsRecv,
		LatencyP50Ms: m.LatHist.Quantile(0.50),
		LatencyP95Ms: m.LatHist.Quantile(0.95),
		LatencyP99Ms: m.LatHist.Quantile(0.99),
	}

	b, _ := json.MarshalIndent(summary, "", "  ")
	_ = os.WriteFile(jsonPath, b, 0644)
}

/* =======================
   Clients
   ======================= */

type ClientRunner struct {
	run func(ctx context.Context)
}

func (c *ClientRunner) Run(ctx context.Context) { c.run(ctx) }

func mqttOpts(cfg Config, pool []Broker, id string) *mqtt.ClientOptions {
	host := cfg.BrokerHost
	port := cfg.BrokerPort
	if len(pool) > 0 {
		b := pickBroker(pool, id)
		host = b.Host
		port = b.Port
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", host, port))
	opts.SetClientID(id)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(false)
	opts.SetKeepAlive(30 * time.Second)
	return opts
}

func NewSubscriber(cfg Config, pool []Broker, id int, m *Metrics, latCh chan<- float64) *ClientRunner {
	clientID := fmt.Sprintf("sub-%d", id)
	return &ClientRunner{run: func(ctx context.Context) {
		opts := mqttOpts(cfg, pool, clientID)
		opts.SetDefaultPublishHandler(func(_ mqtt.Client, msg mqtt.Message) {
			atomic.AddUint64(&m.MsgsRecv, 1)
			ts := int64(binary.BigEndian.Uint64(msg.Payload()[:8]))
			latCh <- float64(time.Now().UnixNano()-ts) / 1e6
		})
		c := mqtt.NewClient(opts)
		if token := c.Connect(); token.Wait() && token.Error() == nil {
			atomic.AddUint64(&m.ConnOK, 1)
			atomic.AddInt64(&m.SubsConnected, 1)
		}
		for i := 0; i < cfg.TopicCount; i++ {
			c.Subscribe(fmt.Sprintf("%s/%d", cfg.TopicPrefix, i), 0, nil)
		}
		<-ctx.Done()
		c.Disconnect(100)
	}}
}

func NewPublisher(cfg Config, pool []Broker, id int, m *Metrics) *ClientRunner {
	clientID := fmt.Sprintf("pub-%d", id)
	payload := make([]byte, cfg.PayloadKB*1024)
	return &ClientRunner{run: func(ctx context.Context) {
		opts := mqttOpts(cfg, pool, clientID)
		c := mqtt.NewClient(opts)
		if token := c.Connect(); token.Wait() && token.Error() == nil {
			atomic.AddUint64(&m.ConnOK, 1)
			atomic.AddInt64(&m.PubsConnected, 1)
		}
		topic := fmt.Sprintf("%s/%d", cfg.TopicPrefix, id%cfg.TopicCount)
		t := time.NewTicker(time.Duration(float64(time.Second) / cfg.PubRatePerSec))
		for {
			select {
			case <-ctx.Done():
				c.Disconnect(100)
				return
			case <-t.C:
				binary.BigEndian.PutUint64(payload[:8], uint64(time.Now().UnixNano()))
				c.Publish(topic, 0, false, payload)
				atomic.AddUint64(&m.PubsSent, 1)
			}
		}
	}}
}

/* =======================
   Helpers
   ======================= */

func mustLoadBrokers(path string) *BrokersFile {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	var bf BrokersFile
	if err := json.Unmarshal(b, &bf); err != nil {
		log.Fatal(err)
	}
	return &bf
}

func pickBroker(b []Broker, key string) Broker {
	h := fnv.New32a()
	h.Write([]byte(key))
	return b[int(h.Sum32())%len(b)]
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func getenvInt(k string, d int) int {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	n, _ := strconv.Atoi(v)
	return n
}
