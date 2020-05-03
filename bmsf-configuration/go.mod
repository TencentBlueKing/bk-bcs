module bk-bscp

go 1.12

require (
	github.com/Tencent/bk-bcs v0.0.0-20200316023358-95388ded8ae7
	github.com/apache/thrift v0.12.0
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bluele/gcache v0.0.0-20190518031135-bc40bd653833
	github.com/coreos/bbolt v1.3.4 // indirect
	github.com/coreos/etcd v3.3.10+incompatible
	github.com/go-ini/ini v0.0.0-20190707052557-8659100d2d9e
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gofrs/flock v0.0.0-20190224121256-392e7fae8f1b
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.3.5
	github.com/google/uuid v0.0.0-20190227210549-0cd6bf5da1e1
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.9.0
	github.com/jinzhu/gorm v0.0.0-20190630075019-836fb2c19d84
	github.com/mattn/go-runewidth v0.0.0-20181210065943-3ee7d812e62a // indirect
	github.com/nats-io/nats-server/v2 v2.1.6 // indirect
	github.com/nats-io/nats.go v1.9.2
	github.com/olekukonko/tablewriter v0.0.0-20181026071410-e6d60cf7ba1f
	github.com/pmezard/go-difflib v1.0.0
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/prometheus/common v0.6.0 // indirect
	github.com/prometheus/procfs v0.0.3 // indirect
	github.com/sergi/go-diff v0.0.0-20191119141955-58c5cb1602ee
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/spf13/cobra v0.0.0-20190607144823-f2b07da1e2c3
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/tidwall/gjson v0.0.0-20190715145443-c5e72cdf74df
	github.com/tmc/grpc-websocket-proxy v0.0.0-20200122045848-3419fae592fc // indirect
	github.com/ugorji/go v1.1.7 // indirect
	go.etcd.io/bbolt v1.3.4 // indirect
	go.uber.org/zap v1.15.0 // indirect
	golang.org/x/net v0.0.0-20190628185345-da137c7871d7
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc v1.21.0
	gopkg.in/ini.v1 v1.55.0 // indirect
	gopkg.in/redis.v5 v5.0.0-20170304113825-a16aeec10ff4
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
	gopkg.in/yaml.v2 v2.2.4
)

replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4

replace go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
