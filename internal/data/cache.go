package data

import (
	"github.com/dgraph-io/ristretto"
	"time"
)

// LocalCache 本地缓存接口（与OrderCacheRepo中定义一致）
type LocalCache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
}

// RistrettoCache 基于Ristretto的本地缓存实现
type RistrettoCache struct {
	cache *ristretto.Cache
}

// NewRistrettoCache 创建本地缓存实例
func NewRistrettoCache() (*RistrettoCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 100000, // 必须设置，且不能为0（建议设为预期缓存项的10倍）
		MaxCost:     10000,  // 最大缓存项数量（根据内存调整）
		BufferItems: 64,     // 缓冲区大小（提高命中率）
	})
	if err != nil {
		return nil, err
	}
	return &RistrettoCache{cache: cache}, nil
}

// Get 从本地缓存获取数据
func (r *RistrettoCache) Get(key string) (interface{}, bool) {
	return r.cache.Get(key)
}

// Set 写入本地缓存（带TTL过期）
func (r *RistrettoCache) Set(key string, value interface{}, ttl time.Duration) {
	// 第三个参数"1"是cost，简单场景下固定为1即可
	r.cache.SetWithTTL(key, value, 1, ttl)
}
