package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store provides Redis/Valkey access for ephemeral state
type Store struct {
	client *redis.Client
}

// Config holds Redis connection configuration
type Config struct {
	Addr     string
	Password string
	DB       int
}

// NewStore creates a new Redis store
func NewStore(ctx context.Context, cfg Config) (*Store, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Store{client: client}, nil
}

// Close closes the Redis connection
func (s *Store) Close() error {
	return s.client.Close()
}

// Client returns the underlying Redis client
func (s *Store) Client() *redis.Client {
	return s.client
}

// AgentPresence manages agent presence tracking
type AgentPresence struct {
	store *Store
}

// NewAgentPresence creates a new agent presence manager
func NewAgentPresence(store *Store) *AgentPresence {
	return &AgentPresence{store: store}
}

// RegisterAgent registers an agent as present
func (ap *AgentPresence) RegisterAgent(ctx context.Context, tenantID, projectID, agentID string, ttl time.Duration) error {
	key := fmt.Sprintf("agents:%s:%s:%s", tenantID, projectID, agentID)
	return ap.store.client.Set(ctx, key, "present", ttl).Err()
}

// UnregisterAgent removes agent presence
func (ap *AgentPresence) UnregisterAgent(ctx context.Context, tenantID, projectID, agentID string) error {
	key := fmt.Sprintf("agents:%s:%s:%s", tenantID, projectID, agentID)
	return ap.store.client.Del(ctx, key).Err()
}

// ListAgents lists all agents for a tenant and project
func (ap *AgentPresence) ListAgents(ctx context.Context, tenantID, projectID string) ([]string, error) {
	pattern := fmt.Sprintf("agents:%s:%s:*", tenantID, projectID)
	
	keys, err := ap.store.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	
	// Extract agent IDs from keys
	agents := make([]string, 0, len(keys))
	prefix := fmt.Sprintf("agents:%s:%s:", tenantID, projectID)
	
	for _, key := range keys {
		if len(key) > len(prefix) {
			agentID := key[len(prefix):]
			agents = append(agents, agentID)
		}
	}
	
	return agents, nil
}

// IsAgentPresent checks if an agent is currently present
func (ap *AgentPresence) IsAgentPresent(ctx context.Context, tenantID, projectID, agentID string) (bool, error) {
	key := fmt.Sprintf("agents:%s:%s:%s", tenantID, projectID, agentID)
	exists, err := ap.store.client.Exists(ctx, key).Result()
	return exists > 0, err
}

// RateLimiter manages rate limiting using token buckets
type RateLimiter struct {
	store *Store
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(store *Store) *RateLimiter {
	return &RateLimiter{store: store}
}

// Allow checks if a request is allowed under rate limits
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	// Simple token bucket implementation using Redis
	current, err := rl.store.client.Get(ctx, key).Int()
	if err == redis.Nil {
		// First request, set initial count
		if err := rl.store.client.Set(ctx, key, 1, window).Err(); err != nil {
			return false, err
		}
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get rate limit: %w", err)
	}

	if current >= limit {
		return false, nil
	}

	// Increment counter
	newCount, err := rl.store.client.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to increment rate limit: %w", err)
	}

	// Set expiry if this is the first increment after key creation
	if newCount == 1 {
		rl.store.client.Expire(ctx, key, window)
	}

	return newCount <= int64(limit), nil
}
