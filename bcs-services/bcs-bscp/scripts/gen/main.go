package main

import (
	"gorm.io/gen"

	"bscp.io/pkg/dal/table"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:       "./pkg/dal/gen",
		Mode:          gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable: true,
	})

	// 需要 Gen 的模型这里添加
	g.ApplyBasic(
		table.Audit{},
		table.TemplateSpace{},
	)

	g.Execute()
}
