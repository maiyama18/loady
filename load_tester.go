package main

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// LoadTester ...
type LoadTester struct {
	Connections int
	RPS         float64
	Duration    time.Duration
	Clients     []*LoadClient
	Logger      *log.Logger
}

// NewLoadTester ...
func NewLoadTester(url string, connections int, rps float64, duration time.Duration, logger *log.Logger) *LoadTester {
	clients := []*LoadClient{}
	for i := 0; i < connections; i++ {
		clients = append(clients, NewLoadClient(i, url, logger))
	}

	return &LoadTester{
		Connections: connections,
		RPS:         rps,
		Duration:    duration,
		Clients:     clients,
		Logger:      logger,
	}
}

// Run ...
func (lt *LoadTester) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), lt.Duration)
	defer cancel()

	lt.Logger.Printf("request start with rps %f and connections %d\n", lt.RPS, lt.Connections)

	interval := time.Duration(1000*(float64(lt.Connections)/lt.RPS)) * time.Millisecond
	delayUnit := time.Duration(1000/lt.RPS) * time.Millisecond

	resultCh := make(chan *RequestResult)
	for i := 0; i < lt.Connections; i++ {
		delay := time.Duration(i) * delayUnit
		go lt.Clients[i].Run(ctx, resultCh, delay, interval)
	}

	results := []*RequestResult{}
loop:
	for {
		select {
		case result := <-resultCh:
			lt.Logger.Println("got")
			results = append(results, result)
		case <-ctx.Done():
			break loop
		}
	}

	stats := processResults(results)

	lt.Logger.Printf("RequestCount: %v\n", stats.RequestCount)
	lt.Logger.Printf("ErrorCount: %v\n", stats.ErrorCount)
	lt.Logger.Printf("MeanResponseTime: %v\n", stats.MeanResponseTime)
}

// LoadClient ...
type LoadClient struct {
	ID         int
	URL        string
	HTTPClient *http.Client
	Logger     *log.Logger
}

// NewLoadClient ...
func NewLoadClient(id int, url string, logger *log.Logger) *LoadClient {
	tr := &http.Transport{
		MaxIdleConns:    1,
		IdleConnTimeout: 100 * time.Second,
	}
	return &LoadClient{
		ID:         id,
		URL:        url,
		HTTPClient: &http.Client{Transport: tr},
		Logger:     logger,
	}
}

// Run ...
func (lc *LoadClient) Run(ctx context.Context, resultCh chan *RequestResult, delay, interval time.Duration) {
	<-time.After(delay)

	result, _ := lc.Request()
	resultCh <- result

	lc.Logger.Printf("status: %d (client%d)\n", result.StatusCode, lc.ID)

	for {
		select {
		case <-time.After(interval):
			result, err := lc.Request()
			if err != nil {
				break
			}
			resultCh <- result

			lc.Logger.Printf("status: %d (client%d)\n", result.StatusCode, lc.ID)
		case <-ctx.Done():
			return
		}
	}
}

// Request ...
func (lc *LoadClient) Request() (*RequestResult, error) {
	req, err := http.NewRequest("GET", lc.URL, nil)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	resp, err := lc.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)

	responseTime := time.Now().Sub(start)

	return NewRequestResult(resp.StatusCode, responseTime), nil
}

// RequestResult ...
type RequestResult struct {
	StatusCode   int
	ResponseTime time.Duration
}

// NewRequestResult ...
func NewRequestResult(statusCode int, responseTime time.Duration) *RequestResult {
	return &RequestResult{
		StatusCode:   statusCode,
		ResponseTime: responseTime,
	}
}

// RequestStats ...
type RequestStats struct {
	RequestCount     int
	ErrorCount       int
	MeanResponseTime time.Duration
}

func processResults(results []*RequestResult) *RequestStats {
	errCount := 0
	resTimeSumMicroSec := 0.0
	for _, result := range results {
		if result.StatusCode != http.StatusOK {
			errCount++
		}
		resTimeSumMicroSec += float64(result.ResponseTime / time.Microsecond)
	}

	return &RequestStats{
		RequestCount:     len(results),
		ErrorCount:       errCount,
		MeanResponseTime: time.Duration((resTimeSumMicroSec / float64(len(results)))) * time.Microsecond,
	}
}
