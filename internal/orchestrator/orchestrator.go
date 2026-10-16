// Package orchestrator 是 ReAct 编排循环。
//
// 遵循 §10 契约：
//   - 模型只提议 tool_call，禁止直接触达文件系统
//   - 事件流走 stderr 控制层，最终答复走 stdout 内容层
//   - 最大循环次数防止无限循环
package orchestrator

import (
	"bufio"
	"fmt"
	"os"

	"aiharness/internal/events"
	"aiharness/internal/model"
	"aiharness/internal/tools"
)

const maxIterations = 10

// Orchestrator 是 ReAct 循环。
type Orchestrator struct {
	model model.Model
}

func New(m model.Model) *Orchestrator {
	return &Orchestrator{model: m}
}

// Run 执行任务，返回退出码。
func (o *Orchestrator) Run(task string) int {
	if task == "" {
		events.EmitError(
			"EMPTY_TASK",
			"任务内容为空",
			"请提供要执行的任务描述。",
			[]events.ErrorAction{
				{Label: "重新输入任务", Verb: "retry"},
			},
			false,
			64, // EX_USAGE
		)
		return 64
	}

	events.EmitControl("▸ 开始任务：%s", task)

	var history []string
	content := bufio.NewWriter(os.Stdout)
	defer content.Flush()

	for i := 1; i <= maxIterations; i++ {
		events.EmitControl("▸ 第 %d 轮：调用模型", i)

		resp, err := o.model.Complete(task, history)
		if err != nil {
			events.EmitError(
				"MODEL_ERROR",
				fmt.Sprintf("模型调用失败：%v", err),
				"任务终止，未产生副作用。",
				[]events.ErrorAction{
					{Label: "重试", Verb: "retry"},
				},
				true,
				69, // EX_UNAVAILABLE
			)
			return 69
		}

		if resp.ToolCall != nil {
			events.EmitControl("▸ %s", resp.ToolCall.Display())

			result := tools.DefaultRegistry.Execute(resp.ToolCall.Name, resp.ToolCall.Arguments)
			if err := result.Validate(); err != nil {
				events.EmitError(
					"ENVELOPE_INVALID",
					fmt.Sprintf("工具返回 envelope 不合法: %v", err),
					"工具执行结果格式异常。",
					nil,
					false,
					70,
				)
				return 70
			}

			// 把工具调用 + 结果完整喂回模型（真正的 ReAct）
			history = append(history, fmt.Sprintf("我调用了工具 %s，参数: %v", resp.ToolCall.Name, resp.ToolCall.Arguments))

			if !result.OK {
				events.EmitControl("  ✗ %s: %s", result.Error.Code, result.Error.Message)
				history = append(history, fmt.Sprintf("工具执行失败，错误码 %s: %s", result.Error.Code, result.Error.Message))
			} else {
				// 截断长输出，避免 context 爆炸
				output := result.Data.Body
				if len(output) > 2000 {
					output = output[:2000] + "\n... (truncated)"
				}
				events.EmitControl("  ✓ 完成（%d bytes）", len(result.Data.Body))
				history = append(history, fmt.Sprintf("工具输出:\n%s", output))
			}
			continue
		}

		if resp.Done {
			events.EmitControl("✓ 任务完成")
			// 最终答复走 stdout 内容层
			content.WriteString(resp.Text + "\n")
			content.Flush()
			return 0
		}

		// 无 tool_call 且未 done = 异常状态
		events.EmitError(
			"ORCHESTRATOR_ASSERT",
			"模型返回既无 tool_call 也未标记完成",
			"循环异常终止。",
			nil,
			false,
			70, // EX_SOFTWARE
		)
		return 70
	}

	events.EmitError(
		"MAX_ITERATIONS",
		fmt.Sprintf("超过最大循环次数 %d", maxIterations),
		"任务未完成，可能模型陷入循环。",
		[]events.ErrorAction{
			{Label: "缩小任务范围", Verb: "refine_task"},
		},
		false,
		70, // EX_SOFTWARE — 逻辑错误，重跑不会好
	)
	return 70
}
