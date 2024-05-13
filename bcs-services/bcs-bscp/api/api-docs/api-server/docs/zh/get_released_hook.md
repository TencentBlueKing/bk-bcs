### 描述

该接口提供版本：v1.0.0+

获取服务配置脚本，未命名版本 release_id 为 0，pre_hook 或 post_hook 为 null 则表示未设置脚本。

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述   |
| -------- | -------- | ---- | ------ |
| biz_id   | uint32   | 是   | 业务ID |
| app_id   | uint32   | 是   | 应用ID |
| release_id | uint32   | 是   | 版本ID，未命名版本 id 为 0 |

### 调用示例

### 响应示例

```json
{
    "data": {
        "pre_hook": {
            "hook_id": 1,
            "hook_name": "alkaid-test-hook-4",
            "hook_revision_id": 1,
            "hook_revision_name": "v20230727155543",
            "type": "pre_hook",
            "content": "echo \"hello alkaid\""
        },
        "post_hook": {
            "hook_id": 2,
            "hook_name": "alkaid-test-hook-5",
            "hook_revision_id": 2,
            "hook_revision_name": "v20230727155603",
            "type": "post_hook",
            "content": "echo \"hello alkaid2\""
        }
    }
}
```