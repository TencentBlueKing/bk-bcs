# GPU资源注入插件 (gpuinjector)

## 背景

在 Kubernetes 集群中使用 GPU 资源时，通常需要为 GPU 容器配置额外的 CPU、内存等资源，以确保 GPU 工作负载能够正常运行。不同型号的 GPU 对 CPU、内存等资源的需求不同，手动配置这些资源容易出错且效率低下。

gpuinjector 插件作为 bcs-webhook-server 的一个插件，通过 Mutating Webhook 机制，在 Pod 创建时自动根据 GPU 类型和数量，为 GPU 容器注入相应的 CPU、内存等资源，简化 GPU 工作负载的配置管理。

## 功能特性

* **自动资源注入**：根据 GPU 类型和数量，自动为 GPU 容器注入 CPU、内存等资源
* **多 GPU 类型支持**：支持配置多种 GPU 类型（如 V100、T4、A100 等），每种类型可配置不同的资源系数
* **资源系数配置**：支持为每种 GPU 类型配置 CPU、内存、存储等资源的系数，支持扩展资源
* **Annotation 注入**：支持为 Pod 注入自定义的 Annotation，用于网络、存储等配置
* **Namespace 特定规则**：支持为特定 namespace 配置独立的 GPU 资源注入规则，实现更细粒度的资源管理
* **资源共享计算**：自动计算 Pod 中 GPU 容器和非 GPU 容器之间的资源分配，确保资源合理分配

## 工作原理

### Hook 流程

gpuinjector 插件通过以下流程工作：

1. **识别 Pod**：拦截 Pod 的创建请求（Create 操作）
2. **检查 Annotation**：检查 Pod 是否包含 `task.bkbcs.tencent.com/gpu-type` annotation
3. **匹配配置**：
   - 优先检查 namespace 特定规则（`NamespaceGPUResourceMap`）
   - 如果 namespace 特定规则未匹配，则使用默认规则（`GPUResourceMap`）
4. **计算资源**：根据 GPU 数量和配置的资源系数，计算需要注入的资源
5. **生成 Patch**：生成 JSON Patch 操作，注入资源和 Annotation

### 资源计算逻辑

1. **总资源计算**：根据 Pod 中所有 GPU 容器的 GPU 数量总和，计算 Pod 所需的总资源
   - 总资源 = GPU 数量 × 资源系数
2. **资源扣除**：从总资源中扣除非 GPU 容器已申请的资源
3. **按 GPU 分配**：将剩余资源平均分配给每个 GPU，得到每个 GPU 容器应注入的资源

## 配置说明

### 配置文件格式

gpuinjector 插件的配置文件为 JSON 格式，包含以下字段：

```json
{
  "resourceMap": {
    "GPU类型": {
      "GPU资源名称": {
        "resourceList": [
          {
            "name": "资源名称",
            "coefficient": 资源系数,
            "unit": "单位"
          }
        ],
        "annotations": {
          "annotation键": "annotation值"
        }
      }
    }
  },
  "namespaceResourceMap": {
    "namespace名称": {
      "GPU类型": {
        "GPU资源名称": {
          "resourceList": [...],
          "annotations": {...}
        }
      }
    }
  }
}
```

### 配置字段说明

#### resourceMap

默认的 GPU 资源映射配置，格式为：`map[GPU类型]map[资源名称]InjectInfo`

- **GPU类型**：GPU 的型号标识，如 "V100"、"T4"、"A100" 等
- **资源名称**：Kubernetes 资源名称，如 "nvidia.com/gpu"、"amd.com/gpu" 等
- **InjectInfo**：注入信息
  - **resourceList**：资源列表，每个资源包含：
    - **name**：资源名称（如 "cpu"、"memory"、"ephemeral-storage" 或扩展资源）
    - **coefficient**：资源系数（浮点数），表示每个 GPU 需要多少该资源
    - **unit**：资源单位（如 "m" 表示毫核，"Gi" 表示 Gibibyte）
  - **annotations**：需要注入的 Pod Annotation（可选）

#### namespaceResourceMap

Namespace 特定的 GPU 资源映射配置，格式为：`map[Namespace]map[GPU类型]map[资源名称]InjectInfo`

- **Namespace名称**：Kubernetes Namespace 名称
- 其他字段与 `resourceMap` 相同

**优先级**：如果 Pod 的 namespace 在 `namespaceResourceMap` 中有配置，且 GPU 类型和资源名称都匹配，则优先使用 namespace 特定规则；否则使用 `resourceMap` 中的默认规则。

### 配置示例

#### 基础配置示例

```json
{
  "resourceMap": {
    "V100": {
      "nvidia.com/gpu": {
        "resourceList": [
          {
            "name": "cpu",
            "coefficient": 4000,
            "unit": "m"
          },
          {
            "name": "memory",
            "coefficient": 8,
            "unit": "Gi"
          },
          {
            "name": "nvidia.com/shm",
            "coefficient": 1,
            "unit": "Gi"
          }
        ],
        "annotations": {
          "tke.cloud.tencent.com/networks": "tke-route-eni"
        }
      }
    },
    "T4": {
      "nvidia.com/gpu": {
        "resourceList": [
          {
            "name": "cpu",
            "coefficient": 2000,
            "unit": "m"
          },
          {
            "name": "memory",
            "coefficient": 4,
            "unit": "Gi"
          }
        ]
      }
    }
  }
}
```

#### 包含 Namespace 特定规则的配置示例

```json
{
  "resourceMap": {
    "V100": {
      "nvidia.com/gpu": {
        "resourceList": [
          {
            "name": "cpu",
            "coefficient": 4000,
            "unit": "m"
          },
          {
            "name": "memory",
            "coefficient": 8,
            "unit": "Gi"
          }
        ],
        "annotations": {
          "default.annotation": "default-value"
        }
      }
    }
  },
  "namespaceResourceMap": {
    "gpu-training": {
      "V100": {
        "nvidia.com/gpu": {
          "resourceList": [
            {
              "name": "cpu",
              "coefficient": 6000,
              "unit": "m"
            },
            {
              "name": "memory",
              "coefficient": 12,
              "unit": "Gi"
            },
            {
              "name": "nvidia.com/shm",
              "coefficient": 2,
              "unit": "Gi"
            }
          ],
          "annotations": {
            "training.annotation": "training-value",
            "tke.cloud.tencent.com/networks": "tke-route-eni"
          }
        }
      }
    },
    "gpu-inference": {
      "T4": {
        "nvidia.com/gpu": {
          "resourceList": [
            {
              "name": "cpu",
              "coefficient": 1000,
              "unit": "m"
            },
            {
              "name": "memory",
              "coefficient": 2,
              "unit": "Gi"
            }
          ]
        }
      }
    }
  }
}
```

## 使用示例

### 基础使用

在 Pod 的 Annotation 中指定 GPU 类型：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: gpu-pod
  namespace: default
  annotations:
    task.bkbcs.tencent.com/gpu-type: "V100"
spec:
  containers:
  - name: gpu-container
    image: nvidia/cuda:11.0-base
    resources:
      requests:
        nvidia.com/gpu: "2"
      limits:
        nvidia.com/gpu: "2"
```

注入后的 Pod（假设配置了 V100 的资源系数：CPU 4000m/GPU，Memory 8Gi/GPU）：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: gpu-pod
  namespace: default
  annotations:
    task.bkbcs.tencent.com/gpu-type: "V100"
    tke.cloud.tencent.com/networks: "tke-route-eni"
spec:
  containers:
  - name: gpu-container
    image: nvidia/cuda:11.0-base
    resources:
      requests:
        nvidia.com/gpu: "2"
        cpu: "8"          # 2 GPU × 4000m = 8000m = 8
        memory: "16Gi"    # 2 GPU × 8Gi = 16Gi
      limits:
        nvidia.com/gpu: "2"
        cpu: "8"
        memory: "16Gi"
```

### 使用 Namespace 特定规则

在特定 namespace 中使用不同的资源配置：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: training-pod
  namespace: gpu-training
  annotations:
    task.bkbcs.tencent.com/gpu-type: "V100"
spec:
  containers:
  - name: gpu-container
    image: nvidia/cuda:11.0-base
    resources:
      requests:
        nvidia.com/gpu: "2"
      limits:
        nvidia.com/gpu: "2"
```

如果配置了 `namespaceResourceMap` 中 `gpu-training` namespace 的特定规则，则会使用该规则进行注入，而不是默认的 `resourceMap` 规则。

### 多容器 Pod 示例

Pod 中包含 GPU 容器和非 GPU 容器时，资源会自动分配：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mixed-pod
  namespace: default
  annotations:
    task.bkbcs.tencent.com/gpu-type: "V100"
spec:
  containers:
  - name: gpu-container
    image: nvidia/cuda:11.0-base
    resources:
      requests:
        nvidia.com/gpu: "2"
  - name: cpu-container
    image: busybox
    resources:
      requests:
        cpu: "1000m"
        memory: "2Gi"
```

假设配置为每个 GPU 需要 4000m CPU 和 8Gi Memory：
- 总资源需求：2 GPU × (4000m CPU + 8Gi Memory) = 8000m CPU + 16Gi Memory
- 扣除非 GPU 容器资源：8000m - 1000m = 7000m CPU，16Gi - 2Gi = 14Gi Memory
- GPU 容器分配：7000m / 2 = 3500m CPU per GPU，14Gi / 2 = 7Gi Memory per GPU

注入后的 GPU 容器资源：
- CPU: 3500m × 2 = 7000m
- Memory: 7Gi × 2 = 14Gi

## 资源类型说明

### 标准资源

支持 Kubernetes 标准资源：
- **cpu**：CPU 资源，单位通常为 "m"（毫核）
- **memory**：内存资源，单位通常为 "Mi"、"Gi" 等
- **ephemeral-storage**：临时存储，单位通常为 "Gi"

### 扩展资源

支持 Kubernetes 扩展资源，如：
- **nvidia.com/shm**：NVIDIA 共享内存
- **tke.cloud.tencent.com/qgpu-core**：腾讯云 QGPU 核心数
- 其他自定义扩展资源

扩展资源会直接按照系数和 GPU 数量计算，不参与资源共享计算。

## 注意事项

1. **Annotation 要求**：Pod 必须包含 `task.bkbcs.tencent.com/gpu-type` annotation 才会触发注入
2. **GPU 资源识别**：插件会识别 Pod 中容器的 GPU 资源请求，支持的资源名称需要在配置中定义
3. **资源冲突**：如果 Pod 中不同容器使用了不同的 GPU 资源类型（如同时使用 nvidia.com/gpu 和 amd.com/gpu），插件会报错
4. **资源不足**：如果非 GPU 容器申请的资源超过 GPU 容器应分配的资源，插件会报错
5. **Namespace 规则优先级**：Namespace 特定规则的优先级高于默认规则，但只有在完全匹配（namespace、GPU 类型、资源名称都匹配）时才会生效
6. **仅支持 Create 操作**：插件只处理 Pod 的创建请求，更新操作不会触发注入

## 配置最佳实践

1. **合理设置资源系数**：根据实际 GPU 工作负载的需求，合理设置 CPU、内存等资源的系数
2. **使用 Namespace 隔离**：为不同用途的 namespace（如训练、推理）配置不同的资源规则
3. **扩展资源配置**：根据实际需求配置扩展资源，如 NVIDIA 共享内存等
4. **Annotation 注入**：利用 Annotation 注入功能，自动配置网络、存储等基础设施相关配置
5. **测试验证**：在生产环境使用前，建议在测试环境充分验证配置的正确性
