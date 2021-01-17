module github.com/Tencent/bk-bcs/bcs-services/bcs-logbeat-sidecar

go 1.14

replace (
	bitbucket.org/ww/goautoneg => github.com/adjust/goautoneg v0.0.0-20150426214442-d788f35a0315
	github.com/Tencent/bk-bcs/bcs-common => ../../../bk-bcs/bcs-common
	github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator => ../../bcs-k8s/bcs-gamedeployment-operator
	github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator => ../../bcs-k8s/bcs-gamestatefulset-operator
	github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-watch => ../../bcs-k8s/bcs-k8s-watch
	github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs => ../../bcs-k8s/kubebkbcs
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes => ../../bcs-k8s/kubernetes
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common => ../../bcs-k8s/kubernetes/common
	github.com/Tencent/bk-bcs/bcs-services/bcs-logbeat-sidecar => ./
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0 // indirect
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	k8s.io/api => k8s.io/api v0.0.0-20181126151915-b503174bad59
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181126155829-0cd23ebeb688
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20181126123746-eddba98df674
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181126152608-d082d5923d3c
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20181109181836-c59034cc13d5
	k8s.io/kubernetes => k8s.io/kubernetes v1.13.1
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.1.9
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs v0.0.0-00010101000000-000000000000
	github.com/containerd/containerd v1.4.3 // indirect
	github.com/docker/docker v20.10.0-rc1+incompatible // indirect
	github.com/fsouza/go-dockerclient v1.6.5
	github.com/googleapis/gnostic v0.0.0-00010101000000-000000000000 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/client_model v0.2.0
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools/v3 v3.0.3 // indirect
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
)
