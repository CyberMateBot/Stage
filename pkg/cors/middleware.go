package cors

import (
	"net/http"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

// Wrap adds CORS headers for browser clients (Telegram Mini App / Vite dev server).
func Wrap(next http.Handler, cfg config.ConfigCORS) http.Handler {
	allowed := normalizeOrigins(cfg.AllowedOrigins)
	allowAll := len(allowed) == 1 && allowed[0] == "*"
	methods := strings.Join(cfg.AllowedMethods, ", ")
	headers := strings.Join(cfg.AllowedHeaders, ", ")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if allowAll || originAllowed(origin, allowed) {
				w.Header().Set("Access-Control-Allow-Origin", pickAllowOrigin(origin, allowAll))
				w.Header().Add("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func normalizeOrigins(origins []string) []string {
	out := make([]string, 0, len(origins))
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o != "" {
			out = append(out, o)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}

func originAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == origin {
			return true
		}
	}
	return false
}

func pickAllowOrigin(requestOrigin string, allowAll bool) string {
	if allowAll {
		if requestOrigin != "" {
			return requestOrigin
		}
		return "*"
	}
	return requestOrigin
}
