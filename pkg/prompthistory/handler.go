package prompthistory

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

const pathPrefix = "/v1/prompts/history/telegram/"

// Wrap adds prompt history HTTP routes.
func Wrap(next http.Handler, store *Store) http.Handler {
	if store == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/prompts/history":
			handleSave(w, r, store)
			return
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, pathPrefix):
			telegramID := strings.TrimPrefix(r.URL.Path, pathPrefix)
			if telegramID != "" && !strings.Contains(telegramID, "/") {
				handleList(w, r, store, telegramID)
				return
			}
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, pathPrefix):
			telegramID := strings.TrimPrefix(r.URL.Path, pathPrefix)
			if telegramID != "" && !strings.Contains(telegramID, "/") {
				handleDelete(w, r, store, telegramID)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func handleSave(w http.ResponseWriter, r *http.Request, store *Store) {
	var req saveRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.TelegramID) == "" {
		writeError(w, http.StatusBadRequest, "telegramId is required")
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		writeError(w, http.StatusBadRequest, "prompt is required")
		return
	}

	item, err := store.Insert(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save prompt history")
		return
	}
	writeJSON(w, http.StatusOK, saveResponse{Item: item})
}

func handleList(w http.ResponseWriter, r *http.Request, store *Store, telegramID string) {
	items, err := store.ListByTelegram(r.Context(), telegramID, 200)
	if err != nil {
		slog.ErrorContext(r.Context(), "load prompt history",
			slog.String("telegram_id", telegramID),
			slog.Any("error", err),
		)
		writeError(w, http.StatusInternalServerError, "failed to load prompt history")
		return
	}
	if items == nil {
		items = []Item{}
	}
	writeJSON(w, http.StatusOK, listResponse{Items: items})
}

func handleDelete(w http.ResponseWriter, r *http.Request, store *Store, telegramID string) {
	if err := store.DeleteByTelegram(r.Context(), telegramID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to clear prompt history")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	defer r.Body.Close()
	return json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(dst)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
