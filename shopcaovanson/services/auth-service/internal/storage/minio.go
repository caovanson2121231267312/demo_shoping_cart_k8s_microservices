package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/shopcaovanson/auth-service/internal/config"
)

var (
	ErrStorageDisabled = fmt.Errorf("object storage is not configured")
	ErrInvalidImage    = fmt.Errorf("invalid image file")
)

type Object struct {
	Reader      io.ReadCloser
	Size        int64
	ContentType string
}

type AvatarStorage struct {
	client *minio.Client
	bucket string
	region string
}

func NewAvatarStorage(cfg *config.Config) (*AvatarStorage, error) {
	if !cfg.MinIOEnabled {
		return nil, nil
	}
	if cfg.MinIOEndpoint == "" || cfg.MinIOAccessKey == "" || cfg.MinIOSecretKey == "" || cfg.MinIOBucket == "" {
		return nil, fmt.Errorf("minio configuration is incomplete")
	}

	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
		Region: cfg.MinIORegion,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.MinIOBucket)
	if err != nil {
		return nil, fmt.Errorf("check minio bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.MinIOBucket, minio.MakeBucketOptions{Region: cfg.MinIORegion}); err != nil {
			return nil, fmt.Errorf("create minio bucket: %w", err)
		}
	}

	return &AvatarStorage{
		client: client,
		bucket: cfg.MinIOBucket,
		region: cfg.MinIORegion,
	}, nil
}

func (s *AvatarStorage) Enabled() bool {
	return s != nil
}

func AvatarObjectKey(userID uuid.UUID, ext string) string {
	return fmt.Sprintf("avatars/%s/avatar%s", userID.String(), ext)
}

func (s *AvatarStorage) Upload(ctx context.Context, userID uuid.UUID, reader io.Reader, size int64, contentType, ext string) (string, error) {
	if s == nil {
		return "", ErrStorageDisabled
	}
	key := AvatarObjectKey(userID, ext)
	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"x-amz-meta-owner": userID.String(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("upload avatar: %w", err)
	}
	return key, nil
}

func (s *AvatarStorage) Delete(ctx context.Context, key string) error {
	if s == nil {
		return ErrStorageDisabled
	}
	if strings.TrimSpace(key) == "" {
		return nil
	}
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

func (s *AvatarStorage) Open(ctx context.Context, key string) (*Object, error) {
	if s == nil {
		return nil, ErrStorageDisabled
	}
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("open avatar: %w", err)
	}
	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, fmt.Errorf("stat avatar: %w", err)
	}
	contentType := info.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return &Object{
		Reader:      obj,
		Size:        info.Size,
		ContentType: contentType,
	}, nil
}

func DetectImageExt(filename string, contentType string) (ext string, mime string, err error) {
	ext = strings.ToLower(filepath.Ext(filename))
	allowed := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".webp": "image/webp",
		".gif":  "image/gif",
	}
	if mime, ok := allowed[ext]; ok {
		return ext, mime, nil
	}
	ct := strings.ToLower(strings.TrimSpace(contentType))
	for e, m := range allowed {
		if ct == m {
			return e, m, nil
		}
	}
	return "", "", ErrInvalidImage
}
