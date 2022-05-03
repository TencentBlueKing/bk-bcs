module github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/mesosv2

go 1.17

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20211020081756-b0c6bdef9dcb
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.2
	k8s.io/code-generator v0.18.5
	sigs.k8s.io/controller-runtime v0.6.0
)
