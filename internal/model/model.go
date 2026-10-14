// Package model 是 LLM 模型层的抽象。
//
// M2 阶段是 stub：返回固定的 tool_call 提议，不接真实 API。
// 接口设计遵循 §10.3 envelope 契约。
package model

import "fmt"

// ToolCall 是模型提议的一次工具调用。
type ToolCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// Display 返回人类可读的工具调用一行描述。
// 格式约定：tool_name(arg1=val1, arg2=val2)
// 每个工具在 M3 可覆盖自己的渲染（通过 registry 挂载），
// M2 用通用实现。
func (tc *ToolCall) Display() string {
	args := ""
	for k, v := range tc.Arguments {
		if args != "" {
			args += ", "
		}
		args += fmt.Sprintf("%s=%v", k, v)
	}
	return fmt.Sprintf("%s(%s)", tc.Name, args)
}

// Response 是模型对一轮对话的返回。
type Response struct {
	Text     string      `json:"text"`      // 最终答复（若无 tool_call）
	ToolCall *ToolCall   `json:"tool_call"` // 提议的工具调用（若有）
	Done     bool        `json:"done"`      // 是否完成
}

// Model 是 LLM 模型接口。
type Model interface {
	// Complete 接收用户任务，返回模型响应。
	Complete(task string, history []string) (*Response, error)
}

// StubModel 是 M2 阶段的占位模型：先提议 read_file，再返回最终答复。
type StubModel struct {
	callCount int
}

func NewStubModel() *StubModel {
	return &StubModel{}
}

func (m *StubModel) Complete(task string, history []string) (*Response, error) {
	m.callCount++

	// 第一轮：提议 read_file 工具调用
	if m.callCount == 1 {
		return &Response{
			Text: "",
			ToolCall: &ToolCall{
				Name: "read_file",
				Arguments: map[string]any{
					"path": "README.md",
				},
			},
			Done: false,
		}, nil
	}

	// 第二轮：返回最终答复
	return &Response{
		Text:     fmt.Sprintf("任务 %q 已完成（stub 模拟）。读取了 README.md 的内容，摘要如下：这是 Harness 项目的 README。", task),
		ToolCall: nil,
		Done:     true,
	}, nil
}
