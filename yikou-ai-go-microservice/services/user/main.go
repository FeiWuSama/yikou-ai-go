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
	"yikou-ai-go-microservice/services/user/config"
	"yikou-ai-go-microservice/services/user/dal"
	"yikou-ai-go-microservice/services/user/handler"
	"yikou-ai-go-microservice/services/user/kitex_gen/userservice"
	userLogic "yikou-ai-go-microservice/services/user/logic"
	"yikou-ai-go-microservice/services/user/router"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp", ":9090")
	if err != nil {
		log.Fatal(err)
	}

	// 初始化配置
	cfg := config.InitConfig()
	// 初始化数据库
	db := dal.InitDB(cfg)
	// 初始化Redis
	redisClient := dal.InitRedis(cfg)
	// 创建用户服务实例
	userService := userLogic.NewUserService(db, redisClient)
	// 创建Handler实例并注入依赖
	rpcHandler := handler.NewUserServiceImpl(userService)
	userHandler := handler.NewUserHandler(userService)

	kitexServer := userservice.NewServer(rpcHandler,
		kServer.WithServiceAddr(addr))
	go func() {
		err := kitexServer.Run()

		if err != nil {
			log.Println(err.Error())
		}
	}()

	// 启动 Hertz
	hertzServer := server.Default(server.WithHostPorts(":8080"))
	// 注册路由...
	go func() {
		router.CustomizedRegister(hertzServer, db, redisClient, userHandler)
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
