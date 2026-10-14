#!/usr/bin/env python3
"""D7 检查：源码无 emoji 表情（白名单符号除外）。

依据：HARNESS_INTERFACE_SPEC.md §10.7
退出码：0 = 通过，1 = 发现 emoji。
"""
import pathlib
import sys

EMOJI_RANGES = [
    (0x1F300, 0x1F5FF),  # Misc Symbols and Pictographs
    (0x1F600, 0x1F64F),  # Emoticons
    (0x1F680, 0x1F6FF),  # Transport and Map
    (0x1F900, 0x1F9FF),  # Supplemental Symbols and Pictographs
    (0x1FA00, 0x1FA6F),
    (0x1FA70, 0x1FAFF),
]

# §10.7 符号白名单
WHITELIST = {
    "\u2713",  # ✓ 成功 / 完成
    "\u26A0",  # ⚠ 警告
    "\u2717",  # ✗ 错误 / 失败
    "\u2726",  # ✦ 启动 / banner
    "\u25B8",  # ▸ 进行中
}


def main() -> int:
    bad = []
    # 扫 Go 源码 + 脚本
    patterns = ["**/*.go", "scripts/*.sh", "scripts/*.py"]
    files = []
    for p in patterns:
        files.extend(pathlib.Path(".").glob(p))

    for f in sorted(set(files)):
        if not f.is_file():
            continue
        text = f.read_text(encoding="utf-8")
        for lineno, ch in enumerate(text, 1):
            if ch in WHITELIST:
                continue
            if any(lo <= ord(ch) <= hi for lo, hi in EMOJI_RANGES):
                bad.append(f"{f}:{lineno} {ch!r}")
    if bad:
        print("FAIL: 发现 emoji 表情（§10.7 白名单外）:", file=sys.stderr)
        for line in bad:
            print(f"  {line}", file=sys.stderr)
        return 1
    print("OK: 无 emoji，符号均在 §10.7 白名单内")
    return 0


if __name__ == "__main__":
    sys.exit(main())
