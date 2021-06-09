# InplaceUpdate特性描述

原地升级更新策略在更新过程中，保持 pod 的生命周期不变，只是重启 pod 中的容器，可主要用于以下场景：  

* pod 中有多个容器，只想更新其中的一个容器，保持 pod 的 ipc 共享内存等不发生变化  
* 在更新过程中保持 pod 状态不变，不重新调度，仅仅重启和更新 pod 中的一个或多个容器，加快更新速度

原地升级根据Pod的更新生命周期，提供了以下特性：

* 优雅更新时间: gracePeriodSeconds
* 灰度发布策略: partition
* 优雅更新Hook: PreInplaceUpdateStrategy

## 优雅更新时间

原地重启的速度可能会特别快，在容器重启过程中，service 来不及更新 endpoints ，亦即来不及把正在原地重启的 pod 从 endpoints 中剔除，这样会导致在原地重启过程中的 service 流量受损。  为了解决这一问题，我们在原地重启的策略中加入了 gracePeriodSeconds 的参数。  

假如在原地重启的更新策略下，配置了 spec/updateStrategy/inPlaceUpdateStrategy/gracePeriodSeconds 为 30 秒，那么 bcs-gamestatefulset-operator 在更新一个 pod 前，会先把这个 pod 设置为 unready 状态，30 秒过后才会真正去重启 pod 中的容器，那么在这 30 秒的时间内 k8s 会把该 pod 实例从 service 的 endpoints 中剔除。等原地重启完成后，bcs-gamestatefulset-operator 才会再把该 pod 设为 ready 状态，之后 k8s 就会重新把该 pod 实例加入到 endpoints 当中。这样，在整个原地重启过程中，能保证 service 流量的无损服务。  

gracePeriodSeconds 的默认值为 0 ，如果不设置，bcs-gamestatefulset-operator 会马上原地重启 pod 中的容器。  

## 灰度发布策略 与 优雅更新Hook

InplaceUpdate 同样支持 partition 配置，用于实现灰度发布策略，详见 [Feature - 自动化分步骤灰度发布](./自动化分步骤灰度发布.md)。

InplaceUpdate 支持 PreInplaceUpdate 配置，用于实现更新前的Hook确认，详见 [Feature - PreDeleteHook PreInplaceHook优雅删除和更新Pod](./PreDeleteHook%20PreInplaceHook优雅删除和更新Pod.md)  

## 示例

```shell
# 创建gamedeployment
$ kubectl apply -f doc/example/inplace-update.yaml

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
