package utils

import (
	"math/rand"
	"sort"
)

// ShuffleOptions 打乱选项列表顺序（用于考试/练习中选项随机化）
func ShuffleOptions(options []map[string]string) {
	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})
}

// ShuffleOptionsCopy 打乱选项列表顺序并返回新切片（不修改原切片）
func ShuffleOptionsCopy(options []map[string]string) []map[string]string {
	result := make([]map[string]string, len(options))
	copy(result, options)
	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	return result
}

// SortString 对字符串按字符排序（用于多选题答案比较）
func SortString(s string) string {
	chars := []rune(s)
	sort.Slice(chars, func(i, j int) bool {
		return chars[i] < chars[j]
	})
	return string(chars)
}
