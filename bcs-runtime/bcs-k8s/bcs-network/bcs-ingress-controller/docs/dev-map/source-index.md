# 源文件索引（Source Index）

> 按目录分组，列出每个源文件的路径和职责描述。
> 与 module-index 配合使用：先从 module-index 确定模块内关联文件范围，
> 再从此处查看各文件的职责，确定需要修改的具体文件。

<!-- dev-map:auto -->
## hostnetportcontroller/

| 文件 | 职责描述 |
|------|---------|
| [`hostnetportcontroller/controller_test.go`](../../hostnetportcontroller/controller_test.go) | controller_test.go 单元测试 |
| [`hostnetportcontroller/node_controller.go`](../../hostnetportcontroller/node_controller.go) | node_controller 控制器逻辑 |
| [`hostnetportcontroller/nodefilter.go`](../../hostnetportcontroller/nodefilter.go) | nodefilter 控制器逻辑 |
| [`hostnetportcontroller/nodefilter_test.go`](../../hostnetportcontroller/nodefilter_test.go) | nodefilter_test.go 单元测试 |
| [`hostnetportcontroller/pod_controller.go`](../../hostnetportcontroller/pod_controller.go) | pod_controller 控制器逻辑 |
| [`hostnetportcontroller/podfilter.go`](../../hostnetportcontroller/podfilter.go) | podfilter 控制器逻辑 |
| [`hostnetportcontroller/podfilter_test.go`](../../hostnetportcontroller/podfilter_test.go) | podfilter_test.go 单元测试 |
| [`hostnetportcontroller/pool_controller.go`](../../hostnetportcontroller/pool_controller.go) | HostNetPortPool CRD Reconcile 控制器 |

## ingresscontroller/

| 文件 | 职责描述 |
|------|---------|
| [`ingresscontroller/endpointfilter.go`](../../ingresscontroller/endpointfilter.go) | endpointfilter 控制器逻辑 |
| [`ingresscontroller/ingress_controller.go`](../../ingresscontroller/ingress_controller.go) | Ingress CRD Reconcile 控制器 |
| [`ingresscontroller/multiclusereps_filter.go`](../../ingresscontroller/multiclusereps_filter.go) | multiclusereps_filter 控制器逻辑 |
| [`ingresscontroller/podfilter.go`](../../ingresscontroller/podfilter.go) | podfilter 控制器逻辑 |
| [`ingresscontroller/utils.go`](../../ingresscontroller/utils.go) | utils 控制器逻辑 |
| [`ingresscontroller/utils_test.go`](../../ingresscontroller/utils_test.go) | utils_test.go 单元测试 |

## internal/apiclient/

| 文件 | 职责描述 |
|------|---------|
| [`internal/apiclient/bkmapiclient.go`](../../internal/apiclient/bkmapiclient.go) | bkmapiclient 业务逻辑 |
| [`internal/apiclient/const.go`](../../internal/apiclient/const.go) | const 业务逻辑 |
| [`internal/apiclient/helper.go`](../../internal/apiclient/helper.go) | helper 业务逻辑 |
| [`internal/apiclient/portbindingitem_helper.go`](../../internal/apiclient/portbindingitem_helper.go) | portbindingitem_helper 业务逻辑 |
| [`internal/apiclient/types.go`](../../internal/apiclient/types.go) | types 业务逻辑 |

## internal/apiclient/xrequests/

| 文件 | 职责描述 |
|------|---------|
| [`internal/apiclient/xrequests/requests.go`](../../internal/apiclient/xrequests/requests.go) | requests 业务逻辑 |
| [`internal/apiclient/xrequests/utils.go`](../../internal/apiclient/xrequests/utils.go) | utils 业务逻辑 |

## internal/check/

| 文件 | 职责描述 |
|------|---------|
| [`internal/check/binding.go`](../../internal/check/binding.go) | Ingress SSL 证书 Binding 展开与去重 |
| [`internal/check/binding_test.go`](../../internal/check/binding_test.go) | binding_test.go 单元测试 |
| [`internal/check/certificatechecker.go`](../../internal/check/certificatechecker.go) | SSL 证书过期周期性检查器 |
| [`internal/check/certificatechecker_test.go`](../../internal/check/certificatechecker_test.go) | certificatechecker_test.go 单元测试 |
| [`internal/check/checkrunner.go`](../../internal/check/checkrunner.go) | checkrunner 一致性检查逻辑 |
| [`internal/check/checkrunner_test.go`](../../internal/check/checkrunner_test.go) | checkrunner_test.go 单元测试 |
| [`internal/check/hostnet_segment_checker.go`](../../internal/check/hostnet_segment_checker.go) | hostnet_segment_checker 一致性检查逻辑 |
| [`internal/check/hostnet_segment_checker_test.go`](../../internal/check/hostnet_segment_checker_test.go) | hostnet_segment_checker_test.go 单元测试 |
| [`internal/check/interface.go`](../../internal/check/interface.go) | interface 一致性检查逻辑 |
| [`internal/check/lbcachechecker.go`](../../internal/check/lbcachechecker.go) | lbcachechecker 一致性检查逻辑 |
| [`internal/check/listenerchecker.go`](../../internal/check/listenerchecker.go) | listenerchecker 一致性检查逻辑 |
| [`internal/check/port_leak_checker.go`](../../internal/check/port_leak_checker.go) | port_leak_checker 一致性检查逻辑 |
| [`internal/check/portbindchecker.go`](../../internal/check/portbindchecker.go) | portbindchecker 一致性检查逻辑 |
| [`internal/check/portbindchecker_test.go`](../../internal/check/portbindchecker_test.go) | portbindchecker_test.go 单元测试 |

## internal/cloud/

| 文件 | 职责描述 |
|------|---------|
| [`internal/cloud/interface.go`](../../internal/cloud/interface.go) | 云负载均衡适配器接口定义 |

## internal/cloud/aws/

| 文件 | 职责描述 |
|------|---------|
| [`internal/cloud/aws/aga_support_test.go`](../../internal/cloud/aws/aga_support_test.go) | aga_support_test.go 单元测试 |
| [`internal/cloud/aws/aga_supporter.go`](../../internal/cloud/aws/aga_supporter.go) | aga_supporter 云厂商适配逻辑 |
| [`internal/cloud/aws/constant.go`](../../internal/cloud/aws/constant.go) | constant 云厂商适配逻辑 |
| [`internal/cloud/aws/elb.go`](../../internal/cloud/aws/elb.go) | elb 云厂商适配逻辑 |
| [`internal/cloud/aws/error.go`](../../internal/cloud/aws/error.go) | error 云厂商适配逻辑 |
| [`internal/cloud/aws/helper.go`](../../internal/cloud/aws/helper.go) | helper 云厂商适配逻辑 |
| [`internal/cloud/aws/helper_test.go`](../../internal/cloud/aws/helper_test.go) | helper_test.go 单元测试 |
| [`internal/cloud/aws/sdk.go`](../../internal/cloud/aws/sdk.go) | sdk 云厂商适配逻辑 |
| [`internal/cloud/aws/sdkhelper.go`](../../internal/cloud/aws/sdkhelper.go) | sdkhelper 云厂商适配逻辑 |
| [`internal/cloud/aws/validate.go`](../../internal/cloud/aws/validate.go) | validate 云厂商适配逻辑 |
| [`internal/cloud/aws/validate_test.go`](../../internal/cloud/aws/validate_test.go) | validate_test.go 单元测试 |

## internal/cloud/azure/

| 文件 | 职责描述 |
|------|---------|
| [`internal/cloud/azure/azure.go`](../../internal/cloud/azure/azure.go) | azure 云厂商适配逻辑 |
| [`internal/cloud/azure/constant.go`](../../internal/cloud/azure/constant.go) | constant 云厂商适配逻辑 |
| [`internal/cloud/azure/error.go`](../../internal/cloud/azure/error.go) | error 云厂商适配逻辑 |
| [`internal/cloud/azure/helper.go`](../../internal/cloud/azure/helper.go) | helper 云厂商适配逻辑 |
| [`internal/cloud/azure/helper_test.go`](../../internal/cloud/azure/helper_test.go) | helper_test.go 单元测试 |
| [`internal/cloud/azure/resourcehelper.go`](../../internal/cloud/azure/resourcehelper.go) | resourcehelper 云厂商适配逻辑 |
| [`internal/cloud/azure/sdk.go`](../../internal/cloud/azure/sdk.go) | sdk 云厂商适配逻辑 |
| [`internal/cloud/azure/uitl.go`](../../internal/cloud/azure/uitl.go) | uitl 云厂商适配逻辑 |
| [`internal/cloud/azure/util_test.go`](../../internal/cloud/azure/util_test.go) | util_test.go 单元测试 |
| [`internal/cloud/azure/validate.go`](../../internal/cloud/azure/validate.go) | validate 云厂商适配逻辑 |

## internal/cloud/gcp/

| 文件 | 职责描述 |
|------|---------|
| [`internal/cloud/gcp/constant.go`](../../internal/cloud/gcp/constant.go) | constant 云厂商适配逻辑 |
| [`internal/cloud/gcp/gclb.go`](../../internal/cloud/gcp/gclb.go) | gclb 云厂商适配逻辑 |
| [`internal/cloud/gcp/helper.go`](../../internal/cloud/gcp/helper.go) | helper 云厂商适配逻辑 |
| [`internal/cloud/gcp/sdk.go`](../../internal/cloud/gcp/sdk.go) | sdk 云厂商适配逻辑 |
| [`internal/cloud/gcp/sdkhelper.go`](../../internal/cloud/gcp/sdkhelper.go) | sdkhelper 云厂商适配逻辑 |
| [`internal/cloud/gcp/util.go`](../../internal/cloud/gcp/util.go) | util 云厂商适配逻辑 |
| [`internal/cloud/gcp/validate.go`](../../internal/cloud/gcp/validate.go) | validate 云厂商适配逻辑 |

## internal/cloud/mock/

| 文件 | 职责描述 |
|------|---------|
| [`internal/cloud/mock/mockcloud.go`](../../internal/cloud/mock/mockcloud.go) | mockcloud 云厂商适配逻辑 |

## internal/cloud/namespacedlb/

| 文件 | 职责描述 |
|------|---------|
| [`internal/cloud/namespacedlb/namespacedclient.go`](../../internal/cloud/namespacedlb/namespacedclient.go) | 按 Namespace 隔离的云客户端 |
| [`internal/cloud/namespacedlb/namespacedclient_test.go`](../../internal/cloud/namespacedlb/namespacedclient_test.go) | namespacedclient_test.go 单元测试 |

## internal/cloud/namespacedssl/

| 文件 | 职责描述 |
|------|---------|
| [`internal/cloud/namespacedssl/namespacedclient.go`](../../internal/cloud/namespacedssl/namespacedclient.go) | 按 Namespace 隔离的 SSL 证书 API 客户端 |
| [`internal/cloud/namespacedssl/namespacedclient_test.go`](../../internal/cloud/namespacedssl/namespacedclient_test.go) | namespacedclient_test.go 单元测试 |

## internal/cloud/tencentcloud/

| 文件 | 职责描述 |
|------|---------|
| [`internal/cloud/tencentcloud/api.go`](../../internal/cloud/tencentcloud/api.go) | api 云厂商适配逻辑 |
| [`internal/cloud/tencentcloud/clb.go`](../../internal/cloud/tencentcloud/clb.go) | clb 云厂商适配逻辑 |
| [`internal/cloud/tencentcloud/constant.go`](../../internal/cloud/tencentcloud/constant.go) | constant 云厂商适配逻辑 |
| [`internal/cloud/tencentcloud/helper.go`](../../internal/cloud/tencentcloud/helper.go) | helper 云厂商适配逻辑 |
| [`internal/cloud/tencentcloud/helperbatch.go`](../../internal/cloud/tencentcloud/helperbatch.go) | helperbatch 云厂商适配逻辑 |
| [`internal/cloud/tencentcloud/sdk.go`](../../internal/cloud/tencentcloud/sdk.go) | sdk 云厂商适配逻辑 |
| [`internal/cloud/tencentcloud/sdkhelper.go`](../../internal/cloud/tencentcloud/sdkhelper.go) | sdkhelper 云厂商适配逻辑 |
| [`internal/cloud/tencentcloud/sharedratelimit.go`](../../internal/cloud/tencentcloud/sharedratelimit.go) | CLB/SSL API 共享限流器 |
| [`internal/cloud/tencentcloud/sharedratelimit_test.go`](../../internal/cloud/tencentcloud/sharedratelimit_test.go) | sharedratelimit_test.go 单元测试 |
| [`internal/cloud/tencentcloud/sslclient.go`](../../internal/cloud/tencentcloud/sslclient.go) | 腾讯云 SSL 证书 API 客户端 |
| [`internal/cloud/tencentcloud/sslclient_test.go`](../../internal/cloud/tencentcloud/sslclient_test.go) | sslclient_test.go 单元测试 |
| [`internal/cloud/tencentcloud/util.go`](../../internal/cloud/tencentcloud/util.go) | util 云厂商适配逻辑 |
| [`internal/cloud/tencentcloud/util_test.go`](../../internal/cloud/tencentcloud/util_test.go) | util_test.go 单元测试 |
| [`internal/cloud/tencentcloud/validate.go`](../../internal/cloud/tencentcloud/validate.go) | validate 云厂商适配逻辑 |

## internal/cloudcollector/

| 文件 | 职责描述 |
|------|---------|
| [`internal/cloudcollector/cloudcollector.go`](../../internal/cloudcollector/cloudcollector.go) | cloudcollector 云厂商适配逻辑 |
| [`internal/cloudcollector/helper.go`](../../internal/cloudcollector/helper.go) | helper 云厂商适配逻辑 |
| [`internal/cloudcollector/statuscache.go`](../../internal/cloudcollector/statuscache.go) | statuscache 云厂商适配逻辑 |

## internal/cloudnode/

| 文件 | 职责描述 |
|------|---------|
| [`internal/cloudnode/interface.go`](../../internal/cloudnode/interface.go) | interface 云厂商适配逻辑 |

## internal/cloudnode/native/

| 文件 | 职责描述 |
|------|---------|
| [`internal/cloudnode/native/defaultnodeclient.go`](../../internal/cloudnode/native/defaultnodeclient.go) | defaultnodeclient 云厂商适配逻辑 |

## internal/common/

| 文件 | 职责描述 |
|------|---------|
| [`internal/common/common.go`](../../internal/common/common.go) | common 业务逻辑 |

## internal/conflicthandler/

| 文件 | 职责描述 |
|------|---------|
| [`internal/conflicthandler/conflict.go`](../../internal/conflicthandler/conflict.go) | conflict 业务逻辑 |
| [`internal/conflicthandler/resource.go`](../../internal/conflicthandler/resource.go) | resource 业务逻辑 |
| [`internal/conflicthandler/resource_test.go`](../../internal/conflicthandler/resource_test.go) | resource_test.go 单元测试 |
| [`internal/conflicthandler/util.go`](../../internal/conflicthandler/util.go) | util 业务逻辑 |

## internal/constant/

| 文件 | 职责描述 |
|------|---------|
| [`internal/constant/constant.go`](../../internal/constant/constant.go) | 共享常量与 Annotation Key 定义 |
| [`internal/constant/webhook.go`](../../internal/constant/webhook.go) | Webhook 相关常量 |

## internal/eventer/

| 文件 | 职责描述 |
|------|---------|
| [`internal/eventer/eventer.go`](../../internal/eventer/eventer.go) | eventer 业务逻辑 |

## internal/generator/

| 文件 | 职责描述 |
|------|---------|
| [`internal/generator/ingressconverter.go`](../../internal/generator/ingressconverter.go) | Ingress → Listener 转换入口 |
| [`internal/generator/ingressconverter_test.go`](../../internal/generator/ingressconverter_test.go) | ingressconverter_test.go 单元测试 |
| [`internal/generator/listenerconverter.go`](../../internal/generator/listenerconverter.go) | listenerconverter 业务逻辑 |
| [`internal/generator/mappingconverter.go`](../../internal/generator/mappingconverter.go) | mappingconverter 业务逻辑 |
| [`internal/generator/namespace_scope_exempt_test.go`](../../internal/generator/namespace_scope_exempt_test.go) | namespace_scope_exempt_test.go 单元测试 |
| [`internal/generator/ruleconverter.go`](../../internal/generator/ruleconverter.go) | ruleconverter 业务逻辑 |
| [`internal/generator/util.go`](../../internal/generator/util.go) | util 业务逻辑 |
| [`internal/generator/util_test.go`](../../internal/generator/util_test.go) | util_test.go 单元测试 |

## internal/hostnetportpoolcache/

| 文件 | 职责描述 |
|------|---------|
| [`internal/hostnetportpoolcache/cache.go`](../../internal/hostnetportpoolcache/cache.go) | HostNetPortPool 内存缓存主逻辑 |
| [`internal/hostnetportpoolcache/cache_test.go`](../../internal/hostnetportpoolcache/cache_test.go) | cache_test.go 单元测试 |
| [`internal/hostnetportpoolcache/types.go`](../../internal/hostnetportpoolcache/types.go) | types 缓存相关逻辑 |

## internal/httpsvr/

| 文件 | 职责描述 |
|------|---------|
| [`internal/httpsvr/aga_support.go`](../../internal/httpsvr/aga_support.go) | aga_support HTTP API 处理器 |
| [`internal/httpsvr/check_bind_status.go`](../../internal/httpsvr/check_bind_status.go) | check_bind_status HTTP API 处理器 |
| [`internal/httpsvr/hostnetportpool.go`](../../internal/httpsvr/hostnetportpool.go) | hostnetportpool HTTP API 处理器 |
| [`internal/httpsvr/httpserver.go`](../../internal/httpsvr/httpserver.go) | go-restful HTTP 路由注册入口 InitRouters() |
| [`internal/httpsvr/ingress.go`](../../internal/httpsvr/ingress.go) | ingress HTTP API 处理器 |
| [`internal/httpsvr/listener.go`](../../internal/httpsvr/listener.go) | listener HTTP API 处理器 |
| [`internal/httpsvr/node.go`](../../internal/httpsvr/node.go) | node HTTP API 处理器 |
| [`internal/httpsvr/portpool.go`](../../internal/httpsvr/portpool.go) | portpool HTTP API 处理器 |
| [`internal/httpsvr/readiness_probe.go`](../../internal/httpsvr/readiness_probe.go) | readiness_probe HTTP API 处理器 |
| [`internal/httpsvr/response.go`](../../internal/httpsvr/response.go) | response HTTP API 处理器 |

## internal/ingresscache/

| 文件 | 职责描述 |
|------|---------|
| [`internal/ingresscache/cache.go`](../../internal/ingresscache/cache.go) | cache 缓存相关逻辑 |
| [`internal/ingresscache/interface.go`](../../internal/ingresscache/interface.go) | interface 缓存相关逻辑 |
| [`internal/ingresscache/util.go`](../../internal/ingresscache/util.go) | util 缓存相关逻辑 |

## internal/metrics/

| 文件 | 职责描述 |
|------|---------|
| [`internal/metrics/check.go`](../../internal/metrics/check.go) | check Prometheus 指标 |
| [`internal/metrics/certificate.go`](../../internal/metrics/certificate.go) | SSL 证书过期天数与查询成功指标 |
| [`internal/metrics/certificate_test.go`](../../internal/metrics/certificate_test.go) | certificate_test.go 单元测试 |
| [`internal/metrics/hostnetportpool.go`](../../internal/metrics/hostnetportpool.go) | hostnetportpool Prometheus 指标 |
| [`internal/metrics/listener_controller.go`](../../internal/metrics/listener_controller.go) | listener_controller Prometheus 指标 |
| [`internal/metrics/metric.go`](../../internal/metrics/metric.go) | Prometheus metrics 注册中心 |
| [`internal/metrics/nodeinfoexporter.go`](../../internal/metrics/nodeinfoexporter.go) | nodeinfoexporter Prometheus 指标 |
| [`internal/metrics/portpool.go`](../../internal/metrics/portpool.go) | portpool Prometheus 指标 |
| [`internal/metrics/webhook.go`](../../internal/metrics/webhook.go) | webhook Prometheus 指标 |

## internal/nodecache/

| 文件 | 职责描述 |
|------|---------|
| [`internal/nodecache/nodecache.go`](../../internal/nodecache/nodecache.go) | nodecache 缓存相关逻辑 |

## internal/option/

| 文件 | 职责描述 |
|------|---------|
| [`internal/option/option.go`](../../internal/option/option.go) | CLI 参数与 ControllerOption 配置 |
| [`internal/option/option_test.go`](../../internal/option/option_test.go) | option_test.go 单元测试 |

## internal/portpoolcache/

| 文件 | 职责描述 |
|------|---------|
| [`internal/portpoolcache/metric.go`](../../internal/portpoolcache/metric.go) | metric 缓存相关逻辑 |
| [`internal/portpoolcache/pool.go`](../../internal/portpoolcache/pool.go) | pool 缓存相关逻辑 |
| [`internal/portpoolcache/poolcache.go`](../../internal/portpoolcache/poolcache.go) | PortPool 内存缓存主逻辑 |
| [`internal/portpoolcache/poolcache_test.go`](../../internal/portpoolcache/poolcache_test.go) | poolcache_test.go 单元测试 |
| [`internal/portpoolcache/poolitem.go`](../../internal/portpoolcache/poolitem.go) | poolitem 缓存相关逻辑 |
| [`internal/portpoolcache/port.go`](../../internal/portpoolcache/port.go) | port 缓存相关逻辑 |
| [`internal/portpoolcache/types.go`](../../internal/portpoolcache/types.go) | types 缓存相关逻辑 |

## internal/utils/

| 文件 | 职责描述 |
|------|---------|
| [`internal/utils/patch_util.go`](../../internal/utils/patch_util.go) | patch_util 业务逻辑 |
| [`internal/utils/utils.go`](../../internal/utils/utils.go) | utils 业务逻辑 |

## internal/webhookserver/

| 文件 | 职责描述 |
|------|---------|
| [`internal/webhookserver/annotationparse.go`](../../internal/webhookserver/annotationparse.go) | annotationparse Webhook 处理逻辑 |
| [`internal/webhookserver/annotationparse_test.go`](../../internal/webhookserver/annotationparse_test.go) | annotationparse_test.go 单元测试 |
| [`internal/webhookserver/convert.go`](../../internal/webhookserver/convert.go) | convert Webhook 处理逻辑 |
| [`internal/webhookserver/errors.go`](../../internal/webhookserver/errors.go) | errors Webhook 处理逻辑 |
| [`internal/webhookserver/mutating.go`](../../internal/webhookserver/mutating.go) | mutating Webhook 处理逻辑 |
| [`internal/webhookserver/mutating_ingress.go`](../../internal/webhookserver/mutating_ingress.go) | mutating_ingress Webhook 处理逻辑 |
| [`internal/webhookserver/mutating_node.go`](../../internal/webhookserver/mutating_node.go) | mutating_node Webhook 处理逻辑 |
| [`internal/webhookserver/portallocate.go`](../../internal/webhookserver/portallocate.go) | portallocate Webhook 处理逻辑 |
| [`internal/webhookserver/scheme.go`](../../internal/webhookserver/scheme.go) | scheme Webhook 处理逻辑 |
| [`internal/webhookserver/utils.go`](../../internal/webhookserver/utils.go) | utils Webhook 处理逻辑 |
| [`internal/webhookserver/validate.go`](../../internal/webhookserver/validate.go) | validate Webhook 处理逻辑 |
| [`internal/webhookserver/validate_delete.go`](../../internal/webhookserver/validate_delete.go) | validate_delete Webhook 处理逻辑 |
| [`internal/webhookserver/validate_delete_portpool.go`](../../internal/webhookserver/validate_delete_portpool.go) | validate_delete_portpool Webhook 处理逻辑 |
| [`internal/webhookserver/validate_test.go`](../../internal/webhookserver/validate_test.go) | validate_test.go 单元测试 |
| [`internal/webhookserver/webhookserver.go`](../../internal/webhookserver/webhookserver.go) | Admission Webhook 服务入口 |

## internal/worker/

| 文件 | 职责描述 |
|------|---------|
| [`internal/worker/cache.go`](../../internal/worker/cache.go) | cache 业务逻辑 |
| [`internal/worker/cache_test.go`](../../internal/worker/cache_test.go) | cache_test.go 单元测试 |
| [`internal/worker/event.go`](../../internal/worker/event.go) | event 业务逻辑 |
| [`internal/worker/event_test.go`](../../internal/worker/event_test.go) | event_test.go 单元测试 |
| [`internal/worker/options.go`](../../internal/worker/options.go) | options 业务逻辑 |
| [`internal/worker/synchronizer.go`](../../internal/worker/synchronizer.go) | synchronizer 业务逻辑 |
| [`internal/worker/synchronizer_test.go`](../../internal/worker/synchronizer_test.go) | synchronizer_test.go 单元测试 |
| [`internal/worker/util.go`](../../internal/worker/util.go) | util 业务逻辑 |

## listenercontroller/

| 文件 | 职责描述 |
|------|---------|
| [`listenercontroller/listenerHelper.go`](../../listenercontroller/listenerHelper.go) | listenerHelper 控制器逻辑 |
| [`listenercontroller/listener_bypass_controller.go`](../../listenercontroller/listener_bypass_controller.go) | listener_bypass_controller 控制器逻辑 |
| [`listenercontroller/listener_controller.go`](../../listenercontroller/listener_controller.go) | Listener CRD Reconcile 控制器 |

## namespacecontroller/

| 文件 | 职责描述 |
|------|---------|
| [`namespacecontroller/namespacecontroller.go`](../../namespacecontroller/namespacecontroller.go) | Namespace 变更监听控制器 |
| [`namespacecontroller/utils.go`](../../namespacecontroller/utils.go) | utils 控制器逻辑 |

## nodecontroller/

| 文件 | 职责描述 |
|------|---------|
| [`nodecontroller/nodecontroller.go`](../../nodecontroller/nodecontroller.go) | Node 元数据缓存控制器 |

## portbindingcontroller/

| 文件 | 职责描述 |
|------|---------|
| [`portbindingcontroller/event.go`](../../portbindingcontroller/event.go) | event 控制器逻辑 |
| [`portbindingcontroller/nodefilter.go`](../../portbindingcontroller/nodefilter.go) | nodefilter 控制器逻辑 |
| [`portbindingcontroller/nodeportbindinghandler.go`](../../portbindingcontroller/nodeportbindinghandler.go) | nodeportbindinghandler 控制器逻辑 |
| [`portbindingcontroller/podfilter.go`](../../portbindingcontroller/podfilter.go) | podfilter 控制器逻辑 |
| [`portbindingcontroller/podportbindinghandler.go`](../../portbindingcontroller/podportbindinghandler.go) | podportbindinghandler 控制器逻辑 |
| [`portbindingcontroller/portbinding.go`](../../portbindingcontroller/portbinding.go) | portbinding 控制器逻辑 |
| [`portbindingcontroller/portbindingbypasscontroller.go`](../../portbindingcontroller/portbindingbypasscontroller.go) | portbindingbypasscontroller 控制器逻辑 |
| [`portbindingcontroller/portbindingcontroller.go`](../../portbindingcontroller/portbindingcontroller.go) | PortBinding CRD Reconcile 控制器 |
| [`portbindingcontroller/portbindingitem.go`](../../portbindingcontroller/portbindingitem.go) | portbindingitem 控制器逻辑 |
| [`portbindingcontroller/util.go`](../../portbindingcontroller/util.go) | util 控制器逻辑 |

## portpoolcontroller/

| 文件 | 职责描述 |
|------|---------|
| [`portpoolcontroller/portpool.go`](../../portpoolcontroller/portpool.go) | portpool 控制器逻辑 |
| [`portpoolcontroller/portpoolcontroller.go`](../../portpoolcontroller/portpoolcontroller.go) | PortPool CRD Reconcile 控制器 |
| [`portpoolcontroller/portpoolitem.go`](../../portpoolcontroller/portpoolitem.go) | portpoolitem 控制器逻辑 |
| [`portpoolcontroller/util.go`](../../portpoolcontroller/util.go) | util 控制器逻辑 |

## bcs-ingress-inspector/

| 文件 | 职责描述 |
|------|---------|
| [`bcs-ingress-inspector/main.go`](../../bcs-ingress-inspector/main.go) | 独立诊断二进制入口 |
| [`bcs-ingress-inspector/option/option.go`](../../bcs-ingress-inspector/option/option.go) | 诊断工具 CLI 参数 |
| [`bcs-ingress-inspector/portbindingcontroller/portbindingcontroller.go`](../../bcs-ingress-inspector/portbindingcontroller/portbindingcontroller.go) | PortBinding 诊断控制器 |

## cli-util/validate-listener-name/

| 文件 | 职责描述 |
|------|---------|
| [`cli-util/validate-listener-name/main.go`](../../cli-util/validate-listener-name/main.go) | Listener 名称校验 CLI 入口 |
| [`cli-util/validate-listener-name/pkg/handler.go`](../../cli-util/validate-listener-name/pkg/handler.go) | 校验逻辑处理器 |
| [`cli-util/validate-listener-name/pkg/options.go`](../../cli-util/validate-listener-name/pkg/options.go) | CLI 参数定义 |

## root/

| 文件 | 职责描述 |
|------|---------|
| [`main.go`](../../main.go) | 程序入口，注册所有 Controller、Checker、HTTP 服务与 Webhook |
| [`main_test.go`](../../main_test.go) | 主程序单元测试（如 parseExemptNamespaces） |

<!-- /dev-map:auto -->

---

*由 harness-generating 自动扫描生成，最后更新：2026-06-11（gardening 增量同步证书过期特性）*