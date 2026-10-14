#!/usr/bin/env bash
# 冷启动 bench（SPEC §10.5）。
# Go 单二进制：无解释器层，直接测端到端 wall-clock。
set -euo pipefail

cd "$(dirname "$0")/.."

# 1. 编译
bin=$(mktemp -d)/harness
go build -o "$bin" ./cmd/harness

# 2. 用 Python 精确测量（避免两次 python3 -c 的额外开销）
python3 - "$bin" << 'PYEOF'
import subprocess, sys, time, statistics

bin = sys.argv[1]
N = 30
WARMUP = 5

times = []
for i in range(N):
    t0 = time.perf_counter_ns()
    subprocess.run([bin, "--version"], capture_output=True)
    t1 = time.perf_counter_ns()
    times.append((t1 - t0) / 1_000_000)

# 丢弃 warmup
samples = sorted(times[WARMUP:])
p50 = statistics.median(samples)
p95 = samples[int(len(samples) * 0.95) - 1]

print(f"samples: {len(samples)} (warmup {WARMUP} dropped)")
print(f"P50 startup: {p50:.1f} ms   (target < 300ms)")
print(f"P95 startup: {p95:.1f} ms   (target < 300ms)")

if p95 >= 300:
    print(f"FAIL: P95 >= 300ms", file=sys.stderr)
    sys.exit(1)
print("OK: 端到端冷启动 < 300ms")
PYEOF

rm -rf "$(dirname "$bin")"
