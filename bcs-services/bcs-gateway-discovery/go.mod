module github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery

go 1.14

replace (
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210204122842-0a7ba2fb84ff
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/kevholditch/gokong => github.com/DeveloperJim/gokong v1.9.2
	github.com/micro/go-micro/v2 => github.com/OvertimeDog/go-micro/v2 v2.9.3
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210204084037-834463e85666
	github.com/google/uuid v1.2.0
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/kevholditch/gokong v0.0.0-00010101000000-000000000000
	github.com/micro/go-micro/v2 v2.9.1
)
