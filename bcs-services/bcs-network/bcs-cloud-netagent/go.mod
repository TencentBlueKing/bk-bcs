module github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netagent

go 1.14

replace (
	github.com/Tencent/bk-bcs => ../../../../bk-bcs
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes => ../../../bcs-k8s/kubernetes
	github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netagent => ./
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/Tencent/bk-bcs v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes v0.0.0-00010101000000-000000000000
	github.com/containernetworking/plugins v0.6.0
	github.com/prometheus/client_golang v1.0.0
	github.com/vishvananda/netlink v1.0.0
	golang.org/x/sys v0.0.0-20200519105757-fe76b779f299
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.2
)
