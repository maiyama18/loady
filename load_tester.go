package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

type LoadClient struct {
	ID         int
	URL        string
	HTTPClient *http.Client
	Logger     *log.Logger
}

func NewLoadClient(id int, url string, logger *log.Logger) *LoadClient {
	return &LoadClient{
		ID:         id,
		URL:        url,
		HTTPClient: http.DefaultClient,
		Logger:     logger,
	}
}

func (lc *LoadClient) Run(ctx context.Context, interval time.Duration) {
	for {
		select {
		case <-time.After(interval):
			lc.Logger.Printf("clientID: %d\n", lc.ID)
		case <-ctx.Done():
			return
		}
	}
}

type LoadTester struct {
	Connections int
	RPS         int
	Duration    time.Duration
	Clients     []*LoadClient
	Logger      *log.Logger
}

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
