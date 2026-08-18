## 二、接口定义规范（Proto）


### 2.1 文件组织

```
api/proto/
└── {module}/
    └── {module}.proto    # 每个业务模块一个 proto 文件
```

### 2.2 命名约定

| 元素 | 规则 | 示例 |
|------|------|------|
| Package | 小写模块名 | `user`, `order` |
| Service | `{Module}Service` | `UserService` |
| RPC 方法 | `{Action}{Resource}` | `CreateUser`, `ListOrders` |
| Request | `{Action}{Resource}Request` | `CreateUserRequest` |
| Response | `{Action}{Resource}Response` | `CreateUserResponse` |
| 字段 | snake_case | `user_id`, `created_at` |

### 2.3 HTTP 映射标准

```protobuf
service UserService {
  // CreateUser 创建用户
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
    option (google.api.http) = {
      post: "/api/v1/users"
      body: "*"
    };
  }

  // GetUser 获取用户详情
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (google.api.http) = {
      get: "/api/v1/users/{user_id}"
    };
  }

  // ListUsers 获取用户列表
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse) {
    option (google.api.http) = {
      get: "/api/v1/users"
    };
  }

  // UpdateUser 更新用户
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {
    option (google.api.http) = {
      put: "/api/v1/users/{user_id}"
      body: "*"
    };
  }

  // DeleteUser 删除用户
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse) {
    option (google.api.http) = {
      delete: "/api/v1/users/{user_id}"
    };
  }
}
```

### 2.4 URL 设计规则

| 规则 | 说明 | 示例 |
|------|------|------|
| 资源用名词复数 | RESTful 风格 | `/api/v1/users` |
| 层级不超过 3 层 | 保持简洁 | `/api/v1/groups/{id}/members` |
| 路径参数用 `{field_name}` | 与 message 字段对应 | `/{user_id}` |
| 版本号在路径中 | 方便迁移 | `/api/v1/`, `/api/v2/` |
| 自定义操作用动词 | 非 CRUD 操作 | `/api/v1/users/{id}:activate` |

---
