module github.com/Tencent/bk-bcs/bcs-services/bcs-client

go 1.14

replace (
	bitbucket.org/ww/goautoneg => github.com/adjust/goautoneg v0.0.0-20150426214442-d788f35a0315
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20220329091816-5b868e90d386
	github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator => ../../bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator
	github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator => ../../bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator
	github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs => github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs v0.0.0-20210117140338-aeaed29b1997
	github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager => ../bcs-log-manager
	github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager => ../bcs-mesh-manager
	github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager => ../bcs-user-manager
	github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server => ../bcs-webhook-server
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/mholt/caddy => github.com/caddyserver/caddy v0.11.1
	github.com/openshift/api => github.com/openshift/api v0.0.0-20180801171038-322a19404e37
	github.com/tencentcloud/tencentcloud-sdk-go => github.com/tencentcloud/tencentcloud-sdk-go v1.0.132
	github.com/ugorji/go v1.1.4 => github.com/ugorji/go v0.0.0-20181204163529-d75b2dcb6bc8
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	golang.org/x/net => golang.org/x/net v0.0.0-20210119194325-5f4716e94777
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	istio.io/istio => istio.io/istio v0.0.0-20200812220246-25bea56c0eb0
	k8s.io/api => k8s.io/api v0.0.0-20181126151915-b503174bad59
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181126155829-0cd23ebeb688
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20181126123746-eddba98df674
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181126152608-d082d5923d3c
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	k8s.io/kubectl => k8s.io/kubectl v0.16.15
	k8s.io/kubernetes => k8s.io/kubernetes v1.13.1
)

require (
	github.com/Tencent/bk-bcs v1.20.11 // indirect
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20220329091816-5b868e90d386
	github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager v0.0.0-00010101000000-000000000000
	github.com/bitly/go-simplejson v0.5.0
	github.com/docker/docker v17.12.0-ce-rc1.0.20181223114339-d147fe0582f4+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.2
	github.com/miekg/dns v1.1.30 // indirect
	github.com/moby/term v0.0.0-20200611042045-63b9a826fb74
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli v1.22.4
	google.golang.org/grpc v1.41.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/apiextensions-apiserver v0.18.6
)
