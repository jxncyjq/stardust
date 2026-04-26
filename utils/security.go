package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// VerifyHMAC 校验HMAC签名
// message: 原始消息（如 AccessKey+Timestamp）
// secret: 秘钥（sk）
// sign: 待校验的签名（十六进制字符串）
// 返回 true 表示校验通过，false 表示校验失败
func VerifyHMAC(message, secret, sign string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	expectedSign := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expectedSign), []byte(sign))
}

// GenerateHMAC 生成HMAC签名
// message: 原始消息（如 AccessKey+Timestamp）
// secret: 秘钥（sk）
// 返回签名（十六进制字符串）
func GenerateHMAC(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
