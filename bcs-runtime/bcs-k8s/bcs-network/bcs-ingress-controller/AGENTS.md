# AGENTS.md -- bcs-ingress-controller

## Project Snapshot

BCS Ingress Controller -- Kubernetes Operator managing network extension CRDs
(PortPool, PortBinding, Listener, Ingress, HostNetPortPool).
Go 1.20+, controller-runtime v0.6.3, go-restful HTTP API, Prometheus metrics.
go.mod lives at bcs-network/ parent; CRD types in ../../kubernetes/apis/networkextension/v1/.
Completed features: HostNetPortPool hostNetwork port allocation; Namespace scope exemption for
federated cluster ingresses (see "Namespace Scope Exemption" section below).

## Build and Test

Build binary (run from bcs-ingress-controller/ dir):

    cd .. && make ingress-controller

Build + Docker image push (run from bcs-network/ dir):

    make ingress-controller
    VERSION=$(git describe --always)-$(date +%y.%m.%d)
    BUILD_DIR=../../../build/bcs.${VERSION}/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller
    docker build -t mirrors.tencent.com/<your_repository>/bcs-ingress-controller:latest -f ${BUILD_DIR}/Dockerfile ${BUILD_DIR}
    docker push mirrors.tencent.com/<your_repository>/bcs-ingress-controller:latest

Unit tests with coverage report:

    cd .. && make test-ingress-controller

Run single package tests:

    cd .. && go test -v -run TestReconcile ./bcs-ingress-controller/hostnetportcontroller/...

## K8s Cluster Deploy

    export KUBECONFIG=<path_to_your_kubeconfig>
    kubectl rollout restart -n bcs-system deployment/bcsingresscontroller

## Code Conventions

- Code comments: **English**. Docs / PR / chat: **Chinese**
- Format: gofmt / goimports, explicit error handling, no naked returns
- Constants in internal/constant/constant.go -- never use string literals for annotation keys
- Controllers: controller-runtime Reconcile, must be **idempotent**
- Tests: **table-driven**, colocated *_test.go, fake client from controller-runtime/pkg/client/fake
- Metrics: init() registration in internal/metrics/*.go, namespace bkbcs_ingressctrl
- Logging: bcs-common/common/blog -- not stdlib log or klog
- HTTP: go-restful WebService, routes registered in internal/httpsvr/httpserver.go InitRouters()
- **Naming**: Function names MUST NOT exceed 35 characters. Keep function names concise and descriptive.

## Patterns -- DO

**New controller** -- follow portpoolcontroller/portpool_controller.go:
- Struct: ctx, client.Client, domain cache, record.EventRecorder
- Constructor NewXxxReconciler(ctx, cli, cache, eventer)
- SetupWithManager(mgr) with For(primaryCRD) + Watches(source.Kind)
- Reconcile(): fetch resource -> handle IsNotFound gracefully -> business logic -> update status
- Register in main.go via SetupWithManager(mgr)

**New metrics** -- follow internal/metrics/portpool.go:
- Package-level var block with prometheus.NewGaugeVec / CounterVec / HistogramVec
- init() calls metrics.Registry.MustRegister(...)
- Exported helpers: ReportXxx(...), CleanXxx(...), IncreaseXxx(...)

**New HTTP endpoint** -- follow internal/httpsvr/portpool.go:
- Handler method (h *HttpServerClient) handlerName(req, resp)
- Register in InitRouters() at internal/httpsvr/httpserver.go

**New cache** -- follow internal/portpoolcache/:
- Thread-safe with sync.RWMutex, types in separate types.go
- RebuildFromAPIServer() for cold-start / leader election recovery

**New constants** -- append to internal/constant/constant.go with exported const + godoc comment

## Namespace Scope Exemption

Flag `--is_namespace_scope` (bool) restricts each Ingress to services in the same namespace.
Flag `--namespace_scope_exempt_namespaces` (string, comma-separated) lists namespaces that bypass
both restrictions: they may reference cross-namespace services AND use the controller's global
cloud credentials (env `TENCENTCLOUD_ACCESS_KEY_ID` / `TENCENTCLOUD_ACESS_KEY`) instead of
requiring a per-namespace Secret / ControllerConfig.

**Data flow for exempt namespaces:**

    ControllerOption.NamespaceScopeExemptNamespaces (string)
      └─ parseExemptNamespaces() → map[string]struct{}
           ├─ IngressConverterOpt.ExemptNamespaces              (generator/ingressconverter.go)
           │    └─ IngressListenerConverter.ExemptNamespaces
           │         ├─ RuleConverter.exemptNamespaces          (generator/ruleconverter.go)
           │         │    └─ isIngressNamespaceExempt() skips cross-ns override
           │         └─ MappingConverter.exemptNamespaces       (generator/mappingconverter.go)
           │              └─ isIngressNamespaceExempt() skips cross-ns override
           └─ newNamespacedLBWithExempt() → NamespacedLB
                └─ NamespacedLB.defaultClient + exemptNamespaces (namespacedlb/namespacedclient.go)
                     └─ getNsClient(): exempt ns → defaultClient, others → per-ns secret lookup

**Adding a new cloud provider to the exemption pattern:**

1. Add `initXxxClient(ctx, opts, cli, ew, exemptNsMap)` in main.go following `initTencentCloudClient`.
2. Call `newNamespacedLBWithExempt(...)` in the `isNamespaceScope` branch.
3. Add the new case to `initClient`'s switch.

**Secret key names for per-ns Secret (note historical typo in key name):**

    kubectl create secret generic ingress-secret.networkextension.bkbcs.tencent.com \
      -n <namespace> \
      --from-literal=TENCENTCLOUD_ACCESS_KEY_ID="<secretID>" \
      --from-literal=TENCENTCLOUD_ACESS_KEY="<secretKey>"   # ACESS (not ACCESS) -- matches code constant

**Unit tests:**
- Generator layer: `internal/generator/namespace_scope_exempt_test.go` (14 cases)
- NamespacedLB layer: `internal/cloud/namespacedlb/namespacedclient_test.go` (7 test funcs)
- Flag parsing: `main_test.go::TestParseExemptNamespaces` (7 cases)

## initClient Decomposition

`initClient` (main.go) is a pure dispatcher -- keep function cyclomatic complexity low (≤ 5 for simple dispatch functions):

    initClient()
      parseExemptNamespaces()           # string → map[string]struct{}
      switch cloud →
        initTencentCloudClient()
        initAWSClient()
        initGCPClient()
        initAzureClient()

Each `initXxxClient` calls `newNamespacedLBWithExempt` when `IsNamespaceScope=true`.
Do NOT add business logic back into `initClient` -- put it in the per-cloud function.

**Note:** All functions should maintain reasonable cyclomatic complexity. Complicated functions (complexity > 10) should be decomposed into smaller, focused functions.

## Patterns -- DON'T

- Import log / klog directly -- use blog
- Hardcode annotation keys -- add constant in internal/constant/constant.go
- Skip error checking on k8s API calls (client.Get, client.Update, etc.)
- Create controllers without registering in main.go
- Assume go.mod is in this directory -- it is at ../go.mod, run go mod tidy from ../
- Add cloud init logic directly into `initClient` -- add a new `initXxxClient` function instead
- Use `TENCENTCLOUD_ACCESS_KEY` (correct spelling) in Secrets -- the code constant is `TENCENTCLOUD_ACESS_KEY` (typo, one C)

## Key Files

| Purpose | Path |
|---------|------|
| Entry point, wires all controllers / checkers / HTTP | main.go |
| All constants and annotation keys | internal/constant/constant.go |
| CLI flags and controller config | internal/option/option.go |
| HTTP API route registration | internal/httpsvr/httpserver.go |
| Webhook admission handlers | internal/webhookserver/ |
| CRD Go type definitions | ../../kubernetes/apis/networkextension/v1/ |
| CRD YAML manifests | ../../kubernetes/config/crd/bases/ |
| Generated clientset / listers / informers | ../../kubernetes/generated/ |
| Build Makefile | ../Makefile |
| Ingress → Listener conversion entry | internal/generator/ingressconverter.go |
| Rule-based listener conversion (L7) | internal/generator/ruleconverter.go |
| Port-mapping listener conversion (L4) | internal/generator/mappingconverter.go |
| Namespaced cloud client (per-ns credential) | internal/cloud/namespacedlb/namespacedclient.go |

## Controllers and CRDs

| Controller | CRD / Resource | Directory |
|------------|---------------|-----------|
| Ingress | networkextension.Ingress | ingresscontroller/ |
| Listener | networkextension.Listener | listenercontroller/ |
| PortPool | networkextension.PortPool | portpoolcontroller/ |
| PortBinding | networkextension.PortBinding | portbindingcontroller/ |
| HostNetPortPool | networkextension.HostNetPortPool | hostnetportcontroller/ |
| Namespace | core/v1.Namespace | namespacecontroller/ |
| Node | core/v1.Node | nodecontroller/ |

## Module Layout

    bcs-ingress-controller/
      main.go                     # Wires everything
      {name}controller/           # One dir per controller
        controller.go             # Reconciler + SetupWithManager
        *_test.go                 # Colocated tests
      internal/
        constant/                 # Shared constants
        metrics/                  # Prometheus metrics (one file per subsystem)
        httpsvr/                  # REST API handlers + route registration
        check/                    # Periodic consistency checkers
        cloud/                    # Cloud adapters (aws/ azure/ gcp/ tencentcloud/)
        webhookserver/            # Admission webhooks
        hostnetportpoolcache/     # HostNetPortPool in-memory cache
        portpoolcache/            # PortPool in-memory cache
        generator/                # Ingress to Listener conversion
        ingresscache/             # Service/workload cache for Ingress
        nodecache/                # Node metadata cache
        apiclient/                # External API client helpers
        option/                   # CLI option parsing
        worker/                   # Worker/sync utilities
      bcs-ingress-inspector/      # Separate diagnostic binary (NOT main controller)
      specs/                      # Feature design documents
      benchmark/                  # Performance test scripts and fixtures

## JIT Search Commands

Find all Reconcile implementations:

    rg -n "func.*Reconcile\(" --type go
    # Or more specifically:
    rg -n "\) Reconcile\(" --type go

Find annotation / constant definitions:

    rg -n "const\b" internal/constant/constant.go

Find Prometheus metric definitions:

    rg -n "prometheus\.New" internal/metrics/

Find HTTP routes:

    rg -n "ws\.Route" internal/httpsvr/

Find CRD type structs:

    rg -n "type .*(Spec|Status) struct" ../../kubernetes/apis/networkextension/v1/

Find where controllers are registered:

    rg -n "SetupWithManager" main.go

Find webhook handlers:

    rg -n "Handle\(" internal/webhookserver/

Find all test files:

    rg -l "_test\.go" --type go

Find namespace-scope exemption touch-points:

    rg -n "ExemptNamespace\|exemptNamespace\|isExempt\|isIngressNamespaceExempt" --type go

Find per-cloud init functions:

    rg -n "^func init.*Client\b" main.go

## Common Gotchas

- go.mod is at **bcs-network/** level -- always run go mod tidy from ../
- CRD types live in a **separate Go module** (../../kubernetes/), linked via replace directive
- controller-runtime v0.6.3 + effective k8s client v0.18.6 -- old API style, no generics
- bcs-ingress-inspector/ is a **separate binary** with its own main.go, not part of this controller
- The cloud/ adapters have vendor-specific SDKs -- only import the one you need

## Pre-PR Checklist (updated)

Run tests before submitting:

    cd .. && make test-ingress-controller && echo "All tests OK"

Quick exemption-feature regression (fast, no cluster needed):

    go test -count=1 -run 'TestParseExemptNamespaces|TestRuleConverter|TestMappingConverter|TestIsExempt|TestGetNsClient|TestReloadNsClient|TestNewNamespacedLB' \
      ./internal/cloud/namespacedlb/... ./internal/generator/... . 2>&1 | grep -E 'PASS|FAIL|ok'

Verify:
- New constants added to internal/constant/constant.go
- New controllers registered in main.go via SetupWithManager
- New metrics registered via init() in internal/metrics/
- New HTTP routes added in internal/httpsvr/httpserver.go InitRouters()
- CRD type changes: regenerate deepcopy and manifests in ../../kubernetes/
- gofmt / goimports clean
- All functions maintain reasonable cyclomatic complexity (decompose functions with complexity > 10)

## Completed Features

- ~~hostnet-port-allocation~~: HostNetPortPool dynamic port allocation for hostNetwork Pods (done)
- Design docs: specs/001-hostnet-port-allocation/
- ~~namespace-scope-exemption~~: Federated cluster ingresses in whitelisted namespaces bypass
  cross-namespace service binding restriction and per-namespace cloud credential requirement.
  Key flag: `--namespace_scope_exempt_namespaces`; key files: internal/generator/{rule,mapping,listener,ingress}converter.go,
  internal/cloud/namespacedlb/namespacedclient.go, main.go (parseExemptNamespaces/newNamespacedLBWithExempt/initXxxClient).
