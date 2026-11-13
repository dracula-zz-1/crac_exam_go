package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"crac_exam_v3/backend/config"
	"crac_exam_v3/backend/models"
	"crac_exam_v3/backend/services"
	"crac_exam_v3/backend/store"
	"crac_exam_v3/backend/utils"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	utils.Init(config.LogDir())
	utils.Info("main", "应用启动", "version", config.AppVersion, "data_dir", config.DataDir())

	db, err := store.Open(config.DBPath())
	if err != nil {
		utils.Error("main", "数据库初始化失败", err)
		os.Exit(1)
	}
	defer db.Close()

	// 上次运行遗留的“进行中”考试不计入成绩
	store.CleanupAbandonedExams(db)

	application.RegisterEvent[models.ImportProgress]("import:progress")
	application.RegisterEvent[models.BankUpdateStatus]("bank:update-status")
	application.RegisterEvent[models.AppUpdateStatus]("app:update-status")
	application.RegisterEvent[models.AppUpdateProgress]("app:update-progress")
	application.RegisterEvent[[]string]("files-dropped")

	app := application.New(application.Options{
		Name:        config.AppName,
		Description: config.AppDesc,
		Services: []application.Service{
			application.NewService(services.NewUserService(db)),
			application.NewService(services.NewQuestionService(db)),
			application.NewService(services.NewExamService(db)),
			application.NewService(services.NewPracticeService(db)),
			application.NewService(services.NewStatisticsService(db)),
			application.NewService(services.NewSettingsService(db)),
			application.NewService(services.NewImportService(db)),
			application.NewService(services.NewUpdateService(db)),
			application.NewService(services.NewAppUpdateService(db)),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		ErrorHandler: func(err error) {
			utils.Error("application", "未处理错误", err)
		},
		PanicHandler: func(details *application.PanicDetails) {
			utils.Error("application", "捕获到未处理的 panic", details.Error, "stack", details.FullStackTrace)
		},
		OnShutdown: func() {
			_ = db.Close()
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	win := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            config.AppName,
		Width:            1280,
		Height:           840,
		MinWidth:         1100,
		MinHeight:        720,
		BackgroundColour: application.NewRGB(247, 245, 240),
		URL:              "/",
		EnableFileDrop:   true,
	})

	// 原生拖拽文件（只处理路径，真正导入由前端确认）
	win.OnWindowEvent(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {
		files := event.Context().DroppedFiles()
		if len(files) > 0 {
			app.Event.Emit("files-dropped", files)
		}
	})

	if err := app.Run(); err != nil {
		utils.Error("main", "应用退出", err)
		os.Exit(1)
	}
}
