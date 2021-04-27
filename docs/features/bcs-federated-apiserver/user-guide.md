#### 所有member集群 指定namespace下name的pod资源
```shell
curl -k  --header "Authorization: Bearer ${TOKEN}" https://xxx.xxx.xxx.xxx:xxx/apis/aggregation.federated.bkbcs.tencent.com/v1alpha1/namespaces/default/podaggregations/descheduler-descheduler-helm-chart-1611736920-8jtk9
[root@centos ~]# kubectl agg pod -n default descheduler-descheduler-helm-chart-1611736920-8jtk9 -o wide
NAMESPACE       NAME                                                            READY   STATUS          RESTARTS  AGE       IP                  NODE                                    NOMINATED NODE      READINESS GATES
default         descheduler-descheduler-helm-chart-1611736920-8jtk9             0/1     Succeeded       0         76d       10.xxx.x.xxx        ip-bcs-k8s-xxxxx                        <none>              <none>
```

#### 返回所有member集群的某namespace下的所有Pod资源(结果是个Pod的List)
```shell
curl -k  --header "Authorization: Bearer ${TOKEN}" https://xxx.xxx.xxx.xxx:xxx/apis/aggregation.
federated.bkbcs.tencent.com/v1alpha1/namespaces/default/podaggregations
[root@centos ~]# kubectl agg pod -n default -o wide
NAMESPACE       NAME                                                            READY   STATUS          RESTARTS  AGE       IP                  NODE                                    NOMINATED NODE      READINESS GATES
default         hellop-0                                                        1/1     Running         0         117d      xx.xxx.xx.xxx       ip-x-xxx-xx-xxx-n-bcs-k8s-15000         <none>              <none>
default         descheduler-descheduler-helm-chart-1611736920-8jtk9             0/1     Succeeded       0         76d       xx.xxx.xx.xxx       ip-x-xxx-xx-xxx-n-bcs-k8s-15000         <none>              <none>
```

#### 返回所有member集群的某namespace下的label的所有Pod资源(结果是个Pod的List)
```shell
curl -k  --header "Authorization: Bearer ${TOKEN}" https://xxx.xxx.xxx.xxx:xxx/apis/aggregation.federated.bkbcs.tencent.com/v1alpha1/namespaces/default/podaggregations?labelSelector=job-name%3Ddescheduler-descheduler-helm-chart-1611736920
[root@centos ~]# kubectl agg pod -n default -l job-name=descheduler-descheduler-helm-chart-1611736920 -o wide
NAMESPACE       NAME                                                            READY   STATUS          RESTARTS  AGE       IP                  NODE                                    NOMINATED NODE      READINESS GATES
default         descheduler-descheduler-helm-chart-1611736920-8jtk9             0/1     Succeeded       0         76d       xx.xxx.x.xxx        ip-x-xxx-xx-xxx-n-bcs-k8s-xxxxx         <none>              <none>
```

#### 返回所有member集群下所有Pod资源(结果是个Pod的List)
```shell
curl -k  --header "Authorization: Bearer Bearer ${TOKEN}" https://xxx.xxx.xxx.xxx:xxx/apis/aggregation.federated.bkbcs.tencent.com/v1alpha1/podaggregations
[root@centos ~]# kubectl agg pod -A -o wide
NAMESPACE       NAME                                                            READY   STATUS          RESTARTS  AGE       IP                  NODE                                    NOMINATED NODE      READINESS GATES
bcs-system      bcs-k8s-watch-76c78c49f-8n6dj                                   1/1     Running         0         152d      xx.xxx.x.xxx        ip-x-xxx-xx-xxx-m-bcs-k8s-xxxxx         <none>              <none>
archie-test     hellop-0                                                        1/1     Running         0         117d      xx.xxx.x.xxx        ip-x-xxx-xx-xxx-n-bcs-k8s-xxxxx         <none>              <none>
default         hellop-0                                                        1/1     Running         0         117d      xx.xxx.x.xxx        ip-x-xxx-xx-xxx-n-bcs-k8s-xxxxx         <none>              <none>
bcs-system      bcs-gamestatefulset-operator-84dc56d775-2mccv                   1/1     Running         0         112d      xx.xxx.x.xxx        ip-x-xxx-xx-xxx-n-bcs-k8s-xxxxx         <none>              <none>
bcs-system      bcs-gamedeployment-operator-7d75b6d89-lgwbk                     1/1     Running         0         112d      xx.xxx.x.xxx        ip-x-xxx-xx-xxx-n-bcs-k8s-xxxxx         <none>              <none>
bcs-system      bcs-hook-operator-6775d7d8d9-7qr74                              1/1     Running         0         112d      xx.xxx.x.xxx        ip-x-xxx-xx-xxx-n-bcs-k8s-xxxxx         <none>              <none>
test-ns         deploy1-6799fc88d8-qmdp7                                        1/1     Running         0         78d       xx.xxx.x.xxx        ip-x-xxx-xx-xxx-n-bcs-k8s-xxxxx         <none>              <none>
test-ns         deploy1-6799fc88d8-wkdqb                                        1/1     Running         0         78d       xx.xxx.x.xxx        ip-x-xxx-xx-xxx-m-bcs-k8s-xxxxx         <none>              <none>
bcs-system      bcs-kube-agent-867f6fdf8d-w2vd4                                 1/1     Running         0         77d       xx.xxx.x.xxx        ip-x-xxx-xx-xxx-m-bcs-k8s-xxxxx         <none>              <none>
default         descheduler-descheduler-helm-chart-1611736920-8jtk9             0/1     Succeeded       0         76d       xx.xxx.x.xxx        ip-x-xxx-xx-xxx-n-bcs-k8s-xxxxx         <none>              <none>
bcs-system      bcs-k8s-watch-d56cf4d68-n8rkl                                   1/1     Running         0         76d       xx.xxx.x.xxx        ip-x-xxx-xx-xxx-m-bcs-k8s-xxxxx         <none>              <none>
```

#### 返回所有member集群的label为xxxx的所有Pod资源(结果是个Pod的List)
```shell
curl -k  --header "Authorization: Bearer ${TOKEN}" https://xxx.xxx.xxx.xxx:xxx/aggregation.federated.
bkbcs.tencent.com/v1alpha1/podaggregations?labelSelector=job-name%3Ddescheduler-descheduler-helm-chart-1611736920
[root@centos ~]# kubectl agg pod -A -l job-name=descheduler-descheduler-helm-chart-1611736920 -o wide
NAMESPACE       NAME                                                            READY   STATUS          RESTARTS  AGE       IP                  NODE                                    NOMINATED NODE      READINESS GATES
default         descheduler-descheduler-helm-chart-1611736920-8jtk9             0/1     Succeeded       0         76d       xx.xxx.x.xxx        ip-x-xxx-xx-xxx-n-bcs-k8s-xxxxx         <none>              <none>
```

> 注： 以上结果是Pod的List； kubectl agg pod 命令，支持-o wide、 --show-labels 选项，以展示详细信息、标签信息等