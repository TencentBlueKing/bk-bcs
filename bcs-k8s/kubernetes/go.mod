module github.com/Tencent/bk-bcs/bcs-k8s/kubernetes

go 1.14

require (
	github.com/Tencent/bk-bcs v0.0.0-00010101000000-000000000000
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.2
	k8s.io/code-generator v0.18.5
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Tencent/bk-bcs => ../../
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes => ./
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)
