module github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network

go 1.24.0

replace (
	bitbucket.org/ww/goautoneg => github.com/adjust/goautoneg v0.0.0-20150426214442-d788f35a0315
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20230607093333-1f5cd2719e19
	github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated => github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated v0.0.0-20210117140338-aeaed29b1997
	github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes => ../kubernetes
	github.com/containernetworking/cni => github.com/containernetworking/cni v0.6.0
	github.com/containernetworking/plugins => github.com/containernetworking/plugins v0.6.0
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/golang/protobuf => github.com/golang/protobuf v1.4.3
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/api => google.golang.org/api v0.10.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.26.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.26.2
	k8s.io/client-go => k8s.io/client-go v0.26.2
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.90.1
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20221012153701-172d655c2280
	k8s.io/kubelet => k8s.io/kubelet v0.26.2
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20230607093333-1f5cd2719e19
	github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-mesos/mesosv2 v0.0.0-20210117140338-aeaed29b1997
	github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go v1.37.27
	github.com/aws/aws-sdk-go-v2 v1.13.0
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.16.0
	github.com/aws/smithy-go v1.10.0
	github.com/containernetworking/cni v1.1.2
	github.com/containernetworking/plugins v1.2.0
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/go-iptables v0.4.3
	github.com/emicklei/go-restful v2.16.0+incompatible
	github.com/fsnotify/fsnotify v1.6.0
	github.com/fsouza/go-dockerclient v1.7.3
	github.com/go-logr/logr v1.4.2
	github.com/golang/glog v1.2.4
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.4
	github.com/google/uuid v1.5.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.16.0
	github.com/tencentcloud/tencentcloud-sdk-go v1.0.132
	github.com/vishvananda/netlink v1.2.1-beta.2
	golang.org/x/oauth2 v0.30.0
	golang.org/x/sys v0.38.0
	golang.org/x/time v0.12.0 // indirect
	google.golang.org/api v0.149.0
	google.golang.org/genproto v0.0.0-20231211222908-989df2bf70f3 // indirect
	google.golang.org/grpc v1.79.3
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.26.2
	k8s.io/apimachinery v0.27.4
	k8s.io/client-go v0.26.2
	k8s.io/kubelet v0.26.2
	sigs.k8s.io/controller-runtime v0.14.6
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/containerd/containerd v1.7.29 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/docker v25.0.13+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/juju/ratelimit v1.0.1 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/micro/go-micro/v2 v2.9.1 // indirect
	github.com/miekg/dns v1.1.50 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/parnurzeal/gorequest v0.2.16 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/ugorji/go/codec v1.2.3 // indirect
	github.com/vishvananda/netns v0.0.0-20210104183010-2eb08e3e575f // indirect
	go.mongodb.org/mongo-driver v1.5.3 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/apiextensions-apiserver v0.26.1
	k8s.io/klog v1.0.0 // indirect
	k8s.io/klog/v2 v2.90.1 // indirect
	k8s.io/kube-openapi v0.0.0-20221012153701-172d655c2280 // indirect
	k8s.io/utils v0.0.0-20230220204549-a5ecb0141aa5 // indirect
	moul.io/http2curl v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.9.1
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.5.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork v1.1.0
	github.com/Azure/go-autorest/autorest/to v0.2.0
)

require (
	github.com/deckarep/golang-set v1.8.0
	github.com/hashicorp/go-multierror v1.1.0
	golang.org/x/sync v0.18.0
	google.golang.org/genproto/googleapis/api v0.0.0-20231120223509-83a465c0220f
)

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.5.1 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.2.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common v0.0.0-20220330120237-0bbed74dcf6d // indirect
	github.com/Tencent/bk-bcs/bcs-services/pkg v0.0.0-20220926153300-4e631deaebe4 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/emicklei/go-restful-openapi v1.4.1 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.19.3 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20200820163806-098863c1fc24 // indirect
	github.com/moby/patternmatcher v0.6.1 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/sys/user v0.3.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.45.0 // indirect
	go.opentelemetry.io/otel v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
	golang.org/x/term v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240401170217-c3f982113cda // indirect
)
