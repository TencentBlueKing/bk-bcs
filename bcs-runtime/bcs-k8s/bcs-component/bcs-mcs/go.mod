module github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs

go 1.14

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20211220083546-9911225681e0
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	google.golang.org/grpc v1.43.0 // indirect
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/apiserver v0.20.1
	k8s.io/client-go v0.20.2
	k8s.io/code-generator v0.20.1
	k8s.io/component-base v0.20.2
	k8s.io/klog/v2 v2.4.0
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/mcs-api v0.1.0
)

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.4
	github.com/googleapis/gnostic v0.5.1 => github.com/googleapis/gnostic v0.4.1
	google.golang.org/grpc => google.golang.org/grpc v1.27.1
)
