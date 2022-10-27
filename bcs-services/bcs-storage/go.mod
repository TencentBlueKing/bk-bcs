module github.com/Tencent/bk-bcs/bcs-services/bcs-storage

go 1.17

replace (
	github.com/Tencent/bk-bcs/bcs-common => ../../bcs-common
	github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2 => github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2 v0.0.0-20210517125505-0f40c4b365cb
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/haproxytech/client-native => github.com/abstractmj/client-native v1.2.8
	github.com/mholt/caddy => github.com/caddyserver/caddy v0.11.1
	github.com/micro/go-micro/v2 => github.com/OvertimeDog/go-micro/v2 v2.9.3
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/deckarep/golang-set v1.8.0
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/google/uuid v1.3.0
	github.com/micro/go-micro/v2 v2.9.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/prometheus/client_golang v1.12.1
	go.mongodb.org/mongo-driver v1.9.0
	golang.org/x/net v0.0.0-20220909164309-bea034e7d591
)

require github.com/Tencent/bk-bcs/bcs-common v0.0.0-20220919094211-a1b246e54e5a

require (
	github.com/BurntSushi/toml v1.1.0 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20220824120805-4b6e5c587895 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cloudflare/circl v1.2.0 // indirect
	github.com/coreos/etcd v3.3.18+incompatible // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/emicklei/go-restful-openapi v1.4.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/go-git/go-git/v5 v5.4.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/spec v0.19.3 // indirect
	github.com/go-openapi/swag v0.19.5 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/go-hclog v0.9.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/micro/cli/v2 v2.1.2 // indirect
	github.com/micro/go-plugins/broker/rabbitmq/v2 v2.9.1 // indirect
	github.com/micro/go-plugins/broker/stan/v2 v2.9.1 // indirect
	github.com/miekg/dns v1.1.50 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nats-io/jwt v0.3.2 // indirect
	github.com/nats-io/nats.go v1.10.0 // indirect
	github.com/nats-io/nkeys v0.1.4 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nats-io/stan.go v0.8.2 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/uber/jaeger-client-go v2.29.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/ugorji/go/codec v1.2.3 // indirect
	github.com/xanzy/ssh-agent v0.3.2 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.0.2 // indirect
	github.com/xdg-go/stringprep v1.0.2 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.etcd.io/bbolt v1.3.5 // indirect
	go.uber.org/atomic v1.5.0 // indirect
	go.uber.org/multierr v1.3.0 // indirect
	go.uber.org/tools v0.0.0-20190618225709-2cfd321de3ee // indirect
	go.uber.org/zap v1.13.0 // indirect
	golang.org/x/crypto v0.0.0-20220829220503-c86fa9a7ed90 // indirect
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/sync v0.0.0-20220907140024-f12130a52804 // indirect
	golang.org/x/sys v0.0.0-20220909162455-aba9fc2a8ff2 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.12 // indirect
	google.golang.org/genproto v0.0.0-20210402141018-6c239bbf2bb1 // indirect
	google.golang.org/grpc v1.41.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
)
