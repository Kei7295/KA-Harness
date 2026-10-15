// Package events 提供分层输出与统一错误模板。
//
// 分层规则（HARNESS_INTERFACE_SPEC.md §3.1）：
//   - 控制层（banner / 错误 / 审批 / 进度）→ stderr，直写不缓冲
//   - 内容层（文件正文 / 最终答复 / diff）→ stdout，bufio 行缓冲 + defer flush
package events

import (
	"bufio"
	"fmt"
	"os"
)

// EmitControl 写一条控制层事件到 stderr，直写不缓冲。
func EmitControl(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// EmitBanner 打印启动 banner（控制层）。
func EmitBanner(version string) {
	EmitControl("✦ KA Harness v%s", version)
}

// ContentWriter 返回一个行缓冲的内容层 writer，调用方必须 defer Flush。
func ContentWriter() *bufio.Writer {
	return bufio.NewWriter(os.Stdout)
}

// ErrorAction 是错误模板里的一条可行动作。
type ErrorAction struct {
	Label string
	Verb  string
}

// EmitError 走统一错误模板（SPEC §3）：
// 三要素 = 发生了什么 / 影响 / 下一步可行动作。
// 第一行遵循 `<program>: <message>`（P4-R2 GCC 风格）。
func EmitError(code, message, impact string, actions []ErrorAction, retryable bool, exitCode int) {
	EmitControl("✗  harness: %s", message)
	EmitControl("   影响：%s", impact)
	EmitControl("   下一步：")
	for i, a := range actions {
		EmitControl("     %d) %s", i+1, a.Label)
	}
	// cause / code / retryable 默认折叠，不打印（P6 简洁）。
}
