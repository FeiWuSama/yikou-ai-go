//go:build wireinject

package wire

import (
	"context"
	"fmt"
	"strconv"
	"time"
	workflow "workspace-yikou-ai-go/biz/handler/workflow"
	appLogic "workspace-yikou-ai-go/biz/logic/app"
	chatHistoryLogic "workspace-yikou-ai-go/biz/logic/chathistory"
	downloadLogic "workspace-yikou-ai-go/biz/logic/download"
	screenLogic "workspace-yikou-ai-go/biz/logic/screenshot"
	userLogic "workspace-yikou-ai-go/biz/logic/user"
	"workspace-yikou-ai-go/biz/service/chathistory"
	"workspace-yikou-ai-go/biz/service/download"
	"workspace-yikou-ai-go/biz/service/screenshot"
	"workspace-yikou-ai-go/biz/service/user"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	hertzConfig "github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/wire"
	"github.com/hertz-contrib/cors"
	hertzPrometheus "github.com/hertz-contrib/monitor-prometheus"
	"github.com/hertz-contrib/swagger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"workspace-yikou-ai-go/biz/ai"
	"workspace-yikou-ai-go/biz/ai/agent"
	"workspace-yikou-ai-go/biz/ai/aitools"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/cache"
	"workspace-yikou-ai-go/biz/core"
	"workspace-yikou-ai-go/biz/core/messagehandler"
	"workspace-yikou-ai-go/biz/core/parser"
	"workspace-yikou-ai-go/biz/core/saver"
	"workspace-yikou-ai-go/biz/dal"
	"workspace-yikou-ai-go/biz/graph/node"
	appHandler "workspace-yikou-ai-go/biz/handler/app"
	chatHistoryHandler "workspace-yikou-ai-go/biz/handler/chathistory"
	static "workspace-yikou-ai-go/biz/handler/static"
	userHandler "workspace-yikou-ai-go/biz/handler/user"
	"workspace-yikou-ai-go/biz/manager"
	"workspace-yikou-ai-go/biz/model/api/common"
	"workspace-yikou-ai-go/biz/monitor"
	"workspace-yikou-ai-go/biz/router"
	application "workspace-yikou-ai-go/biz/service/app"
	"workspace-yikou-ai-go/config"
	"workspace-yikou-ai-go/docs"
	pkg "workspace-yikou-ai-go/pkg/errors"
)

// 配置依赖
var configSet = wire.NewSet(
	config.InitConfig,
)

// 数据库依赖
var dbSet = wire.NewSet(
	dal.InitDB,
	dal.InitRedis,
	dal.InitCOSClient,
)

// 缓存依赖
var cacheSet = wire.NewSet(
	cache.InitCacheManager,
)

// Prometheus 监控依赖
var metricsSet = wire.NewSet(
	NewPrometheusRegistry,
	monitor.NewAiModelMetricsCollector,
)

func NewPrometheusRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

// Service依赖
var serviceSet = wire.NewSet(
	core.NewYiKouAiCodegenFacade,
	appLogic.NewAppService,
	wire.Bind(new(application.IAppService), new(*appLogic.AppService)),
	userLogic.NewUserService,
	wire.Bind(new(user.IUserService), new(*userLogic.UserService)),
	chatHistoryLogic.NewChatHistoryService,
	wire.Bind(new(chathistory.IChatHistoryService), new(*chatHistoryLogic.ChatHistoryService)),
	screenLogic.NewScreenshotService,
	wire.Bind(new(screenshot.IScreenshotService), new(*screenLogic.ScreenshotService)),
	downloadLogic.NewProjectDownloadService,
	wire.Bind(new(download.IProjectDownloadService), new(*downloadLogic.ProjectDownloadService)),
)

// Handler依赖
var handlerSet = wire.NewSet(
	appHandler.NewAppHandler,
	userHandler.NewUserHandler,
	chatHistoryHandler.NewChatHistoryHandler,
	static.NewStaticResourceHandler,
	workflow.NewWorkflowHandler,
)

var llmSet = wire.NewSet(
	llm.NewReasoningChatModel, llm.NewChatModel)

type NodeInitializer struct{}

func InitAllNodes(
	cfg *config.Config,
	chatModel *llm.ChatModelWrapper,
	cosManager *manager.CosManager,
	facade *core.YiKouAiCodegenFacade,
	metricsCollector *monitor.AiModelMetricsCollector,
) *NodeInitializer {
	node.InitImagePlanNode(chatModel, metricsCollector)
	node.InitImageCollectorPlanNode(chatModel, cosManager, metricsCollector)
	node.InitContentImageCollectorNode(cfg)
	node.InitDiagramCollectorNode(cosManager)
	node.InitLogoCollectorNode(cfg, cosManager)
	node.InitRouterNode(chatModel, metricsCollector)
	node.InitCodeQualityCheckNode(chatModel, metricsCollector)
	node.InitCodeGeneratorNode(facade)
	node.InitImageCollectorNode(cfg, chatModel, metricsCollector)
	return &NodeInitializer{}
}

func CustomRecoveryHandler(ctx context.Context, c *app.RequestContext, err interface{}, stack []byte) {
	logger.Errorf("panic recovered: %v\n%s", err, stack)
	c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.SystemError.WithMessage(fmt.Sprintf("%v", err))))
	c.Abort()
}

// Server依赖
func InitServer(
	serverConfig *config.Config,
	appHandler *appHandler.AppHandler,
	userHandler *userHandler.UserHandler,
	chatHistoryHandler *chatHistoryHandler.ChatHistoryHandler,
	staticResourceHandler *static.StaticResourceHandler,
	workflowHandler *workflow.WorkflowHandler,
	cacheManager *cache.CacheManager,
	db *gorm.DB,
	redisClient *redis.Client,
	userService user.IUserService,
	registry *prometheus.Registry,
	_ *NodeInitializer,
) *server.Hertz {
	basePath := serverConfig.Server.ContextPath
	// 动态补充swagger前缀
	docs.SwaggerInfo.BasePath = basePath
	// 初始化swagger路径
	swaggerPath := fmt.Sprintf("http://localhost:%d%s/swagger/doc.json", serverConfig.Server.Port, basePath)
	url := swagger.URL(swaggerPath)

	// 构建服务器选项
	serverOpts := []hertzConfig.Option{
		server.WithHostPorts(":" + strconv.Itoa(serverConfig.Server.Port)),
		server.WithBasePath(serverConfig.Server.ContextPath),
	}

	// Prometheus 监控指标
	if serverConfig.Server.EnableMetric {
		metricPath := serverConfig.Server.MetricPath
		if metricPath == "" {
			metricPath = "/prometheus"
		}
		metricPort := serverConfig.Server.MetricPort
		if metricPort == 0 {
			metricPort = 9090
		}
		// 使用 NewServerTracer 创建监控 tracer
		// 通过 WithRegistry 选项共享 Registry，使自定义指标与 Hertz 指标在同一端点暴露
		promTracer := hertzPrometheus.NewServerTracer(
			":"+strconv.Itoa(metricPort),
			metricPath,
			hertzPrometheus.WithRegistry(registry),
		)
		serverOpts = append(serverOpts, server.WithTracer(promTracer))
		logger.Infof("Prometheus 监控已启用，指标地址: http://localhost:%d%s", metricPort, metricPath)
	}

	h := server.New(serverOpts...)
	// 全局异常处理
	h.Use(recovery.Recovery(recovery.WithRecoveryHandler(CustomRecoveryHandler)))
	// 处理跨域问题
	h.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
	// 注册路由
	router.CustomizedRegister(h, db, redisClient, appHandler, userHandler, chatHistoryHandler, staticResourceHandler, workflowHandler, cacheManager, userService, url)
	return h
}

// 初始化所有依赖（依赖图）
func InitializeApp() (*server.Hertz, error) {
	panic(wire.Build(
		configSet,
		dbSet,
		cacheSet,
		metricsSet,
		serviceSet,
		handlerSet,
		InitServer,
		llmSet,
		ai.NewYiKouAiCodegenService,
		wire.Bind(new(ai.IYiKouAiCodegenService), new(*ai.YiKouAiCodegenService)),
		parser.NewCodeParserExecutor,
		saver.NewCodeFileSaverExecutor,
		agent.NewCodeGenAgentFactory,
		agent.NewCodeGenTypeRoutingAgentFactory,
		agent.NewChatSummaryAgentFactory,
		aitools.NewToolManager,
		messagehandler.NewStreamHandlerExecutor,
		manager.NewCosManager,
		InitAllNodes,
	))
}
