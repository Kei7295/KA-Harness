package tools

import (
	"fmt"
	"os"
)

func init() {
	DefaultRegistry.Register(Tool{
		Name:        "read_file",
		Description: "读取文件内容",
		Handler:     readFileHandler,
	})
}

func readFileHandler(args map[string]any) Envelope {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return Failure("TOOL_INVALID_ARGS", "read_file 需要 string 类型的 path 参数", false)
	}

	resolved, err := CheckPath(path)
	if err != nil {
		return Failure("FS_PATH_OUT_OF_SCOPE", fmt.Sprintf("路径越界: %s", path), false)
	}

	content, err := os.ReadFile(resolved)
	if err != nil {
		if os.IsNotExist(err) {
			return Failure("FS_NOT_FOUND", fmt.Sprintf("文件不存在: %s", path), false)
		}
		return Failure("INTERNAL_ASSERT", fmt.Sprintf("读取失败: %v", err), true)
	}

	return Success("text/plain", string(content))
}
