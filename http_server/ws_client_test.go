package httpServer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jxncyjq/stardust/codec"
)

// TestWebSocketClient 测试 WebSocket 客户端功能
// 验证: 客户端连接、发送消息、接收响应
func TestWebSocketClient(t *testing.T) {
	jsonCodec := codec.NewJsonCodec()
	handler := &TestMessageHandler{}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		t.Logf("Server received: %s", msg)

		message, _ := jsonCodec.Decode(msg)
		response, _ := handler.HandlerMessage(message)
		conn.WriteMessage(websocket.TextMessage, []byte(response))
		t.Logf("Server sent: %s", response)
	}))
	defer server.Close()

	// 客户端连接
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()
	t.Log("Client connected")

	// 客户端发送消息
	testMsg := `{"type":"ping","data":"hello"}`
	conn.WriteMessage(websocket.TextMessage, []byte(testMsg))
	t.Logf("Client sent: %s", testMsg)

	// 客户端接收响应
	_, response, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	t.Logf("Client received: %s", response)

	respMsg, _ := jsonCodec.Decode(response)
	if respMsg.GetType() != "echo" {
		t.Errorf("Expected type 'echo', got '%s'", respMsg.GetType())
	}
}

// TestWebSocketClientMultipleMessages 测试客户端多次消息收发
func TestWebSocketClientMultipleMessages(t *testing.T) {
	jsonCodec := codec.NewJsonCodec()
	handler := &TestMessageHandler{}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			message, _ := jsonCodec.Decode(msg)
			response, _ := handler.HandlerMessage(message)
			conn.WriteMessage(websocket.TextMessage, []byte(response))
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	// 发送多条消息
	messages := []string{"msg1", "msg2", "msg3"}
	for _, msg := range messages {
		testMsg := `{"type":"` + msg + `","data":"test"}`
		conn.WriteMessage(websocket.TextMessage, []byte(testMsg))

		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, response, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Read failed for %s: %v", msg, err)
		}
		t.Logf("Sent: %s, Received: %s", msg, response)
	}
}
