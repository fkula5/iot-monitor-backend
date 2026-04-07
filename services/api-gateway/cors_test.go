package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/stretchr/testify/assert"
)

func TestCORSConfiguration(t *testing.T) {
	r := chi.NewRouter()

	corsAllowedOrigins := []string{"http://localhost:5173"}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   corsAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Test case for AllowedHeaders including X-CSRF-Token
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "X-CSRF-Token")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Check for canonicalized header
	assert.Contains(t, rr.Header().Get("Access-Control-Allow-Headers"), "X-Csrf-Token", "X-CSRF-Token should be allowed (canonicalized)")

	// Test case for ExposedHeaders
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Contains(t, rr.Header().Get("Access-Control-Expose-Headers"), "Link", "Link should be exposed")
}
