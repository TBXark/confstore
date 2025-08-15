package confstore

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_Load_WithCustomHTTPClientTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = io.WriteString(w, "{}")
	}))
	defer ts.Close()

	client := &http.Client{Timeout: 10 * time.Millisecond}
	_, err := Load[map[string]any](ts.URL, WithHTTPClientOption(client))
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	var nerr net.Error
	if !errors.As(err, &nerr) || !nerr.Timeout() {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}

func Test_Save_WithCustomHTTPClientTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := &http.Client{Timeout: 10 * time.Millisecond}
	payload := map[string]any{"k": "v"}
	err := Save(ts.URL, &payload, WithHTTPClientOption(client))
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	var nerr net.Error
	if !errors.As(err, &nerr) || !nerr.Timeout() {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}
