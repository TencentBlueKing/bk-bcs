# PreDeleteHook & PreInplaceHook特性描述

在与腾讯的众多游戏业务的交流过程中，我们发现，许多业务场景下，在删除 pod 实例(比如缩容实例数、HPA)前或发布更新 pod 版本前，业务希望能够实现优雅的 pod 退出，在删除前能够加入一些 hook 勾子，通过这些 hook 判断是否已经可以正常删除或更新 pod。如果 hook 返回 ok，那么就正常删除或更新 pod，如果返回不 ok，那么就继续等待，直到 hook 返回 ok。  发散来看，这其实并非游戏业务的独特需求，而是大多数不同类型业务的普遍需求。  然而，原生的 kubernetes 只支持 pod 级别的 preStop 和 postStart ，远不能满足这种更精细化的 hook 需求。

BCS 团队结合业务需求，在 bcs-gamedeployment-operator 中通过 GameDeployment 和 HookRun 两个 controller，在 GameDeployment 这个 kubernetes workload 层面提供了 pod 删除或更新前的 PreDeleteUpdateHook/PreInplaceUpdateHook 功能，实现了应用实例的优雅删除和更新。 

## 实现原理

> bcs-gamedeployment-operator 通过与 bcs-hook-operator 的联动来实现 PreDeleteUpdateHook/PreInplaceUpdateHook。如果想要配置 PreDeleteUpdateHook/PreInplaceUpdateHook，需要同时部署 bcs-gamedeployment-operator 和 bcs-hook-operator 这两个 bcs 组件。

bcs-hook-operator 定义了 HookRun 和 HookTemplate 两个 CRD，其原理及实现请参考：[bcs-hook-operator](../../../bcs-hook-operator/README.md) 。
在 GameDeployment 应用中配置了 preDeleteUpdateStrategy.hook/preInplaceUpdateStrategy.hook 后，
bcs-gamedeployment-operator 在删除或更新一个 pod 前，会根据用户配置的 HookTemplate 创建一个 HookRun, 
bcs-hook-operator watch 到 HookRun crd 后就会操作这个 HookRun 进行 hook 调用，并维护这个 HookRun 的状态。

bcs-gamedeployment-operator watch 这个HookRun的状态，如果HookRun的运行结果符合预期，那bcs-gamedeployment-operator 
会正常地删除或更新这个 pod，如果不符合预期，那么就拒绝删除或更新 pod，并在 status 中记录状态。

以下以PreDeleteHook为例：

如果想要配置 GameDeployment 的 PreDeleteHook 功能，用户需先定义并创建 HookTemplate，然后在 GameDeployment 中配置 PreDeleteHook:  

```yaml
apiVersion: tkex.tencent.com/v1alpha1
kind: GameDeployment
metadata:
  name: test-gamedeployment
  labels:
    app: test-gamedeployment
spec:
  replicas: 8
  ......
  preDeleteUpdateStrategy:
    hook:
      templateName: test
  updateStrategy:
    type: InplaceUpdate
    #partition: 1
    maxUnavailable: 2
```

**注意：**   
考虑到 HotPatch 并不会实际销毁 pod 容器，所以当前配置了 PreDeleteHook 后，只支持 pod删除以及 RollingUpdate前的 hook ，HotPatchUpdate 前不会触发 PreDeleteHook，后续考虑把其做成可配置化。（配置了 preInplaceUpdateStrategy.hook 后，只支持pod原地更新前的 hook）  

## 使用示例

### 前置条件

在集群外运行一个示例的 webserver ：

```
$ ./gamedeployment-canary
$ curl http://1.1.1.1:9091
{"name":"bryan","male":"yes","age":45}
```

### 符合预期的 Hook

#### 创建 HookTemplate 模板和 GameDeployment 应用

定义一个 HookTemplate: [hook-template-success.yaml](./hook-template-success.yaml), hook 类型为 Webhook。  
定义一个 GameDeployment: [gamedeployment.yaml](./gamedeployment.yaml), 更新策略为 InplaceUpdate，并配置了 PreDeleteHook 的策略。  

创建 Hook 模板和应用：  

```shell
# 创建 hook 模板
$ kubectl apply -f hook-template-success.yaml
hooktemplate.tkex.tencent.com/test created

# 首次创建应用
$ kubectl apply -f gamedeployment.yaml
gamedeployment.tkex.tencent.com/test-gamedeployment created

$ kubectl get gamedeployment
NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
test-gamedeployment   6         6         6               6       6       73s
$ 
```

#### 缩容 GameDeployment

scale 缩容应用，operator 会创建 HookRun, GameDeployment 并未马上删除实例。  

```shell
# scale GameDeployment，删除一个实例
$ kubectl scale --replicas=5 gamedeployment/test-gamedeployment
  gamedeployment.tkex.tencent.com/test-gamedeployment scaled

# 查看 HookRun 资源, 已经创建一个 HookRun
$ kubectl get hookrun
  NAME                                        PHASE     AGE
  pre-delete-hook-test-gamedeployment-67864c6f65-s9rbp-test   Running   10s

# 确认 GameDeployment 状态，并未马上缩容
$ kubectl get gamedeployment
  NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
  test-gamedeployment   5         6         6               6       6       13m
$ kubectl get pod | grep gamedeployment
test-gamedeployment-2nrrt         1/1     Running   0          13m
test-gamedeployment-mw5nw         1/1     Running   0          13m
test-gamedeployment-pqrw5         1/1     Running   0          13m
test-gamedeployment-q78jd         1/1     Running   0          13m
test-gamedeployment-s9rbp         1/1     Running   0          13m
test-gamedeployment-xv26j         1/1     Running   0          13m
```

一段时间后，HookRun 运行完成，再次确认状态。因为 HookRun 运行成功，完成缩容：  

```shell
$ kubectl get gamedeployment
  NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
  test-gamedeployment   5         5         5               5       5       15m
$ kubectl get pod | grep gamedeployment
  test-gamedeployment-2nrrt         1/1     Running       0          15m
  test-gamedeployment-mw5nw         1/1     Running       0          15m
  test-gamedeployment-pqrw5         1/1     Running       0          15m
  test-gamedeployment-q78jd         1/1     Running       0          15m
  test-gamedeployment-s9rbp         1/1     Terminating   0          15m
  test-gamedeployment-xv26j         1/1     Running       0          15m

# pod 被删除后，其对应的 HookRun 也被删除
$ kubectl get hookrun
  No resources found in default namespace.
```

#### 更新 GameDeployment

原地更新 GameDeployment。因 partition 为 2，一共会更新 3 个实例。  
operator 会先创建 HookRun, 并不会马上进行实际更新操作。

```
$ kubectl patch gamedeployment test-gamedeployment --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"python:3.5"}]'
  gamedeployment.tkex.tencent.com/test-gamedeployment patched

# 查看 HookRun 资源。因为 maxUnavailable 配置为 2，每批只更新两个实例，故而已经创建 2 个 HookRun
$ kubectl get hookrun
  NAME                                        PHASE     AGE
  pre-delete-hook-test-gamedeployment-67864c6f65-q78jd-test   Running   5s
  pre-delete-hook-test-gamedeployment-67864c6f65-xv26j-test   Running   5s

# 确认 GameDeployment 状态，并未马上进行原地更新
$ kubectl get gamedeployment
NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
test-gamedeployment   5         0         0               5       5       28m
```

一段时间后，首批 2 个 pod 完成原地更新：  

```
$ kubectl get gamedeployment
NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
test-gamedeployment   5         2         2               5       5       32m

# 已经删除已更新的 2 个 pod 的 HookRun
$ kubectl get hookrun
  No resources found in default namespace.
```

第 2 批 pod 的更新，一共 1 个 pod 需要更新：  

```
# 先创建 HookRun，等待 HookRun 的状态
$ kubectl get hookrun
  NAME                                        PHASE     AGE
  pre-delete-hook-test-gamedeployment-67864c6f65-pqrw5-test   Running   54s

# 最终，HookRun 符合预期，完成第 2 批共 1 个 pod 的更新：  
$ kubectl get gamedeployment
  NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
  test-gamedeployment   5         3         3               5       5       40m
$ kubectl get pods | grep gamedeploy
  test-gamedeployment-2nrrt         1/1     Running   0          46m
  test-gamedeployment-mw5nw         1/1     Running   0          46m
  test-gamedeployment-pqrw5         1/1     Running   1          46m
  test-gamedeployment-q78jd         1/1     Running   1          46m
  test-gamedeployment-xv26j         1/1     Running   1          46m

# 已完成更新的 Pod 对应的 HookRun 也已被清理：
$ kubectl get hookrun
  No resources found in default namespace.
```

### 不符合预期的 Hook 

#### 创建 HookTemplate 模板和 GameDeployment 应用

定义一个 HookTemplate: [hook-template-fail.yaml](../example/hook-template-fail.yaml), hook 类型为 Webhook。    
定义一个 GameDeployment: [gamedeployment.yaml](../example/gamedeployment.yaml), 更新策略为 InplaceUpdate，并配置了 PreDeleteHook 的策略。  

创建 Hook 模板和应用：  

```shell
# 创建 hook 模板
$ kubectl apply -f hook-template-fail.yaml
hooktemplate.tkex.tencent.com/test created

# 首次创建应用
$ kubectl apply -f gamedeployment.yaml
gamedeployment.tkex.tencent.com/test-gamedeployment created

$ kubectl get gamedeployment
NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
test-gamedeployment   6         6         6               6       6       73s
$ 
```

#### 缩容 GameDeployment

scale 缩容应用，operator 会创建 HookRun, GameDeployment 并未马上删除实例。  

```shell
# scale GameDeployment，删除一个实例
$ kubectl scale --replicas=5 gamedeployment/test-gamedeployment
gamedeployment.tkex.tencent.com/test-gamedeployment scaled

# 查看 HookRun 资源, 已经创建一个 HookRun
$ kubectl get hookrun
  NAME                                        PHASE     AGE
  pre-delete-hook-test-gamedeployment-67864c6f65-f9tkt-test   Running   9s

# 确认 GameDeployment 状态，并未马上缩容
$ kubectl get gamedeployment
  NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
  test-gamedeployment   5         6         6               6       6       85s
```

一段时间后，HookRun 运行完成，再次确认状态。因为 HookRun 运行失败，结果不符合预期，缩容失败：

```
$ kubectl get gamedeployment
NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
test-gamedeployment   5         6         6               6       6       4m33s

$ kubectl get hookrun
  NAME                                        PHASE    AGE
  pre-delete-hook-test-gamedeployment-67864c6f65-f9tkt-test   Failed   3m29s
```

kubectl get gamedeployment test-gamedeployment -o yaml 检查 GameDeployment preDeleteHookCondition 状态。显示 test-gamedeployment-f9tkt 这个 pod 因为 PreDeleteHook 的状态为 Failed, 不符合预期而删除失败。   

```yaml
apiVersion: tkex.tencent.com/v1alpha1
kind: GameDeployment
metadata:
  ...
spec:
  ...
  updateStrategy:
    inPlaceUpdateStrategy:
      gracePeriodSeconds: 30
    maxUnavailable: 2
    partition: 2
    type: InplaceUpdate
status:
  availableReplicas: 6
  collisionCount: 0
  labelSelector: app=test-gamedeployment
  observedGeneration: 2
  preDeleteHookCondition:
  - phase: Failed
    podName: test-gamedeployment-f9tkt
    startTime: "2020-11-19T10:25:17Z"
  readyReplicas: 6
  replicas: 6
  updateRevision: test-gamedeployment-67864c6f65
  updatedReadyReplicas: 6
  updatedReplicas: 6
```

此时，如果人工介入后解决了 hook 失败的问题，使得 hook 能够符合预期了，就可以触发重试缩容。通过配置 preDeleteUpdateStrategy 下面的retry 参数为 true，即可触发进行重试。

```shell
# 触发重试
$kubectl patch gamedeployment test-gamedeployment --type='json' -p='[{"op": "replace", "path": "/spec/preDeleteUpdateStrategy/retry", "value":true}]'
gamedeployment.tkex.tencent.com/test-gamedeployment patched

# operator 重新创建该 pod 的 preDelete HookRun
$ kubectl get hookrun
  NAME                                        PHASE     AGE
  pre-delete-hook-test-gamedeployment-67864c6f65-f9tkt-test   Running   4s
```

如果这次 HookRun 的结果符合预期，缩容就会成功。  

#### 更新 GameDeployment

使用流程及结果与缩容 GameDeployment 类似，在此不再详述。  

### 在 PreDeleteHook 时配置可变参数

HookTemplate 及 HookRun 还支持自定义变量，GameDeployment 在创建 PreDeleteHook HookRun 时也会传入重要的几个变量，示例如下：  

#### 前置条件

使用上面提到的二进制测试程序 gamedeployment-canary 打成一个镜像 canary-hook:test，使用这个镜像运行一个 pod 后，在 pod 中会在 9091 端口运行这个二进制程序。假设这个 pod 的 IP 为 1.1.1.1，那么调用该 pod 的接口会返回以下结果：  

```
$ curl http://1.1.1.1:9091
{"name":"bryan","male":"yes","age":45}
```

#### 定义带有可变参数的 HookTemplate 和 GameDeployment

定义一个带有可变参数的 HookTemplate: [hook-template-args.yaml](../example/hook-template-args.yaml)。  
定义一个配置有 service-name 参数的 GameDeployment： [gamedeployment-args.yaml](../example/gamedeployment-args.yaml)。  

其中，HookTemplate 中配置了 PodIP, PodName, PodNamespace, service-name 一共 4 个变量 key ，但没有配置具体的 value，期望由创建 HookRun 时传入并渲染。  在 GameDeployment 的 preDeleteUpdateStrategy 中配置了 service-name 这个变量的 key 和 value。  PodIP, PodName, PodNamespace 这 3 个参数不需要在 GameDeployment 的 preDeleteUpdateStrategy 中直接配置，因为这 3 个参数是可变的，bcs-gamedeployment-operator 会在创建每个 pod 的 PreDeleteHookRun 时，把这 3 个参数的 key 和 value 传进到 HookRun 中，并由 HookRun渲染进 url 或 header 当中去。  
因为这个 HookTemplate 的 url 定义为 http://{{ args.PodIP }}:9091，只配置了 PodIP 这个待渲染的变量，所以会把 PodIP 这个 key 的 value 渲染进 url 。

#### 创建 HookTemplate 和 GameDeployment

```
$ kubectl apply -f hook-template-args.yaml
hooktemplate.tkex.tencent.com/test created

$ kubectl apply -f gamedeployment-args.yaml
  gamedeployment.tkex.tencent.com/test-gamedeployment created

$ kubectl get gamedeployment
  NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
  test-gamedeployment   4         4         4               4       4       13s
```

#### 缩容 GameDeployment

```
$ kubectl scale --replicas=2 gamedeployment/test-gamedeployment
  gamedeployment.tkex.tencent.com/test-gamedeployment scaled

# operator 已创建两个 PreDelete HookRun
$ kubectl get hookrun
  NAME                                        PHASE     AGE
  pre-delete-hook-test-gamedeployment-767bfd8bcc-94pn2-test   Running   8s
  pre-delete-hook-test-gamedeployment-767bfd8bcc-twlkl-test   Running   8s
```

#### 查看 PreDelete HookRun 的状态

kubectl get hookrun pre-delete-hook-test-gamedeployment-767bfd8bcc-94pn2-test -o yaml

```yaml
apiVersion: tkex.tencent.com/v1alpha1
kind: HookRun
metadata:
  creationTimestamp: "2020-11-20T01:46:55Z"
  generation: 1
  labels:
    InstanceID: 94pn2
    PodControllerRevision: test-gamedeployment-767bfd8bcc
    gamedeployment-type: PreDelete
  name: test-gamedeployment-767bfd8bcc-94pn2-test
  namespace: default
  ownerReferences:
  - apiVersion: tkex.tencent.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: GameDeployment
    name: test-gamedeployment
    uid: f041dc44-291a-4e43-866f-18ffd3ed4c33
  resourceVersion: "43458406"
  selfLink: /apis/tkex.tencent.com/v1alpha1/namespaces/default/hookruns/test-gamedeployment-767bfd8bcc-94pn2-test
  uid: 872ec6c8-cdb0-42f0-ac1d-55237b9b2bda
spec:
  args:
  - name: PodIP
    value: 1.1.1.1    # fake IP
  - name: PodName
    value: test-gamedeployment-94pn2
  - name: PodNamespace
    value: default
  - name: service-name
    value: test-gamedeployment-svc.default.svc.cluster.local
  metrics:
  - count: 3
    failureLimit: 2
    interval: 60s
    name: webtest
    provider:
      web:
        jsonPath: '{$.age}'
        url: http://{{ args.PodIP }}:9091
    successCondition: asInt(result) > 30
status:
  metricResults:
  - count: 1
    measurements:
    - finishedAt: "2020-11-20T01:46:55Z"
      phase: Successful
      startedAt: "2020-11-20T01:46:55Z"
      value: "32"
    name: webtest
    phase: Running
    successful: 1
  phase: Running
  startedAt: "2020-11-20T01:46:55Z"
```

可见，bcs-gamedeployment-operator 创建的 HookRun 中，已经把 PodIP, PodName, PodNamespace, service-name 的 key 和 value 配置进来，当 HookRun 实际去使用 provider 调用 hook 时，如果 url 或 header 中配置了模板，就会把对应的 value 渲染进去。

#### 缩容成功

```shell
$ kubectl get gamedeployment
  NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
  test-gamedeployment   2         2         2               2       2       13m
$ kubectl get pod | grep gamedeployment
  test-gamedeployment-rk5jv         1/1     Running   0          13m
  test-gamedeployment-xzzpk         1/1     Running   0          13m
$ kubectl get hookrun
  No resources found in default namespace.
```
