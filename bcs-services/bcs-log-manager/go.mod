module github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager

go 1.14

replace (
	github.com/Tencent/bk-bcs => ../../
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.0.0-20181126151915-b503174bad59
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181126155829-0cd23ebeb688
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20181126123746-eddba98df674
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181126152608-d082d5923d3c
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.2.0
)

require (
	github.com/Tencent/bk-bcs v0.0.0-00010101000000-000000000000
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/micro/go-micro/v2 v2.9.1
	github.com/prometheus/client_golang v1.7.1
	golang.org/x/sys v0.0.0-20200918174421-af09f7315aff // indirect
	google.golang.org/genproto v0.0.0-20200513103714-09dca8ec2884
	google.golang.org/grpc v1.29.1
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.6.3
)
