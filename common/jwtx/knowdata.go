package jwtx

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var keyManager *KeyManager

type Claims struct {
	ClientId  string `json:"clientId"`
	UserId    int64  `json:"userId"`
	EmpCode   string `json:"empCode"`
	Roles     string `json:"roles"` // 0:普通用户 1:普通管理员 2:超级管理员
	CompanyId int64  `json:"companyId"`
	UserName  string `json:"userName"`

	IsAdmin int64 `json:"isAdmin"`

	jwt.RegisteredClaims
}

// GenerateTokenWithContext 使用context生成JWT token
func GenerateTokenWithContext(ctx context.Context, clientId string, userId int64, empCode string, isAdmin int64, userName string, roles string, expireDuration time.Duration) (string, error) {
	claims := Claims{
		ClientId: clientId,
		UserId:   userId,
		EmpCode:  empCode,
		Roles:    roles,
		IsAdmin:  isAdmin, // 0:普通用户 1:普通管理员 2:超级管理员

		UserName: userName,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	key, err := getSecretKey(ctx)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}

// ParseToken 解析JWT token
func ParseTokenKnowdata(tokenString string) (*Claims, error) {
	return ParseTokenKnowdataWithContext(context.Background(), tokenString)
}

// ParseTokenKnowdataWithContext 使用context解析JWT token
func ParseTokenKnowdataWithContext(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		key, err := getSecretKey(ctx)
		if err != nil {
			return nil, err
		}
		return key, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// SetKeyManager 设置密钥管理器（启动时必须调用）
func SetKeyManager(km *KeyManager) {
	keyManager = km
}

// getSecretKey 仅从 KeyManager 获取 JWT 签名密钥
func getSecretKey(ctx context.Context) ([]byte, error) {
	if keyManager == nil {
		return nil, errors.New("JWT KeyManager 未初始化")
	}
	key, err := keyManager.GetOrCreateSecretKey(ctx)
	if err != nil {
		return nil, err
	}
	return []byte(key), nil
}
