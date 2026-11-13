// 命令行导入题库到应用默认数据目录（无需启动 GUI）：
//
//	go run ./tools/importbank <题库.zip>
//
// 数据目录与桌面应用一致，可用 CRAC_EXAM_DATA_DIR 覆盖。
package main

import (
	"fmt"
	"os"

	"crac_exam_v3/backend/config"
	"crac_exam_v3/backend/services"
	"crac_exam_v3/backend/store"
	"crac_exam_v3/backend/utils"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "用法: go run ./tools/importbank <题库.zip>")
		os.Exit(2)
	}
	utils.Init(config.LogDir())
	db, err := store.Open(config.DBPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "打开数据库失败:", err)
		os.Exit(1)
	}
	defer db.Close()

	res, err := services.NewImportService(db).ImportZip(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "导入失败:", err)
		os.Exit(1)
	}
	fmt.Printf("数据库: %s\n", config.DBPath())
	fmt.Printf("导入完成: %s（%d 题，耗时 %dms）\n", res.Message, res.Imported, res.ElapsedMS)
	for _, w := range res.Warnings {
		fmt.Println("警告:", w)
	}
}
