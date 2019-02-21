package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"time"
)

const usage = `
usage:
udb-loadtest -url <URL> -connections <CONNECTIONS> -rps <RPS> -duration <DURATION>
`

func main() {
	options, err := parseOptions()
	if err != nil {
		os.Stdout.WriteString(usage)
		os.Exit(1)
	}

	loadTester := NewLoadTester(
		options.URL,
		options.Connections,
		options.RPS,
		time.Duration(options.Duration)*time.Second,
		log.New(os.Stdout, "[INFO]", log.LstdFlags),
	)

	loadTester.Run()
}

// Options ...
type Options struct {
	URL         string
	Connections int
	RPS         int
	Duration    int
}

func parseOptions() (*Options, error) {
	url := flag.String("url", "", "url to access")
	connections := flag.Int("connections", 0, "number of connections")
	rps := flag.Int("rps", 0, "request per second")
	duration := flag.Int("duration", 0, "duration to stress load")

	flag.Parse()

	if *url == "" || *connections == 0 || *rps == 0 || *duration == 0 {
		return nil, errors.New("some options are not given")
	}

	return &Options{
		URL:         *url,
		Connections: *connections,
		RPS:         *rps,
		Duration:    *duration,
	}, nil
}
