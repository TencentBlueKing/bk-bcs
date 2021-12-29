module github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-clusternet-controller

go 1.14

require (
	github.com/clusternet/clusternet v0.5.0
	github.com/onsi/ginkgo v1.14.1 // indirect
	github.com/onsi/gomega v1.10.2 // indirect
	k8s.io/apimachinery v0.21.2
	k8s.io/apiserver v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/controller-manager v0.21.2
	k8s.io/klog/v2 v2.8.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
