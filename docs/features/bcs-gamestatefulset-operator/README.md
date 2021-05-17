## gamestatefulset-operator

bcs-gamestatefulset-operator 是针对游戏 gameserver 实现的管理有状态应用的增强版 statefulset，基于原生statefulset改造，
支持原地重启、镜像热更新、滚动更新等多种更新策略。

### 功能迭代

* [done]增加原地重启 InplaceUpdate 更新策略，并支持原地重启过程当中的 gracePeriodSeconds
* [done]增加镜像热更新 HotPatchUpdate 更新策略
* [done]支持并行滚动更新
* [done]支持HPA
* [done]集成腾讯云CLB，实现有状态端口段动态转发
* [done]支持分步骤自动化灰度发布，在灰度过程中加入 hook 校验
* [done]优雅地删除和更新应用实例 PreDeleteHook 
* [done]强制删除NodeLost上被主动驱逐的Terminating状态pod，促使Pod快速重建
* [todo]扩展 kubectl，支持 kubectl gamestatefulset子命令

### 特性

基于 CRD+Operator 开发的一种自定义的 K8S 工作负载（GameStatefulSet），核心特性包括：

* 兼容 StatefulSet 所有特性
* 支持 Operator 高可用部署
* 支持Node失联时，Pod的自动漂移（StatefulSet不支持）  
* 支持容器原地升级
* 支持容器镜像热更新
* 支持自动并行滚动更新(StatefulSet 只支持按序滚动更新)
* 支持 HPA
* 支持分步骤自动化灰度发布
* 优雅地删除和更新应用实例 PreDeleteHook

### 特性介绍

GameStatefulSet 是 kubernetes 原生 StatefulSet 的增强实现，一个典型的 GameStatefulSet 定义如下：  

```yaml
apiVersion: tkex.tencent.com/v1alpha1
kind: GameStatefulSet
metadata:
  name: test-gamestatefulset
spec:
  serviceName: "test"
  podManagementPolicy: Parallel
  replicas: 5
  selector:
    matchLabels:
      app: test
  preDeleteUpdateStrategy:
    hook:
      templateName: test
  updateStrategy:
    type: InplaceUpdate
    rollingUpdate:
      partition: 1
    inPlaceUpdateStrategy:
      gracePeriodSeconds: 30
    canary:
      steps:
      - partition: 3
      - pause: {}
      - partition: 1
      - pause: {duration: 60}
      - hook:
          templateName: test
      - pause: {}
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: python
        image: python:latest
        imagePullPolicy: IfNotPresent
        command: ["python"]
        args: ["-m", "http.server", "8000" ]
        ports:
        - name: http
          containerPort: 8000
```

* podManagementPolicy  
支持 "OrderedReady" 和 "Parallel" 两种方式，定义和 StatefulSet 一致，默认为 OrderedReady。与 StatefulSet 不同的是，如果配置为 Parallel，
那么不仅删除和创建 pod 实例是并行的，实例更新也是并行的，即自动并行更新。
* preDeleteUpdateStrategy  
见下文 "优雅地删除和更新应用实例 PreDeleteHook" 章节。  
* updateStrategy/type  
支持 RollingUpdate, OnDelete, InplaceUpdate, HotPatchUpdate 四种更新方式，相比原生 StatefulSet,新增 InplaceUpdate, HotPatchUpdate
两种更新模式。  
* updateStrategy/rollingUpdate/partition  
控制灰度发布的个数，与 StatefulSet 含义一致。为了兼容，InplaceUpdate 和 HotPatchUpdate 的灰度发布个数也由这个参数配置。  
* updateStrategy/inPlaceUpdateStrategy  
见下方 InplaceUpdate 介绍。  
* updateStrategy/canary  
智能式分步骤灰度发布，见下文详细介绍。  

#### 滚动发布 RollingUpdate

同 StatefulSet

#### 手动更新 OnDelete

同 StatefulSet

#### 原地重启 InplaceUpdate

原地重启更新策略在更新过程中，保持 pod 的生命周期不变，只是重启 pod 中的容器，可主要用于以下场景：  

* pod 中有多个容器，只想更新其中的一个容器，保持 pod 的 ipc 共享内存等不发生变化
* 在更新过程中保持 pod 状态不变，不重新调度，仅仅重启和更新 pod 中的一个或多个容器，加快更新速度  

原地重启的速度可能会特别快，在容器重启过程中，service 来不及更新 endpoints ，亦即来不及把正在原地重启的 pod 从 
endpoints 中剔除，这样会导致在原地重启过程中的 service 流量受损。 为了解决这一问题，我们在原地重启的策略中加入了 
gracePeriodSeconds 的参数。  假如在原地重启的更新策略下，配置了 spec/updateStrategy/inPlaceUpdateStrategy/
gracePeriodSeconds 为 30 秒，那么 bcs-gamestatefulset-operator在更新一个 pod 前，会先把这个 pod 设置为 unready 
状态，30 秒过后才会真正去重启 pod 中的容器，那么在这 30 秒的时间内 k8s 会把该 pod 实例从service 的 endpoints 中剔除。
等原地重启完成后，bcs-gamestatefulset-operator 才会再把该 pod 设为 ready 状态，之后 k8s 就会重新把该 pod实例加入到 
endpoints 当中。这样，在整个原地重启过程中，能保证 service 流量的无损服务。  gracePeriodSeconds 的默认值为 0 ，如果
不设置，bcs-gamestatefulset-operator 会马上原地重启 pod 中的容器。  InplaceUpdate 同样支持 partition 配置，用于
实现灰度发布策略。为了兼容旧版本，InplaceUpdate 沿用 RollingUpdate 的 partition 配置字段：
spec/updateStrategy/rollingUpdate/partition。

操作范例请参照[inPlaceUpdate.md](./inPlaceUpdate.md)

#### 镜像热更新 HotPatchUpdate 

镜像热更新 HotPatchUpdate 更新策略在更新过程中，保持 pod 及其容器的生命周期都不变，只是新容器的镜像版本。更新完成后，
用户需要通过向pod 容器发送信号或命令，reload 或 重启 pod 容器中的进程，最终实现 pod 容器的更新。  该功能需要配合 bcs 
定制的 kubelet 和 dockerd 版本才能使用。  HotPatchUpdate 同样支持 partition 配置，用于实现灰度发布策略。为了兼容
旧版本，HotPatchUpdate 沿用 RollingUpdate 的 partition 配置字段：spec/updateStrategy/rollingUpdate/partition。

操作范例请参照[hotPatchUpdate.md](./hotPatchUpdate.md)

#### 智能式分步骤灰度发布

GameStatefulSet 支持智能化的分步骤灰度发布功能，允许用户在 GameStatefulSet 定义中配置多个灰度发布的步骤，这些步骤可以
是 "灰度发布部分实例"、"永久暂停灰度发布"、"暂停指定的时间段后再继续灰度发布"、"外部 Hook 调用以决定是否暂停灰度发布"，
通过配置这些不同的灰度发布步骤，可以达到自动化的分步骤灰度发布能力，实现分批灰度发布的智能控制。  
GameStatefulSet 的智能式分步骤灰度发布的使用与 GameDeployment 一致，详见：[智能式分步骤灰度发布auto-canary-update.md](../bcs-gamedeployment-operator/features/canary/auto-canary-update.md)

#### 强制删除NodeLost节点中被主动驱逐的Terminating状态pod
支持对 NodeLost 节点中GameStatefulSet 下 被主动驱逐（处于Terminating状态）的 Pod 进行强制删除 (删除Etcd中该资源)，促使Pod快速重建，降低业务损失时间。
在GameStatefulSet的资源定义中, 如果 spec.template.metadata.annotations 存在 pod.gamestatefulset.bkbcs.tencent.
com/node-lost-force-delete: "true" 时，则执行强制删除策略；否则，保持与原生 StatefulSet 一致的策略，即 node lost 
后，即使主动驱逐了该Pod，但因kubelet工作异常，不会删除成功、Etcd 中也不会删除该 Pod，进而新的"替代" Pod 也不会被创建。

#### 优雅地删除和更新应用实例 PreDeleteHook

在与腾讯的众多游戏业务的交流过程中，我们发现，许多业务场景下，在删除 pod 实例(比如缩容实例数、HPA)前或发布更新 pod 版本前，
业务希望能够实现优雅的 pod 退出，在删除前能够加入一些 hook 勾子，通过这些 hook 判断是否已经可以正常删除或更新 pod。如果
hook 返回 ok，那么就正常删除或更新 pod，如果返回不 ok，那么就继续等待，直到 hook 返回 ok。发散来看，这其实并非游戏业务
的独特需求，而是大多数不同类型业务的普遍需求。然而，原生的 kubernetes 只支持 pod 级别的 preStop 和 postStart ，远不
能满足这种更精细化的 hook 需求。BCS 团队结合业务需求，抽象并开发了 bcs-hook-operator ，用于实现多种形式的 hook 控制，
在 bcs-gamedeployment-operator 和 game-statefulset-operator这两个 workload 层面实现了与 bcs-hook-operator 
的联动，提供了 pod 删除或更新前的 PreDeleteHook 功能，实现了应用实例的优雅删除和更新。GameStatefulSet 的 PreDeleteHook 
的使用与 GameDeployment 一致，详见：[应用实例的优雅删除和更新 PreDeleteHook](../bcs-gamedeployment-operator/features/preDeleteHook/pre-delete-hook.md)

### 信息初始化

初始化依赖信息，安装 gamestatefulset-operator，helm chart信息位于bk-bcs/install/helm/bcs-gamestatefulset-operator

```shell
helm upgrade bcs-gamestatefulset-operator helm/bcs-gamestatefulset-operator -n bcs-system --install
```

### 使用案例

* 扩缩容
* 滚动升级
* 原地重启
* 镜像热更新

#### 扩缩容

```shell
# 创建gamestatefulset
$ kubectl create -f doc/example/gamestatefulset-sample.yml

# check pod status
$ kubectl get pod -n test | grep web 
web-0                              1/1     Running   0         21s

# 执行扩缩容
$ kubectl scale --replicas=3 gamestatefulset/web -n test 

# check pod status
$ kubectl get pod -n test | grep web 
web-0   1/1     Running   0          2m
web-1   1/1     Running   0          13s
web-2   1/1     Running   0          10s
```

#### 滚动升级 RollingUpdate

```shell
# 创建gamestatefulset
$ kubectl apply -f doc/example/rolling-update.yaml

# 检查 pod 状态
$ kubectl get pods
  NAME                              READY   STATUS             RESTARTS   AGE
  test-gamestatefulset-0            1/1     Running            0          22s
  test-gamestatefulset-1            1/1     Running            0          20s
  test-gamestatefulset-2            1/1     Running            0          18s
  test-gamestatefulset-3            1/1     Running            0          15s
  test-gamestatefulset-4            1/1     Running            0          13s

# 查看 gamestatefulset 状态
$ kubectl get gamestatefulset
NAME                   REPLICAS   READYREPLICAS   CURRENTREPLICAS   UPDATEDREPLICAS   AGE
test-gamestatefulset   5          5               5                 5                 4m2s

# 执行灰度滚动升级，灰度两个实例
$ kubectl patch gamestatefulset test-gamestatefulset --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"python:latest"}]'
gamestatefulset.tkex.tencent.com/test-gamestatefulset patched
# 也可以在调整yaml文件之后 kubectl apply -f doc/example/rolling-update.yaml

# 检查 pod 状态，后面两个实例完成更新
$ kubectl get pods
  NAME                              READY   STATUS             RESTARTS   AGE
  test-gamestatefulset-0            1/1     Running            0          8m59s
  test-gamestatefulset-1            1/1     Running            0          8m57s
  test-gamestatefulset-2            1/1     Running            0          8m55s
  test-gamestatefulset-3            1/1     Running            0          75s
  test-gamestatefulset-4            1/1     Running            0          108s

# 查看 gamestatefulset 状态，CURRENTREPLICAS为3，UPDATEDREPLICAS 为2
$ kubectl get gamestatefulset
  NAME                   REPLICAS   READYREPLICAS   CURRENTREPLICAS   UPDATEDREPLICAS   AGE
  test-gamestatefulset   5          5               3                 2                 9m40s

# 若想进一步完成全部更新，把 partition 设为 0 后，重复上面的滚动更新过程
```

#### 镜像热更新 HotPatchUpdate

**注意：**  
为了兼容老版本，HotPatchUpdate 的 partition 配置沿用 spec/updateStrategy/rollingUpdate/partition 字段的值。  
HotPatchUpdate 需要结合 bcs 定制的 kubelet 和 dockerd 版本才能使用，直接用官方的 k8s 和 docker 版本不能生效。  

```shell
# 创建gamestatefulset
$ kubectl apply -f doc/example/hotpatch-update.yaml

# 检查 pod 状态
$ kubectl get pods
  NAME                              READY   STATUS             RESTARTS   AGE
  test-gamestatefulset-0            1/1     Running            0          33s
  test-gamestatefulset-1            1/1     Running            0          31s
  test-gamestatefulset-2            1/1     Running            0          29s
  test-gamestatefulset-3            1/1     Running            0          27s
  test-gamestatefulset-4            1/1     Running            0          24s

# 查看 gamestatefulset 状态
$ kubectl get gamestatefulset
  NAME                   REPLICAS   READYREPLICAS   CURRENTREPLICAS   UPDATEDREPLICAS   AGE
  test-gamestatefulset   5          5               5                 5                 59s

# 执行 HotPatchUpdate 镜像热更新更新，灰度两个实例
$ kubectl patch gamestatefulset test-gamestatefulset --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"main:v2"}]'
  gamestatefulset.tkex.tencent.com/test-gamestatefulset patched
# 也可以在调整yaml文件之后 kubectl apply -f doc/example/hotpatch-update.yaml

# 在节点上查看容器状态，容器的生命周期没有变化，但有两个实例的镜像版本已经发生变化
$ docker ps | grep gamestatefulset | grep main
  20425c7de46b        main:v2                                                    "/main"                  4 minutes ago        Up 4 minutes                            k8s_main_test-gamestatefulset-4_default_28988f05-ac0c-492e-94db-157ee1b49cfc_0
  b91a0c3a008c        main:v2                                                    "/main"                  4 minutes ago        Up 4 minutes                            k8s_main_test-gamestatefulset-3_default_58763289-15da-40aa-9f2d-2fdcb6386e2d_0
  2e3a362769e1        7c3c2d7947cb                                               "/main"                  4 minutes ago        Up 4 minutes                            k8s_main_test-gamestatefulset-2_default_48257245-1160-499a-81bd-e9dd269e142c_0
  13fd34bc84a5        7c3c2d7947cb                                               "/main"                  4 minutes ago        Up 4 minutes                            k8s_main_test-gamestatefulset-1_default_00052779-4ab8-4668-a0af-ca50ea25e2ca_0
  b683c08dc285        7c3c2d7947cb                                               "/main"                  4 minutes ago        Up 4 minutes                            k8s_main_test-gamestatefulset-0_default_eb2ebbb2-91a4-42fd-a1f8-0e9284b83dca_0

# 查看 pod 状态，生命周期没变, RESTARTS 也没有增加
$ kubectl get pods
  NAME                              READY   STATUS             RESTARTS   AGE
  test-gamestatefulset-0            1/1     Running            0          7m6s
  test-gamestatefulset-1            1/1     Running            0          7m4s
  test-gamestatefulset-2            1/1     Running            0          7m2s
  test-gamestatefulset-3            1/1     Running            0          7m
  test-gamestatefulset-4            1/1     Running            0          6m58s

# 查看 gamestatefulset 状态, CURRENTREPLICAS为3，UPDATEDREPLICAS 为2
# kubectl get gamestatefulset
 NAME                   REPLICAS   READYREPLICAS   CURRENTREPLICAS   UPDATEDREPLICAS   AGE
 test-gamestatefulset   5          5               3                 2                 8m19s

# 若想进一步完成全部更新，把 partition 设为 0 后，重复上面的镜像热更新过程
```

### 后续规划

bcs 将根据业务场景，持续增强 GameStatefulSet 的能力，增加更多发布场景下的更新策略，并与 service，ingress 等形成联动。