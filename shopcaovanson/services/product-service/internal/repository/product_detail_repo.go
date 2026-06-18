package repository

import (
	"context"
	"time"

	"github.com/caovanson/shopcaovanson/product-service/internal/domain"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProductDetailRepository interface {
	GetByProductID(ctx context.Context, productID uuid.UUID) (*domain.ProductDetail, error)
	Upsert(ctx context.Context, detail *domain.ProductDetail) error
	DeleteByProductID(ctx context.Context, productID uuid.UUID) error
}

type productDetailRepo struct {
	collection *mongo.Collection
}

func NewProductDetailRepository(db *mongo.Database) ProductDetailRepository {
	return &productDetailRepo{collection: db.Collection("product_details")}
}

func (r *productDetailRepo) GetByProductID(ctx context.Context, productID uuid.UUID) (*domain.ProductDetail, error) {
	var doc struct {
		Specifications  map[string]string `bson:"specifications"`
		RichDescription string            `bson:"rich_description"`
		UpdatedAt       time.Time         `bson:"updated_at"`
	}
	err := r.collection.FindOne(ctx, bson.M{"_id": productID.String()}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &domain.ProductDetail{
		ProductID:       productID,
		Specifications:  doc.Specifications,
		RichDescription: doc.RichDescription,
		UpdatedAt:       doc.UpdatedAt,
	}, nil
}

func (r *productDetailRepo) Upsert(ctx context.Context, detail *domain.ProductDetail) error {
	detail.UpdatedAt = time.Now().UTC()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": detail.ProductID.String()},
		bson.M{"$set": bson.M{
			"specifications":   detail.Specifications,
			"rich_description": detail.RichDescription,
			"updated_at":       detail.UpdatedAt,
		}},
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *productDetailRepo) DeleteByProductID(ctx context.Context, productID uuid.UUID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": productID.String()})
	return err
}
