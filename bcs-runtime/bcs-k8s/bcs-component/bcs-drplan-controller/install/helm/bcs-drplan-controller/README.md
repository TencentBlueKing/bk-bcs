# BCS DR Plan Controller Helm Chart

Kubernetes Operator for Disaster Recovery Plan orchestration and execution.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+

## Installing the Chart

### Install with default configuration

```bash
helm install bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  --namespace bcs-system \
  --create-namespace
```

### Install with custom values

```bash
helm install bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  --namespace bcs-system \
  --create-namespace \
  --set image.repository=your-registry/bcs-drplan-controller \
  --set image.tag=v1.0.0 \
  --set controller.logLevel=4
```

### Install from values file

```bash
helm install bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  --namespace bcs-system \
  --create-namespace \
  -f custom-values.yaml
```

## Uninstalling the Chart

```bash
helm uninstall bcs-drplan-controller --namespace bcs-system
```

**Note**: CRDs are kept by default. To remove CRDs:

```bash
kubectl delete crd drworkflows.dr.bkbcs.tencent.com
kubectl delete crd drplans.dr.bkbcs.tencent.com
kubectl delete crd drplanexecutions.dr.bkbcs.tencent.com
```

## Configuration

The following table lists the configurable parameters of the chart and their default values.

### Basic Configuration

| Parameter          | Description                   | Default                 |
| ------------------ | ----------------------------- | ----------------------- |
| `replicaCount`     | Number of controller replicas | `1`                     |
| `image.repository` | Image repository              | `bcs-drplan-controller` |
| `image.pullPolicy` | Image pull policy             | `IfNotPresent`          |
| `image.tag`        | Image tag                     | Chart appVersion        |
| `imagePullSecrets` | Image pull secrets            | `[]`                    |

### Controller Configuration

| Parameter                       | Description                             | Default |
| ------------------------------- | --------------------------------------- | ------- |
| `controller.logLevel`           | Log level (0=Info, 4=Debug)             | `0`     |
| `controller.leaderElection`     | Enable leader election                  | `true`  |
| `controller.metricsAddr`        | Metrics bind address                    | `:8080` |
| `controller.healthProbeAddr`    | Health probe bind address               | `:8081` |
| `controller.reconcileFrequency` | Reconcile frequency for DRPlanExecution | `30s`   |
| `controller.enableHTTP2`        | Enable HTTP/2                           | `false` |

### Webhook Configuration

| Parameter                          | Description                                       | Default |
| ---------------------------------- | ------------------------------------------------- | ------- |
| `webhook.enabled`                  | Enable webhook server                             | `true`  |
| `webhook.port`                     | Webhook port                                      | `9443`  |
| `webhook.certificate.autoGenerate` | Auto-generate self-signed certificates (10 years) | `true`  |
| `webhook.certificate.certPem`      | Custom certificate (base64 encoded)               | `""`    |
| `webhook.certificate.keyPem`       | Custom private key (base64 encoded)               | `""`    |
| `webhook.certificate.caBundle`     | Custom CA bundle (base64 encoded)                 | `""`    |

### Resource Configuration

| Parameter                   | Description    | Default |
| --------------------------- | -------------- | ------- |
| `resources.limits.cpu`      | CPU limit      | `500m`  |
| `resources.limits.memory`   | Memory limit   | `512Mi` |
| `resources.requests.cpu`    | CPU request    | `100m`  |
| `resources.requests.memory` | Memory request | `128Mi` |

### RBAC Configuration

| Parameter                    | Description                  | Default |
| ---------------------------- | ---------------------------- | ------- |
| `rbac.create`                | Create RBAC resources        | `true`  |
| `rbac.additionalRules`       | Additional ClusterRole rules | `[]`    |
| `serviceAccount.create`      | Create ServiceAccount        | `true`  |
| `serviceAccount.name`        | ServiceAccount name          | `""`    |
| `serviceAccount.annotations` | ServiceAccount annotations   | `{}`    |

### Metrics Configuration

| Parameter                         | Description            | Default     |
| --------------------------------- | ---------------------- | ----------- |
| `metrics.enabled`                 | Enable metrics service | `true`      |
| `metrics.type`                    | Service type           | `ClusterIP` |
| `metrics.port`                    | Service port           | `8080`      |
| `metrics.serviceMonitor.enabled`  | Create ServiceMonitor  | `false`     |
| `metrics.serviceMonitor.interval` | Scrape interval        | `30s`       |

### Security Configuration

| Parameter                                  | Description                | Default |
| ------------------------------------------ | -------------------------- | ------- |
| `podSecurityContext.runAsNonRoot`          | Run as non-root user       | `true`  |
| `podSecurityContext.runAsUser`             | User ID                    | `65532` |
| `podSecurityContext.fsGroup`               | FS group ID                | `65532` |
| `securityContext.allowPrivilegeEscalation` | Allow privilege escalation | `false` |
| `securityContext.readOnlyRootFilesystem`   | Read-only root filesystem  | `true`  |

### Scheduling Configuration

| Parameter           | Description         | Default |
| ------------------- | ------------------- | ------- |
| `nodeSelector`      | Node selector       | `{}`    |
| `tolerations`       | Tolerations         | `[]`    |
| `affinity`          | Affinity rules      | `{}`    |
| `priorityClassName` | Priority class name | `""`    |

## Examples

### Example 1: Development deployment with debug logging

```yaml
# dev-values.yaml
controller:
  logLevel: 4
  leaderElection: false

resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 200m
    memory: 256Mi
```

```bash
helm install bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  -f dev-values.yaml \
  --namespace bcs-system \
  --create-namespace
```

### Example 2: Production deployment with monitoring

```yaml
# prod-values.yaml
replicaCount: 2

image:
  repository: your-registry.com/bcs-drplan-controller
  tag: v1.0.0

controller:
  logLevel: 0
  leaderElection: true
  reconcileFrequency: 30s

resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi

metrics:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 30s

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchLabels:
            app.kubernetes.io/name: bcs-drplan-controller
        topologyKey: kubernetes.io/hostname
```

```bash
helm install bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  -f prod-values.yaml \
  --namespace bcs-system \
  --create-namespace
```

### Example 3: Without cert-manager (manual certificates)

```yaml
# manual-cert-values.yaml
webhook:
  enabled: true
  certManager:
    enabled: false
```

```bash
# Create webhook certificate manually
kubectl create secret tls bcs-drplan-controller-webhook-cert \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key \
  --namespace bcs-system

# Install chart
helm install bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  -f manual-cert-values.yaml \
  --namespace bcs-system
```

## Upgrading

```bash
helm upgrade bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  --namespace bcs-system \
  -f your-values.yaml
```

## Troubleshooting

### Check controller logs

```bash
kubectl logs -l control-plane=controller-manager -n bcs-system -f
```

### Check CRD installation

```bash
kubectl get crd | grep dr.bkbcs.tencent.com
```

### Check webhook configuration

```bash
kubectl get mutatingwebhookconfiguration | grep drplan
kubectl get validatingwebhookconfiguration | grep drplan
```

### Debug webhook certificate

```bash
kubectl get certificate -n bcs-system
kubectl describe certificate bcs-drplan-controller-serving-cert -n bcs-system
```

## Support

For issues and questions, please visit:
- GitHub: https://github.com/Tencent/bk-bcs
- Documentation: https://github.com/Tencent/bk-bcs/tree/master/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/docs
