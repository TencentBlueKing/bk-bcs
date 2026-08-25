## 五、Agentic Loop（调用链路）


```
Handler
  └─ Runner.Run(ctx, req)
       └─ Agent.Run(ctx, req)
            ├─ LLM 推理 → 生成 ToolCall 指令
            │     └─ Tool.Execute() → 结果注入下一轮 LLM
            └─ LLM 推理 → 生成 FinalAnswer → 事件流结束
```

### 5.1 事件消费规范

```go
import (
    "github.com/trpc-go/trpc-agent-go/runner"
    "github.com/trpc-go/trpc-agent-go/runner/event"
)

r := runner.New(myAgent)
eventCh, err := r.Run(ctx, &runner.Request{Input: userInput})
if err != nil {
    return nil, errs.New(ErrCodeAgentFailed, "agent run failed")
}

var finalAnswer string
for e := range eventCh {
    switch {
    case event.IsRunnerCompletion(e):          // ✅ 以此作为结束信号
        finalAnswer = e.Content
    case event.IsToolCall(e):
        log.InfoContextf(ctx, "tool call: name=%s", e.ToolName)
    case event.IsToolResult(e):
        log.InfoContextf(ctx, "tool result: name=%s", e.ToolName)
    }
}

// ❌ 错误：不能用 Response.Done 字段判断结束
```

### 5.2 Agentic Loop 规则

- 必须以 `event.IsRunnerCompletion(e)` 作为循环终止条件
- 每次 Tool 调用和结果都应记录日志，包含 tool 名称
- `WithMaxIterations` 为必填项，推荐设为 10~20，防止无限循环
- 循环结束后检查 `finalAnswer` 是否为空，为空表示 Agent 未正常完成

---
