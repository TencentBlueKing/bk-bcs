module github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch

go 1.17

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
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.4.0
)

require (
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common v0.0.0-20210223080803-f27f3f3c01c4 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.3 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20200820163806-098863c1fc24 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/nxadm/tail v1.4.4 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.15.0 // indirect
	github.com/prometheus/procfs v0.2.0 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/ugorji/go/codec v1.2.3 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9 // indirect
	golang.org/x/oauth2 v0.0.0-20210113205817-d3ed898aa8a3 // indirect
	golang.org/x/sys v0.0.0-20220128215802-99c3d69c2c27 // indirect
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1 // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6 // indirect
	k8s.io/kubernetes v1.13.1 // indirect
	moul.io/http2curl v1.0.0 // indirect
	sigs.k8s.io/testing_frameworks v0.1.2 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
