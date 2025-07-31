package common

import (
	"math/rand"
	"time"
)

// 生成随机字符串，用于生成上传ID
func RandString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// 生成唯一的上传ID
func GenerateUploadID() string {
	return time.Now().Format("20060102150405") + "_" + RandString(8)
}
