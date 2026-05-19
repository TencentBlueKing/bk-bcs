# BCS DR Plan Controller

![License](https://img.shields.io/badge/license-Apache--2.0-blue)
![Go Version](https://img.shields.io/badge/go-1.21+-blue)
![Kubernetes](https://img.shields.io/badge/kubernetes-1.19+-blue)

Kubernetes Operator for Disaster Recovery Plan orchestration and execution.

## 🎯 Features

- **Multi-Workflow Orchestration**: Organize multiple workflows into stages with dependency management
- **Parallel Execution**: Execute workflows in parallel within a stage
- **5 Action Types**: HTTP, Job, Localization (Clusternet), Subscription (Clusternet), KubernetesResource
- **Rollback Support**: Automatic and custom rollback mechanisms
- **Parameter Templates**: Template-based parameter substitution
- **Execution History**: Track and audit all executions
- **Event Logging**: Kubernetes Events and structured klog logging

## 📋 Architecture

```
┌─────────────┐         ┌─────────────┐         ┌──────────────────┐
│ DRWorkflow  │◄────────│   DRPlan    │────────►│ DRPlanExecution  │
│             │ 1:N     │             │ 1:N     │                  │
│ (Template)  │         │  (Instance) │         │   (Record)       │
└─────────────┘         └─────────────┘         └──────────────────┘
       │                                                │
       │ contains                                       │ contains
       ▼                                                ▼
┌─────────────┐                                 ┌──────────────────┐
│   Action    │                                 │  ActionStatus    │
│  (5 types)  │                                 │   (Status)       │
└─────────────┘                                 └──────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Kubernetes cluster 1.19+
- kubectl configured
- (Optional) cert-manager for webhook certificates

### Installation

#### Option 1: Using Helm (Recommended)

```bash
helm install bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  --namespace bcs-system \
  --create-namespace
```

#### Option 2: Using Kustomize

```bash
# Install CRDs
make install

# Deploy controller
make deploy IMG=your-registry/bcs-drplan-controller:v1.0.0
```

#### Option 3: Manual deployment

```bash
# Apply CRDs
kubectl apply -f config/crd/bases/

# Apply RBAC and deployment
kubectl apply -f config/rbac/
kubectl apply -f config/manager/
```

### Usage Example

1. **Create a DRWorkflow**:

```bash
kubectl apply -f config/samples/drworkflow-http.yaml
```

2. **Create a DRPlan**:

```bash
kubectl apply -f config/samples/drplan-simple.yaml
```

3. **Trigger execution**:

```bash
kubectl apply -f config/samples/drplanexecution.yaml
```

4. **Check status**:

```bash
# Check plan status
kubectl get drplan simple-dr-plan -o yaml

# Check execution status
kubectl get drplanexecution -l dr.bkbcs.tencent.com/plan=simple-dr-plan

# Check events
kubectl get events --field-selector involvedObject.name=simple-dr-plan-exec-001
```

5. **Revert execution**:

```bash
kubectl apply -f - <<EOF
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: simple-dr-plan-revert-001
spec:
  planRef: simple-dr-plan
  operationType: Revert
  revertExecutionRef: simple-dr-plan-exec-001  # Required!
EOF
```

6. **Cancel running execution**:

```bash
kubectl annotate drplanexecution <execution-name> dr.bkbcs.tencent.com/cancel=true
```

## 📚 Documentation

- [User Guide](docs/user-guide.md) - Complete user documentation
- [Specification](specs/001-drplan-action-executor/spec.md) - Feature specification
- [Data Model](specs/001-drplan-action-executor/data-model.md) - CRD data model
- [Quick Start](specs/001-drplan-action-executor/quickstart.md) - Quick start guide

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Kubebuilder 3.0+
- Docker (for building images)

### Build

```bash
# Build binary
make build

# Build with BCS version info
make bcs-build

# Run tests
make test

# Run linter
make bcs-lint
```

### Run Locally

```bash
# Install CRDs
make install

# Run controller locally
make run
```

### Build Docker Image

```bash
# Build image
make docker-build IMG=your-registry/bcs-drplan-controller:v1.0.0

# Push image
make docker-push IMG=your-registry/bcs-drplan-controller:v1.0.0
```

## 🏗️ Project Structure

```
bcs-drplan-controller/
├── api/v1alpha1/           # CRD type definitions
├── cmd/                    # Main entry point
├── config/                 # Kubernetes manifests
│   ├── crd/               # CRD definitions
│   ├── rbac/              # RBAC configuration
│   ├── manager/           # Controller deployment
│   └── samples/           # Example resources
├── docs/                   # Documentation
├── internal/
│   ├── controller/        # Reconcilers
│   ├── executor/          # Action executors
│   ├── webhook/           # Admission webhooks
│   └── utils/             # Utility functions
├── install/helm/          # Helm chart
└── specs/                 # Design specifications
```

## 🔍 Supported Action Types

| Action Type            | Description             | Rollback                |
| ---------------------- | ----------------------- | ----------------------- |
| **HTTP**               | HTTP/HTTPS requests     | Custom                  |
| **Job**                | Kubernetes Jobs         | Auto (delete Job)       |
| **Localization**       | Clusternet Localization | Auto (delete CR)        |
| **Subscription**       | Clusternet Subscription | Auto (delete CR)        |
| **KubernetesResource** | Generic K8s resources   | Auto (delete) or Custom |

## 📊 Monitoring

### Metrics

Controller exposes Prometheus metrics on `:8080/metrics`:

- `drplan_executions_total` - Total number of executions
- `drplan_execution_duration_seconds` - Execution duration
- `drplan_action_failures_total` - Total action failures

### Logging

Structured logging with klog:
- **Info level**: Key events (execution start/complete, stage transitions)
- **V(4) level**: Debug information (parameter substitution, action details, HTTP request/response)

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md).

## 📝 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙋 Support

- GitHub Issues: https://github.com/Tencent/bk-bcs/issues
- Documentation: https://github.com/Tencent/bk-bcs/tree/master/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/docs
