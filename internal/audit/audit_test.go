package audit

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAuditEvent_Marshal(t *testing.T) {
	event := AuditEvent{
		Timestamp: 1234567890,
		Metrics:   []string{"Alloc", "Frees"},
		IPAddress: "192.168.0.42",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal audit event: %v", err)
	}

	var decoded AuditEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal audit event: %v", err)
	}

	if decoded.Timestamp != event.Timestamp {
		t.Errorf("Expected timestamp %d, got %d", event.Timestamp, decoded.Timestamp)
	}
	if len(decoded.Metrics) != len(event.Metrics) {
		t.Errorf("Expected %d metrics, got %d", len(event.Metrics), len(decoded.Metrics))
	}
	if decoded.IPAddress != event.IPAddress {
		t.Errorf("Expected IP %s, got %s", event.IPAddress, decoded.IPAddress)
	}
}

func TestAuditManager_AddObserver(t *testing.T) {
	manager := NewAuditManager()

	if manager.IsEnabled() {
		t.Error("Expected manager to be disabled initially")
	}

	tmpFile, err := ioutil.TempFile("", "audit-test-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	auditor, err := NewFileAuditor(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create file auditor: %v", err)
	}

	manager.AddObserver(auditor)

	if !manager.IsEnabled() {
		t.Error("Expected manager to be enabled after adding observer")
	}
}

func TestFileAuditor_Notify(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "audit-test-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	auditor, err := NewFileAuditor(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create file auditor: %v", err)
	}

	event := AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   []string{"metric1", "metric2"},
		IPAddress: "127.0.0.1",
	}

	if err := auditor.Notify(event); err != nil {
		t.Fatalf("Failed to notify file auditor: %v", err)
	}

	data, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read audit file: %v", err)
	}

	var decoded AuditEvent
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line in audit file, got %d", len(lines))
	}

	if err := json.Unmarshal([]byte(lines[0]), &decoded); err != nil {
		t.Fatalf("Failed to unmarshal audit event from file: %v", err)
	}

	if decoded.IPAddress != event.IPAddress {
		t.Errorf("Expected IP %s, got %s", event.IPAddress, decoded.IPAddress)
	}
	if len(decoded.Metrics) != len(event.Metrics) {
		t.Errorf("Expected %d metrics, got %d", len(event.Metrics), len(decoded.Metrics))
	}
}

func TestFileAuditor_MultipleNotifications(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "audit-test-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	auditor, err := NewFileAuditor(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create file auditor: %v", err)
	}

	for i := 0; i < 3; i++ {
		event := AuditEvent{
			Timestamp: time.Now().Unix(),
			Metrics:   []string{"metric1"},
			IPAddress: "127.0.0.1",
		}
		if err := auditor.Notify(event); err != nil {
			t.Fatalf("Failed to notify file auditor: %v", err)
		}
	}

	data, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read audit file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines in audit file, got %d", len(lines))
	}
}

func TestFileAuditor_InvalidPath(t *testing.T) {
	_, err := NewFileAuditor("")
	if err == nil {
		t.Error("Expected error for empty file path")
	}

	_, err = NewFileAuditor("/invalid/path/that/does/not/exist/audit.log")
	if err == nil {
		t.Error("Expected error for invalid file path")
	}
}

func TestHTTPAuditor_Notify(t *testing.T) {
	var receivedEvent AuditEvent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		if err := json.Unmarshal(body, &receivedEvent); err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	auditor, err := NewHTTPAuditor(server.URL)
	if err != nil {
		t.Fatalf("Failed to create HTTP auditor: %v", err)
	}

	event := AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   []string{"metric1", "metric2"},
		IPAddress: "192.168.1.1",
	}

	if err := auditor.Notify(event); err != nil {
		t.Fatalf("Failed to notify HTTP auditor: %v", err)
	}

	if receivedEvent.IPAddress != event.IPAddress {
		t.Errorf("Expected IP %s, got %s", event.IPAddress, receivedEvent.IPAddress)
	}
	if len(receivedEvent.Metrics) != len(event.Metrics) {
		t.Errorf("Expected %d metrics, got %d", len(event.Metrics), len(receivedEvent.Metrics))
	}
}

func TestHTTPAuditor_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	auditor, err := NewHTTPAuditor(server.URL)
	if err != nil {
		t.Fatalf("Failed to create HTTP auditor: %v", err)
	}

	event := AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   []string{"metric1"},
		IPAddress: "127.0.0.1",
	}

	if err := auditor.Notify(event); err == nil {
		t.Error("Expected error when server returns 500")
	}
}

func TestHTTPAuditor_InvalidURL(t *testing.T) {
	_, err := NewHTTPAuditor("")
	if err == nil {
		t.Error("Expected error for empty URL")
	}
}

func TestHTTPAuditor_NetworkError(t *testing.T) {
	auditor, err := NewHTTPAuditor("http://localhost:99999/audit")
	if err != nil {
		t.Fatalf("Failed to create HTTP auditor: %v", err)
	}

	event := AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   []string{"metric1"},
		IPAddress: "127.0.0.1",
	}

	if err := auditor.Notify(event); err == nil {
		t.Error("Expected network error for invalid URL")
	}
}

func TestAuditManager_Notify(t *testing.T) {
	manager := NewAuditManager()

	tmpFile, err := ioutil.TempFile("", "audit-test-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	auditor, err := NewFileAuditor(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create file auditor: %v", err)
	}

	manager.AddObserver(auditor)

	event := AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   []string{"metric1", "metric2"},
		IPAddress: "192.168.0.1",
	}

	manager.Notify(event)

	time.Sleep(100 * time.Millisecond)

	data, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read audit file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected audit file to contain data")
	}

	var decoded AuditEvent
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if err := json.Unmarshal([]byte(lines[0]), &decoded); err != nil {
		t.Fatalf("Failed to unmarshal audit event from file: %v", err)
	}

	if decoded.IPAddress != event.IPAddress {
		t.Errorf("Expected IP %s, got %s", event.IPAddress, decoded.IPAddress)
	}
}

func TestAuditManager_MultipleObservers(t *testing.T) {
	manager := NewAuditManager()

	tmpFile, err := ioutil.TempFile("", "audit-test-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	fileAuditor, err := NewFileAuditor(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create file auditor: %v", err)
	}
	manager.AddObserver(fileAuditor)

	var httpReceived bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpReceived = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpAuditor, err := NewHTTPAuditor(server.URL)
	if err != nil {
		t.Fatalf("Failed to create HTTP auditor: %v", err)
	}
	manager.AddObserver(httpAuditor)

	event := AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   []string{"metric1"},
		IPAddress: "127.0.0.1",
	}

	manager.Notify(event)

	time.Sleep(200 * time.Millisecond)

	data, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read audit file: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected audit file to contain data")
	}

	if !httpReceived {
		t.Error("Expected HTTP auditor to receive event")
	}
}

func TestAuditManager_ConcurrentNotify(t *testing.T) {
	manager := NewAuditManager()

	tmpFile, err := ioutil.TempFile("", "audit-test-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	auditor, err := NewFileAuditor(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create file auditor: %v", err)
	}
	manager.AddObserver(auditor)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			event := AuditEvent{
				Timestamp: time.Now().Unix(),
				Metrics:   []string{"metric1"},
				IPAddress: "127.0.0.1",
			}
			manager.Notify(event)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	time.Sleep(200 * time.Millisecond)

	data, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read audit file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 10 {
		t.Errorf("Expected 10 lines in audit file, got %d", len(lines))
	}
}
