module github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager

go 1.14

require (
	github.com/Tencent/bk-bcs v0.0.0-00010101000000-000000000000
	github.com/go-logr/logr v0.2.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.4.2
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/kubernetes-client/go v0.0.0-20200222171647-9dac5e4c5400
	github.com/micro/go-micro/v2 v2.9.1
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	google.golang.org/genproto v0.0.0-20200722002428-88e341933a54
	google.golang.org/grpc v1.31.0
	istio.io/istio v0.0.0-20200821014521-882778a67948
	k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.14.10
	sigs.k8s.io/controller-runtime v0.6.1
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/Tencent/bk-bcs => ../../
	github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager => ./
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/go-logr/logr => github.com/go-logr/logr v0.1.0
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	istio.io/istio => istio.io/istio v0.0.0-20200807215558-ae959de3c67a
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
)
