module github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network

go 1.17

replace (
	bitbucket.org/ww/goautoneg => github.com/adjust/goautoneg v0.0.0-20150426214442-d788f35a0315
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20230607093333-1f5cd2719e19
	github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated => github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated v0.0.0-20210117140338-aeaed29b1997
	github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes => ../kubernetes
	github.com/containernetworking/cni => github.com/containernetworking/cni v0.6.0
	github.com/containernetworking/plugins => github.com/containernetworking/plugins v0.6.0
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/go-logr/logr => github.com/go-logr/logr v0.1.0
	github.com/golang/protobuf => github.com/golang/protobuf v1.4.3
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/api => google.golang.org/api v0.10.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.18.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.6
	k8s.io/client-go => k8s.io/client-go v0.18.6
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.0.0
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	k8s.io/kubelet => k8s.io/kubelet v0.18.6
)

require (
	cloud.google.com/go/compute v1.5.0 // indirect
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20230607093333-1f5cd2719e19
	github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-mesos/mesosv2 v0.0.0-20210117140338-aeaed29b1997
	github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go v1.37.27
	github.com/aws/aws-sdk-go-v2 v1.13.0
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.16.0
	github.com/aws/smithy-go v1.10.0
	github.com/containernetworking/cni v0.6.0
	github.com/containernetworking/plugins v0.6.0
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/go-iptables v0.4.3
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/fsnotify/fsnotify v1.5.4
	github.com/fsouza/go-dockerclient v1.7.3
	github.com/go-logr/logr v1.2.3
	github.com/golang/glog v1.0.0
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.11.0
	github.com/tencentcloud/tencentcloud-sdk-go v1.0.132
	github.com/vishvananda/netlink v1.1.0
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	golang.org/x/sys v0.0.0-20220909162455-aba9fc2a8ff2
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/api v0.70.0
	google.golang.org/genproto v0.0.0-20220222213610-43724f9ea8cf
	google.golang.org/grpc v1.46.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.23.1
	k8s.io/apimachinery v0.23.1
	k8s.io/client-go v0.23.1
	k8s.io/kubelet v0.18.6
	sigs.k8s.io/controller-runtime v0.6.3
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Microsoft/hcsshim v0.8.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/containerd/containerd v1.4.4 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/docker v20.10.7+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/zapr v0.1.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gax-go/v2 v2.1.1 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/juju/ratelimit v1.0.1 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/micro/go-micro/v2 v2.9.1 // indirect
	github.com/miekg/dns v1.1.50 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v1.1.4 // indirect
	github.com/parnurzeal/gorequest v0.2.16 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/ugorji/go/codec v1.2.3 // indirect
	github.com/vishvananda/netns v0.0.0-20191106174202-0a2b9b5464df // indirect
	go.mongodb.org/mongo-driver v1.5.3 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/crypto v0.0.0-20220829220503-c86fa9a7ed90 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/net v0.0.0-20220909164309-bea034e7d591 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.12 // indirect
	gomodules.xyz/jsonpatch/v2 v2.0.1 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/apiextensions-apiserver v0.20.0
	k8s.io/klog v1.0.0 // indirect
	k8s.io/klog/v2 v2.30.0 // indirect
	k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65 // indirect
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b // indirect
	moul.io/http2curl v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.0.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.0.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork v1.0.0
	github.com/Azure/go-autorest/autorest/to v0.2.0
)

require (
	github.com/deckarep/golang-set v1.8.0
	github.com/hashicorp/go-multierror v1.1.0
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.0.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v0.4.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common v0.0.0-20220330120237-0bbed74dcf6d // indirect
	github.com/Tencent/bk-bcs/bcs-services/pkg v0.0.0-20220926153300-4e631deaebe4 // indirect
	github.com/containerd/cgroups v0.0.0-20200531161412-0dbf7f05ba59 // indirect
	github.com/emicklei/go-restful-openapi v1.4.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/spec v0.19.3 // indirect
	github.com/go-openapi/swag v0.19.5 // indirect
	github.com/golang-jwt/jwt v3.2.1+incompatible // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20200820163806-098863c1fc24 // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.5.0 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/pkg/browser v0.0.0-20210115035449-ce105d075bb4 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
)
