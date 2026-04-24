package router

import (
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	enum "yikou-ai-go-microservice/pkg/commonenum"
	commonmiddleware "yikou-ai-go-microservice/pkg/commonmiddleware"
	"yikou-ai-go-microservice/services/app/cache"
	"yikou-ai-go-microservice/services/app/handler"
	"yikou-ai-go-microservice/services/app/middleware"
	"yikou-ai-go-microservice/services/user/kitex_gen/userservice"
)

func CustomizedRegister(
	r *server.Hertz,
	db *gorm.DB,
	redisClient *redis.Client,
	cacheManager *cache.CacheManager,
	userRpcClient userservice.Client,
	appHandler *handler.AppHandler,
) {
	appRoute := r.Group("/app")
	{
		appRoute.POST("/good/list/page/vo",
			middleware.CacheMiddleware(cacheManager, middleware.CacheMiddlewareConfig{
				CacheName:  "good_app_page",
				TTL:        5 * time.Minute,
				KeyBuilder: middleware.DefaultKeyBuilder,
				Condition:  middleware.PageCondition(10),
			}),
			appHandler.ListGoodApp,
		)

		appRoute.GET("/get/vo",
			commonmiddleware.AuthMiddleware(enum.UserRole, db, redisClient),
			appHandler.GetAppVo,
		)

		appRoute.POST("/my/list/page/vo",
			commonmiddleware.AuthMiddleware(enum.UserRole, db, redisClient),
			appHandler.ListMyApp,
		)

		appRoute.POST("/add",
			commonmiddleware.AuthMiddleware(enum.UserRole, db, redisClient),
			appHandler.AddApp,
		)

		appRoute.POST("/update",
			commonmiddleware.AuthMiddleware(enum.UserRole, db, redisClient),
			appHandler.UpdateApp,
		)

		appRoute.POST("/delete",
			commonmiddleware.AuthMiddleware(enum.UserRole, db, redisClient),
			appHandler.DeleteApp,
		)

		appRoute.GET("/chat/gen/code",
			commonmiddleware.AuthMiddleware(enum.UserRole, db, redisClient),
			middleware.RateLimitMiddleware(redisClient, userRpcClient, middleware.RateLimitConfig{
				Rate:         5,
				RateInterval: 60,
				LimitType:    middleware.RateLimitTypeUSER,
				Message:      "AI对话请求过于频繁，请稍后再试",
			}),
			appHandler.ChatToGenCode,
		)

		appRoute.POST("/chat/gen/stop",
			commonmiddleware.AuthMiddleware(enum.UserRole, db, redisClient),
			appHandler.StopStream,
		)

		appRoute.POST("/deploy",
			commonmiddleware.AuthMiddleware(enum.UserRole, db, redisClient),
			appHandler.DeployApp,
		)

		appRoute.GET("/download/:appId",
			commonmiddleware.AuthMiddleware(enum.UserRole, db, redisClient),
			appHandler.DownloadAppCode,
		)

		appRoute.POST("/admin/update",
			commonmiddleware.AuthMiddleware(enum.AdminRole, db, redisClient),
			appHandler.AdminUpdateApp,
		)

		appRoute.POST("/admin/delete",
			commonmiddleware.AuthMiddleware(enum.AdminRole, db, redisClient),
			appHandler.AdminDeleteApp,
		)

		appRoute.GET("/admin/get/vo",
			commonmiddleware.AuthMiddleware(enum.AdminRole, db, redisClient),
			appHandler.AdminGetAppVo,
		)

		appRoute.POST("/admin/list/page/vo",
			commonmiddleware.AuthMiddleware(enum.AdminRole, db, redisClient),
			appHandler.AdminListApp,
		)
	}
}
