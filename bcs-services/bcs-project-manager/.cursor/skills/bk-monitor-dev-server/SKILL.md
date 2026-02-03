---
name: bk-monitor-dev-server
description: 帮助配置和启动本地开发服务器。当用户提到启动开发服务器、dev server、pnpm dev 时使用。
---

# 本地开发服务器配置

## 工作流程

### Step 0: 检查配置（优先）

```bash
node .cursor/skills/dev-server/scripts/setup-config.js
```

- `CONFIG_VALID` → 跳到 Step 5
- `CONFIG_INVALID:*` → 需配置，向用户索要 URL/Cookie/Token

### Step 1-4: 环境配置

1. 环境检查：`bash .cursor/skills/dev-server/scripts/check-env.sh`
2. 安装依赖：`pnpm i`
3. 生成配置（详见 config-guide.md）
4. 配置 hosts

### Step 5: 启动服务

- monitor-pc: `make dev-pc`
- trace (Vue3): `make dev-vue3`

### Step 6: 验证

等待 `Compiled successfully`，访问 `http://appdev.xxx.com:7001`

---

## 常见问题速查

| 问题          | 可能原因          | 解决方案                       |
| ------------- | ----------------- | ------------------------------ |
| 端口被占用    | 其他进程占用 7001 | 服务会自动换端口               |
| API 401/403   | Cookie/Token 过期 | 重新获取 Cookie 和 X-CSRFToken |
| API 连接失败  | 代理配置错误      | 检查 devProxyUrl               |
| 页面空白      | hosts 未配置      | 添加 hosts 映射                |
| 依赖安装失败  | 未使用 pnpm       | 使用 `pnpm i`                  |

详细排查见 [references/troubleshooting.md](references/troubleshooting.md)

---

## 📦 可用资源

- `skill://bk-monitor-dev-server/references/config-guide.md`
- `skill://bk-monitor-dev-server/references/troubleshooting.md`
- `skill://bk-monitor-dev-server/scripts/check-env.sh`
- `skill://bk-monitor-dev-server/scripts/setup-config.js`


---
## 📦 可用资源

- `skill://bk-monitor-dev-server/scripts/check-env.sh`
- `skill://bk-monitor-dev-server/scripts/generate-config.js`
- `skill://bk-monitor-dev-server/scripts/setup-config.js`
- `skill://bk-monitor-dev-server/references/config-guide.md`
- `skill://bk-monitor-dev-server/references/troubleshooting.md`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
