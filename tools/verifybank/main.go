// 题库完整性校验工具（不启动 GUI）：
//
//	go run ./tools/verifybank <题库.zip>
//
// 校验项：
//   - 题目数量、J/P/I 联合唯一、I 唯一、字段完整性
//   - 题型与答案一致性、答案字符合法性、选项完整性
//   - A/B/C 分类与分类 PDF 的 [I] 集合一致
//   - 图片数量、格式、尺寸、图片题号与附图标记一致
//   - 题干/选项/答案能否在其原始 [I] 题目块中找到（防止解析截断/错位）
//   - 标记数量与解析题目数量一致
//
// 任一项失败会以非零状态退出，便于 CI/回归使用。
package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"crac_exam_v3/backend/importer"
	"crac_exam_v3/backend/services"
	"crac_exam_v3/backend/store"
)

type checker struct {
	failures []string
	warnings []string
}

func (c *checker) fail(format string, args ...any) {
	c.failures = append(c.failures, fmt.Sprintf(format, args...))
}

func (c *checker) warn(format string, args ...any) {
	c.warnings = append(c.warnings, fmt.Sprintf(format, args...))
}

func norm(s string) string {
	return strings.Join(strings.Fields(s), "")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "用法: go run ./tools/verifybank <题库.zip>")
		os.Exit(2)
	}
	zipPath := os.Args[1]
	dir := os.Getenv("CRAC_VERIFY_DIR")
	keep := dir != ""
	if !keep {
		var err error
		dir, err = os.MkdirTemp("", "crac-bank-*")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(dir)
	}
	os.Setenv("CRAC_EXAM_DATA_DIR", dir)

	c := &checker{}

	// 解析 ZIP（独立于写库，保留文本用于来源校验）
	bundleDir, err := os.MkdirTemp(dir, "bundle-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(bundleDir)
	bundle, _, err := importer.ExtractBundle(zipPath, bundleDir)
	if err != nil {
		fmt.Println("FAIL: 解压 ZIP:", err)
		os.Exit(1)
	}
	totalText, err := importer.ExtractText(bundle.Total)
	if err != nil {
		fmt.Println("FAIL: 读取总题库文本:", err)
		os.Exit(1)
	}

	// 导入临时数据库
	dbPath := filepath.Join(dir, "verify.db")
	_ = os.Remove(dbPath)
	db, err := store.Open(dbPath)
	if err != nil {
		fmt.Println("FAIL: 打开数据库:", err)
		os.Exit(1)
	}
	defer db.Close()
	result, err := services.NewImportService(db).ImportZip(zipPath)
	if err != nil {
		fmt.Println("FAIL: 导入:", err)
		os.Exit(1)
	}
	fmt.Printf("导入: success=%v imported=%d duplicateJ=%d duplicateJPI=%d elapsed=%dms\n",
		result.Success, result.Imported, result.DuplicateJ, result.DuplicateJPI, result.ElapsedMS)
	if result.DuplicateJPI != 0 {
		c.fail("联合唯一键重复未被去重：duplicateJPI=%d", result.DuplicateJPI)
	}

	type question struct {
		id                int64
		j, p, i, q, t     string
		a, b, cc, d, f    string
		la, lb, lc, qtype int
	}
	rows, err := db.Query(`SELECT id,J,P,I,Q,T,A,B,C,D,F,LA,LB,LC,type FROM questions ORDER BY id`)
	if err != nil {
		panic(err)
	}
	var qs []question
	for rows.Next() {
		var q question
		if err := rows.Scan(&q.id, &q.j, &q.p, &q.i, &q.q, &q.t, &q.a, &q.b, &q.cc, &q.d, &q.f,
			&q.la, &q.lb, &q.lc, &q.qtype); err != nil {
			panic(err)
		}
		qs = append(qs, q)
	}
	rows.Close()

	fmt.Printf("题目总数: %d\n", len(qs))
	if len(qs) != result.Imported {
		c.fail("数据库题目数 %d != 导入返回 %d", len(qs), result.Imported)
	}

	// 唯一约束
	var distinctJPI, distinctI int
	db.QueryRow(`SELECT COUNT(*) FROM (SELECT J,P,I FROM questions GROUP BY J,P,I)`).Scan(&distinctJPI)
	db.QueryRow(`SELECT COUNT(*) FROM (SELECT I FROM questions GROUP BY I)`).Scan(&distinctI)
	fmt.Printf("唯一 J/P/I: %d, 唯一 I: %d\n", distinctJPI, distinctI)
	if distinctJPI != len(qs) {
		c.fail("J/P/I 存在重复：%d 个唯一组合 / %d 道题", distinctJPI, len(qs))
	}
	if distinctI != len(qs) {
		c.fail("I 存在重复：%d 个唯一 I / %d 道题", distinctI, len(qs))
	}

	var idx int
	db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_questions_jpi'`).Scan(&idx)
	if idx != 1 {
		c.fail("缺少 idx_questions_jpi 唯一索引")
	}

	// 字段完整性 / 题型答案
	answerRe := regexp.MustCompile(`^[A-D]{1,4}$`)
	dupOptionQuestions := 0
	for _, q := range qs {
		if strings.TrimSpace(q.j) == "" || strings.TrimSpace(q.p) == "" || strings.TrimSpace(q.i) == "" ||
			strings.TrimSpace(q.q) == "" || strings.TrimSpace(q.t) == "" {
			c.fail("题目 id=%d 存在空 J/P/I/Q/T", q.id)
		}
		for name, v := range map[string]string{"A": q.a, "B": q.b, "C": q.cc, "D": q.d} {
			if strings.TrimSpace(v) == "" {
				c.fail("题目 %s (I=%s) 选项 %s 为空", q.j, q.i, name)
			}
		}
		if !answerRe.MatchString(q.t) {
			c.fail("题目 %s (I=%s) 答案格式无效: %q", q.j, q.i, q.t)
		}
		if len([]rune(q.t)) > 1 {
			seen := map[rune]bool{}
			for _, r := range q.t {
				if seen[r] {
					c.fail("题目 %s (I=%s) 答案重复字符: %q", q.j, q.i, q.t)
				}
				seen[r] = true
			}
		}
		wantType := 1
		if len([]rune(q.t)) > 1 {
			wantType = 2
		}
		if q.qtype != wantType {
			c.fail("题目 %s (I=%s) 题型=%d 与答案 %q 不一致", q.j, q.i, q.qtype, q.t)
		}
		opts := map[string]bool{}
		for _, v := range []string{norm(q.a), norm(q.b), norm(q.cc), norm(q.d)} {
			opts[v] = true
		}
		if len(opts) < 4 {
			dupOptionQuestions++
		}
	}
	if dupOptionQuestions > 0 {
		c.warn("%d 道题存在完全相同的选项文本，请人工抽查", dupOptionQuestions)
	}

	// 分类：与分类 PDF 的 [I] 集合逐一比对
	catExpected := map[string]map[string]bool{}
	for _, item := range []struct{ key, path string }{
		{"A", bundle.A}, {"B", bundle.B}, {"C", bundle.C},
	} {
		text, err := importer.ExtractText(item.path)
		if err != nil {
			c.fail("读取 %s 类 PDF 失败: %v", item.key, err)
			continue
		}
		catExpected[item.key] = importer.ParseIValues(text)
	}
	catCounts := map[string]int{"A": 0, "B": 0, "C": 0}
	catMismatch := 0
	for _, q := range qs {
		iVal := norm(q.i)
		got := map[string]bool{"A": q.la == 1, "B": q.lb == 1, "C": q.lc == 1}
		for _, cat := range []string{"A", "B", "C"} {
			want := catExpected[cat][iVal]
			if got[cat] != want {
				catMismatch++
				if catMismatch <= 10 {
					c.fail("分类不一致 I=%s J=%s %s: 题目=%v PDF=%v", q.i, q.j, cat, got[cat], want)
				}
			}
			if got[cat] {
				catCounts[cat]++
			}
		}
		if !got["A"] && !got["B"] && !got["C"] {
			c.fail("题目 I=%s J=%s 不属于任何类别", q.i, q.j)
		}
	}
	fmt.Printf("分类: A=%d B=%d C=%d (PDF 唯一 I: A=%d B=%d C=%d), 成员不一致=%d\n",
		catCounts["A"], catCounts["B"], catCounts["C"],
		len(catExpected["A"]), len(catExpected["B"]), len(catExpected["C"]), catMismatch)
	if catMismatch > 0 {
		c.fail("分类成员不一致共 %d 项", catMismatch)
	}

	// 图片：解码、尺寸、题号集合
	imgTokens := imageTokens(bundle.Images)
	imgs, err := importer.ExtractImages(bundle.Images)
	if err != nil {
		c.warn("图片提取失败: %v", err)
	}
	imgCount := 0
	for _, page := range imgs {
		imgCount += len(page)
	}
	imgQuestions := 0
	for _, q := range qs {
		if q.f == "" {
			continue
		}
		imgQuestions++
		if !strings.HasPrefix(q.f, "data:image/") {
			c.fail("题目 I=%s 图片不是 data URL", q.i)
			continue
		}
		raw := q.f[strings.Index(q.f, ",")+1:]
		data, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			c.fail("题目 I=%s 图片 base64 解码失败: %v", q.i, err)
			continue
		}
		cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			c.fail("题目 I=%s 图片无法解码: %v", q.i, err)
			continue
		}
		if cfg.Width <= 0 || cfg.Height <= 0 {
			c.fail("题目 I=%s 图片尺寸无效: %s", q.i, format)
		}
	}
	fmt.Printf("图片: 题目带图=%d, PDF 提取=%d, 附图标记=%d\n", imgQuestions, imgCount, len(imgTokens))
	if imgCount != len(imgTokens) {
		c.warn("附图标记数量 %d != 图片数量 %d（导入器会给出警告）", len(imgTokens), imgCount)
	}

	// 该题库存在“一题多个 J 码”的写法（如 LY0014、LY0017），
	// 解析时会按同一题拼接，因此 J 标记数可能略多于题目数。
	multiJExtra := 0
	for _, q := range qs {
		if n := strings.Count(q.j, "、"); n > 0 {
			multiJExtra += n
		}
	}
	fmt.Printf("多 J 码题目额外标记: %d\n", multiJExtra)

	// 每个题目字段都必须能在其 [I] 块中找到（按 I 定位，避免重复 J 干扰），
	// 并进一步按标记位置重新抽取 Q/T/A-D，逐字段核对位置与内容。
	blockByI, blocksWithoutI := buildBlocksByI(totalText)
	missing := 0
	fieldMismatch := 0
	for _, q := range qs {
		block, ok := blockByI[norm(q.i)]
		if !ok {
			missing++
			if missing <= 10 {
				c.fail("在原始文本中找不到 I=%s 的题目块", q.i)
			}
			continue
		}
		fields := map[string]string{
			"Q": q.q, "T": q.t, "A": q.a, "B": q.b, "C": q.cc, "D": q.d,
		}
		for name, v := range fields {
			if norm(v) == "" {
				continue
			}
			if !strings.Contains(block, norm(v)) {
				missing++
				if missing <= 20 {
					c.fail("题目 I=%s 字段 %s 不在原始题目块中: %q", q.i, name, truncate(v, 80))
				}
			}
		}
		extracted := extractBlockFields(block)
		for _, name := range []string{"Q", "T", "A", "B", "C", "D"} {
			if got, want := extracted[name], norm(fields[name]); got != want {
				fieldMismatch++
				if fieldMismatch <= 20 {
					c.fail("题目 I=%s 字段 %s 位置/内容不一致：文本=%q 解析=%q",
						q.i, name, truncate(got, 80), truncate(want, 80))
				}
			}
		}
	}
	fmt.Printf("原始文本块: 有 I 的块=%d, 无 I 的块=%d, 字段缺失=%d, 字段位置不一致=%d\n",
		len(blockByI), blocksWithoutI, missing, fieldMismatch)
	if blocksWithoutI != multiJExtra {
		c.fail("无 I 的 J 块=%d，与多 J 码额外标记=%d 不一致", blocksWithoutI, multiJExtra)
	}

	// 标记数量
	markerCounts := countMarkers(totalText)
	fmt.Printf("标记数量: J=%d P=%d I=%d Q=%d T=%d A=%d B=%d C=%d D=%d\n",
		markerCounts["J"], markerCounts["P"], markerCounts["I"], markerCounts["Q"],
		markerCounts["T"], markerCounts["A"], markerCounts["B"], markerCounts["C"], markerCounts["D"])
	for _, m := range []string{"J", "P", "I", "Q", "T", "A", "B", "C", "D"} {
		want := len(qs)
		if m == "J" {
			want += multiJExtra
		}
		if markerCounts[m] != want {
			c.fail("标记 %s 数量=%d，期望 %d", m, markerCounts[m], want)
		}
	}

	for _, w := range result.Warnings {
		c.warn("导入警告: %s", w)
	}

	fmt.Println()
	if len(c.warnings) > 0 {
		fmt.Println("WARN:")
		for _, w := range c.warnings {
			fmt.Println("  -", w)
		}
	}
	if len(c.failures) > 0 {
		fmt.Println("FAIL:")
		for _, f := range c.failures {
			fmt.Println("  -", f)
		}
		os.Exit(1)
	}
	fmt.Println("RESULT: PASS")
}

func buildBlocksByI(text string) (map[string]string, int) {
	re := regexp.MustCompile(`\[J\]|J\]`)
	locs := re.FindAllStringIndex(text, -1)
	out := map[string]string{}
	withoutI := 0
	for i, loc := range locs {
		start := loc[0]
		end := len(text)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		block := text[start:end]
		values := importer.ParseIValues(block)
		if len(values) == 0 {
			withoutI++
			continue
		}
		var keys []string
		for k := range values {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out[keys[0]] = norm(block)
	}
	return out, withoutI
}

// extractBlockFields 在一个题目块内按标记位置重新抽取字段，用于独立核对解析结果。
func extractBlockFields(block string) map[string]string {
	re := regexp.MustCompile(`\[[JPIQTFABCD]\]|[JPIQTFABCD]\]`)
	locs := re.FindAllStringIndex(block, -1)
	out := map[string]string{}
	for i, loc := range locs {
		marker := block[loc[0]:loc[1]]
		field := ""
		if marker[0] == '[' {
			field = string(marker[1])
		} else {
			field = string(marker[0])
		}
		end := len(block)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		switch field {
		case "Q", "T", "A", "B", "C", "D":
			out[field] = norm(block[loc[1]:end])
		}
	}
	return out
}

func imageTokens(path string) []string {
	pages, err := importer.ExtractPageTexts(path)
	if err != nil {
		return nil
	}
	var out []string
	for _, p := range pages {
		out = append(out, importer.ParseImageTokens(p)...)
	}
	return out
}

func countMarkers(text string) map[string]int {
	out := map[string]int{}
	for _, m := range []string{"J", "P", "I", "Q", "T", "A", "B", "C", "D"} {
		re := regexp.MustCompile(`\[` + m + `\]|` + m + `\]`)
		out[m] = len(re.FindAllString(text, -1))
	}
	return out
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
