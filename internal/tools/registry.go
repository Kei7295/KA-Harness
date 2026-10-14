package tools

import (
	"fmt"
	"sort"
)

// Handler 是工具执行函数。
type Handler func(args map[string]any) Envelope

// Tool 是注册的工具。
type Tool struct {
	Name        string
	Description string
	Handler     Handler
}

// Registry 是工具注册表。
type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

// Register 注册一个工具。
func (r *Registry) Register(t Tool) {
	r.tools[t.Name] = t
}

// Get 获取工具。
func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// List 列出所有注册的工具名（排序）。
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.tools))
	for n := range r.tools {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Execute 执行指定工具，返回 envelope。
func (r *Registry) Execute(name string, args map[string]any) Envelope {
	t, ok := r.tools[name]
	if !ok {
		return Failure("TOOL_NOT_FOUND", fmt.Sprintf("工具 %q 未注册", name), false)
	}
	return t.Handler(args)
}
