# CODEBUDDY.md

This file provides guidance to CodeBuddy Code when working with code in this repository.

## Overview

BCS Cluster Autoscaler is a Kubernetes component that automatically adjusts cluster size based on pod resource requirements and node utilization. It is built on top of the upstream `k8s.io/autoscaler/cluster-autoscaler` and customized for Blueking Container Service (BCS).

**Go Version:** 1.17  
**Kubernetes Version:** 1.16.15  
**License:** MIT (Tencent)

## Build Commands

```bash
# Build binary locally
make build

# Build binary only (skip tests)
make build-binary

# Run unit tests with race detection
make test-unit

# Format code
make format

# Clean build artifacts
make clean

# Build in Docker container (isolated environment)
make build-in-docker

# Run tests in Docker
make test-in-docker

# Build Docker image
make make-image

# Full release (build in Docker + create image + push)
REGISTRY=myregistry REPOSITORY=bcs TAG=v1.0 make release
```

### Build Variables

- `REGISTRY` - Docker registry URL
- `REPOSITORY` - Image repository name (default: `bcs`)
- `TAG` - Image tag (default: `dev`)
- `BUILD_TAGS` - Cloud provider selection (e.g., `clusterapi`)
- `LDFLAGS` - Build flags including encryption key for BCS provider

## Running Locally

```bash
./bcs-cluster-autoscaler \
  --cloud-provider=bcs \
  --cloud-config=/path/to/config \
  --kubeconfig=/path/to/kubeconfig \
  --scan-interval=10s \
  --v=4
```

## Architecture

### Directory Structure

```
├── main.go                    # Entry point, flag parsing, main loop
├── core/                      # Core autoscaling logic
│   ├── buffered_autoscaler.go # Main autoscaler engine (RunOnce method)
│   ├── scale_up.go            # Scale-up decision logic
│   ├── scale_down.go          # Scale-down decision logic
│   ├── webhook_*.go           # Webhook mode implementations
│   ├── cron.go                # Cron-scheduled scaling
│   └── nodegroup_set_manager.go # NodeGroupSet management
├── cloudprovider/
│   ├── bcs/                   # BCS cloud provider implementation
│   │   ├── cloud_provider.go  # Provider interface implementation
│   │   ├── node_group.go      # NodeGroup operations
│   │   ├── cache.go           # Node group caching
│   │   └── clustermanager/    # Cluster manager client (protobuf)
│   └── builder/               # Cloud provider factory
├── simulator/                 # Pod scheduling simulation
├── estimator/                 # Node count estimation
├── context/                   # Extended autoscaling context
├── scalingconfig/             # Configuration options
├── metrics/                   # Prometheus metrics
├── helm/                      # Helm chart for deployment
└── examples/                  # Configuration examples
```

### Core Flow

1. **main.go** initializes configuration, sets up Kubernetes client, and starts the main loop
2. **BufferedAutoscaler.RunOnce()** (`core/buffered_autoscaler.go`) executes each scan interval:
   - Refresh cloud provider state
   - Identify unschedulable pods
   - Attempt scale-up if needed (`scale_up.go`)
   - Evaluate and execute scale-down if needed (`scale_down.go`)
3. **CloudProvider** (`cloudprovider/bcs/`) interfaces with BCS cluster manager for node operations

### Key Components

**BufferedAutoscaler** (`core/buffered_autoscaler.go:52`)
- Main orchestrator implementing `core.Autoscaler` interface
- Manages cluster state, scale-up/down timing, and node info cache
- Supports webhook modes for custom scaling decisions

**BCS CloudProvider** (`cloudprovider/bcs/cloud_provider.go`)
- Implements `cloudprovider.CloudProvider` interface
- Uses `NodeGroupCache` for efficient node group lookups
- Communicates with BCS cluster manager via gRPC

**NodeGroupSetManager** (`core/nodegroup_set_manager.go`)
- Manages groups of node groups with shared policies
- Supports priority-based expansion and buffer ratios

### Extension Points

- **Expanders**: `least-waste`, `priority`, `price`, `random` - control node group selection strategy
- **Estimators**: `binpacking` (default), `clusterresource` - estimate required node count
- **Webhook Modes**: `Web`, `ConfigMap` - external scaling decision integration

## Key Configuration Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--scan-interval` | 10s | Cluster evaluation frequency |
| `--scale-down-enabled` | true | Enable automatic scale-down |
| `--scale-down-unneeded-time` | 10m | Time before node eligible for removal |
| `--scale-down-delay-after-add` | 20m | Delay scale-down after scale-up |
| `--cloud-provider` | bcs | Cloud provider type |
| `--namespace` | bcs-system | Namespace for CA resources |
| `--max-nodes-total` | 0 | Maximum cluster nodes (0=unlimited) |
| `--buffer-cpu-ratio` | - | CPU buffer ratio for scale-up decisions |
| `--buffer-mem-ratio` | - | Memory buffer ratio for scale-up decisions |
| `--webhook-mode` | - | Enable webhook-based scaling decisions |

## Testing

```bash
# Run all unit tests
make test-unit

# Run specific package tests
go test -race ./core
go test -race ./cloudprovider/bcs

# Run tests in Docker (isolated)
make test-in-docker
```

Key test files:
- `core/scale_up_test.go` - Scale-up logic tests
- `core/scale_down_test.go` - Scale-down logic tests
- `cloudprovider/bcs/cloud_provider_test.go` - Provider tests
- `simulator/cluster_test.go` - Simulation tests

## Metrics

Prometheus metrics exposed on `:8085/metrics`:

- `cluster_autoscaler_webhook_*` - Webhook mode metrics
- `cluster_autoscaler_scale_task` - Scale task status
- `cluster_autoscaler_resource_used_ratio` - Resource utilization
- `cluster_autoscaler_nodegroup_set_*` - NodeGroupSet metrics

## Deployment

**Helm:**
```bash
helm install bcs-ca ./helm/bcs-cluster-autoscaler \
  --namespace=bcs-system \
  --values=custom-values.yaml
```

**Manifest:**
```bash
kubectl apply -f mainifeasts/ca.yaml
```

## Dependencies

This project uses Go modules with extensive `replace` directives in `go.mod` to pin Kubernetes component versions to 1.16.15. The upstream autoscaler is replaced with a forked version:
```
k8s.io/autoscaler/cluster-autoscaler => github.com/OvertimeDog/cluster-autoscaler
```
