package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
)

type MemoryStore interface {
	Write(ctx context.Context, sessionID string, msgs []*schema.Message) error
	Read(ctx context.Context, sessionID string) ([]*schema.Message, error)
}

type RedisMemoryStore struct {
	redisClient *redis.Client
	sessionID   string
	ttl         time.Duration
}

func NewRedisMemoryStore(redisClient *redis.Client, sessionID string) *RedisMemoryStore {
	return &RedisMemoryStore{
		redisClient: redisClient,
		sessionID:   sessionID,
		ttl:         24 * time.Hour,
	}
}

func (r *RedisMemoryStore) Write(ctx context.Context, sessionID string, msgs []*schema.Message) error {
	key := fmt.Sprintf("memory:%s", sessionID)
	data, err := EncodeMessagesToJSON(msgs)
	if err != nil {
		return err
	}
	return r.redisClient.Set(ctx, key, data, r.ttl).Err()
}

func (r *RedisMemoryStore) Read(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	key := fmt.Sprintf("memory:%s", sessionID)
	data, err := r.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return DecodeMessagesFromJSON(data)
}

func EncodeMessagesToJSON(msgs []*schema.Message) ([]byte, error) {
	return json.Marshal(msgs)
}

func DecodeMessagesFromJSON(data []byte) ([]*schema.Message, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var msgs []*schema.Message
	err := json.Unmarshal(data, &msgs)
	return msgs, err
}

type LimitedMemoryStore struct {
	store   MemoryStore
	maxMsgs int
}

func NewLimitedMemoryStore(store MemoryStore, maxMsgs int) *LimitedMemoryStore {
	return &LimitedMemoryStore{
		store:   store,
		maxMsgs: maxMsgs,
	}
}

func (s *LimitedMemoryStore) Write(ctx context.Context, sessionID string, msgs []*schema.Message) error {
	if len(msgs) > s.maxMsgs {
		msgs = msgs[len(msgs)-s.maxMsgs:]
	}
	return s.store.Write(ctx, sessionID, msgs)
}

func (s *LimitedMemoryStore) Read(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	return s.store.Read(ctx, sessionID)
}

type MemoryStoreHelper struct {
	store MemoryStore
}

func NewMemoryStoreHelper(store MemoryStore) *MemoryStoreHelper {
	return &MemoryStoreHelper{store: store}
}

func (h *MemoryStoreHelper) GetHistory(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	if h.store == nil {
		return nil, nil
	}
	return h.store.Read(ctx, sessionID)
}

func (h *MemoryStoreHelper) SaveHistory(ctx context.Context, sessionID string, userMsg, aiMsg string) error {
	if h.store == nil {
		return nil
	}

	history, err := h.store.Read(ctx, sessionID)
	if err != nil {
		return err
	}

	history = append(history, schema.UserMessage(userMsg), schema.AssistantMessage(aiMsg, nil))

	return h.store.Write(ctx, sessionID, history)
}
