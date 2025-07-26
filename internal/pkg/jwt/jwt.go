package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims 自定义JWT声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWT JWT工具类
type JWT struct {
	secret []byte
}

// NewJwt 创建JWT实例
func NewJwt(secret string) *JWT {
	return &JWT{
		secret: []byte(secret),
	}
}

// GenerateToken 生成JWT token
func (j *JWT) GenerateToken(userID uint, username string, expiration time.Duration) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ParseToken 解析JWT token
func (j *JWT) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken 刷新token
func (j *JWT) RefreshToken(tokenString string, expiration time.Duration) (string, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 更新过期时间
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(expiration))
	claims.IssuedAt = jwt.NewNumericDate(time.Now())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}
