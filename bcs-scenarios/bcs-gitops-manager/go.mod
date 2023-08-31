module github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager

go 1.16

require (
	github.com/Tencent/bk-bcs/bcs-common v0.0.0-20230811095616-815d33d32e2d
	github.com/Tencent/bk-bcs/bcs-common/pkg/auth v0.0.0-20230811095616-815d33d32e2d
	github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapiv4 v0.0.0-20230811095616-815d33d32e2d
	github.com/Tencent/bk-bcs/bcs-common/pkg/otel v0.0.0-20230613090449-9c5bf107fe88
	github.com/Tencent/bk-bcs/bcs-services/cluster-resources v0.0.0-20230811095616-815d33d32e2d
	github.com/Tencent/bk-bcs/bcs-services/pkg v0.0.0-20230701022721-8cbd62252af8
	github.com/agiledragon/gomonkey/v2 v2.10.1
	github.com/argoproj-labs/argocd-vault-plugin v1.15.0
	github.com/argoproj/argo-cd/v2 v2.6.2
	github.com/asim/go-micro/plugins/sync/etcd/v4 v4.7.0
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/envoyproxy/protoc-gen-validate v0.9.1
	github.com/go-micro/plugins/v4/client/grpc v1.1.0
	github.com/go-micro/plugins/v4/registry/etcd v1.1.0
	github.com/go-micro/plugins/v4/server/grpc v1.2.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.3
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.3
	github.com/hashicorp/vault/api v1.9.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.66.0
	github.com/prometheus/client_golang v1.15.1
	github.com/spf13/cobra v1.6.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.12.0
	github.com/stretchr/testify v1.8.2
	go-micro.dev/v4 v4.10.2
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.31.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.28.0
	go.opentelemetry.io/otel v1.14.0
	go.opentelemetry.io/otel/trace v1.14.0
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4
	google.golang.org/grpc v1.53.0
	google.golang.org/protobuf v1.30.0
	gopkg.in/go-playground/webhooks.v5 v5.17.0
	k8s.io/api v0.27.2
	k8s.io/apimachinery v0.27.2
	k8s.io/client-go v0.27.2
	k8s.io/kubernetes v1.24.2
)

replace (
	github.com/argoproj/gitops-engine => github.com/argoproj/gitops-engine v0.7.1-0.20221004132320-98ccd3d43fd9
	// https://github.com/golang/go/issues/33546#issuecomment-519656923
	github.com/go-check/check => github.com/go-check/check v0.0.0-20180628173108-788fd7840127
	github.com/micro/go-micro => go-micro.dev/v4 v4.7.0
	// https://github.com/kubernetes/kubernetes/issues/79384#issuecomment-505627280
	k8s.io/api => k8s.io/api v0.24.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.24.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.24.1
	k8s.io/apiserver => k8s.io/apiserver v0.24.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.24.1
	k8s.io/client-go => k8s.io/client-go v0.24.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.24.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.24.1
	k8s.io/code-generator => k8s.io/code-generator v0.24.1
	k8s.io/component-base => k8s.io/component-base v0.24.1
	k8s.io/component-helpers => k8s.io/component-helpers v0.24.1
	k8s.io/controller-manager => k8s.io/controller-manager v0.24.1
	k8s.io/cri-api => k8s.io/cri-api v0.24.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.24.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.24.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.24.1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.24.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.24.1
	k8s.io/kubectl => k8s.io/kubectl v0.24.1
	k8s.io/kubelet => k8s.io/kubelet v0.24.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.24.1
	k8s.io/metrics => k8s.io/metrics v0.24.1
	k8s.io/mount-utils => k8s.io/mount-utils v0.24.1
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.24.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.24.1
)
