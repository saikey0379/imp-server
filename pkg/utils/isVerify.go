package utils

import (
	"fmt"
	"regexp"
)

func IsValidDomainName(domain string) bool {
	// 使用正则表达式检查域名格式
	pattern := `^([a-zA-Z0-9-_]+\.)+[a-zA-Z]{2,}$`
	match, err := regexp.MatchString(pattern, domain)
	if err != nil {
		fmt.Println("Regex error:", err)
		return false
	}
	return match
}

func IsValidPort(port int) bool {
	if port > 0 && port < 65535 {
		return true
	}
	return false
}

func IsValidEmail(email string) bool {
	// 定义邮箱格式的正则表达式
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	// 使用正则表达式匹配邮箱格式
	match, err := regexp.MatchString(emailPattern, email)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	return match
}
