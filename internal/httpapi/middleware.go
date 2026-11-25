package httpapi

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.status == 0 && status >= http.StatusOK {
		r.status = status
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(body)
}

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}

		defer func() {
			status := rec.status
			if status == 0 {
				status = http.StatusOK
			}
			logger.Info("запрос",
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"duration", time.Since(start),
			)
		}()

		next.ServeHTTP(rec, r)
	})
}

func recoverPanics(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w}

		defer func() {
			reason := recover()
			if reason == nil {
				return
			}
			if reason == http.ErrAbortHandler {
				panic(reason)
			}

			logger.Error("паника при обработке запроса",
				"method", r.Method,
				"path", r.URL.Path,
				"error", reason,
				"stack", string(debug.Stack()),
			)

			if rec.status != 0 {
				panic(http.ErrAbortHandler)
			}

			w.Header().Del("Content-Length")
			writeJSON(logger, w, http.StatusInternalServerError, `{"error":"внутренняя ошибка сервера"}`)
		}()

		next.ServeHTTP(rec, r)
	})
}
