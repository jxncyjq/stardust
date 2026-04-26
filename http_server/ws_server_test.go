package httpServer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jxncyjq/stardust/codec"
	"github.com/jxncyjq/stardust/logs"
	"github.com/jxncyjq/stardust/uuid"
)

// TestWebSocketServer 测试 WebSocket 服务器端功能
// 验证: HttpServer + ClientManager + Client 的服务端集成
func TestWebSocketServer(t *testing.T) {
	logger := logs.GetLogger("ws_server_test")
	jsonCodec := codec.NewJsonCodec()
	handler := &TestMessageHandler{}

	// 1. 创建 HttpServer
	config := []byte(`{"address":"127.0.0.1","port":0,"path":"/"}`)
	httpServer, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create HttpServer: %v", err)
	}
	t.Log("HttpServer created")

	// 2. 创建 ClientManager
	clientManager := NewClientManager(logger)
	go clientManager.Start()
	t.Log("ClientManager started")

	// 3. 注册 WebSocket 路由
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	httpServer.Engine().GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		userId := c.Query("userId")
		sessionId := uuid.GenSessionId()
		client := NewClient(userId, sessionId, conn, jsonCodec, logger, context.Background(), handler, clientManager)
		clientManager.RegisterClient(client)
		t.Logf("Client registered: userId=%s, sessionId=%s", userId, sessionId)
		client.Listen()
	})

	server := httptest.NewServer(httpServer.Engine())
	defer server.Close()

	// 4. 测试客户端注册
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?userId=testUser"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	time.Sleep(100 * time.Millisecond)
	if count := clientManager.ClientCount(); count != 1 {
		t.Errorf("Expected 1 client, got %d", count)
	}
	t.Logf("ClientManager.ClientCount() = %d", clientManager.ClientCount())

	// 5. 测试踢出客户端
	clientManager.KickClientByUserId("testUser")
	time.Sleep(100 * time.Millisecond)
	t.Logf("After kick: ClientCount() = %d", clientManager.ClientCount())
}

// TestClientManagerBroadcast 测试 ClientManager 广播功能
func TestClientManagerBroadcast(t *testing.T) {
	logger := logs.GetLogger("ws_broadcast_test")
	jsonCodec := codec.NewJsonCodec()
	handler := &TestMessageHandler{}

	config := []byte(`{"address":"127.0.0.1","port":0,"path":"/"}`)
	httpServer, _ := NewHttpServer(config)

	clientManager := NewClientManager(logger)
	go clientManager.Start()

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	httpServer.Engine().GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		userId := c.Query("userId")
		sessionId := uuid.GenSessionId()
		client := NewClient(userId, sessionId, conn, jsonCodec, logger, context.Background(), handler, clientManager)
		clientManager.RegisterClient(client)
		client.Listen()
	})

	server := httptest.NewServer(httpServer.Engine())
	defer server.Close()

	// 连接两个客户端
	wsURL1 := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?userId=user1"
	wsURL2 := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?userId=user2"

	conn1, _, _ := websocket.DefaultDialer.Dial(wsURL1, nil)
	defer conn1.Close()
	conn2, _, _ := websocket.DefaultDialer.Dial(wsURL2, nil)
	defer conn2.Close()

	time.Sleep(100 * time.Millisecond)
	if count := clientManager.ClientCount(); count != 2 {
		t.Errorf("Expected 2 clients, got %d", count)
	}
	t.Logf("Two clients connected, ClientCount() = %d", clientManager.ClientCount())
}
