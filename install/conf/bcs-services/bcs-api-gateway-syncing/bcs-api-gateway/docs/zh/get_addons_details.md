### 描述

获取集群组件详情

### 路径参数

| 参数名称        | 参数类型   | 必选 | 描述      |
|-------------|--------|----|---------|
| projectCode | string | 是  | 项目代码    |
| clusterID   | string | 是  | 所在的集群ID |
| name        | string | 是  | 组件名称    |

### 调用示例

```sh
curl -X GET -H 'x-bkapi-authorization: {"access_token": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/helmmanager/v1/projects/projecttest/clusters/clustertest/addons/test_addons
```

### 响应字段说明

data中的字段说明如下：

| 字段名              | 类型               | 描述                                             |
|------------------|------------------|------------------------------------------------|
| name             | 	required string | 	组件名称                                          |
| chartName        | 	required string | 	chart name                                    |
| description      | 	optional string | 	组件描述                                          |
| logo             | 	optional string | 	logo                                          |
| docsLink         | 	optional string | 	文档链接                                          |
| version          | 	required string | 	组件最新版本                                        |
| currentVersion   | 	optional string | 	组件当前安装版本，空代表没安装                               |
| namespace        | 	required string | 	部署的命名空间                                       |
| defaultValues    | 	optional string | 	默认配置，空代表可以直接安装，不需要填写自定义配置                     |
| currentValues    | 	optional string | 	当前部署配置                                        |
| status           | 	optional string | 	部署状态，同 Helm release 状态，空代表没安装                 |
| message          | 	optional string | 	部署信息，部署异常则显示报错信息                              |
| supportedActions | 	repeated string | 	组件支持的操作，目前有 install, upgrade, stop, uninstall |
| releaseName      | 	optional string | 	组件在集群中的 release name                          |


### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "result": true,
  "data": {
    "name": "bk-log-collector",
    "chartName": "bk-log-collector",
    "description": "提供容器日志采集",
    "logo": "",
    "docsLink": "",
    "version": "0.3.5-alpha-x86.12",
    "currentVersion": "0.3.5-alpha-x86.12",
    "namespace": "kube-system",
    "defaultValues": "bkunifylogbeat:\n  ipcPath: /var/run/ipc.state.report",
    "currentValues": "",
    "status": "deployed",
    "message": "",
    "supportedActions": [
      "install",
      "upgrade",
      "uninstall"
    ],
    "releaseName": "bk-log-collector"
  },
  "requestID": "d7aaba73b9b567ef2afe819061532c39",
  "web_annotations": null
}
```