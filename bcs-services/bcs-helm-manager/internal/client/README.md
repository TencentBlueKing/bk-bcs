# helmctl

bcs helm 命令行工具

## 安装

```bash
git clone https://github.com/TencentBlueKing/bk-bcs
cd bcs-services/bcs-helm-manager
make helmctl
```
根据编译信息找到 `helmctl` 可执行文件所在目录。

## 配置文件

配置文件默认读取 `/etc/bcs/helmctl.yaml`，也可以通过 `--config=config.yaml` 指定配置文件。
示例：
```yaml
config:
  apiserver: ""
  token: ""
```

## 使用文档

### 1. 检查连通性

```bash
$ helmctl available
ok
```

### 2. 获取仓库

```bash
$ helmctl get repo -p <project_code>
```

### 3. 获取 Chart

```bash
$ helmctl get chart -p <project_code> 
$ helmctl get chart -p <project_code> <chart_name>
```

### 4. 获取 Chart 版本

```bash
$ helmctl get chartVersion -p <project_code> <chart_name>
$ helmctl get chartVersion -p <project_code> <chart_name> <version>
```

### 5. 删除 Chart

```bash
$ helmctl delete chart -p <project_code> <chart_name>
```

### 6. 删除 Chart 版本

```bash
$ helmctl delete chartVersion -p <project_code> <chart_name> <version>
```

### 7. 获取 Release

```bash
$ helmctl get release -p <project_code> -c <cluster_id> -n <namespace>
$ helmctl get release -p <project_code> -c <cluster_id> -n <namespace> <name>
$ helmctl get release -p <project_code> -c <cluster_id> -A
```

### 8. 获取 Release 历史版本

```bash
$ helmctl history -p <project_code> -c <cluster_id> -n <namespace> <release_name>
```

### 9. 安装 Release

```bash
$ helmctl install -p <project_code> -c <cluster_id> -n <namespace> <release_name> <chart_name> <version> -f values.yaml --args=--wait=true --args=--timeout=600s
```

### 10. 更新 Release

```bash
$ helmctl upgrade -p <project_code> -c <cluster_id> -n <namespace> <release_name> <chart_name> <version> -f values.yaml --args=--wait=true --args=--timeout=600s
```

### 11. 回滚 Release

```bash
$ helmctl rollback -p <project_code> -c <cluster_id> -n <namespace> <release_name> <revision>
```

### 12. 卸载 Release

```bash
$ helmctl uninstall -p <project_code> -c <cluster_id> -n <namespace> <release_name>
```

### 13. 差异化显示revison Release revison

```bash
$ helmctl diff revision -p <project_code> -c <cluster_id> -n <namespace> <release_name> <revision1> <revision2>
```

### 14. 差异化显示回滚版本 Release rollback

```bash
$ helmctl diff rollback -p <project_code> -c <cluster_id> -n <namespace> <release_name> <revision>
```

### 15. 差异化显示values.yaml更新 Release upgrade

```bash
$ helmctl diff upgrade -p <project_code> -c <cluster_id> -n <namespace> <release_name> <chart_name> <version> -f values.yaml
```
