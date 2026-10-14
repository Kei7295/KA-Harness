#!/usr/bin/env bash
# M1 自动化设计检查（D1 / D2 / D6 / D7）。
set -euo pipefail

cd "$(dirname "$0")/.."

bin=$(mktemp -d)/harness
trap 'rm -rf "$(dirname "$bin")"' EXIT
go build -o "$bin" ./cmd/harness

fail=0

echo "== D1: 错误行首 'harness: ' 前缀 =="
out=$("$bin" run "demo task" 2>&1 >/dev/null || true)
if echo "$out" | grep -q "^✗  harness: "; then
  echo "OK"
else
  echo "FAIL: 错误行缺少 'harness: ' 前缀"
  echo "--- 实际输出 ---"
  echo "$out"
  fail=1
fi

echo
echo "== D2: 退出码 ∈ 白名单 {0,1,64,65,69,70,75,77,130} =="
set +e
"$bin" run "demo task" >/dev/null 2>&1
code=$?
set -e
echo "run 退出码: $code"
case "$code" in
  0|1|64|65|69|70|75|77|130) echo "OK" ;;
  *) echo "FAIL: 退出码 $code 不在白名单"; fail=1 ;;
esac

echo
echo "== D6: 业务代码无平台分支 (runtime.GOOS / runtime.GOARCH) =="
if grep -rE "runtime\.GOOS|runtime\.GOARCH" internal/ ; then
  echo "FAIL: 发现平台分支，应封装在 internal/platform/ 下"
  fail=1
else
  echo "OK"
fi

echo
echo "== D7: 源码无 emoji 表情（白名单符号 ✓⚠✗○▸✦⏸ 除外） =="
if python3 scripts/check_emoji.py; then
  :
else
  fail=1
fi

echo
if [ "$fail" -eq 0 ]; then
  echo "ALL CHECKS PASSED"
else
  echo "CHECKS FAILED"
fi
exit "$fail"
