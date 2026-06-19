package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/caovanson/shopcaovanson/chat-service/internal/domain"
	"github.com/caovanson/shopcaovanson/chat-service/internal/rasa"
	"github.com/caovanson/shopcaovanson/chat-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrRoomNotFound       = errors.New("room not found")
	ErrNotParticipant     = errors.New("not a room participant")
	ErrInvalidParticipant = errors.New("invalid participant")
)

type ChatService struct {
	repo repository.ChatRepository
}

func NewChatService(repo repository.ChatRepository) *ChatService {
	return &ChatService{repo: repo}
}

func (s *ChatService) ListRooms(ctx context.Context, userID string) ([]domain.ChatRoom, error) {
	return s.repo.FindRoomsByUser(ctx, userID)
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
		return existing, nil
	}

	room := &domain.ChatRoom{
		Participants: []string{userID, rasa.BotUserID},
		RoomType:     domain.RoomTypeSupport,
	}
	if err := s.repo.CreateRoom(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *ChatService) GetMessages(ctx context.Context, userID, roomIDStr string, beforeCursor string, limit int) ([]domain.ChatMessage, error) {
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
	if !contains(room.Participants, userID) {
		return nil, ErrNotParticipant
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

func (s *ChatService) SaveMessage(ctx context.Context, userID, roomIDStr, content string) (*domain.ChatMessage, error) {
	return s.saveMessageAs(ctx, userID, roomIDStr, content)
}

func (s *ChatService) SaveBotMessage(ctx context.Context, roomIDStr, content string) (*domain.ChatMessage, error) {
	return s.saveMessageAs(ctx, rasa.BotUserID, roomIDStr, content)
}

func (s *ChatService) saveMessageAs(ctx context.Context, senderID, roomIDStr, content string) (*domain.ChatMessage, error) {
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
		Type:     domain.MessageTypeText,
		ReadBy:   readBy,
	}
	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *ChatService) ValidateRoomAccess(ctx context.Context, userID, roomIDStr string) error {
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
