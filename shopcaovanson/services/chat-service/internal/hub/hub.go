package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/caovanson/shopcaovanson/chat-service/internal/domain"
	"github.com/caovanson/shopcaovanson/chat-service/internal/service"
	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"
)

type Hub struct {
	chatSvc     *service.ChatService
	redis       *redis.Client
	mu          sync.RWMutex
	rooms       map[string]map[*websocket.Conn]string
	subscribers map[string]context.CancelFunc
}

func NewHub(chatSvc *service.ChatService, redisClient *redis.Client) *Hub {
	return &Hub{
		chatSvc:     chatSvc,
		redis:       redisClient,
		rooms:       make(map[string]map[*websocket.Conn]string),
		subscribers: make(map[string]context.CancelFunc),
	}
}

func (h *Hub) Register(conn *websocket.Conn, userID, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = make(map[*websocket.Conn]string)
	}
	h.rooms[roomID][conn] = userID

	if _, subscribed := h.subscribers[roomID]; !subscribed {
		ctx, cancel := context.WithCancel(context.Background())
		h.subscribers[roomID] = cancel
		go h.listenRedis(ctx, roomID)
	}

	h.broadcastLocal(roomID, domain.WSServerMessage{
		Type:   "user_joined",
		RoomID: roomID,
		UserID: userID,
	}, conn)
}

func (h *Hub) Unregister(conn *websocket.Conn, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[roomID]
	if !ok {
		return
	}
	delete(room, conn)
	if len(room) == 0 {
		delete(h.rooms, roomID)
		if cancel, ok := h.subscribers[roomID]; ok {
			cancel()
			delete(h.subscribers, roomID)
		}
	}
}

func (h *Hub) HandleMessage(ctx context.Context, userID string, msg domain.WSClientMessage) error {
	switch msg.Type {
	case "join":
		return h.chatSvc.ValidateRoomAccess(ctx, userID, msg.RoomID)
	case "message":
		saved, err := h.chatSvc.SaveMessage(ctx, userID, msg.RoomID, msg.Content)
		if err != nil {
			return err
		}
		out := domain.WSServerMessage{
			Type:      "message",
			RoomID:    msg.RoomID,
			SenderID:  saved.SenderID,
			Content:   saved.Content,
			CreatedAt: saved.CreatedAt.UTC().Format(time.RFC3339),
		}
		payload, err := json.Marshal(out)
		if err != nil {
			return err
		}
		channel := fmt.Sprintf("chat:%s", msg.RoomID)
		return h.redis.Publish(ctx, channel, payload).Err()
	case "typing":
		if err := h.chatSvc.ValidateRoomAccess(ctx, userID, msg.RoomID); err != nil {
			return err
		}
		out := domain.WSServerMessage{
			Type:   "typing",
			RoomID: msg.RoomID,
			UserID: userID,
		}
		payload, err := json.Marshal(out)
		if err != nil {
			return err
		}
		channel := fmt.Sprintf("chat:%s", msg.RoomID)
		return h.redis.Publish(ctx, channel, payload).Err()
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

func (h *Hub) listenRedis(ctx context.Context, roomID string) {
	channel := fmt.Sprintf("chat:%s", roomID)
	pubsub := h.redis.Subscribe(ctx, channel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var out domain.WSServerMessage
			if err := json.Unmarshal([]byte(msg.Payload), &out); err != nil {
				log.Printf("redis payload decode error: %v", err)
				continue
			}
			h.broadcastLocal(roomID, out, nil)
		}
	}
}

func (h *Hub) broadcastLocal(roomID string, msg domain.WSServerMessage, exclude *websocket.Conn) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, ok := h.rooms[roomID]
	if !ok {
		return
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for conn := range room {
		if exclude != nil && conn == exclude {
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("websocket write error: %v", err)
		}
	}
}
