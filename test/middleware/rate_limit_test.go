package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/assidik12/catalyst/internal/delivery/http/middleware"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit(t *testing.T) {
	// Setup a dummy HTTP handler
	dummyHandler := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}

	// Wrap with RateLimit(2, 2)
	handler := middleware.RateLimit(2, 2)(dummyHandler)

	ip := "192.168.1.1:1234"

	// Request 1
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = ip
	rec1 := httptest.NewRecorder()
	handler(rec1, req1, nil)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Request 2
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = ip
	rec2 := httptest.NewRecorder()
	handler(rec2, req2, nil)
	assert.Equal(t, http.StatusOK, rec2.Code)

	// Request 3
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.RemoteAddr = ip
	rec3 := httptest.NewRecorder()
	handler(rec3, req3, nil)
	assert.Equal(t, http.StatusTooManyRequests, rec3.Code)
}
