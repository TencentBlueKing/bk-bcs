module github.com/Tencent/bk-bcs/bcs-common/test

go 1.17

replace (
	github.com/Tencent/bk-bcs/bcs-common => ../../../../bcs-common
	github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2 => github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2 v0.0.0-20220221082336-47ffb7ebf0b5
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/go-logr/logr => github.com/go-logr/logr v1.2.0
	github.com/haproxytech/client-native => github.com/abstractmj/client-native v1.2.8
	github.com/mholt/caddy => github.com/caddyserver/caddy v0.11.1
	github.com/micro/go-micro/v2 => github.com/OvertimeDog/go-micro/v2 v2.9.3
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210818040851-76fdc539dc33
