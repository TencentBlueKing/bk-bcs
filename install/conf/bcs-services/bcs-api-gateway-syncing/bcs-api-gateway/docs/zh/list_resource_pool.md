### 描述

资源池列表


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
            "id": "resourcepool-xx",
            "name": "BCS_Test",
            "comment": "",
            "labels": {},
            "annotations": {},
            "createTime": "1658916900",
            "updateTime": "1658916900"
        }
    ]
}
```

### 响应参数说明
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|  id            |   string         |       资源池唯一ID                       |
|  name            |   string         |       资源池名称                       |
|  comment            |   string         |       资源池备注                       |
|  labels            |   map[string]string         |      资源池labels，预留扩展用，目前空                      |