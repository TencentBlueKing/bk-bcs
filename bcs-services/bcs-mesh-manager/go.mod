module github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager

go 1.14

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210117140338-aeaed29b1997
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.2.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.4.3
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/kubernetes-client/go v0.0.0-20200222171647-9dac5e4c5400
	github.com/micro/go-micro/v2 v2.9.1
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	google.golang.org/genproto v0.0.0-20200806141610-86f49bd18e98
	google.golang.org/grpc v1.31.0
	gopkg.in/yaml.v2 v2.3.0 // indirect
	istio.io/istio v0.0.0-20200812220246-25bea56c0eb0 // indirect
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.17.2 // indirect
	k8s.io/kubernetes v1.14.10
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/structured-merge-diff v1.0.2 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/go-logr/logr => github.com/go-logr/logr v0.1.0
	github.com/openshift/api => github.com/openshift/api v0.0.0-20180801171038-322a19404e37
	github.com/ugorji/go v1.1.4 => github.com/ugorji/go v0.0.0-20181204163529-d75b2dcb6bc8
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	istio.io/istio => istio.io/istio v0.0.0-20200812220246-25bea56c0eb0
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	k8s.io/kubectl => k8s.io/kubectl v0.16.15
)
