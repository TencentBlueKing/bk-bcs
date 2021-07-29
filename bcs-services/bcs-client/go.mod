module github.com/Tencent/bk-bcs/bcs-services/bcs-client

go 1.14

replace (
	bitbucket.org/ww/goautoneg => github.com/adjust/goautoneg v0.0.0-20150426214442-d788f35a0315
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210625040556-0385f88cbfd6
	github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator => ../../bcs-k8s/bcs-gamedeployment-operator
	github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator => ../../bcs-k8s/bcs-gamestatefulset-operator
	github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs => github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs v0.0.0-20210117140338-aeaed29b1997
	github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager => ../bcs-log-manager
	github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager => ../bcs-mesh-manager
	github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager => ../bcs-user-manager
	github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server => ../../bcs-services/bcs-webhook-server
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/openshift/api => github.com/openshift/api v0.0.0-20180801171038-322a19404e37
	github.com/ugorji/go v1.1.4 => github.com/ugorji/go v0.0.0-20181204163529-d75b2dcb6bc8
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
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
	contrib.go.opencensus.io/exporter/prometheus v0.2.0 // indirect
	fortio.org/fortio v1.6.3 // indirect
	github.com/DATA-DOG/go-sqlmock v1.4.1 // indirect
	github.com/Masterminds/sprig/v3 v3.1.0 // indirect
	github.com/Masterminds/squirrel v1.2.0 // indirect
	github.com/Tencent/bk-bcs v1.20.11
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210223080803-f27f3f3c01c4
	github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager v0.0.0-00010101000000-000000000000
	github.com/bitly/go-simplejson v0.5.0
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/cheggaaa/pb/v3 v3.0.4 // indirect
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/containernetworking/cni v0.7.0-alpha1 // indirect
	github.com/containernetworking/plugins v0.7.3 // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible // indirect
	github.com/d4l3k/messagediff v1.2.1 // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20181223114339-d147fe0582f4+incompatible
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/envoyproxy/go-control-plane v0.9.7-0.20200811182123-112a4904c4b0 // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/sync v0.0.0-20180314180146-1d60e4601c6f // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/hashicorp/vault/api v1.0.3 // indirect
	github.com/howeyc/fsnotify v0.9.0 // indirect
	github.com/jmoiron/sqlx v1.2.0 // indirect
	github.com/kubernetes-client/go v0.0.0-20200222171647-9dac5e4c5400 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lestrrat-go/jwx v1.0.3 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mattn/go-shellwords v1.0.10 // indirect
	github.com/mholt/archiver v3.1.1+incompatible // indirect
	github.com/micro/go-micro/v2 v2.9.1 // indirect
	github.com/miekg/dns v1.1.30 // indirect
	github.com/moby/term v0.0.0-20200611042045-63b9a826fb74
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/tsdb v0.7.1 // indirect
	github.com/rubenv/sql-migrate v0.0.0-20200212082348-64f95ea68aa3 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/ulikunitz/xz v0.5.7 // indirect
	github.com/urfave/cli v1.22.4
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	github.com/yl2chen/cidranger v1.0.0 // indirect
	go.uber.org/zap v1.15.0 // indirect
	google.golang.org/grpc v1.31.1
	google.golang.org/grpc/examples v0.0.0-20200818224027-0f73133e3aa3 // indirect
	gopkg.in/d4l3k/messagediff.v1 v1.2.1 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	istio.io/api v0.0.0-20200819225923-c78f387f78a2 // indirect
	istio.io/client-go v0.0.0-20200812230733-f5504d568313 // indirect
	istio.io/gogo-genproto v0.0.0-20200720193312-b523a30fe746 // indirect
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/code-generator v0.19.0 // indirect
	sigs.k8s.io/service-apis v0.0.0-20200731055707-56154e7bfde5 // indirect
)
