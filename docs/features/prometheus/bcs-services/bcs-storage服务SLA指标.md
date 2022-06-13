# bcs-storage服务SLA指标

指标说明请参考github.com/Tencent/bk-bcs/docs/features/prometheus/bcs-services/bcs-storage服务metrics指标.md

## 指标聚合

### 查询请求

#### 数据查询成功率

```
# 数据查询请求成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/dynamic.*",method=~"GET|POST",status="2xx"} or bkbcs_storage_api_request_total_num{handler=~".*/metric/clusters.*",method="GET",status="2xx"} or bkbcs_storage_api_request_total_num{handler=~".*/alarms.*",method="GET",status="2xx"} or bkbcs_storage_api_request_total_num{handler=~".*/events.*",method="GET",status="2xx"} or bkbcs_storage_api_request_total_num{handler=~".*/cluster_config.*",method="GET",status="2xx"} or bkbcs_storage_api_request_total_num{handler=~".*/host.*",method="GET",status="2xx"}) by (instance) / sum(bkbcs_storage_api_request_total_num{handler=~".*/dynamic.*",method=~"GET|POST"} or bkbcs_storage_api_request_total_num{handler=~".*/metric/clusters.*",method="GET"} or bkbcs_storage_api_request_total_num{handler=~".*/alarms.*",method="GET"} or bkbcs_storage_api_request_total_num{handler=~".*/events.*",method="GET"} or bkbcs_storage_api_request_total_num{handler=~".*/cluster_config.*",method="GET"} or bkbcs_storage_api_request_total_num{handler=~".*/host.*",method="GET"}) by (instance)

# 动态数据查询成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/dynamic.*",method=~"GET|POST",status="2xx"})/sum(bkbcs_storage_api_request_total_num{handler=~".*/dynamic.*",method=~"GET|POST"})

# metric数据查询成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/metric/clusters.*",method="GET",status="2xx"})/sum(bkbcs_storage_api_request_total_num{handler=~".*/metric/clusters.*",method="GET"})

# event数据查询成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/events.*",method="GET",status="2xx"})/sum(bkbcs_storage_api_request_total_num{handler=~".*/events.*",method="GET"})

# alarm数据查询成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/alarms.*",method="GET",status="2xx"})/sum(bkbcs_storage_api_request_total_num{handler=~".*/alarms.*",method="GET"})

# 集群配置查询成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/cluster_config.*",method="GET",status="2xx"})/sum(bkbcs_storage_api_request_total_num{handler=~".*/cluster_config.*",method="GET"})

# host配置查询成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/host.*",method="GET",status="2xx"})/sum(bkbcs_storage_api_request_total_num{handler=~".*/host.*",method="GET"})
```

#### 数据查询延迟

```
# 数据查询请求延时<50ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/dynamic.*",method=~"GET|POST",status="2xx",le="0.05"} or bkbcs_storage_api_request_duration_time_bucket{handler=~"/metric/clusters.*",method="GET",status="2xx",le="0.05"} or bkbcs_storage_api_request_duration_time_bucket{handler=~"/alarms.*",method="GET",status="2xx",le="0.05"} or bkbcs_storage_api_request_duration_time_bucket{handler=~"/events.*",method="GET",status="2xx",le="0.05"} or bkbcs_storage_api_request_duration_time_bucket{handler=~"/cluster_config.*",method="GET",status="2xx",le="0.05"} or bkbcs_storage_api_request_duration_time_bucket{handler=~"/host.*",method="GET",status="2xx",le="0.05"}) by (instance) / sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/dynamic.*",method=~"GET|POST"} or bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~"/metric/clusters.*",method="GET"} or bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~"/alarms.*",method="GET"} or bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~"/events.*",method="GET"} or bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~"/cluster_config.*",method="GET"} or bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~"/host.*",method="GET"}) by (instance)

# 动态数据查询延时<50ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/dynamic.*",method=~"GET|POST",status="2xx",le="0.05"})/sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/dynamic.*",method=~"GET|POST"})

# metric数据查询延时<50ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/metric/clusters.*",method="GET",status="2xx",le="0.05"})/sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/metric/clusters.*",method="GET"})

# event数据查询延时<50ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/event.*",method="GET",status="2xx",le="0.05"})/sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/mevent.*",method="GET"})

# alarm数据查询延时<50ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/alarms.*",method="GET",status="2xx",le="0.05"})/sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/alarms.*",method="GET"})

# 集群配置查询延时<50ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/cluster_config.*",method="GET",status="2xx",le="0.05"})/sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/cluster_config.*",method="GET"})

# host配置查询延时<50ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/host.*",method="GET",status="2xx",le="0.05"})/sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/host.*",method="GET"})
```

### 上报更新请求

#### API上报/更新请求成功率

```
# API上报/更新请求成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/dynamic.*",method="PUT",status="2xx"} or bkbcs_storage_api_request_total_num{handler=~".*/metric/clusters.*",method="PUT",status="2xx"} or bkbcs_storage_api_request_total_num{handler=~".*/alarms.*",method="POST",status="2xx"} or bkbcs_storage_api_request_total_num{handler=~".*/events.*",method="PUT",status="2xx"} or bkbcs_storage_api_request_total_num{handler=~".*/cluster_config.*",method="PUT",status="2xx"} or bkbcs_storage_api_request_total_num{handler=~".*/host.*",method="PUT",status="2xx"}) by (instance) / sum(bkbcs_storage_api_request_total_num{handler=~".*/dynamic.*",method="PUT"} or bkbcs_storage_api_request_total_num{handler=~".*/metric/clusters.*",method="PUT"} or bkbcs_storage_api_request_total_num{handler=~".*/alarms.*",method="POST"} or bkbcs_storage_api_request_total_num{handler=~".*/events.*",method="PUT"} or bkbcs_storage_api_request_total_num{handler=~".*/cluster_config.*",method="PUT"} or bkbcs_storage_api_request_total_num{handler=~".*/host.*",method="PUT"}) by (instance)

# 动态数据上报/更新成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/dynamic.*",method="PUT",status="2xx"})/sum(bkbcs_storage_api_request_total_num{handler=~".*/dynamic.*",method="PUT"})

# metric数据上报/更新成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/metric/clusters.*",method="PUT",status="2xx"})/sum(bkbcs_storage_api_request_total_num{handler=~".*/metric/clusters.*",method="PUT"})

# event数据上报/更新成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/events.*",method="PUT",status="2xx"})/sum(bkbcs_storage_api_request_total_num{handler=~".*/events.*",method="PUT"})

# alarm数据上报/更新成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/alarms.*",method="POST",status="2xx"})/sum(bkbcs_storage_api_request_total_num{handler=~".*/alarms.*",method="POST"})

# 集群配置上报/更新成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/cluster_config.*",method="GET",status="2xx"})/sum(bkbcs_storage_api_request_total_num{handler=~".*/cluster_config.*",method="GET"})

# host配置上报/更新成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/host.*",method="PUT",status="2xx"})/sum(bkbcs_storage_api_request_total_num{handler=~".*/host.*",method="PUT"})
```

#### 数据上报/更新请求延时

```
# 数据上报/更新请求延时<250ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/dynamic.*",method=~"GET|POST",status="2xx",le="0.05"} or bkbcs_storage_api_request_duration_time_bucket{handler=~"/metric/clusters.*",method="GET",status="2xx",le="0.05"} or bkbcs_storage_api_request_duration_time_bucket{handler=~"/alarms.*",method="GET",status="2xx",le="0.05"} or bkbcs_storage_api_request_duration_time_bucket{handler=~"/events.*",method="GET",status="2xx",le="0.05"} or bkbcs_storage_api_request_duration_time_bucket{handler=~"/cluster_config.*",method="GET",status="2xx",le="0.05"} or bkbcs_storage_api_request_duration_time_bucket{handler=~"/host.*",method="GET",status="2xx",le="0.05"}) by (instance) / sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/dynamic.*",method=~"GET|POST"} or bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~"/metric/clusters.*",method="GET"} or bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~"/alarms.*",method="GET"} or bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~"/events.*",method="GET"} or bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~"/cluster_config.*",method="GET"} or bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~"/host.*",method="GET"}) by (instance)

# 动态数据上报/更新请求延时<250ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/dynamic.*",method="PUT",status="2xx",le="0.25"})/ sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/dynamic.*",method="PUT"})

# metric数据上报/更新请求延时<250ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/metric/clusters.*",method="PUT",status="2xx",le="0.25"})/ sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/metric/clusters.*",method="PUT"})

# event数据上报/更新请求延时<250ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/event.*",method="PUT",status="2xx",le="0.25"})/sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/event.*",method="PUT"})

# alarm数据上报/更新请求延时<250ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/alarms.*",method="POST",status="2xx",le="0.25"})/ sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/alarms.*",method="POST"})

# 集群配置上报/更新请求延时<250ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/cluster_config.*",method="PUT",status="2xx",le="0.25"})/sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/cluster_config.*",method="PUT"})

# host配置上报/更新请求延时<250ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/host.*",method="PUT",status="2xx",le="0.25"})/sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/host.*",method="PUT"})
```

### 删除请求

#### 数据删除请求成功率

```
# 数据删除请求成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/dynamic.*",method="DELETE",status="2xx"} or bkbcs_storage_api_request_total_num{handler=~".*/metric/clusters.*",method="DELETE",status="2xx"} or bkbcs_storage_api_request_total_num{handler=~".*/host.*",method="PUT",status="2xx"}) by (instance) / sum(bkbcs_storage_api_request_total_num{handler=~".*/dynamic.*",method="DELETE"} or bkbcs_storage_api_request_total_num{handler=~".*/metric/clusters.*",method="DELETE"} or bkbcs_storage_api_request_total_num{handler=~".*/host.*",method="DELETE"}) by (instance)

# 动态数据删除成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/dynamic.*",method="DELETE",status="2xx"})/sum(bkbcs_storage_api_request_total_num{handler=~".*/dynamic.*",method="DELETE"})

# metric数据删除成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/metric/clusters.*",method="DELETE",status="2xx"})/sum(bkbcs_storage_api_request_total_num{handler=~".*/metric/clusters.*",method="DELETE"})

# host配置删除成功率
sum(bkbcs_storage_api_request_total_num{handler=~".*/host.*",method="DELETE",status="2xx"})/sum(bkbcs_storage_api_request_total_num{handler=~".*/host.*",method="DELETE"})
```

#### 数据删除延迟

```
# 数据删除请求延时<250ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/dynamic.*",method=~"DELETE",status="2xx",le="0.25"} or bkbcs_storage_api_request_duration_time_bucket{handler=~"/metric/clusters.*",method="DELETE",status="2xx",le="0.25"} or bkbcs_storage_api_request_duration_time_bucket{handler=~"/host.*",method="DELETE",status="2xx",le="0.25"}) by (instance) / sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/dynamic.*",method=~"DELETE"} or bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~"/metric/clusters.*",method="DELETE"} or bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~"/alarms.*",method="POST"} or bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~"/host.*",method="DELETE"}) by (instance)

# 动态数据删除延迟<250ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/dynamic.*",method="DELETE",status="2xx",le="0.25"})/sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/dynamic.*",method="DELETE"})

# metric数据删除延迟<250ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/metric/clusters.*",method="DELETE",status="2xx",le="0.25"})/sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/metric/clusters.*",method="DELETE"})

# host配置删除延迟<250ms
sum(bkbcs_storage_api_request_duration_time_bucket{handler=~".*/host.*",method="DELETE",status="2xx",le="0.25"})/sum(bkbcs_storage_api_request_duration_time_count{status="2xx",handler=~".*/host.*",method="DELETE"})
```

