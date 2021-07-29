module github.com/Tencent/bk-bcs/bcs-services/bcs-logbeat-sidecar

go 1.14

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210518090424-99527484a283
	github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs v0.0.0-20210518090424-99527484a283
	github.com/containerd/containerd v1.4.3 // indirect
	github.com/docker/docker v20.10.0-rc1+incompatible // indirect
	github.com/fsouza/go-dockerclient v1.6.5
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/client_model v0.2.0
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools/v3 v3.0.3 // indirect
	k8s.io/api v0.18.5
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.2
)
