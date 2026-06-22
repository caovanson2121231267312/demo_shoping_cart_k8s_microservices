package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/caovanson/shopcaovanson/chat-service/internal/domain"
	"github.com/caovanson/shopcaovanson/chat-service/internal/rasa"
	"github.com/caovanson/shopcaovanson/chat-service/internal/repository"
	"github.com/caovanson/shopcaovanson/chat-service/internal/staff"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrRoomNotFound       = errors.New("room not found")
	ErrNotParticipant     = errors.New("not a room participant")
	ErrInvalidParticipant = errors.New("invalid participant")
	ErrStaffOnly          = errors.New("staff access required")
)

type ChatService struct {
	repo        repository.ChatRepository
	staffUserID string
}

func NewChatService(repo repository.ChatRepository) *ChatService {
	return &ChatService{
		repo:        repo,
		staffUserID: staff.SupportUserID(),
	}
}

func (s *ChatService) StaffUserID() string {
	return s.staffUserID
}

func (s *ChatService) ListRooms(ctx context.Context, userID string) ([]domain.ChatRoom, error) {
	return s.repo.FindRoomsByUser(ctx, userID)
}

func (s *ChatService) ListSupportRoomsForStaff(ctx context.Context) ([]domain.SupportRoomSummary, error) {
	rooms, err := s.repo.FindAllSupportRooms(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.SupportRoomSummary, 0, len(rooms))
	for i := range rooms {
		room := rooms[i]
		fixed, err := s.ensureStaffSupportRoom(ctx, &room)
		if err != nil {
			return nil, err
		}
		out = append(out, domain.SupportRoomSummary{
			ChatRoom:   *fixed,
			CustomerID: customerIDFromRoom(fixed, s.staffUserID),
		})
	}
	return out, nil
}

func (s *ChatService) CreateRoom(ctx context.Context, userID, participantID string) (*domain.ChatRoom, error) {
	if participantID == "" || participantID == userID {
		return nil, ErrInvalidParticipant
	}

	existing, err := s.repo.FindDirectRoom(ctx, userID, participantID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	room := &domain.ChatRoom{
		Participants: []string{userID, participantID},
		RoomType:     domain.RoomTypeDirect,
	}
	if err := s.repo.CreateRoom(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *ChatService) GetOrCreateSupportRoom(ctx context.Context, userID string) (*domain.ChatRoom, error) {
	existing, err := s.repo.FindSupportRoom(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return s.ensureStaffSupportRoom(ctx, existing)
	}

	room := &domain.ChatRoom{
		Participants: []string{userID, s.staffUserID},
		RoomType:     domain.RoomTypeSupport,
	}
	if err := s.repo.CreateRoom(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *ChatService) GetOrCreateSupportRoomForCustomer(ctx context.Context, customerID string) (*domain.ChatRoom, error) {
	if customerID == "" || customerID == s.staffUserID {
		return nil, ErrInvalidParticipant
	}
	existing, err := s.repo.FindSupportRoom(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return s.ensureStaffSupportRoom(ctx, existing)
	}

	room := &domain.ChatRoom{
		Participants: []string{customerID, s.staffUserID},
		RoomType:     domain.RoomTypeSupport,
	}
	if err := s.repo.CreateRoom(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *ChatService) ensureStaffSupportRoom(ctx context.Context, room *domain.ChatRoom) (*domain.ChatRoom, error) {
	participants := make([]string, 0, len(room.Participants)+1)
	seen := make(map[string]bool)
	changed := false

	for _, p := range room.Participants {
		if p == rasa.BotUserID {
			changed = true
			continue
		}
		if seen[p] {
			changed = true
			continue
		}
		seen[p] = true
		participants = append(participants, p)
	}

	if !seen[s.staffUserID] {
		participants = append(participants, s.staffUserID)
		changed = true
	}

	if changed {
		if err := s.repo.UpdateParticipants(ctx, room.ID, participants); err != nil {
			return nil, err
		}
		room.Participants = participants
	}
	return room, nil
}

func customerIDFromRoom(room *domain.ChatRoom, staffUserID string) string {
	for _, p := range room.Participants {
		if p != staffUserID && p != rasa.BotUserID {
			return p
		}
	}
	return ""
}

func (s *ChatService) GetMessages(ctx context.Context, userID, userRole, roomIDStr string, beforeCursor string, limit int) ([]domain.ChatMessage, error) {
	if err := s.ValidateRoomAccess(ctx, userID, userRole, roomIDStr); err != nil {
		return nil, err
	}

	roomID, err := primitive.ObjectIDFromHex(roomIDStr)
	if err != nil {
		return nil, ErrRoomNotFound
	}

	var before *primitive.ObjectID
	if beforeCursor != "" {
		id, err := primitive.ObjectIDFromHex(beforeCursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor")
		}
		before = &id
	}

	return s.repo.GetMessages(ctx, roomID, before, limit)
}

func (s *ChatService) SaveMessage(ctx context.Context, userID, roomIDStr, content, msgType string) (*domain.ChatMessage, error) {
	if msgType == "" {
		msgType = domain.MessageTypeText
	}
	if msgType != domain.MessageTypeText && msgType != domain.MessageTypeImage {
		return nil, fmt.Errorf("unsupported message type")
	}
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("empty content")
	}
	return s.saveMessageAs(ctx, userID, roomIDStr, content, msgType)
}

func (s *ChatService) SaveBotMessage(ctx context.Context, roomIDStr, content string) (*domain.ChatMessage, error) {
	return s.saveMessageAs(ctx, rasa.BotUserID, roomIDStr, content, domain.MessageTypeText)
}

func (s *ChatService) saveMessageAs(ctx context.Context, senderID, roomIDStr, content, msgType string) (*domain.ChatMessage, error) {
	roomID, err := primitive.ObjectIDFromHex(roomIDStr)
	if err != nil {
		return nil, ErrRoomNotFound
	}

	room, err := s.repo.FindRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, ErrRoomNotFound
	}
	if !contains(room.Participants, senderID) {
		return nil, ErrNotParticipant
	}

	readBy := []string{senderID}
	msg := &domain.ChatMessage{
		RoomID:   roomID,
		SenderID: senderID,
		Content:  content,
		Type:     msgType,
		ReadBy:   readBy,
	}
	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *ChatService) ToggleReaction(ctx context.Context, userID, userRole, roomIDStr, messageIDStr, emoji string) (*domain.ChatMessage, error) {
	emoji = strings.TrimSpace(emoji)
	if emoji == "" {
		return nil, fmt.Errorf("emoji required")
	}
	if err := s.ValidateRoomAccess(ctx, userID, userRole, roomIDStr); err != nil {
		return nil, err
	}

	messageID, err := primitive.ObjectIDFromHex(messageIDStr)
	if err != nil {
		return nil, ErrRoomNotFound
	}
	roomID, err := primitive.ObjectIDFromHex(roomIDStr)
	if err != nil {
		return nil, ErrRoomNotFound
	}

	msg, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if msg == nil || msg.RoomID != roomID {
		return nil, ErrRoomNotFound
	}

	reactions := msg.Reactions
	if reactions == nil {
		reactions = make(map[string][]string)
	}
	users := reactions[emoji]
	found := false
	next := make([]string, 0, len(users))
	for _, uid := range users {
		if uid == userID {
			found = true
			continue
		}
		next = append(next, uid)
	}
	if !found {
		next = append(next, userID)
	}
	if len(next) == 0 {
		delete(reactions, emoji)
	} else {
		reactions[emoji] = next
	}

	if err := s.repo.UpdateMessageReactions(ctx, messageID, reactions); err != nil {
		return nil, err
	}
	msg.Reactions = reactions
	return msg, nil
}

func (s *ChatService) ValidateRoomAccess(ctx context.Context, userID, userRole, roomIDStr string) error {
	roomID, err := primitive.ObjectIDFromHex(roomIDStr)
	if err != nil {
		return ErrRoomNotFound
	}
	room, err := s.repo.FindRoomByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return ErrRoomNotFound
	}

	if staff.IsStaffRole(userRole) && room.RoomType == domain.RoomTypeSupport {
		if !contains(room.Participants, userID) {
			participants := append(append([]string{}, room.Participants...), userID)
			if err := s.repo.UpdateParticipants(ctx, room.ID, participants); err != nil {
				return err
			}
		}
		return nil
	}

	if !contains(room.Participants, userID) {
		return ErrNotParticipant
	}
	return nil
}

func (s *ChatService) Repo() repository.ChatRepository {
	return s.repo
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
