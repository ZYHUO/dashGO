package utils

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// GenerateUUID 生成 UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateToken 生成随机 token
func GenerateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}

// HashPassword 使用 bcrypt 加密密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// MD5 计算 MD5
func MD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// SHA256 计算 SHA256
func SHA256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// GetServerKey 生成服务器密�?(用于 Shadowsocks 2022)
// size: 16 for aes-128, 32 for aes-256/chacha20
func GetServerKey(createdAt int64, size int) string {
	// 使用 createdAt 作为种子生成固定的服务器密钥
	seed := fmt.Sprintf("xboard-ss2022-server-key-%d", createdAt)
	hash := sha256.Sum256([]byte(seed))
	// 取前 size 字节并编码为 base64
	return base64.StdEncoding.EncodeToString(hash[:size])
}

// UUIDToBase64 �?UUID 转换�?Base64 密钥 (用于 Shadowsocks 2022)
// size: 16 for aes-128, 32 for aes-256/chacha20
func UUIDToBase64(uuidStr string, size int) string {
	// 移除 UUID 中的连字�?
	cleanUUID := strings.ReplaceAll(uuidStr, "-", "")
	
	// 使用 UUID 作为种子生成用户密钥
	seed := fmt.Sprintf("xboard-ss2022-user-key-%s", cleanUUID)
	hash := sha256.Sum256([]byte(seed))
	// 取前 size 字节并编码为 base64
	return base64.StdEncoding.EncodeToString(hash[:size])
}

// GenerateSS2022Password 生成完整�?SS2022 密码
// cipher: 加密方式 (2022-blake3-aes-128-gcm, 2022-blake3-aes-256-gcm, 2022-blake3-chacha20-poly1305)
// createdAt: 服务器创建时间戳
// userUUID: 用户 UUID
// 返回格式: serverKey:userKey (用于客户�? �?serverKey (用于服务�?
func GenerateSS2022Password(cipher string, createdAt int64, userUUID string) string {
	var keySize int
	switch cipher {
	case "2022-blake3-aes-128-gcm":
		keySize = 16
	case "2022-blake3-aes-256-gcm", "2022-blake3-chacha20-poly1305":
		keySize = 32
	default:
		// �?SS2022 加密方式，直接返�?UUID
		return userUUID
	}
	
	serverKey := GetServerKey(createdAt, keySize)
	userKey := UUIDToBase64(userUUID, keySize)
	return serverKey + ":" + userKey
}

// GetSS2022ServerPassword 获取 SS2022 服务端密�?(仅服务器密钥)
func GetSS2022ServerPassword(cipher string, createdAt int64) string {
	var keySize int
	switch cipher {
	case "2022-blake3-aes-128-gcm":
		keySize = 16
	case "2022-blake3-aes-256-gcm", "2022-blake3-chacha20-poly1305":
		keySize = 32
	default:
		return ""
	}
	return GetServerKey(createdAt, keySize)
}

// GetSS2022UserPassword 获取 SS2022 用户密钥 (仅用户密钥，用于服务端用户列�?
func GetSS2022UserPassword(cipher string, userUUID string) string {
	var keySize int
	switch cipher {
	case "2022-blake3-aes-128-gcm":
		keySize = 16
	case "2022-blake3-aes-256-gcm", "2022-blake3-chacha20-poly1305":
		keySize = 32
	default:
		return userUUID
	}
	return UUIDToBase64(userUUID, keySize)
}

// RandomPort 从端口范围中随机选择一个端�?
func RandomPort(portRange string) int {
	parts := strings.Split(portRange, "-")
	if len(parts) != 2 {
		return 0
	}
	var start, end int
	fmt.Sscanf(parts[0], "%d", &start)
	fmt.Sscanf(parts[1], "%d", &end)
	if start >= end {
		return start
	}
	b := make([]byte, 4)
	rand.Read(b)
	return start + int(b[0])%(end-start+1)
}

// GenerateNumericCode 生成数字验证�?
func GenerateNumericCode(length int) string {
	const digits = "0123456789"
	code := make([]byte, length)
	for i := range code {
		b := make([]byte, 1)
		rand.Read(b)
		code[i] = digits[int(b[0])%10]
	}
	return string(code)
}
