package core

import (
	"encoding/json"
	"os"
	"time"

	"github.com/go-redis/redis"
)

var (
	redisClient *redis.Client
	dockerURL   = "159.89.134.228:6379"
)

// InitializeRedis ...
func InitializeRedis() error {
	if os.Getenv("GO_ENV") != "production" {
		dockerURL = "localhost:6379"
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:       dockerURL,
		PoolSize:   100,
		MaxRetries: 2,
		Password:   "",
		DB:         0,
	})

	ping, err := redisClient.Ping().Result()
	if err == nil && len(ping) > 0 {
		println("Connected to Redis")
		return nil
	}
	println("Redis Connection Failed")
	return err
}

func GetValue(key string) (interface{}, error) {
	var deserializedValue interface{}
	serializedValue, err := redisClient.Get(key).Result()
	json.Unmarshal([]byte(serializedValue), &deserializedValue)
	return deserializedValue, err
}

func SetValue(key string, value interface{}) (bool, error) {
	serializedValue, _ := json.Marshal(value)
	err := redisClient.Set(key, string(serializedValue), 0).Err()
	return true, err
}

func SetValueWithTTL(key string, value interface{}, ttl int) (bool, error) {
	serializedValue, _ := json.Marshal(value)
	err := redisClient.Set(key, string(serializedValue), time.Duration(ttl)*time.Second).Err()
	return true, err
}

func RPush(key string, valueList []string) (bool, error) {
	err := redisClient.RPush(key, valueList).Err()
	return true, err
}

func RpushWithTTL(key string, valueList []string, ttl int) (bool, error) {
	err := redisClient.RPush(key, valueList, ttl).Err()
	return true, err
}

func LRange(key string) (bool, error) {
	err := redisClient.LRange(key, 0, -1).Err()
	return true, err
}

func ListLength(key string) int64 {
	return redisClient.LLen(key).Val()
}

func Publish(channel string, message string) {
	redisClient.Publish(channel, message)
}

func GetKeyListByPattern(pattern string) []string {
	return redisClient.Keys(pattern).Val()
}

func IncrementValue(key string) int64 {
	return redisClient.Incr(key).Val()
}

func DelKey(key string) error {
	return redisClient.Del(key).Err()
}
