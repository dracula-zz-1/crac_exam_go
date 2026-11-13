// 命令行题库导入校验工具：
//
//	go run ./tools/verifyimport <题库.zip>
//
// 解析 ZIP、写入临时 SQLite，并输出统计与样例题，便于在不启动 GUI 的情况下验证真实题库。
package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"crac_exam_v3/backend/services"
	"crac_exam_v3/backend/store"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "用法: go run ./tools/verifyimport <题库.zip>")
		os.Exit(2)
	}
	zipPath := os.Args[1]

	dir := os.Getenv("CRAC_VERIFY_DIR")
	keep := dir != ""
	if !keep {
		var err error
		dir, err = os.MkdirTemp("", "crac-verify-*")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(dir)
	} else {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			panic(err)
		}
	}
	os.Setenv("CRAC_EXAM_DATA_DIR", dir)

	db, err := store.Open(filepath.Join(dir, "verify.db"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	res, err := services.NewImportService(db).ImportZip(zipPath)
	if err != nil {
		fmt.Printf("IMPORT FAILED: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("success=%v imported=%d total=%d elapsed=%dms duplicateJ=%d duplicateJPI=%d\n",
		res.Success, res.Imported, res.Total, res.ElapsedMS, res.DuplicateJ, res.DuplicateJPI)
	fmt.Println("stats:", res.Stats)
	for _, w := range res.Warnings {
		fmt.Println("warning:", w)
	}

	var dupGroups, idxCount int
	db.QueryRow(`SELECT COUNT(*) FROM (SELECT J,P,I FROM questions GROUP BY J,P,I HAVING COUNT(*) > 1)`).Scan(&dupGroups)
	db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_questions_jpi'`).Scan(&idxCount)
	fmt.Printf("unique index idx_questions_jpi=%d, duplicate J/P/I groups=%d\n", idxCount, dupGroups)

	counts, err := store.CategoryCounts(db)
	if err != nil {
		panic(err)
	}
	for _, c := range counts {
		fmt.Printf("category %s: total=%d single=%d multiple=%d withImage=%d\n",
			c.Category, c.Total, c.Single, c.Multiple, c.WithImg)
	}

	rows, err := db.Query(`SELECT id, J, P, I, substr(Q,1,60), T, type,
		length(F), LA, LB, LC FROM questions ORDER BY id LIMIT 5`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	fmt.Println("samples:")
	for rows.Next() {
		var id int64
		var j, p, i, q, t string
		var typ, flen, la, lb, lc int
		if err := rows.Scan(&id, &j, &p, &i, &q, &t, &typ, &flen, &la, &lb, &lc); err != nil {
			panic(err)
		}
		fmt.Printf("  id=%d J=%s P=%s I=%s type=%d T=%s imgBytes=%d flags=%d/%d/%d Q=%q\n",
			id, j, p, i, typ, t, flen, la, lb, lc, strings.TrimSpace(q))
	}

	// 题目答案分布
	rows2, err := db.Query(`SELECT T, COUNT(*) FROM questions GROUP BY T ORDER BY COUNT(*) DESC LIMIT 12`)
	if err != nil {
		panic(err)
	}
	defer rows2.Close()
	fmt.Println("answer distribution:")
	for rows2.Next() {
		var t string
		var n int
		if err := rows2.Scan(&t, &n); err != nil {
			panic(err)
		}
		fmt.Printf("  T=%-4s %d\n", t, n)
	}

	fmt.Println("duplicate J:")
	rows3, err := db.Query(`SELECT J, COUNT(*) FROM questions GROUP BY J HAVING COUNT(*) > 1 ORDER BY COUNT(*) DESC`)
	if err != nil {
		panic(err)
	}
	defer rows3.Close()
	for rows3.Next() {
		var j string
		var n int
		if err := rows3.Scan(&j, &n); err != nil {
			panic(err)
		}
		fmt.Printf("  J=%-10s %d\n", j, n)
	}

	fmt.Println("image J:")
	rows4, err := db.Query(`SELECT J FROM questions WHERE F <> '' ORDER BY J`)
	if err != nil {
		panic(err)
	}
	defer rows4.Close()
	var js []string
	for rows4.Next() {
		var j string
		if err := rows4.Scan(&j); err != nil {
			panic(err)
		}
		js = append(js, j)
	}
	fmt.Printf("  count=%d %v\n", len(js), js)

	verifyRealFlows(db)
}

// verifyRealFlows 用真实题库跑一遍登录、首页统计、三类考试与练习取数。
func verifyRealFlows(db *sql.DB) {
	fmt.Println("=== real data flows ===")
	user, err := services.NewUserService(db).Login("验证用户", "VERIFY-001")
	if err != nil {
		panic(err)
	}
	dash, err := services.NewStatisticsService(db).GetDashboard(user.ID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("dashboard: total=%d images=%d errors=%d favorites=%d\n",
		dash.TotalQuestions, dash.WithImages, dash.ErrorCount, dash.FavoriteCount)

	practice, err := services.NewPracticeService(db).GetPracticeData(user.ID, "A", "practice")
	if err != nil {
		panic(err)
	}
	withImg := 0
	for _, q := range practice.Questions {
		if q.F != "" {
			withImg++
		}
	}
	fmt.Printf("practice A: questions=%d withImage=%d first=%s T=%s\n",
		len(practice.Questions), withImg, practice.Questions[0].J, practice.Questions[0].T)

	for _, cat := range []string{"A", "B", "C"} {
		start, err := services.NewExamService(db).CreateExam(user.ID, cat)
		if err != nil {
			panic(fmt.Sprintf("create %s exam: %v", cat, err))
		}
		details, err := store.GetExamDetails(db, start.ExamID)
		if err != nil {
			panic(err)
		}
		answers := map[string]string{}
		for _, d := range details {
			answers[fmt.Sprintf("%d", d.QuestionID)] = d.CorrectAnswer
		}
		result, err := services.NewExamService(db).SubmitExam(start.ExamID, answers)
		if err != nil {
			panic(fmt.Sprintf("submit %s exam: %v", cat, err))
		}
		fmt.Printf("exam %s: questions=%d score=%d/%d passed=%v duration=%ds\n",
			cat, start.Total, result.Score, result.TotalCount, result.Passed, result.DurationSeconds)
	}
}
