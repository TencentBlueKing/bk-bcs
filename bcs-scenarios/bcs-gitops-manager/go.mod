module github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager

go 1.18

replace (
        github.com/Tencent/bk-bcs/bcs-common/pkg/audit => ../../bcs-common/pkg/audit
        github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v1.2.11
)

require (
        github.com/Tencent/bk-bcs/bcs-common v0.0.0-20231007115947-fb72f2248970
        github.com/Tencent/bk-bcs/bcs-common/pkg/audit v0.0.0-20231027105519-ebbe20c3f975
        github.com/Tencent/bk-bcs/bcs-common/pkg/auth v0.0.0-20230811095616-815d33d32e2d
        github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapiv4 v0.0.0-20230811095616-815d33d32e2d
        github.com/Tencent/bk-bcs/bcs-common/pkg/otel v0.0.0-20230613090449-9c5bf107fe88
        github.com/Tencent/bk-bcs/bcs-services/cluster-resources v0.0.0-20230811095616-815d33d32e2d
        github.com/Tencent/bk-bcs/bcs-services/pkg v0.0.0-20230701022721-8cbd62252af8
        github.com/argoproj/argo-cd/v2 v2.6.2
        github.com/argoproj/gitops-engine v0.7.1-0.20221208230615-917f5a0f16d5
        github.com/asim/go-micro/plugins/sync/etcd/v4 v4.7.0
        github.com/envoyproxy/protoc-gen-validate v0.10.1
        github.com/gin-gonic/gin v1.9.0
        github.com/go-micro/plugins/v4/client/grpc v1.1.0
        github.com/go-micro/plugins/v4/registry/etcd v1.1.0
        github.com/go-micro/plugins/v4/server/grpc v1.2.0
        github.com/gogo/protobuf v1.3.2
        github.com/golang/protobuf v1.5.3
        github.com/google/uuid v1.3.1
        github.com/gorilla/mux v1.8.0
        github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
        github.com/grpc-ecosystem/grpc-gateway v1.16.0
        github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0
        github.com/pkg/errors v0.9.1
        github.com/prometheus-operator/prometheus-operator/pkg/client v0.66.0
        github.com/prometheus/client_golang v1.15.1
        github.com/spf13/cobra v1.6.1
        github.com/spf13/pflag v1.0.5
        go-micro.dev/v4 v4.10.2
        go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.31.0
        go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.43.0
        go.opentelemetry.io/otel v1.17.0
        go.opentelemetry.io/otel/trace v1.17.0
        google.golang.org/genproto/googleapis/api v0.0.0-20230530153820-e85fd2cbaebc
        google.golang.org/grpc v1.57.0
        google.golang.org/protobuf v1.31.0
        gopkg.in/go-playground/webhooks.v5 v5.17.0
        k8s.io/api v0.27.2
        k8s.io/apimachinery v0.27.2
        k8s.io/client-go v0.27.2
        k8s.io/kubernetes v1.24.2
)

require (
        cloud.google.com/go/compute v1.19.1 // indirect
        cloud.google.com/go/compute/metadata v0.2.3 // indirect
        code.gitea.io/sdk/gitea v0.15.1 // indirect
        github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
        github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd // indirect
        github.com/Masterminds/goutils v1.1.1 // indirect
        github.com/Masterminds/semver/v3 v3.2.0 // indirect
        github.com/Masterminds/sprig/v3 v3.2.2 // indirect
        github.com/Microsoft/go-winio v0.6.1 // indirect
        github.com/ProtonMail/go-crypto v0.0.0-20220824120805-4b6e5c587895 // indirect
        github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common v0.0.0-20220330120237-0bbed74dcf6d // indirect
        github.com/TencentBlueKing/bk-audit-go-sdk v0.0.5 // indirect
        github.com/TencentBlueKing/gopkg v1.1.0 // indirect
        github.com/TencentBlueKing/iam-go-sdk v0.1.3 // indirect
        github.com/acomagu/bufpipe v1.0.3 // indirect
        github.com/argoproj/pkg v0.13.7-0.20221221191914-44694015343d // indirect
        github.com/beorn7/perks v1.0.1 // indirect
        github.com/bitly/go-simplejson v0.5.0 // indirect
        github.com/blang/semver/v4 v4.0.0 // indirect
        github.com/bombsimon/logrusr/v2 v2.0.1 // indirect
        github.com/bradleyfalzon/ghinstallation/v2 v2.1.0 // indirect
        github.com/bytedance/sonic v1.8.0 // indirect
        github.com/cenkalti/backoff/v4 v4.2.1 // indirect
        github.com/cespare/xxhash/v2 v2.2.0 // indirect
        github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
        github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
        github.com/cloudflare/circl v1.2.0 // indirect
        github.com/coreos/go-oidc v2.2.1+incompatible // indirect
        github.com/coreos/go-semver v0.3.0 // indirect
        github.com/coreos/go-systemd/v22 v22.3.2 // indirect
        github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
        github.com/davecgh/go-spew v1.1.1 // indirect
        github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
        github.com/docker/distribution v2.8.2+incompatible // indirect
        github.com/emicklei/go-restful/v3 v3.10.2 // indirect
        github.com/emirpasic/gods v1.18.1 // indirect
        github.com/evanphx/json-patch v5.6.0+incompatible // indirect
        github.com/evanphx/json-patch/v5 v5.6.0 // indirect
        github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
        github.com/fatih/camelcase v1.0.0 // indirect
        github.com/felixge/httpsnoop v1.0.3 // indirect
        github.com/fsnotify/fsnotify v1.6.0 // indirect
        github.com/fvbommel/sortorder v1.0.1 // indirect
        github.com/gfleury/go-bitbucket-v1 v0.0.0-20220301131131-8e7ed04b843e // indirect
        github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
        github.com/gin-contrib/sse v0.1.0 // indirect
        github.com/go-acme/lego/v4 v4.4.0 // indirect
        github.com/go-errors/errors v1.0.1 // indirect
        github.com/go-git/gcfg v1.5.0 // indirect
        github.com/go-git/go-billy/v5 v5.3.1 // indirect
        github.com/go-git/go-git/v5 v5.4.2 // indirect
        github.com/go-logr/logr v1.2.4 // indirect
        github.com/go-logr/stdr v1.2.2 // indirect
        github.com/go-openapi/jsonpointer v0.19.6 // indirect
        github.com/go-openapi/jsonreference v0.20.2 // indirect
        github.com/go-openapi/swag v0.22.4 // indirect
        github.com/go-playground/locales v0.14.1 // indirect
        github.com/go-playground/universal-translator v0.18.1 // indirect
        github.com/go-playground/validator/v10 v10.11.2 // indirect
        github.com/go-redis/cache/v8 v8.4.3 // indirect
        github.com/go-redis/redis/v8 v8.11.5 // indirect
        github.com/go-resty/resty/v2 v2.7.0 // indirect
        github.com/go-sql-driver/mysql v1.7.1 // indirect
        github.com/gobwas/glob v0.2.3 // indirect
        github.com/gobwas/httphead v0.1.0 // indirect
        github.com/gobwas/pool v0.2.1 // indirect
        github.com/gobwas/ws v1.0.4 // indirect
        github.com/goccy/go-json v0.10.0 // indirect
        github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
        github.com/golang-migrate/migrate/v4 v4.16.2 // indirect
        github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
        github.com/google/btree v1.0.1 // indirect
        github.com/google/gnostic v0.6.9 // indirect
        github.com/google/go-cmp v0.5.9 // indirect
        github.com/google/go-github/v35 v35.3.0 // indirect
        github.com/google/go-github/v45 v45.2.0 // indirect
        github.com/google/go-querystring v1.1.0 // indirect
        github.com/google/gofuzz v1.2.0 // indirect
        github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
        github.com/gorilla/handlers v1.5.1 // indirect
        github.com/gorilla/websocket v1.4.2 // indirect
        github.com/gosimple/slug v1.13.1 // indirect
        github.com/gosimple/unidecode v1.0.1 // indirect
        github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
        github.com/hashicorp/errwrap v1.1.0 // indirect
        github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
        github.com/hashicorp/go-hclog v1.2.2 // indirect
        github.com/hashicorp/go-multierror v1.1.1 // indirect
        github.com/hashicorp/go-retryablehttp v0.7.1 // indirect
        github.com/hashicorp/go-version v1.5.0 // indirect
        github.com/huandu/xstrings v1.3.2 // indirect
        github.com/imdario/mergo v0.3.13 // indirect
        github.com/inconshreveable/mousetrap v1.0.1 // indirect
        github.com/itchyny/gojq v0.12.9 // indirect
        github.com/itchyny/timefmt-go v0.1.4 // indirect
        github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
        github.com/jeremywohl/flatten v1.0.1 // indirect
        github.com/jonboulle/clockwork v0.2.2 // indirect
        github.com/josharian/intern v1.0.0 // indirect
        github.com/json-iterator/go v1.1.12 // indirect
        github.com/juju/ratelimit v1.0.1 // indirect
        github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
        github.com/kevinburke/ssh_config v1.2.0 // indirect
        github.com/klauspost/compress v1.15.11 // indirect
        github.com/klauspost/cpuid/v2 v2.1.0 // indirect
        github.com/ktrysmt/go-bitbucket v0.9.55 // indirect
        github.com/leodido/go-urn v1.2.2 // indirect
        github.com/lib/pq v1.10.6 // indirect
        github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
        github.com/mailru/easyjson v0.7.7 // indirect
        github.com/mattbaird/jsonpatch v0.0.0-20200820163806-098863c1fc24 // indirect
        github.com/mattn/go-isatty v0.0.17 // indirect
        github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
        github.com/microsoft/azure-devops-go-api/azuredevops v1.0.0-b5 // indirect
        github.com/miekg/dns v1.1.50 // indirect
        github.com/mitchellh/copystructure v1.2.0 // indirect
        github.com/mitchellh/go-homedir v1.1.0 // indirect
        github.com/mitchellh/go-wordwrap v1.0.1 // indirect
        github.com/mitchellh/hashstructure v1.1.0 // indirect
        github.com/mitchellh/mapstructure v1.5.0 // indirect
        github.com/mitchellh/reflectwalk v1.0.2 // indirect
        github.com/moby/spdystream v0.2.0 // indirect
        github.com/moby/term v0.5.0 // indirect
        github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
        github.com/modern-go/reflect2 v1.0.2 // indirect
        github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
        github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
        github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
        github.com/nxadm/tail v1.4.8 // indirect
        github.com/opencontainers/go-digest v1.0.0 // indirect
        github.com/opencontainers/selinux v1.10.0 // indirect
        github.com/openzipkin/zipkin-go v0.3.0 // indirect
        github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
        github.com/parnurzeal/gorequest v0.2.16 // indirect
        github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
        github.com/pelletier/go-toml/v2 v2.0.6 // indirect
        github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
        github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
        github.com/pquerna/cachecontrol v0.1.0 // indirect
        github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.66.0 // indirect
        github.com/prometheus/client_model v0.4.0 // indirect
        github.com/prometheus/common v0.42.0 // indirect
        github.com/prometheus/procfs v0.9.0 // indirect
        github.com/r3labs/diff v1.1.0 // indirect
        github.com/robfig/cron/v3 v3.0.1 // indirect
        github.com/russross/blackfriday v1.5.2 // indirect
        github.com/russross/blackfriday/v2 v2.1.0 // indirect
        github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414 // indirect
        github.com/sergi/go-diff v1.2.0 // indirect
        github.com/shopspring/decimal v1.2.0 // indirect
        github.com/sirupsen/logrus v1.9.3 // indirect
        github.com/spf13/cast v1.5.0 // indirect
        github.com/stretchr/testify v1.8.4 // indirect
        github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
        github.com/ugorji/go/codec v1.2.9 // indirect
        github.com/urfave/cli/v2 v2.3.0 // indirect
        github.com/valyala/bytebufferpool v1.0.0 // indirect
        github.com/valyala/fasttemplate v1.2.2 // indirect
        github.com/vmihailenco/go-tinylfu v0.2.2 // indirect
        github.com/vmihailenco/msgpack/v5 v5.3.4 // indirect
        github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
        github.com/xanzy/go-gitlab v0.60.0 // indirect
        github.com/xanzy/ssh-agent v0.3.2 // indirect
        github.com/xlab/treeprint v0.0.0-20181112141820-a009c3971eca // indirect
        go.etcd.io/etcd/api/v3 v3.5.4 // indirect
        go.etcd.io/etcd/client/pkg/v3 v3.5.4 // indirect
        go.etcd.io/etcd/client/v3 v3.5.4 // indirect
        go.mongodb.org/mongo-driver v1.10.0 // indirect
        go.opentelemetry.io/otel/exporters/jaeger v1.3.0 // indirect
        go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.17.0 // indirect
        go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.17.0 // indirect
        go.opentelemetry.io/otel/exporters/zipkin v1.3.0 // indirect
        go.opentelemetry.io/otel/metric v1.17.0 // indirect
        go.opentelemetry.io/otel/sdk v1.17.0 // indirect
        go.opentelemetry.io/proto/otlp v1.0.0 // indirect
        go.starlark.net v0.0.0-20200306205701-8dd3e2ee1dd5 // indirect
        go.uber.org/atomic v1.9.0 // indirect
        go.uber.org/multierr v1.6.0 // indirect
        go.uber.org/zap v1.24.0 // indirect
        golang.org/x/arch v0.0.0-20210923205945-b76863e36670 // indirect
        golang.org/x/crypto v0.7.0 // indirect
        golang.org/x/exp v0.0.0-20230315142452-642cacee5cc0 // indirect
        golang.org/x/mod v0.10.0 // indirect
        golang.org/x/net v0.10.0 // indirect
        golang.org/x/oauth2 v0.8.0 // indirect
        golang.org/x/sync v0.2.0 // indirect
        golang.org/x/sys v0.11.0 // indirect
        golang.org/x/term v0.9.0 // indirect
        golang.org/x/text v0.10.0 // indirect
        golang.org/x/time v0.3.0 // indirect
        golang.org/x/tools v0.9.1 // indirect
        google.golang.org/appengine v1.6.7 // indirect
        google.golang.org/genproto v0.0.0-20230526203410-71b5a4ffd15e // indirect
        google.golang.org/genproto/googleapis/rpc v0.0.0-20230530153820-e85fd2cbaebc // indirect
        gopkg.in/inf.v0 v0.9.1 // indirect
        gopkg.in/square/go-jose.v2 v2.6.0 // indirect
        gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
        gopkg.in/warnings.v0 v0.1.2 // indirect
        gopkg.in/yaml.v2 v2.4.0 // indirect
        gopkg.in/yaml.v3 v3.0.1 // indirect
        k8s.io/apiextensions-apiserver v0.27.2 // indirect
        k8s.io/apiserver v0.24.2 // indirect
        k8s.io/cli-runtime v0.24.2 // indirect
        k8s.io/cloud-provider v0.24.1 // indirect
        k8s.io/component-base v0.27.2 // indirect
        k8s.io/component-helpers v0.24.2 // indirect
        k8s.io/cri-api v0.0.0 // indirect
        k8s.io/klog v1.0.0 // indirect
        k8s.io/klog/v2 v2.100.1 // indirect
        k8s.io/kube-aggregator v0.24.2 // indirect
        k8s.io/kube-openapi v0.0.0-20230501164219-8b0f38b5fd1f // indirect
        k8s.io/kubectl v0.24.2 // indirect
        k8s.io/mount-utils v0.24.1 // indirect
        k8s.io/utils v0.0.0-20230505201702-9f6742963106 // indirect
        moul.io/http2curl v1.0.0 // indirect
        sigs.k8s.io/controller-runtime v0.15.0 // indirect
        sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
        sigs.k8s.io/kustomize/api v0.11.4 // indirect
        sigs.k8s.io/kustomize/kyaml v0.13.6 // indirect
        sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
        sigs.k8s.io/yaml v1.3.0 // indirect
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