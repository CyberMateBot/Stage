package health

import "net/http"

// Wrap handles GET /health and GET /healthz (Railway / load balancers).
func Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && (r.URL.Path == "/health" || r.URL.Path == "/healthz") {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
