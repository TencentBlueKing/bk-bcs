module github.com/Tencent/bk-bcs/bcs-services/bcs-monitor

go 1.17

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20230202032803-47379be3c4c1
	github.com/Tencent/bk-bcs/bcs-services/pkg v0.0.0-20230202032803-47379be3c4c1
	github.com/TencentBlueKing/bkmonitor-kits v0.2.0
	github.com/chonla/format v0.0.0-20220105105701-1119f4a3f36f
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/fsnotify/fsnotify v1.5.4
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-contrib/requestid v0.0.4
	github.com/gin-contrib/sse v0.1.0
	github.com/gin-gonic/gin v1.7.7
	github.com/go-kit/log v0.2.0
	github.com/go-micro/plugins/v4/registry/etcd v1.1.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-resty/resty/v2 v2.7.0
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/hashicorp/go-version v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/oklog/run v1.1.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.46.0
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.46.0
	github.com/prometheus/client_golang v1.12.1
	github.com/prometheus/common v0.34.0
	github.com/prometheus/prometheus v1.8.2-0.20220308163432-03831554a519
	github.com/spf13/cobra v1.4.0
	github.com/spf13/viper v1.11.0
	github.com/stretchr/testify v1.7.1
	github.com/swaggo/files v0.0.0-20210815190702-a29dd2bc99b2
	github.com/swaggo/gin-swagger v1.4.3
	github.com/swaggo/swag v1.8.1
	github.com/thanos-io/thanos v0.26.0
	go-micro.dev/v4 v4.8.1
	go.uber.org/automaxprocs v1.5.1
	golang.org/x/sync v0.0.0-20220907140024-f12130a52804
	google.golang.org/grpc v1.47.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v0.23.5
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.60.1
)

require (
	cloud.google.com/go/compute v1.5.0 // indirect
	cloud.google.com/go/trace v0.1.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.0.0 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20220824120805-4b6e5c587895 // indirect
	github.com/TencentBlueKing/iam-go-sdk v0.0.8 // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.42.31 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cloudflare/circl v1.2.0 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/edsrzf/mmap-go v1.1.0 // indirect
	github.com/efficientgo/tools/extkingpin v0.0.0-20210609125236-d73259166f20 // indirect
	github.com/elastic/go-sysinfo v1.1.1 // indirect
	github.com/elastic/go-windows v1.0.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/facette/natsort v0.0.0-20181210072756-2cd4dd1e2dcb // indirect
	github.com/felixge/fgprof v0.9.1 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/go-git/go-git/v5 v5.4.2 // indirect
	github.com/go-kit/kit v0.12.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gogo/status v1.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20211214055906-6f57359322fd // indirect
	github.com/googleapis/gax-go/v2 v2.3.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/providers/kit/v2 v2.0.0-20201002093600-73cf2ae9d891 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0-rc.2.0.20201207153454-9f6bf00c00a7 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lightstep/lightstep-tracer-common/golang/gogo v0.0.0-20190605223551-bc2310a04743 // indirect
	github.com/lightstep/lightstep-tracer-go v0.18.1 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/miekg/dns v1.1.50 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opentracing-contrib/go-stdlib v1.0.0 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/parnurzeal/gorequest v0.2.16 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/exporter-toolkit v0.7.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/santhosh-tekuri/jsonschema v1.2.4 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/smartystreets/goconvey v1.7.2 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/uber/jaeger-client-go v2.30.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
	github.com/vimeo/galaxycache v0.0.0-20210323154928-b7e5d71c067a // indirect
	github.com/xanzy/ssh-agent v0.3.2 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.elastic.co/apm v1.11.0 // indirect
	go.elastic.co/apm/module/apmhttp v1.11.0 // indirect
	go.elastic.co/apm/module/apmot v1.11.0 // indirect
	go.elastic.co/fastjson v1.1.0 // indirect
	go.etcd.io/etcd/api/v3 v3.5.4 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.4 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.4.0 // indirect
	go.opentelemetry.io/otel v1.5.0 // indirect
	go.opentelemetry.io/otel/bridge/opentracing v1.5.0 // indirect
	go.opentelemetry.io/otel/sdk v1.5.0 // indirect
	go.opentelemetry.io/otel/trace v1.5.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/goleak v1.1.12 // indirect
	golang.org/x/crypto v0.0.0-20220829220503-c86fa9a7ed90 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/net v0.0.0-20220909164309-bea034e7d591 // indirect
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5 // indirect
	golang.org/x/sys v0.0.0-20220909162455-aba9fc2a8ff2 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11 // indirect
	golang.org/x/tools v0.1.12 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/api v0.74.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	helm.sh/helm/v3 v3.8.2 // indirect
	howett.net/plist v0.0.0-20181124034731-591f970eefbb // indirect
	k8s.io/apiextensions-apiserver v0.23.5 // indirect
	k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65 // indirect
	k8s.io/utils v0.0.0-20211116205334-6203023598ed // indirect
	moul.io/http2curl v1.0.0 // indirect
	sigs.k8s.io/json v0.0.0-20211020170558-c049b76a60c6 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

require (
	cloud.google.com/go/iam v0.3.0 // indirect
	github.com/OneOfOne/xxhash v1.2.6 // indirect
	github.com/clusternet/clusternet v0.13.0
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-openapi/spec v0.20.6 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/go-playground/validator/v10 v10.11.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/urfave/cli/v2 v2.8.1 // indirect
	go.etcd.io/etcd/client/v3 v3.5.4 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	google.golang.org/genproto v0.0.0-20220607223854-30acc4cbd2aa // indirect
)

replace (
	github.com/go-resty/resty/v2 => github.com/ifooth/resty/v2 v2.0.0-20230223083514-3015979960de
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v1.8.2-0.20220308163432-03831554a519
	github.com/thanos-io/thanos => github.com/ifooth/thanos v0.26.1-0.20230320020638-51ebe892541f
	github.com/vimeo/galaxycache => github.com/thanos-community/galaxycache v0.0.0-20211122094458-3a32041a1f1e
)
