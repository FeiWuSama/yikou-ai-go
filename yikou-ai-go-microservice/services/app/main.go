package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"yikou-ai-go-microservice/services/screenshot/logic/download"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/kitex/client"
	kServer "github.com/cloudwego/kitex/server"
	"github.com/hertz-contrib/registry/nacos"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"yikou-ai-go-microservice/services/app/cache"
	"yikou-ai-go-microservice/services/app/config"
	"yikou-ai-go-microservice/services/app/core"
	"yikou-ai-go-microservice/services/app/core/messagehandler"
	"yikou-ai-go-microservice/services/app/core/parser"
	"yikou-ai-go-microservice/services/app/core/saver"
	"yikou-ai-go-microservice/services/app/dal"
	"yikou-ai-go-microservice/services/app/handler"
	"yikou-ai-go-microservice/services/app/kitex_gen/chathistory/chathistoryservice"
	appLogic "yikou-ai-go-microservice/services/app/logic/app"
	chatHistoryLogic "yikou-ai-go-microservice/services/app/logic/chathistory"
	"yikou-ai-go-microservice/services/app/router"
	"yikou-ai-go-microservice/services/user/kitex_gen/userservice"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp", ":9092")
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.InitConfig()
	db := dal.InitDB(cfg)
	redisClient := dal.InitRedis(cfg)

	clientConfig := constant.ClientConfig{
		NamespaceId:         cfg.Nacos.NamespaceId,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              cfg.Nacos.LogDir,
		CacheDir:            cfg.Nacos.CacheDir,
		LogLevel:            cfg.Nacos.LogLevel,
		Username:            cfg.Nacos.Username,
		Password:            cfg.Nacos.Password,
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      cfg.Nacos.Host,
			ContextPath: "/nacos",
			Port:        uint64(cfg.Nacos.Port),
			Scheme:      "http",
		},
	}

	nacosClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		log.Fatalf("创建 Nacos 客户端失败: %v", err)
	}

	nacosRegistry := nacos.NewNacosRegistry(nacosClient)

	chatHistorySvc := chatHistoryLogic.NewChatHistoryService(db)
	chatHistoryRpcHandler := handler.NewChatHistoryServiceImpl(chatHistorySvc, redisClient, db)

	kitexServer := kServer.NewServer(kServer.WithServiceAddr(addr))
	if err := chathistoryservice.RegisterService(kitexServer, chatHistoryRpcHandler); err != nil {
		log.Fatalf("Failed to register ChatHistoryService: %v", err)
	}

	go func() {
		fmt.Println("App Service Kitex Server starting on :9092...")
		if err := kitexServer.Run(); err != nil {
			log.Printf("Kitex server error: %v", err)
		}
	}()

	hertzServer := server.Default(
		server.WithHostPorts(":8082"),
		server.WithRegistry(nacosRegistry, &registry.Info{
			ServiceName: "app-service",
			Addr:        utils.NewNetAddr("tcp", "localhost:8082"),
			Weight:      10,
			Tags:        map[string]string{"env": "dev", "version": "1.0.0"},
		}),
	)

	initHertzRoutes(hertzServer, db, redisClient, cfg)

	go func() {
		fmt.Println("App Service Hertz Server starting on :8082...")
		hertzServer.Spin()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down App Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	hertzServer.Shutdown(ctx)
	kitexServer.Stop()

	fmt.Println("App Service stopped")
}

func initHertzRoutes(h *server.Hertz, db *gorm.DB, redisClient *redis.Client, cfg *config.Config) {
	cacheManager := cache.InitCacheManager(redisClient)

	userRpcClient := userservice.MustNewClient("user-service", client.WithHostPorts("127.0.0.1:9090"))

	chatHistorySvc := chatHistoryLogic.NewChatHistoryService(db)

	codeParserExecutor := parser.NewCodeParserExecutor()
	codeFileSaverExecutor := saver.NewCodeFileSaverExecutor()
	aiCodeGenFacade := core.NewYiKouAiCodegenFacade(codeParserExecutor, codeFileSaverExecutor)

	streamHandlerExecutor := messagehandler.NewStreamHandlerExecutor(chatHistorySvc, nil)

	appSvc := appLogic.NewAppService(aiCodeGenFacade, chatHistorySvc, streamHandlerExecutor, db)

	projectDownloadSvc := download.NewProjectDownloadService()

	appHandler := handler.NewAppHandler(appSvc, userRpcClient, chatHistorySvc, projectDownloadSvc)

	router.CustomizedRegister(h, db, redisClient, cacheManager, userRpcClient, appHandler)
}
