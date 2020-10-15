module github.com/Tencent/bk-bcs/bcs-k8s/kubernetes

go 1.14

replace github.com/Tencent/bk-bcs/bcs-k8s/kubernetes => ./

require (
	github.com/go-logr/logr v0.2.0
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.2
	k8s.io/code-generator v0.18.5
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.6.0
)
