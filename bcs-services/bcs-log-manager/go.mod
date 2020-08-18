module github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager

go 1.14

replace (
	github.com/Tencent/bk-bcs => ../../
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes => ../../bcs-k8s/kubernetes/
	github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager => ./
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/Tencent/bk-bcs v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes v0.0.0-00010101000000-000000000000
	k8s.io/api v0.18.8 // indirect
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.0+incompatible
)
