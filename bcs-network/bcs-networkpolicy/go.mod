module github.com/Tencent/bk-bcs/bcs-network/bcs-networkpolicy

go 1.14

replace (
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210517123645-82ef0026bf95
	github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated => github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated v0.0.0-20210117140338-aeaed29b1997
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes => github.com/Tencent/bk-bcs/bcs-k8s/kubernetes v0.0.0-20210421085141-bce3cfad89cc
	github.com/containernetworking/cni => github.com/containernetworking/cni v0.6.0
	github.com/containernetworking/plugins => github.com/containernetworking/plugins v0.6.0
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/go-logr/logr => github.com/go-logr/logr v0.1.0
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
	k8s.io/api => k8s.io/api v0.18.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.6
	k8s.io/client-go => k8s.io/client-go v0.18.6
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated v0.0.0-00010101000000-000000000000 // indirect
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes v0.0.0-00010101000000-000000000000 // indirect
	github.com/Tencent/bk-bcs/bcs-mesos/mesosv2 v0.0.0-20210117140338-aeaed29b1997
	github.com/aws/aws-sdk-go v1.34.28 // indirect
	github.com/containernetworking/cni v0.6.0 // indirect
	github.com/containernetworking/plugins v0.6.0 // indirect
	github.com/coreos/go-iptables v0.4.3
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/fsouza/go-dockerclient v1.6.0
	github.com/go-logr/logr v0.2.0 // indirect
	github.com/golang/mock v1.4.4 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.6 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/prometheus/client_golang v1.9.0
	github.com/tencentcloud/tencentcloud-sdk-go v1.0.132 // indirect
	github.com/vishvananda/netlink v1.1.0 // indirect
	golang.org/x/sys v0.0.0-20201214210602-f9fddec55a1e // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/genproto v0.0.0-20200715011427-11fb19a81f2c // indirect
	google.golang.org/grpc v1.31.0 // indirect
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.3 // indirect
)
