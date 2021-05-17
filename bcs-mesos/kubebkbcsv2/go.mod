module github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2

go 1.14

replace (
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common 82ef0026bf95
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/onsi/ginkgo v1.13.0 // indirect
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.2
	k8s.io/code-generator v0.18.5
	sigs.k8s.io/controller-runtime v0.6.0
)
