package httpServer

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestSSEHandlerSendsEvents(t *testing.T) {
	config := []byte(`{"address":"127.0.0.1","port":0,"path":"/","worker_id":1}`)
	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("NewHttpServer() error = %v", err)
	}

	type payload struct {
		Msg string `json:"msg"`
	}

	server.Get("events", "", NewSSEHandler("events", SSEOptions{
		Handler: func(c *gin.Context, send SSESendFunc) {
			_ = send("update", payload{Msg: "hello"})
			_ = send("", "plain text")
		},
	}))

	ts := httptest.NewServer(server.Engine())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/api/events", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/events error = %v", err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want %q", ct, "text/event-stream")
	}

	var lines []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	want := []string{
		"event: update",
		`data: {"msg":"hello"}`,
		"",
		"data: plain text",
		"",
	}
	if len(lines) < len(want) {
		t.Fatalf("got %d lines, want at least %d\nlines: %v", len(lines), len(want), lines)
	}
	for i, w := range want {
		if lines[i] != w {
			t.Errorf("line[%d] = %q, want %q", i, lines[i], w)
		}
	}
}

func TestSSEHandlerRetryDirective(t *testing.T) {
	config := []byte(`{"address":"127.0.0.1","port":0,"path":"/","worker_id":1}`)
	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("NewHttpServer() error = %v", err)
	}

	server.Get("retry", "", NewSSEHandler("retry", SSEOptions{
		RetryMs: 3000,
		Handler: func(c *gin.Context, send SSESendFunc) {
			_ = send("ping", "1")
		},
	}))

	ts := httptest.NewServer(server.Engine())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/api/retry", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/retry error = %v", err)
	}
	defer resp.Body.Close()

	var lines []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	found := false
	for _, l := range lines {
		if strings.HasPrefix(l, "retry:") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("no retry directive found in: %v", lines)
	}
}
