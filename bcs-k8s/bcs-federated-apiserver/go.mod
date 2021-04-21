module github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver

go 1.13

require (
	github.com/go-openapi/loads v0.19.4
	github.com/go-openapi/spec v0.19.3
	github.com/gogo/protobuf v1.3.2 // indirect
	go.uber.org/automaxprocs v1.4.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/apiserver v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd
	sigs.k8s.io/apiserver-builder-alpha v1.18.0
	sigs.k8s.io/kubefed v0.7.0
)

replace github.com/markbates/inflect => github.com/markbates/inflect v1.0.4

replace k8s.io/api v0.20.2 => k8s.io/api v0.18.4

replace k8s.io/apimachinery v0.20.2 => k8s.io/apimachinery v0.18.4

replace k8s.io/apiserver v0.20.2 => k8s.io/apiserver v0.18.4

replace k8s.io/client-go v0.20.2 => k8s.io/client-go v0.18.4

replace github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0

replace github.com/onsi/ginkgo v1.14.2 => github.com/onsi/ginkgo v1.11.0

replace k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6

replace github.com/kubernetes-incubator/reference-docs => github.com/kubernetes-sigs/reference-docs v0.0.0-20170929004150-fcf65347b256

replace github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg => ./pkg
