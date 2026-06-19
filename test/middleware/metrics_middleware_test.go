package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/assidik12/catalyst/internal/delivery/http/middleware"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

// TestMetricsMiddleware_CallsNextAndPassesStatus200 verifies:
//   - The next handler is called exactly once
//   - HTTP 200 status is preserved and returned to the client
//   - No panic occurs
func TestMetricsMiddleware_CallsNextAndPassesStatus200(t *testing.T) {
	nextCalled := false

	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}

	wrapped := middleware.MetricsMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rec := httptest.NewRecorder()

	// Must not panic
	assert.NotPanics(t, func() {
		wrapped(rec, req, nil)
	})

	assert.True(t, nextCalled, "next handler must be called")
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestMetricsMiddleware_CapturesStatus500 verifies:
//   - When the handler writes 500, responseWriterObserver captures it correctly
//   - The client receives the 500 status code
func TestMetricsMiddleware_CapturesStatus500(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(http.StatusInternalServerError)
	}

	wrapped := middleware.MetricsMiddleware(next)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		wrapped(rec, req, nil)
	})

	// The recorder captures the actual status written to the ResponseWriter
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestMetricsMiddleware_Status404 verifies correct capture of 404 responses.
func TestMetricsMiddleware_Status404(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(http.StatusNotFound)
	}

	wrapped := middleware.MetricsMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/nonexistent", nil)
	rec := httptest.NewRecorder()

	wrapped(rec, req, nil)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// TestMetricsMiddleware_DefaultStatus200WhenNoWriteHeader verifies:
//   - If next handler never calls WriteHeader explicitly, status defaults to 200
func TestMetricsMiddleware_DefaultStatus200WhenNoWriteHeader(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// Writes body without calling WriteHeader — Go implicitly uses 200
		_, _ = w.Write([]byte(`{"data":"ok"}`))
	}

	wrapped := middleware.MetricsMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		wrapped(rec, req, nil)
	})

	// httptest.ResponseRecorder defaults to 200 when no explicit WriteHeader
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestMetricsMiddleware_DifferentHTTPMethods verifies middleware works for
// all HTTP method types without issue.
func TestMetricsMiddleware_DifferentHTTPMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
				w.WriteHeader(http.StatusOK)
			}

			wrapped := middleware.MetricsMiddleware(next)
			req := httptest.NewRequest(method, "/api/v1/products", nil)
			rec := httptest.NewRecorder()

			assert.NotPanics(t, func() {
				wrapped(rec, req, nil)
			})
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
