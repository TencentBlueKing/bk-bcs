## 描述

bcs-storage 提供了对集群存储资源（如告警数据、集群配置、自定义资源等）的管理功能，支持对这些资源进行创建、读取、更新和删除（CRUD）操作，以及其他操作。

## 访问路径获取

bcs-storage 中的路径会通过网关 rewrite，并非路由前缀和各组件接口的简单叠加，具体改写规则可以查看 bcs 网关服务中的 apisix
配置文件，可以参考以下示例：

```shell
# 拷贝 apisix 配置文件到容器外
kubectl -n bcs-system cp bcs-services-stack-bk-micro-gateway-xxxxx:/usr/local/apisix/conf/apisix.yaml ./apisix.yaml -c apisix
```

下面是一个文件示例，包含了与 storage 相关的部分：

```yaml
# storage 服务的路由配置
- id: bcs-api-gateway.prod.storage
  name: storage
  labels:
    gateway.bk.tencent.com/gateway: bcs-api-gateway
    gateway.bk.tencent.com/stage: prod
  priority: -982
  uris:
    - /bcsapi/v4/storage    # 匹配 /bcsapi/v4/storage 路径的请求
    - /bcsapi/v4/storage/*bk_api_subpath_match_param_name # 通配符匹配以 /bcsapi/v4/storage/ 开头的路径
  enable_websocket: true
  plugins:
    # 认证、日志、监控等插件配置
    bkbcs-auth: { ... }        # 负责认证
    file-logger: { ... }       # 记录访问日志
    prometheus: { ... }        # 用于监控

    # 路径改写插件，用于将原始路径重写为实际的服务路径
    proxy-rewrite:
      regex_uri:
        - /bcsapi/v4/storage/(.*)   # 匹配原始路径的正则表达式
        - /bcsstorage/v1/$1         # 重写后的路径

    request-id:
      algorithm: uuid
      header_name: X-Request-Id
      include_in_response: true

  status: 1
  service_id: bcs-api-gateway.prod.storage
```

`proxy-rewrite` 插件配置用于将路径 `/bcsapi/v4/storage/...` 重写为 `/bcsstorage/v1/...`，通过正则匹配实现。

所以在外部访问的时候应该使用 `/bcsapi/v4/storage/...` 而不是 swagger 文档中所使用的，参考时需要注意使用正确路径替换。

## 可以访问获取的资源

可以参考 swagger
文档：[bcs-storage.swagger.json](https://github.com/TencentBlueKing/bk-bcs/blob/master/bcs-services/bcs-storage/pkg/proto/bcs-storage.swagger.json)

> 文档中 v2 的接口，但推荐使用 v1，只需要将 v2 相关路径根据上述的访问路径获取方法进行替换即可，入参和结果等均可参考 v2 的接口

详情访问示例可以参考其他 storage 相关的接口文档中的示例
