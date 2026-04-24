package router

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"yikou-ai-go-microservice/services/gateway/proxy"
)

func RegisterRoutes(h *server.Hertz, reverseProxy *proxy.ReverseProxy) {
	h.NoRoute(reverseProxy.Handler)
}
