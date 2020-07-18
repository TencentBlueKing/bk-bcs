module github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netcontroller

go 1.14

replace (
	github.com/Tencent/bk-bcs => ../../../../bk-bcs
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes => ../../../bcs-k8s/kubernetes
	github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netcontroller => ./
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/go-logr/logr => github.com/go-logr/logr v0.1.0
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/Tencent/bk-bcs v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go v1.16.11
	github.com/go-logr/logr v0.2.0
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.114+incompatible
	github.com/vishvananda/netlink v1.0.0
	google.golang.org/grpc v1.26.0
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.2
	sigs.k8s.io/controller-runtime v0.6.0
)
