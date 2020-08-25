module github.com/Tencent/bk-bcs/bcs-k8s/kubernetes

go 1.14

replace github.com/Tencent/bk-bcs/bcs-k8s/kubernetes => ./

require (
	github.com/go-logr/logr v0.2.0 // indirect
	github.com/onsi/ginkgo v1.13.0 // indirect
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.2 // indirect
	k8s.io/code-generator v0.18.5
	sigs.k8s.io/controller-runtime v0.6.0
)
