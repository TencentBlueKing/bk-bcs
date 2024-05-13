### 描述

获取 release 历史记录

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| projectCode         | string       | 是     | 项目英文名     |
| clusterID         | string       | 是     | 集群 ID     |
| namespace         | string       | 是     | 命名空间     |
| name         | string       | 是     | release 名称     |


### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"access_token": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/helmmanager/v1/projects/projecttest/clusters/BCS-K8S-00000/namespaces/default/releases/release-test/history
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "result": true,
    "data": [
        {
            "revision": 1,
            "name": "nginx",
            "namespace": "default",
            "updateTime": "2022-11-18 16:58:30",
            "description": "Install complete",
            "status": "deployed",
            "chart": "nginx",
            "chartVersion": "0.0.1",
            "appVersion": "0.0.1",
            "values": "enable: true\nenv: {}\nreplicas: 1\ntag: 3.8\n"
        }
    ],
    "requestID": "300c969e-ad23-49c1-a16b-56a47fd106cb"
}
```

### 响应参数
| 参数名称     | 参数类型     | 描述             |
| ------------ | ------------  | ---------------- |
| revision         | int      | release 原生版本     |
| status         | string      | release 状态，详情见下表     |

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