#!/usr/bin/env bash
# KA Harness 集成测试集
set -uo pipefail

cd "$(dirname "$0")/.."
go build -o /tmp/ka-harness ./cmd/harness

PASS=0
FAIL=0

run_test() {
  local name="$1"
  local expected="$2"
  shift 2
  local output
  output=$("$@" 2>&1)
  local code=$?

  if [ "$expected" = "exit_0" ] && [ "$code" -eq 0 ]; then
    echo "  ✓ $name (exit 0)"
    PASS=$((PASS+1))
  elif [ "$expected" = "exit_nonzero" ] && [ "$code" -ne 0 ]; then
    echo "  ✓ $name (exit $code, expected nonzero)"
    PASS=$((PASS+1))
  else
    echo "  ✗ $name (exit $code, expected $expected)"
    FAIL=$((FAIL+1))
  fi
}

echo "=== 1. 基本命令 ==="
run_test "--version 输出版本" exit_0 /tmp/ka-harness --version
run_test "--help 输出帮助" exit_0 /tmp/ka-harness --help
run_test "无命令显示帮助" exit_0 /tmp/ka-harness

echo ""
echo "=== 2. run 子命令 ==="
run_test "正常任务执行" exit_0 /tmp/ka-harness run "test task"
run_test "空 task 报错" exit_nonzero /tmp/ka-harness run ""
run_test "未知命令报错" exit_nonzero /tmp/ka-harness unknowncmd

echo ""
echo "=== 3. 分层输出验证 ==="
# stdout 应该只有最终答复，stderr 应该有控制层事件
stdout=$(/tmp/ka-harness run "test" 2>/dev/null)
stderr=$(/tmp/ka-harness run "test" 2>&1 >/dev/null)

if echo "$stdout" | grep -q "stub 模拟"; then
  echo "  ✓ 内容层在 stdout"
  PASS=$((PASS+1))
else
  echo "  ✗ 内容层不在 stdout"
  FAIL=$((FAIL+1))
fi

if echo "$stderr" | grep -q "✦ KA Harness"; then
  echo "  ✓ 控制层 banner 在 stderr"
  PASS=$((PASS+1))
else
  echo "  ✗ 控制层 banner 不在 stderr"
  FAIL=$((FAIL+1))
fi

echo ""
echo "=== 4. 符号白名单验证 ==="
# 只检查是否有 emoji（Unicode emoji 范围），中文标点和符号不算
bad=$(echo "$stderr" | python3 -c "
import sys
text = sys.stdin.read()
emoji_ranges = [
    (0x1F300, 0x1F5FF), (0x1F600, 0x1F64F),
    (0x1F680, 0x1F6FF), (0x1F900, 0x1F9FF),
]
bad_chars = set()
for ch in text:
    cp = ord(ch)
    for lo, hi in emoji_ranges:
        if lo <= cp <= hi:
            bad_chars.add(ch)
print(''.join(bad_chars))
")
if [ -z "$bad" ]; then
  echo "  ✓ 无 emoji，符号均在白名单内"
  PASS=$((PASS+1))
else
  echo "  ✗ 发现 emoji: $bad"
  FAIL=$((FAIL+1))
fi

echo ""
echo "=== 5. 工具执行验证 ==="
if echo "$stderr" | grep -q "read_file(path=README.md)"; then
  echo "  ✓ read_file 工具调用输出正确"
  PASS=$((PASS+1))
else
  echo "  ✗ read_file 工具调用缺失"
  FAIL=$((FAIL+1))
fi

echo ""
echo "=== 结果 ==="
echo "PASS: $PASS / $((PASS+FAIL))"
echo "FAIL: $FAIL / $((PASS+FAIL))"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
