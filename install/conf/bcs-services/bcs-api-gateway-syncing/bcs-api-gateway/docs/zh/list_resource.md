### 通过资源池ID获取资源列表

这是一个描述

### 输入参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| resource_pool_id         | string       | 是     | resourcepool-xxx     |


### 调用示例
```python
from bkapi.bcs_api_gateway.shortcuts import get_client_by_request

client = get_client_by_request(request)
result = client.api.api_test({}, path_params={}, headers=None, verify=True)
```

### 响应示例
```json
{
    "code": 0,
    "message": "OK",
    "result": true,
    "data": [
        {
            "id": "resource-xx",
            "innerIP": "10.0.0.0",
            "innerIPv6": "",
            "resourceType": "CVM",
            "resourceProvider": "yunti",
            "poolID": "resourcepool-xx",
            "labels": {
                "instanceID": "ins-4mhs1lim",
                "instanceType": "IT5.21XLARGE320",
                "subnet": "",
                "vpc": "vpc-faovacc5",
                "zone": "ap-nanjing-2"
            },
            "annotations": {},
            "createTime": "1658917665",
            "updateTime": "1659508895",
            "status": {
                "phase": "IDLE",
                "updateTime": "1659508895",
                "consumeOrderID": "",
                "devicePoolID": "",
                "clusterID": "",
                "returnOrderID": ""
            }
        }
    ]
}
```

### 响应参数说明
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|  id           |   string         |    资源唯一ID              |
|  innerIP           |   string         |      内网IP地址            |
|  innerIPv6           |   string         |    内网IPv6地址，目前为空              |
|  resourceType           |   string         |    资源类型，目前有CVM, BareMetal              |
|  resourceProvider           |   string         |    资源提供方，目前有yunti, bk_cmdb             |
|  poolID           |   string         |    所属资源池ID              |
|  labels           |  map[string]string         |    不同的资源有差异，公共的有  zone， instanceTyped等          |