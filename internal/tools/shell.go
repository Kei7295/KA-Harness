package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

func init() {
	DefaultRegistry.Register(Tool{
		Name:        "shell.exec",
		Description: "在工作目录下执行 shell 命令",
		Handler:     shellExecHandler,
	})
}

func shellExecHandler(args map[string]any) Envelope {
	command, ok := args["command"].(string)
	if !ok || command == "" {
		return Failure("TOOL_INVALID_ARGS", "shell.exec 需要 string 类型的 command 参数", false)
	}

	workDir, ok := args["work_dir"].(string)
	if !ok || workDir == "" {
		workDir = "."
	}

	// 解析工作目录（走路径白名单）
	resolved, err := CheckPath(workDir)
	if err != nil {
		return Failure("FS_PATH_OUT_OF_SCOPE", fmt.Sprintf("工作目录越界: %s", workDir), false)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Dir = resolved

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return Failure("INTERNAL_ASSERT", fmt.Sprintf("执行失败: %v", err), true)
		}
	}

	data := map[string]any{
		"exit_code": exitCode,
		"stdout":    stdout.String(),
		"stderr":    stderr.String(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return Failure("INTERNAL_ASSERT", fmt.Sprintf("序列化失败: %v", err), true)
	}

	return Success("application/json", string(jsonData))
}
