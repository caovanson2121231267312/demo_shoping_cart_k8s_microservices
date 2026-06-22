package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	RoomTypeDirect  = "direct"
	RoomTypeSupport = "support"

	MessageTypeText  = "text"
	MessageTypeImage = "image"
)

type ChatRoom struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Participants []string           `bson:"participants" json:"participants"`
	RoomType     string             `bson:"room_type" json:"room_type"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}

type ChatMessage struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoomID    primitive.ObjectID `bson:"room_id" json:"room_id"`
	SenderID  string             `bson:"sender_id" json:"sender_id"`
	Content   string             `bson:"content" json:"content"`
	Type      string             `bson:"type" json:"type"`
	Reactions map[string][]string `bson:"reactions,omitempty" json:"reactions,omitempty"`
	ReadBy    []string           `bson:"read_by" json:"read_by"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type CreateRoomRequest struct {
	ParticipantID string `json:"participant_id"`
}

type AdminSupportRoomRequest struct {
	CustomerID string `json:"customer_id"`
}

type SupportRoomSummary struct {
	ChatRoom
	CustomerID string `json:"customer_id"`
}

type WSClientMessage struct {
	Type      string `json:"type"`
	RoomID    string `json:"room_id"`
	Content   string `json:"content,omitempty"`
	MsgType   string `json:"msg_type,omitempty"`
	MessageID string `json:"message_id,omitempty"`
	Emoji     string `json:"emoji,omitempty"`
}

type WSServerMessage struct {
	Type      string              `json:"type"`
	RoomID    string              `json:"room_id,omitempty"`
	MessageID string              `json:"message_id,omitempty"`
	SenderID  string              `json:"sender_id,omitempty"`
	UserID    string              `json:"user_id,omitempty"`
	Content   string              `json:"content,omitempty"`
	MsgType   string              `json:"msg_type,omitempty"`
	Emoji     string              `json:"emoji,omitempty"`
	Reactions map[string][]string `json:"reactions,omitempty"`
	CreatedAt string              `json:"created_at,omitempty"`
}
