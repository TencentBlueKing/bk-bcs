module github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery

go 1.14

replace (
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210517123645-82ef0026bf95
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.18+incompatible
	github.com/kevholditch/gokong => github.com/DeveloperJim/gokong v1.9.2
	github.com/micro/go-micro/v2 => github.com/OvertimeDog/go-micro/v2 v2.9.3
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.2.0
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/kevholditch/gokong v0.0.0-00010101000000-000000000000
	github.com/micro/go-micro/v2 v2.9.1
	github.com/prometheus/client_golang v1.9.0
	k8s.io/api v0.21.0 // indirect
)
