module github.com/Tencent/bk-bcs

go 1.14

replace (
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210517123645-82ef0026bf95
	github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2 => github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2 v0.0.0-20210517125505-0f40c4b365cb
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/mholt/caddy => github.com/caddyserver/caddy v0.11.1
	github.com/micro/go-micro/v2 => github.com/OvertimeDog/go-micro/v2 v2.9.3
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210128033108-0471fd5e2976
	github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2 v0.0.0-00010101000000-000000000000
	github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager v0.0.0-20210322074635-395741b9b978
	github.com/andygrunwald/megos v0.0.0-20180424065632-0fccaea93714
	github.com/bitly/go-simplejson v0.5.0
	github.com/container-storage-interface/spec v0.3.0
	github.com/containerd/continuity v0.0.0-20190827140505-75bee3e2ccb6 // indirect
	github.com/containernetworking/cni v0.6.0
	github.com/coredns/coredns v1.3.0
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964
	github.com/dbdd4us/qcloudapi-sdk-go v0.0.0-20190530123522-c8d9381de48c
	github.com/deckarep/golang-set v1.7.1
	github.com/dnstap/golang-dnstap v0.0.0-20170829151710-2cf77a2b5e11 // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20181223114339-d147fe0582f4+incompatible // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/evanphx/json-patch v4.9.0+incompatible // indirect
	github.com/farsightsec/golang-framestream v0.0.0-20181102145529-8a0cb8ba8710 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/fsouza/go-dockerclient v1.6.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/go-openapi/analysis v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/loads v0.19.4 // indirect
	github.com/go-openapi/runtime v0.19.7 // indirect
	github.com/go-openapi/spec v0.19.4 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.4.3
	github.com/google/cadvisor v0.32.0
	github.com/google/uuid v1.2.0
	github.com/gorilla/websocket v1.4.2
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645 // indirect
	github.com/haproxytech/client-native v1.2.6
	github.com/haproxytech/models v1.2.4
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/json-iterator/go v1.1.10
	github.com/kr/pretty v0.2.0 // indirect
	github.com/lithammer/go-jump-consistent-hash v1.0.1
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20200820163806-098863c1fc24 // indirect
	github.com/mattn/go-sqlite3 v1.14.0
	github.com/mesos/mesos-go v0.0.10
	github.com/mholt/caddy v0.11.1
	github.com/micro/go-micro/v2 v2.9.1
	github.com/miekg/dns v1.1.27
	github.com/mitchellh/mapstructure v1.1.2
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/opencontainers/runc v1.0.0-rc6.0.20181203215513-96ec2177ae84 // indirect
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pborman/uuid v1.2.0
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.15.0
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/tencentcloud/tencentcloud-sdk-go v1.0.132
	github.com/tmc/grpc-websocket-proxy v0.0.0-20200427203606-3cfed13b9966 // indirect
	github.com/ugorji/go/codec v1.2.3
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.5.1
	go4.org v0.0.0-20190313082347-94abd6928b1d
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	google.golang.org/grpc v1.33.1
	k8s.io/api v0.18.16
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.16
	k8s.io/client-go v0.18.6
	k8s.io/kubelet v0.18.16
	k8s.io/kubernetes v1.13.0
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800 // indirect
	moul.io/http2curl v1.0.0 // indirect
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/structured-merge-diff v0.0.0-20190525122527-15d366b2352e // indirect
	sigs.k8s.io/testing_frameworks v0.1.1 // indirect
)
