# bcs-storage服务metrics指标

## 指标

### 访问bcs-storage服务metrics指标
####  bkbcs\_storage\_http\_request\_total
* 标识API访问请求数指标
* label：handler,method,code,cluster_id,resource_type

#### bkbcs\_storage\_http\_request\_duration\_seconds
* 标识API访问请求的延迟指标
* label：handler, method, code,cluster_id, resource_type


#### bkbcs\_storage\_http\_requests\_inflight
* 标识接口维度的并发访问数
* label：handler, method

#### bkbcs\_storage\_http\_response\_size\_bytes 
* 标识接口访问返回的body体大小
* label：handler,method,code

## 指标聚合
### bcs-storage服务API接口指标聚合
#### storage API 全量聚合
```
storage接口请求总数
sum(bkbcs_storage_http_request_total)
 
storage API 每秒请求数
sum(rate(bkbcs_storage_http_request_total[5m]))
 
storage API 时延
sum(bkbcs_storage_http_request_duration_seconds_sum{}) / sum(bkbcs_storage_http_request_duration_seconds_count{})
 
storage API 成功率
sum(bkbcs_storage_http_request_total{code="2xx"}) / sum(bkbcs_storage_http_request_total)
```

#### storage API接口维度指标聚合
```
storage接口维度请求总数
sum(bkbcs_storage_http_request_total{}) by(handler)
 
storage接口维度时延
(sum(bkbcs_storage_http_request_duration_seconds_sum) by (handler)) / (sum(bkbcs_storage_http_request_duration_seconds_count) by (handler))
 
storage接口维度qps
sum(rate(bkbcs_storage_http_request_total[5m])) by (handler)
sum(irate(bkbcs_storage_http_request_total[5m])) by (handler)
 
storage接口维度成功率
sum(bkbcs_storage_http_request_total{status="2xx"}) by(handler) / sum(bkbcs_storage_http_request_total) by(handler)
``` 

#### storage API接口方法维度指标聚合
```
storage接口方法维度请求总数
sum(bkbcs_storage_http_request_total{}) by(handler,method)

storage接口方法维度时延
(sum(bkbcs_storage_http_request_duration_seconds_sum) by (handler,method)) / (sum(bkbcs_storage_http_request_duration_seconds_count) by (handler,method))

storage接口方法维度qps
sum(rate(bkbcs_storage_http_request_total[5m])) by (handler,method)
sum(irate(bkbcs_storage_http_request_total[5m])) by (handler,method)

storage接口方法维度成功率
sum(bkbcs_storage_http_request_total{code="2xx"}) by(handler,method) / sum(bkbcs_storage_http_request_total) by(handler,method)

```

#### 接口维度并发量聚合
```
接口维度并发量
sum(bkbcs_storage_http_requests_inflight) by(handler, method)
```

#### dynamic库接口聚合指标
通过`cluster_id` 和 `resource_type`能够唯一确定dynamic数据库的表类型

```
表类型请求总数
sum(bkbcs_storage_http_request_total{}) by(cluster_id, resource_type)

表类型的请求延迟
(sum(bkbcs_storage_http_request_duration_seconds_sum) by(cluster_id, resource_type)) / (sum(bkbcs_storage_http_request_duration_seconds_count) by(cluster_id, resource_type))

表类型的请求qps
sum(rate(bkbcs_storage_http_request_total[5m])) by(cluster_id, resource_type)

表类型请求的成功率
sum(bkbcs_storage_http_request_total{code="2xx"}) by(cluster_id, resource_type) / sum(bkbcs_storage_http_request_total) by(cluster_id, resource_type)

```
   