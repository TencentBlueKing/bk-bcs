# 项目代码生成

项目代码生成

## 1. 创建hack/boilderplate.go.txt文件

注入Liscense信息

## 2. 创建资源

修改/pkg/apis/apis.go中

go run ../../../../../../vendor/k8s.io/code-generator/cmd/deepcopy-gen/main.go -O zz_generated.deepcopy -i ./... -h ../../../../hack/boilerplate.go.txt

deepcopy-gen -O zz_generated.deepcopy -i ./... -h ../../../../hack/boilerplate.go.txt

## 3. client代码生成

```shell
go run vendor/k8s.io/code-generator/cmd/client-gen/main.go --go-header-file="./bcs-k8s/bcs-gamestatefulset-operator/hack/boilerplate.go.txt" --input="tkex/v1alpha1" --input-base="bcs-gamestatefulset-operator/pkg/apis" --clientset-path="bcs-gamestatefulset-operator/pkg/clientset"
```

```shell
client-gen --go-header-file="./bcs-k8s/bcs-gamestatefulset-operator/hack/boilerplate.go.txt" --input="tkex/v1alpha1" --input-base="bcs-gamestatefulset-operator/pkg/apis" --clientset-path="bcs-gamestatefulset-operator/pkg/clientset"
```

## 4. lister代码生成

```shell
go run vendor/k8s.io/code-generator/cmd/lister-gen/main.go --go-header-file="./bcs-k8s/bcs-gamestatefulset-operator/hack/boilerplate.go.txt" --input-dirs="bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1" --output-package="bcs-gamestatefulset-operator/pkg/listers"
```

```shell
lister-gen --go-header-file="./bcs-k8s/bcs-gamestatefulset-operator/hack/boilerplate.go.txt" --input-dirs="bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1" --output-package="bcs-gamestatefulset-operator/pkg/listers"
```

生成代码仅有internalversion，迁移至tkex/v1alpha1，修正了list中Watch和List中context参数。

## 5. informer代码生成

```shell
informer-gen --go-header-file="./bcs-k8s/bcs-gamestatefulset-operator/hack/boilerplate.go.txt" --input-dirs="bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"  --internal-clientset-package="bcs-gamestatefulset-operator/pkg/clientset/internalclientset" --versioned-clientset-package="bcs-gamestatefulset-operator/pkg/clientset/internalclientset" --listers-package="bcs-gamestatefulset-operator/pkg/listers" --output-package="bcs-gamestatefulset-operator/pkg/informers" --single-directory=true
```