### 描述
查询是否有权限和申请链接，可以结合返回的web_annotations.perms字段交互

### 请求方法和URL
`POST {BK_BCS_BSCP_API}/api/v1/auth/iam/permission/check`

### 输入参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| biz_id         | uint32       | 是     | 业务ID     |
| basic         | []resource       | 是     |  基础资源信息     |
| gen_apply_url         | bool       | 是     | 是否生产申请链接     |

### 输入参数 basic resource
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| type         | string       | 是     | 类型，目前有 biz    |
| action         | string       | 是     | 权限操作，目前有  find_business_resource  |
| resource_id         | string       | 是     |  资源ID，和type对应，如业务，则为业务ID   |

### 调用示例
```json
{"biz_id": 2, "basic": {"type": "biz", "action": "find_business_resource", "resource_id": "2"}, "gen_apply_url": true}
```

### 响应示例
```json
{
    "error": {
        "code": "PERMISSION_DENIED",
        "message": "no permission",
        "data": {
            "apply_url": "http:/{host}/apply-custom-perm?system_id=bk-bscp&cache_id=60cb8e7a122a47dba402fa8f7367fd15",
            "resources": [
                {
                    "type": "biz",
                    "type_name": "业务",
                    "action": "find_business_resource",
                    "action_name": "业务访问",
                    "resource_id": "2",
                    "resource_name": "蓝鲸"
                }
            ]
        },
        "details": []
    }
}
```
