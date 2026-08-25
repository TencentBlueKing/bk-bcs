## 九、Filter（拦截器）规范


Filter 分为 Server Filter（处理入站请求）和 Client Filter（处理出站请求），执行顺序与 yaml 中配置顺序一致。

### 9.1 自定义 Server Filter

```go
package filter

import (
    "context"
    "time"

    "trpc.group/trpc-go/trpc-go"
    "trpc.group/trpc-go/trpc-go/filter"
    "trpc.group/trpc-go/trpc-go/log"
)

func LogFilter(ctx context.Context, req interface{}, next filter.ServerHandleFunc) (interface{}, error) {
    start := time.Now()
    rsp, err := next(ctx, req)
    log.InfoContextf(ctx, "method=%s cost=%v err=%v",
        trpc.Message(ctx).ServerRPCName(), time.Since(start), err)
    return rsp, err
}

// 必须在 trpc.NewServer() 之前完成注册
func init() {
    filter.Register("log", LogFilter)
}
```

### 9.2 yaml 中启用 Filter

```yaml
server:
  filter:
    - recovery    # 先执行 panic 恢复
    - log         # 再执行请求日志

client:
  filter:
    - recovery
```

### 9.3 常用内置 Filter

| Filter | 引入包 | 功能 |
|--------|--------|------|
| recovery | `trpc.group/trpc-go/trpc-filter/recovery` | panic 恢复，避免单次请求崩溃整个服务 |
| debuglog | `trpc.group/trpc-go/trpc-filter/debuglog` | 自动打印请求/响应日志 |
| validation | `trpc.group/trpc-go/trpc-filter/validation` | proto validate 参数自动校验 |
| auth | `trpc.group/trpc-go/trpc-filter/auth` | 统一鉴权 |

### 9.4 使用规则

- 插件包必须同时 `import _ "插件包路径"` **和** 在 yaml 的 `filter` 数组中配置名称，缺一不可
- 鉴权逻辑统一通过 `trpc-filter/auth` 实现，禁止在每个 Handler 中重复编写
- `filter.Register` 必须在 `trpc.NewServer()` 之前调用

---
