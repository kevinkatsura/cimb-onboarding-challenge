package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// LokiSyncer buffers logs and periodically POSTs them to Loki's HTTP API.
type LokiSyncer struct {
	url        string
	client     *http.Client
	labels     map[string]string
	batch      [][]string
	batchMutex sync.Mutex
}

func NewLokiSyncer(url string, labels map[string]string) *LokiSyncer {
	s := &LokiSyncer{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Second},
		labels: labels,
		batch:  make([][]string, 0),
	}
	go s.flushRoutine()
	return s
}

func (s *LokiSyncer) Write(p []byte) (n int, err error) {
	now := fmt.Sprintf("%d", time.Now().UnixNano())
	logLine := string(bytes.TrimSpace(p))

	s.batchMutex.Lock()
	s.batch = append(s.batch, []string{now, logLine})
	s.batchMutex.Unlock()

	return len(p), nil
}

func (s *LokiSyncer) flushRoutine() {
	ticker := time.NewTicker(2 * time.Second)
	for range ticker.C {
		s.flush()
	}
}

func (s *LokiSyncer) flush() {
	s.batchMutex.Lock()
	if len(s.batch) == 0 {
		s.batchMutex.Unlock()
		return
	}
	currentBatch := s.batch
	s.batch = make([][]string, 0)
	s.batchMutex.Unlock()

	payload := map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"stream": s.labels,
				"values": currentBatch,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", s.url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err == nil {
		resp.Body.Close()
	}
}

func (s *LokiSyncer) Sync() error {
	s.flush()
	return nil
}
