module github.com/Tencent/bk-bcs/bcs-services/bcs-logbeat-sidecar

go 1.14

replace (
	github.com/Tencent/bk-bcs => ../..
	github.com/Tencent/bk-bcs/bcs-services/bcs-logbeat-sidecar => ./
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	k8s.io/api => k8s.io/api v0.0.0-20181126151915-b503174bad59
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181126155829-0cd23ebeb688
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20181126123746-eddba98df674
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181126152608-d082d5923d3c
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.1.9
)

require (
	github.com/Tencent/bk-bcs v0.0.0-20200805130634-8a6c639f4a4c
	github.com/fsouza/go-dockerclient v1.6.5
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/client_model v0.2.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
)
