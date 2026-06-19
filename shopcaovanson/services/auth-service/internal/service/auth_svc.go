package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopcaovanson/auth-service/internal/config"
	"github.com/shopcaovanson/auth-service/internal/domain"
	"github.com/shopcaovanson/auth-service/internal/kafka"
	"github.com/shopcaovanson/auth-service/internal/middleware"
	"github.com/shopcaovanson/auth-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidRefresh     = errors.New("invalid refresh token")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrInvalidVerifyToken = errors.New("invalid or expired verification token")
	ErrInvalidResetOTP    = errors.New("invalid or expired otp")
	ErrAccountDisabled    = errors.New("account is disabled")
)

const passwordResetOTPTTL = 10 * time.Minute

type AuthService struct {
	cfg          *config.Config
	userRepo     repository.UserRepository
	refreshRepo  repository.RefreshTokenRepository
	redis        *redis.Client
	kafka        *kafka.Producer
	privateKey   *rsa.PrivateKey
}

func NewAuthService(
	cfg *config.Config,
	userRepo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	redisClient *redis.Client,
	kafkaProducer *kafka.Producer,
	privateKey *rsa.PrivateKey,
) *AuthService {
	return &AuthService{
		cfg:         cfg,
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
		redis:       redisClient,
		kafka:       kafkaProducer,
		privateKey:  privateKey,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, fullName string) (*domain.RegisterResponse, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	fullName = strings.TrimSpace(fullName)

	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if len(password) < 8 {
		return nil, ErrWeakPassword
	}
	if fullName == "" {
		return nil, errors.New("full_name is required")
	}

	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, repository.ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cfg.BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	verifyToken, err := generateVerificationToken()
	if err != nil {
		return nil, err
	}
	expiresAt := now.Add(24 * time.Hour)
	user := &domain.User{
		ID:                    uuid.New(),
		Email:                 email,
		PasswordHash:          string(hash),
		FullName:              fullName,
		Role:                  domain.RoleCustomer,
		IsActive:              true,
		EmailVerified:         false,
		VerificationToken:     &verifyToken,
		VerificationExpiresAt: &expiresAt,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	verifyURL := fmt.Sprintf("%s/auth/verify-email?token=%s", strings.TrimRight(s.cfg.FrontendURL, "/"), verifyToken)
	if err := s.kafka.PublishVerificationRequested(kafka.VerificationRequestedEvent{
		UserID:    user.ID.String(),
		Email:     user.Email,
		FullName:  user.FullName,
		Token:     verifyToken,
		VerifyURL: verifyURL,
	}); err != nil {
		return nil, fmt.Errorf("publish user.verification_requested: %w", err)
	}

	return &domain.RegisterResponse{
		Message:        "Đăng ký thành công. Vui lòng kiểm tra email để xác nhận tài khoản.",
		Email:          user.Email,
		RequiresVerify: true,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.TokenPair, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrAccountDisabled
	}
	if !user.EmailVerified {
		return nil, ErrEmailNotVerified
	}

	return s.issueTokenPair(ctx, user)
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) (*domain.TokenPair, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrInvalidVerifyToken
	}

	user, err := s.userRepo.GetByVerificationToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidVerifyToken
		}
		return nil, err
	}
	if user.EmailVerified {
		return s.issueTokenPair(ctx, user)
	}
	if user.VerificationExpiresAt != nil && time.Now().UTC().After(*user.VerificationExpiresAt) {
		return nil, ErrInvalidVerifyToken
	}

	if err := s.userRepo.MarkEmailVerified(ctx, user.ID); err != nil {
		return nil, err
	}
	user.EmailVerified = true

	_ = s.kafka.PublishUserRegistered(kafka.UserRegisteredEvent{
		UserID:   user.ID.String(),
		Email:    user.Email,
		FullName: user.FullName,
	})

	return s.issueTokenPair(ctx, user)
}

func (s *AuthService) ResendVerification(ctx context.Context, email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil
		}
		return err
	}
	if user.EmailVerified {
		return nil
	}

	verifyToken, err := generateVerificationToken()
	if err != nil {
		return err
	}
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	if err := s.userRepo.SetVerificationToken(ctx, user.ID, verifyToken, expiresAt); err != nil {
		return err
	}

	verifyURL := fmt.Sprintf("%s/auth/verify-email?token=%s", strings.TrimRight(s.cfg.FrontendURL, "/"), verifyToken)
	return s.kafka.PublishVerificationRequested(kafka.VerificationRequestedEvent{
		UserID:    user.ID.String(),
		Email:     user.Email,
		FullName:  user.FullName,
		Token:     verifyToken,
		VerifyURL: verifyURL,
	})
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) (*domain.ForgotPasswordResponse, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	resp := &domain.ForgotPasswordResponse{
		Message: "Nếu email tồn tại, mã OTP đặt lại mật khẩu đã được gửi.",
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return resp, nil
		}
		return nil, err
	}
	if !user.IsActive || !user.EmailVerified {
		return resp, nil
	}

	rateKey := fmt.Sprintf("pwd_reset_rate:%s", email)
	set, err := s.redis.SetNX(ctx, rateKey, "1", time.Minute).Result()
	if err != nil {
		return nil, fmt.Errorf("redis rate limit: %w", err)
	}
	if !set {
		return resp, nil
	}

	otp, err := generateOTP()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().UTC().Add(passwordResetOTPTTL)
	otpHash := hashToken(otp)

	if err := s.userRepo.SetPasswordResetOTP(ctx, user.ID, otpHash, expiresAt); err != nil {
		return nil, err
	}

	resetURL := fmt.Sprintf("%s/auth/reset-password?email=%s", strings.TrimRight(s.cfg.FrontendURL, "/"), url.QueryEscape(email))
	if err := s.kafka.PublishPasswordResetRequested(kafka.PasswordResetRequestedEvent{
		UserID:         user.ID.String(),
		Email:          user.Email,
		FullName:       user.FullName,
		OTP:            otp,
		ResetURL:       resetURL,
		ExpiresMinutes: int(passwordResetOTPTTL.Minutes()),
	}); err != nil {
		return nil, fmt.Errorf("publish user.password_reset_requested: %w", err)
	}

	resp.Email = user.Email
	return resp, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, email, otp, newPassword string) (*domain.ResetPasswordResponse, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	otp = strings.TrimSpace(otp)

	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if len(otp) != 6 {
		return nil, ErrInvalidResetOTP
	}
	if len(newPassword) < 8 {
		return nil, ErrWeakPassword
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidResetOTP
		}
		return nil, err
	}
	if user.PasswordResetOTPHash == nil || user.PasswordResetExpiresAt == nil {
		return nil, ErrInvalidResetOTP
	}
	if time.Now().UTC().After(*user.PasswordResetExpiresAt) {
		return nil, ErrInvalidResetOTP
	}
	if hashToken(otp) != *user.PasswordResetOTPHash {
		return nil, ErrInvalidResetOTP
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.cfg.BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	if err := s.userRepo.UpdatePasswordHash(ctx, user.ID, string(hash)); err != nil {
		return nil, err
	}

	if err := s.refreshRepo.RevokeAllForUser(ctx, user.ID); err != nil {
		return nil, err
	}
	iter := s.redis.Scan(ctx, 0, "refresh_token:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		storedUserID, err := s.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		if storedUserID == user.ID.String() {
			_ = s.redis.Del(ctx, key).Err()
		}
	}

	return &domain.ResetPasswordResponse{
		Message: "Đặt lại mật khẩu thành công. Bạn có thể đăng nhập ngay.",
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, error) {
	tokenHash := hashToken(refreshToken)

	redisKey := fmt.Sprintf("refresh_token:%s", tokenHash)
	userIDStr, err := s.redis.Get(ctx, redisKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrInvalidRefresh
		}
		return "", fmt.Errorf("redis get refresh token: %w", err)
	}

	stored, err := s.refreshRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return "", ErrInvalidRefresh
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil || stored.UserID != userID {
		return "", ErrInvalidRefresh
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	if refreshToken != "" {
		tokenHash := hashToken(refreshToken)
		redisKey := fmt.Sprintf("refresh_token:%s", tokenHash)

		if err := s.redis.Del(ctx, redisKey).Err(); err != nil {
			return fmt.Errorf("redis delete refresh token: %w", err)
		}

		if err := s.refreshRepo.RevokeByTokenHash(ctx, tokenHash); err != nil {
			if errors.Is(err, repository.ErrRefreshTokenNotFound) {
				return nil
			}
			return err
		}
		return nil
	}

	if err := s.refreshRepo.RevokeAllForUser(ctx, userID); err != nil {
		return err
	}

	iter := s.redis.Scan(ctx, 0, "refresh_token:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		storedUserID, err := s.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		if storedUserID == userID.String() {
			_ = s.redis.Del(ctx, key).Err()
		}
	}

	return nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile := user.ToProfile()
	return &profile, nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, fullName string) (*domain.UserProfile, error) {
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return nil, errors.New("full_name is required")
	}

	user, err := s.userRepo.UpdateFullName(ctx, userID, fullName)
	if err != nil {
		return nil, err
	}
	profile := user.ToProfile()
	return &profile, nil
}

func (s *AuthService) issueTokenPair(ctx context.Context, user *domain.User) (*domain.TokenPair, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, tokenHash, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.cfg.RefreshTokenTTL)

	rt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		Revoked:   false,
		CreatedAt: now,
	}

	if err := s.refreshRepo.Create(ctx, rt); err != nil {
		return nil, err
	}

	redisKey := fmt.Sprintf("refresh_token:%s", tokenHash)
	if err := s.redis.Set(ctx, redisKey, user.ID.String(), s.cfg.RefreshTokenTTL).Err(); err != nil {
		return nil, fmt.Errorf("redis set refresh token: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) generateAccessToken(user *domain.User) (string, error) {
	now := time.Now().UTC()
	claims := middleware.Claims{
		Email: user.Email,
		Role:  user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessTokenTTL)),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

func generateRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	raw = hex.EncodeToString(b)
	hash = hashToken(raw)
	return raw, hash, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func validateEmail(email string) error {
	if email == "" || !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return ErrInvalidEmail
	}
	return nil
}

func generateVerificationToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate verification token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func generateOTP() (string, error) {
	var n uint32
	if err := binary.Read(rand.Reader, binary.BigEndian, &n); err != nil {
		return "", fmt.Errorf("generate otp: %w", err)
	}
	return fmt.Sprintf("%06d", n%1000000), nil
}
