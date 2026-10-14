package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CheckPath 校验路径是否在基准目录内（§10.9）。
// 返回规范化后的绝对路径，或 error。
func CheckPath(input string) (string, error) {
	base, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("无法获取工作目录: %w", err)
	}

	abs, err := filepath.Abs(filepath.Join(base, input))
	if err != nil {
		return "", fmt.Errorf("路径解析失败: %w", err)
	}

	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		resolved = abs
	}

	rel, err := filepath.Rel(base, resolved)
	if err != nil {
		return "", fmt.Errorf("路径越界: %w", err)
	}

	if rel == ".." || strings.HasPrefix(rel, "../") {
		return "", fmt.Errorf("%w: %s", ErrPathOutOfScope, input)
	}

	return resolved, nil
}

var ErrPathOutOfScope = fmt.Errorf("FS_PATH_OUT_OF_SCOPE")
