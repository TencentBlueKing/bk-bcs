## 智能式分步骤灰度发布

GameDeployment/GameStatefulSet支持智能化的分步骤灰度发布功能，允许用户在 GameDeployment 定义中配置多个灰度发布的步骤，
这些步骤可以是"灰度发布部分实例"、"暂停灰度发布"、"暂停指定的时间段后再继续灰度发布"、"外部 Hook 调用以决定是否暂停灰度发布"，
通过配置这些不同的灰度发布步骤，可以达到自动化的分步骤灰度发布能力，实现灰度发布的智能控制。  

### 发布步骤

用户可以在 GameDeployment 定义中配置不同的发布步骤，以控制灰度发布的进程。目前，可以在灰度发布中配置以下几种步骤： 

* 灰度指定数量的实例
* 暂停灰度，直到用户触发后才能继续后续步骤
* 暂停指定的时间段，到时后再继续后续步骤
* 外部 Hook 调用：
  * 如果返回的结果满足预期，则继续执行后续步骤，
  * 如果返回结果不满足预期，则自动暂停灰度，由用户手动介入来决定是继续灰度发布还是进行回滚。目前支持 WebHook，Prometheus 两种 Hook 方式。

### 基本定义、原理及实现

#### 基本定义

例如，用户可在 GameDeployment 中定义以下的灰度发布步骤：  
```yaml
apiVersion: tkex.tencent.com/v1alpha1
kind: GameDeployment
metadata:
  name: test-gamedeployment
  labels:
    app: test-gamedeployment
spec:
  replicas: 4
  ……
  updateStrategy:
    type: InplaceUpdate
    maxUnavailable: 2
    canary:
      steps:
        - partition: 3
        - pause: {}
        - partition: 2
        - pause: {duration: 60}
        - hook:
            templateName: test
    inPlaceUpdateStrategy:
      gracePeriodSeconds: 30
```

在这个 GameDeployment 配置中，期望实例数为 4，使用原地重启的更新策略，配置了 5 个灰度发布的步骤：  
* 步骤 0：partition 为 3，灰度 1 个实例；
* 步骤 1：暂停灰度发布，等待用户介入；
* 步骤 2：partition 为 2，灰度 2 个实例；
* 步骤 3：暂停灰度发布 60 秒，60 秒过后继续后续步骤；
* 步骤 4：使用名为 test 的 HookTemplate 进行 hook 调用，如果返回的结果满足预期，则继续执行后续步骤，如果返回结果不满足预期，则暂停灰度，等待

用户手动介入来决定是继续灰度发布还是进行回滚操作。  
如果不需要分步骤灰度发布，那么无需配置 spec.updateStrategy.canary ，仍然按照README.md指引即可。

#### hook 步骤的实现

bcs-gamedeployment-operator 通过与 bcs-hook-operator 的联动来实现灰度发布中的 hook 步骤。如果想要配置分步骤灰度发布中
的 hook 步骤，需要同时部署 bcs-gamedeployment-operator 和 bcs-hook-operator 这两个 bcs 组件。   
bcs-hook-operator 定义了 HookRun 和 HookTemplate 两个 CRD，其原理及实现请参考：[bcs-hook-operator](../../../bcs-hook-operator/README.md)。    
如果在一个 GameDeployment 应用中配置了 hook 调用的步骤，那么 bcs-gamedeployment-operator 就会根据指定 name 的
HookTemplate 创建一个 HookRun crd，bcs-hook-operator watch 到 crd 后就会操作这个 HookRun 进行 hook 调用，
并维护这个 HookRun 的状态。GameDeployment watch这个HookRun的状态，根据 HookRun 的状态来判断是否继续或暂停灰度发布。  

### 使用示例

#### 前置条件

在集群外运行一个示例的 webserver ：

```shell
$ ./gamedeployment-canary
$ curl http://1.1.1.1:9091
{"name":"bryan","male":"yes","age":45}
```

定义一个 [GameDeployment](gamedeployment.yaml)，更新策略为 InplaceUpdate, 配置了 5 个灰度发布的步骤。  
定义一个 [HookTemplate](hooktemplate.yaml), hook 类型为 Webhook。  
根据该 webserver api 的返回及该 HookTemplate 的定义可知，该 hook 调用不会符合预期。

#### 创建应用和 HookTemplate 模板

```shell
# 创建 hook 模板
$ kubectl apply -f hooktemplate.yaml
hooktemplate.tkex.tencent.com/test created

# 首次创建应用
$ kubectl apply -f gamedeployment.yaml
gamedeployment.tkex.tencent.com/test-gamedeployment created

# 确认应用创建成功，4 个实例都为 ready 状态：
$ kubectl get gamedeployment
NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
test-gamedeployment   4         4         4               4       4       56s
```

#### 灰度更新应用

```shell
$ kubectl patch gamedeployment test-gamedeployment --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"python:3.5"}]'
  gamedeployment.tkex.tencent.com/test-gamedeployment patched
```

因为灰度策略中第 0 个步骤设置了 partition 为 3，只更新 1 个实例，第 1 个步骤设置了暂停灰度发布，一段时间后查看
GameDeployment 状态：  

```shell
$ kubectl get gamedeployment
NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
test-gamedeployment   4         1         1               4       4       3m
```

可见应用灰度了 1 个实例后就停止了，使用 kubectl get gamedeployment test-gamedeployment -o yaml 
查看该 GameDeployment 详细状态如下：

```yaml
apiVersion: tkex.tencent.com/v1alpha1
kind: GameDeployment
  ...
  updateStrategy:
    canary:
      steps:
      - partition: 3
      - pause: {}
      - partition: 2
      - pause:
          duration: 60
      - hook:
          templateName: test
    inPlaceUpdateStrategy:
      gracePeriodSeconds: 30
    maxUnavailable: 2
    paused: true
    type: InplaceUpdate
status:
  availableReplicas: 4
  canary:
    revision: test-gamedeployment-67864c6f65
  collisionCount: 0
  currentStepHash: 79cffb7b9
  currentStepIndex: 1
  labelSelector: app=test-gamedeployment
  observedGeneration: 3
  pauseConditions:
  - reason: PausedByCanaryPauseStep
    startTime: "2020-11-10T08:41:36Z"
  readyReplicas: 4
  replicas: 4
  updateRevision: test-gamedeployment-9d78dd5db
  updatedReadyReplicas: 1
  updatedReplicas: 1
```

status.currentStepIndex 为 1，表示当前处于第 1 个步骤；  
updateStrategy.paused 被设为了 true，表示已经暂停该 GameDeployment 的灰度发布，暂停在第 1 个步骤；  
status.pauseConditions 表示暂停原因和暂停时间。  

人工介入，继续灰度发布：  

```shell
$ kubectl patch gamedeployment test-gamedeployment --type='json' -p='[{"op": "replace", "path": "/spec/updateStrategy/paused", "value":false}]'
gamedeployment.tkex.tencent.com/test-gamedeployment patched
```

GameDeployment 会继续执行第 2 个灰度步骤：partition 为 2，灰度 2 个实例。  
执行完步骤 2 后，查看 GameDeployment 状态，可见已完成步骤 2：  

```shell
$ kubectl get gamedeployment
NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
test-gamedeployment   4         2         2               4       4       29m
```

此时，继续使用 kubectl get gamedeployment test-gamedeployment -o yaml 查看该 GameDeployment 详细状态如下：  

```yaml
apiVersion: tkex.tencent.com/v1alpha1
kind: GameDeployment
  ......
  updateStrategy:
    canary:
      steps:
      - partition: 3
      - pause: {}
      - partition: 2
      - pause:
          duration: 60
      - hook:
          templateName: test
    inPlaceUpdateStrategy:
      gracePeriodSeconds: 30
    maxUnavailable: 2
    paused: true
    type: InplaceUpdate
status:
  availableReplicas: 4
  canary:
    revision: test-gamedeployment-67864c6f65
  collisionCount: 0
  currentStepHash: 79cffb7b9
  currentStepIndex: 3
  labelSelector: app=test-gamedeployment
  observedGeneration: 6
  pauseConditions:
  - reason: PausedByCanaryPauseStep
    startTime: "2020-11-10T09:01:05Z"
  readyReplicas: 4
  replicas: 4
  updateRevision: test-gamedeployment-9d78dd5db
  updatedReadyReplicas: 2
  updatedReplicas: 2
```

此时可见 currentStepIndex 为 3，灰度发布暂停在了步骤 3。因为步骤 3 定义了暂停时间为 60 s，所以 60 s 后 
GameDeployment 会继续下一个步骤。  

60 秒时间过了以后，查看 GameDeployment 详细状态：  

```yaml
apiVersion: tkex.tencent.com/v1alpha1
kind: GameDeployment
  ......
  updateStrategy:
    canary:
      steps:
      - partition: 3
      - pause: {}
      - partition: 2
      - pause:
          duration: 60
      - hook:
          templateName: test
    inPlaceUpdateStrategy:
      gracePeriodSeconds: 30
    maxUnavailable: 2
    paused: true
    type: InplaceUpdate
status:
  availableReplicas: 4
  canary:
    currentStepHookRun: test-gamedeployment-9d78dd5db-4-test
    revision: test-gamedeployment-67864c6f65
  collisionCount: 0
  currentStepHash: 79cffb7b9
  currentStepIndex: 4
  labelSelector: app=test-gamedeployment
  observedGeneration: 8
  pauseConditions:
  - reason: PausedByStepBasedHook
    startTime: "2020-11-10T09:02:05Z"
  readyReplicas: 4
  replicas: 4
  updateRevision: test-gamedeployment-9d78dd5db
  updatedReadyReplicas: 2
  updatedReplicas: 2
```

currentStepIndex 为 4，可见 GameDeployment 在步骤 3 暂停 60 秒后，自动地继续执行步骤 4 了，然后因为步骤 4 的 
hook 调用不符合预期，所以被暂停在了步骤 4。暂停的原因为 PausedByStepBasedHook，暂停的时间为 2020-11-10T09:02:05Z 。  
在步骤 4 中，bcs-gamedeployment-operator 创建并执行了一个 HookRun，可查看其状态如下：  

```shell
# kubectl get hookrun
NAME                                   PHASE      AGE
canary-step-hook-test-gamedeployment-9d78dd5db-4-test   Failed     37m
```

使用 kubectl get hookrun test-gamedeployment-9d78dd5db-4-test -o yaml 命令查看到其状态为 Failed:  

```yaml
apiVersion: tkex.tencent.com/v1alpha1
kind: HookRun
......
spec:
  metrics:
  - count: 2
    interval: 60s
    name: webtest
    provider:
      web:
        jsonPath: '{$.age}'
        url: http://1.1.1.1:9091
    successCondition: asInt(result) < 30
status:
  metricResults:
  - count: 1
    failed: 1
    measurements:
    - finishedAt: "2020-11-10T09:02:05Z"
      phase: Failed
      startedAt: "2020-11-10T09:02:05Z"
      value: "32"
    name: webtest
    phase: Failed
  phase: Failed
  startedAt: "2020-11-10T09:02:05Z"
```

此时，需要人工介入以决定是否继续灰度发布。  
用户如果想回滚，可以执行

```shell
kubectl patch gamedeployment test-gamedeployment --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"python:latest"}]'
```

以回滚到上一个版本。  

用户如果想继续灰度发布，可以执行 

```shell
kubectl patch gamedeployment test-gamedeployment --type='json' -p='[{"op": "replace", "path": "/spec/updateStrategy/paused", "value":false}]'
```

以继续下一步骤。  

假设继续灰度发布：  

```shell
$ kubectl patch gamedeployment test-gamedeployment --type='json' -p='[{"op": "replace", "path": "/spec/updateStrategy/paused", "value":false}]'
gamedeployment.tkex.tencent.com/test-gamedeployment patched

$ kubectl get gamedeployment
NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
test-gamedeployment   4         4         2               2       4       70m
```

GameDeployment 因为已经执行完定义的所有 5 个灰度步骤，此时就会灰度所有实例以完成应用的整个发布。  
一段时间后查看应用状态，已更新所有 4 个实例：

```shell
$ kubectl get gamedeployment
NAME                  DESIRED   UPDATED   UPDATED_READY   READY   TOTAL   AGE
test-gamedeployment   4         4         4               4       4       73m
```

#### 补充

在上述的 hook 配置中，hook 调用是不符合预期的，所以会中断灰度发布。如果定义的 hook 调用是符合预期的，比如HookTemplate
中的successCondition 改为 "asInt(result) > 30"，此时，步骤 4 的 hook 调用是符合预期的，那么 GameDeployment 在执
行完后步骤 4 后就继续以下步骤，不会暂停在步骤 4。因步骤 4 是最后一个步骤，那么 GameDeployment 在执行完步骤4后，就会继续
完成应用的所有实例的更新。
