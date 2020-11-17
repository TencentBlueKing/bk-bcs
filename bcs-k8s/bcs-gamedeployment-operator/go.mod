module github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator

go 1.14

replace (
	bitbucket.org/ww/goautoneg => github.com/adjust/goautoneg v0.0.0-20150426214442-d788f35a0315
	github.com/Tencent/bk-bcs => ../../../bk-bcs
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes => ../../bcs-k8s/kubernetes
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.4
	k8s.io/api => k8s.io/api v0.0.0-20181213150558-05914d821849
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181213153335-0fe22c71c476 // indirect
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20181127025237-2b1284ed4c93
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20181213151703-3ccfe8365421
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181213151034-8d9ed539ba31
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20181109181836-c59034cc13d5
	k8s.io/kubernetes => k8s.io/kubernetes v1.13.1
)

require (
	bitbucket.org/ww/goautoneg v0.0.0-20120707110453-75cd24fc2f2c // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/Tencent/bk-bcs/bcs-k8s/kubernetes v0.0.0-00010101000000-000000000000
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.10+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/elazarl/go-bindata-assetfs v1.0.0 // indirect
	github.com/emicklei/go-restful-swagger12 v0.0.0-20170926063155-7524189396c6 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.13.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20200820163806-098863c1fc24
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/soheilhy/cmux v0.1.4 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20200122045848-3419fae592fc // indirect
	github.com/ugorji/go v1.1.7 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.etcd.io/bbolt v1.3.3 // indirect
	go.uber.org/zap v1.13.0 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/grpc v1.27.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	k8s.io/api v0.18.4
	k8s.io/apimachinery v0.18.5
	k8s.io/apiserver v0.18.4
	k8s.io/client-go v0.18.4
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v0.0.0-00010101000000-000000000000
	k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89
)