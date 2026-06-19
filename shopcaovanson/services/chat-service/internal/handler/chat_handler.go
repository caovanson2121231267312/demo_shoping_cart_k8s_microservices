package handler

import (
	"context"
	"strconv"

	"github.com/caovanson/shopcaovanson/chat-service/internal/domain"
	"github.com/caovanson/shopcaovanson/chat-service/internal/hub"
	"github.com/caovanson/shopcaovanson/chat-service/internal/middleware"
	"github.com/caovanson/shopcaovanson/chat-service/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type ChatHandler struct {
	svc        *service.ChatService
	hub        *hub.Hub
	jwtPubKey  string
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

func (h *ChatHandler) GetMessages(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	roomID := c.Params("id")
	before := c.Query("before")
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	messages, err := h.svc.GetMessages(c.Context(), userID, roomID, before, limit)
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

	var activeRoom string

	defer func() {
		if activeRoom != "" {
			h.hub.Unregister(c, activeRoom)
		}
		_ = c.Close()
	}()

	for {
		var msg domain.WSClientMessage
		if err := c.ReadJSON(&msg); err != nil {
			break
		}

		switch msg.Type {
		case "join":
			if activeRoom != "" {
				h.hub.Unregister(c, activeRoom)
			}
			if err := h.svc.ValidateRoomAccess(context.Background(), userID, msg.RoomID); err != nil {
				_ = c.WriteJSON(fiber.Map{"error": err.Error()})
				continue
			}
			activeRoom = msg.RoomID
			h.hub.Register(c, userID, activeRoom)
		default:
			if err := h.hub.HandleMessage(context.Background(), userID, msg); err != nil {
				_ = c.WriteJSON(fiber.Map{"error": err.Error()})
			}
		}
	}
}
