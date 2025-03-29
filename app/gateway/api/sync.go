package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"

	"github.com/crazyfrankie/cloudstorage/app/gateway/mws"
)

type FileChangeEvent struct {
	EventType string    `json:"event_type"`
	FileId    int64     `json:"file_id"`
	FolderId  int64     `json:"folder_id"`
	UserId    int32     `json:"user_id"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
}

// ConnectionManager WebSocket 连接管理器
type ConnectionManager struct {
	// 按用户ID分组的连接
	connections map[int32][]*websocket.Conn
	mu          sync.RWMutex
	reader      *kafka.Reader
	stopCh      chan struct{}
}

func NewConnectionManager() *ConnectionManager {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"localhost:9092"},
		Topic:       "file-changes",
		GroupID:     "gateway-websocket",
		StartOffset: kafka.LastOffset,
	})

	cm := &ConnectionManager{
		connections: make(map[int32][]*websocket.Conn),
		reader:      reader,
		stopCh:      make(chan struct{}),
	}

	// 启动消息监听
	go cm.listenForMessages()

	return cm
}

func (cm *ConnectionManager) listenForMessages() {
	for {
		select {
		case <-cm.stopCh:
			return
		default:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			message, err := cm.reader.ReadMessage(ctx)
			cancel()

			if err != nil {
				if !errors.Is(err, context.DeadlineExceeded) {
					log.Printf("Error reading Kafka message: %v", err)
				}
				time.Sleep(time.Second)
				continue
			}

			// 解析消息
			var event FileChangeEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				log.Printf("Error unmarshaling event: %v", err)
				continue
			}

			// 发送给相应用户的所有连接
			cm.broadcast(event)
		}
	}
}

func (cm *ConnectionManager) AddConnection(userId int32, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.connections[userId] = append(cm.connections[userId], conn)

	// 清理关闭的连接
	go func() {
		cm.removeConnection(userId, conn)
	}()
}

func (cm *ConnectionManager) removeConnection(userId int32, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conns := cm.connections[userId]
	for i, c := range conns {
		if c == conn {
			cm.connections[userId] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

func (cm *ConnectionManager) broadcast(event FileChangeEvent) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 获取该用户的所有连接
	conns := cm.connections[event.UserId]
	for _, conn := range conns {
		if err := conn.WriteJSON(event); err != nil {
			log.Printf("Error sending event to client: %v", err)
			// 连接可能已断开，后台会自动清理
		}
	}
}

func (cm *ConnectionManager) Close() error {
	close(cm.stopCh)
	return cm.reader.Close()
}

// SyncHandler 处理 WebSocket 连接
type SyncHandler struct {
	connManager *ConnectionManager
}

func NewSyncHandler(cm *ConnectionManager) *SyncHandler {
	return &SyncHandler{connManager: cm}
}

func (h *SyncHandler) RegisterRoute(r *gin.Engine) {
	r.GET("/api/sync/ws", mws.Auth(), h.HandleWebSocket())
}

// HandleWebSocket 处理 WebSocket 连接
func (h *SyncHandler) HandleWebSocket() gin.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源，生产环境应该限制
		},
	}

	return func(c *gin.Context) {
		claims := c.MustGet("claims").(*mws.Claim)
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Error upgrading to WebSocket: %v", err)
			return
		}

		// 注册连接
		h.connManager.AddConnection(claims.UserId, conn)

		// 发送初始连接成功消息
		conn.WriteJSON(map[string]string{
			"type":    "connected",
			"message": "WebSocket连接成功",
		})

		// 保持连接
		for {
			// 读取客户端消息，主要是为了检测断开
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}
}
