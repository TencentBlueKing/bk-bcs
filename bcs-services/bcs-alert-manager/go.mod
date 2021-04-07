module github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager

go 1.14

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210128033108-0471fd5e2976
	github.com/envoyproxy/protoc-gen-validate v0.1.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.4
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/micro/go-micro/v2 v2.9.1
	github.com/parnurzeal/gorequest v0.2.16
	github.com/prometheus/client_golang v1.9.0
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.33.1
	k8s.io/api v0.18.6
)
