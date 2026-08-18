## 三、Proto 文件规范


### 3.1 文件结构

```protobuf
syntax = "proto3";

package {module};

option go_package = "github.com/{org}/{project}/stub/{module};{module}";

// {Module} 提供 {模块} 相关功能
service {Module} {
  // Create{Resource} 创建{资源}
  rpc Create{Resource}(Create{Resource}Request) returns (Create{Resource}Response);
}
```

### 3.2 命名约定

| 元素 | 规则 | 示例 |
|------|------|------|
| package | 小写，无下划线 | `user`, `order` |
| Service | PascalCase，无需 Service 后缀 | `User`, `Order` |
| RPC 方法 | PascalCase：`{Action}{Resource}` | `CreateUser`, `ListOrders` |
| Message | PascalCase：`{Action}{Resource}Request/Response` | `CreateUserRequest` |
| 字段 | snake_case | `user_id`, `created_at` |
| Enum | PascalCase，值为 UPPER_SNAKE_CASE | `Status`, `STATUS_ACTIVE` |

### 3.3 标准 CRUD RPC 模式

```protobuf
service {Module} {
  rpc Create{Resource}(Create{Resource}Request) returns (Create{Resource}Response);
  rpc Get{Resource}(Get{Resource}Request) returns (Get{Resource}Response);
  rpc List{Resource}s(List{Resource}sRequest) returns (List{Resource}sResponse);
  rpc Update{Resource}(Update{Resource}Request) returns (Update{Resource}Response);
  rpc Delete{Resource}(Delete{Resource}Request) returns (Delete{Resource}Response);
}
```

### 3.4 版本兼容性规则

| 规则 | 说明 |
|------|------|
| 禁止删除字段 | 标记 `reserved` 代替删除 |
| 禁止修改字段编号 | 编号一旦分配不可变更 |
| 新增字段使用 `optional` | 末尾追加，保持向后兼容 |
| Enum 新增值末尾追加 | 保留 0 值为默认值 |

---
