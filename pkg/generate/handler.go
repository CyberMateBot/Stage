package generate

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"github.com/twelvepills-936/tgapp-/pkg/ai"
)

const (
	pathGenerateText  = "/v1/generate/text"
	pathGenerateImage = "/v1/generate/image"
)

// Wrap adds POST /v1/generate/text and POST /v1/generate/image.
func Wrap(next http.Handler, svc *ai.Service) http.Handler {
	if svc == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == pathGenerateText:
			handleText(w, r, svc)
			return
		case r.Method == http.MethodPost && r.URL.Path == pathGenerateImage:
			handleImage(w, r, svc)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleText(w http.ResponseWriter, r *http.Request, svc *ai.Service) {
	var req ai.TextRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := svc.GenerateText(r.Context(), req)
	if err != nil {
		writeServiceError(w, r, "generate text", err)
		return
	}

	writeJSON(w, http.StatusOK, out)
}

func handleImage(w http.ResponseWriter, r *http.Request, svc *ai.Service) {
	var req ai.ImageRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := svc.GenerateImage(r.Context(), req)
	if err != nil {
		writeServiceError(w, r, "generate image", err)
		return
	}

	writeJSON(w, http.StatusOK, out)
}

func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

func writeServiceError(w http.ResponseWriter, r *http.Request, op string, err error) {
	switch {
	case errors.Is(err, ai.ErrPromptEmpty):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ai.ErrNotConfigured):
		writeError(w, http.StatusServiceUnavailable, err.Error())
	default:
		var pe *ai.ProviderError
		if errors.As(err, &pe) {
			slog.ErrorContext(r.Context(), op+" provider error",
				slog.String("provider", pe.Provider),
				slog.Int("status", pe.Status),
				slog.String("message", pe.Message),
			)
			writeError(w, http.StatusBadGateway, pe.Error())
			return
		}
		slog.ErrorContext(r.Context(), op, slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "generation failed")
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
