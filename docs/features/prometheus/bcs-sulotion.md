# BCS系统prometheus方案

BCS接入prometheus系统，实现数据上报，完成系统状态统计，异常数据告警功能。

状态：初稿

* 命名规则pre：bkbcs
* 层级、集群相关：mesos/k8s/service
* 模块相关：dns，storage，api等

例如mesos-datawatch数据索引命名范例

bkbcs_mesos_datawatch_sync_total: mesos-datawatch同步总次数

## 接入方案

![接入模型](./prometheus-intergration.png)

## bcs-api

bcs-api模块相关metrics指标

* api_requests_total: API请求总数，类型CounterVec，Labels区分不同api类型
* api_requests_err_total: API请求错误数，类型CounterVec，Labels区分不同api类型
* api_requests_latency_milliseconds: API请求延迟统计，类型HistogramVec，Labels区分不同api类型

## bcs-storage

* storage_requests_total: API请求总数，类型CounterVec，Labels区分不同api类型
* storage_requests_err_total: API请求错误数，类型CounterVec，Labels区分不同api类型
* storage_requests_latency_milliseconds: API请求延迟统计，类型HistogramVec，Labels区分不同api类型

## bcs-dns

* dns_total: 域名缓存数量，Gauge
* dns_request_total: dns请求次数统计，CounterVec，success与failed使用label区分
* dns_request_proxy_total: 转发给bcs service DNS次数，CounterVec，success/failure
* dns_request_out_proxy_total: 转发给外部DNS次数，CounterVec，success/failure
* dns_request_latency_milliseconds: dns查找延时，Histogram
* dns_storage_notify_total: zookeeper通知次数统计，CounterVec，Add/Update/Delete
* dns_storage_operator_total: 写入etcd操作统计，CounterVec，增删改查区分
* dns_storage_operator_latency_milliseconds: 写入etcd操作延时，HistogramVec，增删改查区分

## bcs-health

## bcs-loadbalance

* loadbalance_tunnel_state: tgw ip tunnel状态，0为异常，1为正常，Gauge
* loadbalance_zookeeper_state: 与zookeeper链接状态，0为异常，1为正常，Gauge
* loadbalance_zookeeper_notify_total: zookeeper事件通知次数，CounterVec，事件类型区分
* loadbalance_render_cfg_total：渲染haproxy/nginx配置次数，CounterVec，success/failure区分
* loadbalance_refresh_cfg_total: 刷新haproxy，nginx配置次数，CounterVec，success/failure区分
* loadbalance_restart_proxy_total：重启haproxy/nginx配置次数，CounterVec，success/failure区分

## bcs-mesos-driver

## bcs-scheduler

* bkbcs_scheduler_schedule_taskgroup_total: 容器调度次数，CounterVec，labels区分namespace、application、taskgroup、type(launch、reschedule、scale、update)
* bkbcs_scheduler_schedule_taskgroup_latency_ms: 容器调度耗时，类型Histogram，labels区分namespace、application、taskgroup、type(launch、reschedule、scale、update)
* bkbcs_scheduler_operate_application_total：操作应用次数，类型CounterVec，labels区分namespace、application、type(launch、delete、scale、update、rollingupdate)
* bkbcs_scheduler_operate_application_latency_second: 操作应用耗时，类型Histogram，labels区分namespace、application、type(launch、delete、scale、update、rollingupdate)
* bkbcs_scheduler_object_resource_info: 容器各类资源信息，类型GaugeVec，labels区分resource(service,deployment,application,configmap,secret)、namespace、name，
如果resource=application value表示application状态，0表示Staging、Deploying、Operating、RollingUpdate; 1表示Running；2表示Finish；3表示Abnormal,Error
* bkbcs_scheduler_taskgroup_info：容器taskgroup资源信息，类型GaugeVec，labels区分namespace、application、taskgroup，
value表示taskgroup状态，0表示Staging、Starting;1表示Running;2表示Finish、Killing、Killed;3表示Error、Failed;4表示Lost
* bkbcs_scheduler_agent_cpu_resource_total: slave机器cpu总数，类型GaugeVec，labels区分InnerIP
* bkbcs_scheduler_agent_cpu_resource_remain：slave机器cpu剩余数量，类型GaugeVec，labels区分InnerIP
* bkbcs_scheduler_agent_memory_resource_total: slave机器memory总数，类型GaugeVec，labels区分InnerIP，单位：MB
* bkbcs_scheduler_agent_memory_resource_remain：slave机器memory剩余数量，类型GaugeVec，labels区分InnerIP，单位：MB
* bkbcs_scheduler_storage_operator_total: scheduler操作存储次数，CounterVec，labels区分operator(create、delete、update、fetch)
* bkbcs_scheduler_storage_operator_latency_ms: scheudler操作存储耗时，HistogramVec，labels区分operator(create、delete、update、fetch)
* bkbcs_scheduler_taskgroup_report_total: 接收taskgroup上报次数，CounterVec，labels区分namespace、application、taskgroup、status(Staging、Starting、Running、Finish、Error、Killing、Killed、Failed、Lost)

## bcs-mesos-datawatch

* bkbcs_datawatch_mesos_storage_total: storage API请求总数，类型CounterVec，Labels区分不同同步数据类型
* bkbcs_datawatch_mesos_storage_latency_total: storage API请求延迟统计，类型HistogramVec，Labels区分不同api类型
* bkbcs_datawatch_mesos_sync_total: zookeeper事件触发次数统计，CounterVec，Labels区分不同事件
* bkbcs_datawatch_mesos_storage_state: bcs-storage服务发现状态，Gauge，正常为1，异常为0
* bkbcs_datawatch_mesos_cluster_state：datawatch集群状态，正常为1，其他皆为异常
* bkbcs_datawatch_mesos_role_state：datawatch角色状态，master为1，其余为slave

## bcs-contaienr-executor

* bkbcs_executor_slave_connection: 与slave链接数，0无连接，1有链接，Gauge
* bkbcs_executor_taskgroup_status_report_total: taskgroup上报次数，CounterVec，labels区分taskgroup
* bkbcs_executor_taskgroup_status_ack_total: scheduler确认taskgroup数据，CounterVec，labels区分taskgroup

## bcs-netservice

总体数据：
* ip_pools_total: 管理的地址池数量，Gauge
* ip_available_total: 可用IP数量，GaugeVec，不同地址池Labels拆分
* ip_active_total: 已用IP数量，GaugeVec，不同地址池使用Labels拆分
* ip_reserved_total: 保留IP数量，GaugeVec，不同地址池使用Labels区分

请求数据：
* ip_request_total: IP地址申请和释放请求次数，CounterVec，lease/release区分
* ip_request_err_total: IP地址申请和释放请求错误次数，CounterVec，lease/release区分
* ip_request_latency_seconds: IP地址申请和释放请求延时，HistogramVec，lease/release区分

存储对接统计：

* storage_operator_total: 与存储交互总次数，包含错误，CounterVec，lock/unlock/lease/release
* storage_operator_latency_seconds: 与存储交互总次数，CounterVec，lock/unlock/lease/release

