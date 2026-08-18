package token

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
)

// deviceTokenBytes 设备凭证的随机字节数
const deviceTokenBytes = 32

// GenerateDeviceToken 生成设备长期凭证：32 字节随机数的 base64url 编码。
// 不用 JWT——JWT 硬编码 360 分钟有效期，不适合设备长期凭证；
// 不透明 token + hash 落库还支持服务端吊销。
func GenerateDeviceToken() (string, error) {
	buf := make([]byte, deviceTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// HashDeviceToken 计算设备凭证的 sha256 hex，落库只存这个值
func HashDeviceToken(token string) string {
	if token == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// SecureCompare 常量时间字符串比较，用于 enrollment key 校验
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
