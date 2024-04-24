module github.com/TencentBlueKing/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-external-privilege

go 1.14

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/mholt/caddy => github.com/caddyserver/caddy v0.11.1
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210818040851-76fdc539dc33
