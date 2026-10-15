// Package cli 是 M1 接入层：命令解析与分发。
//
// 性能约束（SPEC §10.5 / D5）：
//   - --version / --help 路径只 import stdlib + 本包。
//   - 不引入 cobra/viper/urfave 等第三方 CLI 库（拖启动几十毫秒）。
package cli

import (
	"flag"
	"fmt"
	"os"

	"aiharness/internal/events"
	"aiharness/internal/model"
	"aiharness/internal/orchestrator"
)

const version = "0.1.0"

// Main 是 CLI 入口。
func Main() {
	if len(os.Args) < 2 {
		events.EmitBanner(version)
		printUsage()
		return
	}

	switch os.Args[1] {
	case "--version", "-version":
		fmt.Printf("harness %s\n", version)
		return
	case "--help", "-h", "help":
		printUsage()
		return
	case "run":
		cmdRun(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "harness: unknown command %q\n", os.Args[1])
		printUsage()
		os.Exit(64) // EX_USAGE
	}
}

func printUsage() {
	fmt.Fprintln(os.Stdout, `usage: ka-harness [--version] [--help] <command> ...

KA Harness — AI Agent engineering framework

commands:
  run <task>   Run a task in natural language

options:
  --version    Show version and exit
  --help       Show this help message`)
}

func cmdRun(argv []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	_ = fs.Bool("allow-write-in-ci", false,
		"Allow write operations in non-TTY/CI (not recommended for interactive use)")
	if err := fs.Parse(argv); err != nil {
		os.Exit(64)
	}
	task := fs.Arg(0)

	events.EmitBanner(version)

	m := model.NewStubModel()
	o := orchestrator.New(m)
	code := o.Run(task)
	os.Exit(code)
}
