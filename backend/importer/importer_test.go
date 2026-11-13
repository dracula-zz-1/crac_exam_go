package importer

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/signintech/gopdf"
)

// testFont 返回一个可用的中文字体路径。
// 英文版 Windows（例如 GitHub Actions 的 windows-latest）没有安装中文字体，
// 这时跳过依赖「生成中文 PDF」的用例，其余逻辑仍由其它用例覆盖。
func testFont(t *testing.T) string {
	t.Helper()
	for _, path := range []string{
		`C:\Windows\Fonts\simhei.ttf`,
		`C:\Windows\Fonts\simsun.ttf`,
		`C:\Windows\Fonts\msyh.ttf`,
		`C:\Windows\Fonts\Deng.ttf`,
		`C:\Windows\Fonts\simkai.ttf`,
	} {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	t.Skip("系统中没有中文字体（simhei/msyh 等），跳过需要生成中文 PDF 的测试")
	return ""
}

func TestParseQuestionsBasic(t *testing.T) {
	text := strings.Join([]string{
		"1",
		"[J]LK0001",
		"[P]1",
		"[I]I001",
		"[Q]我国的无线电管理机构是？",
		"[T]A",
		"[A]国务院",
		"[B]工业和信息化部",
		"[C]广电总局",
		"[D]公安部",
		"[J]LK0002",
		"[P]2",
		"[I]I002",
		"[Q]以下哪些属于正确答案？",
		"[T]CA",
		"[A]甲",
		"[B]乙",
		"[C]丙",
		"[D]丁",
		"[J]LK0003",
		"[P]3",
		"[I]I003",
		"[Q]不完整题目（缺答案）",
		"[A]甲",
		"[B]乙",
		"[C]丙",
		"[D]丁",
	}, "\n")

	questions, stats := ParseQuestions(text)
	if len(questions) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(questions))
	}
	if stats.Skipped != 1 {
		t.Fatalf("expected 1 skipped, got %d", stats.Skipped)
	}
	if questions[0].J != "LK0001" || questions[0].T != "A" || questions[0].Type != 1 {
		t.Fatalf("unexpected first question: %+v", questions[0])
	}
	if questions[1].T != "AC" || questions[1].Type != 2 {
		t.Fatalf("multi answer should be normalized/sorted: %+v", questions[1])
	}
	if questions[0].Q != "我国的无线电管理机构是？" {
		t.Fatalf("unexpected stem: %q", questions[0].Q)
	}
}

func TestParseQuestionsMarkerVariants(t *testing.T) {
	text := "J]LK0100\nP]1\nI]I100\nQ]页面清理测试\nT]B\nA]a\nB]b\nC]c\nD]d\n"
	questions, _ := ParseQuestions(text)
	if len(questions) != 1 {
		t.Fatalf("expected 1 question, got %d", len(questions))
	}
	if questions[0].J != "LK0100" || questions[0].T != "B" {
		t.Fatalf("unexpected question: %+v", questions[0])
	}
	if questions[0].Q != "页面清理测试" {
		t.Fatalf("unexpected stem: %q", questions[0].Q)
	}
}

func TestStripPageNumberOnlyRemovesLeadingPageNumber(t *testing.T) {
	// 该题库页码固定在页首；页尾的纯数字可能是题目内容，不能删除。
	got := stripPageNumber("7\n已知阻抗为\n50\n欧\n7", 7)
	if strings.Contains(got, "\n7\n") || strings.HasPrefix(got, "7") {
		t.Fatalf("leading page number not removed: %q", got)
	}
	if !strings.HasSuffix(got, "7") {
		t.Fatalf("trailing content number should be kept: %q", got)
	}
	if !strings.Contains(got, "50") {
		t.Fatalf("inline number should be kept: %q", got)
	}
}

func TestParseIValues(t *testing.T) {
	values := ParseIValues("[I]I001\n[I] I002 \n[I]I003[I]I004")
	for _, v := range []string{"I001", "I002", "I003", "I004"} {
		if !values[v] {
			t.Fatalf("missing I value %q in %v", v, values)
		}
	}
}

func TestImportZipEndToEnd(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CRAC_EXAM_DATA_DIR", dir)

	questions := []struct {
		j, i, q, ans string
		qtype        int
	}{
		{"LK0001", "I001", "我国的无线电管理机构是？", "A", 1},
		{"LK0002", "I002", "以下哪些属于正确答案？", "AC", 2},
		{"LK0003", "I003", "无线电波的传播方式有哪些？", "B", 1},
	}

	var totalLines []string
	for _, q := range questions {
		totalLines = append(totalLines,
			"[J]"+q.j, "[P]P", "[I]"+q.i, "[Q]"+q.q, "[T]"+q.ans,
			"[A]选项甲", "[B]选项乙", "[C]选项丙", "[D]选项丁")
	}

	totalPDF := filepath.Join(dir, "total.pdf")
	makeTextPDF(t, totalPDF, totalLines)

	// 分类：I001 同时属于 A、B；I002 属于 B；I003 属于 C
	aPDF := filepath.Join(dir, "a.pdf")
	makeTextPDF(t, aPDF, []string{"[I]I001"})
	bPDF := filepath.Join(dir, "b.pdf")
	makeTextPDF(t, bPDF, []string{"[I]I001", "[I]I002"})
	cPDF := filepath.Join(dir, "c.pdf")
	makeTextPDF(t, cPDF, []string{"[I]I003"})

	imagesPDF := filepath.Join(dir, "images.pdf")
	makeImagesPDF(t, imagesPDF, []string{"LK0001", "LK0002", "LK0003"})

	zipPath := filepath.Join(dir, "bank.zip")
	makeZip(t, zipPath, map[string][]byte{
		"题库/总题库.pdf":     readFile(t, totalPDF),
		"题库/A类题库.pdf":    readFile(t, aPDF),
		"题库/B类题库.pdf":    readFile(t, bPDF),
		"题库/C类题库.pdf":    readFile(t, cPDF),
		"题库/总题库附图标记.pdf": readFile(t, imagesPDF),
	})

	parsed, result, err := ImportZip(zipPath, nil)
	if err != nil {
		t.Fatalf("ImportZip failed: %v", err)
	}
	if len(parsed) != 3 {
		t.Fatalf("expected 3 questions, got %d (warnings=%v)", len(parsed), result.Warnings)
	}

	// 验证图片（gopdf 生成的 PDF 中每页一张 PNG）
	withImage := 0
	for _, q := range parsed {
		if strings.HasPrefix(q.F, "data:image/png;base64,") {
			withImage++
		}
	}
	if withImage != 3 {
		t.Fatalf("expected 3 questions with image, got %d; warnings=%v", withImage, result.Warnings)
	}

	byJ := map[string]int{}
	for _, q := range parsed {
		byJ[q.J] = q.LA*100 + q.LB*10 + q.LC
	}
	if byJ["LK0001"] != 110 {
		t.Fatalf("LK0001 category flags wrong: %d", byJ["LK0001"])
	}
	if byJ["LK0002"] != 10 {
		t.Fatalf("LK0002 should be B only: %d", byJ["LK0002"])
	}
	if byJ["LK0003"] != 1 {
		t.Fatalf("LK0003 should be C only: %d", byJ["LK0003"])
	}
	for _, q := range parsed {
		if q.J == "LK0002" && q.Type != 2 {
			t.Fatalf("LK0002 should be multiple choice")
		}
		if q.J == "LK0001" && !strings.Contains(q.Q, "无线电管理机构") {
			t.Fatalf("Chinese text extraction failed, got %q", q.Q)
		}
	}
}

// TestRealBankZipIfProvided 用真实题库做回归验证。
// 设置 CRAC_TEST_ZIP 指向题库 ZIP 后运行；仓库不内置该文件。
func TestRealBankZipIfProvided(t *testing.T) {
	zipPath := os.Getenv("CRAC_TEST_ZIP")
	if zipPath == "" {
		t.Skip("set CRAC_TEST_ZIP to run real bank regression")
	}
	t.Setenv("CRAC_EXAM_DATA_DIR", t.TempDir())

	questions, result, err := ImportZip(zipPath, nil)
	if err != nil {
		t.Fatalf("import real bank: %v", err)
	}
	if len(questions) < 1000 {
		t.Fatalf("too few questions: %d", len(questions))
	}
	if result.Stats["withImage"] < 50 {
		t.Fatalf("expected at least 50 image questions, got %+v", result.Stats)
	}
	if result.DuplicateJPI != 0 {
		t.Fatalf("real bank should have unique J/P/I, got %d duplicates", result.DuplicateJPI)
	}
	if questions[0].J != "LY0001" || !strings.Contains(questions[0].Q, "无线电管理") {
		t.Fatalf("unexpected first question: %+v", questions[0])
	}
	// 行内数字不能被当作页码删掉
	found := false
	for _, q := range questions {
		if strings.Contains(q.Q, "50Ω") || strings.Contains(q.A, "50.0") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("inline numbers seem to be lost during PDF text extraction")
	}
}

func TestDuplicateJPIIsSkipped(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CRAC_EXAM_DATA_DIR", dir)

	lines := []string{
		"[J]DUP", "[P]P1", "[I]I1", "[Q]第一条题目", "[T]A", "[A]a", "[B]b", "[C]c", "[D]d",
		"[J]DUP", "[P]P1", "[I]I1", "[Q]第二条题目", "[T]B", "[A]a", "[B]b", "[C]c", "[D]d",
	}
	totalPDF := filepath.Join(dir, "total.pdf")
	makeTextPDF(t, totalPDF, lines)
	catPDF := filepath.Join(dir, "cat.pdf")
	makeTextPDF(t, catPDF, []string{"[I]I1"})
	imagesPDF := filepath.Join(dir, "images.pdf")
	makeTextPDF(t, imagesPDF, []string{"DUP"})

	zipPath := filepath.Join(dir, "bank.zip")
	makeZip(t, zipPath, map[string][]byte{
		"总题库.pdf":     readFile(t, totalPDF),
		"总题库附图标记.pdf": readFile(t, imagesPDF),
		"A类题库.pdf":    readFile(t, catPDF),
		"B类题库.pdf":    readFile(t, catPDF),
		"C类题库.pdf":    readFile(t, catPDF),
	})

	questions, result, err := ImportZip(zipPath, nil)
	if err != nil {
		t.Fatalf("ImportZip: %v", err)
	}
	if len(questions) != 1 {
		t.Fatalf("duplicate J/P/I should be skipped, got %d questions", len(questions))
	}
	if result.DuplicateJPI != 1 {
		t.Fatalf("DuplicateJPI = %d, want 1", result.DuplicateJPI)
	}
	if result.DuplicateJ != 0 {
		t.Fatalf("after dedupe J should be unique, got DuplicateJ=%d", result.DuplicateJ)
	}
	if questions[0].Q != "第一条题目" {
		t.Fatalf("should keep first occurrence, got %q", questions[0].Q)
	}
}

func TestImportZipRejectsNonZipAndMissingFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CRAC_EXAM_DATA_DIR", dir)

	txt := filepath.Join(dir, "x.pdf")
	makeTextPDF(t, txt, []string{"[J]A"})
	if _, _, err := ImportZip(txt, nil); err == nil || !strings.Contains(err.Error(), "ZIP") {
		t.Fatalf("expected ZIP-only error, got %v", err)
	}

	zipPath := filepath.Join(dir, "empty.zip")
	makeZip(t, zipPath, map[string][]byte{"readme.txt": []byte("hi")})
	_, _, err := ImportZip(zipPath, nil)
	if err == nil || !strings.Contains(err.Error(), "缺少") {
		t.Fatalf("expected missing-file error, got %v", err)
	}
}

func TestDuplicateJWarns(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CRAC_EXAM_DATA_DIR", dir)

	lines := []string{}
	for i := 0; i < 2; i++ {
		// 同 J、不同 P/I：联合键唯一，但 J 重复，应触发图片关联提示
		lines = append(lines,
			"[J]SAME", fmt.Sprintf("[P]P%d", i+1), fmt.Sprintf("[I]I%d", i+1),
			"[Q]题目", "[T]A", "[A]a", "[B]b", "[C]c", "[D]d")
	}
	totalPDF := filepath.Join(dir, "total.pdf")
	makeTextPDF(t, totalPDF, lines)
	aPDF := filepath.Join(dir, "a.pdf")
	makeTextPDF(t, aPDF, []string{"[I]I1"})
	bPDF := filepath.Join(dir, "b.pdf")
	makeTextPDF(t, bPDF, []string{"[I]I1"})
	cPDF := filepath.Join(dir, "c.pdf")
	makeTextPDF(t, cPDF, []string{"[I]I1"})
	imagesPDF := filepath.Join(dir, "images.pdf")
	makeTextPDF(t, imagesPDF, []string{"SAME", "SAME"})

	zipPath := filepath.Join(dir, "bank.zip")
	makeZip(t, zipPath, map[string][]byte{
		"总题库.pdf":     readFile(t, totalPDF),
		"A类题库.pdf":    readFile(t, aPDF),
		"B类题库.pdf":    readFile(t, bPDF),
		"C类题库.pdf":    readFile(t, cPDF),
		"总题库附图标记.pdf": readFile(t, imagesPDF),
	})

	_, result, err := ImportZip(zipPath, nil)
	if err != nil {
		t.Fatalf("ImportZip failed: %v", err)
	}
	if result.DuplicateJ != 1 {
		t.Fatalf("expected 1 duplicate J, got %d", result.DuplicateJ)
	}
}

// ---------- 测试辅助 ----------

func makeTextPDF(t *testing.T, path string, lines []string) {
	t.Helper()
	var pdf gopdf.GoPdf
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	if err := pdf.AddTTFFont("simhei", testFont(t)); err != nil {
		t.Fatalf("add font: %v", err)
	}
	if err := pdf.SetFont("simhei", "", 11); err != nil {
		t.Fatalf("set font: %v", err)
	}
	pdf.AddPage()
	if err := pdf.SetFont("simhei", "", 11); err != nil {
		t.Fatalf("set font: %v", err)
	}
	y := 30.0
	for _, line := range lines {
		if y > 800 {
			pdf.AddPage()
			if err := pdf.SetFont("simhei", "", 11); err != nil {
				t.Fatalf("set font: %v", err)
			}
			y = 30
		}
		pdf.SetXY(30, y)
		if err := pdf.Cell(nil, line); err != nil {
			t.Fatalf("write cell: %v", err)
		}
		y += 18
	}
	if err := pdf.WritePdf(path); err != nil {
		t.Fatalf("write pdf: %v", err)
	}
}

func makeImagesPDF(t *testing.T, path string, tokens []string) {
	t.Helper()
	var pdf gopdf.GoPdf
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	if err := pdf.AddTTFFont("simhei", testFont(t)); err != nil {
		t.Fatalf("add font: %v", err)
	}
	if err := pdf.SetFont("simhei", "", 12); err != nil {
		t.Fatalf("set font: %v", err)
	}

	pdf.AddPage()
	if err := pdf.SetFont("simhei", "", 12); err != nil {
		t.Fatalf("set font: %v", err)
	}

	img := testPNG(t)
	for i, token := range tokens {
		if i > 0 {
			pdf.AddPage()
			if err := pdf.SetFont("simhei", "", 12); err != nil {
				t.Fatalf("set font: %v", err)
			}
		}
		pdf.SetXY(30, 30)
		if err := pdf.Cell(nil, token); err != nil {
			t.Fatalf("write token: %v", err)
		}
		if err := pdf.ImageFrom(img, 30, 80, &gopdf.Rect{W: 120, H: 120}); err != nil {
			t.Fatalf("place image: %v", err)
		}
	}
	if err := pdf.WritePdf(path); err != nil {
		t.Fatalf("write images pdf: %v", err)
	}
}

func testPNG(t *testing.T) image.Image {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 24, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 10), G: uint8(y * 10), B: 180, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return img
}

func makeZip(t *testing.T, path string, files map[string][]byte) {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, data := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write zip file: %v", err)
	}
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}
