package httpServer

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jxncyjq/stardust/codec"
	"github.com/jxncyjq/stardust/logs"
)

func TestWebSocketHandlerRegistersThroughHttpServer(t *testing.T) {
	logger := logs.GetLogger("ws_handler_test")
	manager := NewClientManager(logger)
	go manager.Start()
	t.Cleanup(manager.Stop)

	config := []byte(`{"address":"127.0.0.1","port":0,"path":"/","worker_id":1}`)
	server, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("NewHttpServer() error = %v", err)
	}

	server.Get("ws", "", NewWebSocketHandler("ws", WebSocketOptions{
		Codec:   codec.NewJsonCodec(),
		Logger:  logger,
		Manager: manager,
		Handler: &TestMessageHandler{},
	}))

	testServer := httptest.NewServer(server.Engine())
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/api/ws?userId=testUser"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("websocket.Dial(%q) error = %v", wsURL, err)
	}
	defer conn.Close()

	time.Sleep(100 * time.Millisecond)
	if count := manager.ClientCount(); count != 1 {
		t.Fatalf("ClientManager.ClientCount() = %d, want %d", count, 1)
	}

	const request = `{"type":"hello","data":"world"}`
	if err := conn.WriteMessage(websocket.TextMessage, []byte(request)); err != nil {
		t.Fatalf("Conn.WriteMessage(%q) error = %v", request, err)
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, response, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Conn.ReadMessage() error = %v", err)
	}
	const want = `{"type":"echo","data":"hello"}`
	if string(response) != want {
		t.Errorf("Conn.ReadMessage() = %q, want %q", string(response), want)
	}
}
