package services

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"io"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"crac_exam_go/backend/models"
	"crac_exam_go/backend/utils"

	"github.com/gen2brain/go-fitz"
	"github.com/oliverpool/unipdf/v3/extractor"
	"github.com/oliverpool/unipdf/v3/model"
)

// PDFImporter PDF 导入器
type PDFImporter struct {
	specialMapping map[string]string
	requiredPDFs   map[string]string
}

// ProgressCallback 进度回调函数类型
type ProgressCallback func(current, total int, message string)

// NewPDFImporter 创建 PDF 导入器实例
func NewPDFImporter() *PDFImporter {
	return &PDFImporter{
		specialMapping: map[string]string{"LK060": "LK0603"},
		requiredPDFs: map[string]string{
			"total":  "总题库.pdf",
			"images": "总题库附图标记.pdf",
			"a":      "A类题库.pdf",
			"b":      "B类题库.pdf",
			"c":      "C类题库.pdf",
		},
	}
}

// ProcessPDFData 处理 PDF 文件导入（主入口）
func (i *PDFImporter) ProcessPDFData(zipPath string, progress ProgressCallback) ([]*models.Question, error) {
	totalStart := time.Now()
	utils.Info("PDFImporter", "开始处理 PDF 文件", map[string]interface{}{
		"zip_path": zipPath,
	})

	// 解压 ZIP 文件
	if progress != nil {
		progress(0, 100, "正在解压 ZIP 文件...")
	}
	utils.Info("PDFImporter", "开始解压 ZIP 文件", map[string]interface{}{
		"zip_path": zipPath,
	})
	pdfDataMap, err := i.extractPDFsFromZip(zipPath)
	if err != nil {
		utils.Error("PDFImporter", "解压 ZIP 文件失败", err, nil)
		return nil, fmt.Errorf("解压 ZIP 文件失败：%v", err)
	}
	utils.Info("PDFImporter", "ZIP 文件解压完成", map[string]interface{}{
		"pdfs_found": len(pdfDataMap),
		"keys":       []string{"total", "images", "a", "b", "c"},
	})

	// 并发处理各个 PDF 文件
	type textResult struct {
		text string
		err  error
	}
	type imgResult struct {
		images map[string]string
		err    error
	}
	type catResult struct {
		a, b, c []string
	}

	textChan := make(chan textResult, 1)
	imgChan := make(chan imgResult, 1)
	catChan := make(chan catResult, 1)

	// 启动 goroutine 处理总题库文本
	go func() {
		if progress != nil {
			progress(10, 100, "正在提取总题库文本...")
		}
		text, err := i.readTotalPDFTextParallel(pdfDataMap["total"])
		textChan <- textResult{text: text, err: err}
	}()

	// 启动 goroutine 处理图片
	go func() {
		if progress != nil {
			progress(20, 100, "正在提取图片...")
		}
		images, err := i.extractImagesFromPDF(pdfDataMap["images"])
		imgChan <- imgResult{images: images, err: err}
	}()

	// 启动 goroutine 处理分类题库
	go func() {
		if progress != nil {
			progress(30, 100, "正在处理分类题库...")
		}
		var a, b, c []string
		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer wg.Done()
			a, _ = i.processCategoryPDFParallel(pdfDataMap["a"], "A 类题库.pdf")
		}()
		go func() {
			defer wg.Done()
			b, _ = i.processCategoryPDFParallel(pdfDataMap["b"], "B 类题库.pdf")
		}()
		go func() {
			defer wg.Done()
			c, _ = i.processCategoryPDFParallel(pdfDataMap["c"], "C 类题库.pdf")
		}()
		wg.Wait()
		catChan <- catResult{a: a, b: b, c: c}
	}()

	// 等待文本提取完成
	textRes := <-textChan
	if textRes.err != nil {
		utils.Error("PDFImporter", "读取总题库失败", textRes.err, nil)
		return nil, fmt.Errorf("读取总题库失败：%v", textRes.err)
	}

	// 等待图片提取完成
	imgRes := <-imgChan
	if imgRes.err != nil {
		utils.Warn("PDFImporter", "提取图片失败", map[string]interface{}{
			"error": imgRes.err,
		})
		imgRes.images = make(map[string]string)
	}

	// 等待分类题库处理完成
	catRes := <-catChan

	// 打印分类数据日志
	utils.Info("PDFImporter", "分类题库数据", map[string]interface{}{
		"a_class_count": len(catRes.a),
		"b_class_count": len(catRes.b),
		"c_class_count": len(catRes.c),
	})
	if len(catRes.a) > 0 && len(catRes.a) <= 20 {
		utils.Info("PDFImporter", "A 类题库数据样例", map[string]interface{}{
			"examples": catRes.a,
		})
	}
	if len(catRes.b) > 0 && len(catRes.b) <= 20 {
		utils.Info("PDFImporter", "B 类题库数据样例", map[string]interface{}{
			"examples": catRes.b,
		})
	}
	if len(catRes.c) > 0 && len(catRes.c) <= 20 {
		utils.Info("PDFImporter", "C 类题库数据样例", map[string]interface{}{
			"examples": catRes.c,
		})
	}

	// 解析题目
	if progress != nil {
		progress(60, 100, "正在解析题目...")
	}
	questions := i.parseQuestions(textRes.text, imgRes.images)

	// 打印图片提取日志
	utils.Info("PDFImporter", "图片提取结果", map[string]interface{}{
		"images_map_size": len(imgRes.images),
	})
	if len(imgRes.images) > 0 && len(imgRes.images) <= 20 {
		keys := make([]string, 0, len(imgRes.images))
		for k := range imgRes.images {
			keys = append(keys, k)
		}
		utils.Info("PDFImporter", "图片数据样例", map[string]interface{}{
			"j_values": keys,
		})
	}

	// 处理分类标记
	aSet := make(map[string]bool)
	bSet := make(map[string]bool)
	cSet := make(map[string]bool)
	for _, v := range catRes.a {
		aSet[v] = true
	}
	for _, v := range catRes.b {
		bSet[v] = true
	}
	for _, v := range catRes.c {
		cSet[v] = true
	}

	laCount, lbCount, lcCount := 0, 0, 0
	for idx := range questions {
		if aSet[questions[idx].I] {
			questions[idx].LA = 1
			laCount++
		}
		if bSet[questions[idx].I] {
			questions[idx].LB = 1
			lbCount++
		}
		if cSet[questions[idx].I] {
			questions[idx].LC = 1
			lcCount++
		}
	}

	// 打印分类标记日志
	utils.Info("PDFImporter", "分类标记统计", map[string]interface{}{
		"la_count":        laCount,
		"lb_count":        lbCount,
		"lc_count":        lcCount,
		"total_questions": len(questions),
	})

	// 打印前 10 道题目的分类信息
	if len(questions) > 0 {
		sampleData := make([]map[string]interface{}, 0, 10)
		for i := 0; i < len(questions) && i < 10; i++ {
			sampleData = append(sampleData, map[string]interface{}{
				"id":    i + 1,
				"J":     questions[i].J,
				"I":     questions[i].I,
				"LA":    questions[i].LA,
				"LB":    questions[i].LB,
				"LC":    questions[i].LC,
				"F_len": len(questions[i].F),
			})
		}
		utils.Info("PDFImporter", "前 10 道题目样例", map[string]interface{}{
			"questions": sampleData,
		})
	}

	// 统计信息
	imageCount := 0
	for _, q := range questions {
		if q.F != "" {
			imageCount++
		}
	}

	utils.Info("PDFImporter", "PDF 文件处理完成", map[string]interface{}{
		"total_questions":  len(questions),
		"image_questions":  imageCount,
		"a_class":          laCount,
		"b_class":          lbCount,
		"c_class":          lcCount,
		"duration_seconds": time.Since(totalStart).Seconds(),
	})

	if progress != nil {
		progress(90, 100, fmt.Sprintf("解析完成，共 %d 道题目", len(questions)))
	}

	if progress != nil {
		progress(100, 100, "PDF 处理完成")
	}

	return questions, nil
}

// readTotalPDFTextParallel 并行读取总题库 PDF 文本（go-fitz，快速）
func (i *PDFImporter) readTotalPDFTextParallel(pdfData []byte) (string, error) {
	start := time.Now()
	utils.Debug("PDFImporter", "开始并行提取总题库文本", nil)

	doc, err := fitz.NewFromMemory(pdfData)
	if err != nil {
		return "", err
	}
	defer doc.Close()

	totalPages := doc.NumPage()
	utils.Debug("PDFImporter", "总题库页数", map[string]interface{}{
		"pages": totalPages,
	})

	type pageResult struct {
		pageNum int
		text    string
		err     error
	}

	resultChan := make(chan pageResult, totalPages)
	var wg sync.WaitGroup

	for pageNum := 0; pageNum < totalPages; pageNum++ {
		wg.Add(1)
		go func(pn int) {
			defer wg.Done()
			globalSem <- struct{}{}
			defer func() { <-globalSem }()

			text, err := doc.Text(pn)
			if err != nil {
				resultChan <- pageResult{pageNum: pn, err: err}
				return
			}

			resultChan <- pageResult{pageNum: pn, text: text, err: nil}
		}(pageNum)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	results := make([]pageResult, 0, totalPages)
	for res := range resultChan {
		results = append(results, res)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].pageNum < results[j].pageNum
	})

	var allLines []string
	for _, res := range results {
		if res.err == nil {
			lines := strings.Split(res.text, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !isPureNumber(line) {
					allLines = append(allLines, line)
				}
			}
		}
	}

	result := strings.Join(allLines, "\n")

	// 打印前 50 行样例
	sampleLines := strings.SplitN(result, "\n", 50)
	utils.Info("PDFImporter", "PDF 文本提取样例（前 50 行）", map[string]interface{}{
		"sample": sampleLines,
	})

	utils.Debug("PDFImporter", "总题库文本提取完成", map[string]interface{}{
		"lines":            len(allLines),
		"total_chars":      len(result),
		"duration_seconds": time.Since(start).Seconds(),
	})
	return result, nil
}

// processCategoryPDFParallel 并行处理分类题库 PDF
func (i *PDFImporter) processCategoryPDFParallel(pdfData []byte, name string) ([]string, error) {
	start := time.Now()
	utils.Debug("PDFImporter", fmt.Sprintf("开始并行提取分类题库 %s...", name), nil)

	doc, err := fitz.NewFromMemory(pdfData)
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	totalPages := doc.NumPage()
	utils.Debug("PDFImporter", fmt.Sprintf("  %s 总页数", name), map[string]interface{}{
		"pages": totalPages,
	})

	type pageResult struct {
		pageNum int
		values  []string
		err     error
	}

	resultChan := make(chan pageResult, totalPages)
	var wg sync.WaitGroup

	for pageNum := 0; pageNum < totalPages; pageNum++ {
		wg.Add(1)
		go func(pn int) {
			defer wg.Done()
			globalSem <- struct{}{}
			defer func() { <-globalSem }()

			text, err := doc.Text(pn)
			if err != nil {
				resultChan <- pageResult{pageNum: pn, err: err}
				return
			}

			var pageValues []string
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "[I]") {
					value := strings.TrimSpace(line[3:])
					if value != "" {
						pageValues = append(pageValues, value)
					}
				}
			}

			resultChan <- pageResult{pageNum: pn, values: pageValues, err: nil}
		}(pageNum)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	results := make([]pageResult, 0, totalPages)
	for res := range resultChan {
		results = append(results, res)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].pageNum < results[j].pageNum
	})

	var allValues []string
	for _, res := range results {
		if res.err == nil {
			allValues = append(allValues, res.values...)
		}
	}

	utils.Debug("PDFImporter", fmt.Sprintf("分类题库 %s 提取完成", name), map[string]interface{}{
		"values":           len(allValues),
		"duration_seconds": time.Since(start).Seconds(),
	})
	return allValues, nil
}

// extractJValuesFromImagesPDF 从附图 PDF 中提取 J 值
func (i *PDFImporter) extractJValuesFromImagesPDF(pdfData []byte) ([]string, error) {
	start := time.Now()
	utils.Debug("PDFImporter", "开始提取附图 J 值", nil)

	doc, err := fitz.NewFromMemory(pdfData)
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	var jValues []string
	for pageNum := 0; pageNum < doc.NumPage(); pageNum++ {
		text, err := doc.Text(pageNum)
		if err != nil {
			continue
		}
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				jValues = append(jValues, line)
			}
		}
	}

	utils.Debug("PDFImporter", "附图 J 值提取完成", map[string]interface{}{
		"count":            len(jValues),
		"duration_seconds": time.Since(start).Seconds(),
	})
	return jValues, nil
}

// extractImagesFromPDF 从 PDF 中提取图片（unipdf，正确性高）
func (i *PDFImporter) extractImagesFromPDF(pdfData []byte) (map[string]string, error) {
	start := time.Now()
	utils.Debug("PDFImporter", "开始提取图片", nil)

	// 提取 J 值
	jValues, err := i.extractJValuesFromImagesPDF(pdfData)
	if err != nil {
		return nil, err
	}
	utils.Debug("PDFImporter", "J 值数量", map[string]interface{}{
		"count": len(jValues),
	})

	for idx, jv := range jValues {
		if mapped, ok := i.specialMapping[jv]; ok {
			jValues[idx] = mapped
		}
	}

	// 使用 unipdf 提取图片
	reader, err := model.NewPdfReader(bytes.NewReader(pdfData))
	if err != nil {
		return nil, err
	}

	numPages, err := reader.GetNumPages()
	if err != nil {
		return nil, err
	}

	var allImagesData [][]byte
	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page, err := reader.GetPage(pageNum)
		if err != nil {
			continue
		}

		ex, err := extractor.New(page)
		if err != nil {
			continue
		}

		opts := extractor.ImageExtractOptions{}
		pageImages, err := ex.ExtractPageImages(&opts)
		if err != nil {
			continue
		}

		for _, imgMark := range pageImages.Images {
			if imgMark.Image == nil {
				continue
			}

			goImg, err := imgMark.Image.ToGoImage()
			if err != nil {
				continue
			}

			var buf bytes.Buffer
			if err := png.Encode(&buf, goImg); err != nil {
				continue
			}
			allImagesData = append(allImagesData, buf.Bytes())
		}
	}

	utils.Debug("PDFImporter", "提取到原始图片数据", map[string]interface{}{
		"count": len(allImagesData),
	})

	imagesMap := make(map[string]string)
	minLen := len(allImagesData)
	if len(jValues) < minLen {
		minLen = len(jValues)
	}

	for idx := 0; idx < minLen; idx++ {
		imgBase64 := base64.StdEncoding.EncodeToString(allImagesData[idx])
		imagesMap[jValues[idx]] = fmt.Sprintf("data:image/png;base64,%s", imgBase64)
	}

	utils.Debug("PDFImporter", "图片提取完成", map[string]interface{}{
		"count":            len(imagesMap),
		"duration_seconds": time.Since(start).Seconds(),
	})
	return imagesMap, nil
}

// parseQuestions 从文本内容解析题目数据（使用正则表达式，与 Python 版本一致）
func (i *PDFImporter) parseQuestions(textContent string, imagesMap map[string]string) []*models.Question {
	start := time.Now()
	utils.Debug("PDFImporter", "开始解析题目", nil)

	// 预处理文本内容，过滤页码行
	lines := strings.Split(textContent, "\n")
	filteredLines := []string{}
	for _, line := range lines {
		cleanLine := strings.ReplaceAll(line, " ", "")
		cleanLine = strings.ReplaceAll(cleanLine, "\t", "")
		cleanLine = strings.ReplaceAll(cleanLine, "\r", "")
		cleanLine = strings.TrimSpace(cleanLine)

		// 忽略纯数字行（页码）
		isPageNumber := false
		if cleanLine != "" {
			isPageNumber = true
			for _, c := range cleanLine {
				if c < '0' || c > '9' {
					isPageNumber = false
					break
				}
			}
		}

		if cleanLine != "" && !isPageNumber {
			filteredLines = append(filteredLines, line)
		}
	}
	filteredText := strings.Join(filteredLines, "\n")

	var questions []*models.Question
	currentQuestion := make(map[string]string)

	// 使用正则表达式匹配字段标记：支持 [X] 和 X] 两种格式
	// 只匹配标记位置，内容在后续处理中提取
	pattern := regexp.MustCompile(`\[?([JPIQTFABCD])\]`)
	matches := pattern.FindAllStringIndex(filteredText, -1)

	for idx, match := range matches {
		field := filteredText[match[0]:match[1]]
		// 提取字段字母
		fieldMatch := regexp.MustCompile(`([JPIQTFABCD])`).FindStringSubmatch(field)
		if len(fieldMatch) < 2 {
			continue
		}
		fieldCode := fieldMatch[1]

		// 提取字段内容：从当前标记结束到下一个标记开始
		var value string
		if idx+1 < len(matches) {
			value = filteredText[match[1]:matches[idx+1][0]]
		} else {
			value = filteredText[match[1]:]
		}

		value = strings.TrimSpace(value)
		// 清理字段值：去除末尾页码和换行符
		cleanValue := cleanFieldValue(value)

		if fieldCode == "J" {
			// 检查是否是同一行中的多个 J 字段
			isSameLineJ := len(currentQuestion) == 1 && currentQuestion["J"] != ""

			if isSameLineJ {
				// 同一行中的多个 J 字段，合并
				currentQuestion["J"] = currentQuestion["J"] + cleanValue
			} else if len(currentQuestion) > 0 {
				// 遇到新的 J 字段，保存当前题目并开始新题目
				questions = append(questions, convertToQuestion(currentQuestion, imagesMap, i.specialMapping))
				currentQuestion = map[string]string{"J": cleanValue}
			} else {
				// 第一个 J 字段
				currentQuestion["J"] = cleanValue
			}
		} else {
			// 正常处理其他字段
			currentQuestion[fieldCode] = cleanValue
		}
	}

	// 添加最后一个题目
	if len(currentQuestion) > 0 {
		questions = append(questions, convertToQuestion(currentQuestion, imagesMap, i.specialMapping))
	}

	for idx := range questions {
		if len(questions[idx].T) == 1 {
			questions[idx].Type = 1
		} else if len(questions[idx].T) > 1 {
			questions[idx].Type = 2
		}
	}

	utils.Debug("PDFImporter", "题目解析完成", map[string]interface{}{
		"count":            len(questions),
		"duration_seconds": time.Since(start).Seconds(),
	})
	return questions
}

// extractPDFsFromZip 从 ZIP 文件中提取 PDF 文件
func (i *PDFImporter) extractPDFsFromZip(zipPath string) (map[string][]byte, error) {
	start := time.Now()
	utils.Debug("PDFImporter", "开始解压 ZIP 文件", nil)

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	pdfDataMap := make(map[string][]byte)
	allPDFNames := []string{}

	for _, f := range r.File {
		name := f.Name
		if strings.Contains(name, "__MACOSX") || strings.Contains(name, ".DS_Store") {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(name), ".pdf") {
			continue
		}

		allPDFNames = append(allPDFNames, name)
		utils.Debug("PDFImporter", "找到 PDF 文件", map[string]interface{}{
			"name": name,
		})

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}

		nameLower := strings.ToLower(name)
		utils.Debug("PDFImporter", "尝试匹配 PDF 文件", map[string]interface{}{
			"name": name,
		})

		for key, required := range i.requiredPDFs {
			requiredLower := strings.ToLower(required)
			utils.Debug("PDFImporter", "检查是否包含", map[string]interface{}{
				"name":     name,
				"required": required,
				"match":    strings.Contains(nameLower, requiredLower),
			})
			if strings.Contains(nameLower, requiredLower) {
				pdfDataMap[key] = data
				utils.Info("PDFImporter", "匹配成功", map[string]interface{}{
					"name": name,
					"key":  key,
				})
				break
			}
		}
	}

	utils.Info("PDFImporter", "ZIP 解压完成", map[string]interface{}{
		"pdfs_found":       len(pdfDataMap),
		"keys_found":       getMapKeys(pdfDataMap),
		"all_pdf_names":    allPDFNames,
		"duration_seconds": time.Since(start).Seconds(),
	})

	// 显示哪些 PDF 文件没找到
	requiredKeys := []string{"total", "images", "a", "b", "c"}
	missingKeys := []string{}
	for _, key := range requiredKeys {
		if _, exists := pdfDataMap[key]; !exists {
			missingKeys = append(missingKeys, key)
		}
	}
	if len(missingKeys) > 0 {
		utils.Warn("PDFImporter", "以下 PDF 文件未找到", map[string]interface{}{
			"missing_keys": missingKeys,
		})
	}

	return pdfDataMap, nil
}

// getMapKeys 获取 map 的所有键
func getMapKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// isPureNumber 检查是否为纯数字
func isPureNumber(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0 && len(s) <= 4
}

// cleanJValue 清理 J 值
func cleanJValue(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "[J]", "")
	return strings.TrimSpace(s)
}

// cleanFieldValue 清理字段值：去除末尾页码和换行符
func cleanFieldValue(value string) string {
	// 去除首尾空白字符
	strippedValue := strings.TrimSpace(value)
	// 去除值末尾的页码信息（数字和空白字符）
	noPageValue := regexp.MustCompile(`\s+\d+\s*$`).ReplaceAllString(strippedValue, "")
	// 完全去除内部的换行符
	noNewlineValue := regexp.MustCompile(`[\r\n]+`).ReplaceAllString(noPageValue, "")
	return noNewlineValue
}

// convertToQuestion 将 map 转换为 Question 对象
func convertToQuestion(qMap map[string]string, imagesMap map[string]string, specialMapping map[string]string) *models.Question {
	question := &models.Question{}

	// 获取 J 值并处理特殊映射
	jValue := qMap["J"]
	if mapped, ok := specialMapping[jValue]; ok {
		jValue = mapped
	}
	question.J = jValue

	// 复制其他字段
	if v, ok := qMap["P"]; ok {
		question.P = v
	}
	if v, ok := qMap["I"]; ok {
		question.I = v
	}
	if v, ok := qMap["Q"]; ok {
		question.Q = v
	}
	if v, ok := qMap["T"]; ok {
		question.T = v
	}
	if v, ok := qMap["A"]; ok {
		question.A = v
	}
	if v, ok := qMap["B"]; ok {
		question.B = v
	}
	if v, ok := qMap["C"]; ok {
		question.C = v
	}
	if v, ok := qMap["D"]; ok {
		question.D = v
	}

	// 处理图片
	if img, ok := imagesMap[question.J]; ok {
		question.F = img
	} else {
		question.F = ""
	}

	// 根据 T 字段长度判断题型
	if len(question.T) == 1 {
		question.Type = 1
	} else if len(question.T) > 1 {
		question.Type = 2
	}

	return question
}

// 全局信号量，限制总并发数
var globalSem chan struct{}

func init() {
	globalSem = make(chan struct{}, runtime.NumCPU())
}
