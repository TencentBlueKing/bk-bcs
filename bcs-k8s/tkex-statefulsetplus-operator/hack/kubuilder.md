# 项目代码生成

项目代码生成

## 1. 创建hack/boilderplate.go.txt文件

注入Liscense信息

## 2. 创建资源

修改/pkg/apis/apis.go中

go run ../../../../../../vendor/k8s.io/code-generator/cmd/deepcopy-gen/main.go -O zz_generated.deepcopy -i ./... -h ../../../../hack/boilerplate.go.txt

## 3. client代码生成

```shell
go run vendor/k8s.io/code-generator/cmd/client-gen/main.go --go-header-file="./bcs-k8s/tkex-statefulsetplus-operator/hack/boilerplate.go.txt" --input="tkex/v1alpha1" --input-base="bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/apis" --clientset-path="bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/clientset"
```

## 4. lister代码生成

```shell
go run vendor/k8s.io/code-generator/cmd/lister-gen/main.go --go-header-file="./bcs-k8s/tkex-statefulsetplus-operator/hack/boilerplate.go.txt" --input-dirs="bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/apis/tkex/v1alpha1" --output-package="bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/listers"
```

## 5. informer代码生成

```shell
go run vendor/k8s.io/code-generator/cmd/informer-gen/main.go --go-header-file="./bcs-k8s/tkex-statefulsetplus-operator/hack/boilerplate.go.txt" --input-dirs="bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/apis/tkex/v1alpha1"  --internal-clientset-package="bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/clientset/internalclientset" --versioned-clientset-package="bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/clientset/internalclientset" --listers-package="bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/listers" --output-package="bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/informers" --single-directory=true

```