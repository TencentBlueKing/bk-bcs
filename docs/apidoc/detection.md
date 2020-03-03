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

## url prefix
### bcs-api
http://{ipaddr}:{port}/bcsapi

示例：
curl -X GET http://127.0.0.1:8081/bcsapi/v4/detection/detectionpods

## api
### api list

* detection
  - [**get all detection pods**](#getDetectionPods)


### getDetectionPods
#### 描述
get all detection pods

#### 请求地址
- /v4/detection/detectionpods

#### 请求方式
- GET

#### 请求示例
curl -X GET http://127.0.0.1:8081/bcsapi/v4/detection/detectionpods

#### 返回结果

```json
{
    "result":true,
    "code":0,
    "message":"success",
    "data":[
        {
            "Ip":"172.32.0.1",
            "Idc":"深圳移动荔景DC"
        },
        {
            "Ip":"172.32.0.1",
            "Idc":"深圳移动荔景DC"
        },
        {
            "Ip":"172.32.0.1",
            "Idc":"深圳移动荔景DC"
        },
        {
            "Ip":"172.32.0.2",
            "Idc":"深圳移动光明DC"
        },
        {
            "Ip":"172.32.0.1",
            "Idc":"深圳移动光明DC"
        },
        {
            "Ip":"172.32.0.3",
            "Idc":"深圳移动光明DC"
        }
    ]
}
```