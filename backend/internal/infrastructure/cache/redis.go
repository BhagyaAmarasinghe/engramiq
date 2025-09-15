package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engramiq/engramiq-backend/internal/config"
	"github.com/redis/go-redis/v9"
)

// Redis wraps the Redis client with common operations
// We use Redis for:
// 1. Session management (refresh tokens)
// 2. Query result caching
// 3. Rate limiting
// 4. Background job queuing
type Redis struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedis creates a new Redis client
func NewRedis(cfg config.RedisConfig) *Redis {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		panic(fmt.Errorf("invalid redis URL: %w", err))
	}

	opt.PoolSize = cfg.PoolSize
	opt.DialTimeout = cfg.DialTimeout

	client := redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		panic(fmt.Errorf("failed to connect to Redis: %w", err))
	}

	return &Redis{
		client: client,
		ctx:    ctx,
	}
}

// Session management methods

// SetRefreshToken stores a refresh token with expiration
func (r *Redis) SetRefreshToken(userID, token string, expiration time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return r.client.Set(r.ctx, key, userID, expiration).Err()
}

// GetRefreshToken retrieves the user ID associated with a refresh token
func (r *Redis) GetRefreshToken(token string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s", token)
	return r.client.Get(r.ctx, key).Result()
}

// DeleteRefreshToken removes a refresh token (logout)
func (r *Redis) DeleteRefreshToken(token string) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return r.client.Del(r.ctx, key).Err()
}

// Query caching methods

// SetQueryCache caches a query result
func (r *Redis) SetQueryCache(siteID, queryHash string, result interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("query_cache:%s:%s", siteID, queryHash)
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return r.client.Set(r.ctx, key, data, ttl).Err()
}

// GetQueryCache retrieves a cached query result
func (r *Redis) GetQueryCache(siteID, queryHash string, result interface{}) error {
	key := fmt.Sprintf("query_cache:%s:%s", siteID, queryHash)
	data, err := r.client.Get(r.ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

// Document processing queue

// EnqueueDocument adds a document to the processing queue
func (r *Redis) EnqueueDocument(documentID string, priority int) error {
	// Use sorted set for priority queue
	key := "document_processing_queue"
	score := float64(time.Now().Unix()) - float64(priority*1000) // Higher priority = lower score
	return r.client.ZAdd(r.ctx, key, redis.Z{
		Score:  score,
		Member: documentID,
	}).Err()
}

// DequeueDocument gets the next document to process
func (r *Redis) DequeueDocument() (string, error) {
	key := "document_processing_queue"
	// Get and remove the item with lowest score (highest priority/oldest)
	result := r.client.ZPopMin(r.ctx, key, 1)
	if result.Err() != nil {
		return "", result.Err()
	}
	
	values := result.Val()
	if len(values) == 0 {
		return "", redis.Nil
	}
	
	return values[0].Member.(string), nil
}

// Rate limiting

// CheckRateLimit implements a sliding window rate limiter
func (r *Redis) CheckRateLimit(identifier string, limit int, window time.Duration) (bool, int, error) {
	key := fmt.Sprintf("rate_limit:%s", identifier)
	now := time.Now()
	windowStart := now.Add(-window).Unix()

	// Remove old entries
	r.client.ZRemRangeByScore(r.ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// Count current entries
	count, err := r.client.ZCard(r.ctx, key).Result()
	if err != nil {
		return false, 0, err
	}

	if count >= int64(limit) {
		return false, int(count), nil
	}

	// Add current request
	r.client.ZAdd(r.ctx, key, redis.Z{
		Score:  float64(now.Unix()),
		Member: now.UnixNano(),
	})

	// Set expiration
	r.client.Expire(r.ctx, key, window)

	return true, int(count) + 1, nil
}

// Component status caching

// SetComponentStatus caches component operational status
func (r *Redis) SetComponentStatus(componentID string, status string) error {
	key := fmt.Sprintf("component_status:%s", componentID)
	return r.client.Set(r.ctx, key, status, 5*time.Minute).Err()
}

// GetComponentStatus retrieves cached component status
func (r *Redis) GetComponentStatus(componentID string) (string, error) {
	key := fmt.Sprintf("component_status:%s", componentID)
	return r.client.Get(r.ctx, key).Result()
}

// Site activity tracking

// IncrementSiteActivity tracks site activity for analytics
func (r *Redis) IncrementSiteActivity(siteID, activityType string) error {
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("site_activity:%s:%s:%s", siteID, today, activityType)
	
	// Increment counter
	if err := r.client.Incr(r.ctx, key).Err(); err != nil {
		return err
	}
	
	// Set expiration to 30 days
	return r.client.Expire(r.ctx, key, 30*24*time.Hour).Err()
}

// GetSiteActivity retrieves activity counts
func (r *Redis) GetSiteActivity(siteID string, date string, activityType string) (int64, error) {
	key := fmt.Sprintf("site_activity:%s:%s:%s", siteID, date, activityType)
	return r.client.Get(r.ctx, key).Int64()
}

// Health check
func (r *Redis) HealthCheck() error {
	return r.client.Ping(r.ctx).Err()
}

// Close closes the Redis connection
func (r *Redis) Close() error {
	return r.client.Close()
}