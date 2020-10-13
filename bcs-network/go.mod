module github.com/Tencent/bk-bcs/bcs-network

go 1.14

replace (
	bitbucket.org/ww/goautoneg => github.com/adjust/goautoneg v0.0.0-20150426214442-d788f35a0315
	github.com/Tencent/bk-bcs => ../../bk-bcs
	github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator => ../bcs-k8s/bcs-gamestatefulset-operator
	github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated => ../bcs-k8s/kubedeprecated
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes => ../bcs-k8s/kubernetes
	github.com/Tencent/bk-bcs/bcs-mesos/mesosv2 => ../bcs-mesos/mesosv2
	github.com/Tencent/bk-bcs/bcs-network => ./
	github.com/containernetworking/cni => github.com/containernetworking/cni v0.6.0
	github.com/containernetworking/plugins => github.com/containernetworking/plugins v0.6.0
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/go-logr/logr => github.com/go-logr/logr v0.1.0
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/Tencent/bk-bcs v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-mesos/mesosv2 v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go v1.23.0
	github.com/containernetworking/cni v0.6.0
	github.com/containernetworking/plugins v0.6.0
	github.com/coreos/go-iptables v0.4.3
	github.com/emicklei/go-restful v2.9.5+incompatible
	github.com/go-logr/logr v0.2.0
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.1.0
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.114+incompatible
	github.com/vishvananda/netlink v1.1.0
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/genproto v0.0.0-20200715011427-11fb19a81f2c
	google.golang.org/grpc v1.29.1
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.3
)
