## 三、Agent 定义规范


### 3.1 推荐方式（llmagent.New）

```go
package agent

import (
    "github.com/trpc-go/trpc-agent-go/agent/llmagent"
    "github.com/trpc-go/trpc-agent-go/model"
    "{project}/internal/{module}/tools"
)

func NewMyAgent(m model.Model) *llmagent.LLMAgent {
    return llmagent.New(
        llmagent.WithName("ops-search-agent"),   // 名称全局唯一
        llmagent.WithModel(m),
        llmagent.WithSystemPrompt("你是一个运维助手，负责帮助用户查询系统状态..."),
        llmagent.WithMaxIterations(10),           // 防止无限循环，必须设置
        llmagent.WithTools(
            tools.NewSearchTool(),
            tools.NewGetMetricsTool(),
        ),
    )
}
```

### 3.2 自定义 Agent（业务编排复杂时）

```go
// 自定义 Agent 必须实现完整的 agent.Agent 接口
type MyAgent struct {
    model   model.Model
    myTools []tool.Tool
}

func (a *MyAgent) Run(ctx context.Context, req *agent.Request) (<-chan agent.Event, error) { ... }
func (a *MyAgent) Tools() []tool.Tool   { return a.myTools }
func (a *MyAgent) Info() agent.Info     { return agent.Info{Name: "my-agent"} }
func (a *MyAgent) SubAgents() []agent.Agent { return nil }
```

### 3.3 Agent 规则

| 规则 | 说明 |
|------|------|
| 名称全局唯一 | 命名格式：`{业务域}-{功能}-agent`，如 `ops-search-agent` |
| 优先 `llmagent.New` | 只有业务编排复杂时才自定义实现接口 |
| System Prompt 静态维护 | 禁止在运行时动态拼接不可控的外部内容 |
| 必须设置 MaxIterations | 防止因 Tool 调用失败陷入无限循环 |

---
