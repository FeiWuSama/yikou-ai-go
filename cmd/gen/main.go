package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"workspace-yikou-ai-go/config"
)

func main() {
	if err := config.InitConfig(); err != nil {
		panic(fmt.Errorf("init config fail: %w", err))
	}

	dsn := config.GlobalConfig.Database.GetDSN()
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(fmt.Errorf("connect db fail: %w", err))
	}

	g := gen.NewGenerator(gen.Config{
		OutPath:      "./biz/dal/query",
		ModelPkgPath: "./model",
		Mode: gen.WithoutContext |
			gen.WithDefaultQuery |
			gen.WithQueryInterface,
	})

	g.UseDB(db)

	g.ApplyBasic(g.GenerateAllTable()...)

	g.Execute()
}
