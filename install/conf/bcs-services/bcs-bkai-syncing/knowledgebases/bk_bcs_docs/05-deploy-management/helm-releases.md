# Helm Release 应用部署、更新、回滚与排障

在将 Helm Chart 上传至仓库后，可通过蓝鲸容器管理平台（BCS）的【部署管理】模块将 Chart 实例化部署到指定集群与命名空间，生成并管理 **Helm Release**。

---

## 一、部署 Helm Chart（创建 Release）

### 1. 操作入口
- **方式 A**：进入【部署管理】→【Chart 仓库】，找到目标 Chart 点击【部署】。
- **方式 B**：进入【部署管理】→【Release 列表】，点击左上角【部署 Helm Chart】。

### 2. 填写部署配置项

| 配置项 | 填写要求与说明 |
| :--- | :--- |
| **Release 名称** | 只能包含小写字母、数字或连字符 `-`，在同一集群的同一命名空间内必须唯一。 |
| **Chart 版本** | 选择要部署的 Chart 目标版本号。 |
| **所属集群** | 选择部署的目标 Kubernetes 集群。 |
| **命名空间** | 选择目标集群下的命名空间（若不存在可先创建）。 |
| **Chart 部署选项（Values）** | 选择 Chart 内置的 values 配置文件（如 `values.yaml`），并可通过**表单模式**或 **YAML 编辑器**修改各项参数键值。 |
| **Helm 部署参数（Flags）** | 支持勾选常用的 `helm install/upgrade` 命令参数，例如 `--timeout`（超时时间，默认如 `300s`~`600s`）、`--wait`（等待所有 Pod 就绪后才标记成功）、`--atomic`（失败自动回滚）、`--skip-crds` 等。 |

### 3. 渲染预览（Preview）与确认部署
- 在真正执行部署前，强烈建议点击底部【预览】按钮：平台会调用 `helm template` 渲染出最终将要下发到集群的所有 Kubernetes YAML 资源清单，方便您检查镜像地址、环境变量、端口及 BCS 变量是否渲染正确。
- 预览无误后点击【部署】按钮，系统将跳转至【Release 列表】展示部署进度，直至状态变为 **正常（Deployed）** 或 **失败（Failed）**。

---

## 二、更新、回滚与卸载 Helm Release

在【部署管理】→【Release 列表】中，可按集群、命名空间或 Release 名称筛选已发布的实例，并执行全生命周期管理：

### 1. 更新（Upgrade）与版本差异对比
1. 点击目标 Release 右侧的【更新】按钮。
2. 可选择升级到新的 **Chart 版本**，或仅修改当前版本的 **`values.yaml` 参数**。
3. **版本对比（Diff）**：点击【预览 / 内容对比】，平台会高亮展示“当前线上已部署版本”与“本次待更新渲染结果”之间的 YAML 差异，确认变更范围符合预期后再点击【更新】。

### 2. 回滚（Rollback）
1. 当新版本更新失败或线上验证发现异常时，点击目标 Release 右侧的【回滚】按钮。
2. 在弹窗中选择历史正常运行的 **Revision（修订版本号）**，并可查看该历史版本与当前版本的配置差异。
3. 点击【确定】即可一键回滚至选定的历史 Revision。

### 3. 查看发布历史与资源状态
- **发布历史（History）**：点击【更新记录 / 历史】，查看每次 Revision 的变更时间、Chart 版本、状态及 Values 快照。
- **关联资源状态**：在 Release 详情页可直接查看该 Release 渲染创建的所有 K8S 资源（Deployment、StatefulSet、Service、Ingress、ConfigMap 等）及关联 Pod 的实时健康状态。

### 4. 卸载（Uninstall）
- 点击【删除 / 卸载】，平台将执行 `helm uninstall` 清理该 Release 在集群中创建的所有关联资源（注意：StatefulSet 动态创建的 PVC 默认不会被自动删除，以保护持久化数据安全）。
