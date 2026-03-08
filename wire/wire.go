//go:build wireinject

package wire

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/wire"
	"github.com/hertz-contrib/cors"
	"github.com/hertz-contrib/swagger"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"strconv"
	"time"
	"workspace-yikou-ai-go/biz/ai/agent"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/ai/skill"
	"workspace-yikou-ai-go/biz/core"
	"workspace-yikou-ai-go/biz/core/parser"
	"workspace-yikou-ai-go/biz/core/saver"
	"workspace-yikou-ai-go/biz/dal"
	appHandler "workspace-yikou-ai-go/biz/handler/app"
	chatHistoryHandler "workspace-yikou-ai-go/biz/handler/chathistory"
	static "workspace-yikou-ai-go/biz/handler/static"
	userHandler "workspace-yikou-ai-go/biz/handler/user"
	"workspace-yikou-ai-go/biz/model/api/common"
	"workspace-yikou-ai-go/biz/router"
	application "workspace-yikou-ai-go/biz/service/app"
	"workspace-yikou-ai-go/biz/service/chathistory"
	user "workspace-yikou-ai-go/biz/service/user"
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
)

// Service依赖
var serviceSet = wire.NewSet(
	core.NewYiKouAiCodegenFacade,
	application.NewAppService,
	wire.Bind(new(application.IAppService), new(*application.AppService)),
	user.NewUserService,
	wire.Bind(new(user.IUserService), new(*user.UserService)),
	chathistory.NewChatHistoryService,
	wire.Bind(new(chathistory.IChatHistoryService), new(*chathistory.ChatHistoryService)),
)

// Handler依赖
var handlerSet = wire.NewSet(
	appHandler.NewAppHandler,
	userHandler.NewUserHandler,
	chatHistoryHandler.NewChatHistoryHandler,
	static.NewStaticResourceHandler,
)

func CustomRecoveryHandler(ctx context.Context, c *app.RequestContext, err interface{}, stack []byte) {
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
	db *gorm.DB,
	redisClient *redis.Client,
) *server.Hertz {
	basePath := serverConfig.Server.ContextPath
	// 动态补充swagger前缀
	docs.SwaggerInfo.BasePath = basePath
	// 初始化swagger路径
	swaggerPath := fmt.Sprintf("http://localhost:%d%s/swagger/doc.json", serverConfig.Server.Port, basePath)
	url := swagger.URL(swaggerPath)
	h := server.New(
		server.WithHostPorts(":"+strconv.Itoa(serverConfig.Server.Port)),
		server.WithBasePath(serverConfig.Server.ContextPath),
	)
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
	router.CustomizedRegister(h, db, redisClient, appHandler, userHandler, chatHistoryHandler, staticResourceHandler, url)
	return h
}

// 初始化所有依赖（依赖图）
func InitializeApp() (*server.Hertz, error) {
	panic(wire.Build(
		configSet,
		dbSet,
		serviceSet,
		handlerSet,
		InitServer,
		llm.NewBaseAiChatModel,
		llm.NewChatModel,
		skill.NewYiKouAiCodegenService,
		wire.Bind(new(skill.IYiKouAiCodegenService), new(*skill.YiKouAiCodegenService)),
		parser.NewCodeParserExecutor,
		saver.NewCodeFileSaverExecutor,
		agent.NewCodeGenAgentFactory,
	))
}
