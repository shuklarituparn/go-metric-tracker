package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
	"strconv"
	"time"
)

const (
	defaultEndpoint       = "localhost:8080"
	defaultReportInterval = "10"
	defaultPollInterval   = "2"
)

type Config interface {
	load() *AppConfig
}

type AppConfig struct {
	Endpoint       string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
}

func Load() *AppConfig {
	var appConfig *AppConfig
	pollTimeDuration, err := strconv.Atoi(defaultPollInterval)
	if err != nil {
		log.Printf("err: while parsing default pollduration in config")
	}
	reportTimeDuration, err := strconv.Atoi(defaultReportInterval)
	if err != nil {
		log.Printf("err: while parsing default report interval in config")
	}
	appConfig = &AppConfig{
		Endpoint:       defaultEndpoint,
		PollInterval:   time.Duration(pollTimeDuration) * time.Second,
		ReportInterval: time.Duration(reportTimeDuration) * time.Second,
	}

	endpoint := flag.String("a", defaultEndpoint, "endpoint address")
	reportInterval := flag.String("r", defaultReportInterval, "report interval (seconds)")
	pollInterval := flag.String("p", defaultPollInterval, "poll interval (seconds)")
	flag.Parse()

	parseReportnterval, err := strconv.Atoi(*reportInterval)
	if err != nil {
		log.Printf("err: converting parseTimeInterval to int")
	}
	parsePollInterval, err := strconv.Atoi(*pollInterval)
	if err != nil {
		log.Printf("err: converting parseTimeInterval to int")
	}
	if *endpoint != "" {
		appConfig.Endpoint = *endpoint
	}
	if *reportInterval != "" {
		appConfig.ReportInterval = time.Duration(parseReportnterval)
	}
	if *pollInterval != "" {
		appConfig.PollInterval = time.Duration(parsePollInterval)

	}

	if err := env.Parse(&appConfig); err != nil {
		log.Printf("err: no env variables are given")
	}

	log.Printf("Starting metrics server on %s", *endpoint)
	log.Printf("Update metrics: POST http://%s/update/<type>/<name>/<value>", *endpoint)
	log.Printf("Get metric value: GET http://%s/value/<type>/<name>", *endpoint)
	log.Printf("View all metrics: GET http://%s/debug", *endpoint)

	return appConfig

}
