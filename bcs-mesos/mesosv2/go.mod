module github.com/Tencent/bk-bcs/bcs-mesos/mesosv2

go 1.13

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	k8s.io/apimachinery v0.18.5
	k8s.io/code-generator v0.18.5
	sigs.k8s.io/controller-runtime v0.6.0
)
