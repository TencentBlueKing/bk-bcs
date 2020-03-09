# bcs apiserver v4 http api

## Response
### http状态码
- 成功：2xx
- 失败：4xx、5xx

### 返回格式
```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```
- code int //"0"表示成功，>0表示失败
- message string
- data interface{}   //成功返回的数据

## header
BCS-ClusterID   //集群id，例如：BCS-xxxxx-xxxxx

## url prefix
### bcs-api
http://{ipaddr}:{port}/bcsapi

示例：
curl -H "BCS-ClusterID: BCS-xxxxx-xxxxx" -X GET http://127.0.0.1:8899/bcsapi/v4/scheduler/mesos/namespaces/defaultGroup/applications

## api
### api list

* Application
  - [**create application**](#createapplication)
  - [**update application**](#updateapplication)
  - [**delete application**](#deleteapplication)
  - [**scale application**](#scaleapplication)
  - [**rollback application**](#rollbackapplication)
  - [**fetch application**](#fetchapplication)
  - [**list applications**](#listapplications)
  - [**list taskgroups**](#listtaskgroups)
  - [**list tasks**](#listtasks)
  - [**rescheduler taskgroup**](#reschedulertaskgroup)
  - [**list versions**](#listversions)
  - [**fetch version**](#fetchversion)
  - [**send application message**](#sendapplicationmessage)
  - [**send taskgroup message**](#sendtaskgroupmessage)
* ConfigMap
  - [**create configmap**](#createconfigmap)
  - [**update configmap**](#updateconfigmap)
  - [**delete configmap**](#deleteconfigmap)
* Secret
  - [**create secret**](#createsecret)
  - [**update secret**](#updatesecret)
  - [**delete secret**](#deletesecret)
* Sercice
  - [**create service**](#createservice)
  - [**update service**](#updateservice)
  - [**delete service**](#deleteservice)
- [**get cluster resources**](#getclusterresources)
- [**get cluster endpoints**](#getclusterendpoints)
- [**get cluster offers**](#getclusteroffers)
- [**health check report**](#healthcheckreport)
* Deployment
  - [**create deployment**](#createdeployment)
  - [**update deployment**](#updatedeployment)
  - [**cancel update deployment**](#cancelupdatedeployment)
  - [**pause update deployment**](#pauseupdatedeployment)
  - [**resume update deployment**](#resumeupdatedeployment)
  - [**scale deployment**](#scaledeployment)
  - [**delete deployment**](#deletedeployment)
* AgentSetting
  - [**get agent setting list**](#getagentsettinglist)
  - [**set agent setting list**](#setagentsettinglist)
  - [**delete agent setting list**](#deleteagentsettinglist)
  - [**update agent setting list**](#updateagentsettinglist)
  - [**disable agent list**](#disableagentlist)
  - [**enable agent list**](#enableagentlist)
  - [**taint agents**](#taintagents)
* [自定义资源定义(CustomResourceDefinition)](#customresourcedefinition)
  - [**create**](#createcrd)
  - [**get**](#getcrd)
  - [**delete**](#deletecrd)
* [自定义资源(CustomResource)](#customresource)
- [**create crd**](#createcrd)
- [**delete crd**](#deletecrd)
* Mesos Json定义
  - [**get deployment definition json**](getDeploymentDef)
  - [**get application definition json**](getApplicationDef)
* 信号与控制
  - [**commit image**](#commitimage)
  - [**send signal to application**](#sendapplicationSignal)
  - [**send signal to taskgroup**](#sendtaskgroupSignal)
  - [**send application/deployment command**](#sendCommand)
  - [**get application/deloyment command**](#getCommand)
  - [**delete application/deployment command**](#deleteCommand)
* Webhook
  - [**create admission webhook**](#createadmission)
  - [**update admission webhook**](#updateadmission)
  - [**get admission webhook**](#getadmission)
  - [**delete admission webhook**](#deleteadmission)

### createApplication
#### 描述
创建application

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/applications

#### 请求方式
- POST

#### 请求参数
Content-type: application/json

``` json
{
  "apiVersion": "v4",
  "kind": "application",
  "appname": "app-test",
  "updatePolicy": {
    "updateDelay": 10,
    "MaxRetries": 10,
    "maxFailovers": 10,
    "action": ""
  },
  "restartPolicy": {
    "policy": "Never",
    "interval": 5,
    "backoff": 10
  },
  "killPolicy": {
    "gracePeriod": 5
  },
  "constraint": {
    "selector": {},
    "IntersectionItem": [
    ]
  },
  "metadata": {
    "annotations": {},
    "labels": {
      "podname": "app-test"
    },
    "name": "app-test",
    "namespace": "defaultGroup"
  },
  "spec": {
    "instance": 2,
    "template": {
      "spec": {
        "containers": [
          {
            "command": "",
            "args": [
            ],
            "parameters": [],
            "type": "MESOS",
            "env": [
            ],
            "image": "dockerhub:8443/bcs/game/gamesvr:v1.1.1",
            "imagePullUser": "",
            "imagePullPasswd": "",
            "imagePullPolicy": "Always",
            "privileged": false,
            "ports": [
              {
                "containerPort": 9999,
                "name": "test-port",
                "protocol": "HTTP"
              }
            ],
            "healthChecks": [
              {
                "type": "HTTP",
                "delaySeconds": 10,
                "gracePeriodSeconds": 10,
                "intervalSeconds": 10,
                "timeoutSeconds": 5,
                "consecutiveFailures": 10,
                "http": {
                  "port": 0,
                  "portName": "test-port",
                  "scheme": "http",
                  "path": "/dd"
                }
              },
              {
                "type": "REMOTE_HTTP",
                "delaySeconds": 10,
                "gracePeriodSeconds": 10,
                "intervalSeconds": 10,
                "timeoutSeconds": 5,
                "consecutiveFailures": 10,
                "http": {
                  "port": 0,
                  "portName": "test-port",
                  "scheme": "http",
                  "path": "/dd",
                  "headers": {
                    "key1": "value1"
                  }
                }
              },
              {
                "type": "REMOTE_TCP",
                "delaySeconds": 10,
                "gracePeriodSeconds": 10,
                "intervalSeconds": 10,
                "timeoutSeconds": 5,
                "consecutiveFailures": 10,
                "tcp": {
                  "port": 0,
                  "portName": "test-port"
                }
              }
            ],
            "resources": {
              "limits": {
                "cpu": "2",
                "memory": "100"
              }
            },
            "volumes": [],
            "secrets": [],
            "configmaps": []
          }
        ],
        "networkMode": "BRIDGE",
        "networkType": "BRIDGE"
      }
    }
  }
}
```

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X POST -d "{application.json}" http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/applications

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### updateApplication
#### 描述
更新application

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/applications?instances=2&args=resource

#### 请求方式
- PUT

#### 请求参数
Content-type: application/json
- instances //udpate数量
- args=resource  //是否动态调整容器的配额，如果args=resource则是，否则不是。此情况下，不会重启容器

``` json
{
  "apiVersion": "v4",
  "kind": "application",
  "appname": "app-test",
  "updatePolicy": {
    "updateDelay": 10,
    "MaxRetries": 10,
    "maxFailovers": 10,
    "action": ""
  },
  "restartPolicy": {
    "policy": "Never",
    "interval": 5,
    "backoff": 10
  },
  "killPolicy": {
    "gracePeriod": 5
  },
  "constraint": {
    "selector": {},
    "IntersectionItem": [
    ]
  },
  "metadata": {
    "annotations": {},
    "labels": {
      "podname": "app-test"
    },
    "name": "app-test",
    "namespace": "defaultGroup"
  },
  "spec": {
    "instance": 2,
    "template": {
      "spec": {
        "containers": [
          {
            "command": "",
            "args": [
            ],
            "parameters": [],
            "type": "MESOS",
            "env": [
            ],
            "image": "dockerhub:8443/bcs/game/gamesvr:v1.1.1",
            "imagePullUser": "",
            "imagePullPasswd": "",
            "imagePullPolicy": "Always",
            "privileged": false,
            "ports": [
              {
                "containerPort": 9999,
                "name": "test-port",
                "protocol": "HTTP"
              }
            ],
            "healthChecks": [
              {
                "type": "HTTP",
                "delaySeconds": 10,
                "gracePeriodSeconds": 10,
                "intervalSeconds": 10,
                "timeoutSeconds": 5,
                "consecutiveFailures": 10,
                "http": {
                  "port": 0,
                  "portName": "test-port",
                  "scheme": "http",
                  "path": "/dd"
                }
              },
              {
                "type": "REMOTE_HTTP",
                "delaySeconds": 10,
                "gracePeriodSeconds": 10,
                "intervalSeconds": 10,
                "timeoutSeconds": 5,
                "consecutiveFailures": 10,
                "http": {
                  "port": 0,
                  "portName": "test-port",
                  "scheme": "http",
                  "path": "/dd",
                  "headers": {
                    "key1": "value1"
                  }
                }
              },
              {
                "type": "REMOTE_TCP",
                "delaySeconds": 10,
                "gracePeriodSeconds": 10,
                "intervalSeconds": 10,
                "timeoutSeconds": 5,
                "consecutiveFailures": 10,
                "tcp": {
                  "port": 0,
                  "portName": "test-port"
                }
              }
            ],
            "resources": {
              "limits": {
                "cpu": "2",
                "memory": "100"
              }
            },
            "volumes": [],
            "secrets": [],
            "configmaps": []
          }
        ],
        "networkMode": "BRIDGE",
        "networkType": "BRIDGE"
      }
    }
  }
}
```

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT -d "{application.json}" http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/applications

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### deleteApplication
#### 描述
删除application

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/applications/{name}?enforce=0

#### 请求方式
- DELETE

#### 请求参数
- ns  //namespace
- name //application name
- enforce  //enum: 1表示强制删除；0表示不强制删除

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X DELETE http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/applications/app-test

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### scaleApplication
#### 描述
scale application

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/applications/{name}/scale/{instances}

#### 请求方式
- PUT

#### 请求参数
- ns //namespace
- name //application name
- instances

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/applications/app-test/scale/5

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### rollbackApplication
#### 描述
回滚application

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/applications/rollback

#### 请求方式
- PUT

#### 请求参数
Content-type: application/json

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT -d "{application.json}" http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/applications/rollback

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### fetchApplication
#### 描述
fetch application

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/applications/{name}

#### 请求方式
- GET

#### 请求参数
- ns  //namespace
- name //application name

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X GET http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/applications/app-test

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### listApplications
#### 描述
list applications

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/applications

#### 请求方式
- GET

#### 请求参数
- ns  //namespace

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X GET http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/applications

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### listTaskgroups
#### 描述
list taskgroups

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/applications/{name}/taskgroups

#### 请求方式
- GET

#### 请求参数
- ns  //namespace
- name // application name

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X GET http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/applications/app-test/taskgroups

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### listTasks
#### 描述
list tasks

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/applications/{name}/tasks

#### 请求方式
- GET

#### 请求参数
- ns  //namespace
- name // application name

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X GET http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/applications/app-test/tasks

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### reschedulerTaskgroup
#### 描述
重新调度taskgroup

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/applications/{appname}/taskgroups/{name}/rescheduler

#### 请求方式
- PUT

#### 请求参数
- name // taskgroup name

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/applications/app-name/taskgroups/taskgroup-name/rescheduler

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### listVersions
#### 描述
list application versions

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/applications/{name}/versions

#### 请求方式
- GET

#### 请求参数
- ns //namespace
- name // application name

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/applications/app-test/versions

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### fetchVersion
#### 描述
fetch application version

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/applications/{name}/versions/{versionid}

#### 请求方式
- GET

#### 请求参数
- ns //namespace
- name // application name
- versionid //version id

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/applications/app-test/versions/version-id

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### sendApplicationMessage
#### 描述
send application message

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/applications/{name}/message

#### 请求方式
- POST

#### 请求参数
- ns //namespace
- name // application name

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT -d "{message.json}" http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/applications/app-test/message

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### sendTaskgroupMessage
#### 描述
send taskgroup message

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/applications/{name}/taskgroups/{taskgroup-name}/message

#### 请求方式
- POST

#### 请求参数
- ns //namespace
- name // application name
- taskgroup-name  //taskgroup name

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT -d "{message.json}" http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/applications/app-test/taskgroups/taskgroup-name/message

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### createConfigmap
#### 描述
创建 configmap

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/configmaps

#### 请求方式
- POST

#### 请求参数
Content-type: application/json

``` json
{
  "apiVersion": "v4",
  "kind": "configmap",
  "metadata": {
    "name": "configmap-test",
    "namespace": "defaultGroup",
    "labels": {
    }
  },
  "datas": {
    "item-one": {
      "type": "file",
      "content": "Y29uZmlnIGNvbnRleHQ="
    },
    "item-two": {
      "type": "file",
      "content": "Y29uZmlnIGNvbnRleHQ="
    }
  }
}
```

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X POST -d "{configmap.json}" http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/configmaps

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### updateConfigmap
#### 描述
更新 configmap

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/configmaps

#### 请求方式
- PUT

#### 请求参数
Content-type: application/json

``` json
{
  "apiVersion": "v4",
  "kind": "configmap",
  "metadata": {
    "name": "configmap-test",
    "namespace": "defaultGroup",
    "labels": {
    }
  },
  "datas": {
    "item-one": {
      "type": "file",
      "content": "Y29uZmlnIGNvbnRleHQ="
    },
    "item-two": {
      "type": "file",
      "content": "Y29uZmlnIGNvbnRleHQ="
    }
  }
}
```

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT -d "{configmap.json}" http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/configmaps

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### deleteConfigmap
#### 描述
删除 configmap

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/configmaps/{name}

#### 请求方式
- DELETE

#### 请求参数
- ns //namespace
- name //

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X DELETE -d "{configmap.json}" http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/configmaps/configmap-name

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### createSecret
#### 描述
创建 secret

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/secrets

#### 请求方式
- POST

#### 请求参数
Content-type: application/json

``` json
{
  "apiVersion": "v4",
  "kind": "secret",
  "metadata": {
    "name": "secret-test",
    "namespace": "defaultGroup",
    "labels": {}
  },
  "type": "",
  "datas": {
    "secret-subkey": {
      "path": "SECRET_ENV_TEST",
      "content": "Y29uZmlnIGNvbnRleHQ="
    }
  }
}
```

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X POST -d "{secret.json}" http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/secrets

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### updateSecret
#### 描述
更新 secret

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/secrets

#### 请求方式
- PUT

#### 请求参数
Content-type: application/json

``` json
{
  "apiVersion": "v4",
  "kind": "secret",
  "metadata": {
    "name": "secret-test",
    "namespace": "defaultGroup",
    "labels": {}
  },
  "type": "",
  "datas": {
    "secret-subkey": {
      "path": "SECRET_ENV_TEST",
      "content": "Y29uZmlnIGNvbnRleHQ="
    }
  }
}
```

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT -d "{secret.json}" http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/secrets

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### deleteSecret
#### 描述
删除 secret

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/secrets/{name}

#### 请求方式
- DELETE

#### 请求参数
- ns //namespace
- name //

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X DELETE -d "{secret.json}" http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/secrets/secret-name

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### createService
#### 描述
创建 service

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/services

#### 请求方式
- POST

#### 请求参数
Content-type: application/json

``` json
{
  "apiVersion": "v4",
  "kind": "service",
  "metadata": {
    "name": "service-test",
    "namespace": "defaultGroup",
    "labels": {
      "BCSGROUP": "external"
    }
  },
  "spec": {
    "selector": {
      "podname": "app-test"
    },
    "ports": [
      {
        "name": "test-port",
        "protocol": "tcp",
        "servicePort": 8899
      }
    ]
  }
}
```

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X POST -d "{service.json}" http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/services

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### updateService
#### 描述
更新 service

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/services

#### 请求方式
- PUT

#### 请求参数
Content-type: application/json

``` json
{
  "apiVersion": "v4",
  "kind": "service",
  "metadata": {
    "name": "service-test",
    "namespace": "defaultGroup",
    "labels": {
      "BCSGROUP": "external"
    }
  },
  "spec": {
    "selector": {
      "podname": "app-test"
    },
    "ports": [
      {
        "name": "test-port",
        "protocol": "tcp",
        "servicePort": 8899
      }
    ]
  }
}
```

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT -d "{service.json}" http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/services

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### deleteService
#### 描述
删除 service

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/services/{name}

#### 请求方式
- DELETE

#### 请求参数
- ns //namespace
- name //

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X DELETE http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/services/secret-name

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### getClusterResources
#### 描述
get cluster resources

#### 请求地址
- /v4/scheduler/mesos/cluster/resources

#### 请求方式
- GET

#### 请求参数

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X GET http://{Bcs-Domain}/v4/scheduler/mesos/cluster/resources

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### getClusterEndpoints
#### 描述
get cluster endpoints

#### 请求地址
- /v4/scheduler/mesos/cluster/endpoints

#### 请求方式
- GET

#### 请求参数

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X GET http://{Bcs-Domain}/v4/scheduler/mesos/cluster/endpoints

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### getClusteroffers
#### 描述
获取集群当前的offers信息

#### 请求地址
- /v4/scheduler/mesos/cluster/current/offers

#### 请求方式
- GET

#### 请求参数

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X GET http://{Bcs-Domain}/v4/scheduler/mesos/cluster/current/offers

#### 返回结果

```json
{
	"result": true,
	"code": 0,
	"message": "success",
	"data": [{
		"id": {
			"value": "f136f5f2-45e6-4bff-8418-d4403fb68936-O17911"
		},
		"framework_id": {
			"value": "350009f5-1caf-41dd-97b1-9f0d9a4643c8-0000"
		},
		"agent_id": {
			"value": "6cd1b3e9-3f95-44fe-b804-10d273903ee9-S0"
		},
		"hostname": "mesos-slave-5",
		"url": {
			"scheme": "http",
			"address": {
				"hostname": "mesos-slave-5",
				"ip": "127.0.0.1",
				"port": 8080
			},
			"path": "/slave(1)"
		},
		"resources": [{
			"name": "cpus",
			"type": 0,
			"scalar": {
				"value": 4
			},
			"role": "*"
		}, {
			"name": "mem",
			"type": 0,
			"scalar": {
				"value": 30878
			},
			"role": "*"
		}, {
			"name": "disk",
			"type": 0,
			"scalar": {
				"value": 795945
			},
			"role": "*"
		}, {
			"name": "ports",
			"type": 1,
			"ranges": {
				"range": [{
					"begin": 31000,
					"end": 32000
				}]
			},
			"role": "*"
		}],
		"attributes": [{
			"name": "InnerIP",
			"type": 3,
			"text": {
				"value": "127.0.0.1"
			}
		}, {
			"name": "City",
			"type": 3,
			"text": {
				"value": "shenzhen"
			}
		}, {
			"name": "ip-resources",
			"type": 0,
			"scalar": {
				"value": 1
			}
		}]
	}]
}
```

### createDeployment
#### 描述
创建deployment

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/deployments

#### 请求方式
- POST

#### 请求参数
Content-type: application/json

``` json
{
  "apiVersion": "v4",
  "kind": "deployment",
  "metadata": {
    "annotations": {},
    "labels": {
      "label_deployment": "label_deployment"
    },
    "name": "deployment-test",
    "namespace": "defaultGroup"
  },
  "updatePolicy": {
    "updateDelay": 10,
    "MaxRetries": 10,
    "maxFailovers": 10,
    "action": ""
  },
  "restartPolicy": {
    "policy": "Always",
    "interval": 5,
    "backoff": 10
  },
  "constraint": {
    "selector": {},
    "IntersectionItem": [
    ]
  },
  "spec": {
    "instance": 2,
    "selector": {
      "podname": "deployment-test"
    },
    "template": {
      "metadata": {
        "labels": {
          "label_deployment": "label_deployment"
        },
        "name": "deployment-test",
        "namespace": "defaultGroup"
      },
      "spec": {
        "containers": [
          {
            "command": "",
            "args": [
            ],
            "parameters": [],
            "type": "MESOS",
            "env": [
            ],
            "image": "dockerhub:8443/bcs/game/gamesvr:v1.1.1",
            "imagePullUser": "",
            "imagePullPasswd": "",
            "imagePullPolicy": "Always",
            "privileged": false,
            "ports": [
              {
                "containerPort": 8899,
                "name": "test-port",
                "protocol": "HTTP"
              }
            ],
            "healthChecks": [
              {
                "protocol": "tcp",
                "path": "/http/path/only",
                "delaySeconds": 10,
                "gracePeriodSeconds": 12,
                "intervalSeconds": 10,
                "timeoutSeconds": 10,
                "consecutiveFailures": 10
              }
            ],
            "resources": {
              "limits": {
                "cpu": "0.1",
                "memory": "1"
              }
            },
            "volumes": [],
            "secrets": [],
            "configmaps": []
          }
        ],
        "networkMode": "BRIDGE",
        "networkType": "BRIDGE"
      }
    },
    "strategy": {
      "type": "RollingUpdate",
      "rollingupdate": {
        "maxUnavilable": 1,
        "maxSurge": 1,
        "upgradeDuration": 60,
        "autoUpgrade": false,
        "rollingOrder": "CreateFirst",
        "pause": false
      }
    }
  }
}
```

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X POST -d "{deployment.json}" http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/deployments

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### updateDeployment
#### 描述
更新deployment

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/deployments

#### 请求方式
- PUT

#### 请求参数
Content-type: application/json

``` json
{
  "apiVersion": "v4",
  "kind": "deployment",
  "metadata": {
    "annotations": {},
    "labels": {
      "label_deployment": "label_deployment"
    },
    "name": "deployment-test",
    "namespace": "defaultGroup"
  },
  "updatePolicy": {
    "updateDelay": 10,
    "MaxRetries": 10,
    "maxFailovers": 10,
    "action": ""
  },
  "restartPolicy": {
    "policy": "Always",
    "interval": 5,
    "backoff": 10
  },
  "constraint": {
    "selector": {},
    "IntersectionItem": [
    ]
  },
  "spec": {
    "instance": 2,
    "selector": {
      "podname": "deployment-test"
    },
    "template": {
      "metadata": {
        "labels": {
          "label_deployment": "label_deployment"
        },
        "name": "deployment-test",
        "namespace": "defaultGroup"
      },
      "spec": {
        "containers": [
          {
            "command": "",
            "args": [
            ],
            "parameters": [],
            "type": "MESOS",
            "env": [
            ],
            "image": "dockerhub:8443/bcs/game/gamesvr:v1.1.1",
            "imagePullUser": "",
            "imagePullPasswd": "",
            "imagePullPolicy": "Always",
            "privileged": false,
            "ports": [
              {
                "containerPort": 8899,
                "name": "test-port",
                "protocol": "HTTP"
              }
            ],
            "healthChecks": [
              {
                "protocol": "tcp",
                "path": "/http/path/only",
                "delaySeconds": 10,
                "gracePeriodSeconds": 12,
                "intervalSeconds": 10,
                "timeoutSeconds": 10,
                "consecutiveFailures": 10
              }
            ],
            "resources": {
              "limits": {
                "cpu": "0.1",
                "memory": "1"
              }
            },
            "volumes": [],
            "secrets": [],
            "configmaps": []
          }
        ],
        "networkMode": "BRIDGE",
        "networkType": "BRIDGE"
      }
    },
    "strategy": {
      "type": "RollingUpdate",
      "rollingupdate": {
        "maxUnavilable": 1,
        "maxSurge": 1,
        "upgradeDuration": 60,
        "autoUpgrade": false,
        "rollingOrder": "CreateFirst",
        "pause": false
      }
    }
  }
}
```

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT -d "{deployment.json}" http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/{ns}/deployments

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### cancelUpdateDeployment
#### 描述
取消更新deployment

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/deployments/{name}/cancelupdate

#### 请求方式
- PUT

#### 请求参数

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/deployments/deployment-name/cancelupdate

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### pauseUpdateDeployment
#### 描述
暂停更新deployment

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/deployments/{name}/pauseupdate

#### 请求方式
- PUT

#### 请求参数

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/deployments/deployment-name/pauseupdate

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### resumeUpdateDeployment
#### 描述
继续更新deployment

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/deployments/{name}/resumeupdate

#### 请求方式
- PUT

#### 请求参数

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/deployments/deployment-name/resumeupdate

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### deleteDeployment
#### 描述
删除deployment

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/deployments/{name}

#### 请求方式
- DELETE

#### 请求参数

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X DELETE http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/deployments/deployment-name

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### scaleDeployment
#### 描述
scale deployment

#### 请求地址
- /v4/scheduler/mesos/namespaces/{ns}/deployments/{name}/scale/{instnaces}

#### 请求方式
- PUT

#### 请求参数
- ns //namespace
- name // deployment name
- instances //taskgroup数量，例如：5

#### 请求示例
curl -H "BCS-ClusterID: {ClusterID}" -X PUT http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/defaultGroup/deployments/deployment-name/scale/5

#### 返回结果

```json
{
    "code": 0,
    "message":"success",
    "data":{
    }
}
```

### getAgentSettingList
#### 描述
获取批量宿主机的属性设置，例如：是否停用

#### 请求地址
- /v4/scheduler/mesos/agentsettings

#### 请求方式
- GET

#### 请求参数
- ips IP列表,多个IP使用逗号隔开；ips为空则查询当前所有设置的机器；ips中的IP如果在数据库中没有设置记录，则返回列表中不会有该IP的数据。

#### 请求示例
- curl -H "BCS-ClusterID: {ClusterID}" -X GET http://{Bcs-Domain}/v4/scheduler/mesos/agentsettings/?ips=127.0.0.1
- curl -H "BCS-ClusterID: {ClusterID}" -X GET http://{Bcs-Domain}/v4/scheduler/mesos/agentsettings/

#### 返回结果

```json
{
    "result":true,
    "code":0,
    "message":"success",
    "data":[
        {
            "innerIP":"127.0.0.1",
            "disabled":false,
            "strings":{
                "attr3":{
                    "value":"333333"
                },
                "attr4":{
                    "value":"444444444"
                },
                "district":{
                    "value":"shenzhen"
                }
            },
            "scalars":{
                "attrScalar1":{
                    "value":999.9
                }
            }
        }
    ]
}
```


### setAgentSettingList
#### 描述
批量设置宿主机的属性，全量更新

#### 请求地址
- /v4/scheduler/mesos/agentsettings

#### 请求方式
- POST

#### 请求参数
- body中带[]BcsClusterAgentSetting

#### 请求示例
- curl -H "BCS-ClusterID: {ClusterID}" -d "[{\"innerIP\":\"127.0.0.1\",\"disabled\":false,\"strings\":{\"attr1\":{\"value\":\"1111\"},\"attr2\":{\"value\":\"22222\"}}}]"  http://{Bcs-Domain}/v4/scheduler/mesos/agentsettings

#### 返回结果

```json
{
    "result":true,
    "code":0,
    "message":"success",
    "data":null
}
```

### deleteAgentSettingList
#### 描述
批量删除宿主机的属性设置

#### 请求地址
- /v4/scheduler/mesos/agentsettings

#### 请求方式
- DELETE

#### 请求参数
- ips IP列表,多个IP使用逗号隔开

#### 请求示例
- curl -H "BCS-ClusterID: {ClusterID}" -X DELETE  http://{Bcs-Domain}/v4/scheduler/mesos/agentsettings?ips=127.0.0.1

#### 返回结果

```json
{
    "result":true,
    "code":0,
    "message":"success",
    "data":null
}
```

### updateAgentSettingList
#### 描述
给批量宿主机修改或者增加属性，增量更新

#### 请求地址
- /v4/scheduler/mesos/agentsettings/update

#### 请求方式
- POST

#### 请求参数
- body中带BcsClusterAgentSettingUpdate,其中包含IPS和属性信息

#### 请求示例
- curl -H "BCS-ClusterID: {ClusterID}" -d "{\"ips\":[\"127.0.0.1\",\"127.0.0.1\"],\"name\":\"attr3\",\"valuetype\":103,\"text\":{\"value\":\"333333\"}}"  http://{Bcs-Domain}/v4/scheduler/mesos/agentsettings/update

#### 返回结果

```json
{
    "result":true,
    "code":0,
    "message":"success",
    "data":null
}
```

### disableAgentList
#### 描述
批量停用宿主机： 停用后的宿主机不会再部署新的容器，但已有的容器不会做处理

#### 请求地址
- /v4/scheduler/mesos/agentsettings/disable

#### 请求方式
- POST

#### 请求参数
- ips IP列表,多个IP使用逗号隔开

#### 请求示例
- curl -H "BCS-ClusterID: {ClusterID}" -d ""  http://{Bcs-Domain}/v4/scheduler/mesos/agentsettings/disable?ips=127.0.0.1

#### 返回结果

```json
{
    "result":true,
    "code":0,
    "message":"success",
    "data":null
}
```

### enableAgentList
#### 描述
批量启用宿主机：机器默认处于启用状态，被disable后处于停用状态，需要enable才会恢复使用

#### 请求地址
- /v4/scheduler/mesos/agentsettings/enable

#### 请求方式
- POST

#### 请求参数
- ips IP列表,多个IP使用逗号隔开

#### 请求示例
- curl -H "BCS-ClusterID: {ClusterID}" -d ""  http://{Bcs-Domain}/v4/scheduler/mesos/agentsettings/enable?ips=127.0.0.1

#### 返回结果

```json
{
    "result":true,
    "code":0,
    "message":"success",
    "data":null
}
```

### taintagents
#### 描述
对node打污点，此node默认不能被应用调度；支持批量操作

#### 请求地址
- /v4/scheduler/mesos/agentsettings/taint

#### 请求方式
- PUT

#### 请求参数，body
```json
[
    {
        "innerIP":"127.0.0.1",
        "noSchedule":{
            "key1":"value1"
        }
    },
    {
        "innerIP":"127.0.0.2",
        "noSchedule":{
            "key1":"value1"
        }
    }
]
```

#### 请求示例
- curl -H "BCS-ClusterID: {ClusterID}" -d ""  http://{Bcs-Domain}/v4/scheduler/mesos/agentsettings/taint

#### 返回结果

```json
{
    "result":true,
    "code":0,
    "message":"success",
    "data":null
}
```

### commitImage
#### 描述
容器快照

#### 请求地址
- /v4/scheduler/mesos/image/commit/{taskgroupid}?image={origin_image}&url={target_image}

#### 请求方式
- POST

#### 请求参数
- taskgroupid   //taskgroup id
- origin_image  //原镜像
- target_image  //目标镜像

#### 请求示例
- curl -H "BCS-ClusterID: {ClusterID}" -d ""  http://{Bcs-Domain}/v4/scheduler/mesos/image/commit/0.app-test.defaultGroup.10001.1527839061316356269?image=docker.hub.com/bcs/server:v1&url=docker.hub.com/bcs/server:v2

#### 返回结果

```json
{
    "result":true,
    "code":0,
    "message":"success",
    "data":null
}
```

### customresourcedefinition
#### createcrd
创建自定义资源类型

#### 请求地址
- /v4/scheduler/mesos/customresourcedefinitions

#### 请求方式
- POST

#### 请求参数
```json
{
  "apiVersion": "v4", //api请求版本，当前为v4，必填
  "kind": "CustomResourceDefinition",//类型，必填
  "metadata": {
    "name": "logcollector.bkbcs.tencent.com" //自定义资源名称，必须唯一，由以下group和plural组合
  },
  "spec": {
    //自定义资源属组和版本信息，用于构成后续操作api部分
    //构成格式：/v4/scheduler/mesos/customresources/{group}/{version}/{plural}
    // 例如/v4/scheduler/mesos/customresources/bkbcs.tencent.com/v1/crontabs
    "group": "bkbcs.tencent.com", 
    "versions": [
      {
        "name": "v1",
        "served": true, //确认该版本是否生效
        "storage": true //落地存储至少一个版本需要标志
      }
    ],
    "scope": "Namespaced",//CRD数据生效范围，Namespaced,Cluster
    "names": {
      "plural": "crontabs", //URL组成部分
      "singular": "crontab", //命令行操作下使用类型，用于bcs-client命令映射与校验
      "kind": "CronTab" //驼峰命名的数据类型，用于填写Json匹配与校验
    }
  }
}
```

#### 请求示例

curl -H "BCS-ClusterID: {ClusterID}" -d "crd.json"  http://{Bcs-Domain}/v4/scheduler/mesos/customresourcedefinitions

#### 返回结果

http Code：200

### getcrd
#### 描述

* 通过名称获取指定自定义资源
* 获取当前注册的所有自定义资源

#### 请求地址

- /v4/scheduler/mesos/customresourcesdefinitions/{name}

#### 请求方式

- GTE

#### 请求参数

- name: 指定CustomResourceDefinition名称，可选

#### 请求示例

- curl -H "BCS-ClusterID: {ClusterID}" -d "crd.json"  http://{Bcs-Domain}/v4/scheduler/mesos/customresourcedefinitions/logcollector.bkbcs.tencent.com

#### 返回结果

```json
{
  "apiVersion": "v4",
  "kind": "CustomResourceDefinition",
  "metadata": {
    "name": "logcollector.bkbcs.tencent.com"
  },
  "spec": {
    "group": "bkbcs.tencent.com", 
    "versions": [
      {
        "name": "v1",
        "served": true,
        "storage": true
      }
    ],
    "scope": "Namespaced",
    "names": {
      "plural": "crontabs",
      "singular": "crontab",
      "kind": "CronTab"
    }
  }
}
```

### deletecrd
#### 描述

删除指定自定义资源

#### 请求地址
- /v4/scheduler/mesos/customresourcedefinitions/{name}

#### 请求方式
- DELETE

#### 请求示例

- curl -H "BCS-ClusterID: {ClusterID}" -X DELETE http://{Bcs-Domain}/v4/scheduler/mesos/customresourcedefinitions/logcollector.bkbcs.tencent.com

#### 返回结果

```json
```

### customresource

基于CRD接口完成资源定义后，可以通过API完成新资源的增删改查。

新资源API规则为: /v4/scheduler/mesos/customresources/{group}/{version}/{plural}

CRD案例说明：

```json
{
  "apiVersion": "v4",
  "kind": "CustomResourceDefinition",
  "metadata": {
    "name": "bklogconfig.bkbcs.tencent.com"
  },
  "spec": {
    "group": "bkbcs.tencent.com", 
    "versions": [
      {
        "name": "v2",
        "served": true,
        "storage": true
      }
    ],
    "scope": "Namespaced",
    "names": {
      "plural": "bklogconfigs",
      "singular": "bklogconfig",
      "kind": "BKLogConfig"
    }
  }
}
```


#### 生成请求接口

* 创建资源
  * Method：POST
  * URL：/v4/scheduler/mesos/customresources/bkbcs.tencent.com/v2/{namespace}/bklogconfigs
  * body: 资源json结构
* 查询资源
  * Method：GET
  * URL：/v4/scheduler/mesos/customresources/bkbcs.tencent.com/v2/{namespace}/bklogconfigs/{name}
* 删除资源
  * Method: DELETE
  * URL：/v4/scheduler/mesos/customresources/bkbcs.tencent.com/v2/{namespace}/bklogconfigs/{name}

#### 请求示例

BKLogConfig结构json文件bklog.json
```json
{
  "apiVersion": "bkbcs.tencent.com/v2",
  "kind": "BKLogConfig",
  "metadata": {
    "name": "myappconfig",
    "namespace": "global"
  },
  "spec": {
    "selector": {
      "app": "loadbalance"
    },
    "stdout": false,
    "logpath": "bcs-lb/logs/bcss-loadbalance.log",
    "dataid": 123456,
    "level": 3
  }
}
```

* 创建
curl -XPOST -H "BCS-ClusterID: {ClusterID}" -d bklog.json  http://{Bcs-Domain}/v4/scheduler/mesos/customresources/bkbcs.tencent.com/v2/global/bklogconfigs

* 查询
- NS下数据查询：curl -H "BCS-ClusterID: {ClusterID}"  http://{Bcs-Domain}/v4/scheduler/mesos/customresources/bkbcs.tencent.com/v2/namespaces/global/bklogconfigs
- 具体实例查询：curl -H "BCS-ClusterID: {ClusterID}"  http://{Bcs-Domain}/v4/scheduler/mesos/customresources/bkbcs.tencent.com/v2/namespaces/global/bklogconfigs/{name}
- 全量数据查询：具体实例查询：curl -H "BCS-ClusterID: {ClusterID}"  http://{Bcs-Domain}/v4/scheduler/mesos/customresources/bkbcs.tencent.com/v2/bklogconfigs

* 更新
curl -XPUT -H "BCS-ClusterID: {ClusterID}" -d bklog.json   http://{Bcs-Domain}/v4/scheduler/mesos/customresources/bkbcs.tencent.com/v2/{namespace}/bklogconfigs/{name}

* 删除
curl -XDELETE -H "BCS-ClusterID: {ClusterID}"  http://{Bcs-Domain}/v4/scheduler/mesos/customresources/bkbcs.tencent.com/v2/namespaces/{namespace}/bklogconfigs/{name}

### getDeploymentDef
#### 描述
获取deployment的定义数据

#### 请求地址
- /v4/scheduler/mesos/definition/deployment/{ns}/{name}

#### 请求方式
- Get

### getApplicationDef
#### 描述
获取application的定义数据:只有通过json定义创建的application才能获取，通过deployment生成的application不能获取起定义数据，只能获取deployment的定义

#### 请求地址
- /v4/scheduler/mesos/definition/application/{ns}/{name}

#### 请求方式
- Get

### sendapplicationSignal

#### 描述
对指定application下所有running状态的taskgroup发送信息；

#### 请求地址
- /v4/scheduler/mesos/namespaces/$(ns)/applications/$(name)/message

#### 请求方式
- POST

#### 请求参数
- 无

### 请求示例
- curl -X POST -d "{\"name\":\"test_deployment\",\"namespace\":\"test\",\"msgtype\": \"signal\",\"msgdata\":{\"processname\":\"myprocess\",\"signal\":12}}" -H'BCS-ClusterID:BCS-xxxxx-xxxxx' http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/test/applications/test_deployment/message
- 以上命令，将会对application(test.test_deployment)下所有running状态的pod容器中执行 /bin/sh -c "killall -12 myprocess"

### 返回结果
```json
{
    "result":true,
    "code":0,
    "message":"success",
    "data":null
}
```
- 补充：某些taskgroup由于异常没有发送信号成功，会在data中返回


### sendtaskgroupSignal

#### 描述
对指定taskgroup发送信息；

#### 请求地址
- /v4/scheduler/mesos/namespaces/$(ns)/applications/$(name)/taskgroups/$(taskgroup)/message

#### 请求方式
- POST

#### 请求参数
- 无

### 请求示例
- curl -X POST -d "{\"name\":\"test_deployment\",\"namespace\":\"test\",\"msgtype\": \"signal\",\"msgdata\":{\"processname\":\"myprocess\",\"signal\":12}}" -H'BCS-ClusterID:BCS-xxxxx-xxxxx' http://{Bcs-Domain}/v4/scheduler/mesos/namespaces/test/applications/test_deployment/taskgroups/0.test_deployment.test.10001.1543390043870110268/message
- 以上命令，将在0.test_deployment.test.10001.1543390043870110268的所有容器中执行 /bin/sh -c "killall -12 myprocess"

### 返回结果
```json
{
    "result":true,
    "code":0,
    "message":"success",
    "data":null
}
```

### sendCommand

#### 描述
对指定的application/deployment(可以指定部分taskgroup)发送命令

#### 请求地址
- /v4/scheduler/mesos/command/application/{ns}/{name}
- /v4/scheduler/mesos/command/deployment/{ns}/{name}
- 注意,必须保证请求地址url和请求data中commandTargetRef的数据保持一致

#### 请求方式
- POST

#### 请求参数
- 无

### 请求示例
- curl -X POST -d "{\"apiVersion\":\"v4\",\"kind\":\"Command\",\"spec\":{\"commandTargetRef\":{\"kind\":\"Application\",\"namespace\":\"test\",\"name\":\"testapp\"},\"command\":[\"ps -aux\"]}}" -H'BCS-ClusterID:BCS-xxxxx-xxxxx' http://{Bcs-Domain}/v4/scheduler/mesos/command/application/test/testapp
- 如果要对部分taskgroup执行命令，请求数据中使用taskgroups指定需要执行命令的taskgroup数组
- 具体参数定义请参考资源模板说明

### 返回结果
```json
{
    "result":true,
    "code":0,
    "message":"success",
    "data":"Application-test-testapp-1547437542873435715"
}
```

### getCommand

#### 描述
查询application/deployment的命令执行情况

#### 请求地址
- /v4/scheduler/mesos/command/application/{ns}/{name}
- /v4/scheduler/mesos/command/deployment/{ns}/{name}

#### 请求方式
- GET

#### 请求参数
- id

### 请求示例
- curl -H'BCS-ClusterID:BCS-xxxxx-xxxxx' http://{Bcs-Domain}/v4/scheduler/mesos/command/application/test/testapp?id="Application-test-testapp-1547437542873435715"

### 返回结果
```json
{
	"result": true,
	"code": 0,
	"message": "success",
	"data": {
		"id": "Application-test-testapp-1547437542873435715",
		"createTime": 1547437542,
		"spec": {
			"commandTargetRef": {
				"kind": "Application|Deployment",
				"name": "testapp",
				"namespace": "test"
			},
			"taskgroups": null,
			"command": ["ps"],
			"env": null,
			"user": "",
			"workingDir": "",
			"privileged": false,
			"reserveTime": 0
		},
		"status": {
			"taskgroups": [{
				"taskgroupId": "0.testapp.test.10001.1547025443000246455",
				"tasks": [{
					"taskId": "1547025443000246455.0.0.testapp.test.10001",
					"status": "staging|running|finish|failed",
					"message": "command in running",
					"commInspect": {
					    "exitCode": 0,
					    "stdout": ,
					    "stderr" ,
					}
				}]
			}]
		}
	}
}

- commInspect只有在finish和failed状态才有意义
- reserveTime为Command信息保留时间，未定义的情况下默认24*60*7(7天)

```

### deleteCommand

#### 描述
删除application/deployment的命令(已经下发的命令会继续执行)

#### 请求地址
- /v4/scheduler/mesos/command/application/{ns}/{name}
- /v4/scheduler/mesos/command/deployment/{ns}/{name}

#### 请求方式
- DELETE

#### 请求参数
- id

### 请求示例
- curl -X DELETE -H'BCS-ClusterID:BCS-xxxxx-xxxxx' http://{Bcs-Domain}/v4/scheduler/mesos/command/application/test/testapp?id="Application-test-testapp-1547437542873435715"

### 返回结果
```json
{
	"result": true,
	"code": 0,
	"message": "success",
	"data": nil
}
```

### createadmission
#### 描述
创建admission webhook

#### 请求地址
- /v4/scheduler/mesos/crd/namespaces/{ns}/admissionwebhook

#### 请求方式
- POST

#### 请求参数
-ns  //namespace

```json
{
  "apiVersion":"v4",
  "kind":"admissionwebhook",
  "metadata":{
    "name":"webhook-test",
    "namespace": "defaultGroup"
  },
  "resourcesRef": {
    "operation": "Create",
    "kind": "Application"
  },
  "admissionWebhooks":[
    {
      "name": "container-sidecar",
      "failurePolicy": "Ignore",
      "clientConfig": {
        "caBundle": "xxxxxxxxx",
        "namespace": "sidecar-webhook-namespace",
        "name": "sidecar-webhook-service"
      }
    },
    {
      "name": "container-images",
      "failurePolicy": "Fail",
      "clientConfig": {
        "caBundle": "xxxxxxxxx",
        "namespace": "image-webhook-namespace",
        "name": "image-webhook-service"
      }
    }
  ]
}
```

#### 请求示例
- curl -H "BCS-ClusterID: {ClusterID}" -d "admission.json"  http://{Bcs-Domain}/v4/scheduler/mesos/crd/namespaces/defaultGroup/admissionwebhook

#### 返回结果

```json
{
    "result":true,
    "code":0,
    "message":"success",
    "data":null
}
```

### updateadmission
#### 描述
更新admission webhook

#### 请求地址
- /v4/scheduler/mesos/crd/namespaces/{ns}/admissionwebhook

#### 请求方式
- PUT

#### 请求参数
-ns  //namespace

```json
{
  "apiVersion":"v4",
  "kind":"admissionwebhook",
  "metadata":{
    "name":"webhook-test",
    "namespace": "defaultGroup"
  },
  "resourcesRef": {
    "operation": "Create",
    "kind": "Application"
  },
  "admissionWebhooks":[
    {
      "name": "container-sidecar",
      "failurePolicy": "Ignore",
      "clientConfig": {
        "caBundle": "xxxxxxxxx",
        "namespace": "sidecar-webhook-namespace",
        "name": "sidecar-webhook-service"
      }
    },
    {
      "name": "container-images",
      "failurePolicy": "Fail",
      "clientConfig": {
        "caBundle": "xxxxxxxxx",
        "namespace": "image-webhook-namespace",
        "name": "image-webhook-service"
      }
    }
  ]
}
```

#### 请求示例
- curl -H "BCS-ClusterID: {ClusterID}" -X PUT -d "admission.json"  http://{Bcs-Domain}/v4/scheduler/mesos/crd/namespaces/defaultGroup/admissionwebhook

#### 返回结果

```json
{
    "result":true,
    "code":0,
    "message":"success",
    "data":null
}
```

### getadmission
#### 描述
获取admission webhook

#### 请求地址
- /v4/scheduler/mesos/crd/namespaces/{ns}/admissionwebhook/{name}

#### 请求方式
- GET

#### 请求参数
-ns  //namespace
-name //name

#### 请求示例
- curl -H "BCS-ClusterID: {ClusterID}" -X GET http://{Bcs-Domain}/v4/scheduler/mesos/crd/namespaces/defaultGroup/admissionwebhook/webhook-test

#### 返回结果

```json
{
    "code": 0,
    "data": {
        "ResourcesRef": {
            "Kind": "Application",
            "Operation": "Create"
        },
        "admissionWebhooks": [
            {
                "ClientConfig": {
                    "CaBundle": "xxxxxxxxx",
                    "Name": "sidecar-webhook-service",
                    "Namespace": "sidecar-webhook-namespace"
                },
                "FailurePolicy": "Ignore",
                "Name": "container-sidecar"
            },
            {
                "ClientConfig": {
                    "CaBundle": "xxxxxxxxx",
                    "Name": "image-webhook-service",
                    "Namespace": "image-webhook-namespace"
                },
                "FailurePolicy": "Fail",
                "Name": "container-images"
            }
        ],
        "apiVersion": "v4",
        "kind": "admissionwebhook",
        "metadata": {
            "creationTimestamp": "0001-01-01T00:00:00Z",
            "name": "webhook-test",
            "namespace": "defaultGroup"
        }
    },
    "message": "success",
    "result": true
}
```

### deleteadmission
#### 描述
删除admission webhook

#### 请求地址
- /v4/scheduler/mesos/crd/namespaces/{ns}/admissionwebhook/{name}

#### 请求方式
- GET

#### 请求参数
-ns  //namespace
-name //name

#### 请求示例
- curl -H "BCS-ClusterID: {ClusterID}" -X DELETE http://{Bcs-Domain}/v4/scheduler/mesos/crd/namespaces/defaultGroup/admissionwebhook/webhook-test

#### 返回结果

```json
{
    "code": 0,
    "data": null,
    "message": "success",
    "result": true
}
```