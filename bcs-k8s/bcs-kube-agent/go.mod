module github.com/Tencent/bk-bcs/bcs-k8s/bcs-kube-agent

go 1.14

replace (
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210201062319-0b02c6d040c6
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	k8s.io/api => k8s.io/api v0.18.6
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-00010101000000-000000000000
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/json-iterator/go v1.1.10
	github.com/parnurzeal/gorequest v0.2.16
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/prometheus/client_golang v1.9.0
	github.com/spf13/cobra v1.1.1 // indirect
	github.com/spf13/viper v1.7.1
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	moul.io/http2curl v1.0.0 // indirect
)
