package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// DefaultRegistry 是默认工具注册表。
var DefaultRegistry = NewRegistry()

func init() {
	DefaultRegistry.Register(Tool{
		Name:        "list_dir",
		Description: "列出目录内容",
		Handler:     listDirHandler,
	})
}

func listDirHandler(args map[string]any) Envelope {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return Failure("TOOL_INVALID_ARGS", "list_dir 需要 string 类型的 path 参数", false)
	}

	resolved, err := CheckPath(path)
	if err != nil {
		return Failure("FS_PATH_OUT_OF_SCOPE", fmt.Sprintf("路径越界: %s", path), false)
	}

	entries, err := os.ReadDir(resolved)
	if err != nil {
		if os.IsNotExist(err) {
			return Failure("FS_NOT_FOUND", fmt.Sprintf("目录不存在: %s", path), false)
		}
		return Failure("INTERNAL_ASSERT", fmt.Sprintf("读取目录失败: %v", err), true)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		names = append(names, name)
	}
	sort.Strings(names)

	body := ""
	for _, n := range names {
		body += n + "\n"
	}

	return Success("text/plain", body)
}

// init 确保 read_file 的 init 先执行（Go 包内 init 顺序按文件名字母序）
var _ = filepath.Join
