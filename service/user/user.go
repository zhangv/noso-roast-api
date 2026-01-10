package user

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"hahacking/config"
	models "hahacking/models/user"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserService interface {
	GetById(int) (*models.User, error)
	GetByMobile(string) (*models.User, error)
	GenerateJWT(uint, string, string) (string, string, error)
	ParseJWT(string) (*Claims, error)
	RefreshToken(string, string, *Claims) (string, string, error)
}
type UserServiceImpl struct {
}

func (u UserServiceImpl) GetById(id int) (*models.User, error) {
	return models.GetById(uint(id))
}

func (u UserServiceImpl) GetByMobile(mobile string) (*models.User, error) {
	return models.GetByMobile(mobile)
}

// Claims 定义JWT声明
type Claims struct {
	Username string `json:"username"`
	Uid      string `json:"uid"`
	jwt.RegisteredClaims
}

// 生成JWT令牌
func (u UserServiceImpl) GenerateJWT(uid uint, username string, userAgent string) (string, string, error) {
	claims := Claims{
		Username: username,
		Uid:      strconv.Itoa(int(uid)),
		RegisteredClaims: jwt.RegisteredClaims{
			// ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Issuer:   "biasura-app",
		},
	}
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(5 * time.Hour * 24)) //访问token只有1小时过期
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtSecret, err := os.ReadFile(config.C.BasePath + "/config/jwt/biasura-private.key")
	if err != nil {
		return "", "", err
	}
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(60 * 24 * time.Hour)) //刷新token有60天过期
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessTokenStr, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}
	refreshTokenStr, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}
	tokenStore := GetTokenStore()
	tokenStore.Add(refreshTokenStr, UserFeature{
		RefreshToken: refreshTokenStr,
		UserAgent:    userAgent,
		ExpiresAt:    time.Now().Add(60 * 24 * time.Hour),
	})

	return accessTokenStr, refreshTokenStr, nil
}

// 兼容php里生成的token
func (u UserServiceImpl) ParseJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		switch token.Method.Alg() {
		case jwt.SigningMethodHS256.Alg(): //对称密钥
			secret, err := os.ReadFile(config.C.BasePath + "/config/jwt/biasura-private.key")
			if err != nil {
				return nil, err
			}
			return secret, nil
		case jwt.SigningMethodRS256.Alg():
			pubKeyPath := config.C.BasePath + "/config/jwt/biasura-public.pem"
			publicKey, err := loadRSAPublicKey(pubKeyPath)
			if err != nil {
				return nil, err
			}
			return publicKey, nil
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 这里可以加上 issuer、audience、jti 等 claim 校验
	return claims, nil
}

// RefreshToken 刷新JWT令牌
func (u UserServiceImpl) RefreshToken(tokenString string, userAgent string, claims *Claims) (string, string, error) {

	// 将uid从string转换为uint
	uid, err := strconv.ParseUint(claims.Uid, 10, 32)
	if err != nil {
		return "", "", fmt.Errorf("invalid uid in token: %v", err)
	}
	newAccessToken, newRefreshToken, err := u.GenerateJWT(uint(uid), claims.Username, userAgent)
	if err != nil {
		return "", "", err
	}
	tokenStore := GetTokenStore()
	tokenStore.Add(newRefreshToken, UserFeature{
		RefreshToken: newRefreshToken,
		UserAgent:    userAgent,
		ExpiresAt:    time.Now().Add(60 * 24 * time.Hour),
	})
	if err != nil {
		return "", "", err
	}
	// 生成新的token
	return newAccessToken, newRefreshToken, nil
}

func loadRSAPublicKey(path string) (*rsa.PublicKey, error) {
	pubBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pubBytes)
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pub.(*rsa.PublicKey), nil
}
