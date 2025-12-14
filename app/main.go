package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Broker struct to hold broker configuration
type Broker struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

// LoadTestConfig holds all configuration for the load test
type LoadTestConfig struct {
	TestCode          string
	TotalUsers        int
	RampUpRate        int    // users per second
	QoS               byte
	MessageSizeKB     int
	PublishRatePerSec int
	TestDurationMins  int
}

// TestMetrics holds metrics for the load test
type TestMetrics struct {
	TotalMessages     int64
	SuccessfulPubs    int64
	FailedPubs        int64
	TotalConnections  int64
	FailedConnections int64
	StartTime         time.Time
	EndTime           time.Time
	RequestsPerMinute map[int]int64 // minute -> successful requests count
	mu                sync.RWMutex
}

// TestResult holds the complete test result for JSON export
type TestResult struct {
	TestCode              string            `json:"testCode"`
	BrokerName            string            `json:"brokerName"`
	BrokerHost            string            `json:"brokerHost"`
	BrokerPort            int               `json:"brokerPort"`
	TestConfiguration     TestConfiguration `json:"testConfiguration"`
	TestMetrics          TestMetricsResult `json:"testMetrics"`
	RequestsPerMinute    map[string]int64  `json:"requestsPerMinute"` // using string key for JSON compatibility
	Timestamp            time.Time         `json:"timestamp"`
}

// TestConfiguration holds the test configuration for JSON export
type TestConfiguration struct {
	TotalUsers        int    `json:"totalUsers"`
	RampUpRate        int    `json:"rampUpRate"`
	QoS               byte   `json:"qos"`
	MessageSizeKB     int    `json:"messageSizeKB"`
	PublishRatePerSec int    `json:"publishRatePerSec"`
	TestDurationMins  int    `json:"testDurationMins"`
}

// TestMetricsResult holds the metrics for JSON export
type TestMetricsResult struct {
	TotalMessages           int64   `json:"totalMessages"`
	SuccessfulPubs          int64   `json:"successfulPubs"`
	FailedPubs              int64   `json:"failedPubs"`
	TotalConnections        int64   `json:"totalConnections"`
	FailedConnections       int64   `json:"failedConnections"`
	TestDurationSeconds     float64 `json:"testDurationSeconds"`
	PublicationSuccessRate  float64 `json:"publicationSuccessRate"`
	ConnectionSuccessRate   float64 `json:"connectionSuccessRate"`
	ThroughputMsgsPerSec    float64 `json:"throughputMsgsPerSec"`
	AverageRequestsPerMin   float64 `json:"averageRequestsPerMin"`
}

func (m *TestMetrics) IncrementSuccessfulPubs() {
	atomic.AddInt64(&m.SuccessfulPubs, 1)
	
	// Track requests per minute
	m.mu.Lock()
	minute := int(time.Since(m.StartTime).Minutes())
	if m.RequestsPerMinute == nil {
		m.RequestsPerMinute = make(map[int]int64)
	}
	m.RequestsPerMinute[minute]++
	m.mu.Unlock()
}

func (m *TestMetrics) IncrementFailedPubs() {
	atomic.AddInt64(&m.FailedPubs, 1)
}

func (m *TestMetrics) IncrementTotalConnections() {
	atomic.AddInt64(&m.TotalConnections, 1)
}

func (m *TestMetrics) IncrementFailedConnections() {
	atomic.AddInt64(&m.FailedConnections, 1)
}

func (m *TestMetrics) PrintStats() {
	duration := m.EndTime.Sub(m.StartTime)
	successRate := float64(m.SuccessfulPubs) / float64(m.TotalMessages) * 100
	connectionSuccessRate := float64(m.TotalConnections-m.FailedConnections) / float64(m.TotalConnections) * 100
	throughput := float64(m.SuccessfulPubs) / duration.Seconds()

	fmt.Printf("\n\n\n=== LOAD TEST RESULTS ===\n")
	fmt.Printf("Test Duration: %v\n", duration)
	fmt.Printf("Total Messages Sent: %d\n", m.TotalMessages)
	fmt.Printf("Successful Publications: %d\n", m.SuccessfulPubs)
	fmt.Printf("Failed Publications: %d\n", m.FailedPubs)
	fmt.Printf("Publication Success Rate: %.2f%%\n", successRate)
	fmt.Printf("Total Connections: %d\n", m.TotalConnections)
	fmt.Printf("Failed Connections: %d\n", m.FailedConnections)
	fmt.Printf("Connection Success Rate: %.2f%%\n", connectionSuccessRate)
	fmt.Printf("Throughput: %.2f messages/sec\n", throughput)
	
	// Print requests per minute breakdown
	fmt.Printf("\n=== SUCCESSFUL REQUESTS PER MINUTE ===\n")
	m.mu.RLock()
	totalMinutes := int(duration.Minutes()) + 1
	for minute := 0; minute < totalMinutes; minute++ {
		count := m.RequestsPerMinute[minute]
		fmt.Printf("Minute %d: %d successful requests\n", minute+1, count)
	}
	
	// Calculate average requests per minute
	if totalMinutes > 0 {
		avgPerMinute := float64(m.SuccessfulPubs) / float64(totalMinutes)
		fmt.Printf("Average: %.2f successful requests/minute\n", avgPerMinute)
	}
	m.mu.RUnlock()
	
	fmt.Printf("========================\n")
}


// ExportToJSON exports the test metrics to a JSON file
func (m *TestMetrics) ExportToJSON(testCode string, broker Broker, config *LoadTestConfig) error {
	// Ensure results directory exists
	testCasesDir := "results"
	if err := os.MkdirAll(testCasesDir, 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %v", err)
	}

	// Create filename using testCode
	filename := filepath.Join(testCasesDir, fmt.Sprintf("%s.json", testCode))

	// Calculate metrics
	duration := m.EndTime.Sub(m.StartTime)
	var successRate, connectionSuccessRate, throughput, avgPerMinute float64

	if m.TotalMessages > 0 {
		successRate = float64(m.SuccessfulPubs) / float64(m.TotalMessages) * 100
	}
	if m.TotalConnections > 0 {
		connectionSuccessRate = float64(m.TotalConnections-m.FailedConnections) / float64(m.TotalConnections) * 100
	}
	if duration.Seconds() > 0 {
		throughput = float64(m.SuccessfulPubs) / duration.Seconds()
	}

	totalMinutes := int(duration.Minutes()) + 1
	if totalMinutes > 0 {
		avgPerMinute = float64(m.SuccessfulPubs) / float64(totalMinutes)
	}

	// Convert RequestsPerMinute to string keys for JSON compatibility
	m.mu.RLock()
	requestsPerMinute := make(map[string]int64)
	for minute, count := range m.RequestsPerMinute {
		requestsPerMinute[fmt.Sprintf("%d", minute+1)] = count
	}
	m.mu.RUnlock()

	// Create test result structure
	testResult := TestResult{
		TestCode:   testCode,
		BrokerName: broker.Name,
		BrokerHost: broker.Host,
		BrokerPort: broker.Port,
		TestConfiguration: TestConfiguration{
			TotalUsers:        config.TotalUsers,
			RampUpRate:        config.RampUpRate,
			QoS:               config.QoS,
			MessageSizeKB:     config.MessageSizeKB,
			PublishRatePerSec: config.PublishRatePerSec,
			TestDurationMins:  config.TestDurationMins,
		},
		TestMetrics: TestMetricsResult{
			TotalMessages:           m.TotalMessages,
			SuccessfulPubs:          m.SuccessfulPubs,
			FailedPubs:              m.FailedPubs,
			TotalConnections:        m.TotalConnections,
			FailedConnections:       m.FailedConnections,
			TestDurationSeconds:     duration.Seconds(),
			PublicationSuccessRate:  successRate,
			ConnectionSuccessRate:   connectionSuccessRate,
			ThroughputMsgsPerSec:    throughput,
			AverageRequestsPerMin:   avgPerMinute,
		},
		RequestsPerMinute: requestsPerMinute,
		Timestamp:         time.Now(),
	}

	// Marshal to JSON with pretty formatting
	jsonData, err := json.MarshalIndent(testResult, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test result to JSON: %v", err)
	}

	// Write JSON file
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON file: %v", err)
	}

	fmt.Printf("Test results exported to JSON: %s\n", filename)
	return nil
}

// ExportAggregatedResultsToJSON exports aggregated test results across all brokers
func ExportAggregatedResultsToJSON(testCode string, brokers []Broker, config *LoadTestConfig, totalSuccessfulPubs, totalFailedPubs, totalConnections, totalFailedConnections int64, testStartTime, testEndTime time.Time) error {
	// Ensure results directory exists
	testCasesDir := "results"
	if err := os.MkdirAll(testCasesDir, 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %v", err)
	}

	// Create filename for aggregated results
	filename := filepath.Join(testCasesDir, fmt.Sprintf("%s-AGGREGATED.json", testCode))

	// Calculate aggregated metrics
	duration := testEndTime.Sub(testStartTime)
	totalMessages := totalSuccessfulPubs + totalFailedPubs
	var successRate, connectionSuccessRate, throughput float64

	if totalMessages > 0 {
		successRate = float64(totalSuccessfulPubs) / float64(totalMessages) * 100
	}
	if totalConnections > 0 {
		connectionSuccessRate = float64(totalConnections-totalFailedConnections) / float64(totalConnections) * 100
	}
	if duration.Seconds() > 0 {
		throughput = float64(totalSuccessfulPubs) / duration.Seconds()
	}

	// Create broker summary
	brokerSummary := make([]map[string]interface{}, len(brokers))
	usersPerBroker := config.TotalUsers / len(brokers)
	remainingUsers := config.TotalUsers % len(brokers)
	
	for i, broker := range brokers {
		users := usersPerBroker
		if i < remainingUsers {
			users++
		}
		brokerSummary[i] = map[string]interface{}{
			"name": broker.Name,
			"host": broker.Host,
			"port": broker.Port,
			"assignedUsers": users,
		}
	}

	// Create aggregated result structure
	aggregatedResult := map[string]interface{}{
		"testCode": testCode,
		"testType": "DISTRIBUTED_LOAD_TEST",
		"brokerCount": len(brokers),
		"brokerSummary": brokerSummary,
		"testConfiguration": TestConfiguration{
			TotalUsers:        config.TotalUsers,
			RampUpRate:        config.RampUpRate,
			QoS:               config.QoS,
			MessageSizeKB:     config.MessageSizeKB,
			PublishRatePerSec: config.PublishRatePerSec,
			TestDurationMins:  config.TestDurationMins,
		},
		"aggregatedMetrics": map[string]interface{}{
			"totalMessages":           totalMessages,
			"successfulPubs":          totalSuccessfulPubs,
			"failedPubs":              totalFailedPubs,
			"totalConnections":        totalConnections,
			"failedConnections":       totalFailedConnections,
			"testDurationSeconds":     duration.Seconds(),
			"publicationSuccessRate":  successRate,
			"connectionSuccessRate":   connectionSuccessRate,
			"throughputMsgsPerSec":    throughput,
		},
		"timestamp": time.Now(),
	}

	// Marshal to JSON with pretty formatting
	jsonData, err := json.MarshalIndent(aggregatedResult, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal aggregated result to JSON: %v", err)
	}

	// Write JSON file
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write aggregated JSON file: %v", err)
	}

	fmt.Printf("Aggregated test results exported to JSON: %s\n", filename)
	return nil
}

func main() {
	// Read broker configuration from brokers.json
	brokers, err := readBrokers("brokers.json")
	if err != nil {
		fmt.Printf("Error reading brokers file: %v\n", err)
		os.Exit(1)
	}

	// Parse environment variables
	config, err := parseConfig()
	if err != nil {
		fmt.Printf("Error parsing configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting MQTT Distributed Load Test\n")
	fmt.Printf("Test Code: %s\n", config.TestCode)
	fmt.Printf("Found %d brokers to test\n", len(brokers))
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Total Users: %d (will be distributed across all brokers)\n", config.TotalUsers)
	fmt.Printf("  Ramp Up Rate: %d users/sec per broker\n", config.RampUpRate)
	fmt.Printf("  QoS: %d\n", config.QoS)
	fmt.Printf("  Message Size: %d KB\n", config.MessageSizeKB)
	fmt.Printf("  Publish Rate: %d msgs/sec per user\n", config.PublishRatePerSec)
	fmt.Printf("  Test Duration: %d minutes\n", config.TestDurationMins)
	fmt.Println("================================")

	// Distribute users equally across all brokers and run simultaneously
	runDistributedLoadTest(brokers, config)
}

// parseConfig reads and validates environment variables
func parseConfig() (*LoadTestConfig, error) {
	config := &LoadTestConfig{}

	// Required environment variables
	testCode := os.Getenv("testCode")
	if testCode == "" {
		return nil, fmt.Errorf("testCode environment variable is required")
	}
	config.TestCode = testCode

	totalUsers, err := strconv.Atoi(os.Getenv("totalUsers"))
	if err != nil || totalUsers <= 0 {
		return nil, fmt.Errorf("totalUsers must be a positive integer")
	}
	config.TotalUsers = totalUsers

	rampUpRate, err := strconv.Atoi(os.Getenv("rampUpRate"))
	if err != nil || rampUpRate <= 0 {
		return nil, fmt.Errorf("rampUpRate must be a positive integer")
	}
	config.RampUpRate = rampUpRate

	qos, err := strconv.Atoi(os.Getenv("qos"))
	if err != nil || qos < 0 || qos > 2 {
		return nil, fmt.Errorf("qos must be 0, 1, or 2")
	}
	config.QoS = byte(qos)

	messageSizeKB, err := strconv.Atoi(os.Getenv("messageSizeKB"))
	if err != nil || messageSizeKB <= 0 {
		return nil, fmt.Errorf("messageSizeKB must be a positive integer")
	}
	config.MessageSizeKB = messageSizeKB

	publishRatePerSec, err := strconv.Atoi(os.Getenv("publishRatePerSec"))
	if err != nil || publishRatePerSec <= 0 {
		return nil, fmt.Errorf("publishRatePerSec must be a positive integer")
	}
	config.PublishRatePerSec = publishRatePerSec

	testDurationMins, err := strconv.Atoi(os.Getenv("testDurationMins"))
	if err != nil || testDurationMins <= 0 {
		return nil, fmt.Errorf("testDurationMins must be a positive integer")
	}
	config.TestDurationMins = testDurationMins

	return config, nil
}

// readBrokers reads the broker configuration from a JSON file
func readBrokers(filename string) ([]Broker, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var brokers []Broker
	err = json.Unmarshal(file, &brokers)
	if err != nil {
		return nil, err
	}

	return brokers, nil
}

// generatePayload creates a payload of specified size in KB
func generatePayload(sizeKB int) string {
	sizeBytes := sizeKB * 1024
	payload := strings.Repeat("A", sizeBytes)
	return payload
}

// runDistributedLoadTest distributes users equally across all brokers and runs them simultaneously
func runDistributedLoadTest(brokers []Broker, config *LoadTestConfig) {
	if len(brokers) == 0 {
		fmt.Printf("No brokers available for testing\n")
		return
	}

	testStartTime := time.Now()

	// Calculate users per broker
	usersPerBroker := config.TotalUsers / len(brokers)
	remainingUsers := config.TotalUsers % len(brokers)
	
	fmt.Printf("\nDistributing %d users across %d brokers:\n", config.TotalUsers, len(brokers))
	
	// Create a slice to hold user distribution
	userDistribution := make([]int, len(brokers))
	for i := range userDistribution {
		userDistribution[i] = usersPerBroker
		// Distribute remaining users among the first few brokers
		if i < remainingUsers {
			userDistribution[i]++
		}
		fmt.Printf("  %s: %d users\n", brokers[i].Name, userDistribution[i])
	}

	// Use WaitGroup to run all broker tests concurrently
	var wg sync.WaitGroup
	
	// Channel to collect results from all brokers
	resultsChan := make(chan struct {
		broker Broker
		metrics *TestMetrics
		config *LoadTestConfig
	}, len(brokers))

	// Start load test for each broker concurrently
	for i, broker := range brokers {
		if userDistribution[i] == 0 {
			continue // Skip brokers with no users assigned
		}
		
		wg.Add(1)
		go func(broker Broker, userCount int, brokerIndex int) {
			defer wg.Done()
			
			fmt.Printf("\nStarting test for broker %s with %d users\n", broker.Name, userCount)
			
			// Create a modified config for this broker with the specific user count
			brokerConfig := *config
			brokerConfig.TotalUsers = userCount
			
			// Run the load test for this broker
			metrics := runLoadTestForBroker(broker, &brokerConfig)
			
			// Send result to channel
			resultsChan <- struct {
				broker Broker
				metrics *TestMetrics
				config *LoadTestConfig
			}{broker, metrics, &brokerConfig}
			
		}(broker, userDistribution[i], i)
	}

	// Wait for all broker tests to complete
	wg.Wait()
	close(resultsChan)
	
	testEndTime := time.Now()

	// Collect and display results for all brokers
	fmt.Printf("\n\n=== DISTRIBUTED LOAD TEST SUMMARY ===\n")
	totalSuccessfulPubs := int64(0)
	totalFailedPubs := int64(0)
	totalConnections := int64(0)
	totalFailedConnections := int64(0)

	for result := range resultsChan {
		fmt.Printf("\nBroker: %s (%d users)\n", result.broker.Name, result.config.TotalUsers)
		result.metrics.PrintStats()
		
		// Export individual broker results
		if err := result.metrics.ExportToJSON(
			fmt.Sprintf("%s-%s", config.TestCode, result.broker.Name), 
			result.broker, 
			result.config); err != nil {
			fmt.Printf("Warning: Failed to export JSON for broker %s: %v\n", result.broker.Name, err)
		}
		
		// Aggregate totals
		totalSuccessfulPubs += result.metrics.SuccessfulPubs
		totalFailedPubs += result.metrics.FailedPubs
		totalConnections += result.metrics.TotalConnections
		totalFailedConnections += result.metrics.FailedConnections
	}

	// Print aggregated results
	fmt.Printf("\n=== OVERALL AGGREGATED RESULTS ===\n")
	totalMessages := totalSuccessfulPubs + totalFailedPubs
	testDuration := testEndTime.Sub(testStartTime)
	fmt.Printf("Total Test Duration: %v\n", testDuration)
	if totalMessages > 0 {
		overallSuccessRate := float64(totalSuccessfulPubs) / float64(totalMessages) * 100
		overallThroughput := float64(totalSuccessfulPubs) / testDuration.Seconds()
		fmt.Printf("Total Messages Sent: %d\n", totalMessages)
		fmt.Printf("Total Successful Publications: %d\n", totalSuccessfulPubs)
		fmt.Printf("Total Failed Publications: %d\n", totalFailedPubs)
		fmt.Printf("Overall Publication Success Rate: %.2f%%\n", overallSuccessRate)
		fmt.Printf("Overall Throughput: %.2f messages/sec\n", overallThroughput)
	}
	if totalConnections > 0 {
		overallConnectionSuccessRate := float64(totalConnections-totalFailedConnections) / float64(totalConnections) * 100
		fmt.Printf("Total Connections: %d\n", totalConnections)
		fmt.Printf("Total Failed Connections: %d\n", totalFailedConnections)
		fmt.Printf("Overall Connection Success Rate: %.2f%%\n", overallConnectionSuccessRate)
	}
	fmt.Printf("======================================\n")
	
	// Export aggregated results
	if err := ExportAggregatedResultsToJSON(config.TestCode, brokers, config, 
		totalSuccessfulPubs, totalFailedPubs, totalConnections, totalFailedConnections,
		testStartTime, testEndTime); err != nil {
		fmt.Printf("Warning: Failed to export aggregated JSON: %v\n", err)
	}
}

// runLoadTestForBroker executes the load test for a single broker and returns metrics
func runLoadTestForBroker(broker Broker, config *LoadTestConfig) *TestMetrics {
	metrics := &TestMetrics{
		StartTime:         time.Now(),
		RequestsPerMinute: make(map[int]int64),
	}

	// Generate payload of specified size
	payload := generatePayload(config.MessageSizeKB)
	topic := fmt.Sprintf("loadtest/%s/%s", config.TestCode, broker.Name)

	// Calculate test parameters
	testDuration := time.Duration(config.TestDurationMins) * time.Minute
	rampUpDuration := time.Duration(config.TotalUsers/config.RampUpRate) * time.Second

	fmt.Printf("Ramp-up phase: %v\n", rampUpDuration)
	fmt.Printf("Test phase: %v\n", testDuration)
	fmt.Printf("Topic: %s\n", topic)
	fmt.Printf("Payload size: %d bytes\n", len(payload))

	var wg sync.WaitGroup
	userStartInterval := time.Duration(1000/config.RampUpRate) * time.Millisecond

	// Start progress monitoring goroutine
	stopMonitoring := make(chan bool)
	go monitorProgress(metrics, stopMonitoring)

	// Start users with proper ramp-up using a ticker
	ticker := time.NewTicker(userStartInterval)
	defer ticker.Stop()
	
	userCount := 0
	
	// Ramp-up phase: start users at the specified rate
	go func() {
		for userCount < config.TotalUsers {
			select {
			case <-ticker.C:
				if userCount < config.TotalUsers {
					wg.Add(1)
					go func(userID int) {
						defer wg.Done()
						runUser(broker, config, metrics, userID, topic, payload, testDuration)
					}(userCount)
					userCount++
					
					if userCount%10 == 0 || userCount == config.TotalUsers {
						fmt.Printf("Started %d/%d users\n", userCount, config.TotalUsers)
					}
				}
			}
		}
	}()

	// Wait for ramp-up to complete
	rampUpTime := time.Duration(config.TotalUsers/config.RampUpRate) * time.Second
	time.Sleep(rampUpTime + time.Second) // Add buffer to ensure all users are started
	
	fmt.Printf("Ramp-up completed. All %d users started.\n", config.TotalUsers)

	// Wait for all users to finish
	wg.Wait()
	stopMonitoring <- true
	metrics.EndTime = time.Now()
	metrics.TotalMessages = atomic.LoadInt64(&metrics.SuccessfulPubs) + atomic.LoadInt64(&metrics.FailedPubs)

	// Return metrics for aggregation (printing and export handled in distributed function)
	return metrics
}

// monitorProgress displays real-time test progress
func monitorProgress(metrics *TestMetrics, stop chan bool) {
	ticker := time.NewTicker(30 * time.Second) // Update every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			elapsed := time.Since(metrics.StartTime)
			successful := atomic.LoadInt64(&metrics.SuccessfulPubs)
			failed := atomic.LoadInt64(&metrics.FailedPubs)
			total := successful + failed
			
			if total > 0 {
				successRate := float64(successful) / float64(total) * 100
				throughput := float64(successful) / elapsed.Seconds()
				
				fmt.Printf("\n[PROGRESS] Elapsed: %v | Total: %d | Success: %d (%.1f%%) | Throughput: %.1f msgs/sec\n",
					elapsed.Round(time.Second), total, successful, successRate, throughput)
			}
		}
	}
}

// runUser simulates a single user performing MQTT operations
func runUser(broker Broker, config *LoadTestConfig, metrics *TestMetrics, userID int, topic, payload string, testDuration time.Duration) {
	clientID := fmt.Sprintf("loadtest-%s-%s-user-%d", config.TestCode, broker.Name, userID)

	// Create MQTT client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker.Host, broker.Port))
	opts.SetClientID(clientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("User %d connection lost: %v", userID, err)
	})

	// Connect to broker
	metrics.IncrementTotalConnections()
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("User %d failed to connect: %v", userID, token.Error())
		metrics.IncrementFailedConnections()
		return
	}
	defer client.Disconnect(250)

	// Calculate publish interval
	publishInterval := time.Duration(1000/config.PublishRatePerSec) * time.Millisecond
	ticker := time.NewTicker(publishInterval)
	defer ticker.Stop()

	endTime := time.Now().Add(testDuration)

	// Publishing loop
	for time.Now().Before(endTime) {
		select {
		case <-ticker.C:
			// Add some randomness to avoid thundering herd
			jitter := time.Duration(rand.Intn(100)) * time.Millisecond
			time.Sleep(jitter)

			messagePayload := fmt.Sprintf("%s-user-%d-time-%d", payload[:min(100, len(payload))], userID, time.Now().UnixNano())
			
			token := client.Publish(topic, config.QoS, false, messagePayload)
			if token.Wait() && token.Error() != nil {
				metrics.IncrementFailedPubs()
			} else {
				metrics.IncrementSuccessfulPubs()
			}
		default:
			// Small sleep to prevent busy waiting
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
