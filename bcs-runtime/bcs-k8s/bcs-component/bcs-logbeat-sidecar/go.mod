module github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-logbeat-sidecar

go 1.14

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/Microsoft/go-winio v0.4.15-0.20200113171025-3fe6c5262873 // indirect
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210908080357-99540f892332
	github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs v0.0.0-20210810131039-5220f346d815
	github.com/containerd/containerd v1.4.3 // indirect
	github.com/docker/docker v20.10.12+incompatible
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/client_model v0.2.0
	google.golang.org/grpc v1.27.1 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.5
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.2
)
