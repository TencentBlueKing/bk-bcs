#!/usr/bin/env python3
"""
PostToolUse hook for tapd-story-pipeline cost tracking.

Triggered by ${CLAUDE_PROJECT_DIR}/.codebuddy/settings.json (合并自
skills/tapd-story-pipeline/scripts/settings.json) when matcher=Task/Agent fires.

输入：stdin 含 hook payload（PostToolUse 事件 JSON）
副作用：在 ${WORK_DIR}/cost-events.jsonl 追加一行
失败语义：任何异常都吞掉并 print {"continue": true}，绝不阻塞主流程
日志输出：写入 ${CLAUDE_PROJECT_DIR}/.codebuddy/hooks-debug.log 便于排查
Payload 示例：
{
    "session_id": "69aea43098bb441983eef8dc78afd2f4",
    "transcript_path": "xxx",
    "cwd": "/",
    "hook_event_name": "PostToolUse",
    "tool_name": "Task",
    "tool_input": {
        "subagent_name": "subagent名字",
        "description": "...",
        "prompt": "... [pipeline-cost-marker] iter_dir=specs/v0.9.x story_id=123 action=execute ts=2026-05-26T11:30:00+08:00 ..."
    },
    "tool_response": {
        "type": "task_tool_result",
        "finalResult": "xxx",
        "toolInfo": [],
        "toolCallBrief": "Execution Summary: 16 tool uses, cost: 48.83s",
        "usage": {
            "inputTokens": 103491,
            "outputTokens": 4238,
            "totalTokens": 107729,
            "cacheTokens": 69696,
            "cachedWriteTokens": 0,
            "cachedMissTokens": 33795,
            "credit": 0
        }
    },
    "generation_id": "...",
    "model": "...",
    "client": "CodeBuddyIDE",
    "version": "4.9.10"
}

关键字段路径：
- credit: tool_response.usage.credit（在 usage 内部，非顶层）
- tokens: tool_response.usage.{inputTokens, outputTokens, cacheTokens}（camelCase）
- duration: 从 tool_response.toolCallBrief 中正则提取 "cost: Xs"
"""
import json
import logging
import os
import re
import sys
import traceback
from datetime import datetime


# ─── 日志配置 ───────────────────────────────────────────────────────────────
# todo(DeveloperJim): runner和pipeline共用的日志文件，后续可能存在写冲突，需要优化
LOG_DIR = os.environ.get("CLAUDE_PROJECT_DIR", "/tmp")
LOG_FILE = os.path.join(LOG_DIR, ".codebuddy", "hooks-debug.log")

# 确保日志目录存在
os.makedirs(os.path.dirname(LOG_FILE), exist_ok=True)

logging.basicConfig(
    filename=LOG_FILE,
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger("hook-log-usage")


# ─── 业务逻辑 ───────────────────────────────────────────────────────────────
SPEC_RE = re.compile(
    r"\[spec-cost-marker\]\s+"
    r"work_dir=(?P<work_dir>\S+)\s+"
    r"stage=(?P<stage>\S+)\s+"
    r"attempt=(?P<attempt>\d+)\s+"
    r"round=(?P<round>\d+)\s+"
    r"ts=(?P<ts>\S+)"
)
PIPELINE_RE = re.compile(
    r"\[pipeline-cost-marker\]\s+"
    r"iter_dir=(?P<iter_dir>\S+)\s+"
    r"story_id=(?P<story_id>\S+)\s+"
    r"action=(?P<action>\S+)\s+"
    r"ts=(?P<ts>\S+)"
)

DURATION_RE = re.compile(r"cost:\s*([0-9.]+)s", re.IGNORECASE)

# 兼容新旧工具名: Claude Code 曾将 Agent 工具命名为 "Task"，现已更名为 "Agent"
ACCEPTED_TOOL_NAMES = {"Task", "Agent"}

# tapd-story-pipeline 各阶段实际调度的 subagent（speckit-execution-agent 承接工具性
# speckit 命令；tech-lead / backend-developer / frontend-developer / qa-engineer 承接
# specify/plan/tasks/implement/validate 等专业化阶段），均需纳入 cost 统计
TRACKED_AGENTS = {
    "speckit-execution-agent",
    "code-reviewer",
    "tech-lead",
    "backend-developer",
    "frontend-developer",
    "qa-engineer",
}

def main():
    try:
        raw_input = sys.stdin.read()
        payload = json.loads(raw_input)
        tool_name = payload.get("tool_name", "<unknown>")
        
        if tool_name not in ACCEPTED_TOOL_NAMES:
            logger.debug("Skipped: tool_name '%s' not in accepted set",
                        tool_name)
            return _ok()

        # 正式开始处理，基于subagent做逻辑拆分
        logger.info("=" * 49)
        subagent = payload.get("tool_input", {}).get("subagent_name", "")
        if subagent in TRACKED_AGENTS:
            handle_speckit_execution_agent(
                payload.get("session_id"), 
                payload.get("tool_input"), 
                payload.get("tool_response", {}))
        elif subagent == "tapd-pipeline-agent":
            handle_tapd_pipeline_agent(
                payload.get("session_id"), 
                payload.get("tool_input"), 
                payload.get("tool_response", {}))
        else:
            logger.info("Skipped: subagent_name '%s' not in accepted set",
                        subagent)
            return _ok()
        
    except Exception as exc:
        logger.error("Hook exception: %s\n%s", exc, traceback.format_exc())
    return _ok()

def handle_speckit_execution_agent(session, inputs, response) -> None:
    """
    处理speckit-execution-agent subagent事件，记录usage信息
    inputs: tool_input，结构参照19行
    response: tool_response，结构参照24行
    """
    # 从prompt中提取marker，确认处于哪个spec阶段
    prompt = inputs.get("prompt", "")
    m = SPEC_RE.search(prompt)
    if not m:
        logger.info("Skipped: no [spec-cost-marker] found in prompt (len=%d)",
                    len(prompt))
        return
    work_dir_rel = m.group("work_dir")
    work_dir_abs = os.path.join(os.environ.get("CLAUDE_PROJECT_DIR"), work_dir_rel)
    if not os.path.isdir(work_dir_abs):
        logger.warning("Skipped: absolute work_dir does not exist: %s", work_dir_abs)
        return
    # 提取usage信息
    usage, duration_sec = normalize_usage_and_duration(response)
    
    # 完成一次性记录
    logger.info("Speckit Marker matched: stage=%s, work_dir=%s, attempt=%s, round=%s, usage=%s",
        m.group("stage"), m.group("work_dir"), m.group("attempt"), 
        m.group("round"), usage)

    record = {
        "ts_event":      datetime.now().astimezone().isoformat(timespec="seconds"),
        "session_id":    session,
        "work_dir":      work_dir_rel,
        "stage":         m.group("stage"),
        "attempt":       int(m.group("attempt")),
        "round":         int(m.group("round")),
        "ts_marker":     m.group("ts"),
        "subagent_name": inputs.get("subagent_name"),
        "duration_sec":  duration_sec,
        "usage":         usage,
        "tool_call_brief": response.get("toolCallBrief", "")
    }
    out_path = os.path.join(work_dir_abs, "cost-events.jsonl")
    with open(out_path, "a", encoding="utf-8") as f:
        f.write(json.dumps(record, ensure_ascii=False) + "\n")

def handle_tapd_pipeline_agent(session, inputs, response) -> None:
    """
    处理tapd-pipeline-agent subagent事件，记录usage信息
    inputs: tool_input，结构参照19行
    response: tool_response，结构参照24行
    """
    # 解析迭代目录的绝对路径
    prompt = inputs.get("prompt", "")
    m = PIPELINE_RE.search(prompt)
    if not m:
        logger.info("Skipped: no [pipeline-cost-marker] found in prompt (len=%d)",
            len(prompt))
        return
    iter_dir_rel = m.group("iter_dir")
    iter_dir_abs = os.path.join(os.environ.get("CLAUDE_PROJECT_DIR", ""), iter_dir_rel)
    if not os.path.isdir(iter_dir_abs):
        logger.warning("Skipped: iter_dir does not exist: %s", iter_dir_abs)
        return

    usage, duration_sec = normalize_usage_and_duration(response)
    # 完成一次性记录
    logger.info("Pipeline Marker matched: action=%s, iter_dir=%s, ts=%s, story_id=%s, usage=%s",
        m.group("action"), m.group("iter_dir"), m.group("ts"), 
        m.group("story_id"), usage)

    # 构建记录（credit 保持在 usage 内部，不单独提取）
    record = {
        "ts_event":      datetime.now().astimezone().isoformat(timespec="seconds"),
        "session_id":    session,
        "iter_dir":      iter_dir_rel,
        "story_id":      m.group("story_id"),
        "action":        m.group("action"),
        "ts_marker":     m.group("ts"),
        "subagent_name": inputs.get("subagent_name"),
        "duration_sec":  duration_sec,
        "usage":         usage,
        "tool_call_brief": response.get("toolCallBrief", "")
    }

    # 写入迭代级 cost events 文件
    out_path = os.path.join(iter_dir_abs, "pipeline-cost-events.jsonl")
    with open(out_path, "a", encoding="utf-8") as f:
        f.write(json.dumps(record, ensure_ascii=False) + "\n")

def normalize_usage_and_duration(response) -> tuple[dict, float]:
    """
    提取usage和duration信息，返回标准化后的usage和duration_sec
    """
    raw_usage = response.get("usage", {})
    usage = {
        "input_tokens":  raw_usage.get("inputTokens", 0),
        "output_tokens": raw_usage.get("outputTokens", 0),
        "total_tokens":  raw_usage.get("totalTokens", 0),
        "cache_tokens":  raw_usage.get("cacheTokens", 0),
        "cached_write_tokens": raw_usage.get("cachedWriteTokens", 0),
        "cached_miss_tokens":  raw_usage.get("cachedMissTokens", 0),
        "credit":        raw_usage.get("credit", 0),
    }
    brief = response.get("toolCallBrief", "")
    dm = DURATION_RE.search(brief)
    duration_sec = float(dm.group(1)) if dm else 0.0
    return usage, duration_sec

def _ok():
    print(json.dumps({"continue": True}))
    return 0

if __name__ == "__main__":
    sys.exit(main())
