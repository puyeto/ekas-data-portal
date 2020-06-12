package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ekas-data-portal/models"
	"github.com/go-redis/redis"
)

var (
	redisClient *redis.Client
)

// InitializeRedis ...
func InitializeRedis() error {
	ctx, _ := context.WithCancel(context.Background())
	// if os.Getenv("GO_ENV") != "production" {
	// 	redisURL = "db-redis-cluster-do-user-4666162-0.db.ondigitalocean.com:25061"
	// }

	opt, _ := redis.ParseURL("rediss://default:wdbsxehbizfl5kbu@db-redis-cluster-do-user-4666162-0.db.ondigitalocean.com:25061/1")
	opt.PoolSize = 100
	opt.MaxRetries = 2
	opt.ReadTimeout = -1

	redisClient = redis.NewClient(opt)

	ping, err := redisClient.Ping(ctx).Result()
	if err == nil && len(ping) > 0 {
		Logger.Info("Connected to Redis")
		return nil
	}
	Logger.Error("Redis Connection Failed")
	return err
}

// GetValue ...
func GetValue(key string) (interface{}, error) {
	var deserializedValue interface{}
	ctx, _ := context.WithCancel(context.Background())
	serializedValue, err := redisClient.Get(ctx, key).Result()
	json.Unmarshal([]byte(serializedValue), &deserializedValue)
	return deserializedValue, err
}

// SetValue ...
func SetValue(key string, value interface{}) (bool, error) {
	ctx, _ := context.WithCancel(context.Background())
	serializedValue, _ := json.Marshal(value)
	err := redisClient.Set(ctx, key, string(serializedValue), 0).Err()
	return true, err
}

// GetLastSeenValue ...
func GetLastSeenValue(key string) (models.DeviceData, error) {
	var deserializedValue models.LastSeenStruct
	ctx, _ := context.WithCancel(context.Background())
	serializedValue, err := redisClient.Get(ctx, key).Result()
	json.Unmarshal([]byte(serializedValue), &deserializedValue)
	return deserializedValue.DeviceData, err
}

// SetValueWithTTL ...
func SetValueWithTTL(key string, value interface{}, ttl int) (bool, error) {
	ctx, _ := context.WithCancel(context.Background())
	serializedValue, _ := json.Marshal(value)
	err := redisClient.Set(ctx, key, string(serializedValue), time.Duration(ttl)*time.Second).Err()
	return true, err
}

// HSet ...
func HSet(key string, value interface{}) (bool, error) {
	ctx, _ := context.WithCancel(context.Background())
	serializedValue, _ := json.Marshal(value)
	err := redisClient.HSet(ctx, key, string(serializedValue), 0).Err()
	return true, err
}

// HGetAll ...
func HGetAll(key string) (interface{}, error) {
	// var deserializedValue interface{}
	ctx, _ := context.WithCancel(context.Background())
	serializedValue, err := redisClient.HGetAll(ctx, key).Result()
	// json.Unmarshal([]byte(serializedValue), &deserializedValue)
	return serializedValue, err
}

// RPush ...
func RPush(key string, valueList []string) (bool, error) {
	ctx, _ := context.WithCancel(context.Background())
	err := redisClient.RPush(ctx, key, valueList).Err()
	return true, err
}

// RpushWithTTL ...
func RpushWithTTL(key string, valueList []string, ttl int) (bool, error) {
	ctx, _ := context.WithCancel(context.Background())
	err := redisClient.RPush(ctx, key, valueList, ttl).Err()
	return true, err
}

// LRange ...
func LRange(key string) (bool, error) {
	ctx, _ := context.WithCancel(context.Background())
	err := redisClient.LRange(ctx, key, 0, -1).Err()
	return true, err
}

// ListLength ...
func ListLength(key string) int64 {
	ctx, _ := context.WithCancel(context.Background())
	return redisClient.LLen(ctx, key).Val()
}

// Publish ...
func Publish(channel string, message string) {
	ctx, _ := context.WithCancel(context.Background())
	redisClient.Publish(ctx, channel, message)
}

// GetKeyListByPattern ...
func GetKeyListByPattern(pattern string) []string {
	ctx, _ := context.WithCancel(context.Background())
	return redisClient.Keys(ctx, pattern).Val()
}

// IncrementValue ...
func IncrementValue(key string) int64 {
	ctx, _ := context.WithCancel(context.Background())
	return redisClient.Incr(ctx, key).Val()
}

// DelKey ...
func DelKey(key string) error {
	ctx, _ := context.WithCancel(context.Background())
	return redisClient.Del(ctx, key).Err()
}

// ListKeys ...
func ListKeys(key string) ([]string, error) {
	ctx, _ := context.WithCancel(context.Background())
	return redisClient.Keys(ctx, key).Result()
}

// SAdd Add values to a set...
func SAdd(key string, members interface{}) (bool, error) {
	ctx, _ := context.WithCancel(context.Background())
	serializedValue, _ := json.Marshal(members)
	err := redisClient.SAdd(ctx, key, string(serializedValue), 0).Err()
	return true, err
}

// LPush push values to the beggining of alist ...
func LPush(key string, members interface{}) (bool, error) {
	ctx, _ := context.WithCancel(context.Background())
	serializedValue, _ := json.Marshal(members)
	err := redisClient.LPush(ctx, key, string(serializedValue)).Err()
	return true, err
}

// ZAdd ...
func ZAdd(key string, score int64, members interface{}) (bool, error) {
	ctx, _ := context.WithCancel(context.Background())
	serializedValue, _ := json.Marshal(members)
	err := redisClient.ZAdd(ctx, key, &redis.Z{
		Score:  float64(score),
		Member: string(serializedValue),
	}).Err()
	return true, err
}
