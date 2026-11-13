package importer

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"crac_exam_v3/backend/config"
	"crac_exam_v3/backend/models"
)

var specialMapping = map[string]string{"LK060": "LK0603"}

// ImportZip 解析题库 ZIP（仅支持 ZIP），返回题目与统计信息。此时尚未写库。
func ImportZip(zipPath string, progress func(models.ImportProgress)) ([]models.Question, *models.ImportResult, error) {
	start := time.Now()
	report := func(stage, message string, percent int) {
		if progress != nil {
			progress(models.ImportProgress{Stage: stage, Message: message, Percent: percent})
		}
	}

	if strings.ToLower(filepath.Ext(zipPath)) != ".zip" {
		return nil, nil, fmt.Errorf("仅支持 ZIP 格式的题库压缩包")
	}
	info, err := os.Stat(zipPath)
	if err != nil {
		return nil, nil, fmt.Errorf("无法读取文件：%w", err)
	}
	if info.IsDir() {
		return nil, nil, fmt.Errorf("请选择 ZIP 文件")
	}
	if info.Size() > 600<<20 {
		return nil, nil, fmt.Errorf("ZIP 文件过大（>600MB）")
	}

	tempDir, err := os.MkdirTemp(config.ImportTempDir(), "import-*")
	if err != nil {
		return nil, nil, fmt.Errorf("创建临时目录失败：%w", err)
	}
	defer os.RemoveAll(tempDir)

	report("unzip", "正在校验并解压题库包…", 5)
	bundle, warnings, err := ExtractBundle(zipPath, tempDir)
	if err != nil {
		return nil, nil, err
	}

	report("parse", "正在解析总题库…", 20)
	totalText, err := ExtractText(bundle.Total)
	if err != nil {
		return nil, nil, fmt.Errorf("解析总题库失败：%w", err)
	}
	questions, parseStats := ParseQuestions(totalText)
	if len(questions) == 0 {
		return nil, nil, fmt.Errorf("未能从总题库中解析出任何题目，请检查 PDF 格式")
	}
	if parseStats.Skipped > 0 {
		warnings = append(warnings, fmt.Sprintf("有 %d 条不完整题目被跳过", parseStats.Skipped))
	}

	// J/P/I 是题目的联合唯一键；重复项只保留第一条，避免写库时触发唯一索引。
	duplicateJPI := 0
	{
		seen := make(map[string]struct{}, len(questions))
		unique := make([]models.Question, 0, len(questions))
		for _, q := range questions {
			key := q.J + "\x1f" + q.P + "\x1f" + q.I
			if _, ok := seen[key]; ok {
				duplicateJPI++
				continue
			}
			seen[key] = struct{}{}
			unique = append(unique, q)
		}
		questions = unique
	}
	if duplicateJPI > 0 {
		warnings = append(warnings, fmt.Sprintf("跳过 %d 道 J/P/I 联合键完全相同的重复题目", duplicateJPI))
	}

	report("category", "正在解析分类题库…", 40)
	categoryValues := map[string]map[string]bool{}
	for _, item := range []struct {
		key   string
		label string
		path  string
	}{
		{"LA", "A", bundle.A},
		{"LB", "B", bundle.B},
		{"LC", "C", bundle.C},
	} {
		text, err := ExtractText(item.path)
		if err != nil {
			return nil, nil, fmt.Errorf("解析 %s 类题库失败：%w", item.label, err)
		}
		values := ParseIValues(text)
		if len(values) == 0 {
			return nil, nil, fmt.Errorf("未能从 %s 类题库中提取题号", item.label)
		}
		categoryValues[item.key] = values
	}

	report("images", "正在提取题目图片（纯 Go 解析）…", 60)
	imageMap, imageWarnings := buildImageMap(bundle.Images)
	warnings = append(warnings, imageWarnings...)

	report("merge", "正在合并题目数据…", 75)
	duplicateJ := 0
	jSeen := map[string]int{}
	stats := map[string]int{
		"single": 0, "multiple": 0, "categoryA": 0, "categoryB": 0, "categoryC": 0,
		"withImage": 0, "noCategory": 0,
	}
	for i := range questions {
		q := &questions[i]
		iValue := compactSpace(q.I)
		if categoryValues["LA"][iValue] {
			q.LA = 1
			stats["categoryA"]++
		}
		if categoryValues["LB"][iValue] {
			q.LB = 1
			stats["categoryB"]++
		}
		if categoryValues["LC"][iValue] {
			q.LC = 1
			stats["categoryC"]++
		}
		if q.LA == 0 && q.LB == 0 && q.LC == 0 {
			stats["noCategory"]++
		}
		if q.Type == 1 {
			stats["single"]++
		} else {
			stats["multiple"]++
		}
		q.F = lookupImage(imageMap, q.J)
		if q.F != "" {
			stats["withImage"]++
		}
		jSeen[q.J]++
		if jSeen[q.J] == 2 {
			duplicateJ++
		}
	}
	if duplicateJ > 0 {
		warnings = append(warnings, fmt.Sprintf(
			"题号 J 有 %d 个重复值（J/P/I 联合唯一，题目本身不重复；仅按 J 关联图片时可能不准确）", duplicateJ))
	}
	stats["duplicateJ"] = duplicateJ
	stats["duplicateJPI"] = duplicateJPI

	report("done", "解析完成，准备写入数据库…", 85)
	return questions, &models.ImportResult{
		Success:      true,
		Message:      "题库解析完成",
		Imported:     len(questions),
		Total:        len(questions),
		Stats:        stats,
		Warnings:     warnings,
		DuplicateJ:   duplicateJ,
		DuplicateJPI: duplicateJPI,
		ElapsedMS:    time.Since(start).Milliseconds(),
	}, nil
}

func buildImageMap(imagesPDF string) (map[string]string, []string) {
	var warnings []string
	out := map[string]string{}

	pageTexts, err := ExtractPageTexts(imagesPDF)
	if err != nil {
		return out, []string{"附图标记 PDF 解析失败，本次导入将不含图片：" + err.Error()}
	}
	var tokens []string
	for _, t := range pageTexts {
		tokens = append(tokens, ParseImageTokens(t)...)
	}

	pages, err := ExtractImages(imagesPDF)
	if err != nil {
		return out, []string{"图片提取失败，本次导入将不含图片：" + err.Error()}
	}
	var images []ImageData
	for _, page := range pages {
		images = append(images, page...)
	}
	if len(images) == 0 {
		return out, []string{"附图 PDF 中未提取到可用图片"}
	}
	if len(tokens) != len(images) {
		warnings = append(warnings, fmt.Sprintf("附图标记数量（%d）与图片数量（%d）不一致，按顺序尽量匹配", len(tokens), len(images)))
	}

	n := len(tokens)
	if len(images) < n {
		n = len(images)
	}
	for i := 0; i < n; i++ {
		key := strings.TrimSpace(tokens[i])
		if key == "" {
			continue
		}
		if mapped, ok := specialMapping[key]; ok {
			key = mapped
		}
		if _, exists := out[key]; exists {
			continue
		}
		out[key] = images[i].DataURL
	}
	return out, warnings
}

func lookupImage(imageMap map[string]string, j string) string {
	j = strings.TrimSpace(j)
	if j == "" {
		return ""
	}
	if v, ok := imageMap[j]; ok {
		return v
	}
	base := j
	if idx := strings.Index(base, "."); idx > 0 {
		base = base[:idx]
	}
	if v, ok := imageMap[base]; ok {
		return v
	}
	if mapped, ok := specialMapping[base]; ok {
		return imageMap[mapped]
	}
	return ""
}

func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
