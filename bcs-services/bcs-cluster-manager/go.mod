module github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager

go 1.14

replace (
	//github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210525083026-bc8c14258b6e // v0.20.16
	github.com/Tencent/bk-bcs/bcs-common => ../../bcs-common
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/coreos/etcd v3.3.25+incompatible => github.com/evanlixin/etcd v3.3.26-0.20210917065228-e1c46c24ee8f+incompatible
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.27.0
	k8s.io/api => k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.6
	k8s.io/client-go => k8s.io/client-go v0.18.6
)

require (
	github.com/RichardKnop/machinery/v2 v2.0.11
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20220123082150-ac3c90791ab4
	github.com/Tencent/bk-bcs/bcs-services/pkg v0.0.0-20220125124309-240e1e103087
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/envoyproxy/protoc-gen-validate v0.4.1
	github.com/ghodss/yaml v1.0.0
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.2.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/jinzhu/gorm v1.9.16
	github.com/kirito41dd/xslice v0.0.1
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/micro/go-micro/v2 v2.9.1
	github.com/parnurzeal/gorequest v0.2.16
	github.com/prometheus/client_golang v1.11.0
	github.com/satori/go.uuid v1.2.0
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.376
    github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm v1.0.376
    github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke v1.0.374
	go.mongodb.org/mongo-driver v1.5.3
	google.golang.org/genproto v0.0.0-20210402141018-6c239bbf2bb1
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	k8s.io/api v0.23.1
	k8s.io/apimachinery v0.23.1
	k8s.io/client-go v11.0.0+incompatible
)
