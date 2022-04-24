module github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common

go 1.13

replace (
	bitbucket.org/ww/goautoneg => github.com/adjust/goautoneg v0.0.0-20150426214442-d788f35a0315
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.20.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.0
	k8s.io/apiserver => k8s.io/apiserver v0.20.0
	k8s.io/client-go => k8s.io/client-go v0.20.0
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd
	k8s.io/kubernetes => k8s.io/kubernetes v1.20.0
)

require (
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.3 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20200820163806-098863c1fc24
	github.com/onsi/ginkgo v1.14.2 // indirect
	github.com/onsi/gomega v1.10.4 // indirect
	github.com/prometheus/client_golang v0.9.1
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9 // indirect
	golang.org/x/oauth2 v0.0.0-20201109201403-9fd604954f58 // indirect
	k8s.io/api v0.20.0
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v0.20.0
	k8s.io/klog v1.0.0
)
