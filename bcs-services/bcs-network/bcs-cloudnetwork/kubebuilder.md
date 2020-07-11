
# 创建client
client-gen --go-header-file="./hack/boilerplate.go.txt" --input="cloud/v1" --input-base="github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/apis" --clientset-path="github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/client"

# 创建lister
lister-gen --go-header-file="./hack/boilerplate.go.txt" --input-dirs="github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/apis/cloud/v1" --output-package="github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/client/lister"

# 创建informer
informer-gen --go-header-file="./hack/boilerplate.go.txt" --input-dirs="github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/apis/cloud/v1" --internal-clientset-package="github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/client/internalclientset" --versioned-clientset-package="github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/client/internalclientset" --listers-package="github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/client/lister" --output-package="github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/client/informers" --single-directory=true