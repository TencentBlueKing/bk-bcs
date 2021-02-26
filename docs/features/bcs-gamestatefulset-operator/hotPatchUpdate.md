# HostPatchUpdate特性描述

镜像热更新 HotPatchUpdate 更新策略在更新过程中，保持 pod 及其容器的生命周期都不变，只是新容器的镜像版本。
更新完成后，用户需要通过向 pod 容器发送信号或命令，reload 或 重启 pod 容器中的进程，最终实现 pod 容器的更新。
该功能需要配合 bcs 定制的 kubelet 和 dockerd 版本才能使用。HotPatchUpdate 同样支持 partition 配置，用于实现灰度发布策略。

> 注意：**  
> HotPatchUpdate 需要结合 bcs 定制的 kubelet 和 dockerd 版本才能使用，直接用官方的 k8s 和 docker 版本不能生效。  

## 示例

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
