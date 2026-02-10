package main

import (
	"workspace-yikou-ai-go/biz/dal"

	"gorm.io/gen"
	"workspace-yikou-ai-go/config"
)

func main() {
	initConfig := config.InitConfig()
	db := dal.InitDB(initConfig)
	g := gen.NewGenerator(gen.Config{
		OutPath:      "./biz/dal/query",
		ModelPkgPath: "./biz/dal/model",
		Mode: gen.WithoutContext |
			gen.WithDefaultQuery |
			gen.WithQueryInterface,
	})

	g.UseDB(db)

	g.ApplyBasic(g.GenerateAllTable()...)

	g.Execute()
}
