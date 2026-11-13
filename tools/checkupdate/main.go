// 命令行检查/应用 CRAC 官方题库更新（使用应用默认数据目录）：
//
//	go run ./tools/checkupdate
//	go run ./tools/checkupdate -apply
//
// 可用 CRAC_EXAM_DATA_DIR 指向其他数据目录。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"crac_exam_v3/backend/config"
	"crac_exam_v3/backend/services"
	"crac_exam_v3/backend/store"
	"crac_exam_v3/backend/utils"
)

func main() {
	apply := flag.Bool("apply", false, "如果发现更新，立即下载并导入")
	flag.Parse()

	utils.Init(config.LogDir())
	db, err := store.Open(config.DBPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "打开数据库失败:", err)
		os.Exit(1)
	}
	defer db.Close()

	svc := services.NewUpdateService(db)
	st, err := svc.Check()
	if err != nil {
		fmt.Fprintln(os.Stderr, "检查更新失败:", err)
		os.Exit(1)
	}
	out, _ := json.MarshalIndent(st, "", "  ")
	fmt.Println(string(out))
	fmt.Println("数据库:", config.DBPath())

	if *apply && st.Available {
		res, err := svc.Apply()
		if err != nil {
			fmt.Fprintln(os.Stderr, "更新失败:", err)
			os.Exit(1)
		}
		fmt.Printf("更新完成: %s（%d 题，保留收藏 %d / 错题 %d / 进度 %d）\n",
			res.Message, res.Imported, res.PreservedFavorites, res.PreservedErrors, res.PreservedProgress)
	}
}
