package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// 常量定义
const (
	// 密钥长度 (32字节 = 256位)
	KeySize = 32

	// 最大允许的序列号差值，用于防止重放攻击
	MaxSequenceDiff = 100

	// 默认会话密钥过期时间（小时）
	DefaultSessionKeyExpiry = 24
)

// 错误定义
var (
	ErrInvalidKey             = errors.New("无效的加密密钥")
	ErrEncryptionFailed       = errors.New("加密失败")
	ErrDecryptionFailed       = errors.New("解密失败")
	ErrInvalidCiphertext      = errors.New("无效的密文格式")
	ErrCiphertextTooShort     = errors.New("密文太短")
	ErrRandomGenerationFailed = errors.New("随机数生成失败")
)

// SequenceManager 消息序列号管理器
// 用于管理用户消息的序列号，防止重放攻击和确保消息顺序
type SequenceManager struct {
	mutex      sync.RWMutex
	sequences  map[int64]int64 // 用户ID -> 当前序列号
	timestamps map[int64]int64 // 用户ID -> 最后消息时间戳
}

// NewSequenceManager 创建新的序列号管理器
func NewSequenceManager() *SequenceManager {
	return &SequenceManager{
		sequences:  make(map[int64]int64),
		timestamps: make(map[int64]int64),
	}
}

// GetNextSequence 获取并递增序列号
// 为指定用户生成下一个序列号，并更新时间戳
func (sm *SequenceManager) GetNextSequence(userID int64) int64 {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	seq := sm.sequences[userID]
	seq++
	sm.sequences[userID] = seq
	sm.timestamps[userID] = time.Now().UnixNano()

	return seq
}

// ValidateSequence 验证序列号
// 检查给定序列号是否有效（防止重放攻击）
// 允许序列号比当前大（未来消息）或者比当前小不超过MaxSequenceDiff（延迟消息）
func (sm *SequenceManager) ValidateSequence(userID, sequence int64) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	lastSeq, exists := sm.sequences[userID]
	if !exists {
		// 如果用户没有序列号记录，则接受任何序列号
		return true
	}

	// 允许序列号比当前大（未来消息）或者比当前小不超过MaxSequenceDiff（延迟消息）
	return sequence > lastSeq || (lastSeq-sequence) < MaxSequenceDiff
}

// UpdateSequence 更新序列号
// 如果新序列号更大，则更新用户的序列号和时间戳
// 返回是否更新成功
func (sm *SequenceManager) UpdateSequence(userID, sequence int64) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	currentSeq, exists := sm.sequences[userID]
	if !exists || sequence > currentSeq {
		sm.sequences[userID] = sequence
		sm.timestamps[userID] = time.Now().UnixNano()
		return true
	}
	return false
}

// GetUserStats 获取用户序列号统计信息
// 返回用户当前序列号和最后更新时间
func (sm *SequenceManager) GetUserStats(userID int64) (sequence int64, lastUpdate time.Time, exists bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	seq, seqExists := sm.sequences[userID]
	ts, tsExists := sm.timestamps[userID]

	if !seqExists || !tsExists {
		return 0, time.Time{}, false
	}

	return seq, time.Unix(0, ts), true
}

// CleanupExpiredSequences 清理过期的序列号
// 删除超过指定时间未活动的用户序列号记录
func (sm *SequenceManager) CleanupExpiredSequences(maxAge time.Duration) int {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now().UnixNano()
	maxAgeNanos := maxAge.Nanoseconds()
	count := 0

	for userID, ts := range sm.timestamps {
		if now-ts > maxAgeNanos {
			delete(sm.sequences, userID)
			delete(sm.timestamps, userID)
			count++
		}
	}

	return count
}

// 全局序列号管理器
var GlobalSequenceManager = NewSequenceManager()

// GenerateKey 生成加密密钥
// 从给定的密钥字符串生成256位加密密钥
func GenerateKey(secret string) []byte {
	if secret == "" {
		// 如果密钥为空，使用当前时间作为备用
		secret = time.Now().String()
	}

	hash := sha256.Sum256([]byte(secret))
	return hash[:]
}

// EncryptMessage 加密消息
// 使用AES-GCM算法加密消息，并返回Base64编码的密文
func EncryptMessage(message []byte, key []byte) (string, error) {
	if len(key) != KeySize {
		return "", fmt.Errorf("%w: 密钥长度应为%d字节", ErrInvalidKey, KeySize)
	}

	if len(message) == 0 {
		// 空消息直接返回空字符串
		return "", nil
	}

	// 创建AES加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	// 创建GCM模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	// 创建随机数
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("%w: %v", ErrRandomGenerationFailed, err)
	}

	// 加密
	ciphertext := aesGCM.Seal(nonce, nonce, message, nil)

	// 转为Base64编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptMessage 解密消息
// 解密Base64编码的密文，使用AES-GCM算法
func DecryptMessage(encryptedMessage string, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("%w: 密钥长度应为%d字节", ErrInvalidKey, KeySize)
	}

	if encryptedMessage == "" {
		// 空密文直接返回空消息
		return []byte{}, nil
	}

	// 解码Base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedMessage)
	if err != nil {
		return nil, fmt.Errorf("%w: Base64解码失败: %v", ErrInvalidCiphertext, err)
	}

	// 创建AES加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	// 创建GCM模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	// 获取随机数大小
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrCiphertextTooShort
	}

	// 提取随机数和密文
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}

// GenerateSessionKey 生成会话密钥
// 生成一个随机的会话密钥，用于WebSocket连接
func GenerateSessionKey() string {
	// 生成随机字节
	bytes := make([]byte, KeySize)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		// 如果随机数生成失败，使用时间戳作为备用
		// 添加一些额外的熵源
		timeStr := time.Now().String()
		hash := sha256.Sum256([]byte(timeStr))
		copy(bytes, hash[:])
	}

	// 转为十六进制字符串
	return hex.EncodeToString(bytes)
}

// VerifyMessageAuthenticity 验证消息真实性
// 检查消息是否来自声称的发送者，并且序列号有效
func VerifyMessageAuthenticity(senderID, sequence int64, signature string, message []byte, key []byte) bool {
	// 验证序列号
	if !GlobalSequenceManager.ValidateSequence(senderID, sequence) {
		return false
	}

	// TODO: 实现消息签名验证
	// 这里可以添加基于HMAC或数字签名的验证逻辑

	return true
}

// SessionKeyManager 会话密钥管理器
// 管理用户会话密钥的生成、存储和过期
type SessionKeyManager struct {
	mutex     sync.RWMutex
	keys      map[int64]string    // 用户ID -> 会话密钥
	expiry    map[int64]time.Time // 用户ID -> 过期时间
	keyToUser map[string]int64    // 会话密钥 -> 用户ID (反向查找)
}

// NewSessionKeyManager 创建新的会话密钥管理器
func NewSessionKeyManager() *SessionKeyManager {
	return &SessionKeyManager{
		keys:      make(map[int64]string),
		expiry:    make(map[int64]time.Time),
		keyToUser: make(map[string]int64),
	}
}

// GenerateKeyForUser 为用户生成新的会话密钥
func (skm *SessionKeyManager) GenerateKeyForUser(userID int64, expiryHours int) string {
	skm.mutex.Lock()
	defer skm.mutex.Unlock()

	// 生成新密钥
	sessionKey := GenerateSessionKey()

	// 设置过期时间
	if expiryHours <= 0 {
		expiryHours = DefaultSessionKeyExpiry
	}
	expiryTime := time.Now().Add(time.Duration(expiryHours) * time.Hour)

	// 如果用户已有密钥，先从反向映射中删除
	if oldKey, exists := skm.keys[userID]; exists {
		delete(skm.keyToUser, oldKey)
	}

	// 存储新密钥
	skm.keys[userID] = sessionKey
	skm.expiry[userID] = expiryTime
	skm.keyToUser[sessionKey] = userID

	return sessionKey
}

// GetKeyForUser 获取用户的会话密钥
func (skm *SessionKeyManager) GetKeyForUser(userID int64) (string, bool) {
	skm.mutex.RLock()
	defer skm.mutex.RUnlock()

	key, exists := skm.keys[userID]
	if !exists {
		return "", false
	}

	// 检查是否过期
	expiryTime, hasExpiry := skm.expiry[userID]
	if hasExpiry && time.Now().After(expiryTime) {
		// 密钥已过期，但不在这里删除，留给清理函数处理
		return "", false
	}

	return key, true
}

// GetUserByKey 通过会话密钥查找用户
func (skm *SessionKeyManager) GetUserByKey(sessionKey string) (int64, bool) {
	skm.mutex.RLock()
	defer skm.mutex.RUnlock()

	userID, exists := skm.keyToUser[sessionKey]
	if !exists {
		return 0, false
	}

	// 检查是否过期
	expiryTime, hasExpiry := skm.expiry[userID]
	if hasExpiry && time.Now().After(expiryTime) {
		return 0, false
	}

	return userID, true
}

// CleanupExpiredKeys 清理过期的会话密钥
func (skm *SessionKeyManager) CleanupExpiredKeys() int {
	skm.mutex.Lock()
	defer skm.mutex.Unlock()

	now := time.Now()
	count := 0

	for userID, expiryTime := range skm.expiry {
		if now.After(expiryTime) {
			// 密钥已过期，删除
			sessionKey := skm.keys[userID]
			delete(skm.keys, userID)
			delete(skm.expiry, userID)
			delete(skm.keyToUser, sessionKey)
			count++
		}
	}

	return count
}

// 全局会话密钥管理器
var GlobalSessionKeyManager = NewSessionKeyManager()
