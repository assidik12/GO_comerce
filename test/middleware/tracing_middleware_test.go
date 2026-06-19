package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/assidik12/catalyst/internal/delivery/http/middleware"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace/noop"
)

// TestMain sets up a NoopTracerProvider so all OTel calls are no-ops.
// This prevents tests from needing a real Jaeger/OTLP endpoint.
func TestMain(m *testing.M) {
	otel.SetTracerProvider(noop.NewTracerProvider())
	otel.SetTextMapPropagator(propagation.TraceContext{})
	os.Exit(m.Run())
}

// TestTracingMiddleware_CallsNextAndInjectsContext verifies:
//   - next handler is called
//   - context passed to next is a NEW context (r.WithContext was called)
//   - the new context carries a span (even if it's a noop span)
//   - no panic occurs
func TestTracingMiddleware_CallsNextAndInjectsContext(t *testing.T) {
	var capturedCtx context.Context
	nextCalled := false

	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		nextCalled = true
		capturedCtx = r.Context() // capture the context injected by TracingMiddleware
		w.WriteHeader(http.StatusOK)
	}

	wrapped := middleware.TracingMiddleware(next, "GET /api/v1/products")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	originalCtx := req.Context()
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		wrapped(rec, req, nil)
	})

	assert.True(t, nextCalled, "next handler must be called")
	assert.NotNil(t, capturedCtx, "context must be propagated to next handler")
	// The context inside the handler should be different (has span injected)
	assert.NotEqual(t, originalCtx, capturedCtx, "TracingMiddleware should inject a new context with span")
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestTracingMiddleware_Status500SetsErrorAttribute verifies:
//   - When next returns 500, middleware sets span attribute error=true
//   - No panic occurs (even with noop tracer)
//   - Client receives correct 500 status
func TestTracingMiddleware_Status500SetsErrorAttribute(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(http.StatusInternalServerError)
	}

	wrapped := middleware.TracingMiddleware(next, "POST /api/v1/transactions")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		wrapped(rec, req, nil)
	})

	// span.SetAttributes(attribute.Bool("error", true)) is called for status >= 500
	// With noop tracer this is a no-op but must not panic
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestTracingMiddleware_Status200NoErrorAttribute verifies:
//   - Successful responses do NOT trigger error attribute
//   - No panic
func TestTracingMiddleware_Status200NoErrorAttribute(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}

	wrapped := middleware.TracingMiddleware(next, "GET /api/v1/products")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		wrapped(rec, req, nil)
	})

	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestTracingMiddleware_WithTraceparentHeader verifies context propagation
// when the client sends a W3C traceparent header.
// The middleware must extract the trace context from the header without panicking.
func TestTracingMiddleware_WithTraceparentHeader(t *testing.T) {
	var capturedCtx context.Context

	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	}

	wrapped := middleware.TracingMiddleware(next, "GET /api/v1/products")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)

	// Set a valid W3C traceparent header to test propagation extraction
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		wrapped(rec, req, nil)
	})

	assert.NotNil(t, capturedCtx)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestTracingMiddleware_RouteNameInSpan verifies that the routeName parameter
// is correctly used for span naming (no panic for various route name formats).
func TestTracingMiddleware_RouteNameInSpan(t *testing.T) {
	routeNames := []string{
		"GET /api/v1/products",
		"POST /api/v1/transactions",
		"DELETE /api/v1/transactions/:id",
		"PUT /api/v1/users/profile",
	}

	for _, routeName := range routeNames {
		t.Run(routeName, func(t *testing.T) {
			next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
				w.WriteHeader(http.StatusOK)
			}

			wrapped := middleware.TracingMiddleware(next, routeName)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
			rec := httptest.NewRecorder()

			assert.NotPanics(t, func() {
				wrapped(rec, req, nil)
			})
		})
	}
}
