### 资源描述
prometheus query

### 参数&返回
- 接口兼容prometheus query接口，具体参考[官方文档](https://prometheus.io/docs/prometheus/latest/querying/api/#instant-queries)
- label 需要带上集群 id(cluster_id=)

### 公共 Header 参数

| 参数名  |  类型  |  是否必须 | 说明 | 
| ------------ | ------------ | ------------ | -------- |
| X-Tenant-Project-Code  | string  | 是 | 项目Code |
| X-Operator  | string  | 否 | 操作人 |

### 示例

```
curl "https://bcs-api-gateway.apigw.com/prod/v4/monitor/query/api/v1/query?query=max(irate(node_disk_io_time_seconds_total%7Bcluster_id%3D%22BCS-K8S-12345%22,instance%3D~%22%5E1.1.1.1%3A%22))%20*%20100" \
     -H 'X-Tenant-Project-Code: xxx'
     -H 'X-Bkapi-Authorization: {"bk_app_code": "xxx", "bk_app_secret": "xxx"}'
```