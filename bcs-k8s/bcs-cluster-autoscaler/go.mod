module github.com/bk-bcs/bcs-k8s/bcs-cluster-autoscaler

go 1.14

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/envoyproxy/protoc-gen-validate v0.1.0
	github.com/go-ini/ini v1.62.0
	github.com/golang/protobuf v1.4.3
	github.com/google/btree v1.0.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.9.5
	github.com/micro/go-micro/v2 v2.9.1
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/tencentcloud/tencentcloud-sdk-go v1.0.62
	golang.org/x/net v0.0.0-20200625001655-4c5254603344
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20191230161307-f3c370f40bfb
	google.golang.org/grpc v1.26.0
	google.golang.org/protobuf v1.26.0-rc.1
	gopkg.in/ini.v1 v1.62.0 // indirect
	k8s.io/api v0.22.0
	k8s.io/apimachinery v0.16.15
	k8s.io/apiserver v0.16.15
	k8s.io/autoscaler/cluster-autoscaler v0.0.0-20200330130154-66383d0c3b27
	k8s.io/client-go v0.16.15
	k8s.io/component-base v0.16.15
	k8s.io/klog v1.0.0
	k8s.io/kube-scheduler v0.16.15 // indirect
	k8s.io/kubernetes v1.16.15
	k8s.io/utils v0.0.0-20200729134348-d5654de09c73 // indirect
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	k8s.io/api => k8s.io/api v0.16.15
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.15
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.15
	k8s.io/apiserver => k8s.io/apiserver v0.16.15
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.15
	k8s.io/client-go => k8s.io/client-go v0.16.15
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.16.15
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.16.15
	k8s.io/code-generator => k8s.io/code-generator v0.16.15
	k8s.io/component-base => k8s.io/component-base v0.16.15
	k8s.io/cri-api => k8s.io/cri-api v0.16.15
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.16.15
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.15
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.16.15
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.16.15
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.16.15
	k8s.io/kubectl => k8s.io/kubectl v0.16.15
	k8s.io/kubelet => k8s.io/kubelet v0.16.15
	k8s.io/kubernetes => k8s.io/kubernetes v1.16.15
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.16.15
	k8s.io/metrics => k8s.io/metrics v0.16.15
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.16.15
)
