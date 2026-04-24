package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	common "yikou-ai-go-microservice/pkg/commonapi"
	pkg "yikou-ai-go-microservice/pkg/errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"yikou-ai-go-microservice/services/gateway/config"
)

// ServiceDiscovery 服务发现结构体
// 基于 Nacos 实现服务实例的发现与缓存
type ServiceDiscovery struct {
	nacosClient naming_client.INamingClient // Nacos 命名服务客户端
	cache       sync.Map                    // 服务实例缓存（并发安全）
	cacheTTL    time.Duration               // 缓存过期时间
}

// ServiceInstance 服务实例信息
type ServiceInstance struct {
	Host string // 服务主机地址
	Port int    // 服务端口
}

// NewServiceDiscovery 创建服务发现实例
// 初始化 Nacos 客户端连接，用于后续的服务实例查询
func NewServiceDiscovery(cfg *config.NacosConfig) (*ServiceDiscovery, error) {
	// 配置 Nacos 客户端参数
	clientConfig := constant.ClientConfig{
		NamespaceId:         cfg.NamespaceId, // 命名空间ID
		TimeoutMs:           5000,            // 请求超时时间
		NotLoadCacheAtStart: true,            // 启动时不加载缓存
		LogDir:              cfg.LogDir,      // 日志目录
		CacheDir:            cfg.CacheDir,    // 缓存目录
		LogLevel:            cfg.LogLevel,    // 日志级别
		Username:            cfg.Username,    // 用户名
		Password:            cfg.Password,    // 密码
	}

	// 配置 Nacos 服务端地址
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      cfg.Host,
			ContextPath: "/nacos",
			Port:        uint64(cfg.Port),
			Scheme:      "http",
		},
	}

	// 创建 Nacos 命名服务客户端
	nacosClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("创建 Nacos 客户端失败: %w", err)
	}

	return &ServiceDiscovery{
		nacosClient: nacosClient,
		cacheTTL:    30 * time.Second, // 缓存30秒过期
	}, nil
}

// GetServiceInstance 获取服务实例
// 优先从缓存获取，缓存过期或不存在则从 Nacos 查询
func (sd *ServiceDiscovery) GetServiceInstance(serviceName string) (*ServiceInstance, error) {
	cacheKey := "service:" + serviceName

	// 尝试从缓存获取
	if cached, ok := sd.cache.Load(cacheKey); ok {
		if entry, ok := cached.(cacheEntry); ok {
			// 检查缓存是否过期
			if time.Since(entry.timestamp) < sd.cacheTTL {
				return entry.instance, nil
			}
		}
	}

	// 缓存不存在或已过期，从 Nacos 查询
	instance, err := sd.selectInstance(serviceName)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	sd.cache.Store(cacheKey, cacheEntry{
		instance:  instance,
		timestamp: time.Now(),
	})

	return instance, nil
}

// cacheEntry 缓存条目
type cacheEntry struct {
	instance  *ServiceInstance // 服务实例
	timestamp time.Time        // 缓存时间戳
}

// selectInstance 从 Nacos 选择一个健康的服务实例
// 当前使用简单策略：选择第一个健康实例
func (sd *ServiceDiscovery) selectInstance(serviceName string) (*ServiceInstance, error) {
	// 查询健康的实例列表
	instances, err := sd.nacosClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		GroupName:   "DEFAULT_GROUP",
		Clusters:    []string{"DEFAULT"},
		HealthyOnly: true, // 只查询健康实例
	})
	if err != nil {
		return nil, fmt.Errorf("获取服务实例失败: %w", err)
	}

	// 检查是否有可用实例
	if len(instances) == 0 {
		return nil, fmt.Errorf("没有可用的服务实例: %s", serviceName)
	}

	// 选择第一个实例
	instance := instances[0]
	return &ServiceInstance{
		Host: instance.Ip,
		Port: int(instance.Port),
	}, nil
}

// ReverseProxy 反向代理结构体
// 负责将请求转发到后端微服务
type ReverseProxy struct {
	discovery *ServiceDiscovery    // 服务发现客户端
	routes    []config.RouteConfig // 路由配置列表
	client    *http.Client         // HTTP 客户端
}

// NewReverseProxy 创建反向代理实例
func NewReverseProxy(discovery *ServiceDiscovery, routes []config.RouteConfig) *ReverseProxy {
	return &ReverseProxy{
		discovery: discovery,
		routes:    routes,
		client: &http.Client{
			Timeout: 30 * time.Second, // 请求超时30秒
		},
	}
}

// Handler 处理所有进入网关的请求
// 根据路径前缀匹配路由配置，转发到对应的后端服务
func (rp *ReverseProxy) Handler(ctx context.Context, c *app.RequestContext) {
	path := string(c.URI().Path())

	// 遍历路由配置，查找匹配的路由
	for _, route := range rp.routes {
		if strings.HasPrefix(path, route.Path) {
			rp.proxyRequest(ctx, c, route)
			return
		}
	}

	// 没有匹配的路由，返回404
	c.String(consts.StatusNotFound, "Route not found")
}

// proxyRequest 执行代理请求
// 将客户端请求转发到后端服务，并返回响应
func (rp *ReverseProxy) proxyRequest(ctx context.Context, c *app.RequestContext, route config.RouteConfig) {
	// 1. 通过服务发现获取后端服务实例
	instance, err := rp.discovery.GetServiceInstance(route.Service)
	if err != nil {
		log.Printf("获取服务实例失败: %v", err)
		c.String(consts.StatusBadGateway, "Service unavailable: %s", route.Service)
		return
	}

	// 2. 构建目标路径
	originalPath := string(c.URI().Path())
	targetPath := originalPath

	// 路径重写逻辑
	// StripPrefix: 去掉指定的前缀（如 /api）
	// StripPath: 去掉整个路由匹配路径
	if route.StripPrefix != "" {
		// 只去掉指定前缀，保留服务路径
		// 例如: /api/app/add -> /app/add
		targetPath = strings.TrimPrefix(originalPath, route.StripPrefix)
	} else if route.StripPath {
		// 去掉整个路由匹配路径
		// 例如: /api/app/add -> /add
		targetPath = strings.TrimPrefix(originalPath, route.Path)
		if route.RewritePath != "" {
			targetPath = route.RewritePath + targetPath
		}
	}

	// 3. 构建目标 URL
	targetURL := fmt.Sprintf("http://%s:%d%s", instance.Host, instance.Port, targetPath)
	// 保留查询参数
	if string(c.URI().QueryString()) != "" {
		targetURL += "?" + string(c.URI().QueryString())
	}

	// 4. 创建转发请求
	var body io.Reader
	if c.Request.Body() != nil {
		body = c.Request.BodyStream()
	}

	req, err := http.NewRequestWithContext(ctx, string(c.Method()), targetURL, body)
	if err != nil {
		hlog.Errorf("创建请求失败: %v", err)
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.SystemError.WithMessage("请求失败")))
		return
	}

	// 5. 复制原始请求头
	c.Request.Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	// 6. 发送请求到后端服务
	resp, err := rp.client.Do(req)
	if err != nil {
		hlog.Errorf("代理请求失败: %v", err)
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.SystemError.WithMessage("请求失败")))
		return
	}
	defer resp.Body.Close()

	// 7. 复制响应头到客户端
	respHeader := &c.Response.Header
	for k, v := range resp.Header {
		for _, vv := range v {
			respHeader.Add(k, vv)
		}
	}

	// 8. 设置响应状态码
	c.Response.SetStatusCode(resp.StatusCode)

	// 9. 读取并设置响应体
	// 使用 io.ReadAll 读取完整响应体，避免流式传输时的连接关闭问题
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		hlog.Errorf("读取响应体失败: %v", err)
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.SystemError.WithMessage("请求失败")))
		return
	}
	c.Response.SetBody(respBody)
}
