package user

import (
	"fmt"
	"hahacking/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	// 初始化配置，使用相对于api根目录的路径
	config.Setup("../../", "test")
}

func TestGenerateJWT_ExpiresAt(t *testing.T) {
	userService := UserServiceImpl{}

	// 生成一个初始token
	uid := uint(123)
	username := "testuser"
	userAgent := "test-user-agent"
	accessToken, refreshToken, err := userService.GenerateJWT(uid, username, userAgent)
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	// 解析token验证内容
	claims, err := userService.ParseJWT(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, "123", claims.Uid)

	// 验证ExpiresAt是否正确设置
	assert.NotNil(t, claims.ExpiresAt)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
	assert.True(t, claims.ExpiresAt.Time.Before(time.Now().Add(61*24*time.Hour))) // 应该在60天+1小时内

	// 验证IssuedAt是否正确设置
	assert.NotNil(t, claims.IssuedAt)
	assert.True(t, claims.IssuedAt.Time.Before(time.Now().Add(time.Second)))
	assert.True(t, claims.IssuedAt.Time.After(time.Now().Add(-time.Second)))

	// 验证Issuer是否正确设置
	assert.Equal(t, "biasura-app", claims.Issuer)

	t.Logf("Token expires at: %v", claims.ExpiresAt.Time)
	t.Logf("Token issued at: %v", claims.IssuedAt.Time)
}

func TestJWTStructure(t *testing.T) {
	userService := UserServiceImpl{}

	// 生成token
	uid := uint(456)
	username := "testuser2"
	userAgent := "test-user-agent"
	accessToken, _, err := userService.GenerateJWT(uid, username, userAgent)
	assert.NoError(t, err)

	// 解析token
	claims, err := userService.ParseJWT(accessToken)
	assert.NoError(t, err)

	// 验证所有字段都正确设置
	fmt.Printf("=== JWT Claims 详细信息 ===\n")
	fmt.Printf("Username: %s\n", claims.Username)
	fmt.Printf("Uid: %s\n", claims.Uid)
	fmt.Printf("ExpiresAt: %v\n", claims.ExpiresAt)
	fmt.Printf("IssuedAt: %v\n", claims.IssuedAt)
	fmt.Printf("Issuer: %s\n", claims.Issuer)
	fmt.Printf("NotBefore: %v\n", claims.NotBefore)
	fmt.Printf("Subject: %s\n", claims.Subject)
	fmt.Printf("Audience: %v\n", claims.Audience)
	fmt.Printf("ID: %s\n", claims.ID)

	// 验证嵌入的RegisteredClaims是否正常工作
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, "456", claims.Uid)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.Equal(t, "biasura-app", claims.Issuer)

	// 验证时间逻辑
	now := time.Now()
	assert.True(t, claims.IssuedAt.Time.Before(now.Add(time.Second)))
	assert.True(t, claims.ExpiresAt.Time.After(now))

	// 验证过期时间大约是60天后
	expectedExpiry := now.Add(60 * 24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time
	timeDiff := actualExpiry.Sub(expectedExpiry)
	assert.True(t, timeDiff < time.Minute, "过期时间应该在60天后的1分钟内")
}
