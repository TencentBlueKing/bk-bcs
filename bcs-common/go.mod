module github.com/Tencent/bk-bcs/bcs-common

go 1.14

replace (
	github.com/Tencent/bk-bcs/bcs-common => ../bcs-common
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	golang.org/x/net v0.0.0-20211209124913-491a49abca63 => golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	google.golang.org/genproto v0.0.0-20210402141018-6c239bbf2bb1 => google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
)

require (
	github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common v0.0.0-20220330120237-0bbed74dcf6d
	github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/kubebkbcsv2 v0.0.0-20220330120237-0bbed74dcf6d
	github.com/TencentBlueKing/iam-go-sdk v0.0.8
	github.com/asim/go-micro/plugins/broker/rabbitmq/v3 v3.7.0
	github.com/asim/go-micro/plugins/broker/stan/v3 v3.7.0
	github.com/asim/go-micro/plugins/registry/etcd/v3 v3.7.0
	github.com/asim/go-micro/v3 v3.7.1
	github.com/bitly/go-simplejson v0.5.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/engine-api v0.4.0
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/emicklei/go-restful-openapi v1.4.1
	github.com/envoyproxy/protoc-gen-validate v0.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/golang/mock v1.5.0
	github.com/golang/protobuf v1.5.2
	github.com/google/go-querystring v1.0.0
	github.com/google/uuid v1.2.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/juju/ratelimit v1.0.1
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/nats-io/stan.go v0.9.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.1
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.1
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	github.com/ugorji/go/codec v1.2.3
	go.etcd.io/etcd/client/v3 v3.5.3
	go.mongodb.org/mongo-driver v1.5.3
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.28.0
	go.opentelemetry.io/otel v1.6.3
	go.opentelemetry.io/otel/exporters/jaeger v1.6.3
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.6.3
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.6.3
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.6.3
	go.opentelemetry.io/otel/exporters/prometheus v0.26.0
	go.opentelemetry.io/otel/metric v0.26.0
	go.opentelemetry.io/otel/sdk v1.6.3
	go.opentelemetry.io/otel/sdk/export/metric v0.26.0
	go.opentelemetry.io/otel/sdk/metric v0.26.0
	go.opentelemetry.io/otel/trace v1.6.3
	golang.org/x/net v0.0.0-20211209124913-491a49abca63
	google.golang.org/genproto v0.0.0-20211118181313-81c1377c94b1
	google.golang.org/grpc v1.45.0
	k8s.io/api v0.23.1
	k8s.io/apimachinery v0.23.1
	k8s.io/client-go v0.23.1
	k8s.io/klog v1.0.0
)
