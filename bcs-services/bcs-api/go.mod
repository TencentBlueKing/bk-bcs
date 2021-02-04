module github.com/Tencent/bk-bcs/bcs-services/bcs-api

go 1.14

replace (
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210122122633-ada65fa64361
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	k8s.io/api => k8s.io/api v0.0.0-20181126151915-b503174bad59
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181126155829-0cd23ebeb688
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20181126123746-eddba98df674
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181126152608-d082d5923d3c
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210204084037-834463e85666
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef
	github.com/dchest/uniuri v0.0.0-20200228104902-7aecb25e1fe5
	github.com/deckarep/golang-set v1.7.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/google/go-cmp v0.5.4
	github.com/googleapis/gnostic v0.0.0-00010101000000-000000000000 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/iancoleman/strcase v0.1.3
	github.com/jinzhu/gorm v1.9.16
	github.com/json-iterator/go v1.1.10
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/parnurzeal/gorequest v0.2.16
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/prometheus/client_golang v1.9.0
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414
	github.com/sirupsen/logrus v1.7.0
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v11.0.0+incompatible
)
