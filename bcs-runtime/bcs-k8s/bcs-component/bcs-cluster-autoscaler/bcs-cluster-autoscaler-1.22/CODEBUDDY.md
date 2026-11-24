# CODEBUDDY.md This file provides guidance to CodeBuddy Code when working with code in this repository.

## Overview

This is **bcs-cluster-autoscaler-1.22**, a customized fork of the Kubernetes Cluster Autoscaler for the Blueking Container Service (BCS). It is based on Kubernetes 1.24/Cluster Autoscaler 1.22 and includes Tencent-specific extensions for integration with the BCS cluster-manager service.

The autoscaler automatically adjusts cluster size by:
- Adding nodes when pods fail to schedule due to insufficient resources
- Removing underutilized nodes when pods can be placed elsewhere

## Common Development Commands

### Building

```bash
# Standard build (local environment)
make build

# Build binary only
make build-binary

# Build with custom tags
BUILD_TAGS=mytag make build

# Build inside Docker container (reproducible)
make build-in-docker

# Build with custom LDFLAGS (e.g., for encryption keys)
bcs_encryption_key=mykey make build
```

### Testing

```bash
# Run unit tests
make test-unit

# Run tests in Docker
make test-in-docker

# Run specific test
go test ./cloudprovider/bcs -run TestNodeGroup

# Run with race detection
GO111MODULE=on go test -race ./...
```

### Docker Images

```bash
# Build Docker image
TAG=v1.0.0 REGISTRY=myregistry.com REPOSITORY=bcs make make-image

# Build and push image
TAG=v1.0.0 REGISTRY=myregistry.com REPOSITORY=bcs make dev-release

# Full release (build in Docker + push)
TAG=v1.0.0 REGISTRY=myregistry.com REPOSITORY=bcs make release

# Build e2e test image
make build-e2e
make container-e2e
```

### Code Quality

```bash
# Format code
make format

# Generate mocks/code
make generate
```

### Deployment

```bash
# Helm chart available in helm/bcs-cluster-autoscaler/
helm install bcs-cluster-autoscaler ./helm/bcs-cluster-autoscaler
```

## Architecture Overview

### Core Components

1. **cloudprovider/bcs/** - Custom BCS cloud provider implementation
   - `cloud_provider.go` - Main provider interface implementing standard Cluster Autoscaler CloudProvider
   - `node_group.go` - BCS-specific node group management
   - `cache.go` - Node group caching layer for performance
   - `clustermanager/` - Integration with BCS cluster-manager service via HTTP API

2. **core/** - Enhanced autoscaling logic (extends upstream Cluster Autoscaler)
   - `buffered_autoscaler.go` - Main autoscaler with buffered resource ratio support
   - `scale_up.go` - Scale-up decision logic with separated pod/buffer paths
   - `scale_down.go` - Scale-down decision logic with custom timing controls and buffer checks
   - `nodegroup_set_manager.go` - NodeGroupSet management, resource snapshots, buffer checking
   - `multi_set_expansion.go` - Multi-Set concurrent expansion orchestration
   - `set_expander.go` - Set-aware Expander wrapper for per-Set expander strategies
   - `webhook_*.go` - Webhook integration for custom scaling policies and external event handling
   - `cron.go` - Cron-based scheduled scaling operations

3. **estimator/** - Resource estimation with buffer support
   - `resource_estimator.go` - ClusterResourceEstimator with per-Set buffer ratio calculation

3. **scalingconfig/** - Custom configuration options
   - `config.go` - Extended options beyond standard Cluster Autoscaler (buffered ratios, webhook mode, batch scaling)
   - `group_config.go` - NodeGroupSet configuration for differential buffered resource ratios

4. **metrics/** - Custom Prometheus metrics for BCS-specific monitoring

### Key Architectural Differences from Upstream

- **BCS Cloud Provider**: Communicates with BCS cluster-manager service instead of direct cloud provider APIs
- **Buffered Resource Ratio**: Supports `BufferedCPURatio`, `BufferedMemRatio`, `BufferedResourceRatio` for preemptive scaling
- **NodeGroupSet Support**: Allows grouping NodeGroups with different buffered resource ratios (see below)
- **Webhook Integration**: External scaling policy enforcement via webhook callbacks (ConfigMap and web modes)
- **Batch Scaling**: `MaxBulkScaleUpCount` and `BatchScaleUpCount` for controlled bulk operations
- **Custom Timing**: `MaxNodeStartupTime`, `MaxNodeStartScheduleTime`, `ScaleDownDelayAfterRemove`

### NodeGroupSet Feature

The NodeGroupSet feature allows you to group different NodeGroups and assign different buffered resource ratios to each group:

**Key Benefits:**
- **Cost Optimization**: Set lower buffer ratios for expensive GPU nodes, higher for cheap CPU nodes
- **Business Isolation**: Configure independent scaling policies for different business units
- **Resource Type Differentiation**: Different ratios for compute-intensive, memory-intensive, storage-intensive nodes
- **Multi-AZ Strategy**: Different scaling triggers for different availability zones

**Architecture:**
```
BufferedAutoscaler
  └── NodeGroupSetManager
        ├── Set "gpu-nodes" (buffer: 10%)
        │     ├── NodeGroup "gpu-v100"
        │     └── NodeGroup "gpu-t4"
        ├── Set "general-purpose" (buffer: 30%)
        │     ├── NodeGroup "worker-m5"
        │     └── NodeGroup "worker-m6"
        └── GlobalSet "__global__" (uses global BufferedRatios)
              └── [unmapped NodeGroups - use cluster-level ratios]
```

**GlobalSetName (`__global__`):**
- NodeGroups not explicitly assigned to any Set get `GlobalSetName = "__global__"`
- These NodeGroups use global cluster-level `BufferedCPURatio`, `BufferedMemRatio`, `BufferedResourceRatio` settings
- `GetAllSets()` does NOT return `__global__` as it's a marker, not a real Set
- Buffer checks for `__global__` NodeGroups use global ratios from command-line flags

**Configuration Example:**
```yaml
# See examples/nodegroup-sets-config.yaml
sets:
  - setName: "gpu-nodes"
    nodeGroupIDs:
      - "gpu-v100-pool"
      - "gpu-t4-pool"
    bufferedRatios:
      cpuRatio: 0.1  # Lower buffer for expensive nodes
      memRatio: 0.1
  - setName: "general-purpose"
    nodeGroupIDs:
      - "worker-m5"
      - "worker-m6"
    bufferedRatios:
      cpuRatio: 0.3  # Higher buffer for general nodes
      memRatio: 0.3
    expander: "least-waste"  # Optional: Set-level expander override
# NodeGroups not listed above use GlobalSetName and global BufferedRatios
```

**Scale-Up Logic (Separated Paths):**
1. **Pod-Triggered Scale-Up**: When pending pods exist
   - Compute expansion options for all NodeGroups
   - Use Expander strategy to select single best NodeGroup
   - Execute scale-up on selected NodeGroup only
   
2. **Buffer-Triggered Scale-Up**: When buffer is insufficient (no pending pods)
   - Filter NodeGroups to only those in triggered Sets
   - Compute expansion options per Set
   - Execute scale-up for ALL triggered Sets concurrently
   
3. **Priority**: Pending pods ALWAYS take precedence over buffer
   - If both exist, handle pods first; buffer waits for next cycle
   - Exception: If pods have no valid options, fall back to buffer expansion

**Key Implementation Files:**
- `scalingconfig/group_config.go` - Data structures (NodeGroupSet, BufferedRatioConfig)
- `core/nodegroup_set_manager.go` - Set management, resource calculation, snapshot
- `core/multi_set_expansion.go` - Multi-Set concurrent expansion logic
- `core/set_expander.go` - Set-aware Expander wrapper
- `core/buffered_autoscaler.go` - Integration with main autoscaler
- `core/scale_up.go` - Separated pod/buffer expansion paths
- `core/scale_down.go` - Prevent scale-down if buffer insufficient
- `estimator/resource_estimator.go` - Per-Set resource estimation with buffer ratios
- `metrics/metrics.go` - Prometheus metrics per set

### Data Flow

1. **Scale-Up Flow**:
   - BufferedAutoscaler main loop runs periodic checks
   - **Pod-triggered path** (priority):
     - Detects unschedulable pods
     - Computes expansion options for each NodeGroup (using ratio=0 estimator, no buffer)
     - Applies Expander strategy to select single best option
     - Executes scale-up on selected NodeGroup
   - **Buffer-triggered path** (when no pending pods):
     - NodeGroupSetManager takes resource snapshot (includes upcoming nodes)
     - Checks each Set's buffer ratios against current capacity
     - Filters NodeGroups to only triggered Sets
     - Computes expansion options per Set (using Set's buffer ratios)
     - Executes scale-up for ALL triggered Sets concurrently
   - Consults webhook (if configured) for approval
   - Calls BCS cloud provider to expand node groups
   - BCS provider communicates with cluster-manager API
   - cluster-manager creates nodes in underlying infrastructure

2. **Scale-Down Flow**:
   - Identifies underutilized nodes (below thresholds)
   - Validates pods can be evicted (respects PDBs, local storage, etc.)
   - **NodeGroupSet buffer check**:
     - Simulates node removal using `SimulateNodeRemoval()`
     - Verifies Set's buffer ratio remains satisfied after removal
     - Blocks scale-down if removal would violate buffer constraints
   - Checks custom delay timers and webhook policies
   - Removes nodes via cluster-manager API
   - Respects `ScaleDownDelayAfterRemove` before next scale-down

3. **BCS Integration**:
   - Authentication via Bearer token (environment variable or config)
   - HTTP client in `cloudprovider/bcs/clustermanager/nodepool.go`
   - Node group metadata cached locally (see `cache.go`)
   - Periodic sync with cluster-manager to update node pool state

### Configuration

Primary configuration is via command-line flags (see `main.go`):
- Standard Cluster Autoscaler flags (see upstream FAQ.md)
- BCS-specific flags: `--buffered-cpu-ratio`, `--buffered-mem-ratio`, `--webhook-mode`, etc.
- Cloud config file (optional): Passed via `--cloud-config`, parsed in `cloudprovider/bcs/cloud_provider.go`

Environment variables commonly used:
- `BCS_CLUSTER_MANAGER_URL` - Cluster manager API endpoint
- `BCS_CLUSTER_MANAGER_TOKEN` - Authentication token
- `bcs_encryption_key` - Encryption key for sensitive data (passed via LDFLAGS)

## Important Implementation Details

### Node Group Discovery
- Node groups are discovered via BCS cluster-manager API (not cloud provider auto-discovery)
- Each node group maps to a BCS NodePool resource
- NodePool templates define node capacity, labels, and launch configuration

### Webhook Mode
The autoscaler supports webhook callbacks for custom scaling policies:
- **ConfigMap mode**: Reads webhook config from Kubernetes ConfigMap
- **Web mode**: HTTP endpoint for external policy decisions
- Webhooks can approve/reject scale-up/scale-down operations
- See `core/webhook_*.go` for implementation

### Buffered Resources
Unlike standard CA, BCS-CA can maintain buffer capacity:
- `BufferedCPURatio`: Extra CPU headroom (e.g., 0.2 = 20% buffer)
- Scale-up triggered before resources are fully exhausted
- Useful for bursty workloads requiring immediate capacity

### Testing BCS Provider
When testing cloudprovider/bcs code:
1. Mock the cluster-manager HTTP client (see `clustermanager/mocks/`)
2. Set up test NodeGroup configs in tests
3. Use `cloudprovider/bcs/*_test.go` for examples
4. Consider integration tests with a test cluster-manager instance

## Dependencies

- **Go 1.17** (see go.mod)
- **Kubernetes 1.24.0** client libraries
- **Upstream Cluster Autoscaler 0.0.0-20220613131812** (vendored with BCS patches)
- Uses Go modules with vendoring (vendor/ directory)
- NEVER manually edit `go.mod` - it's generated (see upstream FAQ for vendor update process)

## Project-Specific Conventions

- All BCS-specific code is under `github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler`
- Upstream Cluster Autoscaler code imported from `k8s.io/autoscaler/cluster-autoscaler`
- Custom extensions in `core/`, `cloudprovider/bcs/`, `scalingconfig/`, `metrics/`
- Copyright header required on all files (see existing files for template)
- Use `klog` (v2) for logging, never `fmt.Println` or `log`

## Troubleshooting Development Issues

### Build Failures
- Ensure Go 1.17 is installed
- Run `go mod tidy` if dependency errors occur
- Check LDFLAGS if encryption key issues arise
- Use `make build-in-docker` for reproducible builds

### Test Failures
- Mock cluster-manager client for unit tests
- Check that test fixtures match current API schemas
- Use `-v` flag for verbose test output
- Run single package tests to isolate issues

### BCS Integration Issues
- Verify BCS_CLUSTER_MANAGER_URL and token are set
- Check network connectivity to cluster-manager
- Review logs for HTTP request/response details
- Ensure NodePool exists in cluster-manager before testing

### Webhook Debugging
- Check ConfigMap format for webhook-mode=configmap
- Verify webhook endpoint is reachable for webhook-mode=web
- Review `core/webhook_utils.go` for callback logic
- Enable debug logging with `-v=4` flag

## Running Locally

```bash
# Build the binary
make build

# Run with minimal config (requires kubeconfig)
./bcs-cluster-autoscaler \
  --kubeconfig=$HOME/.kube/config \
  --cloud-provider=bcs \
  --namespace=bcs-system \
  --v=4

# With BCS cluster-manager
export BCS_CLUSTER_MANAGER_URL=http://localhost:8080
export BCS_CLUSTER_MANAGER_TOKEN=yourtoken
./bcs-cluster-autoscaler \
  --kubeconfig=$HOME/.kube/config \
  --cloud-provider=bcs \
  --namespace=bcs-system \
  --buffered-cpu-ratio=0.2 \
  --buffered-mem-ratio=0.2
```

## Additional Resources

- See `FAQ.md` for comprehensive Cluster Autoscaler behavior documentation
- `README.md` covers basic Cluster Autoscaler concepts
- Upstream documentation: https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler
- BCS documentation: Refer to internal BCS docs for cluster-manager API specs
