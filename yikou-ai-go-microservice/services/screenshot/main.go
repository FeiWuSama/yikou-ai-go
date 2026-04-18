package main

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app/server"
	kServer "github.com/cloudwego/kitex/server"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
	"yikou-ai-go-microservice/services/screenshot/config"
	"yikou-ai-go-microservice/services/screenshot/handler"
	screenshot "yikou-ai-go-microservice/services/screenshot/kitex_gen/screenshotservice"
	logic "yikou-ai-go-microservice/services/screenshot/logic"
	"yikou-ai-go-microservice/services/screenshot/manager"

	"github.com/tencentyun/cos-go-sdk-v5"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp", ":9091")
	if err != nil {
		log.Fatal(err)
	}

	// 初始化配置
	cfg := config.InitConfig()

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
		err := kiteXServer.Run()

		if err != nil {
			log.Println(err.Error())
		}
	}()

	// 启动 Hertz
	hertzServer := server.Default(server.WithHostPorts(":8081"))
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
	kiteXServer.Stop()
}
