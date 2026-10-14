// Package signal 是跨平台信号处理的回调签名层。
//
// M1 阶段不注册真实信号。Windows / Unix 分支实现留到 M4
// 引入子进程时再落（D6：业务代码不得出现平台分支）。
//
// 上层（Loop）只订阅三个回调，不直接处理信号：
//   - OnPaused:    第一次 SIGINT，进入暂停态
//   - OnResumed:   用户选择继续
//   - OnAborted:   第二次 SIGINT / 强制退出
package signal

// Router 是信号事件路由器（M1 空实现，仅定义接口）。
type Router struct {
	onPaused   func()
	onResumed  func()
	onAborted  func(exitCode int)
}

// NewRouter 创建空路由器。
func NewRouter() *Router {
	return &Router{}
}

// OnPaused 注册暂停回调。
func (r *Router) OnPaused(cb func()) { r.onPaused = cb }

// OnResumed 注册恢复回调。
func (r *Router) OnResumed(cb func()) { r.onResumed = cb }

// OnAborted 注册强制退出回调。
func (r *Router) OnAborted(cb func(exitCode int)) { r.onAborted = cb }

// M4 才会实现 Install()/Uninstall()；M1 不写。
