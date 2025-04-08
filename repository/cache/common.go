package cache

import (
	// "context"
	"strconv"

	// util "mall/pkg/utils"
	"sync"
	"time"

	"github.com/go-redis/redis"

	logging "github.com/sirupsen/logrus"

	"mall/conf"
)

// RedisClient Redis缓存客户端单例
var RedisClient *redis.Client

// InitCache 在中间件中初始化redis链接  防止循环导包，所以放在这里
func InitCache() {
	Redis()
}

// Redis 在中间件中初始化redis链接
func Redis() {
	db, _ := strconv.ParseUint(conf.RedisDbName, 10, 64)
	client := redis.NewClient(&redis.Options{
		Addr:     conf.RedisAddr,
		Password: conf.RedisPw,
		DB:       int(db),
	})
	_, err := client.Ping().Result()
	if err != nil {
		logging.Info(err)
		panic(err)
	}
	RedisClient = client
}

var LocalCache *localCache

// CacheItem 缓存项结构
type CacheItem struct {
	Value      interface{}
	Expiration int64 // 过期时间的时间戳（Unix 时间）
}

// LocalCache 本地缓存结构
type localCache struct {
	data sync.Map
}

// NewLocalCache 创建一个新的本地缓存实例
func NewLocalCache() *localCache {
	return &localCache{}
}
func InitLocalCache() {
	LocalCache = NewLocalCache()
}

// Set 设置缓存值
func (c *localCache) Set(key string, value interface{}, ttl time.Duration) {
	expiration := time.Now().Add(ttl).Unix()
	c.data.Store(key, CacheItem{
		Value:      value,
		Expiration: expiration,
	})
}

// Get 获取缓存值
func (c *localCache) Get(key string) (interface{}, bool) {
	item, ok := c.data.Load(key)
	if !ok {
		return nil, false
	}

	cacheItem := item.(CacheItem)
	// 检查是否过期
	if time.Now().Unix() > cacheItem.Expiration {
		c.data.Delete(key) // 删除过期的缓存
		return nil, false
	}

	return cacheItem.Value, true
}

// Delete 删除缓存值
func (c *localCache) Delete(key string) {
	c.data.Delete(key)
}
