package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	"github.com/zenfulcode/zencial/internal/infrastructure/config"
)

// Claims holds JWT token claims.
type Claims struct {
	UserID uuid.UUID
	Role   entity.UserRole
}

// TokenPair holds an access token and refresh token.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// TokenService defines token operations.
type TokenService interface {
	GenerateTokenPair(userID uuid.UUID, role entity.UserRole) (*TokenPair, error)
	ValidateAccessToken(tokenString string) (*Claims, error)
	GenerateRefreshToken() (string, error)
}

type jwtService struct {
	accessSecret    []byte
	accessDuration  time.Duration
	refreshDuration time.Duration
	issuer          string
}

type jwtClaims struct {
	UserID string `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTService creates a JWT-based TokenService.
func NewJWTService(cfg config.JWTConfig) TokenService {
	return &jwtService{
		accessSecret:    []byte(cfg.AccessSecret),
		accessDuration:  cfg.AccessDuration,
		refreshDuration: cfg.RefreshDuration,
		issuer:          cfg.Issuer,
	}
}

func (s *jwtService) GenerateTokenPair(userID uuid.UUID, role entity.UserRole) (*TokenPair, error) {
	expiresAt := time.Now().Add(s.accessDuration)

	claims := jwtClaims{
		UserID: userID.String(),
		Role:   string(role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(s.accessSecret)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	refreshToken, err := s.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *jwtService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.accessSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	return &Claims{
		UserID: userID,
		Role:   entity.UserRole(claims.Role),
	}, nil
}

func (s *jwtService) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
