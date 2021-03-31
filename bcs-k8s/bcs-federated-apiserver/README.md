## bcs-federated-apiserver
### 简介
bcs-federated-apiserver 是联邦方案获取member集群 Pod 等 Workload 资源状态信息 的重要途径。它是 kube-apiserver 在联邦解决方案的扩展。

bcs-federated-apiserver 从 bcs-storage 获取 该联邦集群下 member 集群的 Pod 等资源状态，并通过 federated API 在 kubernetes 
apiserver中公开它们，以供 Saas、 用户 等使用。还可以通过 kubectl agg pod 访问 federated API，从而轻松的获取 member 集群 Pod 的聚合信息。

### 使用方式

#### 获取指定 namespace 下所有 Pod 信息

#### 命令行方式
```shell
kubectl agg pod -n default -o wide
```

##### api方式
```shell
curl -k  --header "Authorization: Bearer ${token}" https://${kubernetes-apiserver-address}:${port} /apis/aggregation.federated.bkbcs.tencent.com/v1alpha1/namespaces/default/podaggregations
```


#### 获取指定 namespace 下，指定 label 的 Pod 信息

#### 命令行方式
```shell
kubectl agg pod -n default -l job-name=descheduler-xxxxxxx -o wide
```

##### api方式
```shell
curl -k  --header "Authorization: Bearer ${token}" https://${kubernetes-apiserver-address}:${port}/apis/aggregation.federated.bkbcs.tencent.com/v1alpha1/namespaces/default/podaggregations?labelSelector=job-name%3Ddescheduler-xxxxxxx
```


#### 获取联邦集群下所有 Pod 信息

#### 命令行方式
```shell
kubectl agg pod -A -o wide
```

##### api方式
```shell
curl -k --header "Authorization: Bearer ${token}" https://${kubernetes-apiserver-address}:${port}/apis/aggregation.federated.bkbcs.tencent.com/v1alpha1/podaggregations
```


#### 获取联邦集群下，指定label的所有 Pod 信息

#### 命令行方式
```shell
kubectl agg pod -A -l job-name=descheduler-xxxxxxx
```

##### api方式
```shell
curl -k --header "Authorization: Bearer ${token}" https://${kubernetes-apiserver-address}:${port}/apis/aggregation.federated.bkbcs.tencent.com/v1alpha1/podaggregations?labelSelector=job-name%3Ddescheduler-xxxxxxx
```


#### 获取联邦集群下，获取指定名称的 Pod 信息
#### 命令行方式
```shell
kubectl agg pod descheduler-xxxxxxx
```

##### api方式
```shell
curl -k --header "Authorization: Bearer ${token}" https://${kubernetes-apiserver-address}:${port}
/apis/aggregation.federated.bkbcs.tencent.com/v1alpha1/namespaces/default/podaggregations/descheduler-xxxxxxx-8jtk9
```

### 调用结果
#### 通过 kubectl 命令行调用
返回与kubectl get pod (-o wide) 一致的返回结果。
```shell
NAMESPACE	NAME                                                            READY   STATUS          RESTARTS  AGE       IP                  NODE                                    NOMINATED NODE      READINESS GATES
bcs-system	bcs-k8s-watch-76c78c49f-xxxxx                                   1/1     Running         0         139d      xx.xxx.0.11         ip-x-xxx-xx-xxx-m-bcs-k8s-yyyyy         <none>              <none>
archie-test     hellop-0                                                        1/1     Running         0         104d      xx.xxx.x.xxx        ip-x-xxx-xx-xxx-n-bcs-k8s-xxxxx         <none>              <none>
```

#### 通过API方式调用
返回AggregatedPodList，Item为Pod的原始数据结构，与官方保持一致。
```yaml
{
  "kind": "PodAggregationList",
  "apiVersion": "aggregation.federated.bkbcs.tencent.com/v1alpha1",
  "metadata": {
    "selfLink": "/apis/aggregation.federated.bkbcs.tencent.com/v1alpha1/namespaces/default/podaggregations/descheduler-xxxxxxxxx"
  },
  "items": [
    {
      "metadata": {
        "name": "descheduler-xxxxxxxxx",
        ...
      },
      "spec": {
        "containers": [
        ...
        ],
      },
      "status": {
      ...
      }
    }
  ]
}
```