package main

import (
	"context"
	"crac_exam_go/backend/services"
	"embed"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 创建应用实例
	err := wails.Run(&options.App{
		Title:         "业余无线电模拟考试系统",
		Width:         1024,
		Height:        768,
		DisableResize: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		OnStartup: func(ctx context.Context) {
			// 初始化数据库
			services.InitDB()
		},
		Bind: []interface{}{
			services.NewUserService(services.GetDB()),
			services.NewSettingsService(services.GetDB()),
			services.NewQuestionsBankService(services.GetDB()),
			services.NewExamService(services.GetDB()),
			services.NewPracticeService(services.GetDB()),
			services.NewFavoriteService(services.GetDB()),
			services.NewStatisticsService(services.GetDB()),
			func() *services.ImportService {
				// 获取可执行文件所在目录
				execPath, _ := os.Executable()
				execDir := filepath.Dir(execPath)

				// Python 脚本目录：优先查找应用目录下的 python_scripts，其次查找开发目录
				scriptDir := filepath.Join(execDir, "python_scripts")
				if _, err := os.Stat(scriptDir); os.IsNotExist(err) {
					// 开发模式：使用绝对路径
					scriptDir = "D:\\crac_new\\crac_exam_go\\python_scripts"
				}

				// Python 解释器路径：优先查找应用目录下的 python.exe，其次使用系统 PATH 中的 python
				pythonPath := filepath.Join(execDir, "python.exe")
				if _, err := os.Stat(pythonPath); os.IsNotExist(err) {
					pythonPath = "python"
				}

				return services.NewImportService(services.GetDB(), pythonPath, scriptDir)
			}(),
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
