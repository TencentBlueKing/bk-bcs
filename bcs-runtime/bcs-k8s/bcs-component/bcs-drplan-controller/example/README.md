# Example：Nginx 部署与双集群差异化

本目录提供一套完整示例：使用 **nginx Helm Chart** 渲染出 YAML，再通过 **DRPlan** 的 Subscription + Localization 下发到两个差异化子集群。

**参数化**：Workflow 中所有 action 的字符串字段均支持模板参数化，使用 `{{ .params.参数名 }}` 语法（如 `{{ .params.managedClusterNamespace }}`）。参数可来自 DRWorkflow 的 `spec.parameters` 默认值、DRPlan 的 `globalParams` 或各 stage 的 `workflows[].params`，执行时合并后参与渲染。支持的 action 类型及可参数化字段包括：Localization（name、namespace、feed 各字段、overrides 的 name/type/value）、Subscription（name、namespace、schedulingStrategy、feeds 各字段）、HTTP（url、method、body、headers）、Job（namespace）、KubernetesResource（manifest、operation）。

## 目录结构

```
example/
├── README.md
├── nginx/                    # nginx Helm Chart
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
│       ├── deployment.yaml
│       ├── configmap.yaml
│       └── service.yaml
└── plan/
    ├── install/              # 部署预案
    │   ├── drplan.yaml                           # 预案：先 Localization，再 Subscription
    │   ├── workflow-subscription.yaml            # 创建 Subscription（关联 YAML 资源）
    │   ├── workflow-localization-cluster-a.yaml  # 集群 A 差异化（副本数 1 + 响应头 X-Cluster: cluster-a）
    │   └── workflow-localization-cluster-b.yaml  # 集群 B 差异化（副本数 2 + 响应头 X-Cluster: cluster-b）
    └── switchover/           # 切换预案
        ├── drplan.yaml       # 预案：A 副本调 0，B 副本调 3
        └── workflow-switchover.yaml  # 通过更高优先级 Localization 覆盖（priority 800）
```

## 使用前提

- Subscription 关联的是 **父集群中已存在的 YAML 资源**（Deployment、ConfigMap、Service），**不是** HelmChart。
- 你需要先使用 `helm template` 等将 `nginx` chart 渲染为 YAML，并 apply 到父集群（或由你的流水线下发到 clusternet 父集群），再执行本 DRPlan。

## 使用步骤

1. **渲染 nginx 并下发到父集群**

   ```bash
   helm template nginx ./nginx -n default > nginx-rendered.yaml
   kubectl apply -f nginx-rendered.yaml -n default
   ```

   确保父集群中存在例如：`Deployment/nginx`、`ConfigMap/nginx-config`、`Service/nginx`（名称以 chart 的 `fullname` 为准）。

2. **按需修改 plan/install 中的资源名与命名空间**

   - `workflow-subscription.yaml` 中 `feeds` 的 `name`/`namespace` 与上面资源一致。
   - **Localization 的 namespace 已参数化**：使用参数 `managedClusterNamespace`，可在 DRPlan 的 stage 里按每个 workflow 传入（见 `drplan.yaml` 中 `localize` stage 的 `params`）。执行时传入的 params 会覆盖 workflow 的 default，无需改 workflow 文件即可适配不同集群。

3. **应用 DRWorkflow 与 DRPlan**

   ```bash
   kubectl apply -f plan/install/workflow-localization-cluster-a.yaml
   kubectl apply -f plan/install/workflow-localization-cluster-b.yaml
   kubectl apply -f plan/install/workflow-subscription.yaml
   kubectl apply -f plan/install/drplan.yaml
   ```

4. **执行预案**（通过 DRPlanExecution 或你的触发方式）

   执行后：先跑两个 Localization workflow（为两集群做差异化），再跑 Subscription workflow（将 nginx 的 YAML 资源下发到子集群）。

## 差异化说明

| 项目       | 集群 A | 集群 B |
|------------|--------|--------|
| Deployment | replicas: 2 | replicas: 5 |
| ConfigMap  | nginx 日志格式带 `(cluster-a)` | nginx 日志格式带 `[cluster-b]` |

Localization 的 `overrides` 使用 `MergePatch` 修改对应 Feed 资源；ConfigMap 的 `value` 需为合法 JSON（字符串内换行、引号需按 JSON 规则转义）。

## 切换预案（plan/switchover）

在完成安装预案后，可通过 **切换预案** 做流量切换：将 A 集群副本数调为 0、B 集群副本数调为 3。  
**不依赖 Patch**：新建 **更高优先级**（priority=800）的 Localization，与安装时创建的 priority=500 的 Localization 针对同一 Feed（Deployment/nginx），Clusternet 会按优先级应用，高优先级覆盖低优先级，从而生效。

1. 应用并执行切换预案前，需已执行过安装预案（或集群中已有对应 Subscription/Localization）。
2. 应用 workflow 与 drplan：
   ```bash
   kubectl apply -f plan/switchover/workflow-switchover.yaml
   kubectl apply -f plan/switchover/drplan.yaml
   ```
3. 通过 DRPlanExecution 执行 `nginx-switchover-plan`：
   ```bash
   kubectl apply -f plan/switchover/drplanexecution-execute.yaml
   ```
   执行顺序：先创建 A 的 scale-to-zero Localization，再创建 B 的 scale-to-three Localization。

## Execution 和 Revert

### 执行预案（Execute）

通过创建 DRPlanExecution CR 触发执行：

```bash
kubectl apply -f - <<EOF
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: nginx-install-plan-exec-001
  namespace: default
spec:
  planRef: nginx-install-plan
  operationType: Execute
EOF
```

### 回滚操作（Revert）

回滚会**删除**之前 Execute 创建的 Localization/Subscription 资源。

**⚠️ 重要：** `revertExecutionRef` 是**必填字段**，必须明确指定要回滚哪个 execution。

```bash
kubectl apply -f - <<EOF
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: nginx-switchover-plan-revert-001
  namespace: default
spec:
  planRef: nginx-switchover-plan
  operationType: Revert
  # revertExecutionRef 是必填字段
  revertExecutionRef: nginx-switchover-plan-exec-001
EOF
```

**优点：**
- ✅ 精确控制回滚目标，避免混淆
- ✅ 支持回滚历史任意版本
- ✅ 类型安全：webhook 会验证目标 execution 存在且状态正确
- ✅ GitOps 友好：声明式且可追踪

### Revert 验证

```bash
# 查看 execution 状态
kubectl get drplanexecution -l dr.bkbcs.tencent.com/plan=nginx-switchover-plan

# 查看 Localization 是否被删除
kubectl get localization -n clusternet-reserved

# 检查 DRPlan 状态
kubectl get drplan nginx-switchover-plan -o yaml

# 查看执行历史
kubectl get drplan nginx-switchover-plan -o jsonpath='{.status.executionHistory}' | jq
```
