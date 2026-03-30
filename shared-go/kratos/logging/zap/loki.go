package zap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// lokiEntry holds a single buffered log entry for Loki push.
type lokiEntry struct {
	pushTS string            // nanosecond timestamp for Loki values array
	line   string            // original formatted log line (ISO ts preserved)
	meta   map[string]string // structured metadata (low-cardinality only)
}

type LokiWriter struct {
	url       string
	labels    map[string]string
	batchWait time.Duration
	batchSize int

	buffer   []lokiEntry
	bufferMu sync.Mutex
	client   *http.Client
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

type lokiPushRequest struct {
	Streams []lokiStream `json:"streams"`
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]any           `json:"values"` // [[timestamp_ns, line, {structured_metadata}]]
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
		buffer:    make([]lokiEntry, 0, 100),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopCh: make(chan struct{}),
	}

	w.wg.Add(1)
	go w.run()

	return w
}

// Write implements io.Writer. p is the already-formatted log line from zap (JSON).
// The original line is sent to Loki verbatim so that the ISO timestamp (ts field)
// is preserved exactly as it appears in file and console outputs.
func (w *LokiWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	pushTS := fmt.Sprintf("%d", time.Now().UnixNano())
	line := string(bytes.TrimRight(p, "\n"))

	// Extract low-cardinality metadata from the JSON log line.
	// High-cardinality fields (span.id, trace.id) are intentionally excluded
	// to avoid creating one Loki stream per span.
	meta := map[string]string{}
	var raw map[string]interface{}
	if json.Unmarshal(p, &raw) == nil {
		for _, k := range []string{"level", "component", "service.name"} {
			if v, ok := raw[k].(string); ok && v != "" {
				meta[k] = v
			}
		}
	}

	w.addEntry(lokiEntry{pushTS: pushTS, line: line, meta: meta})
	return len(p), nil
}

func (w *LokiWriter) addEntry(entry lokiEntry) {
	w.bufferMu.Lock()
	w.buffer = append(w.buffer, entry)
	size := len(w.buffer)
	w.bufferMu.Unlock()

	if size >= 10 {
		go w.flush()
	}
}

func (w *LokiWriter) Close() error {
	close(w.stopCh)
	w.wg.Wait()
	// Final sync flush - send all remaining logs
	if err := w.flush(); err != nil {
		fmt.Printf("loki flush error on close: %v\n", err)
	}
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

	entries := w.buffer
	w.buffer = make([]lokiEntry, 0, 100)
	w.bufferMu.Unlock()

	values := make([][]any, 0, len(entries))
	for _, entry := range entries {
		values = append(values, []any{entry.pushTS, entry.line, entry.meta})
	}

	req := lokiPushRequest{
		Streams: []lokiStream{
			{
				Stream: w.labels,
				Values: values,
			},
		},
	}

	body, err := json.Marshal(req)
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
		return fmt.Errorf("loki returned status %d: %s (entries=%d)", resp.StatusCode, string(respBody), len(entries))
	}

	return nil
}
