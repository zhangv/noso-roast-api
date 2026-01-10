package jwt

import (
	"testing"
	"time"

	"hahacking/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func init() {
	// 初始化配置
	_, err := config.Setup("../../../", "dev")
	if err != nil {
		panic(err)
	}
}

func TestGenerateAndParseJWT(t *testing.T) {

	// 测试生成JWT
	uid := uint(123)
	username := "testuser"
	issuer := "biasura-api-go"
	issuedAt := time.Now()
	audience := jwt.ClaimStrings{"test"}
	token, err := GenerateJWT(uid, username, &Claims{
		Username: username,
		Uid:      "123",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			Audience:  audience,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * 24 * time.Hour)),
		},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 测试解析JWT
	claims, err := ParseJWT(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	// 验证claims内容
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, "123", claims.Uid)
	assert.Equal(t, issuer, claims.Issuer)
	assert.Equal(t, audience, claims.Audience)
	// 验证过期时间（应该在60天后）
	expectedExpireTime := time.Now().Add(60 * 24 * time.Hour)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
	assert.True(t, claims.ExpiresAt.Time.Before(expectedExpireTime.Add(1*time.Hour)))
}

func TestJWTExpiration(t *testing.T) {

	// 生成一个过期的JWT
	nowTime := time.Now()
	expireTime := nowTime.Add(-1 * time.Hour) // 1小时前过期

	claims := Claims{
		Username: "testuser",
		Uid:      "123",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime.Add(-2 * time.Hour)),
			Issuer:    "biasura-api-go",
		},
	}

	expiredToken, err := GenerateJWT(123, "testuser", &claims)
	assert.NoError(t, err)

	// 验证过期token无法解析
	_, err = ParseJWT(expiredToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestLoadRSAKeys(t *testing.T) {
	// 测试加载私钥
	privateKey, err := loadRSAPrivateKey(config.C.BasePath + "/config/jwt/biasura-private.pem")
	assert.NoError(t, err)
	assert.NotNil(t, privateKey)

	// 测试加载公钥
	publicKey, err := loadRSAPublicKey(config.C.BasePath + "/config/jwt/biasura-public.pem")
	assert.NoError(t, err)
	assert.NotNil(t, publicKey)

	// 验证密钥对匹配
	assert.Equal(t, privateKey.PublicKey.N, publicKey.N)
	assert.Equal(t, privateKey.PublicKey.E, publicKey.E)
}
