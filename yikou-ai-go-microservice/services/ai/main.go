package main

import (
	"context"
	"fmt"
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
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"yikou-ai-go-microservice/services/ai/agent"
	main2 "yikou-ai-go-microservice/services/ai/agent/handler"
	"yikou-ai-go-microservice/services/ai/aitools"
	"yikou-ai-go-microservice/services/ai/config"
	kitex_gen "yikou-ai-go-microservice/services/ai/kitex_gen/aiservice"
	"yikou-ai-go-microservice/services/ai/llm"
	"yikou-ai-go-microservice/services/app/kitex_gen/chathistory/chathistoryservice"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp", ":9093")
	if err != nil {
		log.Fatal(err)
	}

	// 初始化配置
	cfg := config.InitConfig()

	// 配置 Nacos 客户端
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

	// 配置 Nacos 服务器
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      cfg.Nacos.Host,
			ContextPath: "/nacos",
			Port:        uint64(cfg.Nacos.Port),
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

	// 初始化 Redis 客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试 Redis 连接
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis 连接失败: %v", err)
	}

	// 初始化 ChatModel
	chatModel := llm.NewChatModel(cfg)
	reasoningChatModel := llm.NewReasoningChatModel(cfg)

	// 初始化 ToolManager
	toolManager, err := aitools.NewToolManager()
	if err != nil {
		log.Fatalf("初始化 ToolManager 失败: %v", err)
	}

	// 初始化 ChatHistory RPC 客户端
	chatHistoryRpcAddr := cfg.RPC.ChatHistoryService
	if chatHistoryRpcAddr == "" {
		chatHistoryRpcAddr = "127.0.0.1:9092"
	}
	chatHistoryRpcClient := chathistoryservice.MustNewClient("chathistory-service", client.WithHostPorts(chatHistoryRpcAddr))

	// 初始化 Agent Factory
	codeGenAgentFactory := agent.NewCodeGenAgentFactory(chatModel, reasoningChatModel, redisClient, toolManager, chatHistoryRpcClient)
	chatSummaryAgentFactory := agent.NewChatSummaryAgentFactory(chatModel)
	codeQualityCheckAgentFactory := agent.NewCodeQualityCheckAgentFactory(chatModel)
	codeGenTypeRoutingFactory := agent.NewCodeGenTypeRoutingAgentFactory(chatModel)

	// 初始化 Handler 并注入依赖
	aiHandler := main2.NewAiServiceImpl(
		codeGenAgentFactory,
		chatSummaryAgentFactory,
		codeQualityCheckAgentFactory,
		codeGenTypeRoutingFactory,
		redisClient,
	)

	// 创建 Kitex Server
	kiteXServer := kitex_gen.NewServer(aiHandler, kServer.WithServiceAddr(addr))

	// 启动 Kitex Server
	go func() {
		fmt.Println("AI Service Kitex Server starting on :9093...")
		err := kiteXServer.Run()
		if err != nil {
			log.Println(err.Error())
		}
	}()

	// 启动 Hertz HTTP Server 并注册到 Nacos
	hertzServer := server.Default(
		server.WithHostPorts(":8083"),
		server.WithRegistry(nacosRegistry, &registry.Info{
			ServiceName: "ai-service",
			Addr:        utils.NewNetAddr("tcp", "localhost:8083"),
			Weight:      10,
			Tags:        map[string]string{"env": "dev", "version": "1.0.0"},
		}),
	)

	// 注册路由...
	go func() {
		fmt.Println("AI Service Hertz Server starting on :8083...")
		hertzServer.Spin()
	}()

	// 等待信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down AI Service...")

	// 优雅关闭 Hertz
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	hertzServer.Shutdown(ctx)

	// 优雅关闭 Kitex
	kiteXServer.Stop()

	// 关闭 Redis 连接
	if err := redisClient.Close(); err != nil {
		log.Printf("关闭 Redis 连接失败: %v", err)
	}

	fmt.Println("AI Service stopped")
}
