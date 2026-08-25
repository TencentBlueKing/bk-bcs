## 四、代码生成（trpc-cli）


### 4.1 工具安装

```bash
go install trpc.group/trpc-go/trpc-cli@latest
```

### 4.2 生成命令

**一键生成项目骨架**（首次初始化）：

```bash
trpc create --protofile={module}.proto --output=. --mod=github.com/{org}/{project}
```

**增量更新 stub**（修改 proto 后）：

```bash
trpc create --protofile={module}.proto --output=stub --rpconly
```

### 4.3 生成文件说明

| 文件 | 用途 | 修改规则 |
|------|------|---------|
| `stub/{pkg}/{module}.pb.go` | 消息类型定义 | **禁止手改** |
| `stub/{pkg}/{module}.trpc.go` | Server/Client 接口 | **禁止手改** |

---
