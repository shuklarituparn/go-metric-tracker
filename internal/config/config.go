package config

import (
	"flag"
	"log"
	"strconv"
	"time"
)


type Config interface{
	load() *AppConfig
}

type AppConfig struct{
	Endpoint string
	ReportInterval time.Duration
	PollInterval time.Duration
}


func Load() *AppConfig{
	endpoint := flag.String("a", "localhost:8080", "endpoint address")
	reportInterval := flag.String("r", "10", "report interval (seconds)")
	pollInterval := flag.String("p", "2", "poll interval (seconds)")
	flag.Parse()

	correctPollDuration, err := strconv.Atoi(*pollInterval)
	if err != nil {
		log.Fatalf("err: during parsing flag pollInterval: %v", err)
	}

	correctReportDuration, err := strconv.Atoi(*reportInterval)
	if err != nil {
		log.Fatalf("err: during parsing flag reportInterval: %v", err)
	}


	log.Printf("Starting metrics server on %s", *endpoint)
	log.Printf("Update metrics: POST http://%s/update/<type>/<name>/<value>", *endpoint)
	log.Printf("Get metric value: GET http://%s/value/<type>/<name>", *endpoint)
	log.Printf("View all metrics: GET http://%s/debug", *endpoint)

	return &AppConfig{
		Endpoint:       *endpoint,
		ReportInterval: time.Duration(correctReportDuration) * time.Second,
		PollInterval:   time.Duration(correctPollDuration) * time.Second,
	}

}