# E2E 测试指南

本文档描述如何对 bcs-drplan-controller 进行端到端（E2E）测试。

---

## 📋 测试级别对比

| 测试类型 | 范围 | 依赖 | 运行时间 | 适用场景 |
|---------|------|------|---------|---------|
| **Unit Tests** | 单个函数/方法 | envtest (fake etcd) | ~6s | 开发阶段，快速验证逻辑 |
| **基础 E2E** | Controller 部署 | Kind + CertManager | ~5min | 验证 controller 基本可用性 |
| **完整 E2E** | DR 功能 | Kind + Clusternet + 示例应用 | ~15min | 发布前完整功能验证 |

---

## 🎯 测试策略

### 1. Unit Tests（单元测试）✅

**已实现**，使用 envtest 框架：

```bash
make test
```

**特点**:
- ✅ 不需要真实集群
- ✅ 快速（6秒）
- ❌ 不测试 Clusternet 集成
- ❌ 不测试真实的 Job/HTTP executor

---

### 2. 基础 E2E Tests（基础端到端）✅

**已实现**，测试 controller 部署和 metrics：

```bash
# 前提：需要 Kind 集群
kind create cluster --name drplan-test

# 运行基础 E2E
make test-e2e
```

**测试内容**:
- ✅ Controller pod 部署成功
- ✅ Metrics 服务可用
- ✅ Health/Readiness 探针正常
- ❌ **不测试 DR 功能**（缺少 Clusternet）

---

### 3. 完整 E2E Tests（完整端到端）⏳

**需要实现**，测试完整的 DR 功能。

## 🚀 完整 E2E 测试环境搭建

### 前提条件

```bash
# 安装工具
brew install kind kubectl helm

# 或者使用 Linux
curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-amd64
chmod +x ./kind && sudo mv ./kind /usr/local/bin/kind
```

---

### 方案 A: 单集群 + Clusternet（推荐，快速验证）

#### Step 1: 创建 Kind 集群

```bash
cat <<EOF | kind create cluster --name drplan-e2e --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF
```

#### Step 2: 安装 Clusternet（模拟多集群环境）

```bash
# 安装 Clusternet Hub（控制平面）
helm repo add clusternet https://clusternet.github.io/charts
helm repo update

helm install clusternet-hub clusternet/clusternet-hub \
  --namespace clusternet-system \
  --create-namespace \
  --set installCRDs=true

# 安装 Clusternet Agent（模拟子集群）
helm install clusternet-agent clusternet/clusternet-agent \
  --namespace clusternet-system \
  --set parentURL=https://clusternet-hub.clusternet-system.svc:443

# 等待 Clusternet 就绪
kubectl wait --for=condition=ready pod \
  -l app=clusternet-hub \
  -n clusternet-system \
  --timeout=5m
```

#### Step 3: 创建虚拟集群（模拟 cluster-a 和 cluster-b）

```bash
# 方式 1: 使用 Clusternet 的虚拟集群功能
kubectl apply -f - <<EOF
apiVersion: clusters.clusternet.io/v1beta1
kind: ManagedCluster
metadata:
  name: cluster-a
spec:
  syncMode: Pull
---
apiVersion: clusters.clusternet.io/v1beta1
kind: ManagedCluster
metadata:
  name: cluster-b
spec:
  syncMode: Pull
EOF

# 方式 2: 使用命名空间模拟（更简单）
kubectl create namespace cluster-a
kubectl create namespace cluster-b
```

#### Step 4: 部署 bcs-drplan-controller

```bash
# 构建镜像
make docker-build IMG=localhost:5001/bcs-drplan-controller:e2e

# 加载到 Kind
kind load docker-image localhost:5001/bcs-drplan-controller:e2e \
  --name drplan-e2e

# 部署 controller
make deploy IMG=localhost:5001/bcs-drplan-controller:e2e

# 验证部署
kubectl get pods -n bcs-drplan-controller-system
```

#### Step 5: 运行测试用例

```bash
cd example/plan/install

# 1. 创建 DRWorkflows
kubectl apply -f workflow-subscription.yaml
kubectl apply -f workflow-localization-cluster-a.yaml
kubectl apply -f workflow-localization-cluster-b.yaml

# 2. 创建 DRPlan
kubectl apply -f drplan.yaml

# 3. 等待 plan 就绪
kubectl wait --for=condition=Ready drplan/nginx-install-plan --timeout=1m

# 4. 执行 Execute 操作
kubectl apply -f - <<EOF
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: nginx-install-exec-001
  namespace: default
spec:
  planRef: nginx-install-plan
  operationType: Execute
EOF

# 5. 等待执行完成（最长 10 分钟）
kubectl wait --for=condition=Complete drplanexecution/nginx-install-exec-001 \
  --timeout=10m

# 6. 验证结果
echo "=== Execution Status ==="
kubectl get drplanexecution nginx-install-exec-001 -o jsonpath='{.status.phase}'
echo

echo "=== Subscription Created ==="
kubectl get subscription nginx-subscription -n default

echo "=== Localization Created ==="
kubectl get localization nginx-loc-cluster-a -n cluster-a
kubectl get localization nginx-loc-cluster-b -n cluster-b

# 7. 测试 Revert 操作
kubectl apply -f - <<EOF
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: nginx-install-revert-001
  namespace: default
spec:
  planRef: nginx-install-plan
  operationType: Revert
  revertExecutionRef: nginx-install-exec-001
EOF

# 8. 等待 Revert 完成
kubectl wait --for=condition=Complete drplanexecution/nginx-install-revert-001 \
  --timeout=10m

# 9. 验证资源已清理
kubectl get subscription nginx-subscription -n default 2>&1 | grep "NotFound" && echo "✅ Subscription deleted"
kubectl get localization nginx-loc-cluster-a -n cluster-a 2>&1 | grep "NotFound" && echo "✅ Localization A deleted"
```

---

### 方案 B: 多 Kind 集群（完整模拟，更真实）

#### Step 1: 创建 3 个 Kind 集群

```bash
# Hub 集群（Clusternet 控制平面）
kind create cluster --name hub

# Cluster A（成员集群 1）
kind create cluster --name cluster-a

# Cluster B（成员集群 2）
kind create cluster --name cluster-b
```

#### Step 2: 安装 Clusternet

```bash
# 在 Hub 集群安装 Clusternet Hub
kubectl config use-context kind-hub
helm install clusternet-hub clusternet/clusternet-hub \
  --namespace clusternet-system \
  --create-namespace \
  --set installCRDs=true

# 在 Cluster A 安装 Agent
kubectl config use-context kind-cluster-a
helm install clusternet-agent clusternet/clusternet-agent \
  --namespace clusternet-system \
  --set parentURL=https://$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' hub-control-plane):6443

# 在 Cluster B 安装 Agent
kubectl config use-context kind-cluster-b
helm install clusternet-agent clusternet/clusternet-agent \
  --namespace clusternet-system \
  --set parentURL=https://$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' hub-control-plane):6443
```

#### Step 3: 注册集群到 Clusternet

```bash
kubectl config use-context kind-hub

# 验证集群已注册
kubectl get managedclusters
# 应该看到 cluster-a 和 cluster-b
```

#### Step 4: 部署 bcs-drplan-controller 到 Hub

```bash
kubectl config use-context kind-hub

# 构建并加载镜像
make docker-build IMG=localhost:5001/bcs-drplan-controller:e2e
kind load docker-image localhost:5001/bcs-drplan-controller:e2e --name hub

# 部署
make deploy IMG=localhost:5001/bcs-drplan-controller:e2e
```

#### Step 5: 部署测试应用（Nginx）

```bash
# 在 Cluster A 和 B 分别部署 Nginx
for cluster in cluster-a cluster-b; do
  kubectl config use-context kind-$cluster
  kubectl create namespace nginx
done

# 切回 Hub 执行 DR Plan
kubectl config use-context kind-hub

# 创建 Subscription 指向两个集群
# ... (使用 example 中的 YAML)
```

---

## 🧪 自动化 E2E 测试脚本

创建 `test/e2e/dr_functionality_test.go`：

```go
//go:build e2e_dr
// +build e2e_dr

package e2e

import (
    "context"
    "time"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
    
    drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

var _ = Describe("DR Functionality", func() {
    var (
        ctx context.Context
        k8sClient client.Client
    )

    BeforeEach(func() {
        ctx = context.Background()
        // Initialize k8sClient (from suite setup)
    })

    Context("Execute and Revert", func() {
        It("should execute a plan successfully", func() {
            // 1. Create DRPlan
            plan := &drv1alpha1.DRPlan{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "test-plan",
                    Namespace: "default",
                },
                Spec: drv1alpha1.DRPlanSpec{
                    Stages: []drv1alpha1.Stage{
                        {
                            Name: "deploy",
                            Workflows: []drv1alpha1.WorkflowReference{
                                {
                                    WorkflowRef: drv1alpha1.ObjectReference{
                                        Name: "test-workflow",
                                    },
                                },
                            },
                        },
                    },
                },
            }
            Expect(k8sClient.Create(ctx, plan)).To(Succeed())

            // 2. Wait for plan to be Ready
            Eventually(func() string {
                _ = k8sClient.Get(ctx, client.ObjectKeyFromObject(plan), plan)
                return plan.Status.Phase
            }, 1*time.Minute).Should(Equal("Ready"))

            // 3. Create Execution
            execution := &drv1alpha1.DRPlanExecution{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "test-exec-001",
                    Namespace: "default",
                },
                Spec: drv1alpha1.DRPlanExecutionSpec{
                    PlanRef:       "test-plan",
                    OperationType: "Execute",
                },
            }
            Expect(k8sClient.Create(ctx, execution)).To(Succeed())

            // 4. Wait for execution to complete
            Eventually(func() string {
                _ = k8sClient.Get(ctx, client.ObjectKeyFromObject(execution), execution)
                return execution.Status.Phase
            }, 10*time.Minute).Should(Equal("Succeeded"))

            // 5. Verify resources created
            // ... (check Subscription, Localization, etc.)
        })

        It("should revert a plan successfully", func() {
            // ... (similar structure for Revert)
        })
    })
})
```

运行 DR 功能测试：

```bash
# 设置环境变量指向 Kind 集群
export KUBECONFIG=~/.kube/config

# 运行 DR E2E 测试
go test -v -tags=e2e_dr ./test/e2e/... -timeout=30m
```

---

## 📊 测试检查清单

### 基本功能

- [ ] DRWorkflow 创建和验证
- [ ] DRPlan 创建和验证
- [ ] DRPlanExecution (Execute) 创建
- [ ] 等待 Execute 完成
- [ ] 验证 Subscription 资源创建
- [ ] 验证 Localization 资源创建
- [ ] DRPlanExecution (Revert) 创建
- [ ] 等待 Revert 完成
- [ ] 验证资源已清理

### 高级功能

- [ ] 参数覆盖（globalParams, stage params）
- [ ] 失败处理（FailFast, Continue）
- [ ] 并发控制（同一 plan 不允许并发执行）
- [ ] Webhook 验证（删除保护、引用检查）
- [ ] ExecutionHistory 记录完整性
- [ ] Finalizer 保护（删除 execution 时更新 history）

### 性能测试

- [ ] 大规模 Plan（10+ stages, 50+ actions）
- [ ] 并发执行多个不同的 Plan
- [ ] 长时间运行（超过 1 小时）

---

## 🐛 调试技巧

### 查看 Controller 日志

```bash
kubectl logs -n bcs-drplan-controller-system \
  deployment/bcs-drplan-controller-controller-manager \
  -c manager \
  -f
```

### 查看 Execution 状态

```bash
# 查看详细状态
kubectl get drplanexecution test-exec-001 -o yaml

# 查看 stage/workflow/action 状态
kubectl get drplanexecution test-exec-001 \
  -o jsonpath='{.status.stageStatuses}' | jq

# 查看失败原因
kubectl get drplanexecution test-exec-001 \
  -o jsonpath='{.status.message}'
```

### 查看 Clusternet 资源

```bash
# 查看 Subscription
kubectl get subscription -A

# 查看 Localization
kubectl get localization -A

# 查看 Clusternet 集群状态
kubectl get managedclusters
```

---

## 🚨 常见问题

### Q1: Clusternet 安装失败

**症状**: `helm install clusternet-hub` 报错

**解决**:
```bash
# 检查 CRD 是否已安装
kubectl get crd | grep clusternet

# 手动安装 CRD
kubectl apply -f https://raw.githubusercontent.com/clusternet/clusternet/main/manifests/crds/
```

### Q2: Execution 一直 Pending

**症状**: `drplanexecution` phase 停留在 Pending

**排查**:
```bash
# 检查 plan 状态
kubectl get drplan <plan-name> -o jsonpath='{.status.phase}'

# 检查是否有其他 execution 正在运行
kubectl get drplan <plan-name> -o jsonpath='{.status.currentExecution}'

# 查看 controller 日志
kubectl logs -n bcs-drplan-controller-system deployment/... -c manager | grep ERROR
```

### Q3: Localization 未创建

**症状**: Execute 成功但 Localization 资源不存在

**排查**:
```bash
# 检查 Clusternet 是否安装
kubectl get crd localizations.apps.clusternet.io

# 检查 RBAC 权限
kubectl auth can-i create localizations --as=system:serviceaccount:bcs-drplan-controller-system:bcs-drplan-controller-controller-manager

# 查看 action 执行状态
kubectl get drplanexecution <exec-name> \
  -o jsonpath='{.status.stageStatuses[*].workflowExecutions[*].actionStatuses[?(@.name=="create-localization")]}'
```

---

## 📝 测试报告模板

```markdown
# E2E 测试报告

**测试日期**: 2026-02-03  
**环境**: Kind v0.20.0, Kubernetes v1.29.0, Clusternet v0.16.0  
**Controller 版本**: v1.0.0  

## 测试结果

| 测试用例 | 结果 | 耗时 | 备注 |
|---------|------|------|------|
| Basic Controller Deployment | ✅ PASS | 2m | - |
| Execute Simple Plan | ✅ PASS | 5m | 3 stages, 10 actions |
| Revert Simple Plan | ✅ PASS | 3m | All resources cleaned |
| Parameter Override | ✅ PASS | 5m | globalParams + stage params |
| Concurrent Execution Block | ✅ PASS | 1m | Webhook rejected |
| Large Plan (50 actions) | ⏳ SKIP | - | Need more time |

## 问题清单

1. **Minor**: Localization 创建延迟 ~10s（Clusternet API 响应慢）
2. **Fixed**: ExecutionHistory 只记录 2 条（已修复）

## 建议

- 增加集成测试覆盖率
- 添加性能基准测试
- 补充文档中的故障排查章节
```

---

## 🎓 总结

**简化版（快速验证）**:
```bash
# 1 个 Kind 集群 + Clusternet + 命名空间模拟
# 适用于：日常开发、快速验证
```

**完整版（发布前验证）**:
```bash
# 3 个 Kind 集群 + 完整 Clusternet 拓扑
# 适用于：发布前、重大变更
```

**生产环境（真实集群）**:
```bash
# 使用真实的 Kubernetes 集群 + Clusternet
# 适用于：生产环境验证、性能测试
```

---

## 📚 参考资料

- [Clusternet 文档](https://clusternet.io/)
- [Kind 文档](https://kind.sigs.k8s.io/)
- [Ginkgo 测试框架](https://onsi.github.io/ginkgo/)
- [Controller-Runtime Testing](https://book.kubebuilder.io/reference/testing)
