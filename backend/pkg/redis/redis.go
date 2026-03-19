package redis

import (
	"backend/config"
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var ctx = context.Background()

// 初始化Redis连接（在main.go中调用）
func InitRedis() error {
	addr := config.RDConfig.Host + ":" + strconv.Itoa(config.RDConfig.Port)

	//创建Redis客户端
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.RDConfig.Password,
		DB:       config.RDConfig.Db,
	})

	//测试连接
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("Redis连接实际失败：%v", err) // 新增：打印真实错误
		return err
	}

	return nil

}

// SetMonitorStatus 缓存监控项状态
func SetMonitorStatus(monitorId int64, status int) error {
	key := "monitor:status:" + strconv.Itoa(int(monitorId))
	return RedisClient.Set(ctx, key, status, 1*time.Hour).Err()
}

// GetMonitorStatus 获取缓存的监控状态
func GetMonitorStatus(monitorId int64) (status int, err error) {
	key := "monitor:status:" + strconv.Itoa(int(monitorId))
	val, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	status, _ = strconv.Atoi(val)
	return status, nil
}

func Set(key string, value interface{}, expire time.Duration) error {
	var val string
	switch v := value.(type) {
	case string:
		val = v
	case []byte:
		val = string(v)
	default:
		jsonVal, err := json.Marshal(v)
		if err != nil {
			return err
		}
		val = string(jsonVal)
	}
	return RedisClient.Set(ctx, key, val, expire).Err()
}

func Get(key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}

func GetJSON(key string, result interface{}) error {
	val, err := Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), result)
}

func Del(key string) error {
	return RedisClient.Del(ctx, key).Err()
}

func Exists(key string) (bool, error) {
	count, err := RedisClient.Exists(ctx, key).Result()
	return count > 0, err
}
