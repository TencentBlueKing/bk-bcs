module github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler

go 1.17

require (
	bou.ke/monkey v1.0.2
	github.com/Microsoft/hcsshim v0.8.7-0.20191101173118-65519b62243c // indirect
	github.com/containerd/cgroups v1.0.2 // indirect
	github.com/coreos/etcd v3.3.18+incompatible // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/go-ini/ini v1.62.0
	github.com/golang/mock v1.4.1
	github.com/gophercloud/gophercloud v0.3.0 // indirect
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.11.1 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/miekg/dns v1.1.27 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/robfig/cron v1.1.0
	github.com/soheilhy/cmux v0.1.4 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/tmc/grpc-websocket-proxy v0.0.0-20200122045848-3419fae592fc // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.etcd.io/bbolt v1.3.4 // indirect
	go.uber.org/zap v1.13.0 // indirect
	google.golang.org/protobuf v1.26.0-rc.1 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/square/go-jose.v2 v2.3.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	k8s.io/api v0.22.0
	k8s.io/apimachinery v0.16.15
	k8s.io/apiserver v0.16.15
	k8s.io/autoscaler/cluster-autoscaler v0.0.0-20200330130154-66383d0c3b27
	k8s.io/client-go v0.16.15
	k8s.io/component-base v0.16.15
	k8s.io/klog v1.0.0
	k8s.io/kube-scheduler v0.16.15 // indirect
	k8s.io/kubernetes v1.16.15
)

replace (
	github.com/google/cadvisor => github.com/google/cadvisor v0.39.2
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	google.golang.org/api => google.golang.org/api v0.14.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.16.15
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.15
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.15
	k8s.io/apiserver => k8s.io/apiserver v0.16.15
	k8s.io/autoscaler/cluster-autoscaler => github.com/OvertimeDog/cluster-autoscaler v0.0.0-20220126030239-e40e4a967f24
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
