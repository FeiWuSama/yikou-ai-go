package agent

import (
	"context"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"strconv"
	"sync"
	"time"
	"workspace-yikou-ai-go/biz/ai/aitools"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/ai/store"
	"workspace-yikou-ai-go/biz/model/enum"
	"workspace-yikou-ai-go/biz/monitor"
	chatHistory "workspace-yikou-ai-go/biz/service/chathistory"
	pkg "workspace-yikou-ai-go/pkg/errors"
)

const MaxAgentInstances = 1000

var (
	serviceCache    = cache.New(30*time.Minute, 10*time.Minute)
	instanceCount   int
	instanceCountMu sync.Mutex
)

type CodeGenAgentFactory struct {
	chatModel                   *llm.ChatModelWrapper
	reasoningStreamingChatModel *llm.ReasoningChatModelWrapper
	redisClient                 *redis.Client
	chatHistoryService          chatHistory.IChatHistoryService
	toolManager                 *aitools.ToolManager
	metricsCollector            *monitor.AiModelMetricsCollector
}

func NewCodeGenAgentFactory(chatModel *llm.ChatModelWrapper, reasoningStreamingChatModel *llm.ReasoningChatModelWrapper,
	redisClient *redis.Client, chatHistoryService chatHistory.IChatHistoryService, toolManager *aitools.ToolManager, metricsCollector *monitor.AiModelMetricsCollector) *CodeGenAgentFactory {
	serviceCache.OnEvicted(func(k string, v interface{}) {
		logger.Debugf("AI服务实例被移除，缓冲键: %v", k)
	})

	return &CodeGenAgentFactory{
		chatModel:                   chatModel,
		reasoningStreamingChatModel: reasoningStreamingChatModel,
		redisClient:                 redisClient,
		chatHistoryService:          chatHistoryService,
		toolManager:                 toolManager,
		metricsCollector:            metricsCollector,
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
	key := buildCacheKey(appId, codeGenType)

	if agent, found := serviceCache.Get(key); found {
		return agent.(*CodeGenAgent), nil
	}

	instanceCountMu.Lock()
	if instanceCount >= MaxAgentInstances {
		c.evictOldest()
	}
	instanceCountMu.Unlock()

	redisStore := store.NewRedisStore(c.redisClient, strconv.Itoa(int(appId)))
	memoryStore := store.NewRedisMemoryStore(c.redisClient, strconv.Itoa(int(appId)))
	limitedMemoryStore := store.NewLimitedMemoryStore(memoryStore, 20)
	_, err := c.chatHistoryService.LoadChatHistoryToMemory(context.Background(), appId, store.NewMemoryStoreHelper(memoryStore), 20)
	if err != nil {
		return nil, err
	}

	var agent *CodeGenAgent
	switch codeGenType {
	case enum.HtmlCodeGen, enum.MultiFileGen:
		agent = NewCodeGenAgent(c.chatModel, redisStore, limitedMemoryStore, codeGenType, c.toolManager, c.metricsCollector)
	case enum.VueCodeGen:
		agent = NewCodeGenAgent(c.reasoningStreamingChatModel, redisStore, limitedMemoryStore, codeGenType, c.toolManager, c.metricsCollector)
	default:
		return nil, pkg.SystemError.WithMessage("不支持的代码生成类型: " + enum.CodeGenTypeTextMap[codeGenType])
	}

	serviceCache.Set(key, agent, cache.DefaultExpiration)
	instanceCountMu.Lock()
	instanceCount++
	instanceCountMu.Unlock()
	return agent, nil
}

func buildCacheKey(appId int64, codeGenType enum.CodeGenTypeEnum) string {
	return strconv.Itoa(int(appId)) + "_" + string(codeGenType)
}
