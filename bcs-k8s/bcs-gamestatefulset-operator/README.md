## gamestatefulset-operator

bcs-gamestatefulset-operator 是针对游戏 gameserver 实现的管理有状态应用的增强版 statefulset，基于原生 statefulset 改造，并参考了
openkruise 的部分实现，支持原地重启、镜像热更新、滚动更新等多种更新策略。

### 重构目标

* [done]本项目 group 重构可用
* [done]增加原地重启 InplaceUpdate 更新策略，并支持原地重启过程当中的 gracePeriodSeconds
* [done]增加镜像热更新 HotPatchUpdate 更新策略
* [done]增加自动并行滚动更新
* [done]支持HPA
* [todo]集成腾讯云CLB，实现有状态端口段动态转发
* [todo]集成BCS无损更新特性：允许不重启容器更新容器内容
* [todo]扩展kubectl，支持kubectl gamestatefulset子命令

### 特性

基于 CRD+Operator 开发的一种自定义的 K8S 工作负载（GameStatefulSet），核心特性包括：

* 兼容StatefulSet所有特性
* 支持Operator高可用部署
* 支持Node失联时，Pod的自动漂移（StatefulSet不支持）  
* 支持容器原地升级
* 支持容器镜像热更新
* 支持 HPA

### 特性介绍

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
InplaceUpdate 同样支持 partition 配置，用于实现灰度发布策略。为了兼容旧版本，InplaceUpdate 沿用 RollingUpdate 的 partition 配置字段：
spec/updateStrategy/rollingUpdate/partition 。

#### 镜像热更新 HotPatchUpdate 
镜像热更新 HotPatchUpdate 更新策略在更新过程中，保持 pod 及其容器的生命周期都不变，只是新容器的镜像版本。更新完成后，用户需要通过向
pod 容器发送信号或命令，reload 或 重启 pod 容器中的进程，最终实现 pod 容器的更新。  
该功能需要配合 bcs 定制的 kubelet 和 dockerd 版本才能使用。  
HotPatchUpdate 同样支持 partition 配置，用于实现灰度发布策略。为了兼容旧版本，HotPatchUpdate 沿用 RollingUpdate 的 partition 配置字段：
spec/updateStrategy/rollingUpdate/partition

### 信息初始化

初始化依赖信息，安装 gamestatefulset-operator

```shell
$ kubectl create -f doc/deploy/01-resources.yaml

$ kubectl create -f doc/deploy/02-namespace.yaml

$ kubectl create -f doc/deploy/03-rbac.yaml

$ kubectl create -f doc/deploy/04-operator-deployment.yaml
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

#### 原地重启 InplaceUpdate
**注意：**  
为了兼容老版本，InplaceUpdate 的 partition 配置沿用 spec/updateStrategy/rollingUpdate/partition 字段的值 

```shell
# 创建gamestatefulset
$ kubectl apply -f doc/example/inplace-update.yaml

# 检查 pod 状态
$ kubectl get pods
  NAME                              READY   STATUS             RESTARTS   AGE
  test-gamestatefulset-0            1/1     Running            0          13s
  test-gamestatefulset-1            1/1     Running            0          11s
  test-gamestatefulset-2            1/1     Running            0          9s
  test-gamestatefulset-3            1/1     Running            0          7s
  test-gamestatefulset-4            1/1     Running            0          4s

# 查看 gamestatefulset 状态
$ kubectl get gamestatefulset
  NAME                   REPLICAS   READYREPLICAS   CURRENTREPLICAS   UPDATEDREPLICAS   AGE
  test-gamestatefulset   5          5               5                 5                 59s

# 执行原地重启更新，灰度两个实例，gracePeriodSeconds 为 30 秒
$ kubectl patch gamestatefulset test-gamestatefulset --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"python:latest"}]'
  gamestatefulset.tkex.tencent.com/test-gamestatefulset patched
# 也可以在调整yaml文件之后 kubectl apply -f doc/example/inplace-update.yaml


# 大概 30 秒后，在 node 节点上查看容器。两个容器实例完成了重启
$ docker ps | grep gamestatefulset | grep python
  b45ac38d9339        32260605cf7a                                               "python -m http.serv…"   56 seconds ago      Up 55 seconds                           k8s_python_test-gamestatefulset-3_default_e8fb9553-a6f9-4a11-9151-29d5ff976f02_1
  676abb166253        32260605cf7a                                               "python -m http.serv…"   56 seconds ago      Up 55 seconds                           k8s_python_test-gamestatefulset-4_default_77b01194-c765-48d5-b7f3-0bd3b13d2668_1
  5146e4fa2f71        7f4efc85a56c                                               "python -m http.serv…"   5 minutes ago       Up 5 minutes                            k8s_python_test-gamestatefulset-2_default_98e79063-5ae8-4f51-b937-9b1932bfa95a_0
  6c6e7c52aa4b        7f4efc85a56c                                               "python -m http.serv…"   5 minutes ago       Up 5 minutes                            k8s_python_test-gamestatefulset-1_default_21c7e5ff-4737-44a9-9bce-2dab5a750db6_0
  31c18ccf777c        7f4efc85a56c                                               "python -m http.serv…"   5 minutes ago       Up 5 minutes                            k8s_python_test-gamestatefulset-0_default_e57f92d8-0d27-417e-8727-40e8a2a689e8_0

# 查看 pod 状态，生命周期没变，后面两个实例的 RESTARTS 次数为 1
$ kubectl get pods
  NAME                              READY   STATUS             RESTARTS   AGE
  test-gamestatefulset-0            1/1     Running            0          7m50s
  test-gamestatefulset-1            1/1     Running            0          7m48s
  test-gamestatefulset-2            1/1     Running            0          7m46s
  test-gamestatefulset-3            1/1     Running            1          7m44s
  test-gamestatefulset-4            1/1     Running            1          7m41s

# 查看 gamestatefulset 状态，CURRENTREPLICAS为3，UPDATEDREPLICAS 为2
$ kubectl get gamestatefulset
  NAME                   REPLICAS   READYREPLICAS   CURRENTREPLICAS   UPDATEDREPLICAS   AGE
  test-gamestatefulset   5          5               3                 2                 9m46s

# 若想进一步完成全部更新，把 partition 设为 0 后，重复上面的原地重启更新过程
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
bcs 将根据业务场景，持续增强 GameStatefulSet 的能力，增加更多发布场景下的更新策略，并与 service，ingress 等形成联动，敬请期待。