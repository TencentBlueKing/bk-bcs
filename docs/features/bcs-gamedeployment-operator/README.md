## gamedeployment-operator

bcs-gamedeployment-operator 是针对游戏 gameserver 实现的管理无状态应用的增强版 deployment，基于 k8s 原生 replicaset 改造，并参考了
openkruise 的部分实现，支持原地重启、镜像热更新、滚动更新、灰度发布等多种更新策略。  

### 功能迭代

* [done]增加滚动更新 RollingUpdate 更新策略
* [done]支持灰度发布
* [done]增加原地重启 InplaceUpdate 更新策略，并支持原地重启过程当中的 gracePeriodSeconds
* [done]增加镜像热更新 HotPatchUpdate 更新策略
* [done]支持HPA
* [done]支持分步骤自动化灰度发布，在灰度过程中加入 hook 校验
* [done]优雅地删除和更新应用实例 PreDeleteHook 
* [todo]扩展kubectl，支持kubectl gamedeployment 子命令

### 特性

基于 CRD+Operator 开发的一种自定义的 K8S 工作负载（GameDeployment），核心特性包括：

* 支持Operator高可用部署
* 支持滚动更新
* 支持设置 partition 灰度发布
* 支持容器原地升级
* 支持容器镜像热更新
* 支持 HPA
* 支持分步骤自动化灰度发布
* 优雅地删除和更新应用实例 PreDeleteHook 

### 特性介绍

基于游戏业务等相关场景，bcs 需要对 deployment 管理的无状态应用添加原地重启、镜像热更新、灰度发布等新的功能，但原生的 deployment 是通过控制多个
replicaset 来实现滚动更新的，由 replicaset 来控制 pod 实例的生命周期。如果在原生的 deployment 上进行定制改造，就无法直接控制 pod 的生命周期。  
因此，bcs 基于 k8s 原生的 replicaset-controller 开发了 bcs-gamedeployment-operator，实现了一种新的 k8s workload：GameDeployment。  
GameDeployment 的配置与原生的 Deployment 基本一致：  

```yaml
apiVersion: tkex.tencent.com/v1alpha1
kind: GameDeployment
metadata:
  name: test-gamedeployment
  labels:
    app: test-gamedeployment
spec:
  replicas: 5
  selector:
    matchLabels:
      app: test-gamedeployment
  template:
    metadata:
      labels:
        app: test-gamedeployment
    spec:
      containers:
      - name: python
        image: python:3.5
        imagePullPolicy: IfNotPresent
        command: ["python"]
        args: ["-m", "http.server", "8000" ]
        ports:
        - name: http
          containerPort: 8000
  preDeleteUpdateStrategy:
    hook:
      templateName: test
  updateStrategy:
    type: InplaceUpdate
    partition: 1
    maxUnavailable: 2
    canary:
      steps:
        - partition: 3
        - pause: {}
        - partition: 1
        - pause: {duration: 60}
        - hook:
            templateName: test
        - pause: {}
    inPlaceUpdateStrategy:
      gracePeriodSeconds: 30
```

* preDeleteUpdateStrategy  
见下文 "优雅地删除和更新应用实例 PreDeleteHook" 章节。  
* updateStrategy/type  
支持 RollingUpdate, InplaceUpdate, HotPatchUpdate 三种更新策略。  
* updateStrategy/partition  
用于实现灰度发布，可参考 statefulset 的 partition 。
* updateStrategy/maxUnavailable  
指在更新过程中每批执行更新的实例数量，在更新过程中这批实例是不可用的。比如一共有 8 个实例，maxUnavailable 设置为 2 ，那么每批滚动或原地重启 2 
个实例，等这 2 个实例更新完成后，再进行下一批更新。可设置为整数值或百分比，默认值为 25% 。
* updateStrategy/maxSurge  
在滚动更新过程中，如果每批都是先删除 maxUnavailable 数量的旧版本 pod 数，再新建新版本的 pod 数，那么在整个更新过程中，总共只有 replicas - maxUnavailable
数量的实例数能够提供服务。在总实例数 replicas 数量比较小的情况下，会影响应用的服务能力。设置 maxSurge 后，会在滚动更新前先多创建 maxSurge 数量的 pod，
然后再逐批进行更新，更新完成后，最后再删掉 maxSurge 数量的 pod ，这样就能保证整个更新过程中可服务的总实例数量。 maxSurge 默认值为 0 。  
因 InplaceUpdate 和 HotPatchUpdate 不会重启 pod ，因此建议只在 RollingUpdate 更新时设置 maxSurge 参数。  
* updateStrategy/inPlaceUpdateStrategy  
见下方 InplaceUpdate 介绍。  
* updateStrategy/canary  
智能式分步骤灰度发布，见下文详细介绍。  

#### 滚动发布 RollingUpdate
kubernetes 原生的 Deployment 通过控制多个 ReplicaSet 来实现应用的滚动发布，GameDeployment 通过直接控制应用 pod 实例的新增和删除来实现滚动
发布，详见下文使用案例。  

#### 原地重启 InplaceUpdate
原地重启更新策略在更新过程中，保持 pod 的生命周期不变，只是重启 pod 中的容器，可主要用于以下场景：  
* pod 中有多个容器，只想更新其中的一个容器，保持 pod 的 ipc 共享内存等不发生变化  
* 在更新过程中保持 pod 状态不变，不重新调度，仅仅重启和更新 pod 中的一个或多个容器，加快更新速度  

原地重启的速度可能会特别快，在容器重启过程中，service 来不及更新 endpoints ，亦即来不及把正在原地重启的 pod 从 endpoints 中剔除，这样会
导致在原地重启过程中的 service 流量受损。  
为了解决这一问题，我们在原地重启的策略中加入了 gracePeriodSeconds 的参数。  
假如在原地重启的更新策略下，配置了 spec/updateStrategy/inPlaceUpdateStrategy/gracePeriodSeconds 为 30 秒，那么 bcs-gamestatefulset-operator
在更新一个 pod 前，会先把这个 pod 设置为 unready 状态，30 秒过后才会真正去重启 pod 中的容器，那么在这 30 秒的时间内 k8s 会把该 pod 实例从
service 的 endpoints 中剔除。等原地重启完成后，bcs-gamestatefulset-operator 才会再把该 pod 设为 ready 状态，之后 k8s 就会重新把该 pod
实例加入到 endpoints 当中。这样，在整个原地重启过程中，能保证 service 流量的无损服务。  
gracePeriodSeconds 的默认值为 0 ，如果不设置，bcs-gamestatefulset-operator 会马上原地重启 pod 中的容器。  
InplaceUpdate 同样支持 partition 配置，用于实现灰度发布策略。  

#### 镜像热更新 HotPatchUpdate 
镜像热更新 HotPatchUpdate 更新策略在更新过程中，保持 pod 及其容器的生命周期都不变，只是新容器的镜像版本。更新完成后，用户需要通过向
pod 容器发送信号或命令，reload 或 重启 pod 容器中的进程，最终实现 pod 容器的更新。  
该功能需要配合 bcs 定制的 kubelet 和 dockerd 版本才能使用。  
HotPatchUpdate 同样支持 partition 配置，用于实现灰度发布策略。  

#### 智能式分步骤灰度发布
GameDeployment 支持智能化的分步骤灰度发布功能，允许用户在 GameDeployment 定义中配置多个灰度发布的步骤，这些步骤可以是 "灰度发布部分实例"、"永久暂停灰度发布"、
"暂停指定的时间段后再继续灰度发布"、"外部 Hook 调用以决定是否暂停灰度发布"，通过配置这些不同的灰度发布步骤，可以达到自动化的分步骤灰度发布能力，实现
灰度发布的智能控制。  
详见：[智能式分步骤灰度发布](features/canary/auto-canary-update.md)

#### 优雅地删除和更新应用实例 PreDeleteHook
在与腾讯的众多游戏业务的交流过程中，我们发现，许多业务场景下，在删除 pod 实例(比如缩容实例数、HPA)前或发布更新 pod 版本前，业务希望能够实现优雅的 pod 退出，
在删除前能够加入一些 hook 勾子，通过这些 hook 判断是否已经可以正常删除或更新 pod。如果 hook 返回 ok，那么就正常删除或更新 pod，如果返回不 ok，
那么就继续等待，直到 hook 返回 ok。  
发散来看，这其实并非游戏业务的独特需求，而是大多数不同类型业务的普遍需求。  
然而，原生的 kubernetes 只支持 pod 级别的 preStop 和 postStart ，远不能满足这种更精细化的 hook 需求。  
这两个 workload 层面实现了与 bcs-hook-operator 的联动，提供了 pod 删除或更新前的 PreDeleteHook 功能，实现了应用实例的优雅删除和更新。    
详见：[应用实例的优雅删除和更新 PreDeleteHook](features/preDeleteHook/pre-delete-hook.md)

### 信息初始化

初始化依赖信息，安装 gamedeployment-operator，helm chart信息位于bk-bcs/install/helm/bcs-gamesdeployment-operator

```shell
helm upgrade bcs-gamedeployment-operator helm/bcs-gamedeployment-operator -n bcs-system --install
```

### 使用案例

* 滚动升级
* 原地重启
* 镜像热更新

#### 滚动升级 RollingUpdate

```shell
# 创建gamedeployment
$ kubectl apply -f example/rolling-update.yaml

# 检查 pod 状态
$ kubectl get pods
  NAME                              READY   STATUS             RESTARTS   AGE
  test-gamedeployment-7hdm5         1/1     Running            0          18s
  test-gamedeployment-7k6ch         1/1     Running            0          18s
  test-gamedeployment-9nbnq         1/1     Running            0          18s
  test-gamedeployment-bnsgw         1/1     Running            0          18s
  test-gamedeployment-fstmn         1/1     Running            0          18s
  test-gamedeployment-gv5cg         1/1     Running            0          18s
  test-gamedeployment-jzjdd         1/1     Running            0          18s
  test-gamedeployment-tsg4g         1/1     Running            0          18s

# 查看 gamedeployment 状态
$ kubectl get gamedeployment
  NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
  test-gamedeployment   8         8         8               8       8       4m11s

# 执行滚动升级，灰度两个实例
$ kubectl patch gamedeployment test-gamedeployment --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"python:latest"}]'
  gamedeployment.tkex.tencent.com/test-gamedeployment patched
# 也可以在调整yaml文件之后 kubectl apply -f doc/example/rolling-update.yaml

# 一段时间后，检查 pod 状态，5 个实例完成更新
$ kubectl get pods
  NAME                              READY   STATUS             RESTARTS   AGE
  test-gamedeployment-5r2mt         1/1     Running            0          39s
  test-gamedeployment-9nbnq         1/1     Running            0          6m6s
  test-gamedeployment-bnsgw         1/1     Running            0          6m6s
  test-gamedeployment-f8mhw         1/1     Running            0          43s
  test-gamedeployment-hml6x         1/1     Running            0          42s
  test-gamedeployment-qmb7b         1/1     Running            0          43s
  test-gamedeployment-ssjxz         1/1     Running            0          42s
  test-gamedeployment-tsg4g         1/1     Running            0          6m6s

# 查看 gamedeployment 状态
$ kubectl get gamedeployment
  NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
  test-gamedeployment   8         5         5               8       8       6m16s

# 若想进一步完成全部更新，把 partition 设为 0 后，重复上面的滚动更新过程
```

#### 原地重启 InplaceUpdate

```shell
# 创建gamedeployment
$ kubectl apply -f example/inplace-update.yaml

# 检查 pod 状态
$ kubectl get pods
  NAME                              READY   STATUS             RESTARTS   AGE
  test-gamedeployment-49m5l         1/1     Running            0          9s
  test-gamedeployment-57rrt         1/1     Running            0          9s
  test-gamedeployment-7wr7h         1/1     Running            0          9s
  test-gamedeployment-cbk77         1/1     Running            0          9s
  test-gamedeployment-n58hm         1/1     Running            0          9s
  test-gamedeployment-n8ld6         1/1     Running            0          9s
  test-gamedeployment-r78df         1/1     Running            0          9s
  test-gamedeployment-wzxm7         1/1     Running            0          8s

# 查看 gamedeployment 状态
$ kubectl get gamedeployment
  NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
  test-gamedeployment   8         8         8               8       8       30s

# 执行原地重启更新，灰度两个实例，gracePeriodSeconds 为 30 秒
$ kubectl patch gamedeployment test-gamedeployment --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"python:latest"}]'
  gamedeployment.tkex.tencent.com/test-gamedeployment patched
# 也可以在调整yaml文件之后 kubectl apply -f doc/example/inplace-update.yaml


# 大概 30s 后，在 node 节点上查看容器。两个容器实例完成了重启
$ docker ps | grep gamedeployment | grep python
  2a5c3b3d7a1f        32260605cf7a                                               "python -m http.serv…"   34 seconds ago       Up 33 seconds                           k8s_python_test-gamedeployment-r78df_default_0cd04138-1925-4879-96d6-3c9cb157dede_1
  50f935da938a        32260605cf7a                                               "python -m http.serv…"   35 seconds ago       Up 34 seconds                           k8s_python_test-gamedeployment-wzxm7_default_96eeb87e-f647-4cfa-9831-a6d5f28ec10d_1
  63d1c8520fec        7f4efc85a56c                                               "python -m http.serv…"   3 minutes ago        Up 3 minutes                            k8s_python_test-gamedeployment-n8ld6_default_590a3a3b-eec6-4c1b-b5dd-40b0a49bdc89_0
  5041b2e48455        7f4efc85a56c                                               "python -m http.serv…"   3 minutes ago        Up 3 minutes                            k8s_python_test-gamedeployment-57rrt_default_5508a171-416a-4584-9e3b-dab308d9842e_0
  467c8160a339        7f4efc85a56c                                               "python -m http.serv…"   3 minutes ago        Up 3 minutes                            k8s_python_test-gamedeployment-n58hm_default_72c71d12-dc3d-4561-a952-f2e2d7a00776_0
  749fecc1f017        7f4efc85a56c                                               "python -m http.serv…"   3 minutes ago        Up 3 minutes                            k8s_python_test-gamedeployment-7wr7h_default_a84a6a20-b32f-430d-bc92-15c1cc9669e1_0
  3fcdd4e6b3fa        7f4efc85a56c                                               "python -m http.serv…"   3 minutes ago        Up 3 minutes                            k8s_python_test-gamedeployment-cbk77_default_50d14026-76c4-422f-a59f-e27c9d8c4a4e_0
  ed324fad792f        7f4efc85a56c                                               "python -m http.serv…"   3 minutes ago        Up 3 minutes                            k8s_python_test-gamedeployment-49m5l_default_304f3ec5-17fd-41da-b77d-2af0b72614f9_0

# 大概 1min 后，在 node 节点上查看容器。两个容器实例完成了重启
$ docker ps | grep gamedeployment | grep python
  196b04939a56        32260605cf7a                                               "python -m http.serv…"   2 seconds ago        Up 1 second                             k8s_python_test-gamedeployment-49m5l_default_304f3ec5-17fd-41da-b77d-2af0b72614f9_1
  43f52566166d        32260605cf7a                                               "python -m http.serv…"   4 seconds ago        Up 3 seconds                            k8s_python_test-gamedeployment-n8ld6_default_590a3a3b-eec6-4c1b-b5dd-40b0a49bdc89_1
  2a5c3b3d7a1f        32260605cf7a                                               "python -m http.serv…"   About a minute ago   Up About a minute                       k8s_python_test-gamedeployment-r78df_default_0cd04138-1925-4879-96d6-3c9cb157dede_1
  50f935da938a        32260605cf7a                                               "python -m http.serv…"   About a minute ago   Up About a minute                       k8s_python_test-gamedeployment-wzxm7_default_96eeb87e-f647-4cfa-9831-a6d5f28ec10d_1
  5041b2e48455        7f4efc85a56c                                               "python -m http.serv…"   4 minutes ago        Up 4 minutes                            k8s_python_test-gamedeployment-57rrt_default_5508a171-416a-4584-9e3b-dab308d9842e_0
  467c8160a339        7f4efc85a56c                                               "python -m http.serv…"   4 minutes ago        Up 4 minutes                            k8s_python_test-gamedeployment-n58hm_default_72c71d12-dc3d-4561-a952-f2e2d7a00776_0
  749fecc1f017        7f4efc85a56c                                               "python -m http.serv…"   4 minutes ago        Up 4 minutes                            k8s_python_test-gamedeployment-7wr7h_default_a84a6a20-b32f-430d-bc92-15c1cc9669e1_0
  3fcdd4e6b3fa        7f4efc85a56c                                               "python -m http.serv…"   4 minutes ago        Up 4 minutes                            k8s_python_test-gamedeployment-cbk77_default_50d14026-76c4-422f-a59f-e27c9d8c4a4e_0

# 最后，查看 pod 状态，生命周期没变，5 个实例的 RESTARTS 次数为 1
$ kubectl get pods | grep gamedeploy
  test-gamedeployment-49m5l         1/1     Running            1          8m6s
  test-gamedeployment-57rrt         1/1     Running            0          8m6s
  test-gamedeployment-7wr7h         1/1     Running            0          8m6s
  test-gamedeployment-cbk77         1/1     Running            0          8m6s
  test-gamedeployment-n58hm         1/1     Running            1          8m6s
  test-gamedeployment-n8ld6         1/1     Running            1          8m6s
  test-gamedeployment-r78df         1/1     Running            1          8m6s
  test-gamedeployment-wzxm7         1/1     Running            1          8m5s

# 查看 gamedeployment 状态
$ kubectl get gamedeployment
  NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
  test-gamedeployment   8         5         5               8       8       8m49s

# 若想进一步完成全部更新，把 partition 设为 0 后，重复上面的原地重启更新过程
```

#### 镜像热更新 HotPatchUpdate
**注意：**  
HotPatchUpdate 需要结合 bcs 定制的 kubelet 和 dockerd 版本才能使用，直接用官方的 k8s 和 docker 版本不能生效。  

```shell
# 创建 gamedeployment
$ kubectl apply -f example/hotpatch-update.yaml

# 检查 pod 状态
$ kubectl get pods
  NAME                              READY   STATUS             RESTARTS   AGE
  test-gamedeployment-4z5ff         1/1     Running            0          20s
  test-gamedeployment-8ncs6         1/1     Running            0          20s
  test-gamedeployment-ftrnm         1/1     Running            0          20s
  test-gamedeployment-klxlp         1/1     Running            0          20s
  test-gamedeployment-mghr5         1/1     Running            0          20s
  test-gamedeployment-mtqn6         1/1     Running            0          20s
  test-gamedeployment-rncjv         1/1     Running            0          20s
  test-gamedeployment-skh2m         1/1     Running            0          20s

# 查看 gamedeployment 状态
$ kubectl get gamedeployment
  NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
  test-gamedeployment   8         8         8               8       8       44s

# 执行 HotPatchUpdate 镜像热更新更新，灰度两个实例
$ kubectl patch gamedeployment test-gamedeployment --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"main:v2"}]'
  gamedeployment.tkex.tencent.com/test-gamedeployment patched
# 也可以在调整yaml文件之后 kubectl apply -f doc/example/hotpatch-update.yaml

# 在节点上查看容器状态，容器的生命周期没有变化，但有 5 个实例的镜像版本已经发生变化
$ docker ps | grep gamedeployment | grep main
  f1ee64fca56f        7c3c2d7947cb                                               "/main"                  2 minutes ago       Up 2 minutes                            k8s_main_test-gamedeployment-skh2m_default_ad3feb3f-48f5-419c-9dc9-89295447247e_0
  8db2c9d576bf        main:v2                                                    "/main"                  2 minutes ago       Up 2 minutes                            k8s_main_test-gamedeployment-rncjv_default_010745a3-54e2-414a-838f-a7646ec4c41d_0
  9a589cb5ff62        main:v2                                                    "/main"                  2 minutes ago       Up 2 minutes                            k8s_main_test-gamedeployment-mtqn6_default_414d6cae-5580-4ffc-8083-bc412556c7bb_0
  a664faa56f81        7c3c2d7947cb                                               "/main"                  2 minutes ago       Up 2 minutes                            k8s_main_test-gamedeployment-8ncs6_default_9525d936-7a42-4383-8431-30b176b3fbd6_0
  e57a8d0d5a5e        main:v2                                                    "/main"                  2 minutes ago       Up 2 minutes                            k8s_main_test-gamedeployment-mghr5_default_a1b46de8-ff5c-4eb8-9563-67aa8d31a5ff_0
  97746655e43a        main:v2                                                    "/main"                  2 minutes ago       Up 2 minutes                            k8s_main_test-gamedeployment-klxlp_default_09e73dcb-ad46-408d-ba82-68212cb8d4f2_0
  475a3d103491        main:v2                                                    "/main"                  2 minutes ago       Up 2 minutes                            k8s_main_test-gamedeployment-ftrnm_default_2582280c-32cd-46a3-9257-d5e9727e8f49_0
  986526c84f38        7c3c2d7947cb                                               "/main"                  2 minutes ago       Up 2 minutes                            k8s_main_test-gamedeployment-4z5ff_default_9d52180d-9dd7-4dcb-a771-34205b29ea45_0

# 查看 pod 状态，生命周期没变, RESTARTS 也没有增加
$ kubectl get pods
  NAME                              READY   STATUS             RESTARTS   AGE
  test-gamedeployment-4z5ff         1/1     Running            0          3m18s
  test-gamedeployment-8ncs6         1/1     Running            0          3m18s
  test-gamedeployment-ftrnm         1/1     Running            0          3m18s
  test-gamedeployment-klxlp         1/1     Running            0          3m18s
  test-gamedeployment-mghr5         1/1     Running            0          3m18s
  test-gamedeployment-mtqn6         1/1     Running            0          3m18s
  test-gamedeployment-rncjv         1/1     Running            0          3m18s
  test-gamedeployment-skh2m         1/1     Running            0          3m18s

# 查看 gamedeployment 状态
$ kubectl get gamedeployment
  NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
  test-gamedeployment   8         5         5               8       8       3m44s

# describe gamedeployment 状态，查看 event 事件
$ kubectl describe gamedeployment test-gamedeployment
……
Events:
  Type    Reason                       Age    From                         Message
  ----    ------                       ----   ----                         -------
  Normal  SuccessfulCreate             4m13s  bcs-gamedeployment-operator  succeed to create pod test-gamedeployment-skh2m
  Normal  SuccessfulCreate             4m13s  bcs-gamedeployment-operator  succeed to create pod test-gamedeployment-ftrnm
  Normal  SuccessfulCreate             4m13s  bcs-gamedeployment-operator  succeed to create pod test-gamedeployment-8ncs6
  Normal  SuccessfulCreate             4m13s  bcs-gamedeployment-operator  succeed to create pod test-gamedeployment-mghr5
  Normal  SuccessfulCreate             4m13s  bcs-gamedeployment-operator  succeed to create pod test-gamedeployment-klxlp
  Normal  SuccessfulCreate             4m13s  bcs-gamedeployment-operator  succeed to create pod test-gamedeployment-mtqn6
  Normal  SuccessfulCreate             4m13s  bcs-gamedeployment-operator  succeed to create pod test-gamedeployment-rncjv
  Normal  SuccessfulCreate             4m13s  bcs-gamedeployment-operator  succeed to create pod test-gamedeployment-4z5ff
  Normal  SuccessfulUpdatePodHotPatch  2m35s  bcs-gamedeployment-operator  successfully update pod test-gamedeployment-klxlp hot-patch
  Normal  SuccessfulUpdatePodHotPatch  2m35s  bcs-gamedeployment-operator  successfully update pod test-gamedeployment-mghr5 hot-patch
  Normal  SuccessfulUpdatePodHotPatch  2m35s  bcs-gamedeployment-operator  successfully update pod test-gamedeployment-mtqn6 hot-patch
  Normal  SuccessfulUpdatePodHotPatch  2m35s  bcs-gamedeployment-operator  successfully update pod test-gamedeployment-ftrnm hot-patch
  Normal  SuccessfulUpdatePodHotPatch  2m34s  bcs-gamedeployment-operator  successfully update pod test-gamedeployment-rncjv hot-patch

# 若想进一步完成全部更新，把 partition 设为 0 后，重复上面的镜像热更新过程
```

### 后续规划

bcs 将根据业务场景，持续增强 GameDeployment 的能力，增加更多发布场景下的更新策略，并与 service，ingress 等形成联动。
