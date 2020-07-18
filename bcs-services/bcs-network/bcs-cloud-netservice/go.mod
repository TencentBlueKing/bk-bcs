module github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netservice

go 1.14

replace (
	github.com/Tencent/bk-bcs => ../../../../bk-bcs
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes => ../../../bcs-k8s/kubernetes
	github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netservice => ./
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/Tencent/bk-bcs v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes v0.0.0-00010101000000-000000000000
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/prometheus/client_golang v1.7.1
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.205+incompatible
	google.golang.org/grpc v1.30.0
	k8s.io/api v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.2
)
