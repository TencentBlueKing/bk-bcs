module github.com/Tencent/bk-bcs/bmsf-mesh

go 1.14

replace (
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210517123645-82ef0026bf95
	github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated => github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated v0.0.0-20210517125505-0f40c4b365cb
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	k8s.io/api v0.18.16
	k8s.io/apimachinery v0.18.16
	k8s.io/client-go v0.18.16
	k8s.io/kubelet v0.18.16
	sigs.k8s.io/controller-runtime v0.6.0
)
