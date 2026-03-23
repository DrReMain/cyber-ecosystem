package zap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/encoding"
)

type LokiWriter struct {
	url       string
	labels    map[string]string
	batchWait time.Duration
	batchSize int

	buffer   []logEntry
	bufferMu sync.Mutex
	client   *http.Client
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

type logEntry struct {
	Timestamp string `json:"ts"`
	Line      string `json:"line"`
}

type lokiPushRequest struct {
	Streams []lokiStream `json:"streams"`
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

func NewLokiWriter(cfg *LokiConfig) *LokiWriter {
	batchWait := time.Duration(cfg.BatchWait) * time.Millisecond
	if batchWait == 0 {
		batchWait = time.Second
	}
	batchSize := cfg.BatchSize
	if batchSize == 0 {
		batchSize = 1024 * 1024 // 1MB default
	}

	w := &LokiWriter{
		url:       cfg.URL,
		labels:    cfg.Labels,
		batchWait: batchWait,
		batchSize: batchSize,
		buffer:    make([]logEntry, 0, 100),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopCh: make(chan struct{}),
	}

	// Start background flusher
	w.wg.Add(1)
	go w.run()

	return w
}

func (w *LokiWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	entry := logEntry{
		Timestamp: fmt.Sprintf("%d", time.Now().UnixNano()),
		Line:      string(p),
	}

	w.bufferMu.Lock()
	w.buffer = append(w.buffer, entry)
	shouldFlush := len(w.buffer) >= 100 // Flush every 100 entries
	w.bufferMu.Unlock()

	if shouldFlush {
		go w.flush()
	}

	return len(p), nil
}

func (w *LokiWriter) Close() error {
	close(w.stopCh)
	w.wg.Wait()
	w.flush()
	return nil
}

func (w *LokiWriter) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.batchWait)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.flush()
		case <-w.stopCh:
			return
		}
	}
}

func (w *LokiWriter) flush() error {
	w.bufferMu.Lock()
	if len(w.buffer) == 0 {
		w.bufferMu.Unlock()
		return nil
	}

	// Take all entries
	entries := w.buffer
	w.buffer = make([]logEntry, 0, 100)
	w.bufferMu.Unlock()

	// Build values array
	values := make([][]string, len(entries))
	for i, entry := range entries {
		values[i] = []string{entry.Timestamp, entry.Line}
	}

	// Build push request
	req := lokiPushRequest{
		Streams: []lokiStream{
			{
				Stream: w.labels,
				Values: values,
			},
		},
	}

	codec := encoding.GetCodec("json")
	body, err := codec.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal loki request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create loki request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send loki request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("loki returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
