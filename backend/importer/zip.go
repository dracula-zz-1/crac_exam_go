package importer

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
)

const (
	maxZipEntries      = 10000
	maxSingleFileBytes = 512 << 20  // 512MB
	maxTotalBytes      = 1536 << 20 // 1.5GB
)

// Bundle 解压后的五份必需 PDF 路径。
type Bundle struct {
	Total  string
	Images string
	A      string
	B      string
	C      string
}

type requiredPDF struct {
	key      string
	label    string
	nameNorm string
	dest     *string
}

// ExtractBundle 校验并解压 ZIP 中的五份题库 PDF。
func ExtractBundle(zipPath, destDir string) (*Bundle, []string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, nil, fmt.Errorf("无法打开 ZIP：%w", err)
	}
	defer r.Close()

	if len(r.File) == 0 {
		return nil, nil, fmt.Errorf("ZIP 文件为空")
	}
	if len(r.File) > maxZipEntries {
		return nil, nil, fmt.Errorf("ZIP 文件条目过多")
	}

	bundle := &Bundle{}
	required := []requiredPDF{
		{key: "images", label: "总题库附图标记.pdf", nameNorm: normalizeName("总题库附图标记.pdf"), dest: &bundle.Images},
		{key: "total", label: "总题库.pdf", nameNorm: normalizeName("总题库.pdf"), dest: &bundle.Total},
		{key: "a", label: "A类题库.pdf", nameNorm: normalizeName("A类题库.pdf"), dest: &bundle.A},
		{key: "b", label: "B类题库.pdf", nameNorm: normalizeName("B类题库.pdf"), dest: &bundle.B},
		{key: "c", label: "C类题库.pdf", nameNorm: normalizeName("C类题库.pdf"), dest: &bundle.C},
	}

	var totalUncompressed uint64
	for _, f := range r.File {
		totalUncompressed += f.UncompressedSize64
		if totalUncompressed > maxTotalBytes {
			return nil, nil, fmt.Errorf("ZIP 解压后体积过大")
		}
	}

	found := map[string]bool{}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := decodeArchiveName(f)
		if isMacMetadata(name) {
			continue
		}
		base := filepath.Base(strings.ReplaceAll(name, "\\", "/"))
		norm := normalizeName(base)
		if norm == "" {
			continue
		}
		for i := range required {
			req := &required[i]
			if found[req.key] || !strings.Contains(norm, req.nameNorm) {
				continue
			}
			dest := filepath.Join(destDir, req.key+".pdf")
			if err := extractZipEntry(f, dest); err != nil {
				return nil, nil, fmt.Errorf("解压 %s 失败：%w", req.label, err)
			}
			*req.dest = dest
			found[req.key] = true
			break
		}
	}

	var missing []string
	for _, req := range required {
		if !found[req.key] {
			missing = append(missing, req.label)
		}
	}
	if len(missing) > 0 {
		return nil, nil, fmt.Errorf("ZIP 中缺少必需文件：%s", strings.Join(missing, "、"))
	}

	// 提示多余文件（不阻断导入）
	var warnings []string
	known := 0
	extra := 0
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := decodeArchiveName(f)
		if isMacMetadata(name) {
			continue
		}
		extra++
		norm := normalizeName(filepath.Base(strings.ReplaceAll(name, "\\", "/")))
		for _, req := range required {
			if strings.Contains(norm, req.nameNorm) {
				known++
				break
			}
		}
	}
	if unused := extra - known; unused > 0 {
		warnings = append(warnings, fmt.Sprintf("ZIP 中另有 %d 个未被使用的文件", unused))
	}

	return bundle, warnings, nil
}

func extractZipEntry(f *zip.File, dest string) error {
	if f.UncompressedSize64 > maxSingleFileBytes {
		return fmt.Errorf("文件过大（%d MB）", f.UncompressedSize64>>20)
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	n, err := io.Copy(out, io.LimitReader(rc, maxSingleFileBytes+1))
	if err != nil {
		return err
	}
	if n > maxSingleFileBytes {
		return fmt.Errorf("文件过大")
	}
	return nil
}

// isMacMetadata 过滤 macOS 的 __MACOSX 资源叉与 ._ 前缀文件。
func isMacMetadata(name string) bool {
	p := strings.ToLower(strings.ReplaceAll(name, "\\", "/"))
	if strings.Contains(p, "__macosx/") {
		return true
	}
	base := filepath.Base(p)
	return strings.HasPrefix(base, "._")
}

func decodeArchiveName(f *zip.File) string {
	name := f.Name
	// 很多 ZIP（尤其是 macOS/跨平台工具）写入的是合法 UTF-8 文件名但没有设置
	// UTF-8 标志位，此时不能按 GBK 解码，否则会把中文名解坏。
	if !utf8.ValidString(name) {
		if decoded, err := simplifiedchinese.GBK.NewDecoder().String(name); err == nil {
			return decoded
		}
	}
	return name
}

func normalizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "　", "")
	return s
}
