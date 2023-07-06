module github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager

go 1.17

replace (
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210818040851-76fdc539dc33
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210818040851-76fdc539dc33
	github.com/envoyproxy/protoc-gen-validate v0.1.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.4
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/micro/go-micro/v2 v2.9.1
	github.com/parnurzeal/gorequest v0.2.16
	github.com/prometheus/client_golang v1.9.0
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.33.1
	k8s.io/api v0.18.6
)

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/coreos/etcd v3.3.18+incompatible // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/fsnotify/fsnotify v1.4.7 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.0.0 // indirect
	github.com/go-git/go-git/v5 v5.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20190725054713-01f96b0aa0cd // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/micro/cli/v2 v2.1.2 // indirect
	github.com/micro/go-plugins/broker/rabbitmq/v2 v2.9.1 // indirect
	github.com/micro/go-plugins/broker/stan/v2 v2.9.1 // indirect
	github.com/miekg/dns v1.1.27 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/nats-io/jwt v0.3.2 // indirect
	github.com/nats-io/nats.go v1.10.0 // indirect
	github.com/nats-io/nkeys v0.1.4 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nats-io/stan.go v0.8.2 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.15.0 // indirect
	github.com/prometheus/procfs v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71 // indirect
	github.com/ugorji/go/codec v1.2.3 // indirect
	github.com/xanzy/ssh-agent v0.2.1 // indirect
	go.uber.org/atomic v1.5.0 // indirect
	go.uber.org/multierr v1.3.0 // indirect
	go.uber.org/tools v0.0.0-20190618225709-2cfd321de3ee // indirect
	go.uber.org/zap v1.13.0 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
	golang.org/x/mod v0.3.0 // indirect
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b // indirect
	golang.org/x/sys v0.0.0-20201214210602-f9fddec55a1e // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/tools v0.0.0-20210106214847-113979e3529a // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.23.1-0.20200526195155-81db48ad09cc // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	honnef.co/go/tools v0.0.1-2019.2.3 // indirect
	k8s.io/apimachinery v0.18.6 // indirect
	k8s.io/klog v1.0.0 // indirect
	moul.io/http2curl v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
)
