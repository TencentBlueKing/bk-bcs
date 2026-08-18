"""tests for log_usage.py — PostToolUse hook for tapd-story-pipeline cost tracking"""
import json
import os
import subprocess
import sys
import tempfile
import unittest


SCRIPT = os.path.join(os.path.dirname(__file__), "log_usage.py")


def run_hook(payload: dict) -> dict:
    """Invoke log_usage.py via subprocess, feeding payload as stdin JSON."""
    proc = subprocess.run(
        [sys.executable, SCRIPT],
        input=json.dumps(payload).encode("utf-8"),
        capture_output=True,
        timeout=10,
    )
    assert proc.returncode == 0, f"hook exited {proc.returncode}: {proc.stderr.decode()}"
    return json.loads(proc.stdout.decode().strip())


class TestHookHappyPath(unittest.TestCase):
    def test_marker_parsed_and_jsonl_appended(self):
        with tempfile.TemporaryDirectory() as cwd:
            work_dir_rel = "specs/v0.9.x/1234567"
            work_dir_abs = os.path.join(cwd, work_dir_rel)
            os.makedirs(work_dir_abs)

            prompt = (
                "你是隔离的 speckit 执行单元。\n"
                "[基础信息]\n"
                "- 需求 ID：1234567\n"
                f"- 工作目录 work_dir：{work_dir_rel}\n"
                "- 当前阶段 stage：specify\n"
                "- 当前 attempt：1\n"
                "- 当前 round：1\n"
                "- 调用时间：2026-05-22T16:30:00+08:00\n"
                "\n"
                f"[spec-cost-marker] work_dir={work_dir_rel} "
                "stage=specify attempt=1 round=1 ts=2026-05-22T16:30:00+08:00\n"
            )
            payload = {
                "session_id": "abcdef123",
                "cwd": cwd,
                "tool_name": "Task",
                "tool_input": {
                    "subagent_name": "speckit-execution-agent",
                    "prompt": prompt,
                },
                "tool_response": {
                    "type": "task_tool_result",
                    "toolCallBrief": "Execution Summary: 9 tool uses, cost: 33.95s",
                    "usage": {
                        "inputTokens": 68485,
                        "outputTokens": 3433,
                        "totalTokens": 71918,
                        "cacheTokens": 15360,
                        "cachedWriteTokens": 0,
                        "cachedMissTokens": 53125,
                        "credit": 1.2,
                    },
                },
            }

            result = run_hook(payload)
            self.assertEqual(result, {"continue": True})

            jsonl_path = os.path.join(work_dir_abs, "cost-events.jsonl")
            self.assertTrue(os.path.isfile(jsonl_path), "cost-events.jsonl not created")

            with open(jsonl_path) as f:
                lines = f.readlines()
            self.assertEqual(len(lines), 1)
            record = json.loads(lines[0])
            self.assertEqual(record["work_dir"], work_dir_rel)
            self.assertEqual(record["stage"], "specify")
            self.assertEqual(record["attempt"], 1)
            self.assertEqual(record["round"], 1)
            self.assertEqual(record["ts_marker"], "2026-05-22T16:30:00+08:00")
            self.assertEqual(record["session_id"], "abcdef123")
            self.assertEqual(record["subagent_name"], "speckit-execution-agent")
            self.assertEqual(record["tool_name"], "Task")
            self.assertAlmostEqual(record["duration_sec"], 33.95, places=2)
            self.assertEqual(record["usage"]["inputTokens"], 68485)
            self.assertEqual(record["usage"]["credit"], 1.2)
            self.assertIn("ts_event", record)


class TestHookEdgeCases(unittest.TestCase):
    def test_non_task_event_silently_skipped(self):
        result = run_hook({
            "tool_name": "Read",
            "cwd": "/tmp",
            "tool_input": {"prompt": "[spec-cost-marker] work_dir=x stage=y attempt=1 round=1 ts=z"},
        })
        self.assertEqual(result, {"continue": True})

    def test_no_marker_silently_skipped(self):
        result = run_hook({
            "tool_name": "Task",
            "cwd": "/tmp",
            "tool_input": {"prompt": "随便一段没有 marker 的 prompt"},
            "tool_response": {"usage": {"inputTokens": 1}},
        })
        self.assertEqual(result, {"continue": True})

    def test_work_dir_not_exist_silently_skipped(self):
        prompt = (
            "[spec-cost-marker] work_dir=does/not/exist "
            "stage=specify attempt=1 round=1 ts=2026-05-22T00:00:00+08:00"
        )
        result = run_hook({
            "tool_name": "Task",
            "cwd": "/tmp",
            "tool_input": {"prompt": prompt},
            "tool_response": {"usage": {}},
        })
        self.assertEqual(result, {"continue": True})

    def test_brief_without_duration_falls_back_to_zero(self):
        with tempfile.TemporaryDirectory() as cwd:
            work_dir_rel = "wd"
            os.makedirs(os.path.join(cwd, work_dir_rel))
            prompt = (
                f"[spec-cost-marker] work_dir={work_dir_rel} "
                "stage=plan attempt=2 round=3 ts=2026-05-22T16:40:00+08:00"
            )
            run_hook({
                "session_id": "s",
                "tool_name": "Task",
                "cwd": cwd,
                "tool_input": {"prompt": prompt},
                "tool_response": {
                    "toolCallBrief": "Execution Summary: 2 tool uses",
                    "usage": {"inputTokens": 100, "outputTokens": 50, "credit": 0.1},
                },
            })
            jsonl = os.path.join(cwd, work_dir_rel, "cost-events.jsonl")
            with open(jsonl) as f:
                record = json.loads(f.readline())
            self.assertEqual(record["duration_sec"], 0.0)
            self.assertEqual(record["stage"], "plan")
            self.assertEqual(record["attempt"], 2)
            self.assertEqual(record["round"], 3)

    def test_malformed_payload_silently_continues(self):
        proc = subprocess.run(
            [sys.executable, SCRIPT],
            input=b"not a json",
            capture_output=True,
            timeout=10,
        )
        self.assertEqual(proc.returncode, 0)
        self.assertEqual(json.loads(proc.stdout.decode().strip()), {"continue": True})


if __name__ == "__main__":
    unittest.main()
