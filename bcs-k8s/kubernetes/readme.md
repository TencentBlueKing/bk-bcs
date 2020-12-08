# how to add new crd in bcs/k8s/kubernetes

## step1: install kubebuilder v2.3.1 (if kubebuilderv2.3.1 already installed, skip this step)

```shell
os=$(go env GOOS)
arch=$(go env GOARCH)

# download kubebuilder and extract it to tmp
curl -L https://go.kubebuilder.io/dl/2.3.1/${os}/${arch} | tar -xz -C /tmp/

# move to a long-term location and put it on your path
# (you'll need to set the KUBEBUILDER_ASSETS env var if you put it somewhere else)
sudo mv /tmp/kubebuilder_2.3.1_${os}_${arch} /usr/local/kubebuilder
export PATH=$PATH:/usr/local/kubebuilder/bin
```

## step2: use kubebuilder builder create api

假设创建的crd为TestCrd，group为test，version为v1alpha

```shell
# 进入文件夹
cd $GOPATH/src/bk-bcs/bcs-k8s/kubernetes

go mod tidy
go mod vendor

# 创建CRD
kubebuilder create api --group test --version v1alpha --kind TestCrd
Create Resource [y/n]
y
Create Controller [y/n]
n
```

**注意**: 创建CRD时，会报以下错误，因为删除了kubebuilder生成测main.go。请忽略

```shell
2020/06/28 19:55:01 failed to create API: error updating main.go: failed to open main.go: open main.go: no such file or directory
```

创建完成之后需要为TestCrd加上注释

```golang
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
```

为TestCrdList加上注释

```golang
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
```

完整内容

```golang
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// NodeNetwork is the Schema for the nodenetworks API
type NodeNetwork struct {}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// NodeNetworkList contains a list of NodeNetwork
type NodeNetworkList struct {}
```

为groupversion_info.go文件加入以下代码

```golang
var (
    SchemeGroupVersion = GroupVersion
)

// Resource is required by pkg/client/listers/...
func Resource(resource string) schema.GroupResource {
    return SchemeGroupVersion.WithResource(resource).GroupResource()
}
```

## step3: generate deepcopy manifests client lister and informer

```shell
./update-coden.sh
```
