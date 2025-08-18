### 资源描述
注册资源的简单描述

### 公共 Header 参数

| 参数名  |  类型  |  是否必须 | 说明 | 
| ------------ | ------------ | ------------ | -------- |
| X-Tenant-Project-Code  | string  | 是 | 项目Code 不能同时为空 |
| X-Operator  | string  | 否 | 操作人 |

### 参数&返回
- 接口兼容prometheus series 接口，具体参考[官方文档](https://prometheus.io/docs/prometheus/latest/querying/api/)