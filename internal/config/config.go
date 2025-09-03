package config

import (
	"flag"
	"log"
	"strconv"
	"time"

	"github.com/caarlos0/env/v6"
)

const (
	defaultEndpoint        = "localhost:8080"
	defaultReportInterval  = "10"
	defaultPollInterval    = "2"
	defaultStoreInterval   = "300"
	defaultFileStoragePath = "/tmp/metrics-store.json"
	defaultRestore         = "true"
)

type Config interface {
	LoadAgentConfig() *AgentConfig
	LoadServerConfig() *ServerConfig
}

type AgentConfig struct {
	Endpoint       string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
}

type ServerConfig struct {
	Endpoint        string        `env:"ADDRESS"`
	StoreInterval   time.Duration `env:"STORE_INTERVAL"`
	FileStoragePath string        `env:"FILE_STORAGE_PATH"`
	Restore         bool          `env:"RESTORE"`
}

func LoadAgentConfig() *AgentConfig {
	pollTimeDuration, err := strconv.Atoi(defaultPollInterval)
	if err != nil {
		log.Printf("err: while parsing default poll duration in config")
		pollTimeDuration = 2
	}
	reportTimeDuration, err := strconv.Atoi(defaultReportInterval)
	if err != nil {
		log.Printf("err: while parsing default report interval in config")
		reportTimeDuration = 10
	}

	appConfig := &AgentConfig{
		Endpoint:       defaultEndpoint,
		PollInterval:   time.Duration(pollTimeDuration) * time.Second,
		ReportInterval: time.Duration(reportTimeDuration) * time.Second,
	}

	endpoint := flag.String("a", defaultEndpoint, "endpoint address")
	reportInterval := flag.String("r", defaultReportInterval, "report interval (seconds)")
	pollInterval := flag.String("p", defaultPollInterval, "poll interval (seconds)")
	flag.Parse()

	if *endpoint != "" {
		appConfig.Endpoint = *endpoint
	}

	if *reportInterval != "" {
		parseReportInterval, err := strconv.Atoi(*reportInterval)
		if err != nil {
			log.Printf("err: converting reportInterval to int: %v", err)
		} else {
			appConfig.ReportInterval = time.Duration(parseReportInterval) * time.Second
		}
	}

	if *pollInterval != "" {
		parsePollInterval, err := strconv.Atoi(*pollInterval)
		if err != nil {
			log.Printf("err: converting pollInterval to int: %v", err)
		} else {
			appConfig.PollInterval = time.Duration(parsePollInterval) * time.Second
		}
	}

	if err := env.Parse(appConfig); err != nil {
		log.Printf("err: parsing env variables: %v", err)
	}

	log.Printf("Starting metrics server on %s", appConfig.Endpoint)
	log.Printf("Update metrics: POST http://%s/update/<type>/<name>/<value>", appConfig.Endpoint)
	log.Printf("Get metric value: GET http://%s/value/<type>/<name>", appConfig.Endpoint)
	log.Printf("View all metrics: GET http://%s/debug", appConfig.Endpoint)
	log.Printf("Poll Interval: %v", appConfig.PollInterval)
	log.Printf("Report Interval: %v", appConfig.ReportInterval)

	return appConfig
}

func LoadServerConfig() *ServerConfig {
	storeIntervalDuration, err := strconv.Atoi(defaultStoreInterval)
	if err != nil {
		log.Printf("err: while parsing default store interval in config")
		storeIntervalDuration = 300

	}

	restoreValue, err := strconv.ParseBool(defaultRestore)
	if err != nil {
		log.Printf("err: while parsing default restore value in config")
		restoreValue = true

	}

	appConfig := &ServerConfig{
		Endpoint:        defaultEndpoint,
		StoreInterval:   time.Duration(storeIntervalDuration) * time.Second,
		FileStoragePath: defaultFileStoragePath,
		Restore:         restoreValue,
	}

	endpoint := flag.String("a", defaultEndpoint, "endpoint address")
	storeInterval := flag.String("i", "", "store interval (seconds)")
	fileStoragePath := flag.String("f", "", "file storage path")
	restore := flag.String("r", "", "restore from file (true/false)")

	flag.Parse()

	if *endpoint != "" {
		appConfig.Endpoint = *endpoint
	}

	if *storeInterval != "" {
		parseStoreInterval, err := strconv.Atoi(*storeInterval)
		if err != nil {
			log.Printf("err: converting store interval to int: %v", err)
		} else {
			appConfig.StoreInterval = time.Duration(parseStoreInterval) * time.Second
		}
	}

	if *restore != "" {
		parseRestore, err := strconv.ParseBool(*restore)
		if err != nil {
			log.Printf("err: converting restore valie to bool: %v", err)
		} else {
			appConfig.Restore = parseRestore
		}
	}
	if *fileStoragePath != "" {
		appConfig.FileStoragePath = *fileStoragePath
	}

	if err := env.Parse(appConfig); err != nil {
		log.Printf("err: parsing env variables: %v", err)
	}

	log.Printf("Starting metrics server on %s", appConfig.Endpoint)
	log.Printf("Update metrics: POST http://%s/update/<type>/<name>/<value>", appConfig.Endpoint)
	log.Printf("Get metric value: GET http://%s/value/<type>/<name>", appConfig.Endpoint)
	log.Printf("View all metrics: GET http://%s/debug", appConfig.Endpoint)
	log.Printf("Storage Interval: %v", appConfig.StoreInterval)
	log.Printf("File Storage Path: %v", appConfig.FileStoragePath)
	log.Printf("Restore value: %v", appConfig.Restore)
	return appConfig
}
