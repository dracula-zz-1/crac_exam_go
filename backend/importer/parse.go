package importer

import (
	"regexp"
	"sort"
	"strings"
	"unicode"

	"crac_exam_v3/backend/models"
)

// 题目标记：支持 "[J]" 与 "J]" 两种写法。
var markerRe = regexp.MustCompile(`\[[JPIQTFABCD]\]|[JPIQTFABCD]\]`)

var iValueRe = regexp.MustCompile(`\[I\]\s*([^\[\r\n]*)`)

var requiredFields = []string{"J", "P", "I", "Q", "T", "A", "B", "C", "D"}

// ParseStats 解析统计。
type ParseStats struct {
	Total    int
	Single   int
	Multiple int
	Skipped  int
}

// ParseQuestions 从“总题库”PDF 文本中解析题目。
// 算法移植自原 Python 版 _parse_questions_from_text。
func ParseQuestions(text string) ([]models.Question, ParseStats) {
	var stats ParseStats
	// 注意：不能在这里丢弃“纯数字行”。该题库的 PDF 会把行内数字
	// （如 50Ω、200 欧）拆成独立的文本对象，纯数字行可能正是题目内容。
	// 页码在 ExtractText 中按页首/页尾精确剔除。
	filtered := text

	locs := markerRe.FindAllStringIndex(filtered, -1)
	if len(locs) == 0 {
		return nil, stats
	}

	current := map[string]string{}
	var maps []map[string]string

	for i, loc := range locs {
		marker := filtered[loc[0]:loc[1]]
		field := markerField(marker)

		end := len(filtered)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		value := cleanFieldValue(filtered[loc[1]:end])

		isSameLineJ := field == "J" && len(current) == 1 && current["J"] != ""
		switch field {
		case "J":
			if isSameLineJ {
				current["J"] += value
			} else {
				if len(current) > 0 {
					maps = append(maps, current)
				}
				current = map[string]string{"J": value}
			}
		default:
			current[field] = value
		}
	}
	if len(current) > 0 {
		maps = append(maps, current)
	}

	out := make([]models.Question, 0, len(maps))
	for _, m := range maps {
		complete := true
		for _, f := range requiredFields {
			if _, ok := m[f]; !ok {
				complete = false
				break
			}
		}
		if !complete {
			stats.Skipped++
			continue
		}
		q := models.Question{
			J: strings.TrimSpace(m["J"]),
			P: strings.TrimSpace(m["P"]),
			I: strings.TrimSpace(m["I"]),
			Q: strings.TrimSpace(m["Q"]),
			A: strings.TrimSpace(m["A"]),
			B: strings.TrimSpace(m["B"]),
			C: strings.TrimSpace(m["C"]),
			D: strings.TrimSpace(m["D"]),
		}
		q.T = normalizeAnswer(m["T"])
		if q.J == "" || q.Q == "" || q.T == "" {
			stats.Skipped++
			continue
		}
		if len([]rune(q.T)) == 1 {
			q.Type = 1
			stats.Single++
		} else {
			q.Type = 2
			stats.Multiple++
		}
		out = append(out, q)
	}
	stats.Total = len(out)
	return out, stats
}

// ParseIValues 从分类题库 PDF 中提取所有 [I] 值（用于判定 A/B/C 类别）。
func ParseIValues(text string) map[string]bool {
	out := map[string]bool{}
	for _, m := range iValueRe.FindAllStringSubmatch(text, -1) {
		v := compactSpace(m[1])
		if v != "" {
			out[v] = true
		}
	}
	return out
}

// ParseImageTokens 从“总题库附图标记”PDF 文本中按出现顺序提取题号。
func ParseImageTokens(text string) []string {
	return strings.Fields(text)
}

func markerField(marker string) string {
	if len(marker) >= 3 && marker[0] == '[' {
		return string(marker[1])
	}
	if len(marker) >= 2 {
		return string(marker[0])
	}
	return ""
}

func compactSpace(s string) string {
	var b strings.Builder
	for _, r := range s {
		if !unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func cleanFieldValue(v string) string {
	v = strings.TrimSpace(v)
	// 不要在这里删除末尾数字：PDF 会把行内数字拆成独立文本行，
	// 例如选项 “1、0、0、0” 的最后一位会被误删。页码已在 ExtractText 按页剔除。
	return joinTextLines(v)
}

// joinTextLines 合并 PDF 的文本对象分段。
// 纯 Go 提取器会在每个 BT（文本对象）处换行，题干里的数字/英文常常单独成行；
// 直接删换行会把 "RMKS IS" 变成 "RMKSIS"，也会让 "50Ω" 与前面的中文粘连。
// 这里按相邻字符类型补一个空格：中文↔ASCII、ASCII↔ASCII 需要空格；
// 中文↔中文、标点相邻、数字与 Ω/μ 等单位符号之间不补。
func joinTextLines(v string) string {
	lines := strings.Split(strings.ReplaceAll(v, "\r", ""), "\n")
	var b strings.Builder
	var prev rune
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		runes := []rune(line)
		first := runes[0]
		if b.Len() > 0 && needsSpace(prev, first) {
			b.WriteByte(' ')
		}
		b.WriteString(line)
		prev = runes[len(runes)-1]
	}
	return strings.TrimSpace(b.String())
}

func needsSpace(prev, next rune) bool {
	if unicode.IsSpace(prev) || unicode.IsSpace(next) {
		return false
	}
	pa, na := isASCIIAlnum(prev), isASCIIAlnum(next)
	pc, nc := unicode.Is(unicode.Han, prev), unicode.Is(unicode.Han, next)
	return (pa && (na || nc)) || (na && pc)
}

func isASCIIAlnum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func normalizeAnswer(s string) string {
	s = strings.ToUpper(compactSpace(s))
	seen := map[rune]bool{}
	var letters []rune
	for _, r := range s {
		if r >= 'A' && r <= 'D' && !seen[r] {
			seen[r] = true
			letters = append(letters, r)
		}
	}
	sort.Slice(letters, func(i, j int) bool { return letters[i] < letters[j] })
	return string(letters)
}
