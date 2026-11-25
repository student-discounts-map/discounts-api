package httpapi

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	NewRouter(slog.New(slog.DiscardHandler)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("код ответа: получили %d, ожидали %d", rec.Code, http.StatusOK)
	}
	if got, want := rec.Body.String(), `{"status":"ok"}`; got != want {
		t.Errorf("тело ответа: получили %s, ожидали %s", got, want)
	}
}

func TestRequestLog(t *testing.T) {
	var logs bytes.Buffer
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/unknown?token=secret", nil)

	NewRouter(slog.New(slog.NewTextHandler(&logs, nil))).ServeHTTP(rec, req)

	for _, want := range []string{"method=GET", "path=/unknown", "status=404"} {
		if !strings.Contains(logs.String(), want) {
			t.Errorf("в логе нет %q: %s", want, logs.String())
		}
	}
	if strings.Contains(logs.String(), "secret") {
		t.Errorf("в лог попали параметры запроса: %s", logs.String())
	}
}

func TestRecoverPanics(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	broken := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("подробности сбоя")
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	logRequests(logger, recoverPanics(logger, broken)).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("код ответа: получили %d, ожидали %d", rec.Code, http.StatusInternalServerError)
	}
	if strings.Contains(rec.Body.String(), "подробности сбоя") {
		t.Errorf("подробности сбоя попали в ответ: %s", rec.Body.String())
	}
	for _, want := range []string{"подробности сбоя", "status=500"} {
		if !strings.Contains(logs.String(), want) {
			t.Errorf("в логе нет %q: %s", want, logs.String())
		}
	}
}
