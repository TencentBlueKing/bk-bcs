module github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager

go 1.14

require (
	github.com/Tencent/bk-bcs v0.0.0-00010101000000-000000000000
	github.com/go-logr/logr v0.2.0
	github.com/kubernetes-client/go v0.0.0-20200222171647-9dac5e4c5400
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	k8s.io/apiextensions-apiserver v0.17.7
	k8s.io/apimachinery v0.17.7
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.14.10
	sigs.k8s.io/controller-runtime v0.5.0
)

replace (
	github.com/Tencent/bk-bcs => ../../
	github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager => ./
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/go-logr/logr => github.com/go-logr/logr v0.1.0
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	k8s.io/client-go => k8s.io/client-go v0.16.7
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
)
