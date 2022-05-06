module github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bmsf-mesh

go 1.17

replace (
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210818040851-76fdc539dc33
	github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated => github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated v0.0.0-20210517125505-0f40c4b365cb
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	k8s.io/api v0.18.16
	k8s.io/apimachinery v0.18.16
	k8s.io/client-go v0.18.16
	sigs.k8s.io/controller-runtime v0.6.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible // indirect
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.4 // indirect
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.9.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.15.0 // indirect
	github.com/prometheus/procfs v0.2.0 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/ugorji/go/codec v1.2.3 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sys v0.0.0-20201214210602-f9fddec55a1e // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.0.1 // indirect
	google.golang.org/appengine v1.6.1 // indirect
	google.golang.org/protobuf v1.23.1-0.20200526195155-81db48ad09cc // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6 // indirect
	k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89 // indirect
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
