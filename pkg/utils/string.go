package utils

import (
	"strings"
)

func SubString(str string, begin int, length int) string {
	// 将字符串的转换成[]rune
	rs := []rune(str)
	lth := len(rs)

	// 简单的越界判断
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}

	end := begin + length
	if end > lth {
		end = lth
	}

	// 返回子串
	return string(rs[begin:end])
}

func IsInArrayStr(str string, arr []string) bool {
	for _, value := range arr {
		if str == value {
			return true
		}
	}
	return false
}

func StrSplitAny(s string, seps string) []string {
	splitter := func(r rune) bool {
		return strings.ContainsRune(seps, r)
	}
	return strings.FieldsFunc(s, splitter)
}

func HasDuplicate(arr []string) bool {
	seen := make(map[string]bool)

	for _, num := range arr {
		if seen[num] {
			return true // 如果元素已经在map中出现过，说明有重复
		}
		seen[num] = true
	}

	return false // 没有重复元素
}
