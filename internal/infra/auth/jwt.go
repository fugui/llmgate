package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"modelgate/internal/repository"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims JWT 声明
type Claims struct {
	UserID uuid.UUID   `json:"user_id"`
	Email  string      `json:"email"`
	Name   string      `json:"name"`
	Role   entity.Role `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager JWT 管理器
type JWTManager struct {
	secret      []byte
	expireHours int
}

func NewJWTManager(secret string, expireHours int) *JWTManager {
	return &JWTManager{
		secret:      []byte(secret),
		expireHours: expireHours,
	}
}

// Generate 生成 JWT Token
func (m *JWTManager) Generate(user *entity.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(m.expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "modelgate",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Validate 验证 JWT Token
func (m *JWTManager) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// ShouldRefresh 判断 Token 是否需要刷新（例如：已过去总有效时间的 75%，即剩余有效时间少于等于 25%）
func (m *JWTManager) ShouldRefresh(claims *Claims) bool {
	if claims == nil || claims.ExpiresAt == nil {
		return false
	}
	expiresAt := claims.ExpiresAt.Time
	totalDuration := time.Duration(m.expireHours) * time.Hour
	remaining := time.Until(expiresAt)
	// 如果剩余时间少于等于总时长的 25%（且 Token 尚未过期），则需要刷新
	return remaining > 0 && remaining <= totalDuration/4
}

// RefreshToken 根据旧的 Claims 生成新的 Token
func (m *JWTManager) RefreshToken(claims *Claims) (string, error) {
	user := &entity.User{
		ID:    claims.UserID,
		Email: claims.Email,
		Name:  claims.Name,
		Role:  claims.Role,
	}
	return m.Generate(user)
}

