module github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler

go 1.14

replace (
	bitbucket.org/ww/goautoneg => github.com/adjust/goautoneg v0.0.0-20150426214442-d788f35a0315
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210201154723-6cbc7cfcb38d
	github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated => github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated v0.0.0-20210117140338-aeaed29b1997
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes => github.com/Tencent/bk-bcs/bcs-k8s/kubernetes v0.0.0-20210608083141-16270a254fb4
	github.com/Tencent/bk-bcs/bcs-network => ../../bcs-network
	github.com/containernetworking/cni => github.com/containernetworking/cni v0.6.0
	github.com/containernetworking/plugins => github.com/containernetworking/plugins v0.6.0
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.18.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.6
	k8s.io/client-go => k8s.io/client-go v0.18.6
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-network v0.0.0-00010101000000-000000000000
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/prometheus/client_golang v1.9.0
	google.golang.org/grpc v1.31.0
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kubernetes v1.13.0
)
