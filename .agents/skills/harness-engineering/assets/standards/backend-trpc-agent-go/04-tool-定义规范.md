## 四、Tool 定义规范


### 4.1 Tool 函数签名

```go
// ✅ 正确：参数结构体每个字段必须带 description tag
type SearchArgs struct {
    Query  string `json:"query"   description:"搜索关键词，不超过 100 个字符"`
    MaxNum int    `json:"max_num" description:"最大返回数量，取值 1-20，默认 10"`
}

type SearchResult struct {
    Items []string `json:"items"`
    Total int      `json:"total"`
}

// 函数签名固定：func(ctx context.Context, args T) (R, error)
func Search(ctx context.Context, args SearchArgs) (SearchResult, error) {
    // 实现搜索逻辑
    return SearchResult{}, nil
}
```

```go
// ❌ 错误：参数字段缺少 description tag，LLM 无法正确构造参数
type BadArgs struct {
    Query  string `json:"query"`    // 缺少 description tag
    MaxNum int    `json:"max_num"`  // 缺少 description tag
}
```

### 4.2 注册 Tool

```go
package tools

import "github.com/trpc-go/trpc-agent-go/tool"

func NewSearchTool() tool.Tool {
    return tool.NewFuncTool(
        "search",              // tool 名称，在同一 Agent 内唯一
        "在知识库中搜索相关内容",  // 描述越清晰，LLM 选择工具越准确
        Search,               // 函数引用
    )
}
```

### 4.3 Tool 规则

| 规则 | 说明 |
|------|------|
| description tag 必填 | 参数结构体每个字段必须有 `description` tag |
| 函数签名固定 | `func(ctx context.Context, args T) (R, error)` |
| 禁止 panic | 所有错误通过 `error` 返回 |
| 参数错误可透传 | 参数校验错误返回描述性 `fmt.Errorf`，LLM 可据此修正后重试 |
| 系统错误不透传 | DB/网络等异常的详细信息不透传给 LLM，返回通用错误消息 |
| tool 名称唯一 | 同一 Agent 内 tool 名称不可重复 |

---
