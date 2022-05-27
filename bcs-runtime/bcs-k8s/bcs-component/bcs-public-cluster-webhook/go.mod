module github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-public-cluster-webhook

go 1.14

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20211112093246-b1af784416d0
	github.com/prometheus/client_golang v1.9.0
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.9.1
	gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0
	k8s.io/api v0.22.3
	k8s.io/apimachinery v0.22.3
)
