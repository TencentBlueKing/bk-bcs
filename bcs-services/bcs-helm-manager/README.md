# HELM MANAGER

作为helm-service下沉的服务, 提供repository, chart, release等维度的helm操作.

## 快速编译镜像

```bash
cd helm/image
sh build.sh -v test
```

## 快速部署

打包chart
```bash
cd helm/chart/
helm package .
```

部署chart
```bash
helm install bcs-helm-manager bcs-helm-manager-0.1.0.tgz -n bcs-system -f values.yaml
```

values.yaml
```yaml
image:
  registry:
  repository: bcs-helm-manager
  tag: test
  pullPolicy: Always

helmmanager:
  repo:
    url: bkrepo.com

volumeMounts: []

volumes: []
```