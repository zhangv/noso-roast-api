package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"hahacking/config"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"username"`
	Uid      string `json:"uid"`
	jwt.RegisteredClaims
}

// 生成JWT令牌
func GenerateJWT(uid uint, username string, claims *Claims) (string, error) {
	cfg := config.Get("jwt")
	signingMethod := cfg["signing_method"].(string)
	switch signingMethod {
	case "RS256":
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		// 加载RSA私钥
		privateKey, err := loadRSAPrivateKey(config.C.BasePath + cfg["private_key"].(string))
		if err != nil {
			return "", fmt.Errorf("failed to load private key: %v", err)
		}
		return token.SignedString(privateKey)
	case "HS256":
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		jwtSecret := []byte(cfg["secret"].(string))
		return token.SignedString(jwtSecret)
	default:
		return "", fmt.Errorf("invalid signing method: %s", signingMethod)
	}

}

// loadRSAPrivateKey 加载RSA私钥
func loadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	privateKeyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 尝试PKCS8格式
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %v", err)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not RSA key")
		}
		return rsaKey, nil
	}

	return privateKey, nil
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

func ParseJWT(tokenString string) (*Claims, error) {
	cfg := config.Get("jwt")
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		switch token.Method.Alg() {
		case jwt.SigningMethodHS256.Alg(): //对称密钥
			secret := []byte(cfg["secret"].(string))
			return secret, nil
		case jwt.SigningMethodRS256.Alg(): //非对称算法
			pubKeyPath := config.C.BasePath + cfg["public_key"].(string)
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

	return claims, nil
}
