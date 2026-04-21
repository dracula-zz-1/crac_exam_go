package main

import (
	"context"
	"crac_exam_go/backend/dao"
	"crac_exam_go/backend/services"
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 初始化数据库
	db, err := dao.GetDB()
	if err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	// 创建数据库表
	if err := dao.InitDatabase(); err != nil {
		log.Fatal("数据库表创建失败:", err)
	}

	// 设置全局数据库实例供 services 使用
	services.SetDB(db)

	// 创建应用实例
	err = wails.Run(&options.App{
		Title:         "业余无线电模拟考试系统",
		Width:         1024,
		Height:        768,
		DisableResize: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		OnStartup: func(ctx context.Context) {
			// 数据库已在 main 中初始化
		},
		OnShutdown: func(ctx context.Context) {
			// 关闭数据库连接
			if err := dao.CloseDB(); err != nil {
				log.Println("关闭数据库连接失败:", err)
			}
		},
		Bind: []interface{}{
			services.NewUserService(services.GetDB()),
			services.NewSettingsService(services.GetDB()),
			services.NewQuestionsBankService(services.GetDB()),
			services.NewExamService(services.GetDB()),
			services.NewPracticeService(services.GetDB()),
			services.NewFavoriteService(services.GetDB()),
			services.NewStatisticsService(services.GetDB()),
			services.NewImportService(services.GetDB()),
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
