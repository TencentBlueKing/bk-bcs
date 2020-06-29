module bk-bcs

go 1.14

replace (
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes => ./bcs-k8s/kubernetes
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/DataDog/zstd v1.3.4 // indirect
	github.com/DeveloperJim/gokong v1.9.1-0.20200511122804-1c0ed1483353
	github.com/Microsoft/go-winio v0.4.11
	github.com/Shopify/sarama v1.20.0 // indirect
	github.com/Shopify/toxiproxy v2.1.4+incompatible // indirect
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes v0.0.0-00010101000000-000000000000
	github.com/andygrunwald/megos v0.0.0-20180424065632-0fccaea93714
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5 // indirect
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/aws/aws-sdk-go v1.16.11
	github.com/bitly/go-simplejson v0.5.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/container-storage-interface/spec v0.3.0
	github.com/containernetworking/cni v0.6.0
	github.com/containernetworking/plugins v0.6.0
	github.com/coredns/coredns v1.3.0
	github.com/coreos/bbolt v1.3.4 // indirect
	github.com/coreos/etcd v3.3.10+incompatible
	github.com/coreos/go-iptables v0.4.3
	github.com/coreos/go-systemd v0.0.0-20190620071333-e64a0ec8b42a // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964
	github.com/dbdd4us/qcloudapi-sdk-go v0.0.0-20190530123522-c8d9381de48c
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/deckarep/golang-set v1.7.1
	github.com/denisenkom/go-mssqldb v0.0.0-20200428022330-06a60b6afbbc // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dnstap/golang-dnstap v0.0.0-20170829151710-2cf77a2b5e11 // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20181223114339-d147fe0582f4+incompatible
	github.com/docker/engine-api v0.4.0
	github.com/docker/libnetwork v0.8.0-dev.2.0.20181207044626-1a06131fb8a0 // indirect
	github.com/eapache/go-resiliency v1.1.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/elastic/beats v5.6.14+incompatible
	github.com/elastic/go-lumber v0.1.0 // indirect
	github.com/elastic/go-ucfg v0.6.5 // indirect
	github.com/elazarl/goproxy v0.0.0-20200426045556-49ad98f6dac1 // indirect
	github.com/emicklei/go-restful v2.9.5+incompatible
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/farsightsec/golang-framestream v0.0.0-20181102145529-8a0cb8ba8710 // indirect
	github.com/flynn/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/frankban/quicktest v1.10.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/fsouza/go-dockerclient v1.3.1
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/go-openapi/analysis v0.19.6 // indirect
	github.com/go-openapi/runtime v0.19.7 // indirect
	github.com/go-openapi/spec v0.19.4 // indirect
	github.com/go-playground/locales v0.12.1 // indirect
	github.com/go-playground/universal-translator v0.16.0 // indirect
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/google/cadvisor v0.32.0
	github.com/google/go-cmp v0.4.0
	github.com/google/go-querystring v1.0.0
	github.com/gorilla/mux v1.6.2
	github.com/gorilla/websocket v1.4.0
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.5 // indirect
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645 // indirect
	github.com/haproxytech/client-native v1.2.6
	github.com/haproxytech/models v1.2.4
	github.com/iancoleman/strcase v0.0.0-20180726023541-3605ed457bf7
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jinzhu/now v1.1.1 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/json-iterator/go v1.1.8
	github.com/juju/ratelimit v1.0.1
	github.com/kevholditch/gokong v6.0.0+incompatible // indirect
	github.com/klauspost/compress v1.4.1 // indirect
	github.com/klauspost/cpuid v1.2.0 // indirect
	github.com/leodido/go-urn v1.1.0 // indirect
	github.com/lithammer/go-jump-consistent-hash v1.0.1
	github.com/mattn/go-sqlite3 v1.11.0
	github.com/mesos/mesos-go v0.0.10
	github.com/mholt/caddy v0.11.1
	github.com/miekg/dns v1.1.1
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/moby/term v0.0.0-20200507201656-73f35e472e8f
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/nranchev/go-libGeoIP v0.0.0-20170629073846-d6d4a9a4c7e8 // indirect
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/opencontainers/runc v1.0.0-rc6.0.20181203215513-96ec2177ae84 // indirect
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pborman/uuid v1.2.0
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.4.1
	github.com/rcrowley/go-metrics v0.0.0-20170417060359-4b2ddcb0bf13 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20180130194729-c4fab1ac1bec
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.4.0
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.114+incompatible
	github.com/tmc/grpc-websocket-proxy v0.0.0-20200427203606-3cfed13b9966 // indirect
	github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8
	github.com/urfave/cli v1.20.0
	github.com/vishvananda/netlink v1.0.0
	github.com/vishvananda/netns v0.0.0-20190625233234-7109fa855b0f // indirect
	go.etcd.io/bbolt v1.3.4 // indirect
	go4.org v0.0.0-20190313082347-94abd6928b1d
	golang.org/x/net v0.0.0-20200520004742-59133d7f0dd7
	golang.org/x/sys v0.0.0-20200519105757-fe76b779f299
	google.golang.org/grpc v1.26.0
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.24.0
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
	gopkg.in/ory-am/dockertest.v3 v3.3.5 // indirect
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.2
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.2
	k8s.io/code-generator v0.18.5
	k8s.io/kubernetes v1.14.10
	sigs.k8s.io/controller-runtime v0.6.0
)
