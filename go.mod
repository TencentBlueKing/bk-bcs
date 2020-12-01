module github.com/Tencent/bk-bcs

go 1.14

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.0.0-20181126151915-b503174bad59
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181126155829-0cd23ebeb688
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20181126123746-eddba98df674
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181126152608-d082d5923d3c
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.1.9
)

require (
	github.com/DataDog/zstd v1.3.4 // indirect
	github.com/DeveloperJim/gokong v1.9.1-0.20200511122804-1c0ed1483353
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5
	github.com/NYTimes/gziphandler v0.0.0-20170623195520-56545f4a5d46 // indirect
	github.com/Shopify/sarama v1.20.0 // indirect
	github.com/Shopify/toxiproxy v2.1.4+incompatible // indirect
	github.com/andygrunwald/megos v0.0.0-20180424065632-0fccaea93714
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5 // indirect
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/aws/aws-sdk-go v1.23.0 // indirect
	github.com/bitly/go-simplejson v0.5.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b // indirect
	github.com/bwmarrin/discordgo v0.20.2 // indirect
	github.com/caddyserver/certmagic v0.10.6 // indirect
	github.com/chzyer/logex v1.1.10 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1 // indirect
	github.com/cloudflare/cloudflare-go v0.10.9 // indirect
	github.com/container-storage-interface/spec v0.3.0
	github.com/containernetworking/cni v0.6.0
	github.com/containernetworking/plugins v0.6.0 // indirect
	github.com/coredns/coredns v1.3.0
	github.com/coreos/bbolt v1.3.4 // indirect
	github.com/coreos/etcd v3.3.18+incompatible
	github.com/coreos/go-iptables v0.4.3 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964
	github.com/dbdd4us/qcloudapi-sdk-go v0.0.0-20190530123522-c8d9381de48c
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/deckarep/golang-set v1.7.1
	github.com/denisenkom/go-mssqldb v0.0.0-20200428022330-06a60b6afbbc // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dnstap/golang-dnstap v0.0.0-20170829151710-2cf77a2b5e11 // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20181223114339-d147fe0582f4+incompatible
	github.com/docker/engine-api v0.4.0
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/libnetwork v0.8.0-dev.2.0.20181207044626-1a06131fb8a0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/eapache/go-resiliency v1.1.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/ef-ds/deque v1.0.4-0.20190904040645-54cb57c252a1 // indirect
	github.com/elastic/beats v5.6.14+incompatible // indirect
	github.com/elastic/go-lumber v0.1.0 // indirect
	github.com/elastic/go-ucfg v0.8.3 // indirect
	github.com/elazarl/goproxy v0.0.0-20200426045556-49ad98f6dac1 // indirect
	github.com/emicklei/go-restful v2.9.5+incompatible
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/evanphx/json-patch/v5 v5.0.0 // indirect
	github.com/farsightsec/golang-framestream v0.0.0-20181102145529-8a0cb8ba8710 // indirect
	github.com/flynn/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/forestgiant/sliceutil v0.0.0-20160425183142-94783f95db6c // indirect
	github.com/frankban/quicktest v1.10.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/fsouza/go-dockerclient v1.6.0
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/go-openapi/analysis v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/loads v0.19.4 // indirect
	github.com/go-openapi/runtime v0.19.7 // indirect
	github.com/go-openapi/spec v0.19.4 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.16.0 // indirect
	github.com/go-sql-driver/mysql v1.4.1
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible // indirect
	github.com/gobwas/httphead v0.0.0-20180130184737-2c6c146eadee // indirect
	github.com/gobwas/pool v0.2.0 // indirect
	github.com/gobwas/ws v1.0.3 // indirect
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/gogo/protobuf v1.2.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.2
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/cadvisor v0.32.0
	github.com/google/go-cmp v0.5.0
	github.com/google/go-querystring v1.0.0
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.1
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.6 // indirect
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645 // indirect
	github.com/haproxytech/client-native v1.2.6
	github.com/haproxytech/models v1.2.4
	github.com/hashicorp/go-version v1.2.1 // indirect
	github.com/hpcloud/tail v1.0.1-0.20180514194441-a1dbeea552b7 // indirect
	github.com/iancoleman/strcase v0.0.0-20180726023541-3605ed457bf7
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jinzhu/now v1.1.1 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/json-iterator/go v1.1.10
	github.com/juju/ratelimit v1.0.1
	github.com/kevholditch/gokong v6.0.0+incompatible // indirect
	github.com/klauspost/compress v1.4.1 // indirect
	github.com/leodido/go-urn v1.1.0 // indirect
	github.com/lib/pq v1.3.0 // indirect
	github.com/lithammer/go-jump-consistent-hash v1.0.1
	github.com/lucas-clemente/quic-go v0.14.1 // indirect
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-sqlite3 v1.11.0
	github.com/mesos/mesos-go v0.0.10
	github.com/mholt/caddy v0.11.1
	github.com/micro/cli/v2 v2.1.2 // indirect
	github.com/micro/go-micro/v2 v2.9.1
	github.com/miekg/dns v1.1.27
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/moby/term v0.0.0-20200507201656-73f35e472e8f
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/nats-io/nats.go v1.9.2 // indirect
	github.com/netdata/go-orchestrator v0.0.0-20190905093727-c793edba0e8f // indirect
	github.com/nlopes/slack v0.6.1-0.20191106133607-d06c2a2b3249 // indirect
	github.com/nranchev/go-libGeoIP v0.0.0-20170629073846-d6d4a9a4c7e8 // indirect
	github.com/olekukonko/tablewriter v0.0.4 // indirect
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/opencontainers/runc v1.0.0-rc6.0.20181203215513-96ec2177ae84 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/parnurzeal/gorequest v0.2.16
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pborman/uuid v1.2.0
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.6.0
	github.com/samuel/go-zookeeper v0.0.0-20180130194729-c4fab1ac1bec
	github.com/satori/go.uuid v1.2.0
	github.com/serenize/snaker v0.0.0-20171204205717-a683aaf2d516 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/soheilhy/cmux v0.1.4 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.5.1
	github.com/stripe/stripe-go/v71 v71.28.0 // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.114+incompatible
	github.com/teris-io/shortid v0.0.0-20171029131806-771a37caa5cf // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20200427203606-3cfed13b9966 // indirect
	github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8
	github.com/urfave/cli v1.22.2
	github.com/vishvananda/netlink v1.0.0 // indirect
	github.com/vishvananda/netns v0.0.0-20190625233234-7109fa855b0f // indirect
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	github.com/xlab/treeprint v0.0.0-20181112141820-a009c3971eca // indirect
	go.etcd.io/bbolt v1.3.4 // indirect
	go.mongodb.org/mongo-driver v1.1.2 // indirect
	go.uber.org/zap v1.13.0 // indirect
	go4.org v0.0.0-20190313082347-94abd6928b1d
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2
	golang.org/x/sys v0.0.0-20200523222454-059865788121 // indirect
	google.golang.org/grpc v1.29.1
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/fsnotify/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.30.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
	gopkg.in/ory-am/dockertest.v3 v3.3.5 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	gopkg.in/telegram-bot-api.v4 v4.6.4 // indirect
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20200410163147-594e756bea31 // indirect
	k8s.io/kubernetes v1.14.10
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800 // indirect
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/testing_frameworks v0.1.1 // indirect
)
