### 资源描述
注册资源的简单描述

### 输入参数说明
|   参数名称   |    参数类型  |  必须  |     参数说明     |
| ------------ | ------------ | ------ | ---------------- |
| app_code   | string | 是 | 应用ID(app id)，可以通过 蓝鲸开发者中心 -> 应用基本设置 -> 基本信息 -> 鉴权信息 获取 |
| app_secret | string | 否 | 安全秘钥(app secret)，可以通过 蓝鲸开发者中心 -> 应用基本设置 -> 基本信息 -> 鉴权信息 获取 |

### 公共 Header 参数

| 参数名  |  类型  |  是否必须 | 说明 | 
| ------------ | ------------ | ------------ | -------- |
| X-Tenant-Project-Id  | string  | 是 | 项目ID |
| X-Tenant-Project-Code  | string  | 是 | 项目Code 不能同时为空 |
| X-Operator  | string  | 否 | 操作人 |

### 参数&返回
- 接口兼容prometheus query_range 接口，具体参考[官方文档](https://prometheus.io/docs/prometheus/latest/querying/api/#range-queries)