package main

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app/server"
	kServer "github.com/cloudwego/kitex/server"
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
		if err := kitexServer.Run(); err != nil {
			log.Printf("Kitex server error: %v", err)
		}
	}()

	// 启动 Hertz
	hertzServer := server.Default(server.WithHostPorts(":8082"))
	// 注册路由...
	go func() {
		hertzServer.Spin()
	}()

	// 等待信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭 Hertz
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	hertzServer.Shutdown(ctx)

	// 优雅关闭 Kitex
	kitexServer.Stop()
}
