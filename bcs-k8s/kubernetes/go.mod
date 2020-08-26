module github.com/Tencent/bk-bcs/bcs-k8s/kubernetes

go 1.14

require (
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a
	github.com/go-logr/logr v0.2.0
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.2
	k8s.io/code-generator v0.18.5
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
    github.com/Tencent/bk-bcs/bcs-k8s/kubernetes => ./
    k8s.io/client-go => k8s.io/client-go v0.0.0-20181213151034-8d9ed539ba31
)
