module github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/istio-policy-controller

go 1.23.0

replace (
	bitbucket.org/ww/goautoneg => github.com/adjust/goautoneg v0.0.0-20150426214442-d788f35a0315
	github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common => github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common v0.0.0-20220117082205-1fdc9e155811
	github.com/Tencent/bk-bcs/bcs-scenarios/kourse => github.com/TencentBlueKing/bk-bcs/bcs-scenarios/kourse v0.0.0-20230620070606-922d3a7a9517
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/mholt/caddy => github.com/caddyserver/caddy v0.11.1
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.26.15
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.26.15
	k8s.io/apiserver => k8s.io/apiserver v0.20.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.0
	k8s.io/client-go => k8s.io/client-go v0.26.15
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.0
	k8s.io/code-generator => k8s.io/code-generator v0.20.0
	k8s.io/component-base => k8s.io/component-base v0.20.0
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.0
	k8s.io/controller-manager => k8s.io/controller-manager v0.20.0
	k8s.io/cri-api => k8s.io/cri-api v0.20.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.0
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.0
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20220603121420-31174f50af60
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.0
	k8s.io/kubectl => k8s.io/kubectl v0.20.0
	k8s.io/kubelet => k8s.io/kubelet v0.20.0
	k8s.io/kubernetes => k8s.io/kubernetes v1.20.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.0
	k8s.io/metrics => k8s.io/metrics v0.20.0
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.0
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20230707084844-155f32c8f606
	github.com/prometheus/client_golang v1.16.0
	k8s.io/api v0.32.1
	k8s.io/apiextensions-apiserver v0.26.10 // indirect
	k8s.io/apimachinery v0.32.1
	k8s.io/client-go v0.32.1
)

require github.com/pkg/errors v0.9.1 // indirect

require (
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
)

require (
	github.com/go-logr/logr v1.4.2
	istio.io/api v1.28.2-0.20251205082437-fde1452f70bc
	istio.io/client-go v1.28.3
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/oauth2 v0.23.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/term v0.34.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	google.golang.org/protobuf v1.36.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/kube-openapi v0.0.0-20241105132330-32ad38e42d3f // indirect
	sigs.k8s.io/controller-runtime v0.14.7
	sigs.k8s.io/structured-merge-diff/v4 v4.7.0 // indirect
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250811230008-5f3141c8851a // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/component-base v0.26.10 // indirect
	sigs.k8s.io/json v0.0.0-20241010143419-9aa6b5e7a4b3 // indirect
)
