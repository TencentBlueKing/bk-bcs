
```bash
# 创建go.mod
export GO111MODULE=on
go mod init bk-bcs/bcs-services/bcs-clb-controller

# 初始化
kubebuilder init --domain bmsf.tencent.com

# 创建crd
export GO111MODULE=auto
kubebuilder create api --group clb --version v1 --kind ClbIngress
kubebuilder create api --group network --version v1 --kind CloudListener

# 创建client
client-gen --go-header-file="./hack/boilerplate.go.txt" --input="clb/v1,network/v1,/mesh/v1" --input-base="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/apis" --clientset-path="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/client"

# 创建lister
lister-gen --go-header-file="./hack/boilerplate.go.txt" --input-dirs="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/clb/v1,bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/network/v1,bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/mesh/v1" --output-package="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/client/lister"

# 创建informer
informer-gen --go-header-file="./hack/boilerplate.go.txt" --input-dirs="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/clb/v1,bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/network/v1,bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/mesh/v1" --internal-clientset-package="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/client/internalclientset" --versioned-clientset-package="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/client/internalclientset" --listers-package="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/client/lister" --output-package="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/client/informers" --single-directory=true

```

```bash
kubebuilder create api --group mesh --version v1 --kind AppSvc
kubebuilder create api --group mesh --version v1 --kind AppNode

client-gen --go-header-file="./hack/boilerplate.go.txt" --input="mesh/v1" --input-base="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/apis" --clientset-path="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/client"

lister-gen --go-header-file="./hack/boilerplate.go.txt" --input-dirs="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/mesh/v1" --output-package="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/client/lister"

informer-gen --go-header-file="./hack/boilerplate.go.txt" --input-dirs="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/mesh/v1" --internal-clientset-package="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/client/internalclientset" --versioned-clientset-package="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/client/internalclientset" --listers-package="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/client/lister" --output-package="github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/client/informers" --single-directory=true
```