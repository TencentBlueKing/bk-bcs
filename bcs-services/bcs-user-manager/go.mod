module github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager

go 1.14

replace (
	github.com/Tencent/bk-bcs/bcs-common => github.com/Tencent/bk-bcs/bcs-common v0.0.0-20210818040851-76fdc539dc33
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.0.0-20181126151915-b503174bad59
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20181126123746-eddba98df674
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181126152608-d082d5923d3c
)

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-00010101000000-000000000000
	github.com/TencentBlueKing/iam-go-sdk v0.0.5
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef
	github.com/dchest/uniuri v0.0.0-20200228104902-7aecb25e1fe5
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/jinzhu/gorm v1.9.16
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/micro/go-micro/v2 v2.9.1
	github.com/parnurzeal/gorequest v0.2.16
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.9.0
	github.com/rs/xid v1.3.0
	github.com/sirupsen/logrus v1.8.1
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	k8s.io/api v0.0.0-00010101000000-000000000000 // indirect
	k8s.io/apimachinery v0.0.0-00010101000000-000000000000
)
