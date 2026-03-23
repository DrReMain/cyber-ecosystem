package encoder

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kratos/kratos/v2/errors"
)

func TestNewErrorEncoder(t *testing.T) {
	// Create a simple buildBody function
	buildBody := func(err *errors.Error) any {
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
	buildBody := func(err *errors.Error) any {
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
	// Create a simple buildBody function
	buildBody := func(v any) (any, error) {
		return map[string]any{
			"data": v,
		}, nil
	}

	encoder := NewResponseEncoder(buildBody)
	if encoder == nil {
		t.Error("NewResponseEncoder should return a non-nil function")
	}
}

func TestNewResponseEncoder_EncodesResponse(t *testing.T) {
	buildBody := func(v any) (any, error) {
		return map[string]any{
			"data":    v,
			"success": true,
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
}

func TestNewResponseEncoder_NilValue(t *testing.T) {
	buildBody := func(v any) (any, error) {
		return v, nil
	}

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
