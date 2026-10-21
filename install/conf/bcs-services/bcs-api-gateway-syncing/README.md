# BCS 网关资源同步镜像

该镜像打包 BCS 网关资源、文档及 MCP Server 配置，在部署时注册网关、发布资源版本，并同步 MCP Server。

## 环境变量依赖


| 变量名                        | 说明                            | 样例                                    |
| ----------------------------- | ------------------------------- | --------------------------------------- |
| BK_APIGW_NAME                 | 网关名称（也用于生成 MCP Server 名称） | bcs-api-gateway                    |
| BKAPI_DESCRIPTION             | 网关描述                        |                                         |
| BK_API_URL_TMPL               | 网关 API 地址模板               | http://bkapi.example.com/api/{api_name} |
| BK_APP_CODE                   | 应用                            | bcs                                     |
| BK_APP_SECRET                 | 应用密钥                        |                                         |
| BKAPI_RELEASE_VERSION         | 网关版本号(推荐使用 appVersion) | 1.0.0                                   |
| BKAPI_RELEASE_COMMENT         | 网关版本说明                    |                                         |
| BKAPI_STAGE_HOST              | 网关部署环境地址                | http://bcs-app.svc                      |
| BK_API_GRANT_PERMISSIONS_APPS | 资源和 MCP 主动授权应用，多个以逗号分隔 | bk_apigateway,bcs-agent             |
| BK_APP_TENANT_ID              | 【多租户】租户ID                 | system or ""                           |
| ENABLE_SYNC_MCP_SERVERS       | 是否同步 MCP，镜像默认 `true`；设为 `false` 仅跳过同步，不关闭已存在的 MCP | true |

## Helm chart 集成

声明一个 Job，在部署时通过镜像拉起容器，使用镜像默认命令 `sync-apigateway.sh`（或显式指定该命令）进行同步。
**运行成功后，会在 /data/apigateway.pub 中生成网关公钥，可以验证网关请求。**

## MCP Server

- `definition.yaml` 使用 `spec_version: 2`，在 `stages[].mcp_servers` 下声明 MCP。
- `prod` 环境创建两个服务，最终名称为 `{BK_APIGW_NAME}-prod-{name}`，例如网关名为 `bcs-api-gateway` 时为 `bcs-api-gateway-prod-mcp`；在网关管理端获取实际访问地址：
  - `mcp`（读写）：接入 `resources.yaml` 全部 **130 个工具**，名称与原服务一致。
  - `mcp-readonly`（只读）：接入 **70 个工具**，即全部 GET 资源，以及只查询、校验或预览、不修改数据的 POST 资源（监控 `*_by_post`、模板详情/变量/预览、Release/组件预览、节点驱逐检查等，清单见测试中的 `READ_ONLY_POSTS`）。新增 POST 资源时需判断是否加入该清单。
- 使用 `streamable_http` 协议，默认启用、**公开**（网关用户可在 MCP 市场看到并申请），分类为 `Official`；读写服务的写操作工具也会对外展示。
- 开启 OAuth2 公开客户端（`bk_app_code=public`）和个人客户端（个人 token，`bk_app_code=personal`）授权，用户可直接以个人身份调用，需 API 网关 >= 1.23.2。后端仍需校验调用方身份与项目/集群权限。
- 工具名与 `resources.yaml` 中的 `operationId` 一致，包括现有的 `list_template_vaiables` 拼写；新增或删除资源时需同步维护清单。
- 工具包含创建/更新项目及配额、创建/更新命名空间与变量、创建/导入/更新集群、节点上下架/驱逐/调度设置、节点池与弹性伸缩配置、任务重试/跳过、安装/升级/停止/卸载集群组件、部署模板、安装/升级/回滚/卸载 Release、删除 Chart、创建 WebConsole 会话等**写操作**，只在读写服务中提供。只需查询的智能体应接入只读服务；读写服务仅向可信应用授权，智能体接入时应按需选择工具并为高风险操作设置人工确认，不能仅依赖提示词限制。
- MCP 授权与 API 资源授权是两个独立步骤，配置为两者都授权给 `BK_APP_CODE` 和 `BK_API_GRANT_PERMISSIONS_APPS` 中的应用。原有资源鉴权配置不变，后端仍需校验调用方身份与项目/集群权限。
- AIDEV 等调用方需使用上述最终 MCP 名称，并单独完成平台侧所需授权。本目录不会修改 `bcs-bkai-syncing` 的 MCP 引用或智能体工具选择。

### 发布前检查

1. 目标 API 网关需支持 MCP Server 同步及 `streamable_http` 协议。
2. **更新 `BKAPI_RELEASE_VERSION`**，使新增的请求 Schema 发布到 `prod` 生效版本。不要使用 `--no-pub` 跳过发布后仍开启 MCP 同步。
3. 基础镜像固定为 `apigw-manager:4.2.4`，其默认脚本在 `create_version_and_release_apigw` 成功后执行 `sync_apigw_stage_mcp_servers`，无需自定义同步脚本。MCP 同步失败会使 Job 失败。
4. 同步后在网关管理端确认读写、只读服务的工具数量分别为 130、70，授权应用列表正确，再使用已授权调用方进行联调。同步配置本身不保证目标部署中的每个后端接口可用。

### 请求参数维护

- Project、Cluster、Cluster Resources 和 Helm 的参数依据仓库内对应服务的 Swagger/proto 补齐；路径参数名保留网关现有占位符，JSON body 不重复要求路径字段。
- 从网关 MCP 导出版本补充的项目、命名空间、变量、集群管理、集群组件及 Helm 接口，资源名统一为 snake_case，文档来自对应导出文档；公共路径参数引用 `components.parameters`，请求体定义放在 `components.schemas`。
- 监控工具声明 `X-Tenant-Project-Code`、`X-Operator` 和 Prometheus 查询参数；GET/POST 均通过 URL 查询参数传值（后端支持 `ParseForm`），不将 POST 查询误声明为 JSON body。`match[]` 使用可重复的数组参数。
- 只有确实无业务请求参数的 `projects_for_iam` 设置 `noneSchema: true`。不能用该字段绕过有参数接口的 Schema 配置。
- 操作记录过滤参数以 `SearchActivitiesForm` 为准（如 `resource_type`、`activity_type`）。WebConsole 请求体依据已有接口文档及 `OpenQuery` 定义，不把应用密钥或 token 暴露为模型输入参数。

## 本地构建与校验

在仓库根目录执行（校验不会访问线上网关）：

```bash
make api-gateway-syncing

docker build -t bcs-api-gateway-syncing:local \
  install/conf/bcs-services/bcs-api-gateway-syncing

python3 -m venv /tmp/bcs-apigw-check
/tmp/bcs-apigw-check/bin/pip install \
  -r install/conf/bcs-services/bcs-api-gateway-syncing/tests/requirements.txt
PYTHONDONTWRITEBYTECODE=1 /tmp/bcs-apigw-check/bin/python -m unittest discover \
  -s install/conf/bcs-services/bcs-api-gateway-syncing/tests -v
```

测试覆盖 OpenAPI 格式与引用、130 个资源及文档与 MCP 清单的一致性、路径参数、无参声明、监控参数、请求体及默认授权策略。

## 参考

- [apigw-manager](https://github.com/TencentBlueKing/bkpaas-python-sdk/tree/master/sdks/apigw-manager)
- [MCP Server 配置说明](https://github.com/TencentBlueKing/bkpaas-python-sdk/blob/master/sdks/apigw-manager/docs/sync_apigateway.md#14-mcp-server-配置说明)
