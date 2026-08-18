## 八、会话管理（Session）


### 8.1 Session 存储

```go
package session

import (
    "time"

    "github.com/trpc-go/trpc-agent-go/session"
    "github.com/redis/go-redis/v9"
)

func NewRedisSessionStore(rdb *redis.Client) session.Store {
    return session.NewRedisStore(rdb,
        session.WithTTL(30*time.Minute),  // 根据业务场景设置，默认 30 分钟
    )
}
```

### 8.2 多轮对话

```go
// Handler 中实现多轮对话
func (h *ChatImpl) Chat(ctx context.Context, req *pb.ChatRequest) (*pb.ChatResponse, error) {
    // 从请求获取 session_id，若为空则创建新会话
    sessionID := req.GetSessionId()
    if sessionID == "" {
        sessionID = uuid.New().String()
    }

    agent := llmagent.New(
        llmagent.WithSessionStore(h.sessionStore),
        llmagent.WithMaxHistoryMessages(20),   // 控制上下文窗口，防止 token 超限
    )

    r := runner.New(agent)
    eventCh, err := r.Run(
        context.WithValue(ctx, session.KeySessionID, sessionID),
        &runner.Request{Input: req.GetInput()},
    )
    // ... 消费事件流
}
```

### 8.3 Session 规则

- 生产环境禁止使用内存 Session（无法水平扩展，重启丢失历史）
- `MaxHistoryMessages` 建议设为 10~20 条，防止 token 超限
- Session TTL 根据业务场景配置，建议 15~60 分钟

---
