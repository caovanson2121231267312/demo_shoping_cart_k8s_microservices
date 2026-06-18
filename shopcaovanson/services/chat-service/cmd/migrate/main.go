package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caovanson/shopcaovanson/chat-service/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run cmd/migrate/main.go [up|down]")
	}

	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database(cfg.MongoDB)

	switch os.Args[1] {
	case "up":
		if err := migrateUp(ctx, db); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
		fmt.Println("migration up completed")
	case "down":
		if err := migrateDown(ctx, db); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
		fmt.Println("migration down completed")
	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}
}

func migrateUp(ctx context.Context, db *mongo.Database) error {
	rooms := db.Collection("chat_rooms")
	_, err := rooms.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "participants", Value: 1}}},
		{Keys: bson.D{{Key: "room_type", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	})
	if err != nil {
		return err
	}

	messages := db.Collection("chat_messages")
	_, err = messages.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "room_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "sender_id", Value: 1}}},
	})
	return err
}

func migrateDown(ctx context.Context, db *mongo.Database) error {
	rooms := db.Collection("chat_rooms")
	for _, name := range []string{"participants_1", "room_type_1", "created_at_-1"} {
		_, _ = rooms.Indexes().DropOne(ctx, name)
	}

	messages := db.Collection("chat_messages")
	for _, name := range []string{"room_id_1_created_at_-1", "sender_id_1"} {
		_, _ = messages.Indexes().DropOne(ctx, name)
	}
	return nil
}
