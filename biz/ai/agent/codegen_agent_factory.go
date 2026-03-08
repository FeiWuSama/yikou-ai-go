package agent

import (
	"context"
	"github.com/bytedance/gopkg/util/logger"
	"strconv"
	"sync"
	"time"
	chatHistory "workspace-yikou-ai-go/biz/service/chathistory"

	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/ai/store"
	"workspace-yikou-ai-go/biz/model/enum"
	pkg "workspace-yikou-ai-go/pkg/errors"
)

const MaxAgentInstances = 1000

var (
	serviceCache    = cache.New(30*time.Minute, 10*time.Minute)
	instanceCount   int
	instanceCountMu sync.Mutex
)

type CodeGenAgentFactory struct {
	chatModel          *llm.BaseAiChatModel
	redisClient        *redis.Client
	chatHistoryService chatHistory.IChatHistoryService
}

func NewCodeGenAgentFactory(chatModel *llm.BaseAiChatModel, redisClient *redis.Client, chatHistoryService chatHistory.IChatHistoryService) *CodeGenAgentFactory {
	serviceCache.OnEvicted(func(k string, v interface{}) {
		logger.Debugf("AI服务实例被移除，appId: %v", k)
	})
	return &CodeGenAgentFactory{
		chatModel:          chatModel,
		redisClient:        redisClient,
		chatHistoryService: chatHistoryService,
	}
}

func (c CodeGenAgentFactory) evictOldest() {
	items := serviceCache.Items()
	oldestKey := ""
	var oldestExpiration int64

	for k, item := range items {
		if item.Expiration == 0 {
			continue
		}
		if oldestKey == "" || item.Expiration < oldestExpiration {
			oldestExpiration = item.Expiration
			oldestKey = k
		}
	}
	if oldestKey != "" {
		serviceCache.Delete(oldestKey)
		instanceCountMu.Lock()
		instanceCount--
		instanceCountMu.Unlock()
	}
}

func (c CodeGenAgentFactory) GetCodeGenAgent(appId int64, codeGenType enum.CodeGenTypeEnum) (*CodeGenAgent, error) {
	key := strconv.Itoa(int(appId))

	if agent, found := serviceCache.Get(key); found {
		return agent.(*CodeGenAgent), nil
	}

	instanceCountMu.Lock()
	if instanceCount >= MaxAgentInstances {
		c.evictOldest()
	}
	instanceCountMu.Unlock()

	redisStore := store.NewRedisStore(c.redisClient, key)
	memoryStore := store.NewRedisMemoryStore(c.redisClient, key)
	limitedMemoryStore := store.NewLimitedMemoryStore(memoryStore, 20)
	_, err := c.chatHistoryService.LoadChatHistoryToMemory(context.Background(), appId, store.NewMemoryStoreHelper(memoryStore), 20)
	if err != nil {
		return nil, err
	}

	var agent *CodeGenAgent
	switch codeGenType {
	case enum.HtmlCodeGen:
		agent = NewCodeGenAgent(c.chatModel, redisStore, limitedMemoryStore)
	case enum.MultiFileGen:
		agent = NewCodeGenAgent(c.chatModel, redisStore, limitedMemoryStore)
	default:
		return nil, pkg.SystemError.WithMessage("不支持的代码生成类型: " + enum.CodeGenTypeTextMap[codeGenType])
	}

	serviceCache.Set(key, agent, cache.DefaultExpiration)
	instanceCountMu.Lock()
	instanceCount++
	instanceCountMu.Unlock()
	return agent, nil
}
