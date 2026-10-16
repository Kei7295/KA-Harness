package tools

import (
	"fmt"
	"os"
)

func init() {
	DefaultRegistry.Register(Tool{
		Name:        "write_file",
		Description: "写入文件内容（覆盖模式）",
		Handler:     writeFileHandler,
	})
}

func writeFileHandler(args map[string]any) Envelope {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return Failure("TOOL_INVALID_ARGS", "write_file 需要 string 类型的 path 参数", false)
	}

	content, ok := args["content"].(string)
	if !ok {
		return Failure("TOOL_INVALID_ARGS", "write_file 需要 string 类型的 content 参数", false)
	}

	resolved, err := CheckPath(path)
	if err != nil {
		return Failure("FS_PATH_OUT_OF_SCOPE", fmt.Sprintf("路径越界: %s", path), false)
	}

	if err := os.WriteFile(resolved, []byte(content), 0644); err != nil {
		return Failure("INTERNAL_ASSERT", fmt.Sprintf("写入失败: %v", err), true)
	}

	return Success("application/json", fmt.Sprintf(`{"written_bytes": %d, "path": %q}`, len(content), path))
}
