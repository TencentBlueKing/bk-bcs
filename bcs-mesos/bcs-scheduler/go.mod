module github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler

go 1.14

replace (
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210625040556-0385f88cbfd6
	github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2 => ../kubebkbcsv2
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2 v0.0.0-00010101000000-000000000000
	github.com/andygrunwald/megos v0.0.0-20180424065632-0fccaea93714
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/fsnotify/fsnotify v1.4.9
	github.com/golang/protobuf v1.5.2
	github.com/parnurzeal/gorequest v0.2.16
	github.com/prometheus/client_golang v1.10.0
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20210428140749-89ef3d95e781
	k8s.io/api v0.18.16
	k8s.io/apiextensions-apiserver v0.18.16
	k8s.io/apimachinery v0.18.16
	k8s.io/client-go v0.18.16
)
