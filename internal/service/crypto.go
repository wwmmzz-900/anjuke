package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"sync"
	"time"
)

// 消息序列号管理器
type SequenceManager struct {
	mutex      sync.RWMutex
	sequences  map[int64]int64 // 用户ID -> 当前序列号
	timestamps map[int64]int64 // 用户ID -> 最后消息时间戳
}

// 创建新的序列号管理器
func NewSequenceManager() *SequenceManager {
	return &SequenceManager{
		sequences:  make(map[int64]int64),
		timestamps: make(map[int64]int64),
	}
}

// 获取并递增序列号
func (sm *SequenceManager) GetNextSequence(userID int64) int64 {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	seq := sm.sequences[userID]
	seq++
	sm.sequences[userID] = seq
	sm.timestamps[userID] = time.Now().UnixNano()
	
	return seq
}

// 验证序列号
func (sm *SequenceManager) ValidateSequence(userID, sequence int64) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	lastSeq := sm.sequences[userID]
	// 允许序列号比当前大（未来消息）或者比当前小不超过100（延迟消息，但不会太旧）
	return sequence > lastSeq || (lastSeq - sequence) < 100
}

// 更新序列号（如果新序列号更大）
func (sm *SequenceManager) UpdateSequence(userID, sequence int64) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	if sequence > sm.sequences[userID] {
		sm.sequences[userID] = sequence
		sm.timestamps[userID] = time.Now().UnixNano()
		return true
	}
	return false
}

// 全局序列号管理器
var GlobalSequenceManager = NewSequenceManager()

// 生成密钥
func GenerateKey(secret string) []byte {
	hash := sha256.Sum256([]byte(secret))
	return hash[:]
}

// 加密消息
func EncryptMessage(message []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	
	// 创建 GCM 模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	// 创建随机数
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	
	// 加密
	ciphertext := aesGCM.Seal(nonce, nonce, message, nil)
	
	// 转为 Base64 编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// 解密消息
func DecryptMessage(encryptedMessage string, key []byte) ([]byte, error) {
	// 解码 Base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedMessage)
	if err != nil {
		return nil, err
	}
	
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	// 创建 GCM 模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	// 获取随机数大小
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("密文太短")
	}
	
	// 提取随机数和密文
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	
	// 解密
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}

// 生成会话密钥
func GenerateSessionKey() string {
	// 生成随机字节
	bytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		// 如果随机数生成失败，使用时间戳作为备用
		bytes = []byte(time.Now().String())
	}
	
	// 转为十六进制字符串
	return hex.EncodeToString(bytes)
}