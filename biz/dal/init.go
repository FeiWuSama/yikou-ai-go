package dal

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"workspace-yikou-ai-go/biz/dal/query"
	"workspace-yikou-ai-go/config"
)

var DB *gorm.DB

func InitDB() error {
	if config.GlobalConfig == nil {
		return fmt.Errorf("配置加载失败")
	}

	dsn := config.GlobalConfig.Database.GetDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	DB = db
	return nil
}

func init() {
	if err := InitDB(); err != nil {
		panic(err)
	}
	query.SetDefault(DB)
}
