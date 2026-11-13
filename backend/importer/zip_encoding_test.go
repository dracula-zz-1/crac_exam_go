package importer

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// 构造一个最小合法题库 ZIP，允许自定义条目名与额外条目。
func makeNamedBankZip(t *testing.T, dir string, nameFn func(string) string, extra map[string][]byte) string {
	t.Helper()
	totalPDF := filepath.Join(dir, "total.pdf")
	makeTextPDF(t, totalPDF, []string{
		"[J]J1", "[P]P1", "[I]I1", "[Q]测试题目", "[T]A", "[A]a", "[B]b", "[C]c", "[D]d",
	})
	catPDF := filepath.Join(dir, "cat.pdf")
	makeTextPDF(t, catPDF, []string{"[I]I1"})
	imagesPDF := filepath.Join(dir, "images.pdf")
	makeTextPDF(t, imagesPDF, []string{"J1"})

	files := map[string][]byte{
		"总题库.pdf":     readFile(t, totalPDF),
		"总题库附图标记.pdf": readFile(t, imagesPDF),
		"A类题库.pdf":    readFile(t, catPDF),
		"B类题库.pdf":    readFile(t, catPDF),
		"C类题库.pdf":    readFile(t, catPDF),
	}
	for k, v := range extra {
		files[k] = v
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, data := range files {
		hdr := &zip.FileHeader{Name: nameFn(name), Method: zip.Deflate}
		hdr.NonUTF8 = true
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			t.Fatalf("create header %q: %v", name, err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("write %q: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	zipPath := filepath.Join(dir, "bank.zip")
	if err := os.WriteFile(zipPath, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write zip: %v", err)
	}
	return zipPath
}

func TestImportZipGBKFilenames(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CRAC_EXAM_DATA_DIR", dir)

	zipPath := makeNamedBankZip(t, dir, func(name string) string {
		encoded, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(name))
		if err != nil {
			t.Fatalf("gbk encode: %v", err)
		}
		return string(encoded)
	}, map[string][]byte{
		"__MACOSX/._总题库.pdf": []byte("metadata"),
		"._总题库.pdf":          []byte("metadata"),
	})

	questions, result, err := ImportZip(zipPath, nil)
	if err != nil {
		t.Fatalf("GBK ZIP import: %v", err)
	}
	if len(questions) != 1 || result.Imported != 1 {
		t.Fatalf("unexpected result: %d questions, %+v", len(questions), result)
	}
}

func TestImportZipUppercaseExtension(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CRAC_EXAM_DATA_DIR", dir)

	zipPath := makeNamedBankZip(t, dir, func(name string) string { return name }, nil)
	upper := filepath.Join(dir, "BANK.ZIP")
	data, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("read zip: %v", err)
	}
	if err := os.WriteFile(upper, data, 0o644); err != nil {
		t.Fatalf("write upper: %v", err)
	}
	if _, _, err := ImportZip(upper, nil); err != nil {
		t.Fatalf("uppercase .ZIP should be accepted: %v", err)
	}
	if _, _, err := ImportZip(strings.Replace(upper, ".ZIP", ".txt", 1), nil); err == nil {
		t.Fatalf("non-zip extension should be rejected")
	}
}
