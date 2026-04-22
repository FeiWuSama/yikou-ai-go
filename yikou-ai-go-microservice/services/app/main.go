package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	kServer "github.com/cloudwego/kitex/server"
	"github.com/hertz-contrib/registry/nacos"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"yikou-ai-go-microservice/services/app/config"
	"yikou-ai-go-microservice/services/app/dal"
	"yikou-ai-go-microservice/services/app/handler"
	"yikou-ai-go-microservice/services/app/kitex_gen/chathistory/chathistoryservice"
	"yikou-ai-go-microservice/services/app/logic/chathistory"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp", ":9092")
	if err != nil {
		log.Fatal(err)
	}

	// 初始化配置
	cfg := config.InitConfig()
	// 初始化数据库
	db := dal.InitDB(cfg)
	// 初始化Redis
	redisClient := dal.InitRedis(cfg)

	// 配置 Nacos 客户端
	clientConfig := constant.ClientConfig{
		NamespaceId:         "public",
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "info",
		Username:            "nacos",
		Password:            "nacos",
	}

	// 配置 Nacos 服务器
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      "localhost",
			ContextPath: "/nacos",
			Port:        8848,
			Scheme:      "http",
		},
	}

	// 创建 Nacos 命名客户端
	nacosClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		log.Fatalf("创建 Nacos 客户端失败: %v", err)
	}

	// 创建 Nacos 注册器
	nacosRegistry := nacos.NewNacosRegistry(nacosClient)

	// 初始化服务层
	chathistoryService := chathistory.NewChatHistoryService(db)

	// 创建Handler实例并注入依赖
	chatHistoryRpcHandler := handler.NewChatHistoryServiceImpl(chathistoryService, redisClient, db)

	// 创建一个 Kitex Server（支持多服务注册）
	kitexServer := kServer.NewServer(kServer.WithServiceAddr(addr))

	// 注册第一个 RPC 服务：ChatHistoryService
	if err := chathistoryservice.RegisterService(kitexServer, chatHistoryRpcHandler); err != nil {
		log.Fatalf("Failed to register ChatHistoryService: %v", err)
	}

	// 注册第二个 RPC 服务：AppService（示例）
	// 注意：需要先生成 AppService 的 Kitex 代码，然后取消下面的注释
	// appRpcHandler := handler.NewAppServiceImpl(appService, db)
	// if err := appservice.RegisterService(kitexServer, appRpcHandler); err != nil {
	//     log.Fatalf("Failed to register AppService: %v", err)
	// }

	// 启动 Kitex Server
	go func() {
		fmt.Println("App Service Kitex Server starting on :9092...")
		if err := kitexServer.Run(); err != nil {
			log.Printf("Kitex server error: %v", err)
		}
	}()

	// 启动 Hertz 并注册到 Nacos
	hertzServer := server.Default(
		server.WithHostPorts(":8082"),
		server.WithRegistry(nacosRegistry, &registry.Info{
			ServiceName: "app-service",
			Addr:        utils.NewNetAddr("tcp", "localhost:8082"),
			Weight:      10,
			Tags:        map[string]string{"env": "dev", "version": "1.0.0"},
		}),
	)

	// 注册路由...
	go func() {
		fmt.Println("App Service Hertz Server starting on :8082...")
		hertzServer.Spin()
	}()

	// 等待信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down App Service...")

	// 优雅关闭 Hertz
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	hertzServer.Shutdown(ctx)

	// 优雅关闭 Kitex
	kitexServer.Stop()

	fmt.Println("App Service stopped")
}
