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
	"github.com/tencentyun/cos-go-sdk-v5"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
	"yikou-ai-go-microservice/pkg/manager"
	"yikou-ai-go-microservice/services/screenshot/config"
	"yikou-ai-go-microservice/services/screenshot/handler"
	screenshot "yikou-ai-go-microservice/services/screenshot/kitex_gen/screenshotservice"
	logic "yikou-ai-go-microservice/services/screenshot/logic"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp", ":9091")
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

	// 初始化COS客户端
	bucketURL, _ := url.Parse("https://" + cfg.COS.Bucket + ".cos." + cfg.COS.Region + ".myqcloud.com")
	baseURL := &cos.BaseURL{
		BucketURL: bucketURL,
	}
	cosClient := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.COS.SecretID,
			SecretKey: cfg.COS.SecretKey,
		},
	})

	// 初始化CosManager
	cosManager := manager.NewCosManager(cosClient, cfg)

	// 初始化ScreenshotService
	screenshotService := logic.NewScreenshotService(cosManager)

	// 初始化Handler并注入依赖
	screenshotHandler := handler.NewScreenshotServiceImpl(screenshotService)

	kiteXServer := screenshot.NewServer(screenshotHandler, kServer.WithServiceAddr(addr))
	go func() {
		fmt.Println("Screenshot Service Kitex Server starting on :9091...")
		err := kiteXServer.Run()

		if err != nil {
			log.Println(err.Error())
		}
	}()

	// 启动 Hertz 并注册到 Nacos
	hertzServer := server.Default(
		server.WithHostPorts(":8081"),
		server.WithRegistry(nacosRegistry, &registry.Info{
			ServiceName: "screenshot-service",
			Addr:        utils.NewNetAddr("tcp", "localhost:8081"),
			Weight:      10,
			Tags:        map[string]string{"env": "dev", "version": "1.0.0"},
		}),
	)

	// 注册路由...
	go func() {
		fmt.Println("Screenshot Service Hertz Server starting on :8081...")
		hertzServer.Spin()
	}()

	// 等待信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down Screenshot Service...")

	// 优雅关闭 Hertz
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	hertzServer.Shutdown(ctx)

	// 优雅关闭 Kitex
	kiteXServer.Stop()

	fmt.Println("Screenshot Service stopped")
}
