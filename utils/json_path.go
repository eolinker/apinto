package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	allowedPattern = regexp.MustCompile(`^\$\.[a-zA-Z0-9_.*\[\]]+$`)
	maxDepth       = 6
	maxArrayCount  = 3
	maxPathLength  = 100
)

func ValidateJSONPath(path string) error {
	path = strings.TrimSpace(path)

	// 1️⃣ 长度限制
	if len(path) == 0 || len(path) > maxPathLength {
		return fmt.Errorf("invalid path length")
	}

	// 2️⃣ 字符白名单
	if !allowedPattern.MatchString(path) {
		return fmt.Errorf("invalid characters in path")
	}

	// 3️⃣ 必须 $. 开头
	if !strings.HasPrefix(path, "$.") {
		return fmt.Errorf("path must start with $.")
	}

	// 4️⃣ 禁止递归
	if strings.Contains(path, "..") {
		return fmt.Errorf("recursive path not allowed")
	}

	// 5️⃣ 深度限制
	depth := strings.Count(path, ".")
	if depth > maxDepth {
		return fmt.Errorf("path depth exceeds limit")
	}

	// 6️⃣ 数组节点限制
	arrayCount := strings.Count(path, "[*]")
	if arrayCount > maxArrayCount {
		return fmt.Errorf("too many array selectors")
	}

	// 7️⃣ 非法组合
	if strings.Contains(path, "[") && !strings.Contains(path, "[*]") {
		return fmt.Errorf("only [*] array selector allowed")
	}

	return nil
}
