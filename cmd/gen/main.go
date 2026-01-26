package main

import (
	"fmt"
	"workspace-yikou-ai-go/biz/dal"

	"gorm.io/gen"
	"workspace-yikou-ai-go/config"
)

func main() {
	if err := config.InitConfig(); err != nil {
		panic(fmt.Errorf("init config fail: %w", err))
	}

	if err := dal.InitDB(); err != nil {
		panic(fmt.Errorf("init db fail: %w", err))
	}

	g := gen.NewGenerator(gen.Config{
		OutPath:      "./biz/dal/query",
		ModelPkgPath: "./biz/dal/model",
		Mode: gen.WithoutContext |
			gen.WithDefaultQuery |
			gen.WithQueryInterface,
	})

	g.UseDB(dal.DB)

	g.ApplyBasic(g.GenerateAllTable()...)

	g.Execute()
}
