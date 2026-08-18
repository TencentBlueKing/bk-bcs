## 三、Proto 文件规范


### 3.1 文件结构

```protobuf
syntax = "proto3";

package {模块名};

option go_package = ".;{模块名}";

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";

// {Module}Service 提供 {模块} 相关功能
service {Module}Service {
  // Create{Resource} 创建{资源}
  rpc Create{Resource}(Create{Resource}Request) returns (Create{Resource}Response) {
    option (google.api.http) = {
      post: "/api/v1/{resources}"
      body: "*"
    };
  }
}
```

### 3.2 命名约定

| 元素 | 规则 | 示例 |
|------|------|------|
| package | 小写，无下划线 | `user`, `order` |
| Service | PascalCase + "Service" 后缀 | `UserService` |
| RPC 方法 | PascalCase：`{Action}{Resource}` | `CreateUser`, `ListOrders` |
| Message | PascalCase：`{Action}{Resource}Request/Response` | `CreateUserRequest` |
| 字段 | snake_case | `user_id`, `created_at` |
| Enum | PascalCase，值为 UPPER_SNAKE_CASE | `Status`, `STATUS_ACTIVE` |

### 3.3 标准 CRUD RPC 模式

```protobuf
service {Module}Service {
  rpc Create{Resource}(Create{Resource}Request) returns (Create{Resource}Response);
  rpc Get{Resource}(Get{Resource}Request) returns (Get{Resource}Response);
  rpc List{Resource}s(List{Resource}sRequest) returns (List{Resource}sResponse);
  rpc Update{Resource}(Update{Resource}Request) returns (Update{Resource}Response);
  rpc Delete{Resource}(Delete{Resource}Request) returns (Delete{Resource}Response);
}
```

### 3.4 HTTP 映射规则（grpc-gateway）

| 操作 | HTTP 方法 | 路径模式 | body |
|------|----------|---------|------|
| Create | POST | `/api/v1/{resources}` | `*` |
| Get | GET | `/api/v1/{resources}/{id}` | 无 |
| List | GET | `/api/v1/{resources}` | 无 |
| Update | PUT | `/api/v1/{resources}/{id}` | `*` |
| Delete | DELETE | `/api/v1/{resources}/{id}` | 无 |
| 自定义操作 | POST | `/api/v1/{resources}/{id}:{action}` | `*` |

### 3.5 版本兼容性规则

| 规则 | 说明 |
|------|------|
| 禁止删除字段 | 标记 `reserved` 代替删除 |
| 禁止修改字段编号 | 编号一旦分配不可变更 |
| 新字段用 `optional` | 保持向后兼容 |
| 新增字段末尾追加 | 不要插入已有字段之间 |
| Enum 新增值末尾追加 | 保留 0 值为默认值 |

---
