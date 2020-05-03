# bcs-storage API 文档 V0.9.0

## 请求不同对象的URL Prefix

### bcs-api

##### 前缀

http://bcs_api_address:port/bcsapi/v4/storage

##### 示例

http://bcs_api_address:port/bcsapi/v4/storage/query/dynamic/clusters/BCS-TEST-10000/taskgroup


## Attention

- 与上一个版本接口有细节方面的不同，以此版本为准

- 部分接口分为mesos/k8s，具体体现在url的前缀中。文档中默认为mesos，若是k8s的接口则将url中mesos替换为k8s


## change log

### V0.9.0

1. 增加query类mesos接口：namespace

### V0.8.0

1. 增加query类k8s接口：daemonset, job, statefulset

### V0.7.0

1. query接口增加extra参数

### V0.6.0

1. 增加host接口

### V0.5.0

1. 增加metric接口

### V0.4.0

1. event结构增加ExtraInfo字段，存储额外信息，同时增加为查询接口的过滤字段
2. 增加alarm接口

### V0.3.1

1. 增加针对所有动态数据操作的接口

### V0.3.0

1. 所有查询类接口的时间参数格式均改为unix时间戳，为int64类型
2. 所有动态数据的query接口field参数，是否支持逗号分隔，改为“是”



### V0.2.0

1. 添加```delete batch namespace resource```和```delete batch cluster resource```两个方法，支持时间过滤删除动态数据，用于watch端处理脏数据

### V0.1.1

1. ```query-mesos ```和```query-k8s```两类接口统一添加```field```参数，用于选择返回的字段
2. watch类接口增加```/mesos```前缀
3. 所有的带```field```参数的接口支持多个参数用逗号```,```分隔

### V0.1.0

1. 第一版文档




## 动态数据

### namespace类

##### get namespace resource

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| ```/mesos/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName}``` |
| METHOD                                   |
| GET                                      |



成功返回示例

```
{
  "result": true,
  "code": 0,
  "message": "Success",
  "data": [
     {
        "_id": xxxxx,
        "clusterId": xxxxx,
        "namespace": xxxxx,
        "resourceName": xxxxx,
        "resourceType": xxxxx,
        "data" : //上报的数据,
        "createTime": "2017-09-26 14:00:00",
        "updateTime": "2017-09-26 14:00:00"
     }
  ]
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10004,
  “message”: “Get resource failed.”,
  “data”: []
}
```



##### put namespace resource

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| ```/mesos/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName}             ``` |
| METHOD                                   |
| PUT                                      |



请求参数:

- 接口见bcs-common/type/storage.go中的BcsStorageDynamicIf


- 将BcsStorageDynamicIf序列化放入body





成功返回示例

```
{
  "result": true,
  "code": 0,
  "message": "Success",
  "data": null
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10003,
  “message”: “Put resource failed.”,
  “data”: null
}
```



##### delete namespace resource

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /mesos/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName} |
| METHOD                                   |
| DELETE                                   |



成功返回示例

```
{
  "result": true,
  "code": 0,
  "message": "Success",
  "data": null
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10005,
  “message”: “Delete resource failed.”,
  “data”: null
}
```



##### list namespace resource

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /mesos/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/{resourceType} |
| METHOD                                   |
| GET                                      |



请求参数

| 参数    | 说明                                       | 必须   | 类型     |
| ----- | ---------------------------------------- | ---- | ------ |
| field | 指定data中单个资源返回的字段，可以逗号分隔多个                | 否    | string |
| extra | 额外条件json的base64编码，其中额外条件层次结构用"."连接，若key中本身含有“.”，则使用其unicode代替(\uff0e) | 否    | string |



成功返回示例

```
{
  “result”: true,
  “code”: 0,
  “message”: “”,
  “data”: [
  {
      “_id”: xxxxx,
      “clusterId”: xxxxx,
      “namespace”: xxxxx,
      “resourceName”: xxxxx,
      “resourceType”: xxxxx,
      “data” : {
          “clusterId”: 1,
      },
      “createTime”: “2017-09-26 16:59:39”,
  	  “updateTime”: “2017-09-26 16:59:39”
  }
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### delete batch namespace resource

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /mesos/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/{resourceType} |
| METHOD                                   |
| DELETE                                   |



请求参数

| 参数              | 说明                                       | 必须   | 类型     |
| --------------- | ---------------------------------------- | ---- | ------ |
| extra           | 额外条件json的base64编码，其中额外条件层次结构用"."连接，若key中本身含有“.”，则使用其unicode代替(\uff0e) | 否    | string |
| updateTimeBegin | updateTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  |
| updateTimeEnd   | updateTime的区间右边界                         | 否    | int64  |



成功返回示例

```
{
  “result”: true,
  “code”: 0,
  “message”: “Success”,
  “data”: nil
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10005,
  “message”: “Delete resource failed.”,
  “data”: []
}
```



### cluster类

##### get cluster resource

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /mesos/dynamic/cluster_resources/clusters/{clusterId}/{resourceType}/{resourceName} |
| METHOD                                   |
| GET                                      |





成功返回示例

```
{
  “result”: true,
  “code”: 0,
  “message”: “Success”,
  “data”: [
     {
        “_id”: xxxxx,
        “clusterId”: xxxxx,
        “resourceName”: xxxxx,
        “resourceType”: xxxxx,
        “data” : //上报的数据,
        “createTime": “2017-09-26 16:59:39”,
        “updateTime”: “2017-09-26 16:59:39”
     }
  ]
}

```



失败返回示例

```
{
  “result”: false,
  “code”: 10004,
  “message”: “Get resource failed.”,
  “data”: []
}
```



##### put cluster resource

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /mesos/dynamic/cluster_resources/clusters/{clusterId}/{resourceType}/{resourceName} |
| METHOD                                   |
| PUT                                      |



请求参数:

- 接口见bcs-common/type/storage.go中的BcsStorageDynamicIf


- 将BcsStorageDynamicIf序列化放入body





成功返回示例

```
{
  "result": true,
  "code": 0,
  "message": "Success",
  "data": null
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10003,
  “message”: “Put resource failed.”,
  “data”: null
}
```



##### delete cluster resource

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /mesos/dynamic/cluster_resources/clusters/{clusterId}/{resourceType}/{resourceName} |
| METHOD                                   |
| DELETE                                   |



成功返回示例

```
{
  "result": true,
  "code": 0,
  "message": "Success",
  "data": null
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10005,
  “message”: “Delete resource failed.”,
  “data”: null
}
```



##### list cluster resource



| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /mesos/dynamic/cluster_resources/clusters/{clusterId}/{resourceType} |
| METHOD                                   |
| GET                                      |



请求参数

| 参数    | 说明                                       | 必须   | 类型     |
| ----- | ---------------------------------------- | ---- | ------ |
| field | 指定data中单个资源返回的字段，可以逗号分隔多个                | 否    | string |
| extra | 额外条件json的base64编码，其中额外条件层次结构用"."连接，若key中本身含有“.”，则使用其unicode代替(\uff0e) | 否    | string |



成功返回示例

```
{
  “result”: true,
  “code”: 0,
  “message”: “”,
  “data”: [
  {
      “_id”: xxxxx,
      “clusterId”: xxxxx,
      “resourceName”: xxxxx,
      “resourceType”: xxxxx,
      “data” : {
          “clusterId”: 1,
      },
      “createTime”: “2017-09-26 16:59:39”,
  	  “updateTime”: “2017-09-26 16:59:39”
  }
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### delete batch cluster resource

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /mesos/dynamic/cluster_resources/clusters/{clusterId}/{resourceType} |
| METHOD                                   |
| DELETE                                   |



请求参数

| 参数              | 说明                                       | 必须   | 类型     |
| --------------- | ---------------------------------------- | ---- | ------ |
| extra           | 额外条件json的base64编码，其中额外条件层次结构用"."连接，若key中本身含有“.”，则使用其unicode代替(\uff0e) | 否    | string |
| updateTimeBegin | updateTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  |
| updateTimeEnd   | updateTime的区间右边界                         | 否    | int64  |



成功返回示例

```
{
  “result”: true,
  “code”: 0,
  “message”: “Success”,
  “data”: nil
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10005,
  “message”: “Delete resource failed.”,
  “data”: []
}
```



### all(针对所有动态数据操作)

##### list resource



| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /mesos/dynamic/all_resources/clusters/{clusterId}/{resourceType} |
| METHOD                                   |
| GET                                      |



请求参数

| 参数    | 说明                                       | 必须   | 类型     |
| ----- | ---------------------------------------- | ---- | ------ |
| field | 指定data中单个资源返回的字段，可以逗号分隔多个                | 否    | string |
| extra | 额外条件json的base64编码，其中额外条件层次结构用"."连接，若key中本身含有“.”，则使用其unicode代替(\uff0e) | 否    | string |



成功返回示例

```
{
  “result”: true,
  “code”: 0,
  “message”: “”,
  “data”: [
  {
      “_id”: xxxxx,
      “clusterId”: xxxxx,
      “resourceName”: xxxxx,
      “resourceType”: xxxxx,
      “data” : {
          “clusterId”: 1,
      },
      “createTime”: “2017-09-26 16:59:39”,
  	  “updateTime”: “2017-09-26 16:59:39”
  }
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### delete batch resource

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /mesos/dynamic/all_resources/clusters/{clusterId}/{resourceType} |
| METHOD                                   |
| DELETE                                   |



请求参数

| 参数              | 说明                                       | 必须   | 类型     |
| --------------- | ---------------------------------------- | ---- | ------ |
| extra           | 额外条件json的base64编码，其中额外条件层次结构用"."连接，若key中本身含有“.”，则使用其unicode代替(\uff0e) | 否    | string |
| updateTimeBegin | updateTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  |
| updateTimeEnd   | updateTime的区间右边界                         | 否    | int64  |



成功返回示例

```
{
  “result”: true,
  “code”: 0,
  “message”: “Success”,
  “data”: nil
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10005,
  “message”: “Delete resource failed.”,
  “data”: []
}
```



### query-mesos

| 参数    | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| ----- | ---------------------------------------- | ---- | ------ | -------------- |
| extra | 额外条件json的base64编码，其中额外条件层次结构用"."连接，若key中本身含有“.”，则使用其unicode代替(\uff0e) | 否    | string | 否              |



##### 1. taskgroup

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/mesos/dynamic/clusters/{clusterId}/taskgroup |
| METHOD                                   |
| GET                                      |



| 参数                  | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| ------------------- | ---------------------------------------- | ---- | ------ | -------------- |
| name                | 名字                                       | 否    | string | 是              |
| namespace           | namespace                                | 否    | string | 是              |
| rcName              | 对应application的name                       | 否    | string | 是              |
| status              | 当前状态                                     | 否    | string | 是              |
| lastStatus          | 上一次状态                                    | 否    | string | 是              |
| hostIp              | 所处节点的ip                                  | 否    | string | 是              |
| hostName            | 所处节点的名字                                  | 否    | string | 是              |
| podIp               | taskgroup的ip                             | 否    | string | 是              |
| startTimeBegin      | startTime的区间左边界 unix时间戳 如1509423130      | 否    | int64  | 否              |
| startTimeEnd        | startTime的区间右边界                          | 否    | int64  | 否              |
| lastUpdateTimeBegin | lastUpdateTime的区间左边界                     | 否    | int64  | 否              |
| lastUpdateTimeEnd   | lastUpdateTime的区间右边界                     | 否    | int64  | 否              |
| field               | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/mesos/dynamic/clusters/BCS-100001/taskgroup?namespace=defaultgroup&hostIp=127.0.0.3,127.0.0.4
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "BCS-100001",
      "createTime": "2017-10-11 09:55:24",
      "data": {
        "containerStatuses": [
          {
            "containerID": "",
            "exitcode": 0,
            "finishTime": "0001-01-01T00:00:00Z",
            "healCheckStatus": [
              {
                "message": "http check by executor",
                "result": false,
                "type": "HTTP"
              },
              {
                "message": "check endpoint 127.0.0.3:31000 resp httpcode 404",
                "result": false,
                "type": "REMOTE_HTTP"
              },
              {
                "message": "check endpoint 127.0.0.3:31000 tcp ok",
                "result": true,
                "type": "REMOTE_TCP"
              }
            ],
            "image": "image.com/myimage:latest",
            "lastStatus": "Starting",
            "lastUpdateTime": "2017-10-09T15:38:07+08:00",
            "message": "container is running, but unhealthy",
            "name": "",
            "restartCount": 0,
            "startTime": "2017-10-09T07:37:45.690518395Z",
            "status": "Running"
          }
        ],
        "hostIP": "127.0.0.3",
        "hostName": "mesos-slave",
        "lastStatus": "Starting",
        "lastUpdateTime": "2017-10-09T15:37:46+08:00",
        "message": "pod is running",
        "metadata": {
          "name": "",
          "namespace": "defaultgroup"
        },
        "podIP": "127.0.0.5",
        "rcname": "app-test",
        "reportTime": "2017-10-11T09:51:16.786921704+08:00",
        "startTime": "2017-10-09T15:37:44+08:00",
        "status": "Running"
      },
      "namespace": "defaultgroup",
      "resourceName": "",
      "resourceType": "taskgroup",
      "updateTime": "2017-10-11 09:55:24"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 2. application

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/mesos/dynamic/clusters/{clusterId}/application |
| METHOD                                   |
| GET                                      |



| 参数                  | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| ------------------- | ---------------------------------------- | ---- | ------ | -------------- |
| name                | 名字                                       | 否    | string | 是              |
| namespace           | namespace                                | 否    | string | 是              |
| instance            | 期望实例数                                    | 否    | int    | 否              |
| buildedInstance     | builded实例数                               | 否    | int    | 否              |
| runningInstance     | 正在运行的实例数                                 | 否    | int    | 否              |
| status              | 当前状态                                     | 否    | string | 是              |
| lastStatus          | 上一次状态                                    | 否    | string | 是              |
| podIp               | taskgroup的ip                             | 否    | string | 是              |
| createTimeBegin     | createTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  | 否              |
| createTimeEnd       | createTime的区间右边界                         | 否    | int64  | 否              |
| lastUpdateTimeBegin | lastUpdateTime的区间左边界                     | 否    | int64  | 否              |
| lastUpdateTimeEnd   | lastUpdateTime的区间右边界                     | 否    | int64  | 否              |
| field               | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/mesos/dynamic/clusters/BCS-TEST-10000/application?createTimeBegin=2017-09-13%2019:40:30&createTimeEnd=2017-09-13%2019:40:40
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "BCS-TEST-10000",
      "createTime": "2017-10-11 10:19:02",
      "data": {
        "buildedInstance": 1,
        "createTime": "2017-09-13T19:40:36+08:00",
        "instance": 1,
        "lastStatus": "Deploying",
        "lastUpdateTime": "2017-09-13T19:40:41+08:00",
        "message": "application is running",
        "metadata": {
          "labels": {

            "io.tencent.bcs.cluster": "BCS-TEST-10000",
            "secret-app": "test-developer-container"
          },
          "name": "test-developer-container",
          "namespace": "developer"
        },
        "pods": [
          {
            "name": ""
          }
        ],
        "reportTime": "2017-10-11T10:17:50.985602723+08:00",
        "runningInstance": 1,
        "status": "Running"
      },
      "namespace": "developer",
      "resourceName": "",
      "resourceType": "application",
      "updateTime": "2017-10-11 10:19:02"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 3. deployment

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/mesos/dynamic/clusters/{clusterId}/deployment |
| METHOD                                   |
| GET                                      |



| 参数                   | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| -------------------- | ---------------------------------------- | ---- | ------ | -------------- |
| name                 | 名字                                       | 否    | string | 是              |
| namespace            | namespace                                | 否    | string | 是              |
| checkTime            | checkTime                                | 否    | int64  | 否              |
| status               | 当前状态                                     | 否    | string | 是              |
| applicationName      | application名字                            | 否    | string | 是              |
| applicationExtName   | applicationExt名字                         | 否    | string | 是              |
| currRollingOperation | 当前roll update的操作                         | 否    | string | 是              |
| isInRolling          | 是否在roll update中(true/false)              | 否    | bool   | 否              |
| lastRollingTimeBegin | lastRollingTime的区间左边界 unix时间戳 如1509423130 | 否    | int64  | 否              |
| lastRollingTimeEnd   | lastRollingTime的区间右边界                    | 否    | int64  | 否              |
| field                | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/mesos/dynamic/clusters/BCS-TEST-10000/deployment?isInRolling=false&status=Running,Staging
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "BCS-TEST-10000",
      "createTime": "2017-10-11 10:27:43",
      "data": {
        "application": {
          "curr_target_instances": 0,
          "name": "test-deployment"
        },
        "application_ext": null,
        "check_time": 0,
        "curr_rolling_operation": "",
        "is_in_rolling": false,
        "last_rolling_time": 0,
        "metadata": {
          "labels": {

            "io.tencent.bcs.cluster": "BCS-TEST-10000",
            "label_deployment": "label_deployment"
          },
          "name": "test-deployment",
          "namespace": "defaultGroup"
        },
        "selector": {
          "podname": "test-deployment"
        },
        "status": "Running",
        "strategy": {
          "rollingupdate": {
            "RollingManully": false,
            "maxSurge": 1,
            "maxUnavilable": 1,
            "rollingOrder": "CreateFirst",
            "upgradeDuration": 60
          },
          "type": "RollingUpdate"
        }
      },
      "namespace": "defaultGroup",
      "resourceName": "test-deployment",
      "resourceType": "deployment",
      "updateTime": "2017-10-11 10:27:43"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 4. service

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/mesos/dynamic/clusters/{clusterId}/service |
| METHOD                                   |
| GET                                      |



| 参数         | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| ---------- | ---------------------------------------- | ---- | ------ | -------------- |
| name       | 名字                                       | 否    | string | 是              |
| namespace  | namespace                                | 否    | string | 是              |
| apiVersion | apiVersion                               | 否    | string | 是              |
| field      | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/mesos/dynamic/clusters/BCS-TEST-10000/service?apiVersion=v1
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "BCS-TEST-10000",
      "createTime": "2017-10-11 10:32:13",
      "data": {
        "apiVersion": "v1",
        "kind": "service",
        "metadata": {
          "labels": {
            "BCSGROUP": "external",
            "io.tencent.bcs.cluster": "BCS-TEST-10000"
          },
          "name": "service-test",
          "namespace": "defaultGroup"
        },
        "spec": {
          "clusterIP": null,
          "ports": [
            {
              "name": "test-port",
              "nodePort": 0,
              "protocol": "http",
              "servicePort": 18800
            }
          ],
          "selector": {
            "podname": ""
          }
        }
      },
      "namespace": "defaultGroup",
      "resourceName": "service-test",
      "resourceType": "service",
      "updateTime": "2017-10-11 10:32:13"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 5. configmap

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/mesos/dynamic/clusters/{clusterId}/configmap |
| METHOD                                   |
| GET                                      |



| 参数         | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| ---------- | ---------------------------------------- | ---- | ------ | -------------- |
| name       | 名字                                       | 否    | string | 是              |
| namespace  | namespace                                | 否    | string | 是              |
| apiVersion | apiVersion                               | 否    | string | 是              |
| field      | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/mesos/dynamic/clusters/BCS-TEST-10000/configmap?apiVersion=v1
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "59dd834798ebbfd671e9a8fd",
      "clusterId": "BCS-TEST-10000",
      "createTime": "2017-10-11 10:34:47",
      "data": {
        "apiVersion": "v1",
        "datas": {
          "config-one": {
            "RemoteUser": "",
            "content": "Y29uZmlnIGNvbnRleHQ=",
            "type": "file"
          },
          "config-two": {
            "RemoteUser": "",
            "content": "Y29uZmlnIGNvbnRleHQ=",
            "type": "file"
          }
        },
        "kind": "configmap",
        "metadata": {
          "labels": {
            "io.tencent.bcs.cluster": "BCS-TEST-10000"
          },
          "name": "test-configmap",
          "namespace": "defaultGroup"
        }
      },
      "namespace": "defaultGroup",
      "resourceName": "test-configmap",
      "resourceType": "configmap",
      "updateTime": "2017-10-11 10:34:47"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 6. secret

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/mesos/dynamic/clusters/{clusterId}/secret |
| METHOD                                   |
| GET                                      |



| 参数         | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| ---------- | ---------------------------------------- | ---- | ------ | -------------- |
| name       | 名字                                       | 否    | string | 是              |
| namespace  | namespace                                | 否    | string | 是              |
| apiVersion | apiVersion                               | 否    | string | 是              |
| field      | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/mesos/dynamic/clusters/BCS-TEST-10000/secret?apiVersion=v1
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "BCS-TEST-10000",
      "createTime": "2017-10-11 10:36:23",
      "data": {
        "apiVersion": "v1",
        "datas": {
          "first-secret": {
            "content": "Y29uZmlnIGNvbnRleHQ=",
            "path": "/path/to/store/in/vault"
          },
          "second-secret": {
            "content": "Y29uZmlnIGNvbnRleHQ=",
            "path": "/path/to/store/in/vault"
          }
        },
        "kind": "secret",
        "metadata": {
          "labels": {
            "io.tencent.bcs.cluster": "BCS-TEST-10000"
          },
          "name": "template-secret",
          "namespace": "defaultGroup"
        }
      },
      "namespace": "defaultGroup",
      "resourceName": "template-secret",
      "resourceType": "secret",
      "updateTime": "2017-10-11 10:36:23"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 7. namespace

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/mesos/dynamic/clusters/{clusterId}/namespace |
| METHOD                                   |
| GET                                      |



请求示例

```
/query/mesos/dynamic/clusters/BCS-TEST-10000/namespace
```



成功返回示例

```
{
  "code": 0,
  "data": [
    "test",
    "test1",
    "test2",
    "test3"
  ],
  "message": "success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}

```

#### 8. IP pool statistic

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/mesos/dynamic/clusters/{clusterId}/ippoolstatic |
| METHOD                                   |
| GET                                      |



请求示例

```
/query/mesos/dynamic/clusters/BCS-TEST-10000/ippoolstatic
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "resourceName": "IPPoolStatic-BCS-TEST-10000",
      "resourceType": "IPPoolStatic",
      "data": {
          "reservedip": 0,
          "activeip": 0,
          "availableip": 8,
          "poolnum": 1
      },
      "createTime": "2020-02-20T11:04:44.162+08:00",
      "updateTime": "2020-03-11T14:32:18.912+08:00",
      "_id": "5e4df74cfcc088392c991783",
      "clusterId": "BCS-TEST-1000"
    }
  ],
  "message": "success",
  "result": true
}
```



失败返回示例

```
{
  "code": 10086,
  "data": {},
  "message": "some failed reason",
  "result": false
}
```

### query-k8s

| 参数    | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| ----- | ---------------------------------------- | ---- | ------ | -------------- |
| extra | 额外条件json的base64编码，其中额外条件层次结构用"."连接，若key中本身含有“.”，则使用其unicode代替(\uff0e) | 否    | string | 否              |



##### 1. pod

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/k8s/dynamic/clusters/{clusterId}/pod |
| METHOD                                   |
| GET                                      |



| 参数              | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| --------------- | ---------------------------------------- | ---- | ------ | -------------- |
| name            | 名字                                       | 否    | string | 是              |
| namespace       | namespace                                | 否    | string | 是              |
| status          | 当前状态                                     | 否    | string | 是              |
| hostIp          | 所处节点的ip                                  | 否    | string | 是              |
| podIp           | pod的ip                                   | 否    | string | 是              |
| startTimeBegin  | startTime的区间左边界 unix时间戳 如1509423130      | 否    | int64  | 否              |
| startTimeEnd    | startTime的区间右边界                          | 否    | int64  | 否              |
| createTimeBegin | createTime的区间左边界                         | 否    | int64  | 否              |
| createTimeEnd   | createTime的区间右边界                         | 否    | int64  | 否              |
| field           | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/k8s/dynamic/clusters/12121/pod?createTimeBegin=2017-10-16%2007:34:50&createTimeEnd=2017-10-16%2007:55:54
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "BCS-TEST-10000",
      "createTime": "2017-10-16 16:15:19",
      "data": {
        "metadata": {
          "annotations": {
            "kubernetes.io/created-by": ""
          },
          "creationTimestamp": "2017-10-16T07:34:52Z",
          "generateName": "",
          "labels": {
            "app": "",
            "pod-template-hash": "",
            "version": ""
          },
          "name": "",
          "namespace": "test",
          "ownerReferences": [
            {
              "apiVersion": "",
              "controller": true,
              "kind": "ReplicaSet",
              "name": "",
              "uid": ""
            }
          ],
          "resourceVersion": "",
          "selfLink": "",
          "uid": ""
        },
        "spec": {
          "containers": [
            {
              "image": "",
              "imagePullPolicy": "Always",
              "name": "",
              "ports": [
                {
                  "containerPort": 80,
                  "protocol": "TCP"
                }
              ],
              "resources": [],
              "terminationMessagePath": ""
            }
          ],
          "dnsPolicy": "ClusterFirst",
          "nodeName": "127.0.0.1",
          "restartPolicy": "Always",
          "securityContext": [],
          "terminationGracePeriodSeconds": 30
        },
        "status": {
          "conditions": [
            {
              "lastProbeTime": null,
              "lastTransitionTime": "2017-10-16T07:34:52Z",
              "status": "True",
              "type": "Initialized"
            },
            {
              "lastProbeTime": null,
              "lastTransitionTime": "2017-10-16T07:34:54Z",
              "status": "True",
              "type": "Ready"
            },
            {
              "lastProbeTime": null,
              "lastTransitionTime": "2017-10-16T07:34:52Z",
              "status": "True",
              "type": "PodScheduled"
            }
          ],
          "containerStatuses": [
            {
              "containerID": "",
              "image": "",
              "imageID": "",
              "lastState": [],
              "name": "",
              "ready": true,
              "restartCount": 0,
              "state": {
                "running": {
                  "startedAt": "2017-10-16T07:34:54Z"
                }
              }
            }
          ],
          "hostIP": "127.0.0.2",
          "phase": "Running",
          "podIP": "",
          "startTime": "2017-10-16T07:34:52Z"
        }
      },
      "namespace": "default",
      "resourceName": "",
      "resourceType": "Pod",
      "updateTime": "2017-10-16 16:15:19"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 2. replicaset

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/k8s/dynamic/clusters/{clusterId}/replicaset |
| METHOD                                   |
| GET                                      |



| 参数                | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| :---------------- | ---------------------------------------- | ---- | ------ | -------------- |
| name              | 名字                                       | 否    | string | 是              |
| namespace         | namespace                                | 否    | string | 是              |
| replicas          | 期望的replicas数                             | 否    | int    | 否              |
| availableReplicas | 可用的replicas数                             | 否    | int    | 否              |
| readyReplicas     | 就绪的replicas数                             | 否    | int    | 否              |
| createTimeBegin   | createTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  | 否              |
| createTimeEnd     | createTime的区间右边界                         | 否    | int64  | 否              |
| field             | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/k8s/dynamic/clusters/BCS-TEST-10000/replicaset
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "BCS-TEST-10000",
      "createTime": "2017-10-16 16:56:43",
      "data": {
        "metadata": {
          "annotations": {
            "deployment.kubernetes.io/desired-replicas": "2",
            "deployment.kubernetes.io/max-replicas": "3",
            "deployment.kubernetes.io/revision": ""
          },
          "creationTimestamp": "2017-10-16T11:55:39Z",
          "generation": 2,
          "labels": {
            "app": "",
            "pod-template-hash": "",
            "version": ""
          },
          "name": "",
          "namespace": "test",
          "resourceVersion": "",
          "selfLink": "",
          "uid": ""
        },
        "spec": {
          "replicas": 2,
          "selector": {
            "matchLabels": {
              "app": "",
              "pod-template-hash": ""
            }
          },
          "template": {
            "metadata": {
              "creationTimestamp": null,
              "labels": {
                "app": "",
                "pod-template-hash": "",
                "version": ""
              }
            },
            "spec": {
              "containers": [
                {
                  "image": "",
                  "imagePullPolicy": "Always",
                  "name": "",
                  "ports": [
                    {
                      "containerPort": 80,
                      "protocol": "TCP"
                    }
                  ],
                  "resources": [],
                  "terminationMessagePath": ""
                }
              ],
              "dnsPolicy": "ClusterFirst",
              "restartPolicy": "Always",
              "securityContext": [],
              "terminationGracePeriodSeconds": 30
            }
          }
        },
        "status": {
          "availableReplicas": 2,
          "fullyLabeledReplicas": 2,
          "observedGeneration": 2,
          "readyReplicas": 2,
          "replicas": 2
        }
      },
      "namespace": "default",
      "resourceName": "",
      "resourceType": "ReplicaSet",
      "updateTime": "2017-10-16 20:13:41"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 3. deployment

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/k8s/dynamic/clusters/{clusterId}/deployment |
| METHOD                                   |
| GET                                      |



| 参数                | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| ----------------- | ---------------------------------------- | ---- | ------ | -------------- |
| name              | 名字                                       | 否    | string | 是              |
| namespace         | namespace                                | 否    | string | 是              |
| replicas          | 期望的replicas数                             | 否    | int    | 否              |
| availableReplicas | 可用的replicas数                             | 否    | int    | 否              |
| updatedReplicas   | 更新的replicas数                             | 否    | int    | 否              |
| strategyType      | update的策略                                | 否    | string | 是              |
| dnsPolicy         | dns策略                                    | 否    | string | 是              |
| restartPolicy     | 重启策略                                     | 否    | bool   | 否              |
| createTimeBegin   | createTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  | 否              |
| createTimeEnd     | createTime的区间右边界                         | 否    | int64  | 否              |
| field             | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/k8s/dynamic/clusters/BCS-TEST-10000/deployment
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "BCS-TEST-10000",
      "createTime": "2017-10-16 20:00:58",
      "data": {
        "metadata": {
          "annotations": {
            "deployment.kubernetes.io/revision": "13"
          },
          "creationTimestamp": "2017-08-01T08:55:09Z",
          "generation": 16,
          "labels": {
            "pod_selector": "",
            "release_version": ""
          },
          "name": "",
          "namespace": "test",
          "resourceVersion": "",
          "selfLink": "",
          "uid": ""
        },
        "spec": {
          "minReadySeconds": 1,
          "replicas": 1,
          "selector": {
            "matchLabels": {
              "pod_selector": ""
            }
          },
          "strategy": {
            "rollingUpdate": {
              "maxSurge": 1,
              "maxUnavailable": 1
            },
            "type": "RollingUpdate"
          },
          "template": {
            "metadata": {
              "creationTimestamp": null,
              "labels": {
                "pod_selector": "",
                "release_version": "16"
              },
              "name": ""
            },
            "spec": {
              "containers": [
                {
                  "args": [
                    "start",
                    "logging"
                  ],
                  "command": [
                    "bash",
                    "/runner/init"
                  ],
                  "env": [
                    {
                      "name": "MYSQL_PASSWORD",
                      "value": ""
                    },
                    {
                      "name": "SLUG_URL",
                      "value": ""
                    }
                  ],
                  "image": "",
                  "imagePullPolicy": "Never",
                  "name": "",
                  "ports": [
                    {
                      "containerPort": 5000,
                      "protocol": "TCP"
                    }
                  ],
                  "resources": {
                    "limits": {
                      "cpu": "250m",
                      "memory": "512Mi"
                    },
                    "requests": {
                      "cpu": "100m",
                      "memory": "64Mi"
                    }
                  },
                  "terminationMessagePath": ""
                }
              ],
              "dnsPolicy": "ClusterFirst",
              "restartPolicy": "Always",
              "securityContext": [],
              "terminationGracePeriodSeconds": 30
            }
          }
        },
        "status": {
          "availableReplicas": 1,
          "conditions": [
            {
              "lastTransitionTime": "2017-08-01T08:55:10Z",
              "lastUpdateTime": "2017-08-01T08:55:10Z",
              "message": "Deployment has minimum availability.",
              "reason": "MinimumReplicasAvailable",
              "status": "True",
              "type": "Available"
            }
          ],
          "observedGeneration": 16,
          "replicas": 1,
          "updatedReplicas": 1
        }
      },
      "namespace": "default",
      "resourceName": "",
      "resourceType": "Deployment",
      "updateTime": "2017-10-16 20:00:58"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 4. service

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/k8s/dynamic/clusters/{clusterId}/service |
| METHOD                                   |
| GET                                      |



| 参数              | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| --------------- | ---------------------------------------- | ---- | ------ | -------------- |
| name            | 名字                                       | 否    | string | 是              |
| namespace       | namespace                                | 否    | string | 是              |
| clusterIp       | clusterIp                                | 否    | string | 是              |
| type            | type                                     | 否    | string | 是              |
| createTimeBegin | createTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  | 否              |
| createTimeEnd   | createTime的区间右边界                         | 否    | int64  | 否              |
| field           | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/k8s/dynamic/clusters/BCS-TEST-10000/service
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "12121",
      "createTime": "2017-10-16 20:07:26",
      "data": {
        "metadata": {
          "creationTimestamp": "2017-08-02T06:33:43Z",
          "name": "",
          "namespace": "test",
          "resourceVersion": "",
          "selfLink": "",
          "uid": ""
        },
        "spec": {
          "clusterIP": "127.0.0.1",
          "ports": [
            {
              "name": "http",
              "port": 80,
              "protocol": "TCP",
              "targetPort": 5000
            }
          ],
          "selector": {
            "pod_selector": ""
          },
          "sessionAffinity": "None",
          "type": "ClusterIP"
        },
        "status": {
          "loadBalancer": []
        }
      },
      "namespace": "default",
      "resourceName": "",
      "resourceType": "Service",
      "updateTime": "2017-10-16 20:07:26"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 5. configmap

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/k8s/dynamic/clusters/{clusterId}/configmap |
| METHOD                                   |
| GET                                      |



| 参数              | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| --------------- | ---------------------------------------- | ---- | ------ | -------------- |
| name            | 名字                                       | 否    | string | 是              |
| namespace       | namespace                                | 否    | string | 是              |
| createTimeBegin | createTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  | 否              |
| createTimeEnd   | createTime的区间右边界                         | 否    | int64  | 否              |
| field           | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/k8s/dynamic/clusters/12121/configmap
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "BCS-TEST-10000",
      "createTime": "2017-10-16 20:10:01",
      "data": {
        "metadata": {
          "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"v1\",\"kind\":\"ConfigMap\",\"metadata\":{\"annotations\":{},\"name\":\"\",\"namespace\":\"default\"}}\n"
          },
          "creationTimestamp": "2017-05-08T12:34:16Z",
          "name": "",
          "namespace": "default",
          "resourceVersion": "",
          "selfLink": "",
          "uid": ""
        }
      },
      "namespace": "default",
      "resourceName": "",
      "resourceType": "ConfigMap",
      "updateTime": "2017-10-16 20:10:01"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 6. secret

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/k8s/dynamic/clusters/{clusterId}/secret |
| METHOD                                   |
| GET                                      |



| 参数              | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| --------------- | ---------------------------------------- | ---- | ------ | -------------- |
| name            | 名字                                       | 否    | string | 是              |
| namespace       | namespace                                | 否    | string | 是              |
| createTimeBegin | createTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  | 否              |
| createTimeEnd   | createTime的区间右边界                         | 否    | int64  | 否              |
| field           | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/k8s/dynamic/clusters/BCS-TEST-10000/secret
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "BCS-TEST-10000",
      "createTime": "2017-10-12 11:14:21",
      "data": {
        "data": {
          "ca.crt": "xxx",
          "namespace": "ZGVmYXVsdA==",
          "token": "xxx"
        },
        "metadata": {
          "annotations": {
            "kubernetes.io/service-account.name": "",
            "kubernetes.io/service-account.uid": ""
          },
          "creationTimestamp": "2017-05-11T10:14:40Z",
          "name": "",
          "namespace": "default",
          "resourceVersion": "",
          "selfLink": "",
          "uid": ""
        },
        "type": "kubernetes.io/service-account-token"
      },
      "namespace": "default",
      "resourceName": "",
      "resourceType": "Secret",
      "updateTime": "2017-10-16 20:11:56"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 7. endpoints

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/k8s/dynamic/clusters/{clusterId}/endpoints |
| METHOD                                   |
| GET                                      |



| 参数              | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| --------------- | ---------------------------------------- | ---- | ------ | -------------- |
| name            | 名字                                       | 否    | string | 是              |
| namespace       | namespace                                | 否    | string | 是              |
| createTimeBegin | createTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  | 否              |
| createTimeEnd   | createTime的区间右边界                         | 否    | int64  | 否              |
| field           | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/k8s/dynamic/clusters/BCS-TEST-10000/endpoints
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "BCS-TEST-10000",
      "createTime": "2017-10-16 20:16:16",
      "data": {
        "metadata": {
          "creationTimestamp": "2017-08-15T03:41:05Z",
          "name": "",
          "namespace": "",
          "resourceVersion": "",
          "selfLink": "",
          "uid": ""
        },
        "subsets": [
          {
            "addresses": [
              {
                "ip": "127.0.0.1",
                "nodeName": "127.0.0.2",
                "targetRef": {
                  "kind": "Pod",
                  "name": "",
                  "namespace": "test",
                  "resourceVersion": "",
                  "uid": ""
                }
              }
            ],
            "ports": [
              {
                "name": "http",
                "port": 5000,
                "protocol": "TCP"
              }
            ]
          }
        ]
      },
      "namespace": "default",
      "resourceName": "",
      "resourceType": "EndPoints",
      "updateTime": "2017-10-16 20:16:16"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 8. ingress

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/k8s/dynamic/clusters/{clusterId}/ingress |
| METHOD                                   |
| GET                                      |



| 参数              | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| --------------- | ---------------------------------------- | ---- | ------ | -------------- |
| name            | 名字                                       | 否    | string | 是              |
| namespace       | namespace                                | 否    | string | 是              |
| createTimeBegin | createTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  | 否              |
| createTimeEnd   | createTime的区间右边界                         | 否    | int64  | 否              |
| field           | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/k8s/dynamic/clusters/BCS-TEST-10000/ingress
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "BCS-TEST-10000",
      "createTime": "2017-10-16 20:17:27",
      "data": {
        "metadata": {
          "creationTimestamp": "2017-08-17T03:50:25Z",
          "generation": 1,
          "name": "",
          "namespace": "test",
          "resourceVersion": "",
          "selfLink": "",
          "uid": ""
        },
        "spec": {
          "rules": [
            {
              "host": "",
              "http": {
                "paths": [
                  {
                    "backend": {
                      "serviceName": "",
                      "servicePort": "http"
                    },
                    "path": "/"
                  }
                ]
              }
            }
          ]
        },
        "status": {
          "loadBalancer": {
            "ingress": [
              {
                "ip": "127.0.0.2"
              },
              {
                "ip": "127.0.0.1"
              }
            ]
          }
        }
      },
      "namespace": "default",
      "resourceName": "",
      "resourceType": "Ingress",
      "updateTime": "2017-10-16 20:17:27"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 9. namespace

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/k8s/dynamic/clusters/{clusterId}/namespace |
| METHOD                                   |
| GET                                      |



| 参数              | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| --------------- | ---------------------------------------- | ---- | ------ | -------------- |
| name            | 名字                                       | 否    | string | 是              |
| status          | 状态                                       | 否    | string | 是              |
| createTimeBegin | createTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  | 否              |
| createTimeEnd   | createTime的区间右边界                         | 否    | int64  | 否              |
| field           | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/k8s/dynamic/clusters/BCS-TEST-10000/namespace
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "BCS-TEST-10000",
      "createTime": "2017-10-16 20:18:46",
      "data": {
        "metadata": {
          "creationTimestamp": "2017-08-15T01:54:29Z",
          "name": "",
          "resourceVersion": "",
          "selfLink": "",
          "uid": ""
        },
        "spec": {
          "finalizers": [
            "kubernetes"
          ]
        },
        "status": {
          "phase": "Active"
        }
      },
      "namespace": "default",
      "resourceName": "",
      "resourceType": "Namespace",
      "updateTime": "2017-10-16 20:18:46"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 10. node

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/k8s/dynamic/clusters/{clusterId}/node |
| METHOD                                   |
| GET                                      |



| 参数              | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| --------------- | ---------------------------------------- | ---- | ------ | -------------- |
| name            | 名字                                       | 否    | string | 是              |
| externalID      | externalID                               | 否    | string | 是              |
| createTimeBegin | createTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  | 否              |
| createTimeEnd   | createTime的区间右边界                         | 否    | int64  | 否              |
| field           | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/k8s/dynamic/clusters/BCS-TEST-10000/node
```



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "BCS-TEST-10000",
      "createTime": "2017-10-16 20:20:30",
      "data": {
        "metadata": {
          "annotations": {
            "volumes.kubernetes.io/controller-managed-attach-detach": "true"
          },
          "creationTimestamp": "2017-04-27T09:21:28Z",
          "labels": {
            "beta.kubernetes.io/arch": "amd64",
            "beta.kubernetes.io/os": "linux",
            "kubernetes.io/hostname": "127.0.0.1"
          },
          "name": "127.0.0.1",
          "resourceVersion": "",
          "selfLink": "",
          "uid": ""
        },
        "spec": {
          "externalID": "127.0.0.1"
        },
        "status": {
          "addresses": [
            {
              "address": "127.0.0.1",
              "type": "LegacyHostIP"
            },
            {
              "address": "127.0.0.1",
              "type": "InternalIP"
            },
            {
              "address": "127.0.0.1",
              "type": "Hostname"
            }
          ],
          "allocatable": {
            "alpha.kubernetes.io/nvidia-gpu": "0",
            "cpu": "16",
            "memory": "65701000Ki",
            "pods": "200"
          },
          "capacity": {
            "alpha.kubernetes.io/nvidia-gpu": "0",
            "cpu": "16",
            "memory": "65701000Ki",
            "pods": "200"
          },
          "conditions": [
            {
              "lastHeartbeatTime": "2017-10-16T12:20:08Z",
              "lastTransitionTime": "2017-08-06T09:18:19Z",
              "message": "kubelet has sufficient disk space available",
              "reason": "KubeletHasSufficientDisk",
              "status": "False",
              "type": "OutOfDisk"
            },
            {
              "lastHeartbeatTime": "2017-10-16T12:20:08Z",
              "lastTransitionTime": "2017-04-27T09:21:28Z",
              "message": "kubelet has sufficient memory available",
              "reason": "KubeletHasSufficientMemory",
              "status": "False",
              "type": "MemoryPressure"
            },
            {
              "lastHeartbeatTime": "2017-10-16T12:20:08Z",
              "lastTransitionTime": "2017-04-27T09:21:28Z",
              "message": "kubelet has no disk pressure",
              "reason": "KubeletHasNoDiskPressure",
              "status": "False",
              "type": "DiskPressure"
            },
            {
              "lastHeartbeatTime": "2017-10-16T12:20:08Z",
              "lastTransitionTime": "2017-08-06T09:18:29Z",
              "message": "kubelet is posting ready status",
              "reason": "KubeletReady",
              "status": "True",
              "type": "Ready"
            }
          ],
          "daemonEndpoints": {
            "kubeletEndpoint": {
              "Port": 10250
            }
          },
          "images": [
            {
              "names": [
                "a",
                "b"
              ],
              "sizeBytes": 1511874035
            },
            {
              "names": [
                "c"
              ],
              "sizeBytes": 1511873989
            }
          ],
          "nodeInfo": {
            "architecture": "amd64",
            "bootID": "",
            "containerRuntimeVersion": "docker://1.13.1",
            "kernelVersion": "",
            "kubeProxyVersion": "v1.5.6",
            "kubeletVersion": "v1.5.6",
            "machineID": "",
            "operatingSystem": "linux",
            "osImage": "",
            "systemUUID": ""
          }
        }
      },
      "namespace": "default",
      "resourceName": "",
      "resourceType": "Node",
      "updateTime": "2017-10-16 20:20:30"
    }
  ],
  "message": "Success",
  "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### 11. daemonset

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/k8s/dynamic/clusters/{clusterId}/daemonset |
| METHOD                                   |
| GET                                      |



| 参数                 | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| ------------------ | ---------------------------------------- | ---- | ------ | -------------- |
| name               | 名字                                       | 否    | string | 是              |
| namespace          | namespace                                | 否    | string | 是              |
| resourceVersion    | resourceVersion                          | 否    | string | 是              |
| uid                | uid                                      | 否    | string | 是              |
| generation         | generation                               | 否    | int    | 否              |
| templateGeneration | templateGeneration                       | 否    | int    |                |
| createTimeBegin    | createTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  | 否              |
| createTimeEnd      | createTime的区间右边界                         | 否    | int64  | 否              |
| field              | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/k8s/dynamic/clusters/BCS-TEST-10000/daemonset
```



成功返回示例

```
{
    "code": 0,
    "data": [
        {
            "resourceType": "DaemonSet",
            "resourceName": "",
            "clusterId": "BCS-TEST-10000",
            "_id": "",
            "data": {
                "status": {
                    "numberReady": 3,
                    "observedGeneration": 1,
                    "updatedNumberScheduled": 3,
                    "currentNumberScheduled": 3,
                    "desiredNumberScheduled": 3,
                    "numberAvailable": 3,
                    "numberMisscheduled": 0
                },
                "metadata": {
                    "name": "",
                    "creationTimestamp": "2018-03-11T09:28:13Z",
                    "selfLink": "",
                    "resourceVersion": "",
                    "uid": "",
                    "annotations": {
                        "io.tencent.paas.webCache": "{\"volumes\": [{\"type\": \"emptyDir\", \"name\": \"\", \"source\": \"\"}], \"isUserConstraint\": false, \"remarkListCache\": [{\"key\": \"\", \"value\": \"\"}], \"labelListCache\": [{\"key\": \"\", \"value\": \"\"}], \"isMetric\": false, \"metricIdList\": [], \"affinityYaml\": \"\"}"
                    },
                    "generation": 1,
                    "labels": {
                        "io.tencent.paas.version": "",
                        "io.tencent.paas.projectid": "",
                        "io.tencent.paas.versionid": "",
                        "io.tencent.bcs.namespace": "test",
                        "io.tencent.bcs.clusterid": "BCS-TEST-10000",
                        "io.tencent.paas.instanceid": "",
                        "io.tencent.paas.templateid": "",
                        "io.tencent.bcs.cluster": "BCS-TEST-10000",
                        "io.tencent.bkdata.baseall.dataid": "",
                        "io.tencent.bkdata.container.stdlog.dataid": ""
                    },
                    "namespace": "test"
                },
                "spec": {
                    "templateGeneration": 1,
                    "updateStrategy": {
                        "type": "OnDelete"
                    },
                    "revisionHistoryLimit": 10,
                    "selector": {
                        "matchLabels": {
                            "io.tencent.bkdata.baseall.dataid": "",
                            "io.tencent.bkdata.container.stdlog.dataid": ""
                        }
                    },
                    "template": {
                        "spec": {
                            "schedulerName": "default-scheduler",
                            "securityContext": [],
                            "terminationGracePeriodSeconds": 10,
                            "containers": [
                                {
                                    "name": "container-1",
                                    "resources": [],
                                    "terminationMessagePath": "",
                                    "terminationMessagePolicy": "File",
                                    "image": "",
                                    "imagePullPolicy": "IfNotPresent"
                                }
                            ],
                            "dnsPolicy": "ClusterFirst",
                            "restartPolicy": "Always"
                        },
                        "metadata": {
                            "creationTimestamp": null,
                            "labels": {
                                "io.tencent.bkdata.baseall.dataid": "",
                                "io.tencent.bkdata.container.stdlog.dataid": ""
                            }
                        }
                    }
                }
            },
            "updateTime": "2018-03-14T07:47:39.208Z",
            "createTime": "2018-03-14T07:47:39.208Z",
            "namespace": "test"
        }
    ],
    "message": "success",
    "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}

```



##### 12. job

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/k8s/dynamic/clusters/{clusterId}/job |
| METHOD                                   |
| GET                                      |



| 参数              | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| --------------- | ---------------------------------------- | ---- | ------ | -------------- |
| name            | 名字                                       | 否    | string | 是              |
| namespace       | namespace                                | 否    | string | 是              |
| resourceVersion | resourceVersion                          | 否    | string | 是              |
| uid             | uid                                      | 否    | string | 是              |
| createTimeBegin | createTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  | 否              |
| createTimeEnd   | createTime的区间右边界                         | 否    | int64  | 否              |
| field           | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/k8s/dynamic/clusters/BCS-TEST-10000/job
```



成功返回示例

```
{
    "code": 0,
    "data": [
        {
            "_id": "",
            "data": {
                "metadata": {
                    "name": "",
                    "namespace": "test",
                    "resourceVersion": "",
                    "selfLink": "",
                    "uid": "",
                    "annotations": {
                        "io.tencent.paas.webCache": "{\"volumes\": [{\"type\": \"emptyDir\", \"name\": \"\", \"source\": \"\"}], \"isUserConstraint\": false, \"remarkListCache\": [{\"key\": \"\", \"value\": \"\"}], \"labelListCache\": [{\"key\": \"\", \"value\": \"\"}], \"isMetric\": false, \"metricIdList\": [], \"affinityYaml\": \"\"}"
                    },
                    "creationTimestamp": "2018-03-14T09:56:18Z",
                    "labels": {
                        "io.tencent.bkdata.container.stdlog.dataid": "",
                        "io.tencent.paas.instanceid": "",
                        "io.tencent.bcs.cluster": "BCS-TEST-10000",
                        "io.tencent.bkdata.baseall.dataid": "",
                        "io.tencent.paas.version": "v1",
                        "io.tencent.bcs.namespace": "test",
                        "io.tencent.paas.projectid": "",
                        "io.tencent.bcs.clusterid": "BCS-TEST-10000",
                        "io.tencent.paas.versionid": "",
                        "io.tencent.paas.templateid": ""
                    }
                },
                "spec": {
                    "parallelism": 1,
                    "selector": {
                        "matchLabels": {
                            "controller-uid": ""
                        }
                    },
                    "template": {
                        "metadata": {
                            "creationTimestamp": null,
                            "labels": {
                                "io.tencent.bkdata.container.stdlog.dataid": "",
                                "job-name": "",
                                "controller-uid": "",
                                "io.tencent.bkdata.baseall.dataid": ""
                            }
                        },
                        "spec": {
                            "imagePullSecrets": [
                                {
                                    "name": ""
                                }
                            ],
                            "restartPolicy": "Never",
                            "schedulerName": "default-scheduler",
                            "securityContext": [],
                            "terminationGracePeriodSeconds": 10,
                            "containers": [
                                {
                                    "image": "",
                                    "imagePullPolicy": "IfNotPresent",
                                    "name": "container-1",
                                    "resources": [],
                                    "terminationMessagePath": "",
                                    "terminationMessagePolicy": "File"
                                }
                            ],
                            "dnsPolicy": "ClusterFirst"
                        }
                    },
                    "activeDeadlineSeconds": 300,
                    "backoffLimit": 6,
                    "completions": 1
                },
                "status": {
                    "conditions": [
                        {
                            "lastTransitionTime": "2018-03-14T10:01:38Z",
                            "message": "Job has reach the specified backoff limit",
                            "reason": "BackoffLimitExceeded",
                            "status": "True",
                            "type": "Failed",
                            "lastProbeTime": "2018-03-14T10:01:38Z"
                        }
                    ],
                    "failed": 6,
                    "startTime": "2018-03-14T09:56:18Z"
                }
            },
            "updateTime": "2018-03-14T10:01:38.892Z",
            "createTime": "2018-03-14T09:56:18.803Z",
            "namespace": "test",
            "resourceType": "Job",
            "resourceName": "",
            "clusterId": "BCS-TEST-10000"
        }
    ],
    "message": "success",
    "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}


```



##### 13. statefulset

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /query/k8s/dynamic/clusters/{clusterId}/statefulset |
| METHOD                                   |
| GET                                      |



| 参数                  | 说明                                       | 必须   | 类型     | 支持逗号(,)分隔符多个查询 |
| ------------------- | ---------------------------------------- | ---- | ------ | -------------- |
| name                | 名字                                       | 否    | string | 是              |
| namespace           | namespace                                | 否    | string | 是              |
| resourceVersion     | resourceVersion                          | 否    | string | 是              |
| uid                 | uid                                      | 否    | string | 是              |
| generation          | generation                               | 否    | int    | 否              |
| podManagementPolicy | podManagementPolicy                      | 否    | string | 是              |
| updateStrategyType  | updateStrategyType                       | 否    | string | 是              |
| serviceName         | serviceName                              | 否    | string | 是              |
| createTimeBegin     | createTime的区间左边界 unix时间戳 如1509423130     | 否    | int64  | 否              |
| createTimeEnd       | createTime的区间右边界                         | 否    | int64  | 否              |
| field               | 指定返回的数据key 深度用点(.)分隔 如field=data.metadata | 否    | string | 是              |



请求示例

```
/query/k8s/dynamic/clusters/BCS-TEST-10000/statefulset
```



成功返回示例

```
{
    "code": 0,
    "data": [
        {
            "resourceName": "",
            "data": {
                "metadata": {
                    "annotations": {
                        "io.tencent.paas.webCache": "{\"volumes\": [{\"type\": \"emptyDir\", \"name\": \"\", \"source\": \"\"}], \"isUserConstraint\": false, \"remarkListCache\": [{\"key\": \"\", \"value\": \"\"}], \"labelListCache\": [{\"key\": \"app\", \"value\": \"rumpetroll\"}], \"isMetric\": false, \"metricIdList\": [], \"affinityYaml\": \"\"}"
                    },
                    "generation": 1,
                    "labels": {
                        "io.tencent.paas.projectid": "",
                        "io.tencent.bkdata.baseall.dataid": "",
                        "io.tencent.bcs.clusterid": "BCS-TEST-10000",
                        "io.tencent.bcs.cluster": "BCS-TEST-10000",
                        "io.tencent.bcs.namespace": "test",
                        "io.tencent.paas.version": "1.0",
                        "io.tencent.paas.templateid": "",
                        "io.tencent.bkdata.container.stdlog.dataid": "",
                        "io.tencent.paas.instanceid": "",
                        "io.tencent.paas.versionid": ""
                    },
                    "name": "",
                    "namespace": "test",
                    "resourceVersion": "",
                    "selfLink": "",
                    "uid": "",
                    "creationTimestamp": "2018-03-14T13:17:35Z"
                },
                "spec": {
                    "replicas": 1,
                    "revisionHistoryLimit": 10,
                    "selector": {
                        "matchLabels": {
                            "io.tencent.bkdata.baseall.dataid": "",
                            "io.tencent.bkdata.container.stdlog.dataid": "",
                            "app": ""
                        }
                    },
                    "serviceName": "test1",
                    "template": {
                        "spec": {
                            "dnsPolicy": "ClusterFirst",
                            "imagePullSecrets": [
                                {
                                    "name": ""
                                }
                            ],
                            "restartPolicy": "Always",
                            "schedulerName": "default-scheduler",
                            "securityContext": [],
                            "terminationGracePeriodSeconds": 10,
                            "containers": [
                                {
                                    "resources": [],
                                    "terminationMessagePath": "",
                                    "terminationMessagePolicy": "File",
                                    "env": [
                                        {
                                            "value": "",
                                            "name": "DOMAIN"
                                        },
                                        {
                                            "name": "MAX_CLIENT",
                                            "value": "2"
                                        }
                                    ],
                                    "image": "",
                                    "imagePullPolicy": "IfNotPresent",
                                    "name": "",
                                    "ports": [
                                        {
                                            "containerPort": 20000,
                                            "name": "port",
                                            "protocol": "TCP"
                                        }
                                    ]
                                }
                            ]
                        },
                        "metadata": {
                            "labels": {
                                "app": "rumpetroll",
                                "io.tencent.bkdata.baseall.dataid": "",
                                "io.tencent.bkdata.container.stdlog.dataid": ""
                            },
                            "creationTimestamp": null
                        }
                    },
                    "updateStrategy": {
                        "type": "OnDelete"
                    },
                    "podManagementPolicy": "OrderedReady"
                },
                "status": {
                    "currentRevision": "",
                    "observedGeneration": 1,
                    "readyReplicas": 1,
                    "replicas": 1,
                    "updateRevision": "",
                    "collisionCount": 0,
                    "currentReplicas": 1
                }
            },
            "updateTime": "2018-03-14T13:17:40.923Z",
            "createTime": "2018-03-14T13:17:35.481Z",
            "_id": "",
            "clusterId": "BCS-TEST-10000",
            "namespace": "test",
            "resourceType": "StatefulSet"
        }
    ],
    "message": "success",
    "result": true
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}



```





## 事件数据

##### list events

| 说明      |
| ------- |
| URL     |
| /events |
| METHOD  |
| GET     |





请求参数

| 参数                  | 示例             | 说明            | 必须      | 类型     |
| ------------------- | -------------- | ------------- | ------- | ------ |
| env                 | mesos          | 编排类型          | 否       | string |
| kind                | pod,rc         | 对象            | 否       | string |
| level               | warning        | 事件类别（等级）      | 否       | string |
| component           | scheduler      | 组件            | 否       | string |
| type                | killing        | 事件类型          | 否       | string |
| clusterId           | BCS-TEST-10000 | 所属集群          | 否       | string |
| timeBegin           | 1509423130     | 开始时间          | 否       | int64  |
| timeEnd             | 1509423130     | 结束时间          | 否       | int64  |
| offset              | 0              | 分页起始          | 否（默认0）  | int    |
| length              | 4              | 分页长度          | 否（默认最大） | int    |
| extraInfo.name      | app            | 额外参数name      | 否       | string |
| extraInfo.namespace | test           | 额外参数namespace | 否       | string |

说明：除```timeBegin```、```timeEnd```、```offset```、```length```以外，其余参数可以通过```,```分隔多个，如不带该参数则不过滤。





返回数据（返回的单个事件中的数据）

| 参数        | 说明       |
| --------- | -------- |
| env       | 编排类型     |
| kind      | 对象       |
| level     | 事件类别（等级） |
| component | 组件       |
| type      | 事件类型     |
| clusterId | 所属集群     |
| eventTime | 事件发生时间   |
| describe  | 事件说明（内容） |
| extraInfo | 额外信息     |
| data      | 事件详细数据   |



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "BCS-10001",
      "component": "kublet",
      "createTime": "2017-09-29 15:17:50",
      "data": {
        "a": "1",
        "b": "2"
      },
      "describe": "wow its killing itself",
      "env": "mesos",
      "eventTime": "2017-09-29 12:00:00",
      "kind": "rc",
      "level": "warning",
      "type": "killing"
    }
  ],
  "message": "Success",
  "result": true
}
```


失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### put events

| 说明      |
| ------- |
| URL     |
| /events |
| METHOD  |
| PUT     |



请求参数

- 接口见bcs-common/type/storage.go中的BcsStorageEventIf


- 将BcsStorageEventIf序列化放入body





成功返回示例

```
{
  "result": true,
  "code": 0,
  "message": "Success",
  "data": null
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10003,
  “message”: “Put resource failed.”,
  “data”: null
}
```



## 告警数据

##### list alarms

| 说明      |
| ------- |
| URL     |
| /alarms |
| METHOD  |
| GET     |





请求参数

| 参数        | 示例             | 说明        | 必须      | 类型     |
| --------- | -------------- | --------- | ------- | ------ |
| clusterId | BCS-TEST-10000 | 所属集群      | 否       | string |
| namespace | defaultGroup   | namespace | 否       | string |
| source    | 127.0.0.4      | 告警来源      | 否       | string |
| module    | bcs-health     | 模块        | 否       | string |
| type      | sms            | 类型        | 否       | string |
| timeBegin | 1509423130     | 开始时间      | 否       | int64  |
| timeEnd   | 1509423130     | 结束时间      | 否       | int64  |
| offset    | 0              | 分页起始      | 否（默认0）  | int    |
| length    | 4              | 分页长度      | 否（默认最大） | int    |

说明：除```timeBegin```、```timeEnd```、```offset```、```length```以外，其余参数可以通过```,```分隔多个，如不带该参数则不过滤。





返回数据（返回的单个告警中的数据）

| 参数           | 说明        |
| ------------ | --------- |
| clusterId    | 所属集群      |
| namespace    | namespace |
| source       | 告警来源      |
| module       | 模块        |
| type         | 类型        |
| receivedTime | 告警时间      |



成功返回示例

```
{
  "code": 0,
  "data": [
    {
      "_id": "",
      "clusterId": "developer-id",
      "createTime": "2017-11-15 19:34:41",
      "message": "developer health test message",
      "module": "",
      "namespace": "test",
      "receivedTime": "2017-11-15 19:34:41",
      "source": "127.0.0.4:52328",
      "type": "weixin"
    }
  ],
  "message": "Success",
  "result": true,
  "total": 1
}
```

失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: []
}
```



##### post alarms

| 说明      |
| ------- |
| URL     |
| /alarms |
| METHOD  |
| POST    |



请求参数

- 接口见bcs-common/type/storage.go中的BcsStorageAlarmIf


- 将BcsStorageAlarmIf序列化放入body



成功返回示例

```
{
  "result": true,
  "code": 0,
  "message": "Success",
  "data": null
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10003,
  “message”: “Put resource failed.”,
  “data”: null
}
```



## Watch类数据

注：watch类数据用于存放在zk上，通过zk的watch功能提供给其他模块消费



##### get resource

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /mesos/watch/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName} |
| METHOD                                   |
| GET                                      |



成功返回示例

```
{
  "result": true,
  "code": 0,
  "message": "Success",
  "data": // 上报的数据
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10004,
  “message”: “Get resource failed.”,
  “data”: null
}
```



##### put resource

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /mesos/watch/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName} |
| METHOD                                   |
| PUT                                      |



请求参数

- 接口见bcs-common/type/storage.go中的BcsStorageWatchIf


- 将BcsStorageWatchIf序列化放入body



成功返回示例

```
{
  "result": true,
  "code": 0,
  "message": "Success",
  "data": null
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10003,
  “message”: “Put resource failed.”,
  “data”: null
}
```



##### delete resource

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /mesos/watch/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName} |
| METHOD                                   |
| DELETE                                   |



成功返回示例

```
{
  "result": true,
  "code": 0,
  "message": "Success",
  "data": null
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10005,
  “message”: “Delete resource failed.”,
  “data”: null
}
```



##### list resource

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /mesos/watch/clusters/{clusterId}/namespaces/{namespace}/{resourceType} |
| METHOD                                   |
| GET                                      |



请求参数

| 参数    | 说明                                       | 必须   | 类型     |
| ----- | ---------------------------------------- | ---- | ------ |
| field | 指定data中单个资源返回的字段                         | 否    | string |
| extra | 额外条件json的base64编码，其中额外条件层次结构用"."连接，若key中本身含有“.”，则使用其unicode代替(\uff0e) | 否    | string |



成功返回示例

```
{
  "result": true,
  "code": 0,
  "message": "Success",
  "data": [
      // 上报的数据1,
      // 上报的数据2
  ]
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: null
}
```



## Metric数据

##### get metric

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /metric/clusters/{clusterId}/namespaces/{namespace}/{type}/{name} |
| METHOD                                   |
| GET                                      |



成功返回示例

```
{
  "message": "Success",
  "result": true,
  "data": {
    "updateTime": "2017-12-08 10:24:35",
    "createTime": "2017-12-08 10:24:35",
    "_id": "",
    "clusterId": "123",
    "namespace": "ns",
    "type": "type",
    "name": "name",
    "data": {}
  },
  "code": 0
}

```



失败返回示例

```
{
  “result”: false,
  “code”: 10004,
  “message”: “Get resource failed.”,
  “data”: null
}
```



##### put metric

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /metric/clusters/{clusterId}/namespaces/{namespace}/{type}/{name} |
| METHOD                                   |
| PUT                                      |



请求参数

- 接口见bcs-common/type/storage.go中的BcsStorageMetricIf


- 将BcsStorageMetricIf序列化放入body

成功返回示例

```
{
  "result": true,
  "code": 0,
  "message": "Success",
  "data": null
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10003,
  “message”: “Put resource failed.”,
  “data”: null
}
```



##### delete metric

| 说明                                       |
| ---------------------------------------- |
| URL                                      |
| /metric/clusters/{clusterId}/namespaces/{namespace}/{type}/{name} |
| METHOD                                   |
| DELETE                                   |



成功返回示例

```
{
  "result": true,
  "code": 0,
  "message": "Success",
  "data": null
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10005,
  “message”: “Delete resource failed.”,
  “data”: null
}
```



##### query metric

| 说明                           |
| ---------------------------- |
| URL                          |
| /metric/clusters/{clusterId} |
| METHOD                       |
| GET                          |



请求参数

| 参数        | 说明               | 必须   | 类型     |
| --------- | ---------------- | ---- | ------ |
| namespace | namespace        | 否    | String |
| type      | type             | 否    | String |
| name      | name             | 否    | String |
| field     | 指定data中单个资源返回的字段 | 否    | string |



成功返回示例

```
{
  "result": true,
  "code": 0,
  "message": "Success",
  "data": [
      // 上报的数据1,
      // 上报的数据2
  ]
}
```



失败返回示例

```
{
  “result”: false,
  “code”: 10006,
  “message”: “List resource failed.”,
  “data”: null
}
```


