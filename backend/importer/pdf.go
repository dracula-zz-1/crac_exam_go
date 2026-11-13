package importer

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// ImageData 从 PDF 中提取出的图片。
type ImageData struct {
	DataURL string
	FileTyp string
}

// ExtractText 使用纯 Go 库提取 PDF 全部文本，并按页剔除页码。
func ExtractText(path string) (string, error) {
	texts, err := ExtractPageTexts(path)
	if err != nil {
		return "", err
	}
	for i := range texts {
		texts[i] = stripPageNumber(texts[i], i+1)
	}
	return strings.Join(texts, "\n"), nil
}

// stripPageNumber 仅剔除与页序一致的页首/页尾纯数字行。
// 不按“纯数字”全局过滤，避免误删题目中的行内数字（如 50Ω、200 欧）。
func stripPageNumber(pageText string, pageNo int) string {
	want := strconv.Itoa(pageNo)
	lines := strings.Split(pageText, "\n")

	// 仅处理页首：该题库的页码固定在每页第一行；页尾数字往往是题目内容。
	for i, line := range lines {
		s := strings.TrimSpace(line)
		if s == "" {
			continue
		}
		if s == want {
			lines = append(lines[:i], lines[i+1:]...)
		}
		break
	}
	return strings.Join(lines, "\n")
}

// ExtractPageTexts 逐页提取文本，保留页序。
func ExtractPageTexts(path string) ([]string, error) {
	f, reader, err := pdf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开 PDF 失败: %w", err)
	}
	defer f.Close()

	pageCount := reader.NumPage()
	out := make([]string, 0, pageCount)
	for i := 1; i <= pageCount; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			out = append(out, "")
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			// 单页失败不应中断整份文档
			out = append(out, "")
			continue
		}
		out = append(out, text)
	}
	return out, nil
}

// ExtractImages 逐页提取图片，页内按对象号排序，尽量与原 PyMuPDF 的 xref 顺序一致。
func ExtractImages(path string) ([][]ImageData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	conf := model.NewDefaultConfiguration()
	pages, err := api.ExtractImagesRaw(ctx, f, nil, conf)
	if err != nil {
		return nil, fmt.Errorf("提取 PDF 图片失败: %w", err)
	}

	out := make([][]ImageData, 0, len(pages))
	for _, page := range pages {
		objNrs := make([]int, 0, len(page))
		for objNr := range page {
			objNrs = append(objNrs, objNr)
		}
		sort.Ints(objNrs)

		pageImages := make([]ImageData, 0, len(objNrs))
		for _, objNr := range objNrs {
			img := page[objNr]
			data, err := io.ReadAll(img.Reader)
			if err != nil || len(data) == 0 {
				continue
			}
			fileType := strings.ToLower(strings.TrimPrefix(img.FileType, "."))
			mime := mimeFor(fileType)
			if mime == "" {
				// 浏览器无法直接显示的格式（jpx/jbig2）跳过
				continue
			}
			pageImages = append(pageImages, ImageData{
				DataURL: "data:" + mime + ";base64," + encodeBase64(data),
				FileTyp: fileType,
			})
		}
		out = append(out, pageImages)
	}
	return out, nil
}

func mimeFor(fileType string) string {
	switch fileType {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "bmp":
		return "image/bmp"
	default:
		return ""
	}
}
