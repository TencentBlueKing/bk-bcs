module github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-driver

go 1.14

replace (
	bitbucket.org/ww/goautoneg => github.com/adjust/goautoneg v0.0.0-20150426214442-d788f35a0315
	github.com/Tencent/bk-bcs => ../../../bk-bcs
	github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-watch => ../../bcs-k8s/bcs-k8s-watch
    github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator => ../../bcs-k8s/bcs-gamedeployment-operator
	github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server => ../../bcs-services/bcs-webhook-server
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes => ../../bcs-k8s/kubernetes
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common => ../../bcs-k8s/kubernetes/common
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0 // indirect
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.0.0-20181213150558-05914d821849
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181213153335-0fe22c71c476 // indirect
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20181127025237-2b1284ed4c93
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20181213151703-3ccfe8365421
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181213151034-8d9ed539ba31
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20181109181836-c59034cc13d5
    k8s.io/kubernetes => k8s.io/kubernetes v1.13.1
)

require (
	github.com/Tencent/bk-bcs v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-watch v0.0.0-00010101000000-000000000000
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/json-iterator/go v1.1.10
	github.com/parnurzeal/gorequest v0.2.16
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
)
