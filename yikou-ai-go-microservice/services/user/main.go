package main

import (
	"context"
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

	// 启动 Hertz 并注册到 Nacos
	hertzServer := server.Default(
		server.WithHostPorts(":8080"),
		server.WithRegistry(nacosRegistry, &registry.Info{
			ServiceName: "user-service",
			Addr:        utils.NewNetAddr("tcp", "localhost:8080"),
			Weight:      10,
			Tags:        map[string]string{"env": "dev", "version": "1.0.0"},
		}),
	)

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
