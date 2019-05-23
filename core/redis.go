package core

import (
	"encoding/json"
	"os"
	"time"

	"github.com/ekas-data-portal/models"
	"github.com/go-redis/redis"
)

var (
	redisClient *redis.Client
	dockerURL   = "157.230.0.138:7001"
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

// GetValue ...
func GetValue(key string) (interface{}, error) {
	var deserializedValue interface{}
	serializedValue, err := redisClient.Get(key).Result()
	json.Unmarshal([]byte(serializedValue), &deserializedValue)
	return deserializedValue, err
}

// SetValue ...
func SetValue(key string, value interface{}) (bool, error) {
	serializedValue, _ := json.Marshal(value)
	err := redisClient.Set(key, string(serializedValue), 0).Err()
	return true, err
}

// GetLastSeenValue ...
func GetLastSeenValue(key string) (models.DeviceData, error) {
	var deserializedValue models.LastSeenStruct
	serializedValue, err := redisClient.Get(key).Result()
	json.Unmarshal([]byte(serializedValue), &deserializedValue)
	return deserializedValue.DeviceData, err
}

// SetValueWithTTL ...
func SetValueWithTTL(key string, value interface{}, ttl int) (bool, error) {
	serializedValue, _ := json.Marshal(value)
	err := redisClient.Set(key, string(serializedValue), time.Duration(ttl)*time.Second).Err()
	return true, err
}

// HSet ...
func HSet(key string, value interface{}) (bool, error) {
	serializedValue, _ := json.Marshal(value)
	err := redisClient.HSet(key, string(serializedValue), 0).Err()
	return true, err
}

// HGetAll ...
func HGetAll(key string) (interface{}, error) {
	// var deserializedValue interface{}
	serializedValue, err := redisClient.HGetAll(key).Result()
	// json.Unmarshal([]byte(serializedValue), &deserializedValue)
	return serializedValue, err
}

// RPush ...
func RPush(key string, valueList []string) (bool, error) {
	err := redisClient.RPush(key, valueList).Err()
	return true, err
}

// RpushWithTTL ...
func RpushWithTTL(key string, valueList []string, ttl int) (bool, error) {
	err := redisClient.RPush(key, valueList, ttl).Err()
	return true, err
}

// LRange ...
func LRange(key string) (bool, error) {
	err := redisClient.LRange(key, 0, -1).Err()
	return true, err
}

// ListLength ...
func ListLength(key string) int64 {
	return redisClient.LLen(key).Val()
}

// Publish ...
func Publish(channel string, message string) {
	redisClient.Publish(channel, message)
}

// GetKeyListByPattern ...
func GetKeyListByPattern(pattern string) []string {
	return redisClient.Keys(pattern).Val()
}

// IncrementValue ...
func IncrementValue(key string) int64 {
	return redisClient.Incr(key).Val()
}

// DelKey ...
func DelKey(key string) error {
	return redisClient.Del(key).Err()
}

// ListKeys ...
func ListKeys(key string) ([]string, error) {
	return redisClient.Keys(key).Result()
}

// SAdd Add values to a set...
func SAdd(key string, members interface{}) (bool, error) {
	serializedValue, _ := json.Marshal(members)
	err := redisClient.SAdd(key, string(serializedValue), 0).Err()
	return true, err
}

// LPush push values to the beggining of alist ...
func LPush(key string, members interface{}) (bool, error) {
	serializedValue, _ := json.Marshal(members)
	err := redisClient.LPush(key, string(serializedValue)).Err()
	return true, err
}

// ZAdd ...
func ZAdd(key string, score int64, members interface{}) (bool, error) {
	serializedValue, _ := json.Marshal(members)
	err := redisClient.ZAdd(key, redis.Z{
		Score:  float64(score),
		Member: string(serializedValue),
	}).Err()
	return true, err
}
