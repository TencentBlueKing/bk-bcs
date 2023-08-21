module bscp.io

go 1.20

replace (
	github.com/TencentBlueKing/iam-go-sdk => github.com/TencentBlueKing/iam-go-sdk v0.0.8
	github.com/go-resty/resty/v2 => github.com/ifooth/resty/v2 v2.0.0-20230223083514-3015979960de
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20230424025302-1a6fd25f4b45
	github.com/TencentBlueKing/iam-go-sdk v0.0.9
	github.com/bluele/gcache v0.0.2
	github.com/dustin/go-humanize v1.0.1
	github.com/go-chi/chi/v5 v5.0.8
	github.com/go-chi/render v1.0.2
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-resty/resty/v2 v2.7.0
	github.com/go-sql-driver/mysql v1.7.1
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/flock v0.8.1
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/golang/protobuf v1.5.3
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.15.2
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jmoiron/sqlx v1.3.5
	github.com/json-iterator/go v1.1.12
	github.com/oklog/run v1.1.0
	github.com/pborman/uuid v1.2.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.15.0
	github.com/sirupsen/logrus v1.9.0
	github.com/smartystreets/goconvey v1.8.0
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.15.0
	github.com/stretchr/testify v1.8.2
	github.com/swaggo/http-swagger v1.3.4
	github.com/swaggo/swag v1.16.1
	github.com/tencentyun/cos-go-sdk-v5 v0.7.41
	github.com/tidwall/gjson v1.14.4
	github.com/tidwall/sjson v1.2.5
	go.etcd.io/etcd/api/v3 v3.5.8
	go.etcd.io/etcd/client/v3 v3.5.8
	go.uber.org/atomic v1.10.0
	go.uber.org/automaxprocs v1.5.2
	golang.org/x/sync v0.1.0
	golang.org/x/time v0.3.0
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1
	google.golang.org/grpc v1.54.0
	google.golang.org/protobuf v1.30.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/datatypes v1.1.1-0.20230130040222-c43177d3cf8c
	gorm.io/driver/mysql v1.5.1
	gorm.io/gen v0.3.22
	gorm.io/gorm v1.25.1
	gorm.io/plugin/dbresolver v1.4.1
	gorm.io/plugin/opentelemetry v0.1.2
	gorm.io/plugin/prometheus v0.0.0-20230504115745-1aec2356381b
	gorm.io/sharding v0.5.3
	k8s.io/klog/v2 v2.100.1
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/bwmarrin/snowflake v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/clbanning/mxj v1.8.4 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.6 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/longbridgeapp/sqlparser v0.3.1 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mozillazg/go-httpheader v0.2.1 // indirect
	github.com/parnurzeal/gorequest v0.2.16 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/smartystreets/assertions v1.13.1 // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/swaggo/files v0.0.0-20220610200504-28940afbdbfe // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/ugorji/go/codec v1.2.3 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.8 // indirect
	go.opentelemetry.io/otel v1.15.1 // indirect
	go.opentelemetry.io/otel/metric v0.38.1 // indirect
	go.opentelemetry.io/otel/trace v1.15.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/crypto v0.6.0 // indirect
	golang.org/x/mod v0.9.0 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	golang.org/x/tools v0.7.0 // indirect
	google.golang.org/grpc/examples v0.0.0-20230426024715-eff0942e95d9 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/hints v1.1.1 // indirect
	moul.io/http2curl v1.0.0 // indirect
)
