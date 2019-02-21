package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

// LoadClient ...
type LoadClient struct {
	ID         int
	URL        string
	HTTPClient *http.Client
	Logger     *log.Logger
}

// NewLoadClient ...
func NewLoadClient(id int, url string, logger *log.Logger) *LoadClient {
	return &LoadClient{
		ID:         id,
		URL:        url,
		HTTPClient: http.DefaultClient,
		Logger:     logger,
	}
}

// Run ...
func (lc *LoadClient) Run(ctx context.Context, interval time.Duration) {
	for {
		select {
		case <-time.After(interval):
			req, err := http.NewRequest("GET", lc.URL, nil)
			if err != nil {
				return
			}

			resp, err := lc.HTTPClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			lc.Logger.Printf("status: %d (client%d)\n", resp.StatusCode, lc.ID)
		case <-ctx.Done():
			return
		}
	}
}

// LoadTester ...
type LoadTester struct {
	Connections int
	RPS         int
	Duration    time.Duration
	Clients     []*LoadClient
	Logger      *log.Logger
}

// NewLoadTester ...
func NewLoadTester(url string, connections, rps int, duration time.Duration, logger *log.Logger) *LoadTester {
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

	lt.Logger.Printf("request start with rps %d and connections %d\n", lt.RPS, lt.Connections)

	// TODO: requestのタイミングをgoroutine間でばらけさせる
	interval := time.Duration(1000*(float64(lt.Connections)/float64(lt.RPS))) * time.Millisecond
	for i := 0; i < lt.Connections; i++ {
		go lt.Clients[i].Run(ctx, interval)
	}

	<-ctx.Done()

	lt.Logger.Println("load tester done")
}
