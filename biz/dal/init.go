package dal

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"workspace-yikou-ai-go/biz/dal/query"
	"workspace-yikou-ai-go/config"
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

//func init() {
//	if err := InitDB(); err != nil {
//		panic(err)
//	}
//	query.SetDefault(DB)
//}
