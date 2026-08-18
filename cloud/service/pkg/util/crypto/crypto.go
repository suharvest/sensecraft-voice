// Package crypto 提供对称加密能力，用于敏感字段（如 asr_servers.api_key）落库前加密。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	// cipherPrefix 密文前缀，用于区分密文与历史明文数据
	cipherPrefix = "gcm:"
)

var (
	ErrEmptyMasterKey = errors.New("crypto: master key is empty")
	ErrInvalidCipher  = errors.New("crypto: invalid cipher text")
)

// Cipher AES-GCM 加解密器
type Cipher struct {
	aead cipher.AEAD
}

// NewCipher 用主密钥构造加解密器。
// 主密钥为任意长度字符串，内部用 SHA-256 派生为 32 字节 AES 密钥。
func NewCipher(masterKey string) (*Cipher, error) {
	if strings.TrimSpace(masterKey) == "" {
		return nil, ErrEmptyMasterKey
	}
	sum := sha256.Sum256([]byte(masterKey))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Cipher{aead: aead}, nil
}

// Encrypt 加密明文，返回 "gcm:" + base64(nonce||ciphertext)。空明文返回空串。
func (c *Cipher) Encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := c.aead.Seal(nil, nonce, []byte(plain), nil)
	out := make([]byte, 0, len(nonce)+len(sealed))
	out = append(out, nonce...)
	out = append(out, sealed...)
	return cipherPrefix + base64.StdEncoding.EncodeToString(out), nil
}

// Decrypt 解密密文。
// 兼容性处理：没有 "gcm:" 前缀的输入视为历史明文原样返回，便于灰度迁移。
func (c *Cipher) Decrypt(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	if !strings.HasPrefix(encoded, cipherPrefix) {
		return encoded, nil
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(encoded, cipherPrefix))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidCipher, err)
	}
	ns := c.aead.NonceSize()
	if len(raw) <= ns {
		return "", ErrInvalidCipher
	}
	plain, err := c.aead.Open(nil, raw[:ns], raw[ns:], nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidCipher, err)
	}
	return string(plain), nil
}

// IsCipherText 判断是否为本包生成的密文
func IsCipherText(s string) bool { return strings.HasPrefix(s, cipherPrefix) }
