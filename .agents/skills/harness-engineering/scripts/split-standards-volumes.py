#!/usr/bin/env python3
"""One-off helper: split large standards presets into thin entry + volume files."""
from __future__ import annotations

import re
import sys
from pathlib import Path

STANDARDS = Path(__file__).resolve().parent.parent / "assets" / "standards"

# Roman/Chinese numeral section titles → slug
SECTION_SLUGS: dict[str, str] = {}

ENTRY_KEEP_SECTIONS = {"本次必读决策表", "规范落地优先级"}


def section_slug(title: str) -> str:
    m = re.match(r"^##\s+([一二三四五六七八九十]+)、(.+)$", title.strip())
    if not m:
        return "misc"
    num, rest = m.group(1), m.group(2)
    num_map = {
        "一": "01", "二": "02", "三": "03", "四": "04", "五": "05",
        "六": "06", "七": "07", "八": "08", "九": "09", "十": "10",
        "十一": "11", "十二": "12", "十三": "13", "十四": "14",
    }
    prefix = num_map.get(num, "99")
    slug = re.sub(r"[^\w\u4e00-\u9fff]+", "-", rest).strip("-").lower()
    slug = slug[:40] if slug else "section"
    return f"{prefix}-{slug}"


def parse_sections(text: str) -> tuple[str, list[tuple[str, str]]]:
    """Return (preamble, [(heading_line, body), ...])."""
    parts = re.split(r"(?m)^(?=## )", text)
    preamble = parts[0].rstrip()
    sections: list[tuple[str, str]] = []
    for part in parts[1:]:
        if not part.strip():
            continue
        lines = part.split("\n", 1)
        heading = lines[0].strip()
        body = lines[1] if len(lines) > 1 else ""
        sections.append((heading, body.rstrip()))
    return preamble, sections


def one_line_desc(heading: str) -> str:
    m = re.match(r"^##\s+[一二三四五六七八九十]+、(.+)$", heading)
    return m.group(1) if m else heading.lstrip("# ").strip()


def build_entry(
    stem: str,
    preamble: str,
    kept: list[tuple[str, str]],
    volumes: list[tuple[str, str, str]],
    extra_blocks: list[str] | None = None,
) -> str:
    lines: list[str] = []
    lines.append(preamble.rstrip())
    lines.append("")
    for heading, body in kept:
        lines.append(heading)
        lines.append(body.rstrip())
        lines.append("")
        lines.append("---")
        lines.append("")

    if extra_blocks:
        for block in extra_blocks:
            lines.append(block.rstrip())
            lines.append("")
            lines.append("---")
            lines.append("")

    lines.append("## 分册目录")
    lines.append("")
    lines.append("| 分册 | 说明 |")
    lines.append("|------|------|")
    for fname, heading, _ in volumes:
        desc = one_line_desc(heading)
        lines.append(f"| [{fname}](./{stem}/{fname}) | {desc} |")
    lines.append("")
    lines.append("---")
    lines.append("")
    lines.append("## 章节快速索引")
    lines.append("")
    lines.append(
        f"> 接入仓 `docs/standards/README.md` 的「章节快速索引」会汇总本入口与各分册标题；"
        f"按任务 **Read 对应分册**（可用 offset/limit），禁止默认全文灌入所有分册。"
    )
    lines.append("")
    return "\n".join(lines).rstrip() + "\n"


QUALITY_OPT_IN = """## 加载说明（opt-in）

> **非凡改代码必读。** 本规范在 code-project 中仍会部署到 `docs/standards/`，但 Agent 仅在 **Code Review / MR 评审 / 质量审查** 任务时按预算加载；日常功能开发不必预读全文。

## 本次必读决策表

> Agent：Review 任务先读本表，再按任务 Read 对应分册；禁止默认全文灌入。

| 类型 | 决策（摘要） | 详见 |
|------|-------------|------|
| 禁止 | 无解释地 `[必须]` 否决；人身攻击式评论 | 分册 05、三 |
| 必须 | 问题分级 `[必须]`/`[建议]`/`[Nit]`；说明「为什么」 | 二、分册 05 |
| 验证 | 评审报告含结论与待办；评分量表仅 Review 需要时读 | 分册 06、四 |"""


def split_file(stem: str, extra_entry: list[str] | None = None) -> None:
    src = STANDARDS / f"{stem}.md"
    text = src.read_text(encoding="utf-8")
    preamble, sections = parse_sections(text)

    kept: list[tuple[str, str]] = []
    body_sections: list[tuple[str, str]] = []
    for heading, body in sections:
        title = heading.lstrip("# ").strip()
        if title in ENTRY_KEEP_SECTIONS:
            kept.append((heading, body))
        else:
            body_sections.append((heading, body))

    vol_dir = STANDARDS / stem
    vol_dir.mkdir(exist_ok=True)
    for old in vol_dir.glob("*.md"):
        old.unlink()

    volumes: list[tuple[str, str, str]] = []
    for heading, body in body_sections:
        fname = section_slug(heading) + ".md"
        vol_path = vol_dir / fname
        content = f"{heading}\n\n{body}\n" if body else f"{heading}\n"
        vol_path.write_text(content, encoding="utf-8")
        volumes.append((fname, heading, body))

    entry = build_entry(stem, preamble, kept, volumes, extra_entry)
    src.write_text(entry, encoding="utf-8")
    line_count = len(entry.splitlines())
    print(f"{stem}: entry {line_count} lines, {len(volumes)} volumes")


def main() -> None:
    split_file("api-grpc-gateway")
    split_file("backend-trpc-go")
    split_file("backend-trpc-agent-go")
    split_file("backend-go-micro")
    split_file("quality-code-review", extra_entry=[QUALITY_OPT_IN])
    for stem in [
        "api-grpc-gateway",
        "backend-trpc-go",
        "backend-trpc-agent-go",
        "backend-go-micro",
        "quality-code-review",
    ]:
        entry = STANDARDS / f"{stem}.md"
        n = len(entry.read_text(encoding="utf-8").splitlines())
        if n > 120:
            print(f"WARN: {stem}.md entry has {n} lines (>120)", file=sys.stderr)
            sys.exit(1)


if __name__ == "__main__":
    main()
