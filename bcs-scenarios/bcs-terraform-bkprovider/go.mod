module github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider

go 1.21

toolchain go1.21.8

replace go-micro.dev/v4/cmd => github.com/micro/go-micro v1.18.0

replace github.com/micro/go-micro => go-micro.dev/v4/cmd v1.18.0

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/grpc-ecosystem/grpc-gateway => github.com/grpc-ecosystem/grpc-gateway v1.16.0
	go-micro.dev/v4 => go-micro.dev/v4 v4.6.0
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20240110084857-febcc2552574
	github.com/Tencent/bk-bcs/bcs-common/pkg/auth v0.0.0-20231130061921-aba183933987
	github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4 v0.0.0-20240129022749-e811fe496ff9
	github.com/asim/go-micro/plugins/client/grpc/v4 v4.7.0
	github.com/asim/go-micro/plugins/registry/etcd/v4 v4.7.0
	github.com/asim/go-micro/plugins/server/grpc/v4 v4.7.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/levigross/grequests v0.0.0-20231203190023-9c307ef1f48d
	github.com/spf13/cobra v1.8.0
	github.com/spf13/pflag v1.0.5
	go-micro.dev/v4 v4.8.1
	google.golang.org/genproto/googleapis/api v0.0.0-20240401170217-c3f982113cda
	google.golang.org/grpc v1.62.1
)

require (
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	go.etcd.io/etcd/api/v3 v3.5.2 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.2 // indirect
	go.etcd.io/etcd/client/v3 v3.5.2 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	google.golang.org/genproto v0.0.0-20240401170217-c3f982113cda // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240325203815-454cdb8f5daa // indirect
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/ProtonMail/go-crypto v1.0.0 // indirect
	github.com/bitly/go-simplejson v0.5.1 // indirect
	github.com/cloudflare/circl v1.3.7 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.5.0 // indirect
	github.com/go-git/go-git/v5 v5.12.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4
	github.com/google/uuid v1.6.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/miekg/dns v1.1.58 // indirect
	github.com/nxadm/tail v1.4.11 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/skeema/knownhosts v1.2.2 // indirect
	github.com/urfave/cli/v2 v2.27.1 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xrash/smetrics v0.0.0-20240312152122-5f08fbb34913 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/mod v0.16.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.19.0 // indirect
	google.golang.org/protobuf v1.33.0
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)
