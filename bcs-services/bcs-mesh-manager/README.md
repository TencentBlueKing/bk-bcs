## bcs-mesh-manager
- istio 组件的安装和升级


## 开发

### proto生成
```
make proto
```

### 本地Linux构建

```
make build
```

### 镜像构建 (buildx)
```
make push-image
```

### 本地执行
```
# 打开ENV_USE_SERVICE_DISCOVERY是不走etcd注册中心，而是直接使用service
ENV_USE_SERVICE_DISCOVERY=true ./bin/bcs-mesh-manager -f ./config/sample/bcs-mesh-manager.json
```

## istio版本配置说明

```yaml
version:
  - name: 1.24
    chartVersion: 1.24-bcs.1 # 对应chart版本
    kubeVersion: xxx # 支持的k8s版本SemVer
    enabled: true  # 是否开放
  - name: 1.22
    chartVersion: 1.22-bcs.1
    kubeVersion: xxx
    enabled: true
  - name: 1.20
    chartVersion: 1.20-bcs.1
    kubeVersion: xxx
    enabled: true
  - name: 1.18
    chartVersion: 1.18-bcs.2
    enabled: true

# 通用配置，
featureConfig:
  outboundTrafficPolicy:
    default: ALLOW_ANY
    enabled: true
  holdApplicationUntilProxyStarts:
    default: true
    enabled: true
  exitOnZeroActiveConnections:
    default: true  # 默认只
    enabled: true  # 是否开放设置
    supportVersion: >=1.12 # 支持的版本 SemVer
... 其他更多配置
```