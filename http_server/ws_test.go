package httpServer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jxncyjq/stardust/codec"
	"github.com/jxncyjq/stardust/logs"
	"github.com/jxncyjq/stardust/uuid"
)

func TestMain(m *testing.M) {
	logs.Init([]byte(`{"level":-1}`))
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// TestMessageHandler 测试消息处理器
type TestMessageHandler struct{}

func (h *TestMessageHandler) HandlerMessage(message codec.IMessage) (string, error) {
	return `{"type":"echo","data":"` + message.GetType() + `"}`, nil
}

func TestWebSocketWithJsonCodec(t *testing.T) {
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

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	testMsg := `{"type":"ping","data":"hello"}`
	conn.WriteMessage(websocket.TextMessage, []byte(testMsg))
	t.Logf("Client sent: %s", testMsg)

	_, response, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	t.Logf("Client received: %s", response)

	respMsg, _ := jsonCodec.Decode(response)
	if respMsg.GetType() != "echo" {
		t.Errorf("Expected type 'echo', got '%s'", respMsg.GetType())
	}
	t.Log("WebSocket test with JsonCodec passed!")
}

// TestWebSocketIntegration 测试 HttpServer + ClientManager + Client 的完整集成
// 关系说明:
// - HttpServer: 提供 HTTP/WebSocket 服务，处理连接升级
// - ClientManager: 管理所有 WebSocket 客户端，支持注册、注销、广播、踢出
// - Client: 单个 WebSocket 连接的封装，负责消息收发
func TestWebSocketIntegration(t *testing.T) {
	logger := logs.GetLogger("ws_integration_test")
	jsonCodec := codec.NewJsonCodec()
	handler := &TestMessageHandler{}

	// 1. 创建 HttpServer - 提供 HTTP 服务
	config := []byte(`{"address":"127.0.0.1","port":0,"path":"/"}`)
	httpServer, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create HttpServer: %v", err)
	}
	t.Log("1. HttpServer created - provides HTTP/WebSocket service")

	// 2. 创建 ClientManager - 管理所有客户端连接
	clientManager := NewClientManager(logger)
	go clientManager.Start()
	t.Log("2. ClientManager started - manages all client connections")

	// 3. 在 HttpServer 上注册 WebSocket 路由
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	httpServer.Engine().GET("/ws", func(c *gin.Context) {
		// WebSocket 连接升级
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		// 4. 创建 Client - 封装单个 WebSocket 连接
		userId := c.Query("userId")
		sessionId := uuid.GenSessionId() // 使用 Snowflake 生成唯一 sessionId
		client := NewClient(userId, sessionId, conn, jsonCodec, logger, context.Background(), handler, clientManager)

		// 5. 将 Client 注册到 ClientManager
		clientManager.RegisterClient(client)
		t.Logf("   Client registered: userId=%s, sessionId=%s", userId, sessionId)

		// 6. 启动 Client 监听 (收发消息)
		client.Listen()
	})
	t.Log("3. WebSocket route /ws registered on HttpServer")

	// 创建测试服务器
	server := httptest.NewServer(httpServer.Engine())
	defer server.Close()

	// 7. 客户端1连接 (sessionId 由服务端自动生成)
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?userId=user1"
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Client 1 dial failed: %v", err)
	}
	defer conn1.Close()
	t.Log("4. Client 1 connected via HttpServer -> registered to ClientManager")

	time.Sleep(100 * time.Millisecond)

	// 8. 验证 ClientManager 管理的客户端数量
	if count := clientManager.ClientCount(); count != 1 {
		t.Errorf("Expected 1 client, got %d", count)
	}
	t.Logf("5. ClientManager.ClientCount() = %d", clientManager.ClientCount())

	// 9. 客户端2连接 (sessionId 由服务端自动生成)
	wsURL2 := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?userId=user2"
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("Client 2 dial failed: %v", err)
	}
	defer conn2.Close()
	t.Log("6. Client 2 connected")

	time.Sleep(100 * time.Millisecond)
	t.Logf("7. ClientManager.ClientCount() = %d", clientManager.ClientCount())

	// 10. Client 发送消息 -> Handler 处理 -> 返回响应
	testMsg := `{"type":"hello","data":"world"}`
	conn1.WriteMessage(websocket.TextMessage, []byte(testMsg))
	t.Logf("8. Client 1 sent: %s", testMsg)

	conn1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, response, err := conn1.ReadMessage()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	t.Logf("9. Client 1 received: %s", response)

	// 11. ClientManager 踢出客户端
	clientManager.KickClientByUserId("user1")
	time.Sleep(100 * time.Millisecond)
	t.Logf("10. ClientManager.KickClientByUserId('user1') -> ClientCount() = %d", clientManager.ClientCount())

	t.Log("WebSocket Integration test passed!")
	t.Log("")
	t.Log("=== 关系总结 ===")
	t.Log("HttpServer: 提供 HTTP 服务，注册 /ws 路由处理 WebSocket 升级")
	t.Log("ClientManager: 管理所有 Client，支持注册/注销/广播/踢出")
	t.Log("Client: 封装单个 WebSocket 连接，处理消息收发")
}
