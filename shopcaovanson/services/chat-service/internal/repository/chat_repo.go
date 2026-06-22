package repository

import (
	"context"
	"errors"
	"time"

	"github.com/caovanson/shopcaovanson/chat-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	roomsCollection    = "chat_rooms"
	messagesCollection = "chat_messages"
	seedMarkerID       = "seed_marker"
)

type ChatRepository interface {
	FindRoomsByUser(ctx context.Context, userID string) ([]domain.ChatRoom, error)
	FindAllSupportRooms(ctx context.Context) ([]domain.ChatRoom, error)
	FindRoomByID(ctx context.Context, roomID primitive.ObjectID) (*domain.ChatRoom, error)
	FindDirectRoom(ctx context.Context, userA, userB string) (*domain.ChatRoom, error)
	FindSupportRoom(ctx context.Context, userID string) (*domain.ChatRoom, error)
	CreateRoom(ctx context.Context, room *domain.ChatRoom) error
	UpdateParticipants(ctx context.Context, roomID primitive.ObjectID, participants []string) error
	GetMessages(ctx context.Context, roomID primitive.ObjectID, before *primitive.ObjectID, limit int) ([]domain.ChatMessage, error)
	FindMessageByID(ctx context.Context, messageID primitive.ObjectID) (*domain.ChatMessage, error)
	UpdateMessageReactions(ctx context.Context, messageID primitive.ObjectID, reactions map[string][]string) error
	CreateMessage(ctx context.Context, msg *domain.ChatMessage) error
	CountRooms(ctx context.Context) (int64, error)
	CountMessages(ctx context.Context) (int64, error)
	IsSeedComplete(ctx context.Context) (bool, error)
	MarkSeedComplete(ctx context.Context) error
}

type mongoChatRepository struct {
	db *mongo.Database
}

func NewChatRepository(db *mongo.Database) ChatRepository {
	return &mongoChatRepository{db: db}
}

func (r *mongoChatRepository) rooms() *mongo.Collection {
	return r.db.Collection(roomsCollection)
}

func (r *mongoChatRepository) messages() *mongo.Collection {
	return r.db.Collection(messagesCollection)
}

func (r *mongoChatRepository) FindRoomsByUser(ctx context.Context, userID string) ([]domain.ChatRoom, error) {
	filter := bson.M{"participants": userID}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.rooms().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rooms []domain.ChatRoom
	if err := cursor.All(ctx, &rooms); err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *mongoChatRepository) FindRoomByID(ctx context.Context, roomID primitive.ObjectID) (*domain.ChatRoom, error) {
	var room domain.ChatRoom
	err := r.rooms().FindOne(ctx, bson.M{"_id": roomID}).Decode(&room)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *mongoChatRepository) FindDirectRoom(ctx context.Context, userA, userB string) (*domain.ChatRoom, error) {
	filter := bson.M{
		"room_type":    domain.RoomTypeDirect,
		"participants": bson.M{"$all": []string{userA, userB}, "$size": 2},
	}
	var room domain.ChatRoom
	err := r.rooms().FindOne(ctx, filter).Decode(&room)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *mongoChatRepository) FindSupportRoom(ctx context.Context, userID string) (*domain.ChatRoom, error) {
	filter := bson.M{
		"room_type":    domain.RoomTypeSupport,
		"participants": userID,
	}
	var room domain.ChatRoom
	err := r.rooms().FindOne(ctx, filter).Decode(&room)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *mongoChatRepository) FindAllSupportRooms(ctx context.Context) ([]domain.ChatRoom, error) {
	filter := bson.M{"room_type": domain.RoomTypeSupport}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.rooms().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rooms []domain.ChatRoom
	if err := cursor.All(ctx, &rooms); err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *mongoChatRepository) UpdateParticipants(ctx context.Context, roomID primitive.ObjectID, participants []string) error {
	_, err := r.rooms().UpdateOne(ctx, bson.M{"_id": roomID}, bson.M{
		"$set": bson.M{"participants": participants},
	})
	return err
}

func (r *mongoChatRepository) CreateRoom(ctx context.Context, room *domain.ChatRoom) error {
	if room.CreatedAt.IsZero() {
		room.CreatedAt = time.Now().UTC()
	}
	res, err := r.rooms().InsertOne(ctx, room)
	if err != nil {
		return err
	}
	room.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *mongoChatRepository) GetMessages(ctx context.Context, roomID primitive.ObjectID, before *primitive.ObjectID, limit int) ([]domain.ChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	filter := bson.M{"room_id": roomID}
	if before != nil {
		filter["_id"] = bson.M{"$lt": *before}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.messages().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []domain.ChatMessage
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *mongoChatRepository) FindMessageByID(ctx context.Context, messageID primitive.ObjectID) (*domain.ChatMessage, error) {
	var msg domain.ChatMessage
	err := r.messages().FindOne(ctx, bson.M{"_id": messageID}).Decode(&msg)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *mongoChatRepository) UpdateMessageReactions(ctx context.Context, messageID primitive.ObjectID, reactions map[string][]string) error {
	_, err := r.messages().UpdateOne(ctx, bson.M{"_id": messageID}, bson.M{
		"$set": bson.M{"reactions": reactions},
	})
	return err
}

func (r *mongoChatRepository) CreateMessage(ctx context.Context, msg *domain.ChatMessage) error {
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now().UTC()
	}
	if msg.Type == "" {
		msg.Type = domain.MessageTypeText
	}
	if msg.ReadBy == nil {
		msg.ReadBy = []string{}
	}
	res, err := r.messages().InsertOne(ctx, msg)
	if err != nil {
		return err
	}
	msg.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *mongoChatRepository) CountRooms(ctx context.Context) (int64, error) {
	return r.rooms().CountDocuments(ctx, bson.M{})
}

func (r *mongoChatRepository) CountMessages(ctx context.Context) (int64, error) {
	return r.messages().CountDocuments(ctx, bson.M{})
}

func (r *mongoChatRepository) IsSeedComplete(ctx context.Context) (bool, error) {
	count, err := r.rooms().CountDocuments(ctx, bson.M{"_id": seedMarkerID})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *mongoChatRepository) MarkSeedComplete(ctx context.Context) error {
	_, err := r.rooms().InsertOne(ctx, bson.M{
		"_id":          seedMarkerID,
		"participants": []string{},
		"room_type":    "seed",
		"created_at":   time.Now().UTC(),
	})
	return err
}
