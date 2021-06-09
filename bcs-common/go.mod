module github.com/Tencent/bk-bcs/bcs-common

go 1.14

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/micro/go-micro/v2 => github.com/OvertimeDog/go-micro/v2 v2.9.3
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/bitly/go-simplejson v0.5.0
	github.com/coreos/bbolt v1.3.4 // indirect
	github.com/coreos/etcd v3.3.18+incompatible
	github.com/docker/engine-api v0.4.0
	github.com/elazarl/goproxy v0.0.0-20210110162100-a92cc753f88e // indirect
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/envoyproxy/protoc-gen-validate v0.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/mock v1.3.1
	github.com/golang/protobuf v1.4.3
	github.com/google/go-querystring v1.0.0
	github.com/google/uuid v1.1.4
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/grpc-gateway v1.9.5
	github.com/juju/ratelimit v1.0.1
	github.com/micro/go-micro/v2 v2.9.1
	github.com/micro/go-plugins/broker/rabbitmq/v2 v2.9.1
	github.com/micro/go-plugins/broker/stan/v2 v2.9.1
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/nats-io/stan.go v0.8.2
	github.com/opentracing/opentracing-go v1.1.0
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414
	github.com/spf13/pflag v1.0.5
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	github.com/ugorji/go/codec v1.2.3
	go.mongodb.org/mongo-driver v1.4.5
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.27.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	moul.io/http2curl v1.0.0 // indirect
)
