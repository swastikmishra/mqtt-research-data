// Key improvements to add to your MQTT benchmark tool
// These changes address the issues identified in your test results

// ============================================================================
// 1. ADD PER-BROKER METRICS TRACKING
// ============================================================================

type BrokerMetrics struct {
	BrokerID         string
	ConnectedSubs    int64
	ConnectedPubs    int64
	ConnectionErrors uint64
	mu               sync.RWMutex
}

type BrokerMetricsMap struct {
	metrics map[string]*BrokerMetrics
	mu      sync.RWMutex
}

func NewBrokerMetricsMap() *BrokerMetricsMap {
	return &BrokerMetricsMap{
		metrics: make(map[string]*BrokerMetrics),
	}
}

func (b *BrokerMetricsMap) GetOrCreate(brokerID string) *BrokerMetrics {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.metrics[brokerID]; !ok {
		b.metrics[brokerID] = &BrokerMetrics{BrokerID: brokerID}
	}
	return b.metrics[brokerID]
}

// Add to WindowRow struct:
type WindowRow struct {
	// ... existing fields ...
	
	// NEW: Per-broker distribution
	BrokerDistribution map[string]int `json:"broker_distribution,omitempty"`
}

// ============================================================================
// 2. ENHANCED CLIENT TRACKING WITH BROKER AWARENESS
// ============================================================================

type TrackedSubscriber struct {
	ClientID      string
	ConnectedTo   string // broker ID
	ConnectedAt   time.Time
	SubscribedAt  time.Time
	LastMessageAt time.Time
	MessageCount  uint64
	IsConnected   bool
	mu            sync.RWMutex
}

func NewSubscriberWithTracking(cfg Config, activeBrokers []Broker, id int, 
	ctr *Counters, latCh chan<- LatSample, brokerMetrics *BrokerMetricsMap) *ClientRunner {
	
	clientID := fmt.Sprintf("sub-%s-%d-%d", cfg.TestName, os.Getpid(), id)
	
	// Determine which broker this client will connect to
	broker := pickBrokerDeterministic(activeBrokers, clientID)
	brokerID := broker.ID
	metrics := brokerMetrics.GetOrCreate(brokerID)
	
	tracker := &TrackedSubscriber{
		ClientID:    clientID,
		ConnectedTo: brokerID,
	}
	
	// ... rest of subscriber logic with tracking updates ...
	
	opts := mqttOpts(cfg, activeBrokers, clientID,
		func() {
			tracker.mu.Lock()
			tracker.IsConnected = true
			tracker.ConnectedAt = time.Now()
			tracker.mu.Unlock()
			
			metrics.mu.Lock()
			metrics.ConnectedSubs++
			metrics.mu.Unlock()
			
			atomic.AddInt64(&ctr.SubsConnected, 1)
		},
		func(err error) {
			tracker.mu.Lock()
			tracker.IsConnected = false
			tracker.mu.Unlock()
			
			metrics.mu.Lock()
			metrics.ConnectedSubs--
			metrics.mu.Unlock()
			
			// ... existing disconnect logic ...
		},
	)
	
	// Return modified ClientRunner with tracking
}

// ============================================================================
// 3. ADAPTIVE STEP SIZE FOR PRECISE CAPACITY MEASUREMENT
// ============================================================================

type AdaptiveRamp struct {
	CurrentStep      int
	BaseStep         int
	MinStep          int
	RecentBreaches   int
	RecentSuccesses  int
	AdaptiveEnabled  bool
}

func NewAdaptiveRamp(baseStep int) *AdaptiveRamp {
	return &AdaptiveRamp{
		CurrentStep:     baseStep,
		BaseStep:        baseStep,
		MinStep:         max(10, baseStep/10),
		AdaptiveEnabled: true,
	}
}

func (ar *AdaptiveRamp) NextStep(slaBreached bool) int {
	if !ar.AdaptiveEnabled {
		return ar.BaseStep
	}
	
	if slaBreached {
		ar.RecentBreaches++
		ar.RecentSuccesses = 0
		
		// Reduce step size as we approach failure
		if ar.RecentBreaches >= 2 && ar.CurrentStep > ar.MinStep {
			ar.CurrentStep = max(ar.MinStep, ar.CurrentStep/2)
			log.Printf("Adaptive ramp: reducing step to %d (approaching capacity)", ar.CurrentStep)
		}
	} else {
		ar.RecentSuccesses++
		ar.RecentBreaches = 0
		
		// Increase step size if stable
		if ar.RecentSuccesses >= 3 && ar.CurrentStep < ar.BaseStep {
			ar.CurrentStep = min(ar.BaseStep, ar.CurrentStep*2)
			log.Printf("Adaptive ramp: increasing step to %d (stable)", ar.CurrentStep)
		}
	}
	
	return ar.CurrentStep
}

// ============================================================================
// 4. DETAILED LATENCY BREAKDOWN
// ============================================================================

type LatencyBreakdown struct {
	ConnectLatencyMs   float64
	SubscribeLatencyMs float64
	MessageLatencyMs   float64
}

func (lb *LatencyBreakdown) Add(breakdown LatencyBreakdown) {
	// Add to histogram per category
}

// Modify subscriber to track connection and subscription timing
func NewSubscriberWithLatencyBreakdown(...) *ClientRunner {
	return &ClientRunner{
		run: func(ctx context.Context) {
			connectStart := time.Now()
			
			client := mqtt.NewClient(opts)
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				atomic.AddUint64(&ctr.SubConnFail, 1)
				return
			}
			
			connectLatency := time.Since(connectStart).Milliseconds()
			// Record connect latency
			
			subStart := time.Now()
			for _, t := range topics {
				if token := client.Subscribe(t, 0, nil); token.Wait() && token.Error() != nil {
					// ... error handling ...
				}
			}
			subscribeLatency := time.Since(subStart).Milliseconds()
			// Record subscribe latency
			
			// ... rest of logic ...
		},
	}
}

// ============================================================================
// 5. RESOURCE MONITORING
// ============================================================================

type ResourceStats struct {
	Timestamp      time.Time
	GoRoutines     int
	HeapAllocMB    float64
	HeapSysMB      float64
	NumGC          uint32
	GCPauseMs      float64
	OpenFDs        int64
	mu             sync.RWMutex
}

func MonitorResources(ctx context.Context, interval time.Duration) chan ResourceStats {
	statsCh := make(chan ResourceStats, 100)
	
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				close(statsCh)
				return
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				
				stats := ResourceStats{
					Timestamp:   time.Now(),
					GoRoutines:  runtime.NumGoroutine(),
					HeapAllocMB: float64(m.HeapAlloc) / 1024 / 1024,
					HeapSysMB:   float64(m.HeapSys) / 1024 / 1024,
					NumGC:       m.NumGC,
					GCPauseMs:   float64(m.PauseNs[(m.NumGC+255)%256]) / 1e6,
				}
				
				// Get open FD count (Linux)
				if fds, err := countOpenFDs(); err == nil {
					stats.OpenFDs = fds
				}
				
				select {
				case statsCh <- stats:
				default:
					// Drop if channel full
				}
			}
		}
	}()
	
	return statsCh
}

func countOpenFDs() (int64, error) {
	// Linux-specific: count files in /proc/self/fd
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return 0, err
	}
	return int64(len(entries)), nil
}

// ============================================================================
// 6. ENHANCED CSV OUTPUT WITH BROKER DISTRIBUTION
// ============================================================================

func (c *WindowCSV) WriteRowWithBrokerStats(r WindowRow, brokerMetrics *BrokerMetricsMap) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Calculate broker distribution
	distribution := make(map[string]int)
	brokerMetrics.mu.RLock()
	for id, metrics := range brokerMetrics.metrics {
		metrics.mu.RLock()
		distribution[id] = int(metrics.ConnectedSubs)
		metrics.mu.RUnlock()
	}
	brokerMetrics.mu.RUnlock()
	
	// Build distribution string for CSV
	distStr := ""
	for id, count := range distribution {
		if distStr != "" {
			distStr += ";"
		}
		distStr += fmt.Sprintf("%s:%d", id, count)
	}
	
	_ = c.w.Write([]string{
		r.TSUTC,
		// ... existing fields ...
		distStr, // NEW: broker distribution
		// ... remaining fields ...
	})
	c.w.Flush()
}

// ============================================================================
// 7. PROMETHEUS METRICS EXPORT
// ============================================================================

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type PrometheusExporter struct {
	connectedSubs     *prometheus.GaugeVec
	connectedPubs     prometheus.Gauge
	p95Latency        prometheus.Gauge
	deliveryRatio     prometheus.Gauge
	messagesReceived  prometheus.Counter
	messagesSent      prometheus.Counter
}

func NewPrometheusExporter() *PrometheusExporter {
	pe := &PrometheusExporter{
		connectedSubs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mqtt_connected_subscribers",
				Help: "Number of connected subscribers per broker",
			},
			[]string{"broker_id"},
		),
		connectedPubs: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mqtt_connected_publishers",
			Help: "Number of connected publishers",
		}),
		p95Latency: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mqtt_p95_latency_ms",
			Help: "P95 message latency in milliseconds",
		}),
		deliveryRatio: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mqtt_delivery_ratio",
			Help: "Message delivery ratio (received/expected)",
		}),
		messagesReceived: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mqtt_messages_received_total",
			Help: "Total messages received by subscribers",
		}),
		messagesSent: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mqtt_messages_sent_total",
			Help: "Total messages sent by publishers",
		}),
	}
	
	prometheus.MustRegister(pe.connectedSubs)
	prometheus.MustRegister(pe.connectedPubs)
	prometheus.MustRegister(pe.p95Latency)
	prometheus.MustRegister(pe.deliveryRatio)
	prometheus.MustRegister(pe.messagesReceived)
	prometheus.MustRegister(pe.messagesSent)
	
	return pe
}

func (pe *PrometheusExporter) Update(ctr *Counters, p95 float64, ratio float64, 
	brokerMetrics *BrokerMetricsMap) {
	
	pe.connectedPubs.Set(float64(atomic.LoadInt64(&ctr.PubsConnected)))
	pe.p95Latency.Set(p95)
	pe.deliveryRatio.Set(ratio)
	pe.messagesReceived.Add(float64(atomic.LoadUint64(&ctr.MsgsRecv)))
	pe.messagesSent.Add(float64(atomic.LoadUint64(&ctr.PubsSent)))
	
	brokerMetrics.mu.RLock()
	for id, metrics := range brokerMetrics.metrics {
		metrics.mu.RLock()
		pe.connectedSubs.WithLabelValues(id).Set(float64(metrics.ConnectedSubs))
		metrics.mu.RUnlock()
	}
	brokerMetrics.mu.RUnlock()
}

func StartPrometheusServer(port int) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Printf("Starting Prometheus metrics server on :%d/metrics", port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			log.Printf("Prometheus server error: %v", err)
		}
	}()
}

// ============================================================================
// 8. IMPROVED ERROR LOGGING AND DIAGNOSTICS
// ============================================================================

type ConnectionError struct {
	Timestamp time.Time
	ClientID  string
	BrokerID  string
	Error     string
	Type      string // "connect", "subscribe", "publish", "disconnect"
}

type ErrorLogger struct {
	errors []ConnectionError
	mu     sync.Mutex
	maxLen int
}

func NewErrorLogger(maxLen int) *ErrorLogger {
	return &ErrorLogger{
		errors: make([]ConnectionError, 0, maxLen),
		maxLen: maxLen,
	}
}

func (el *ErrorLogger) Log(clientID, brokerID, errType, errMsg string) {
	el.mu.Lock()
	defer el.mu.Unlock()
	
	err := ConnectionError{
		Timestamp: time.Now(),
		ClientID:  clientID,
		BrokerID:  brokerID,
		Error:     errMsg,
		Type:      errType,
	}
	
	el.errors = append(el.errors, err)
	if len(el.errors) > el.maxLen {
		el.errors = el.errors[len(el.errors)-el.maxLen:]
	}
	
	// Also log to console for real-time monitoring
	log.Printf("ERROR [%s] client=%s broker=%s: %s", errType, clientID, brokerID, errMsg)
}

func (el *ErrorLogger) WriteToFile(path string) error {
	el.mu.Lock()
	defer el.mu.Unlock()
	
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(el.errors)
}

// ============================================================================
// 9. GRACEFUL SHUTDOWN AND CLEANUP
// ============================================================================

func GracefulShutdown(cancel context.CancelFunc, windowCSV *WindowCSV, 
	errorLogger *ErrorLogger, summaryPath, errorPath string) {
	
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-sigCh
		log.Println("Received interrupt signal, shutting down gracefully...")
		
		cancel() // Cancel context to stop all clients
		
		time.Sleep(2 * time.Second) // Give clients time to disconnect
		
		if err := windowCSV.Close(); err != nil {
			log.Printf("Error closing CSV: %v", err)
		}
		
		if err := errorLogger.WriteToFile(errorPath); err != nil {
			log.Printf("Error writing error log: %v", err)
		}
		
		log.Println("Shutdown complete")
		os.Exit(0)
	}()
}

// ============================================================================
// 10. USAGE IN MAIN
// ============================================================================

func main() {
	// ... existing flag parsing ...
	
	// NEW: Additional flags
	var prometheusPort int
	var adaptiveRamp bool
	flag.IntVar(&prometheusPort, "prometheus-port", 9090, "Port for Prometheus metrics endpoint")
	flag.BoolVar(&adaptiveRamp, "adaptive-ramp", true, "Use adaptive step sizing")
	
	flag.Parse()
	
	// Initialize new components
	brokerMetrics := NewBrokerMetricsMap()
	errorLogger := NewErrorLogger(10000)
	promExporter := NewPrometheusExporter()
	
	if prometheusPort > 0 {
		StartPrometheusServer(prometheusPort)
	}
	
	// Start resource monitoring
	resourceStatsCh := MonitorResources(parentCtx, 10*time.Second)
	go func() {
		for stats := range resourceStatsCh {
			log.Printf("RESOURCES: goroutines=%d heap=%.1fMB fds=%d gc_pause=%.2fms",
				stats.GoRoutines, stats.HeapAllocMB, stats.OpenFDs, stats.GCPauseMs)
		}
	}()
	
	// Setup graceful shutdown
	errorLogPath := filepath.Join(cfg.OutDir, cfg.TestName+".errors.json")
	GracefulShutdown(cancel, windowCSV, errorLogger, summaryPath, errorLogPath)
	
	// Use adaptive ramp in epoch runner
	if adaptiveRamp {
		ramp := NewAdaptiveRamp(cfg.SubStep)
		// Pass ramp to epoch runner and use ramp.NextStep(slaBreached) for step calculation
	}
	
	// ... rest of main logic with new tracking ...
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}