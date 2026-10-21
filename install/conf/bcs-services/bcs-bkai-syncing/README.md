# BCS AIDEV 初始化

在模块仓库维护 BCS 智能体配置，部署时用平台基础镜像中的 `bkai-init` 同步到 AIDEV。本目录不包含应用凭据，也不创建 MCP。

协议版本：bkai-init 0.1.0。

## 同步前

1. 在 BCS API 网关创建并发布 MCP Server，编码与 `agents/ai_bcs.yaml` 中的 `spec.mcps[].code` 一致（当前为 `bk-bcs-apigw-mcp`）。
2. 对接入应用和目标智能体完成授权。公开 MCP 也要授权。`bcs-api-gateway-syncing` 的资源授权不能代替这一步。
3. 核对 `selected_tools[].tool_name` 与已发布 MCP 的工具名。当前清单来自 `bcs-api-gateway/resources.yaml` 的只读 `operationId`。

## 目录

```text
bcs-bkai-syncing/
├── Dockerfile
├── bkai.yaml
├── agents/
│   └── ai_bcs.yaml
├── skills/
│   └── bcs_skill/
│       ├── skill.yaml
│       └── SKILL.md
└── knowledgebases/
    └── bk_bcs_docs/       # BCS 知识库配置与文档
        ├── knowledgebase.yaml
        └── README.md
```

容器内清单路径为 `/bk-bcs/bkai.yaml`。智能体 code 为 `ai-bcs`。

## 构建与校验

```bash
export BKAI_INIT_IMAGE='<可访问镜像仓库>/bk-aidev-init:0.1.0'
export BKAI_SPACE_ID='<目标空间ID>'

docker build --platform linux/amd64 \
  --build-arg BKAI_INIT_IMAGE="$BKAI_INIT_IMAGE" \
  -t bk-bcs-aidev:local install/conf/bcs-services/bcs-aidev

docker run --rm --platform linux/amd64 bk-bcs-aidev:local validate \
  -f /bk-bcs/bkai.yaml --space "$BKAI_SPACE_ID"
```

`validate` 只检查本地配置。连接信息在宿主 export 后按变量名传入容器：

- `BK_API_URL_TMPL`：须包含 `{api_name}/{stage}`
- `BKAI_APP_CODE`、`BKAI_APP_SECRET`

先 `plan` 或 `diff`，确认后再写入。首次需要发布时追加 `--publish --publish_config_only=0`。未加 `--publish` 时只同步草稿。

```bash
docker run --rm --platform linux/amd64 \
  --env BK_API_URL_TMPL --env BKAI_APP_CODE --env BKAI_APP_SECRET \
  bk-bcs-aidev:local sync -f /bk-bcs/bkai.yaml \
  --tenant-id system --space "$BKAI_SPACE_ID" \
  --confirm --publish --publish_config_only=0
```
