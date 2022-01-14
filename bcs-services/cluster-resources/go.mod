module github.com/Tencent/bk-bcs/bcs-services/cluster-resources

go 1.14

replace (
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210525083026-bc8c14258b6e // v0.20.16
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/klauspost/compress => github.com/klauspost/compress v1.13.0
	github.com/micro/go-micro/v2 => github.com/OvertimeDog/go-micro/v2 v2.9.3
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.20.4
	k8s.io/client-go => k8s.io/client-go v0.20.4
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-00010101000000-000000000000
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/alicebob/miniredis v2.5.0+incompatible
	github.com/coreos/etcd v3.3.25+incompatible // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/elazarl/go-bindata-assetfs v1.0.1
	github.com/envoyproxy/protoc-gen-validate v0.6.2
	github.com/go-redis/cache/v8 v8.4.3
	github.com/go-redis/redis/v8 v8.11.4
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.5.2
	github.com/gomodule/redigo v1.8.6 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.1.4
	github.com/googleapis/gnostic v0.4.1
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/micro/go-micro/v2 v2.9.1
	github.com/stretchr/testify v1.7.0
	github.com/yuin/gopher-lua v0.0.0-20210529063254-f4c35e4016d9 // indirect
	go.uber.org/zap v1.19.1
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	golang.org/x/tools v0.1.8 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.33.1
	google.golang.org/protobuf v1.27.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v11.0.0+incompatible
)
