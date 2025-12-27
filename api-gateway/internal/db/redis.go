package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/trackfy/api-gateway/internal/models"
)

type RedisDB struct {
	client *redis.Client
}

// Prefijos para las claves de Redis
const (
	PrefixSession      = "session:"
	PrefixUserSessions = "user_sessions:"
	PrefixRateLimit    = "rate_limit:"
	PrefixConvCache    = "conv_cache:"
	PrefixUserCache    = "user_cache:"
)

func NewRedisDB(url, password string, db int) (*RedisDB, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Info().Str("addr", url).Msg("[RedisDB] Connected successfully")
	return &RedisDB{client: client}, nil
}

func (r *RedisDB) Close() error {
	return r.client.Close()
}

// ==================== SESSIONS ====================

type SessionData struct {
	SessionID  uuid.UUID `json:"session_id"`
	UserID     uuid.UUID `json:"user_id"`
	DeviceID   string    `json:"device_id"`
	DeviceType string    `json:"device_type"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func (r *RedisDB) StoreSession(ctx context.Context, tokenHash string, session *SessionData, ttl time.Duration) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()

	// Guardar sesión por token hash
	pipe.Set(ctx, PrefixSession+tokenHash, data, ttl)

	// Añadir a la lista de sesiones del usuario
	pipe.SAdd(ctx, PrefixUserSessions+session.UserID.String(), tokenHash)
	pipe.Expire(ctx, PrefixUserSessions+session.UserID.String(), 7*24*time.Hour)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisDB) GetSession(ctx context.Context, tokenHash string) (*SessionData, error) {
	data, err := r.client.Get(ctx, PrefixSession+tokenHash).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *RedisDB) DeleteSession(ctx context.Context, tokenHash string, userID uuid.UUID) error {
	pipe := r.client.Pipeline()
	pipe.Del(ctx, PrefixSession+tokenHash)
	pipe.SRem(ctx, PrefixUserSessions+userID.String(), tokenHash)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisDB) DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	// Obtener todas las sesiones del usuario
	tokenHashes, err := r.client.SMembers(ctx, PrefixUserSessions+userID.String()).Result()
	if err != nil {
		return err
	}

	if len(tokenHashes) == 0 {
		return nil
	}

	// Borrar todas las sesiones
	pipe := r.client.Pipeline()
	for _, hash := range tokenHashes {
		pipe.Del(ctx, PrefixSession+hash)
	}
	pipe.Del(ctx, PrefixUserSessions+userID.String())

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisDB) UpdateSessionActivity(ctx context.Context, tokenHash string) error {
	// Actualizar TTL de la sesión
	return r.client.Expire(ctx, PrefixSession+tokenHash, 15*time.Minute).Err()
}

// ==================== RATE LIMITING ====================

func (r *RedisDB) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	fullKey := PrefixRateLimit + key

	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, fullKey)
	pipe.Expire(ctx, fullKey, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	count := int(incr.Val())
	return count <= limit, count, nil
}

// ==================== CONVERSATION CACHE ====================

type ConversationCache struct {
	Messages []models.Message `json:"messages"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (r *RedisDB) CacheConversation(ctx context.Context, conversationID uuid.UUID, messages []models.Message) error {
	cache := ConversationCache{
		Messages:  messages,
		UpdatedAt: time.Now(),
	}
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, PrefixConvCache+conversationID.String(), data, 30*time.Minute).Err()
}

func (r *RedisDB) GetCachedConversation(ctx context.Context, conversationID uuid.UUID) ([]models.Message, error) {
	data, err := r.client.Get(ctx, PrefixConvCache+conversationID.String()).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var cache ConversationCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return cache.Messages, nil
}

func (r *RedisDB) InvalidateConversationCache(ctx context.Context, conversationID uuid.UUID) error {
	return r.client.Del(ctx, PrefixConvCache+conversationID.String()).Err()
}

// ==================== USER CACHE ====================

func (r *RedisDB) CacheUser(ctx context.Context, user *models.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, PrefixUserCache+user.ID.String(), data, 5*time.Minute).Err()
}

func (r *RedisDB) GetCachedUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	data, err := r.client.Get(ctx, PrefixUserCache+userID.String()).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var user models.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// ==================== MEMORIA CORTA FY ====================

// FyMemory almacena el contexto corto de la conversación para Fy
type FyMemory struct {
	UserID         uuid.UUID        `json:"user_id"`
	ConversationID uuid.UUID        `json:"conversation_id"`
	RecentMessages []FyMemoryMessage `json:"recent_messages"`
	LastIntent     string           `json:"last_intent"`
	LastMood       string           `json:"last_mood"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

type FyMemoryMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (r *RedisDB) StoreFyMemory(ctx context.Context, userID, conversationID uuid.UUID, messages []FyMemoryMessage, intent, mood string) error {
	memory := FyMemory{
		UserID:         userID,
		ConversationID: conversationID,
		RecentMessages: messages,
		LastIntent:     intent,
		LastMood:       mood,
		UpdatedAt:      time.Now(),
	}

	data, err := json.Marshal(memory)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("fy_memory:%s:%s", userID.String(), conversationID.String())
	return r.client.Set(ctx, key, data, 30*time.Minute).Err()
}

func (r *RedisDB) GetFyMemory(ctx context.Context, userID, conversationID uuid.UUID) (*FyMemory, error) {
	key := fmt.Sprintf("fy_memory:%s:%s", userID.String(), conversationID.String())
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var memory FyMemory
	if err := json.Unmarshal(data, &memory); err != nil {
		return nil, err
	}

	return &memory, nil
}
