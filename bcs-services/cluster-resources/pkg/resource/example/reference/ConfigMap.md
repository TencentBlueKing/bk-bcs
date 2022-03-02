# ConfigMap

> ConfigMap（CM）用于将非机密的数据保存到键值对中的一种 API 对象，Pod 可以将其用作环境变量，命令行参数或者存储卷中的配置文件。

## 使用 ConfigMap

常见的 ConfigMap 配置示例如下（注意使用该 CM 的 Pod 必须在统一命名空间下）

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: game-demo
data:
  # 类属性键；每一个键都映射到一个简单的值
  player_initial_lives: "3"
  ui_properties_file_name: "user-interface.properties"

  # 类文件键
  game.properties: |
    enemy.types=aliens,monsters
    player.maximum-lives=5
  user-interface.properties: |
    color.good=purple
    color.bad=yellow
    allow.textmode=true
```

Pod 使用 ConfigMap 数据一般使用以下方式：

- 在容器命令和参数内直接引用
- 挂载为容器的环境变量
- 将 ConfigMap 挂载为只读卷文件，使用应用来读取
- 编写代码在 Pod 中运行，使用 Kubernetes API 来读取 ConfigMap

> 注：从 v1.19 开始，可以通过在 ConfigMap 定义中，添加一个 `immutable` 字段以创建 [不可变更的 ConfigMap](https://kubernetes.io/zh/docs/concepts/configuration/configmap/#configmap-immutable)

## 参考资料

1. [Kubernetes / 配置 / ConfigMap](https://kubernetes.io/zh/docs/concepts/configuration/configmap/)
2. [Kubernetes ConfigMap 字段说明](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#configmap-v1-core)
