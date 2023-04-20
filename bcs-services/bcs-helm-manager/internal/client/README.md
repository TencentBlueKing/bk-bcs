# bcs-helm-client 命令行工具

## 配置文件

配置文件默认放在 `/etc/bcs/helmctl.yaml` 文件：
```yaml
config:
  apiserver: ""
  bcs_token: ""
  operator: ""
```

## 使用文档

### 检查连通性 - Available

```bash
bcs-helm-client available --help
```

参数详情:

```yaml 
无参
```

示例:

```
bcs-helm-client available
```



### 获取仓库列表 - ListRepository

```bash
bcs-helm-client list rp --help 
```
参数详情:
```yaml 
  -A, --all                 list all records
      --cluster string      release cluster id for operation
  -c, --config string       config file (default "./etc/bcs/helmctl.yaml")
      --name string         release name for operation
  -n, --namespace string    release namespace for operation
      --num int             list records num (default 20)
  -o, --output string       output format, one of json|wide
  -p, --project string      project id for operation
  -r, --repository string   repository name for operation

```

示例:

```
bcs-helm-client list rp -p [projectname to query]
```



### 获取仓库明细 - GetRepository

```bash
bcs-helm-client get repo --help
```

参数详情:

```yaml 
      --cluster string      release cluster id for operation
  -c, --config string       config file (default "./etc/bcs/helmctl.yaml")
      --name string         release name for operation
  -n, --namespace string    release namespace for operation
  -o, --output string       output format, one of json|wide
  -p, --project string      project id for operation
  -r, --repository string   repository name for operation

```

示例:

```
bcs-helm-client get repo -p [projectname to query] -r [repositoryname to query]
```



### 获取Chart列表V1版 - ListChartV1

```bash
bcs-helm-client list ch --help
```

参数详情:

```yaml 
  -A, --all                 list all records
      --cluster string      release cluster id for operation
  -c, --config string       config file (default "./etc/bcs/helmctl.yaml")
      --name string         release name for operation
  -n, --namespace string    release namespace for operation
      --num int             list records num (default 20)
  -o, --output string       output format, one of json|wide
  -p, --project string      project id for operation
  -r, --repository string   repository name for operation

```

示例:

```
list ch -p [projectname to query] -r [repositoryname to query] --name [name to query]
```





### 获取Chart明细V1版 - GetChartDetailV1

```bash
bcs-helm-client get detail --help
```

参数详情:

```yaml 
      --cluster string      release cluster id for operation
  -c, --config string       config file (default "./etc/bcs/helmctl.yaml")
      --name string         release name for operation
  -n, --namespace string    release namespace for operation
  -o, --output string       output format, one of json|wide
  -p, --project string      project id for operation
  -r, --repository string   repository name for operation

```

示例:

```
bcs-helm-client get detail [chartname to query] -p [projectname to query] -r [repositoryname to query]
```





### 获取ChartVersion明细V1版 - GetVersionDetailV1

### 

```bash
bcs-helm-client get vdv1  --help
```

参数详情:

```yaml 
      --cluster string      release cluster id for operation
  -c, --config string       config file (default "./etc/bcs/helmctl.yaml")
      --name string         release name for operation
  -n, --namespace string    release namespace for operation
  -o, --output string       output format, one of json|wide
  -p, --project string      project id for operation
  -r, --repository string   repository name for operation

```

示例:

```bash
bcs-helm-client get vdv1 [chartname to query] [version to query] -p [projectname to query] -r [repositoryname to query]
```



### 删除Chart - DeleteChart

```bash
bcs-helm-client delete chart --help
```

参数详情:

```yaml 
  -c, --config string       config file (default "./etc/bcs/helmctl.yaml")
  -d, --data string         resource json data
  -f, --file string         resource json file
  -p, --project string      project id for operation
  -r, --repository string   repository name for operation
```

示例:

```bash
bcs-helm-client delete [chartname to delete] -p [projectname to query] -r [repositoryname to query]
```





### 删除ChartVersion - DeleteChartVersion

### 

```bash
bcs-helm-client delete chv --help
```

参数详情:

```yaml 
  -c, --config string       config file (default "./etc/bcs/helmctl.yaml")
  -d, --data string         resource json data
  -f, --file string         resource json file
  -p, --project string      project id for operation
  -r, --repository string   repository name for operation
```

示例:

```bash
bcs-helm-client delete [chartname to delete] [chartversion to delete] -p [projectname to query] -r [repositoryname to query]
```









### 获取ChartRelease列表V1版 - ListReleaseV1

### 

```bash
bcs-helm-client list rl --help
```

参数详情:

```yaml 
  -A, --all                 list all records
      --cluster string      release cluster id for operation
  -c, --config string       config file (default "./etc/bcs/helmctl.yaml")
      --name string         release name for operation
  -n, --namespace string    release namespace for operation
      --num int             list records num (default 20)
  -o, --output string       output format, one of json|wide
  -p, --project string      project id for operation
  -r, --repository string   repository name for operation
```

示例:

```bash
bcs-helm-client list rl -p [projectname to query] -r [repositoryname to query] --name [releasename to query] -n [namespace to query] --cluster [cluster to query]
bcs-helm-client list rl -p project -r repo --name releasename -n default --cluster BCS-K8S-00000
```



### 获取ChartRelease明细V1版 - GetReleaseDetailV1

### 

```bash
bcs-helm-client get release --help
```

子命令Aliases:

```bash
release, rl
```

参数详情:

```yaml 
      --cluster string      release cluster id for operation
  -c, --config string       config file (default "./etc/bcs/helmctl.yaml")
      --name string         release name for operation
  -n, --namespace string    release namespace for operation
  -o, --output string       output format, one of json|wide
  -p, --project string      project id for operation
  -r, --repository string   repository name for operation

```

示例:

```bash
bcs-helm-client get release  --name [releasename to query] -p [projectname to query]  -n [namespace to query] --cluster [cluster to query]
```



### 安装ChartRelease明细V1版 - InstallReleaseV1

### 

```bash
bcs-helm-client install  --help
```

参数详情:

```yaml 
      --args string         args to append to helm command
      --cluster string      release cluster id for operation
  -f, --file strings        value file for installation
  -h, --help                help for installv1
  -n, --namespace string    release namespace for operation
  -p, --project string      project id for operation
  -r, --repository string   repository name for operation
      --sysvar string       sys var file

```

示例:

```bash
bcs-helm-client  install [releasename to install] [chartname to install] [version to install] -n [namespace to install] -p [projectname to install]  -r [repositoryname to install] --cluster [cluster to install] -f [value file for installation]
```



### 卸载ChartRelease明细V1版 - UninstallReleaseV1

### 

```bash
bcs-helm-client uninstall --help
```

参数详情:

```yaml 
      --cluster string     release cluster id for operation
  -h, --help               help for uninstallv1
  -n, --namespace string   release namespace for operation
  -p, --project string     release project for operation
```

示例:

```bash
bcs-helm-client uninstall [releasename to uninstall] -n [namespace to install] -p [projectname to install]  --cluster [cluster to uninstall]
```



### 升级ChartRelease明细V1版 - UpgradeReleaseV1

### 

```bash
bcs-helm-client upgrade --help
```

参数详情:

```yaml 
      --args string         args to append to helm command
      --cluster string      release cluster id for operation
  -f, --file strings        value file for installation
  -h, --help                help for upgradev1
  -n, --namespace string    release namespace for operation
  -p, --project string      project id for operation
  -r, --repository string   repository name for operation
      --sysvar string       sys var file

```

示例:

```bash
bcs-helm-client upgrade   [releasename to upgrade] [chartname to upgrade] [version to upgrade]  -n [namespace to upgrade] -p [projectname to upgrade]  -r [repositoryname to upgrade] --cluster [cluster to upgrade] -f [value file for upgrade]
```



### 回滚ChartRelease明细V1版 - RollbackReleaseV1

### 

```bash
bcs-helm-client rollback --help
```

参数详情:

```yaml 
      --cluster string     release cluster id for operation
  -h, --help               help for rollbackv1
  -n, --namespace string   release namespace for operation
  -p, --project string     project id for operation
```

示例:

```bash
bcs-helm-client  rollback  [releasename to upgrade] [releasename to upgrade(类型为正整数)] -n [namespace to upgrade] -p [projectname to upgrade] --cluster [cluster to upgrade] 
```





### 获取Release历史信息 - GetReleaseHistory

### 

```bash
bcs-helm-client get rlh --help
```

参数详情:

```yaml 
      --cluster string      release cluster id for operation
  -c, --config string       config file (default "./etc/bcs/helmctl.yaml")
      --name string         release name for operation
  -n, --namespace string    release namespace for operation
  -o, --output string       output format, one of json|wide
  -p, --project string      project id for operation
  -r, --repository string   repository name for operation
```

示例:

```bash
bcs-helm-client get rlh [releasename to query] -n [namespace to upgrade] -p [projectname to upgrade] --cluster [cluster to upgrade] 
```





## 如何编译

找到bcs-helm-manager makefile文件,执行下述命令编译 Client 工具
```
make client
```