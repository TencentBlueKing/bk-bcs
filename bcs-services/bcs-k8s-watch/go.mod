module github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch

go 1.14

replace (
	bitbucket.org/ww/goautoneg => github.com/adjust/goautoneg v0.0.0-20150426214442-d788f35a0315
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210818040851-76fdc539dc33
	github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs => github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs v0.0.0-20210117140338-aeaed29b1997
	github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2 => github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2 v0.0.0-20210117140338-aeaed29b1997
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/go-logr/logr => github.com/go-logr/logr v0.1.0
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0 // indirect
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.0.0-20181126151915-b503174bad59
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181126155829-0cd23ebeb688
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20181126123746-eddba98df674
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181126152608-d082d5923d3c
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20181109181836-c59034cc13d5
	k8s.io/kubernetes => k8s.io/kubernetes v1.13.1
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator v0.0.0-20210818040851-76fdc539dc33
	github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator v0.0.0-20210818040851-76fdc539dc33
	github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2 v0.0.0-00010101000000-000000000000
	github.com/cch123/gogctuner v0.0.0-20220122185012-72c5314ae0b1
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/json-iterator/go v1.1.10
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.4
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/satori/go.uuid v1.2.0
	github.com/sheerun/queue v1.0.1
	github.com/spf13/pflag v1.0.5
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	golang.org/x/oauth2 v0.0.0-20210113205817-d3ed898aa8a3 // indirect
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.4.0
)
