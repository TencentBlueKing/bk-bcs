### 资源描述
prometheus query 时间范围查询

### 参数&返回
- 接口兼容prometheus query接口，具体参考[官方文档](https://prometheus.io/docs/prometheus/latest/querying/api/#instant-queries)
- label 需要带上集群 id(cluster_id=)

### 公共 Header 参数

| 参数名  |  类型  |  是否必须 | 说明 | 
| ------------ | ------------ | ------------ | -------- |
| X-Tenant-Project-Code  | string  | 是 | 项目Code |
| X-Operator  | string  | 否 | 操作人 |