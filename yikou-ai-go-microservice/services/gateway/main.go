package main

import (
	"context"
	"fmt"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/cors"
	"github.com/hertz-contrib/registry/nacos"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	common "yikou-ai-go-microservice/pkg/commonapi"
	pkg "yikou-ai-go-microservice/pkg/errors"
	"yikou-ai-go-microservice/services/gateway/config"
	"yikou-ai-go-microservice/services/gateway/proxy"
	"yikou-ai-go-microservice/services/gateway/router"
)

func CustomRecoveryHandler(ctx context.Context, c *app.RequestContext, err interface{}, stack []byte) {
	logger.Errorf("panic recovered: %v\n%s", err, stack)
	c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.SystemError.WithMessage(fmt.Sprintf("%v", err))))
	c.Abort()
}

func main() {
	cfg := config.InitConfig()

	serviceDiscovery, err := proxy.NewServiceDiscovery(&cfg.Nacos)
	if err != nil {
		log.Fatalf("初始化服务发现失败: %v", err)
	}

	reverseProxy := proxy.NewReverseProxy(serviceDiscovery, cfg.Proxy.Routes)

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

	hertzServer := server.Default(
		server.WithHostPorts(fmt.Sprintf(":%d", cfg.Server.Port)),
		server.WithRegistry(nacosRegistry, &registry.Info{
			ServiceName: "gateway-service",
			Addr:        utils.NewNetAddr("tcp", fmt.Sprintf("localhost:%d", cfg.Server.Port)),
			Weight:      10,
			Tags:        map[string]string{"env": "dev", "version": "1.0.0"},
		}),
	)

	// 全局异常处理
	hertzServer.Use(recovery.Recovery(recovery.WithRecoveryHandler(CustomRecoveryHandler)))
	// 处理跨域问题
	hertzServer.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
	// 注

	router.RegisterRoutes(hertzServer, reverseProxy)

	go func() {
		fmt.Printf("Gateway Service starting on :%d...\n", cfg.Server.Port)
		hertzServer.Spin()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down Gateway Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	hertzServer.Shutdown(ctx)

	fmt.Println("Gateway Service stopped")
}
