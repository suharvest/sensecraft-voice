package token

import (
	"crypto/md5"
	"fmt"
	"time"
)

// GenerateSeeedToken 生成Seeed API Token
// 规则: MD5(固定秘钥+当天日期)
// 固定秘钥: 由IT部门提供
// 当天日期: yyyyMMdd格式
// 例子: MD5("***REMOVED***" + "20250718")
func GenerateSeeedToken(secretKey string) string {
	// 获取当天日期，格式为 yyyyMMdd
	today := time.Now().Format("20060102")

	// 拼接固定秘钥和当天日期
	rawToken := secretKey + today

	// 计算MD5
	hash := md5.Sum([]byte(rawToken))
	token := fmt.Sprintf("%x", hash)

	return token
}

// GenerateSeeedTokenWithDate 生成指定日期的Seeed API Token
// 用于测试或历史日期验证
func GenerateSeeedTokenWithDate(secretKey string, date time.Time) string {
	// 格式化日期为 yyyyMMdd
	dateStr := date.Format("20060102")

	// 拼接固定秘钥和日期
	rawToken := secretKey + dateStr

	// 计算MD5
	hash := md5.Sum([]byte(rawToken))
	token := fmt.Sprintf("%x", hash)

	return token
}

// ValidateSeeedToken 验证Token是否正确
func ValidateSeeedToken(secretKey, token string) bool {
	expectedToken := GenerateSeeedToken(secretKey)
	return expectedToken == token
}

// GetTokenInfo 获取Token生成信息
func GetTokenInfo(secretKey string) map[string]string {
	today := time.Now().Format("20060102")
	rawToken := secretKey + today
	token := GenerateSeeedToken(secretKey)

	return map[string]string{
		"secret_key": secretKey,
		"date":       today,
		"raw_token":  rawToken,
		"md5_token":  token,
	}
}
