module github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy

go 1.14

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210621082136-e7b1aa4848c4
	github.com/gorilla/mux v1.8.0
	github.com/lithammer/dedent v1.1.0
	github.com/moby/ipvs v1.0.1
	github.com/prometheus/client_golang v1.9.0
	github.com/spf13/viper v1.8.1
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
)
