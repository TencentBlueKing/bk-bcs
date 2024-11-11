## 描述

集群资源管理接口提供了对集群中各种类型资源（如 Workload、Config、Network、Storage
等）的管理功能，支持对这些资源进行创建、读取、更新和删除（CRUD）操作，以及扩缩容、重启等特定操作。

## 身份认证

cluster-resources 的接口不能使用 admin 的 bearer token，调用权限中心接口鉴权时无法获取用户 ID，导致出现错误：

```bash
curl -X GET 'http://bcs-api.bkdomain/bcsapi/v4/clusterresources/v1/projects/{projectID}/clusters/{clusterID}/workloads/deployments' \
-H 'Authorization: Bearer xxx'

{"code":2,"message":"ctx validate failed: Username/ProjectID/ClusterID required","requestID":"xxx","data":null,"webAnnotations":null}
```

使用蓝鲸的 bk_token 可以正常访问接口，在请求 Header 中添加 `Cookie: bk_token={bk_token}` 和 `User-Agent: Mozillaxxxxxx` 即可：

```bash
curl -X GET \
-H 'Cookie: bk_token=xxx' \
-H 'User-Agent: xxx' \
'http://bcs-api.bkdomain/bcsapi/v4/clusterresources/v1/projects/{projectID}/clusters/{clusterID}/namespaces/{namespace}/workloads/deployments/{name}'
```

## 接口路径规则

接口路径通常遵循以下通用规则：

```
/clusterresources/v1/projects/{projectID}/clusters/{clusterID}/[namespace/{namespace}/]resource_type/{resourceName}[/{action}]
```

- **resource\_type**：指定资源的类型，例如 `workloads`、`configs`、`network`，表示具体的资源类别。

- **resourceName**：指定资源的名称，用于定位特定的资源实例。

- **action**：指定对资源的特定操作，例如扩缩容（`scale`）、重启（`restart`）、重新调度（`reschedule`
  ）、回滚到指定版本（`rollout/{revision}`）等。此参数是可选的，只有在需要对资源执行特定操作时才使用。

| 操作类型   | HTTP 方法 | 路径示例                                       | 描述       |
|--------|---------|--------------------------------------------|----------|
| 创建资源   | POST    | `/resource_type`                           | 创建新资源    |
| 查询资源列表 | GET     | `/resource_type`                           | 查询资源列表   |
| 查询单个资源 | GET     | `/resource_type/{name}`                    | 查询单个资源   |
| 更新资源   | PUT     | `/resource_type/{name}`                    | 更新资源信息   |
| 删除资源   | DELETE  | `/resource_type/{name}`                    | 删除资源     |
| 扩缩容操作  | PUT     | `/resource_type/{name}/scale`              | 调整资源规模   |
| 重启操作   | PUT     | `/resource_type/{name}/restart`            | 重启资源     |
| 重新调度   | PUT     | `/resource_type/{name}/reschedule`         | 重新调度资源   |
| 回滚操作   | PUT     | `/resource_type/{name}/rollout/{revision}` | 回滚到指定版本  |
| 查询历史版本 | GET     | `/resource_type/{name}/history`            | 查询资源历史版本 |

### 通用的参数

大部分资源类型相关的接口都需要传入路径参数：

`/clusterresources/v1/projects/{projectID}/clusters/{clusterID}/namespaces/{namespace}/resource_type`

> **Tips**：对于集群级别的资源（如 CRD），不需要提供 `namespace` 参数，因为这些资源不属于特定的命名空间。
> 即接口路径为 `/clusterresources/v1/projects/{projectID}/clusters/{clusterID}/resource_type`

| 名称        | 类型     | 必选 | 说明    |
|-----------|--------|----|-------|
| projectID | string | 是  | 项目 ID |
| clusterID | string | 是  | 集群 ID |
| namespace | string | 是  | 命名空间  |
| name      | string | 否  | 资源名称  |

一些 PUT 方法的接口还需要传入 Request Body，通用结构如下：

```json
{
  "projectID": "string",
  "clusterID": "string",
  "namespace": "string",
  "name": "string",
  "rawData": {},
  "format": "string"
}
```

| 名称        | 类型     | 说明                        |
|-----------|--------|---------------------------|
| projectID | string | 项目 ID                     |
| clusterID | string | 集群 ID                     |
| namespace | string | 命名空间                      |
| name      | string | 资源名称                      |
| rawData   | object | 资源配置信息                    |
| format    | string | 资源配置格式（manifest/formData） |

### 基础 CRUD 操作

以下以 Pod 资源为例，展示基础的 CRUD 操作：

- **查找 Pod 列表**：\
  `GET /clusterresources/v1/projects/{projectID}/clusters/{clusterID}/namespaces/{namespace}/workloads/pods`
- **查找单个 Pod**：\
  `GET /clusterresources/v1/projects/{projectID}/clusters/{clusterID}/namespaces/{namespace}/workloads/pods/{pod_name}`
- **删除 Pod**：\
  `DELETE /clusterresources/v1/projects/{projectID}/clusters/{clusterID}/namespaces/{namespace}/workloads/pods/{pod_name}`
- **更新 Pod**：\
  `PUT /clusterresources/v1/projects/{projectID}/clusters/{clusterID}/namespaces/{namespace}/workloads/pods/{pod_name}`

更新 Pod 请求通常需要包含 Request Body，具体结构参考通用参数的格式。

### 其他 `action` 的操作

以下是一些具有特定 `action` 的资源操作：

- **扩缩容 Pod**：
  `PUT /clusterresources/v1/projects/{projectID}/clusters/{clusterID}/namespaces/{namespace}/workloads/pods/{pod_name}/scale`\
  通过此接口对指定的 Pod 进行扩缩容操作。

- **重启 Pod**：
  `PUT /clusterresources/v1/projects/{projectID}/clusters/{clusterID}/namespaces/{namespace}/workloads/pods/{pod_name}/restart`\
  重启指定的 Pod。

- **重新调度 Pod**：
  `PUT /clusterresources/v1/projects/{projectID}/clusters/{clusterID}/namespaces/{namespace}/workloads/pods/{pod_name}/reschedule`\
  重新调度指定的 Pod。

- **回滚 Pod 到指定版本**：
  `PUT /clusterresources/v1/projects/{projectID}/clusters/{clusterID}/namespaces/{namespace}/workloads/pods/{pod_name}/rollout/{revision}`\
  回滚指定 Pod 到某个历史版本。

更多详细的接口信息可以参考：[cluster-resources.swagger.json](https://github.com/TencentBlueKing/bk-bcs/blob/master/bcs-services/cluster-resources/swagger/data/cluster-resources.swagger.json)
。具体的接口使用方法可以参考网关下 deployment 相关接口的文档中的示例

## 可管理的集群资源

集群资源管理接口支持以下资源的管理：

- **Workload**
    - Deployment
    - StatefulSet
    - DaemonSet
    - Job
    - CronJob
    - Pod
- **Config**
    - Secret
    - ConfigMap
- **Network**
    - Ingress
    - Service
    - Endpoint
- **Storage**
    - PersistentVolumeClaims
    - PersistentVolumes
    - StorageClasses
- **HPA**
- **CRD**
- **RBAC**
