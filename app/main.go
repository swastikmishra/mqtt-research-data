package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ============================================================================
// 1. CONFIGURATION
// ============================================================================

type Config struct {
	// Broker Connection
	BrokerHost    string
	BrokerPort    int
	BrokersJSON   string
	ClusterHotAdd bool

	// New Flag
	MaxClientsPerBroker int

	// Test Definition
	OutDir       string
	TestName     string
	InitialSubs  int
	SubStep      int
	MaxSubs      int
	Publishers   int
	PubRate      int
	PayloadBytes int
	TopicCount   int
	TopicPrefix  string
	WindowSec    int
	WarmupSec    int

	// SLAs
	SlaMinConnPct     float64
	SlaMinDeliveryPct float64
	SlaMaxP95Ms       int64
	SlaMaxDiscPerMin  int64
	SlaBreachLimit    int
}

type BrokerNode struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

var cfg Config

func init() {
	// Connection Flags
	flag.StringVar(&cfg.BrokerHost, "broker-host", "", "Single broker host")
	flag.IntVar(&cfg.BrokerPort, "broker-port", 1883, "Single broker port")
	flag.StringVar(&cfg.BrokersJSON, "brokers-json", "", "Path to brokers.json")
	flag.BoolVar(&cfg.ClusterHotAdd, "cluster-hot-add-new-clients", false, "Hot-add brokers on scaling")
	flag.IntVar(&cfg.MaxClientsPerBroker, "max-client-per-broker", 0, "Max clients per broker before forcing switch (0 = disabled)")

	// Test Params
	flag.StringVar(&cfg.OutDir, "out-dir", "./results", "Output directory")
	flag.StringVar(&cfg.TestName, "test-name", "benchmark", "Test name")
	flag.IntVar(&cfg.InitialSubs, "initial-subs", 500, "Start subscribers")
	flag.IntVar(&cfg.SubStep, "sub-step", 100, "Subscribers to add per step")
	flag.IntVar(&cfg.MaxSubs, "max-subs", 20000, "Max subscribers")
	flag.IntVar(&cfg.Publishers, "publishers", 50, "Fixed count of publishers")
	flag.IntVar(&cfg.PubRate, "pub-rate", 1, "Msg/sec per publisher")
	flag.IntVar(&cfg.PayloadBytes, "payload-bytes", 10, "Payload size (min 8 bytes)")
	flag.IntVar(&cfg.TopicCount, "topic-count", 10, "Number of topics to spread across")
	flag.StringVar(&cfg.TopicPrefix, "topic-prefix", "bench/topic", "Topic prefix")
	flag.IntVar(&cfg.WindowSec, "window-sec", 60, "Measurement window duration")
	flag.IntVar(&cfg.WarmupSec, "warmup-sec", 10, "Stabilization time before window")

	// SLAs
	flag.Float64Var(&cfg.SlaMinConnPct, "sla-min-connected-sub-pct", 99.0, "Min Connected %")
	flag.Float64Var(&cfg.SlaMinDeliveryPct, "sla-min-delivery-pct", 99.0, "Min Delivery %")
	flag.Int64Var(&cfg.SlaMaxP95Ms, "sla-max-p95-ms", 200, "Max P95 Latency")
	flag.Int64Var(&cfg.SlaMaxDiscPerMin, "sla-max-disc-per-min", 50, "Max Disconnects/min")
	flag.IntVar(&cfg.SlaBreachLimit, "sla-consecutive-breaches", 2, "Stop after N breaches")
}

// ============================================================================
// 2. METRICS & STATE
// ============================================================================

type GlobalStats struct {
	// Counters (Atomic)
	ConnectedSubs int64
	ConnectedPubs int64
	MsgsSent      uint64
	MsgsRecv      uint64
	BytesRecv     uint64 // New: Track bandwidth
	Disconnects   uint64

	// Latency Tracking
	latencies []int64
	mu        sync.Mutex
}

func (s *GlobalStats) RecordLatency(ms int64) {
	s.mu.Lock()
	s.latencies = append(s.latencies, ms)
	s.mu.Unlock()
}

func (s *GlobalStats) RecordBytes(n int) {
	atomic.AddUint64(&s.BytesRecv, uint64(n))
}

func (s *GlobalStats) ResetWindow() {
	atomic.StoreUint64(&s.MsgsSent, 0)
	atomic.StoreUint64(&s.MsgsRecv, 0)
	atomic.StoreUint64(&s.BytesRecv, 0)
	atomic.StoreUint64(&s.Disconnects, 0)
	s.mu.Lock()
	// Pre-allocate to reduce allocation overhead during window
	s.latencies = make([]int64, 0, 10000)
	s.mu.Unlock()
}

// Returns: P95, Min, Max, Avg
func (s *GlobalStats) CalculateLatencyStats() (int64, int64, int64, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := len(s.latencies)
	if count == 0 {
		return 0, 0, 0, 0
	}

	// Sort for P95
	sort.Slice(s.latencies, func(i, j int) bool { return s.latencies[i] < s.latencies[j] })

	// P95
	idx := int(float64(count) * 0.95)
	p95 := s.latencies[idx]

	// Min/Max/Avg
	min := s.latencies[0]
	max := s.latencies[count-1]
	
	var sum int64
	for _, v := range s.latencies {
		sum += v
	}
	avg := sum / int64(count)

	return p95, min, max, avg
}

var stats GlobalStats

// Broker Management
type BrokerManager struct {
	Brokers       []string
	ActiveIndex   int
	CurrentCount  int // Track clients on current broker
	HotAddEnabled bool
}

func (bm *BrokerManager) GetCurrentTarget() string {
	if bm.ActiveIndex >= len(bm.Brokers) {
		return bm.Brokers[len(bm.Brokers)-1]
	}
	return bm.Brokers[bm.ActiveIndex]
}

func (bm *BrokerManager) ActivateNextBroker() bool {
	if !bm.HotAddEnabled {
		return false
	}
	if bm.ActiveIndex < len(bm.Brokers)-1 {
		bm.ActiveIndex++
		bm.CurrentCount = 0 // Reset count for the new node
		log.Printf(">>> [CLUSTER] Hot Adding Broker #%d: %s", bm.ActiveIndex+1, bm.Brokers[bm.ActiveIndex])
		return true
	}
	return false
}

// ============================================================================
// 3. MAIN LOGIC
// ============================================================================

func main() {
	flag.Parse()

	// Validate Payload for Timestamp
	if cfg.PayloadBytes < 8 {
		log.Fatalf("Error: payload-bytes must be at least 8 to support timestamp injection (currently %d)", cfg.PayloadBytes)
	}

	// 1. Setup Environment
	if err := os.MkdirAll(cfg.OutDir, 0755); err != nil {
		log.Fatal(err)
	}

	// 2. Load Brokers
	var brokers []string
	if cfg.BrokersJSON != "" {
		file, err := os.ReadFile(cfg.BrokersJSON)
		if err != nil {
			log.Fatalf("Failed to read brokers json: %v", err)
		}
		var nodes []BrokerNode
		if err := json.Unmarshal(file, &nodes); err != nil {
			log.Fatalf("Invalid brokers json: %v", err)
		}
		for _, n := range nodes {
			brokers = append(brokers, fmt.Sprintf("tcp://%s:%d", n.Host, n.Port))
		}
	} else {
		brokers = []string{fmt.Sprintf("tcp://%s:%d", cfg.BrokerHost, cfg.BrokerPort)}
	}

	bm := &BrokerManager{
		Brokers:       brokers,
		ActiveIndex:   0,
		CurrentCount:  0,
		HotAddEnabled: cfg.ClusterHotAdd,
	}

	log.Printf("Starting Test: %s", cfg.TestName)
	log.Printf("Brokers: %v (HotAdd: %v)", brokers, cfg.ClusterHotAdd)
	if cfg.MaxClientsPerBroker > 0 {
		log.Printf("Load Strategy: Max %d clients per broker", cfg.MaxClientsPerBroker)
	}
	log.Printf("SLA: Conn>%.1f%% Del>%.1f%% P95<%dms Disc<%d",
		cfg.SlaMinConnPct, cfg.SlaMinDeliveryPct, cfg.SlaMaxP95Ms, cfg.SlaMaxDiscPerMin)

	// 3. Setup CSV - Added MinLat, MaxLat, AvgLat, MB/s
	f, _ := os.Create(filepath.Join(cfg.OutDir, cfg.TestName+"_results.csv"))
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{
		"TotalSubs", "ActiveBroker", "ConnPct", "DelPct", 
		"P95Lat_ms", "MinLat_ms", "MaxLat_ms", "AvgLat_ms", 
		"Throughput_MBps", "Disconnects", "Result",
	})
	defer w.Flush()

	// 4. Start Fixed Publishers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// WaitGroup to cleanly stop publishers
	var wg sync.WaitGroup
	startPublishers(ctx, bm, &wg)

	// 5. Subscriber Step Loop
	currentSubs := 0
	consecutiveBreaches := 0

	// Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Main Test Loop
	for targetSubs := cfg.InitialSubs; targetSubs <= cfg.MaxSubs; targetSubs += cfg.SubStep {

		// Check for user abort
		select {
		case <-sigCh:
			log.Println("Test aborted by user")
			return
		default:
		}

		// A. RAMP UP
		needed := targetSubs - currentSubs
		log.Printf("--- STEP: Scaling to %d (Adding %d) ---", targetSubs, needed)

		addedSoFar := 0
		offset := currentSubs

		for addedSoFar < needed {
			// 1. Check Capacity Limit
			if cfg.MaxClientsPerBroker > 0 && bm.CurrentCount >= cfg.MaxClientsPerBroker {
				if bm.ActivateNextBroker() {
					log.Printf("Capacity Reached. Switched to next broker.")
				} else {
					if addedSoFar == 0 {
						log.Printf("Capacity Reached but no more brokers available. Overflowing current.")
					}
				}
			}

			// 2. Calculate Chunk Size
			chunk := needed - addedSoFar
			if cfg.MaxClientsPerBroker > 0 {
				space := cfg.MaxClientsPerBroker - bm.CurrentCount
				if space > 0 && space < chunk {
					chunk = space
				}
			}

			// 3. Spawn Chunk
			spawnSubscribers(chunk, offset, bm)
			
			bm.CurrentCount += chunk
			offset += chunk
			addedSoFar += chunk
		}

		currentSubs = targetSubs

		// B. WARMUP
		log.Printf("Warmup %ds...", cfg.WarmupSec)
		time.Sleep(time.Duration(cfg.WarmupSec) * time.Second)

		// C. MEASUREMENT WINDOW
		log.Printf("Measuring Window %ds...", cfg.WindowSec)
		stats.ResetWindow()
		time.Sleep(time.Duration(cfg.WindowSec) * time.Second)

		// D. SLA CHECK & METRICS CALCULATION
		connPct := (float64(atomic.LoadInt64(&stats.ConnectedSubs)) / float64(targetSubs)) * 100

		sent := atomic.LoadUint64(&stats.MsgsSent)
		recv := atomic.LoadUint64(&stats.MsgsRecv)
		bytesRecv := atomic.LoadUint64(&stats.BytesRecv)
		disc := atomic.LoadUint64(&stats.Disconnects)

		// Throughput: Bytes / WindowSec / 1024 / 1024 = MB/s
		mbps := (float64(bytesRecv) / float64(cfg.WindowSec)) / 1024.0 / 1024.0

		// Delivery %
		delPct := 0.0
		expectedRecv := float64(sent) * (float64(targetSubs) / float64(cfg.TopicCount))
		if expectedRecv > 0 {
			delPct = (float64(recv) / expectedRecv) * 100
		} else if sent == 0 {
			delPct = 100 // No traffic
		}
		// Clamp > 100% due to window timing bleeding
		if delPct > 100.0 {
			delPct = 100.0
		}

		p95, minLat, maxLat, avgLat := stats.CalculateLatencyStats()

		passed := true
		failReason := ""

		if connPct < cfg.SlaMinConnPct {
			passed = false
			failReason += "ConnPct "
		}
		if delPct < cfg.SlaMinDeliveryPct {
			passed = false
			failReason += "DelPct "
		}
		if p95 > cfg.SlaMaxP95Ms {
			passed = false
			failReason += "Latency "
		}
		if int64(disc) > cfg.SlaMaxDiscPerMin {
			passed = false
			failReason += "Disconnects "
		}

		resultStr := "PASS"
		if !passed {
			resultStr = "FAIL"
		}

		// Console Output
		log.Printf("RESULT: %s | Conn: %.1f%% | Del: %.1f%% | MB/s: %.2f", resultStr, connPct, delPct, mbps)
		log.Printf("LATENCY: P95: %dms | Avg: %dms | Min: %dms | Max: %dms", p95, avgLat, minLat, maxLat)

		// CSV Write
		w.Write([]string{
			fmt.Sprintf("%d", targetSubs),
			bm.GetCurrentTarget(),
			fmt.Sprintf("%.2f", connPct),
			fmt.Sprintf("%.2f", delPct),
			fmt.Sprintf("%d", p95),
			fmt.Sprintf("%d", minLat),
			fmt.Sprintf("%d", maxLat),
			fmt.Sprintf("%d", avgLat),
			fmt.Sprintf("%.2f", mbps),
			fmt.Sprintf("%d", disc),
			resultStr,
		})
		w.Flush()

		// E. DECISION LOGIC
		if !passed {
			consecutiveBreaches++
			if bm.HotAddEnabled {
				log.Printf("SLA Breach (%s). Attempting to Hot-Add Broker...", failReason)
				switched := bm.ActivateNextBroker()
				if switched {
					consecutiveBreaches = 0
					log.Println("Broker Added via SLA Logic. Continuing test.")
					continue
				} else {
					log.Println("Cannot add more brokers (limit reached).")
				}
			}

			if consecutiveBreaches >= cfg.SlaBreachLimit {
				log.Printf("Stopping test: %d consecutive SLA breaches.", consecutiveBreaches)
				break
			}
		} else {
			consecutiveBreaches = 0 
		}
	}
}

// ============================================================================
// 4. CLIENT ROUTINES
// ============================================================================

func startPublishers(ctx context.Context, bm *BrokerManager, wg *sync.WaitGroup) {
	log.Printf("Starting %d Publishers...", cfg.Publishers)
	
	for i := 0; i < cfg.Publishers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Round robin connection for pubs
			opts := mqtt.NewClientOptions().AddBroker(bm.Brokers[id%len(bm.Brokers)])
			opts.SetClientID(fmt.Sprintf("pub-%d", id))
			opts.SetAutoReconnect(true)
			client := mqtt.NewClient(opts)

			if token := client.Connect(); token.Wait() && token.Error() != nil {
				log.Printf("Pub %d connect failed: %v", id, token.Error())
				return
			}
			atomic.AddInt64(&stats.ConnectedPubs, 1)

			ticker := time.NewTicker(time.Second / time.Duration(cfg.PubRate))
			
			// Pre-allocate payload
			payload := make([]byte, cfg.PayloadBytes)
			// Fill random data after the first 8 bytes (reserved for timestamp)
			if len(payload) > 8 {
				rand.Read(payload[8:])
			}

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					// Pick partitioned topic
					topicID := id % cfg.TopicCount
					topic := fmt.Sprintf("%s/%d", cfg.TopicPrefix, topicID)

					// 1. INJECT TIMESTAMP (High Precision)
					now := time.Now().UnixNano()
					binary.BigEndian.PutUint64(payload[0:8], uint64(now))

					// Copy to avoid race conditions if needed, though sequential here
					msgPayload := make([]byte, len(payload))
					copy(msgPayload, payload)

					start := time.Now()
					token := client.Publish(topic, 0, false, msgPayload)
					
					// Async send tracking
					go func(t mqtt.Token, s time.Time) {
						if t.Wait() && t.Error() == nil {
							atomic.AddUint64(&stats.MsgsSent, 1)
						}
					}(token, start)
				}
			}
		}(i)
	}
}

func spawnSubscribers(count int, offset int, bm *BrokerManager) {
	limiter := time.NewTicker(20 * time.Millisecond) // Max 50 connects/sec
	targetBroker := bm.GetCurrentTarget()

	for i := 0; i < count; i++ {
		<-limiter.C
		go func(id int) {
			opts := mqtt.NewClientOptions().AddBroker(targetBroker)
			opts.SetClientID(fmt.Sprintf("sub-%d", id))
			opts.SetAutoReconnect(true)
			
			opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
				atomic.AddInt64(&stats.ConnectedSubs, -1)
				atomic.AddUint64(&stats.Disconnects, 1)
			})
			
			opts.SetOnConnectHandler(func(c mqtt.Client) {
				atomic.AddInt64(&stats.ConnectedSubs, 1)

				topicID := id % cfg.TopicCount
				topic := fmt.Sprintf("%s/%d", cfg.TopicPrefix, topicID)

				if t := c.Subscribe(topic, 0, func(c mqtt.Client, m mqtt.Message) {
					atomic.AddUint64(&stats.MsgsRecv, 1)
					
					// Track Data Volume
					payload := m.Payload()
					stats.RecordBytes(len(payload))

					// 2. CALCULATE REAL LATENCY
					if len(payload) >= 8 {
						sentNano := int64(binary.BigEndian.Uint64(payload[0:8]))
						recvNano := time.Now().UnixNano()
						
						latencyNano := recvNano - sentNano
						latencyMs := latencyNano / 1_000_000 // Convert ns to ms
						
						if latencyMs < 0 { latencyMs = 0 } // Clock skew check
						stats.RecordLatency(latencyMs)
					}

				}); t.Wait() && t.Error() != nil {
					// Subscription failed log if needed
				}
			})

			client := mqtt.NewClient(opts)
			client.Connect()
		}(offset + i)
	}
}