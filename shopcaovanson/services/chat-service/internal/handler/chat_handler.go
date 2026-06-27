package handler

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/caovanson/shopcaovanson/chat-service/internal/domain"
	"github.com/caovanson/shopcaovanson/chat-service/internal/hub"
	"github.com/caovanson/shopcaovanson/chat-service/internal/middleware"
	"github.com/caovanson/shopcaovanson/chat-service/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type ChatHandler struct {
	svc       *service.ChatService
	hub       *hub.Hub
	jwtPubKey string
}

func NewChatHandler(svc *service.ChatService, h *hub.Hub, jwtPubKey string) *ChatHandler {
	return &ChatHandler{svc: svc, hub: h, jwtPubKey: jwtPubKey}
}

func (h *ChatHandler) ListRooms(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	rooms, err := h.svc.ListRooms(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": rooms})
}

func (h *ChatHandler) ListSupportRooms(c *fiber.Ctx) error {
	rooms, err := h.svc.ListSupportRoomsForStaff(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": rooms})
}

func (h *ChatHandler) CreateRoom(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req domain.CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	room, err := h.svc.CreateRoom(c.Context(), userID, req.ParticipantID)
	if err != nil {
		switch err {
		case service.ErrInvalidParticipant:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": room})
}

func (h *ChatHandler) CreateSupportRoom(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	room, err := h.svc.GetOrCreateSupportRoom(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": room})
}

func (h *ChatHandler) CreateSupportRoomForCustomer(c *fiber.Ctx) error {
	var req domain.AdminSupportRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	room, err := h.svc.GetOrCreateSupportRoomForCustomer(c.Context(), req.CustomerID)
	if err != nil {
		switch err {
		case service.ErrInvalidParticipant:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}
	return c.JSON(fiber.Map{"data": domain.SupportRoomSummary{
		ChatRoom:   *room,
		CustomerID: req.CustomerID,
	}})
}

func (h *ChatHandler) GetMessages(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	roomID := c.Params("id")
	before := c.Query("before")
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	messages, err := h.svc.GetMessages(c.Context(), userID, userRole, roomID, before, limit)
	if err != nil {
		switch err {
		case service.ErrRoomNotFound, service.ErrNotParticipant:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}
	return c.JSON(fiber.Map{"data": messages})
}

func (h *ChatHandler) WebSocketUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func (h *ChatHandler) WebSocket(c *websocket.Conn) {
	token := c.Query("token")
	claims, err := middleware.ValidateTokenString(token, h.jwtPubKey)
	if err != nil {
		_ = c.WriteJSON(fiber.Map{"error": "unauthorized"})
		_ = c.Close()
		return
	}
	userID := claims.Subject
	userRole := claims.Role
	log.Printf("[chat-ws] connected user=%s role=%s", userID, userRole)

	var activeRoom string

	defer func() {
		log.Printf("[chat-ws] disconnected user=%s room=%s", userID, activeRoom)
		if activeRoom != "" {
			h.hub.Unregister(c, activeRoom)
		}
		_ = c.Close()
	}()

	for {
		mt, raw, err := c.ReadMessage()
		if err != nil {
			log.Printf("[chat-ws] read closed user=%s room=%s err=%v", userID, activeRoom, err)
			break
		}
		if mt != websocket.TextMessage && mt != websocket.BinaryMessage {
			log.Printf("[chat-ws] skip frame user=%s type=%d", userID, mt)
			continue
		}

		var msg domain.WSClientMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("[chat-ws] bad json user=%s raw=%q err=%v", userID, string(raw), err)
			continue
		}
		log.Printf("[chat-ws] recv user=%s type=%s room=%s", userID, msg.Type, msg.RoomID)

		switch msg.Type {
		case "join":
			if activeRoom != "" {
				h.hub.Unregister(c, activeRoom)
			}
			if err := h.svc.ValidateRoomAccess(context.Background(), userID, userRole, msg.RoomID); err != nil {
				log.Printf("[chat-ws] join denied user=%s room=%s err=%v", userID, msg.RoomID, err)
				_ = c.WriteJSON(fiber.Map{"error": err.Error()})
				continue
			}
			activeRoom = msg.RoomID
			h.hub.Register(c, userID, activeRoom)
			log.Printf("[chat-ws] joined user=%s room=%s listeners=%d", userID, activeRoom, h.hub.RoomListenerCount(activeRoom))
			_ = c.WriteJSON(domain.WSServerMessage{
				Type:   "joined",
				RoomID: activeRoom,
				UserID: userID,
			})
		default:
			preview := strings.TrimSpace(msg.Content)
			if len(preview) > 40 {
				preview = preview[:40] + "..."
			}
			log.Printf("[chat-ws] %s user=%s room=%s content=%q", msg.Type, userID, msg.RoomID, preview)
			if err := h.hub.HandleMessage(context.Background(), userID, userRole, msg); err != nil {
				log.Printf("[chat-ws] %s failed user=%s room=%s err=%v", msg.Type, userID, msg.RoomID, err)
				_ = c.WriteJSON(fiber.Map{"error": err.Error()})
			}
		}
	}
}
