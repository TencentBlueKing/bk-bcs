module bcs-egress

go 1.17

replace (
	k8s.io/client-go => k8s.io/client-go v0.16.7
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.5.7
)

require (
	github.com/operator-framework/operator-sdk v0.17.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	k8s.io/api v0.17.7
	k8s.io/apimachinery v0.17.7
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.5.7
)
