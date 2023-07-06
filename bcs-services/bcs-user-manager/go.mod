module github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager

go 1.17

replace (
	github.com/Tencent/bk-bcs/bcs-services/pkg => ../pkg
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	golang.org/x/net => golang.org/x/net v0.0.0-20210119194325-5f4716e94777
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.23.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.23.1
)

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20230202032803-47379be3c4c1
	github.com/Tencent/bk-bcs/bcs-services/pkg v0.0.0-20230103023020-f3d4c576e20b
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/coreos/etcd v3.3.18+incompatible
	github.com/dchest/uniuri v0.0.0-20200228104902-7aecb25e1fe5
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/go-redis/redis/v8 v8.11.4
	github.com/jinzhu/gorm v1.9.16
	github.com/micro/go-micro/v2 v2.9.1
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.11.0
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.7.0
	gopkg.in/go-playground/validator.v9 v9.31.0
	k8s.io/apimachinery v0.23.1
)

require (
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common v0.0.0-20220330120237-0bbed74dcf6d // indirect
	github.com/TencentBlueKing/bk-audit-go-sdk v0.0.5 // indirect
	github.com/TencentBlueKing/iam-go-sdk v0.0.8 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/coreos/bbolt v1.3.4 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/emicklei/go-restful-openapi v1.4.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.1.0 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-logr/logr v1.2.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/spec v0.19.3 // indirect
	github.com/go-openapi/swag v0.19.5 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.11.2 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/juju/ratelimit v1.0.1 // indirect
	github.com/lib/pq v1.10.2 // indirect
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20200820163806-098863c1fc24 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/miekg/dns v1.1.50 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/ugorji/go/codec v1.2.3 // indirect
	go.mongodb.org/mongo-driver v1.5.3 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/crypto v0.5.0 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f // indirect
	golang.org/x/sync v0.0.0-20220907140024-f12130a52804 // indirect
	golang.org/x/sys v0.4.0 // indirect
	golang.org/x/term v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	golang.org/x/tools v0.1.12 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	google.golang.org/grpc v1.41.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.23.1 // indirect
	k8s.io/apiextensions-apiserver v0.20.0 // indirect
	k8s.io/client-go v0.23.1 // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/klog/v2 v2.30.0 // indirect
	k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65 // indirect
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b // indirect
	moul.io/http2curl v1.0.0 // indirect
	sigs.k8s.io/json v0.0.0-20211020170558-c049b76a60c6 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

require (
	github.com/google/uuid v1.3.0
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/parnurzeal/gorequest v0.2.16
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
)
