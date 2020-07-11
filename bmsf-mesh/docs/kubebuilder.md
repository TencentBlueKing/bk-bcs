# kubebuiler

通过kubernetes apiserver builder构建事件api，并支持自定义数据结构

依赖kubebuilder v1.0.7版本

[kubebuilder book](https://book-v1.book.kubebuilder.io)

## 前置条件，Gopkg.toml更改

```bash
required = [
+  "github.com/emicklei/go-restful",
+  "github.com/onsi/ginkgo", # for test framework
+  "github.com/onsi/gomega", # for test matchers
+  "k8s.io/client-go/plugin/pkg/client/auth/gcp", # for development against gcp
+  "k8s.io/code-generator/cmd/client-gen", # for go generate
+  "k8s.io/code-generator/cmd/deepcopy-gen", # for go generate
+  "sigs.k8s.io/controller-tools/cmd/controller-gen", # for crd/rbac generation
+  "sigs.k8s.io/controller-runtime/pkg/client/config",
+  "sigs.k8s.io/controller-runtime/pkg/controller",
+  "sigs.k8s.io/controller-runtime/pkg/handler",
+  "sigs.k8s.io/controller-runtime/pkg/manager",
+  "sigs.k8s.io/controller-runtime/pkg/runtime/signals",
+  "sigs.k8s.io/controller-runtime/pkg/source",
+  "sigs.k8s.io/testing_frameworks/integration", # for integration testing
+  "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1",
]
 [[constraint]]
   name = "k8s.io/api"
-  version = "kubernetes-1.12.0"
+  version = "kubernetes-1.12.3"
 
 [[constraint]]
   name = "k8s.io/apimachinery"
-  version = "kubernetes-1.12.0"
+  version = "kubernetes-1.12.3"
 
 [[constraint]]
   name = "k8s.io/client-go"
-  version = "9.0.0"
+  version = "kubernetes-1.12.3"
```

注意在“提示是否选择下载vendor”时，选择“否”，因为我们将在上一层目录中统一加入vendor依赖

## 1. 创建hack/boilderplate.go.txt文件

注入Liscense信息

## 2. 创建资源

```shell
#create resource，如果有多个，运行多次
kubebuilder create api --group mesh --version v1 --kind AppSvc
kubebuilder create api --group mesh --version v1 --kind AppNode
```

修改/pkg/apis/apis.go中

//go:generate go run ../../vendor/k8s.io/code-generator/cmd/deepcopy-gen/main.go -O zz_generated.deepcopy -i ./... -h ../../hack/boilerplate.go.txt

//go:generate go run ../../../vendor/k8s.io/code-generator/cmd/deepcopy-gen/main.go -O zz_generated.deepcopy -i ./... -h ../../hack/boilerplate.go.txt

## 3. client代码生成

```shell
client-gen --go-header-file="./hack/boilerplate.go.txt" --input="mesh/v1" --input-base="github.com/Tencent/bk-bcs/bmsf-mesh/pkg/apis" --clientset-path="github.com/Tencent/bk-bcs/bmsf-mesh/pkg/client"
```

## 4. lister代码生成

```shell
lister-gen --go-header-file="./hack/boilerplate.go.txt" --input-dirs="github.com/Tencent/bk-bcs/bmsf-mesh/pkg/apis/mesh/v1" --output-package="github.com/Tencent/bk-bcs/bmsf-mesh/pkg/client/lister"
```

## 5. informer代码生成

```shell
informer-gen --go-header-file="./hack/boilerplate.go.txt" --input-dirs="github.com/Tencent/bk-bcs/bmsf-mesh/pkg/apis/mesh/v1"  --internal-clientset-package="github.com/Tencent/bk-bcs/bmsf-mesh/pkg/client/internalclientset" --versioned-clientset-package="github.com/Tencent/bk-bcs/bmsf-mesh/pkg/client/internalclientset" --listers-package="github.com/Tencent/bk-bcs/bmsf-mesh/pkg/client/lister" --output-package="github.com/Tencent/bk-bcs/bmsf-mesh/pkg/client/informers" --single-directory=true
```