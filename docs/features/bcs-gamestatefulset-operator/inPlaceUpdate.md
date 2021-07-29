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
# 创建gamestatefulset
$ kubectl apply -f doc/features/bcs-gamestatefulset-operator/example/inplace-update.yaml

# 检查 pod 状态
$ kubectl get pods
  NAME                     READY   STATUS    RESTARTS   AGE
  test-gamestatefulset-0   1/1     Running   0          81s
  test-gamestatefulset-1   1/1     Running   0          80s
  test-gamestatefulset-2   1/1     Running   0          79s
  test-gamestatefulset-3   1/1     Running   0          78s
  test-gamestatefulset-4   1/1     Running   0          77s

# 查看 gamestatefulset 状态
$ kubectl get gamestatefulset
  NAME                   REPLICAS   READYREPLICAS   CURRENTREPLICAS   UPDATEDREPLICAS   UPDATEDREADY_REPLICAS   AGE
  test-gamestatefulset   5          5               5                 5                 5                       4m1s

# 执行原地重启更新，灰度两个实例，gracePeriodSeconds 为 30 秒
$ kubectl patch gamestatefulset test-gamestatefulset --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"python:latest"}]'
  gamestatefulset.tkex.tencent.com/test-gamestatefulset patched
# 也可以在调整yaml文件之后 kubectl apply -f doc/features/bcs-gamestatefulset-operator/example/inplace-update.yaml


# 在 node 节点上查看容器，两个容器实例逐步完成了重启
$ docker ps | grep gamestatefulset | grep python
  39bf13f0397f   cba42c28d9b8                                                    "python -m http.serv…"   2 seconds ago        Up 1 second                   k8s_python_test-gamestatefulset-3_default_e65b2785-8805-4b5d-8e38-4285f1cc83a2_1
  630b6b45be48   python                                                          "python -m http.serv…"   About a minute ago   Up About a minute             k8s_python_test-gamestatefulset-4_default_30494560-ee37-4f80-a2ae-cbd911d30654_1
  b94970ddac46   3687eb5ea744                                                    "python -m http.serv…"   5 minutes ago        Up 5 minutes                  k8s_python_test-gamestatefulset-2_default_9bfc34a2-4183-43d5-8459-6ef13da0ff5b_0
  3508c1bf9161   3687eb5ea744                                                    "python -m http.serv…"   5 minutes ago        Up 5 minutes                  k8s_python_test-gamestatefulset-1_default_c3fe0700-47e8-469a-b5d6-bc3694ede3a8_0
  eed5920f21d9   3687eb5ea744                                                    "python -m http.serv…"   5 minutes ago        Up 5 minutes                  k8s_python_test-gamestatefulset-0_default_bc1201ee-b22f-4dde-bc73-74a8c9d3fb43_0

# 最后，查看 pod 状态，生命周期没变，2 个实例的 RESTARTS 次数为 1，因为这里partition为3，序号大于等于3的容器都会更新
$ kubectl get pods | grep gamestatefulset
  NAME                     READY   STATUS    RESTARTS   AGE
  test-gamestatefulset-0   1/1     Running   0          9m12s
  test-gamestatefulset-1   1/1     Running   0          9m11s
  test-gamestatefulset-2   1/1     Running   0          9m10s
  test-gamestatefulset-3   1/1     Running   1          9m9s
  test-gamestatefulset-4   1/1     Running   1          9m8s

# 查看 gamestatefulset 状态
$ kubectl get gamestatefulset
  NAME                   REPLICAS   READYREPLICAS   CURRENTREPLICAS   UPDATEDREPLICAS   UPDATEDREADY_REPLICAS   AGE
  test-gamestatefulset   5          5               3                 2                 2                       16m

# 若想进一步完成全部更新，把 partition 设为 0 后，重复上面的原地重启更新过程
```
