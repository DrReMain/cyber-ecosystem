package encoder

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kratos/kratos/v2/errors"
)

func TestNewErrorEncoder(t *testing.T) {
	// Create a simple buildBody function
	buildBody := func(_ context.Context, _ error, err *errors.Error) any {
		return map[string]any{
			"code":    err.Code,
			"message": err.Message,
		}
	}

	encoder := NewErrorEncoder(buildBody)
	if encoder == nil {
		t.Error("NewErrorEncoder should return a non-nil function")
	}
}

func TestNewErrorEncoder_EncodesError(t *testing.T) {
	buildBody := func(_ context.Context, _ error, err *errors.Error) any {
		return map[string]any{
			"code":    err.Code,
			"message": err.Message,
			"reason":  err.Reason,
		}
	}

	encoder := NewErrorEncoder(buildBody)

	// Create a test error
	testErr := errors.NotFound("USER_NOT_FOUND", "user not found")

	// Create test request and response
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()

	// Call the encoder
	encoder(w, req, testErr)

	// Verify response
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}

	if w.Header().Get("Content-Type") == "" {
		t.Error("Content-Type header should be set")
	}
}

func TestNewResponseEncoder(t *testing.T) {
	buildBody := func(v any) (any, error) {
		return map[string]any{"result": v}, nil
	}
	encoder := NewResponseEncoder(buildBody)
	if encoder == nil {
		t.Error("NewResponseEncoder should return a non-nil function")
	}
}

func TestNewResponseEncoder_EncodesResponse(t *testing.T) {
	buildBody := func(v any) (any, error) {
		return map[string]any{
			"success": true,
			"result":  v,
		}, nil
	}
	encoder := NewResponseEncoder(buildBody)

	// Create test request and response
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()

	// Test data
	testData := map[string]string{"name": "test"}

	// Call the encoder
	err := encoder(w, req, testData)
	if err != nil {
		t.Errorf("Encoder should not return error: %v", err)
	}

	// Verify Content-Type header
	if w.Header().Get("Content-Type") == "" {
		t.Error("Content-Type header should be set")
	}

	// Verify body is the wrapped response body
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("response body should be valid json: %v", err)
	}
	if got["success"] != true {
		t.Fatalf("expected success=true, got: %#v", got)
	}
	result, ok := got["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected object result field, got: %#v", got["result"])
	}
	if result["name"] != "test" {
		t.Fatalf("expected result.name=test, got: %#v", result)
	}
}

func TestNewResponseEncoder_NilValue(t *testing.T) {
	buildBody := func(v any) (any, error) { return v, nil }
	encoder := NewResponseEncoder(buildBody)

	// Create test request and response
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Call the encoder with nil value
	err := encoder(w, req, nil)
	if err != nil {
		t.Errorf("Encoder should not return error for nil value: %v", err)
	}
}
