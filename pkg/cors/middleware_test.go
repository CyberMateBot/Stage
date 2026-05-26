package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func TestWrap_AllowAllReflectsOrigin(t *testing.T) {
	h := Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config.ConfigCORS{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "https://my-front.vercel.app")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://my-front.vercel.app" {
		t.Fatalf("Allow-Origin = %q, want reflected origin", got)
	}
}

func TestWrap_AllowList(t *testing.T) {
	h := Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config.ConfigCORS{
		AllowedOrigins: []string{"https://allowed.app"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://blocked.app")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Allow-Origin = %q, want empty for blocked origin", got)
	}
}
