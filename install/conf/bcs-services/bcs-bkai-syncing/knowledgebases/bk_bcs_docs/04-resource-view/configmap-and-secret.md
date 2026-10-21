# 配置管理指南（ConfigMap 与 Secret）

为了遵循云原生“配置与镜像代码分离”的最佳实践，蓝鲸容器管理平台（BCS）在【资源视图】→【配置】模块中提供了 **ConfigMap**（非敏感配置）、**Secret**（敏感凭据）及 **BscpConfigs** 的可视化创建、更新与挂载管理。

---

## 一、ConfigMap 管理（非敏感配置）

**ConfigMap** 用于以键值对（Key-Value）或完整配置文件的形式存储非敏感配置数据（如应用配置文件 `app.yaml`、`nginx.conf` 或环境变量参数）。

### 1. 创建与更新 ConfigMap
1. 进入【资源视图】→【配置】→【ConfigMaps】，选择目标集群与命名空间，点击【创建】。
2. 支持通过**表单**添加多个 `Key` 和 `Value`（Value 可以是单行字符串或多行配置文件内容），或切换至 **YAML 模式**直接编辑：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  APP_ENV: "production"
  LOG_LEVEL: "info"
  nginx.conf: |
    server {
        listen 80;
        server_name localhost;
        location / {
            root /usr/share/nginx/html;
            index index.html;
        }
    }
```

3. 在 ConfigMap 列表中，可随时点击【查看】检查配置详情，或点击【编辑】/【删除】管理存量配置。

---

## 二、Secret 管理（敏感数据与凭据）

**Secret** 用于存储密码、Token、SSH 密钥、TLS 证书及私有镜像仓库拉取凭据等敏感数据。Kubernetes 在存储 Secret 时会对 `data` 字段进行 Base64 编码（或使用 `stringData` 传入明文由系统自动编码）。

### 1. 常见 Secret 类型

| Secret 类型 (`type`) | 用途说明 | 必需字段 |
| :--- | :--- | :--- |
| `Opaque`（默认） | 用户自定义的任意敏感键值对（如数据库密码、API 密钥） | 自定义 Key-Value |
| `kubernetes.io/tls` | 存储 Ingress 或服务使用的 TLS 证书与私钥 | `tls.crt`（证书）、`tls.key`（私钥） |
| `kubernetes.io/dockerconfigjson` | 存储私有容器镜像仓库的登录凭据，供 Pod 的 `imagePullSecrets` 引用 | `.dockerconfigjson` |
| `kubernetes.io/service-account-token` | 关联 ServiceAccount 的访问令牌 | `kubernetes.io/service-account.name` 注解 |

### 2. Secret YAML 示例

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-db-secret
  namespace: default
type: Opaque
stringData:
  DB_USER: "app_reader"
  DB_PASSWORD: "<YOUR_DB_PASSWORD>"
```

---

## 三、在 Pod / 工作负载中使用 ConfigMap 与 Secret

创建好 ConfigMap 或 Secret 后，通常有两种方式注入到业务容器中：

### 方式 1：作为环境变量（`env` / `envFrom`）注入容器

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-app
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: demo-app
  template:
    metadata:
      labels:
        app: demo-app
    spec:
      containers:
        - name: app
          image: nginx:1.24.0
          env:
            # 从 ConfigMap 引用单个键
            - name: LOG_LEVEL
              valueFrom:
                configMapKeyRef:
                  name: app-config
                  key: LOG_LEVEL
            # 从 Secret 引用单个敏感键
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: app-db-secret
                  key: DB_PASSWORD
```

### 方式 2：作为数据卷（`Volume`）挂载为容器内文件

```yaml
      containers:
        - name: app
          image: nginx:1.24.0
          volumeMounts:
            # 将 ConfigMap 中的 nginx.conf 挂载到指定目录
            - name: config-volume
              mountPath: /etc/nginx/conf.d/default.conf
              subPath: nginx.conf
      volumes:
        - name: config-volume
          configMap:
            name: app-config
```

> **重要注意事项（热更新行为差异）**：
> - **环境变量注入**（`env` / `envFrom`）和使用了 **`subPath` 挂载**的文件，在修改 ConfigMap/Secret 后**不会自动更新**到运行中的容器内，必须触发工作负载滚动重启 Pod 才能生效。
> - **普通目录 Volume 挂载**（未使用 `subPath`）在修改 ConfigMap/Secret 后，Kubelet 会在约 1~2 分钟内自动将新内容同步更新到容器挂载目录中。
