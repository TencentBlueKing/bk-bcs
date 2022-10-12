# bcs-user-manager 命令行工具

## 配置文件

配置文件默认放在 `/etc/bcs/bcs-user-manager.yaml` 文件：
```yaml
config:
  apiserver: "${BCS APISERVER地址}"
  bcs_token: "${Token信息}"
```

## 使用文档

### 获取项目列表 - ListProjects

```bash
kubectl-bcs-user-manager list project --help
```
参数详情:
```yaml 
--kind           "项目中集群类型, 允许k8s/mesos"  
--names          "项目中文名称, 长度不能超过64字符, 多个以半角逗号分隔"
--project_code   "项目编码(英文缩写), 全局唯一, 长度不能超过64字符"
--project_ids    "项目ID, 多个以半角逗号分隔"
--search_name    "项目中文名称, 通过此字段模糊查询项目信息"
```


## 如何编译

执行下述命令编译 Client 工具
```
make bin
```