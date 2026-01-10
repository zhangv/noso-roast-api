# JWT非对称算法配置说明

## 概述

本项目使用RSA非对称算法进行JWT签名和验证，提供更高的安全性。

## 密钥文件

- `biasura-private.pem` - RSA私钥，用于JWT签名
- `biasura-public.pem` - RSA公钥，用于JWT验证

## 生成密钥对

```bash
# 生成2048位RSA私钥
openssl genrsa -out biasura-private.pem 2048

# 从私钥生成公钥
openssl rsa -in biasura-private.pem -pubout -out biasura-public.pem
```

## 安全注意事项

1. **私钥保护**：私钥文件权限应设置为600，只有服务进程可读
2. **密钥轮换**：定期更换密钥对，建议每6-12个月轮换一次
3. **备份**：安全备份私钥文件
4. **分发**：公钥可以安全地分发给其他服务进行JWT验证

## 使用示例

### 生成JWT
```go
userService := user.UserServiceImpl{}
token, err := userService.GenerateJWT(uid, username)
```

### 验证JWT
```go
claims, err := userService.ParseJWT(token)
```

## 算法对比

| 特性 | HMAC (对称) | RSA (非对称) |
|------|-------------|--------------|
| 密钥管理 | 需要安全共享密钥 | 私钥签名，公钥验证 |
| 安全性 | 中等 | 高 |
| 性能 | 快 | 较慢 |
| 密钥轮换 | 困难 | 容易 |
| 分布式验证 | 需要共享密钥 | 只需分发公钥 |

## 配置参数

JWT过期时间可在 `app.ini` 中配置：
```ini
[app]
JWTExpireHours = 1440  # 60天
```