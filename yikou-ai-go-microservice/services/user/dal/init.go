package dal

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
	"yikou-ai-go-microservice/services/user/config"
	"yikou-ai-go-microservice/services/user/dal/query"

	"github.com/redis/go-redis/v9"
	"github.com/tencentyun/cos-go-sdk-v5"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//var DB *gorm.DB

func InitDB(config *config.Config) *gorm.DB {
	if config == nil {
		panic(fmt.Errorf("配置加载失败"))
	}

	dsn := config.Database.GetDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(fmt.Errorf("数据库连接失败: %w", err))
	}
	query.SetDefault(db)

	return db
}

func InitRedis(config *config.Config) *redis.Client {
	if config == nil {
		panic(fmt.Errorf("配置加载失败"))
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	return redisClient
}

func InitCOSClient(config *config.Config) *cos.Client {
	if config == nil {
		panic(fmt.Errorf("配置加载失败"))
	}

	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.COS.Bucket, config.COS.Region))
	if err != nil {
		panic(fmt.Errorf("解析COS URL失败: %w", err))
	}

	baseURL := &cos.BaseURL{
		BucketURL: bucketURL,
	}

	client := cos.NewClient(baseURL, &http.Client{
		Timeout: 100 * time.Second,
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.COS.SecretID,
			SecretKey: config.COS.SecretKey,
		},
	})

	return client
}

//func init() {
//	if err := InitDB(); err != nil {
//		panic(err)
//	}
//	query.SetDefault(DB)
//}
