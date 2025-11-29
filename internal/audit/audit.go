package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type AuditEvent struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type AuditObserver interface {
	Notify(event AuditEvent) error
}

type AuditManager struct {
	observers []AuditObserver
	mu        sync.RWMutex
}

func NewAuditManager() *AuditManager {
	return &AuditManager{
		observers: make([]AuditObserver, 0),
	}
}

func (am *AuditManager) AddObserver(observer AuditObserver) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.observers = append(am.observers, observer)
}

func (am *AuditManager) Notify(event AuditEvent) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	for _, observer := range am.observers {
		go func(obs AuditObserver) {
			if err := obs.Notify(event); err != nil {
				log.Printf("Failed to notify audit observer: %v", err)
			}
		}(observer)
	}
}

func (am *AuditManager) IsEnabled() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return len(am.observers) > 0
}

type FileAuditor struct {
	filePath string
	mu       sync.Mutex
}

func NewFileAuditor(filePath string) (*FileAuditor, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit file: %w", err)
	}
	file.Close()

	return &FileAuditor{
		filePath: filePath,
	}, nil
}

func (fa *FileAuditor) Notify(event AuditEvent) error {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	file, err := os.OpenFile(fa.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit file: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}

	return nil
}

type HTTPAuditor struct {
	url    string
	client *http.Client
}

func NewHTTPAuditor(url string) (*HTTPAuditor, error) {
	if url == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	return &HTTPAuditor{
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

func (ha *HTTPAuditor) Notify(event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	resp, err := ha.client.Post(ha.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send audit event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("audit server returned error status: %d", resp.StatusCode)
	}

	return nil
}
