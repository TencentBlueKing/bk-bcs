### 描述

获取 release 列表

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| projectCode         | string       | 是     | 项目英文名     |
| clusterID         | string       | 是     | 集群 ID     |

### query 参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| namespace         | string       | 否     | 命名空间     |
| name         | string       | 否     | release 名称，支持模糊搜索     |
| page         | int       | 否     | 页     |
| size         | int       | 否     | 每页数量     |


### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"access_token": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/helmmanager/v1/projects/projecttest/clusters/BCS-K8S-00000/releases
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "result": true,
    "data": {
        "page": 0,
        "size": 0,
        "total": 1,
        "data": [
            {
                "name": "name",
                "namespace": "namespace",
                "revision": 1,
                "status": "deployed",
                "chart": "nginx",
                "appVersion": "0.0.1",
                "updateTime": "2022-11-18 16:17:45",
                "chartVersion": "0.0.1",
                "createBy": "admin",
                "updateBy": "admin",
                "message": "Install complete",
                "repo": "projecttest",
                "projectCode": "projecttest",
                "clusterID": "BCS-K8S-00000"
            }
        ]
    },
    "requestID": "92127d0f-b3ab-4e87-85b0-60166ef1412f"
}
```

### 响应参数
| 参数名称     | 参数类型     | 描述             |
| ------------ | ------------  | ---------------- |
| revision         | int      | release 原生版本     |
| status         | string      | release 状态，详情见下表     |
| message         | string       | release 部署信息，如果报错将会展示 release 报错信息     |

### release 状态
| status     | 状态     | 描述             |
| ------------ | ------------  | ---------------- |
| deployed         | 正常/部署成功      |      |
| uninstalled         | 已卸载      |      |
| superseded         | 废弃      |      |
| failed         | 失败      |      |
| uninstalling         | 卸载中      |      |
| pending-install         | 安装中      |      |
| pending-upgrade         | 升级中      |      |
| pending-rollback         | 回滚中      |      |
| failed-install         | 安装失败      |      |
| failed-upgrade         | 升级失败      |      |
| failed-rollback         | 回滚失败      |      |
| failed-uninstall         | 卸载失败      |      |
| unknown         | 未知      |   一般是通过 helm 命令手动安装的错误   |