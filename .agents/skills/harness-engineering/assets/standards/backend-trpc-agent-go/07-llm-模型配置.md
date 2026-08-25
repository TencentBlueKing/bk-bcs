## 七、LLM 模型配置


```go
package model

import (
    "os"

    "github.com/trpc-go/trpc-agent-go/model"
    "github.com/trpc-go/trpc-agent-go/model/hunyuan"
)

func NewHunyuanModel() model.Model {
    return hunyuan.New(
        hunyuan.WithHTTPClientName("hunyuan-pro"),          // 绑定 yaml 中的 client 名称
        hunyuan.WithAPIKey(os.Getenv("HUNYUAN_API_KEY")),   // 从环境变量读取
        hunyuan.WithModel("hunyuan-pro"),
        hunyuan.WithMaxTokens(4096),                         // 生产环境必须设置上限
    )
}
```

### 模型配置规则

| 规则 | 说明 |
|------|------|
| API Key 走环境变量 | 通过 `os.Getenv` 读取，禁止硬编码 |
| MaxTokens 必须设置 | 防止单次请求超预算，推荐 2048~8192 |
| Client 名称绑定 | 通过 `WithHTTPClientName` 绑定 yaml 中的 client，统一管理路由和超时 |

---
