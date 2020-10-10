# Bcs Log Manager

bcs-log-manager 结合 bcs-logbeat-sidecar 及其所定义的 BcsLogConfigs 自定义资源，对所有集群的日志配置进行管理，包括日志配置的获取、下发和删除，并提供接口便捷申请蓝鲸数据平台 dataid 与配置数据清洗规则，仅需少量步骤即可快速将日志接入蓝鲸数据平台。

## 特性

- 通过 API Gateway 对集群日志配置进行批量管理操作
- 对接蓝鲸数据平台，通过 API Gateway 或自定义资源 BKDataApiConfig 进行dataid申请与数据清洗策略配置操作

## 资源定义

### BcsLogConfig

详见 bcs-logbeat-sidecar [文档](../../docs/features/bcs-webhook-server/log-controller.md)

### BKDataApiConfig

对接蓝鲸数据平台的申请 dataid、配置数据清洗策略的功能型资源，通过创建资源以发出请求，等待资源更新，获取请求的返回结果。

```yaml
apiVersion: bkbcs.tencent.com/v1
kind: BKDataApiConfig
metadata:
  name: obtain-dataid-test
  namespace: default
spec:
  apiName: v3_access_deploy_plan_post   (v3_access_deploy_plan_post for obtain dataid, 
                                         v3_databus_cleans_post for create data clean strategy)
  accessDeployPlanConfig:               (申请 dataid 的请求内容)
    bkAppCode:                          (蓝鲸应用ID)
    bkUserName:                         (蓝鲸用户名)
    bkAppSecret:                        (蓝鲸应用 secret)
    bkdataAuthenticationMethod:         (内部版为 user)
    bkBizID:                            (业务ID)
    DataScenario:                       (接入方式，填 custom)
    description:                        (数据接入描述)
    appenv:                             (内部版为 ieod)
    accessRawData:                      (数据接入的个性化配置)
      rawDataName:                      (数据名称)
      maintainer:                       (维护者，逗号分割的用户名字符串)
      rawDataAlias:                     (数据别名)
      dataSource:                       (数据源，默认填 svr)
      dataEncoding:                     (数据编码方式，例如 UTF-8)
      sensitivity:                      (数据敏感度，例如 private)
      description:                      (数据描述)
  dataCleanStrategyConfig:              (数据清洗配置)
    bkAppCode:
    bkUsername:
    bkAppSecret:
    bkdataAuthenticationMethod:
    bkBizID:
    rawDataID:                          (数据平台 dataid)
    jSONConfig:                         (清洗规则的json配置，详细请见数据平台文档)
    peConfig:                           (清洗规则的pe配置，详细请见数据平台文档)
    cleanConfigName:                    (数据清洗策略名称)
    resultTableName:                    (清洗导出数据表名称)
    resultTableNameAlias:               (清洗导出数据表别名)
    description:                        (策略描述)
    fields:                             (清洗导出数据表字段描述)
      - fieldName:                      (字段名)
        fieldAlias:                     (字段别名)
        fieldType:                      (字段类型)
        isDimension:                    (是否为维度类型)
        fieldIndex:                     (字段编号，从1开始)
  response:
    errors: ""                          (string errors)
    message: ""                         (return message)
    code: 0                             (return code)
    data: "{\"dataid\":12345}"          (json string)
    result: true                        (true for success)
```

## 接口定义

### Dataid 申请

- 请求地址: /logmanager/v1/dataid
- 请求方式: POST
- 请求数据格式：
```json
{
    "appCode": "",
    "appSecret": "",
    "userName": "",
    "bizID": 0,
    "dataName": "", //对应rawDataName与rawDataAlias
    "maintainers": ""
}
```
- 响应数据格式：
```json
{
    "errCode": 0,
    "errName": "ERROR_OK",
    "message": "",
    "dataID": 0
}
```

### 数据清洗配置

- 请求地址: /logmanager/v1/dataclean
- 请求方式: POST
- 请求数据格式:
```json
{
    // 若设置为true，则会配置默认的容器日志清洗策略，无需填写JSONConfig与fields字段
    "default": false,
    "appCode": "",
    "appSecret": "",
    "userName": "",
    "bizID": 0,
    "dataID": 0,
    "strategyName": "",
    "resultTableName": "",
    "JSONConfig": "",
    "fields": [
        {
            "fieldName": "",
            "fieldAlias": "",
            "fieldType": "",
            "isDimension": false,
            "fieldIndex": 0
        }
    ]
}
```
- 响应数据格式:
```json
{
    "errCode": 0,
    "errName": "ERROR_OK",
    "message": ""
}
```

### 下发日志采集任务

- 请求地址: /logmanager/v1/logcollectiontask
- 请求方式: POST
- 请求数据格式:
```json
{
    // 逗号分割的集群编号
    "clusterIDs": "",
    "config": {
        // 配置名称
        "configName": "",
        // 配置命名空间
        "configNamespace": "",
        "config": {
            "configType": "",
            "appId": "",
            "clusterId": "",
            "stdout": false,
            "stdDataId": "",
            "nonStdDataId": "",
            "logPaths": [],
            "logTags": {},
            "workloadType": "",
            "workloadName": "",
            "workloadNamespace": "",
            "containerConfs": [
                {
                    "containerName": "",
                    "stdout": false,
                    "stdDataId": "",
                    "nonStdDataId": "",
                    "logPaths": [],
                    "logTags": {}
                }
            ],
            "podLabels": false,
            "selector": {
                "matchLabels": {},
                "matchExpressions": [
                    {
                        "key": "",
                        "operator": "",
                        "values": []
                    }
                ]
            }
        }
    }
}
```
- 响应数据格式:
```json
{
    "errCode": 0,
    "errName": "ERROR_OK",
    "message": "",
    // 批量操作时，各失败操作的具体错误信息
    "errResult": [
        {
            "clusterID": "",
            "errCode": 0,
            "errName": "ERROR_OK",
            "message": ""
        }
    ]
}
```

### 删除日志采集任务

- 请求地址: /logmanager/v1/logcollectiontask
- 请求方式: DELETE
- 请求数据格式:
```json
{
    // 逗号分割的集群ID列表，必须指定集群ID
    "clusterIDs": "",
    // 配置名称
    "configName": "",
    // 配置命名空间
    "configNamespace": ""
}
```
- 响应数据格式:
```json
{
    "errCode": 0,
    "errName": "ERROR_OK",
    "message": "",
    // 批量操作时，各失败操作的具体错误信息
    "errResult": [
        {
            "clusterID": "",
            "errCode": 0,
            "errName": "ERROR_OK",
            "message": ""
        }
    ]
}
```

### 获取日志采集任务

- 请求地址: /logmanager/v1/logcollectiontask
- 请求方式: GET
- 请求数据格式:
```json
{
    // 逗号分割的集群ID列表，不指定默认为所有集群
    "clusterIDs": "",
    // 配置名称，不指定默认为所有配置名称
    "configName": "",
    // 配置命名空间，不指定默认为所有命名空间
    "configNamespace": ""
}
```
- 响应数据格式:
```json
{
    "errCode": 0,
    "errName": "ERROR_OK",
    "message": "",
    // 集群日志采集配置列表，每个集群中对应的为列表中的一项
    "data": [
        {
            // 集群ID
            "clusterID": "",
            // 该集群对应请求信息下的日志采集配置列表
            "configs": [
                {
                    "configName": "",
                    "configNamespace": "",
                    "config": {
                        "configType": "",
                        "appId": "",
                        "clusterId": "",
                        "stdout": false,
                        "stdDataId": "",
                        "nonStdDataId": "",
                        "logPaths": [],
                        "logTags": {},
                        "workloadType": "",
                        "workloadName": "",
                        "workloadNamespace": "",
                        "containerConfs": [
                            {
                                "containerName": "",
                                "stdout": false,
                                "stdDataId": "",
                                "nonStdDataId": "",
                                "logPaths": [],
                                "logTags": {}
                            }
                        ],
                        "podLabels": false,
                        "selector": {
                            "matchLabels": {},
                            "matchExpressions": [
                                {
                                    "key": "",
                                    "operator": "",
                                    "values": []
                                }
                            ]
                        }
                    }
                }
            ]
        }
    ]
}
```

## 配置文件

```json
{
    // api-gateway host
    "bcs_api_host": "1.2.3.4:1234",
    // api-gateway auth token
    "api_auth_token": "",
    "use_gateway": true,
    // service cluster kubeconfig
    "kubeconfig": "/path/to/kubeconfig",
    // dataid for system log collection,
    // will apply one if no dataid specified
    "system_dataid": "20770",
    // bkdata api host without http:// and https:// schema
    "bkdata_api_host": "host.without.httpschema",
    "bk_username": "xiaoming",
    "bk_appcode": "xiaoming-test",
    "bk_appsecret": "",
    "bk_bizid": 12345,
    // grpc-gateway & grpc-micro server address
    "address": "1.2.3.4",
    // for example, port is grpc-micro port, port-1 is grpc-gateway port
    "port": "8087",
    // log manager server tls files
    "logmanager_ca_file": "/path/to/ca.crt",
    "logmanager_cert_file": "/path/to/server.crt",
    "logmanager_key_file": "/path/to/server.key",
    // etcd info
    "etcd_hosts": "1.2.3.4:2379",
    "etcd_ca_file": "/path/to/etcd/ca.crt",
    "etcd_cert_file": "/path/to/etcd/server.crt",
    "etcd_key_file": "/path/to/etcd/server.key",
    // log info
    "log_dir": "/data/home/archieai/logtest/log",
    "alsologtostderr": false
}
```