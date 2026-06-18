//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caovanson/shopcaovanson/chat-service/internal/config"
	"github.com/caovanson/shopcaovanson/chat-service/internal/domain"
	"github.com/caovanson/shopcaovanson/chat-service/internal/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var sampleMessages = []string{
	"Xin chào, tôi cần hỗ trợ đơn hàng.",
	"Bạn có thể kiểm tra giúp tôi không?",
	"Đơn hàng của tôi đã giao chưa?",
	"Cảm ơn bạn đã hỗ trợ!",
	"Sản phẩm này còn hàng không?",
	"Tôi muốn đổi size áo.",
	"Phí ship về Hà Nội bao nhiêu?",
	"Khi nào có khuyến mãi tiếp theo?",
	"Tôi nhận được hàng rồi, cảm ơn shop.",
	"Có thể hủy đơn hàng không?",
	"Shop check giúp mã vận đơn nhé.",
	"Tôi muốn mua thêm 2 sản phẩm.",
	"Áo này có màu trắng không?",
	"Đã thanh toán COD, chờ giao hàng.",
	"Bạn tư vấn giúp mình size nhé.",
}

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database(cfg.MongoDB)
	repo := repository.NewChatRepository(db)

	done, err := repo.IsSeedComplete(ctx)
	if err != nil {
		log.Fatalf("seed check: %v", err)
	}
	if done {
		fmt.Println("chat fake data already seeded, skipping")
		os.Exit(0)
	}

	adminID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("admin@shop.com")).String()
	userIDs := make([]string, 0, 10)
	for i := 1; i <= 10; i++ {
		userIDs = append(userIDs, uuid.NewSHA1(uuid.NameSpaceURL, []byte(fmt.Sprintf("customer-%d@shop.com", i))).String())
	}

	rooms := make([]domain.ChatRoom, 0, 20)
	roomDocs := make([]interface{}, 0, 20)
	baseTime := time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)

	for i := 0; i < 20; i++ {
		participant := userIDs[i%len(userIDs)]
		roomType := domain.RoomTypeSupport
		if i%3 == 0 {
			roomType = domain.RoomTypeDirect
		}
		roomID := primitive.NewObjectIDFromTimestamp(baseTime.Add(time.Duration(i) * time.Hour))
		room := domain.ChatRoom{
			ID:           roomID,
			Participants: []string{participant, adminID},
			RoomType:     roomType,
			CreatedAt:    baseTime.Add(time.Duration(i) * time.Hour),
		}
		rooms = append(rooms, room)
		roomDocs = append(roomDocs, room)
	}

	if _, err := db.Collection("chat_rooms").InsertMany(ctx, roomDocs); err != nil {
		log.Fatalf("insert rooms: %v", err)
	}

	messageDocs := make([]interface{}, 0, 400)
	msgIndex := 0
	for _, room := range rooms {
		msgCount := 20
		for j := 0; j < msgCount; j++ {
			sender := room.Participants[j%2]
			msg := domain.ChatMessage{
				RoomID:    room.ID,
				SenderID:  sender,
				Content:   sampleMessages[msgIndex%len(sampleMessages)],
				Type:      domain.MessageTypeText,
				ReadBy:    []string{sender},
				CreatedAt: room.CreatedAt.Add(time.Duration(j+1) * time.Minute),
			}
			messageDocs = append(messageDocs, msg)
			msgIndex++
		}
	}

	if _, err := db.Collection("chat_messages").InsertMany(ctx, messageDocs); err != nil {
		log.Fatalf("insert messages: %v", err)
	}

	if err := repo.MarkSeedComplete(ctx); err != nil {
		log.Fatalf("mark seed: %v", err)
	}

	fmt.Printf("seeded %d rooms and %d messages\n", len(rooms), len(messageDocs))
}
