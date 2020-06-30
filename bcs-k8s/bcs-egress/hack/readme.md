# 项目处理

## 项目初始化

```shell
#初始化项目
operator-sdk new bcs-egress --header-file ./tkex-statefulsetplus-operator/hack/boilerplate.go.txt
cd bcs-egress

#创建bcsegress定义，用于controller
operator-sdk add api --api-version=bkbcs.tencent.com/v1alpha1 --kind=BCSEgress
#创建bcsEgresscontroller，用于Operator
#operator-sdk add api --api-version=bkbcs.tencent.com/v1alpha1 --kind=BCSEgressController

#controller 代码创建
#operator-sdk add controller --api-version=bkbcs.tencent.com/v1alpha1 --kind=BCSEgress
```

## 项目更新

```shell
#重新生成 k8s各项配置
operator-sdk generate k8s

#更新k8s crds
operator-sdk generate crds
```

