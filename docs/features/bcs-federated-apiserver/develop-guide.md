### 安装工具
```shell
wget https://github.com/kubernetes-sigs/apiserver-builder-alpha/releases/download/v1.18.0/apiserver-builder-alpha-v1.18.0-linux-amd64.tar.gz
tar -zxvf apiserver-builder-alpha-v1.18.0-linux-amd64.tar.gz
sudo mkdir -p /usr/local/apiserver-builder/
sudo mv ./bin/apiserver-boot /usr/local/bin/apiserver-builder
```

### 验证工具
```shell
apiserver-boot version
```

### 工程名
```shell
// 在新的branch下操作
mkdir -p ./bcs-k8s/bcs-federated-apiserver
cd ./bcs-k8s/bcs-federated-apiserver
```

### 初始化
```shell
cp ./bcs-k8s/kubernetes/common/hack/boilerplate.go.txt ./
GOROOT=/usr/lib/golang/ apiserver-boot init repo --domain federated.bkbcs.tencent.com
```

### 创建 gvr
```shell
GOROOT=/usr/lib/golang/ apiserver-boot create group version resource --group aggregation --version v1alpha1 --kind PodAggregation
```

### 归档 apis 目录
> 为了调用统一，将 ./bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis 目录中api定义部分 归档至 ./bk-bcs/bcs-k8s/kubernetes/apis
 
### 修改types.go 为自定义的结构体
```yaml
// 在 ./bk-bcs/bcs-k8s/kubernetes/apis/aggregation/v1alpha1/podaggregation_types.go 中填充 
  PodAggregation、PodAggregationList 结构体
type PodAggregation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   v1.PodSpec   `json:"spec,omitempty"`
	Status v1.PodStatus `json:"status,omitempty"`
}

type PodAggregationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodAggregation `json:"items"`
}
```

## 代码生成
### 默认步骤及顺序 （示例，不使用此方式）
```shell
[root@centos ./bk-bcs/bcs-k8s/bcs-federated-apiserver]# GOROOT=/usr/lib/golang/ apiserver-boot build generated
/usr/local/apiserver-builder/apiregister-gen --input-dirs github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis/... --go-header-file boilerplate.go.txt
/usr/local/apiserver-builder/conversion-gen --input-dirs github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis/aggregation/v1alpha1 --input-dirs github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis/aggregation -o /data.go.txt -O zz_generated.conversion --extra-peer-dirs k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/conversion,k8s.io/apimachinery/pkg/runtime
/usr/local/apiserver-builder/deepcopy-gen --input-dirs github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis/aggregation/v1alpha1 --input-dirs github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis/aggregation -o /data/go.txt -O zz_generated.deepcopy
/usr/local/apiserver-builder/openapi-gen --input-dirs github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis/aggregation/v1alpha1 -o /data/go_workspaces/src --go-header-file boilerplate.go.txt -i k8s.io/apimachinery/pkg/apis/meta/apimachinery/pkg/version,k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/util/intstr,k8s.io/api/core/v1,k8s.io/api/apps/v1 --report-filename violations.report --output-package github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg
/usr/local/apiserver-builder/defaulter-gen --input-dirs github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis/aggregation/v1alpha1 --input-dirs github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis/aggregation -o /data/go.txt -O zz_generated.defaults --extra-peer-dirs= k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/conversion,k8s.io/apimachinery/pkg/runtime
/usr/local/apiserver-builder/client-gen -o /data/go_workspaces/src --go-header-file boilerplate.go.txt --input-base github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis --input aggregation/v1alpha1 --clientset-path github.com/pkg/client/clientset_generated --clientset-name clientset
/usr/local/apiserver-builder/lister-gen --input-dirs github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis/aggregation/v1alpha1 -o /data/go_workspaces/src --go-header-file boilerplate.go.txt --output-package github.com/Tencent/nt/listers_generated
/usr/local/apiserver-builder/informer-gen --input-dirs github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis/aggregation/v1alpha1 -o /data/go_workspaces/src --go-header-file boilerplate.go.txt --output-package github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/client/informers_generated --listers-package github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/client/listers_generated --versioned-clientset-package github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/client/clientset_generated/clientset
```

### 代码生成必需步骤
* protobuf: 使用修改命令
* apiregister： 使用默认命令，需调整部分代码
* deepcopy: 使用修改命令
* conversion: 使用修改命令
* openapi: 使用默认命令，需调整部分代码
* client: 使用修改命令，需调整部分代码

#### 生成 protobuf 代码

##### 安装protoc
```shell
wget https://github.com/protocolbuffers/protobuf/releases/download/v3.15.6/protoc-3.15.6-linux-x86_64.zip
unzip protoc-3.15.6-linux-x86_64.zip  -d protoc3
cp -a protoc3/bin/* /usr/local/bin/
cp -a protoc3/include/* /usr/local/include/

go get golang.org/x/tools/cmd/goimports
cp -a ${GOPATH}/bin/goimports  /usr/local/bin/
```

##### 生成代码
 因 apiserver-builder-alpha apiserver-boot命令不支持 --go-header-file 参数，使用命令生成 protobuf 代码会报错( 
 GOROOT=/usr/lib/golang/ apiserver-boot build generated --generator protobuf)，需使用如下 go-to-protobuf 命令：

```shell
GOROOT=/usr/lib/golang/ /usr/local/apiserver-builder/go-to-protobuf --packages github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/aggregation/v1alpha1 --apimachinery-packages -k8s.io/apimachinery/pkg/util/intstr,-k8s.io/apimachinery/pkg/api/resource,-k8s.io/apimachinery/pkg/runtime/schema,-k8s.io/apimachinery/pkg/runtime,-k8s.io/apimachinery/pkg/apis/meta/v1,-sigs.k8s.io/apiserver-builder-alpha/pkg/builders,-k8s.io/api/core/v1 --drop-embedded-fields k8s.io/apimachinery/pkg/apis/meta/v1.TypeMeta,k8s.io/apimachinery/pkg/runtime.Serializer --proto-import=./vendor --vendor-output-base=./vendor/ --go-header-file ./boilerplate.go.txt
```
> 另：由于 type 中定义 resource 引用了 pod 字段，在原始命令中 --apimachinery-packages 部分需增加 -k8s.io/api/core/v1

#### 生成 apiregister 代码
```shell
GOROOT=/usr/lib/golang/ apiserver-boot build generated --generator apiregister
```

#### 生成 deepcopy、conversion 代码
> // 跳转至 /bcs-k8s/kubernetes/ 路径。 在 "默认步骤及顺序" 部分的默认命令基础上 (conversion-gen、deepcopy-gen) ，调整路径 --input-dirs 
> 为 github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/aggregation 相应路径。

#### 生成 openapi 代码
```shell
cd ./bcs-k8s/bcs-federated-apiserver
/usr/local/apiserver-builder/openapi-gen --input-dirs github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/aggregation/v1alpha1 -o /data/go_workspaces/src --go-header-file boilerplate.go.txt -i k8s.io/apimachinery/pkg/apis/meta/apimachinery/pkg/version,k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/util/intstr,k8s.io/api/core/v1,k8s.io/api/apps/v1 --report-filename violations.report --output-package github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg
```

#### 生成 clientset 代码
```shell
// 跳转至 /bcs-k8s/kubernetes/ 路径
GOROOT=/usr/lib/golang/ /usr/local/apiserver-builder/client-gen -o /data/go_workspaces/src --go-header-file boilerplate.go.txt --input-base github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis --input aggregation/v1alpha1 --clientset-path github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset --clientset-name versioned
```

#### 实现 Get、List 接口
> 在 bcs-k8s/bcs-federated-apiserver/pkg/apis/aggregation/podaggregation_rest.go 文件中，实现 
> GetterWithOptions、Lister 接口。

> 注意实现 Getter 接口时，因需要返回联邦 member 集群中所有指定名称的 Pod，在statefulset等场景下，
> 可能返回多个pod，故将返回结果调整为 List。

#### 部分代码增加、调整 (略)
* apiserver 中：cluster 信息、bcs-storage 信息的实现
* apiserver 中：上述内容从 configmap 获取的实现
* kubectl-agg 的实现（调用 生成的 clientSet、Get、List等方法）

### 构建二进制、构建镜像
```shell
GOROOT=/usr/lib/golang/ apiserver-boot build container --image mirrors.tencent.com/test/bcs-federated-apiserver:v0.1.1 --generate=false --targets=apiserver
```

### 生成配置文件
```shell
apiserver-boot build config --name bcs-federated-apiserver --namespace bcs-system --image mirrors.tencent.com/test/bcs-federated-apiserver:v0.1.1
```


### metrics
```shell
 curl -k --header "Authorization: Bearer ${TOKEN}" https://xxx.xxx.xxx.xxx:xxx/metrics
```