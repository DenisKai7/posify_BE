package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"posify-backend/internal/model"
	"posify-backend/internal/repository"
)

type AuthService struct {
	repo *repository.AuthRepo
}

func NewAuthService(repo *repository.AuthRepo) *AuthService {
	return &AuthService{repo: repo}
}

type RegisterRequest struct {
	TenantName string `json:"tenant_name"`
	OwnerName  string `json:"owner_name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type LoginRequest struct {
	TenantName string `json:"tenant_name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	tenant := &model.Tenant{Name: req.TenantName}
	if err := s.repo.CreateTenant(ctx, tenant); err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	user := &model.User{
		TenantID:     tenant.ID,
		FullName:     req.OwnerName,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         "OWNER",
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	token, err := generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	tenant, err := s.repo.GetTenantByName(ctx, req.TenantName)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	user, err := s.repo.GetUserByEmail(ctx, tenant.ID, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, err := generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

func generateToken(user *model.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "posify-dev-secret" // ponytail: replace with env-only in production
	}

	claims := jwt.MapClaims{
		"sub":       user.ID,
		"tenant_id": user.TenantID,
		"role":      user.Role,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
