package main

import (
	"context"
	"log"
	"time"

	"github.com/caovanson/shopcaovanson/chat-service/internal/config"
	"github.com/caovanson/shopcaovanson/chat-service/internal/handler"
	"github.com/caovanson/shopcaovanson/chat-service/internal/hub"
	"github.com/caovanson/shopcaovanson/chat-service/internal/middleware"
	"github.com/caovanson/shopcaovanson/chat-service/internal/repository"
	"github.com/caovanson/shopcaovanson/chat-service/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("mongo ping: %v", err)
	}

	db := mongoClient.Database(cfg.MongoDB)
	ensureIndexes(ctx, db)

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis url: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis ping: %v", err)
	}

	repo := repository.NewChatRepository(db)
	chatSvc := service.NewChatService(repo)
	chatHub := hub.NewHub(chatSvc, redisClient)
	chatHandler := handler.NewChatHandler(chatSvc, chatHub, cfg.JWTPublicKey)
	uploadHandler := handler.NewUploadHandler(cfg)

	app := fiber.New(fiber.Config{AppName: "chat-service"})
	app.Use(recover.New())
	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "chat-service"})
	})

	api := app.Group("/api/chat", middleware.JWTMiddleware(cfg.JWTPublicKey))
	api.Get("/rooms", chatHandler.ListRooms)
	api.Post("/rooms", chatHandler.CreateRoom)
	api.Post("/rooms/support", chatHandler.CreateSupportRoom)
	api.Post("/upload", uploadHandler.Upload)
	api.Get("/media/:filename", uploadHandler.ServeMedia)
	api.Get("/rooms/:id/messages", chatHandler.GetMessages)

	admin := api.Group("/admin", middleware.RequireStaff())
	admin.Get("/support-rooms", chatHandler.ListSupportRooms)
	admin.Post("/support-rooms", chatHandler.CreateSupportRoomForCustomer)

	app.Use("/ws", chatHandler.WebSocketUpgrade)
	app.Get("/ws", websocket.New(chatHandler.WebSocket))

	log.Printf("chat-service listening on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func ensureIndexes(ctx context.Context, db *mongo.Database) {
	rooms := db.Collection("chat_rooms")
	_, _ = rooms.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "participants", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	})

	messages := db.Collection("chat_messages")
	_, _ = messages.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "room_id", Value: 1}, {Key: "created_at", Value: -1}}},
	})
}
