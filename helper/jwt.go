package helper

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"tiket-in-backend/config"
)

type JWTManager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

type JWTClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTManager() *JWTManager {
	ttlMinutes := config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 60)
	return &JWTManager{
		secret: []byte(config.GetEnv("JWT_SECRET", "default-secret-change-me")),
		issuer: config.GetEnv("JWT_ISSUER", "tiket-in-backend"),
		ttl:    time.Duration(ttlMinutes) * time.Minute,
	}
}

func (m *JWTManager) GenerateToken(userID int64, role string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) VerifyToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode signing tidak didukung: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("token tidak valid")
	}
	return claims, nil
}
