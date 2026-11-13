package services

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crac_exam_v3/backend/store"
	"github.com/signintech/gopdf"
)

func TestImportServiceEndToEnd(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)

	// 导入前先制造一些依赖数据，验证导入后会被清理
	oldID := seedQuestion(t, db, "OLD1", 1, 0, 0, 1, "A")
	psvc := NewPracticeService(db)
	psvc.AddFavorite(user.ID, oldID, "A")
	psvc.AddError(user.ID, oldID, "A")
	psvc.SaveProgress(user.ID, "A", 3)

	dir := t.TempDir()
	zipPath := makeTestBankZip(t, dir, 3)
	svc := NewImportService(db)

	result, err := svc.ImportZip(zipPath)
	if err != nil {
		t.Fatalf("ImportZip: %v", err)
	}
	if !result.Success || result.Imported != 3 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.Stats["withImage"] != 3 {
		t.Fatalf("expected 3 images, got %+v", result.Stats)
	}

	count, _ := store.CountQuestions(db)
	if count != 3 {
		t.Fatalf("questions = %d, want 3", count)
	}
	if n, _ := store.CountErrors(db, user.ID); n != 0 {
		t.Fatalf("errors should be cleared by import, got %d", n)
	}
	if n, _ := store.CountFavorites(db, user.ID); n != 0 {
		t.Fatalf("favorites should be cleared by import, got %d", n)
	}

	// 再导入一个更小的题库，验证是替换而不是追加
	zip2 := makeTestBankZip(t, t.TempDir(), 1)
	if _, err := svc.ImportZip(zip2); err != nil {
		t.Fatalf("second import: %v", err)
	}
	count, _ = store.CountQuestions(db)
	if count != 1 {
		t.Fatalf("questions after replace = %d, want 1", count)
	}
}

func TestImportPreservesFavoritesErrorsProgress(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	svc := NewImportService(db)

	first := makeTestBankZip(t, t.TempDir(), 2)
	if _, err := svc.ImportZip(first); err != nil {
		t.Fatalf("first import: %v", err)
	}
	var idJB int64
	if err := db.QueryRow(`SELECT id FROM questions WHERE J = 'JB'`).Scan(&idJB); err != nil {
		t.Fatalf("find JB: %v", err)
	}
	psvc := NewPracticeService(db)
	if _, err := psvc.AddFavorite(user.ID, idJB, "A"); err != nil {
		t.Fatalf("favorite: %v", err)
	}
	if _, err := psvc.AddError(user.ID, idJB, "A"); err != nil {
		t.Fatalf("error: %v", err)
	}
	// 旧题库中 JB 排在第 2 个（索引 1）
	if err := psvc.SaveProgress(user.ID, "A", 1); err != nil {
		t.Fatalf("progress: %v", err)
	}

	second := makeTestBankZip(t, t.TempDir(), 3)
	res, err := svc.ImportZip(second)
	if err != nil {
		t.Fatalf("second import: %v", err)
	}
	if res.PreservedFavorites != 1 || res.PreservedErrors != 1 || res.PreservedProgress != 1 {
		t.Fatalf("preserved counts = %d/%d/%d", res.PreservedFavorites, res.PreservedErrors, res.PreservedProgress)
	}

	var favJ string
	if err := db.QueryRow(`SELECT q.J FROM favorite_questions f JOIN questions q ON q.id = f.question_id
		WHERE f.user_id = ?`, user.ID).Scan(&favJ); err != nil {
		t.Fatalf("favorite not preserved: %v", err)
	}
	if favJ != "JB" {
		t.Fatalf("favorite mapped to %s, want JB", favJ)
	}
	var errJ string
	if err := db.QueryRow(`SELECT q.J FROM error_questions e JOIN questions q ON q.id = e.question_id
		WHERE e.user_id = ?`, user.ID).Scan(&errJ); err != nil {
		t.Fatalf("error not preserved: %v", err)
	}
	if errJ != "JB" {
		t.Fatalf("error mapped to %s, want JB", errJ)
	}
	var idx int
	if err := db.QueryRow(`SELECT current_index FROM practice_progress WHERE user_id = ? AND category = 'A'`,
		user.ID).Scan(&idx); err != nil {
		t.Fatalf("progress not preserved: %v", err)
	}
	if idx != 1 {
		t.Fatalf("progress index = %d, want 1 (JB)", idx)
	}
}

func TestUpdateApplyPreservesAndStoresSource(t *testing.T) {
	db := setupDB(t)
	user := login(t, db)
	svc := NewImportService(db)

	first := makeTestBankZip(t, t.TempDir(), 2)
	if _, err := svc.ImportZip(first); err != nil {
		t.Fatalf("first import: %v", err)
	}
	var idJB int64
	db.QueryRow(`SELECT id FROM questions WHERE J = 'JB'`).Scan(&idJB)
	psvc := NewPracticeService(db)
	psvc.AddFavorite(user.ID, idJB, "A")
	psvc.AddError(user.ID, idJB, "A")
	psvc.SaveProgress(user.ID, "A", 1)

	second := makeTestBankZip(t, t.TempDir(), 3)
	usvc := NewUpdateService(db)
	usvc.pendingPath = second
	usvc.pending = remoteBank{
		URL: "https://www.crac.org.cn/test/bank-v2.zip", Version: "test-v2",
		Size: 1234, ETag: "etag-v2", LastModified: "Fri, 01 Jan 2027 00:00:00 GMT",
		Hash: "hash-v2",
	}
	res, err := usvc.Apply()
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if res.Imported != 3 || res.PreservedFavorites != 1 || res.PreservedErrors != 1 || res.PreservedProgress != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}
	meta, _ := store.GetMetaMap(db, "bank.")
	if meta["source_url"] != "https://www.crac.org.cn/test/bank-v2.zip" {
		t.Fatalf("source URL not stored: %+v", meta)
	}
	if meta["version"] != "test-v2" {
		t.Fatalf("version = %q, want test-v2", meta["version"])
	}
	if meta["etag"] != "etag-v2" {
		t.Fatalf("etag = %q", meta["etag"])
	}
}

func TestCheckOfficialUpdateIfEnabled(t *testing.T) {
	if os.Getenv("CRAC_TEST_UPDATE") == "" {
		t.Skip("set CRAC_TEST_UPDATE=1 to check crac.org.cn")
	}
	zipPath := os.Getenv("CRAC_TEST_ZIP")
	if zipPath == "" {
		t.Skip("set CRAC_TEST_ZIP too")
	}
	db := setupDB(t)
	if _, err := NewImportService(db).ImportZip(zipPath); err != nil {
		t.Fatalf("import: %v", err)
	}
	st, err := NewUpdateService(db).Check()
	if err != nil {
		t.Fatalf("check update: %v", err)
	}
	if st.RemoteURL == "" || !strings.Contains(st.RemoteURL, "crac.org.cn") {
		t.Fatalf("unexpected remote URL: %+v", st)
	}
	if st.Available {
		t.Fatalf("official ZIP should already be up to date: %+v", st)
	}
	if !strings.Contains(st.Message, "最新") {
		t.Fatalf("unexpected message: %s", st.Message)
	}
}

func TestImportServiceFailureKeepsOldBank(t *testing.T) {
	db := setupDB(t)
	login(t, db)
	seedQuestion(t, db, "KEEP1", 1, 0, 0, 1, "A")

	dir := t.TempDir()
	badZip := filepath.Join(dir, "bad.zip")
	makeZipFile(t, badZip, map[string][]byte{"readme.txt": []byte("hello")})

	svc := NewImportService(db)
	if _, err := svc.ImportZip(badZip); err == nil {
		t.Fatalf("expected error for missing PDFs")
	}
	count, _ := store.CountQuestions(db)
	if count != 1 {
		t.Fatalf("old bank should be preserved, got %d questions", count)
	}

	txt := filepath.Join(dir, "notzip.pdf")
	os.WriteFile(txt, []byte("x"), 0o644)
	if _, err := svc.ImportZip(txt); err == nil || !strings.Contains(err.Error(), "ZIP") {
		t.Fatalf("expected ZIP-only error, got %v", err)
	}
}

// ---------- 测试题库 ZIP 生成 ----------

func makeTestBankZip(t *testing.T, dir string, questionCount int) string {
	t.Helper()
	var totalLines, aLines, bLines, cLines, imgLines []string
	for i := 0; i < questionCount; i++ {
		iVal := "I" + string(rune('A'+i))
		jVal := "J" + string(rune('A'+i))
		totalLines = append(totalLines,
			"[J]"+jVal, "[P]P", "[I]"+iVal,
			"[Q]测试题目内容 "+jVal, "[T]A",
			"[A]甲", "[B]乙", "[C]丙", "[D]丁")
		aLines = append(aLines, "[I]"+iVal)
		bLines = append(bLines, "[I]"+iVal)
		cLines = append(cLines, "[I]"+iVal)
		imgLines = append(imgLines, jVal)
	}

	totalPDF := filepath.Join(dir, "total.pdf")
	writePDF(t, totalPDF, totalLines)
	aPDF := filepath.Join(dir, "a.pdf")
	writePDF(t, aPDF, aLines)
	bPDF := filepath.Join(dir, "b.pdf")
	writePDF(t, bPDF, bLines)
	cPDF := filepath.Join(dir, "c.pdf")
	writePDF(t, cPDF, cLines)
	imgPDF := filepath.Join(dir, "img.pdf")
	writeImagePDF(t, imgPDF, imgLines)

	zipPath := filepath.Join(dir, "bank.zip")
	makeZipFile(t, zipPath, map[string][]byte{
		"总题库.pdf":     readBytes(t, totalPDF),
		"总题库附图标记.pdf": readBytes(t, imgPDF),
		"A类题库.pdf":    readBytes(t, aPDF),
		"B类题库.pdf":    readBytes(t, bPDF),
		"C类题库.pdf":    readBytes(t, cPDF),
	})
	return zipPath
}

func writePDF(t *testing.T, path string, lines []string) {
	t.Helper()
	var p gopdf.GoPdf
	p.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	if err := p.AddTTFFont("simhei", `C:\Windows\Fonts\simhei.ttf`); err != nil {
		t.Fatalf("font: %v", err)
	}
	p.AddPage()
	if err := p.SetFont("simhei", "", 11); err != nil {
		t.Fatalf("set font: %v", err)
	}
	y := 30.0
	for _, line := range lines {
		if y > 800 {
			p.AddPage()
			p.SetFont("simhei", "", 11)
			y = 30
		}
		p.SetXY(30, y)
		if err := p.Cell(nil, line); err != nil {
			t.Fatalf("cell: %v", err)
		}
		y += 18
	}
	if err := p.WritePdf(path); err != nil {
		t.Fatalf("write pdf: %v", err)
	}
}

func writeImagePDF(t *testing.T, path string, tokens []string) {
	t.Helper()
	var p gopdf.GoPdf
	p.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	if err := p.AddTTFFont("simhei", `C:\Windows\Fonts\simhei.ttf`); err != nil {
		t.Fatalf("font: %v", err)
	}
	p.AddPage()
	p.SetFont("simhei", "", 12)

	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 12), G: uint8(y * 12), B: 200, A: 255})
		}
	}
	for i, token := range tokens {
		if i > 0 {
			p.AddPage()
			p.SetFont("simhei", "", 12)
		}
		p.SetXY(30, 30)
		if err := p.Cell(nil, token); err != nil {
			t.Fatalf("cell token: %v", err)
		}
		if err := p.ImageFrom(img, 30, 80, &gopdf.Rect{W: 100, H: 100}); err != nil {
			t.Fatalf("image: %v", err)
		}
	}
	if err := p.WritePdf(path); err != nil {
		t.Fatalf("write image pdf: %v", err)
	}
}

func makeZipFile(t *testing.T, path string, files map[string][]byte) {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, data := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create: %v", err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("zip write: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("zip file: %v", err)
	}
}

func readBytes(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return b
}
