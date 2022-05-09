module github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-multi-ns-proxy

go 1.16

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.7
	k8s.io/client-go => k8s.io/client-go v0.22.7
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20220309021702-fe8a9f6843e1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	go.uber.org/zap v1.21.0
	k8s.io/apimachinery v0.23.4
	k8s.io/apiserver v0.23.4
	k8s.io/client-go v0.23.4
)
