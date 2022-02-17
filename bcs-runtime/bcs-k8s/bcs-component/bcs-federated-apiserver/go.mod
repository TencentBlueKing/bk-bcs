module github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-federated-apiserver

go 1.13

require (
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535 // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/onsi/gomega v1.13.0 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/cobra v1.2.0
	github.com/spf13/pflag v1.0.5
	go.uber.org/automaxprocs v1.4.0
	golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6 // indirect
	k8s.io/api v0.21.2
	k8s.io/apiextensions-apiserver v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/apiserver v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/klog/v2 v2.8.0
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b // indirect

)

replace (
	github.com/spf13/afero => github.com/spf13/afero v1.5.1
	google.golang.org/grpc => google.golang.org/grpc v1.27.1
)
