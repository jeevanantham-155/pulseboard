package websocket

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	redisservice "github.com/lords/live-polling/backend/services/redis"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	writeTimeout = 10 * time.Second
	pingInterval = 25 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
}

type Manager struct {
	redis       *redisservice.Client
	frontendURL string
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	clients     map[string]map[*client]struct{}
	subscribed  map[string]struct{}
}

type client struct {
	manager   *Manager
	pollID    string
	conn      *websocket.Conn
	send      chan []byte
	closeOnce sync.Once
}

func NewManager(redisClient *redisservice.Client, frontendURL string) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{redis: redisClient, frontendURL: frontendURL, ctx: ctx, cancel: cancel, clients: make(map[string]map[*client]struct{}), subscribed: make(map[string]struct{})}
}

func (manager *Manager) ServeHTTP(context *gin.Context) {
	pollID := context.Param("pollId")
	if !isValidPollID(pollID) {
		context.JSON(http.StatusBadRequest, map[string]string{"message": "invalid poll id"})
		return
	}
	if context.Request.Header.Get("Origin") != "" && context.Request.Header.Get("Origin") != manager.frontendURL {
		context.JSON(http.StatusForbidden, map[string]string{"message": "origin is not allowed"})
		return
	}

	connection, err := upgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		return
	}
	client := &client{manager: manager, pollID: pollID, conn: connection, send: make(chan []byte, 8)}
	manager.addClient(client)
	go client.writeLoop()
	client.readLoop()
}

func (manager *Manager) addClient(connectionClient *client) {
	manager.mu.Lock()
	if manager.clients[connectionClient.pollID] == nil {
		manager.clients[connectionClient.pollID] = make(map[*client]struct{})
	}
	manager.clients[connectionClient.pollID][connectionClient] = struct{}{}
	shouldSubscribe := false
	if _, exists := manager.subscribed[connectionClient.pollID]; !exists {
		manager.subscribed[connectionClient.pollID] = struct{}{}
		shouldSubscribe = true
	}
	manager.mu.Unlock()
	if shouldSubscribe {
		go manager.subscribe(connectionClient.pollID)
	}
}

func (manager *Manager) removeClient(client *client) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if clients := manager.clients[client.pollID]; clients != nil {
		delete(clients, client)
		if len(clients) == 0 {
			delete(manager.clients, client.pollID)
		}
	}
}

func (manager *Manager) subscribe(pollID string) {
	pubsub := manager.redis.SubscribeResults(manager.ctx, pollID)
	defer pubsub.Close()
	if _, err := pubsub.Receive(manager.ctx); err != nil {
		return
	}
	for {
		message, err := pubsub.ReceiveMessage(manager.ctx)
		if err != nil {
			return
		}
		manager.broadcast(pollID, []byte(message.Payload))
	}
}

func (manager *Manager) broadcast(pollID string, payload []byte) {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	for client := range manager.clients[pollID] {
		select {
		case client.send <- payload:
		default:
			go client.close()
		}
	}
}

func (client *client) readLoop() {
	defer client.close()
	client.conn.SetReadLimit(1024)
	client.conn.SetReadDeadline(time.Now().Add(pingInterval + 5*time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(pingInterval + 5*time.Second))
		return nil
	})
	for {
		if _, _, err := client.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (client *client) writeLoop() {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	defer client.close()
	for {
		select {
		case payload := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := client.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *client) close() {
	client.closeOnce.Do(func() {
		client.manager.removeClient(client)
		_ = client.conn.Close()
	})
}

func (manager *Manager) Close() {
	manager.cancel()
	manager.mu.Lock()
	defer manager.mu.Unlock()
	for _, clients := range manager.clients {
		for client := range clients {
			_ = client.conn.Close()
		}
	}
}

func isValidPollID(value string) bool {
	_, err := primitive.ObjectIDFromHex(value)
	return err == nil
}
