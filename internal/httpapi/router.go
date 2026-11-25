package httpapi

import (
	"io"
	"log/slog"
	"net/http"
)

func NewRouter(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(logger, w, http.StatusOK, `{"status":"ok"}`)
	})
	return logRequests(logger, recoverPanics(logger, mux))
}

func writeJSON(logger *slog.Logger, w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := io.WriteString(w, body); err != nil {
		logger.Error("не удалось отправить ответ", "error", err)
	}
}
