# Helm Chart 仓库与推送指南

**Helm** 是 Kubernetes 的官方包管理工具（类似于 Linux 中的 `yum` 或 `apt`），通过将应用所需的 Deployment、Service、ConfigMap、Ingress 等资源模板及参数打包为 **Helm Chart**，极大地简化了复杂 Kubernetes 应用的分发、部署与版本管理。

---

## 一、BCS Helm Chart 仓库类型说明

蓝鲸容器管理平台（BCS）依托蓝鲸制品库（BKRepo）为每个项目提供两类 Chart 仓库（进入【部署管理】→【Chart 仓库】即可查看）：

1. **项目专属仓库（私有仓库）**：
   - 仓库名称与**项目英文名（`projectCode`）**一致。
   - 仅当前项目有权限的成员可读写（推送或拉取 Chart），操作均需仓库账号与密码认证。
2. **公共仓库（`public-charts`）**：
   - 平台所有项目共享只读权限，内置社区常用开源组件与平台公共 Chart。
   - 若管理员需要将通用 Chart 发布到公共仓库供全平台使用，可在蓝鲸制品库（BKRepo）的 `blueking` 项目下管理 `public-charts` 仓库。

---

## 二、推送自定义 Helm Chart 到项目仓库

### 第一步：安装 `helm` 与 `helm-push` 命令行工具

若本地或构建机尚未安装 `helm`（推荐 Helm v3）及 `cm-push` 插件，请按以下步骤安装：

```bash
# 1. 安装 Helm 3 二进制工具
# 在线自动安装：
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# 或离线手工安装（从 https://github.com/helm/helm/releases 下载对应平台包）：
tar xf helm-v3.xx.xx-linux-amd64.tar.gz
cp -a linux-amd64/helm /usr/bin/

# 2. 安装 ChartMuseum helm-push 插件
# 在线安装：
helm plugin install https://github.com/chartmuseum/helm-push

# 或离线手工安装（从 https://github.com/chartmuseum/helm-push/releases 下载对应平台包）：
helm_push_plugin_dir=$(helm env HELM_PLUGINS)/cm-push
mkdir -p ${helm_push_plugin_dir}
tar xf helm-push_0.xx.xx_linux_amd64.tar.gz -C ${helm_push_plugin_dir}

# 3. 验证插件是否安装成功
helm plugin list
```

### 第二步：在 BCS 控制台获取项目仓库地址与凭据

1. 登录 BCS 控制台 →【部署管理】→【Chart 仓库】。
2. 点击页面右上角的【仓库信息 / 获取仓库配置】按钮，查看当前项目专属仓库的：
   - **仓库 URL**（例如 `http://helm.bkrepo.example.com/{project_code}/{project_code}/`）
   - **用户名（Username）** 与 **密码（Password / Token）**

### 第三步：添加项目仓库并推送 Chart

```bash
# 1. 添加 BCS 项目 Helm 仓库
helm repo add <project_code> <仓库地址> --username=<用户名> --password=<密码>

# 示例：
helm repo add demo-project http://helm.bkrepo.example.com/demo-project/demo-project/ \
  --username=admin-xxx --password=xxxx

# 检查仓库是否添加成功
helm repo list

# 2. 准备或打包您的 Helm Chart（例如目录名为 my-app/ 或打包后的 my-app-1.0.0.tgz）
# 查看 helm-push 插件版本
helm plugin list

# 3. 推送 Chart 到项目仓库
# 若 cm-push 插件版本 >= 0.10.0，使用 helm cm-push 命令：
helm cm-push my-app/ <project_code>

# 若 push 插件版本 < 0.10.0，使用 helm push 命令：
helm push my-app/ <project_code>
```

看到 `Pushing my-app-1.0.0.tgz to <project_code>... Done.` 即表示推送成功。随后回到 BCS【Chart 仓库】页面点击【同步仓库】即可看到新上传的 Chart 版本。

---

## 三、推送 Chart 常见问题（FAQ）

### Q1：添加仓库成功，但推送 Chart 时报错 `Error: 411` 且返回一段 HTML 页面？
- **错误信息示例**：
  ```text
  Error: 411: could not properly parse response JSON: <!DOCTYPE HTML ...>
  ```
- **原因**：本地终端开启了不支持分块传输或篡改 `Content-Length` 头的 HTTP 代理（`http_proxy` / `https_proxy`）。
- **解决**：在终端执行 `unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY` 关闭代理后重新推送，或将制品库域名加入 `no_proxy`。

### Q2：推送 Chart 时报错 `Error: 409: xxx-0.1.22.tgz already exists`？
- **原因**：为保障线上部署可追溯性与制品安全，BCS 仓库默认禁止覆盖已存在的同名同版本 Chart。
- **解决**：修改 Chart 目录下的 `Chart.yaml` 中的 `version` 字段（例如递增为 `0.1.23`）后重新推送；若确需覆盖，需先在【Chart 仓库】页面手动删除该旧版本。
