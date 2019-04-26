# Secret数据定义

**数据结构**json

``` json
{
    "apiVersion":"v4",
    "kind":"secret",
    "metadata":{
        "name":"template-secret",
        "namespace":"defaultGroup",
        "labels":{
        }
    },
    "type": "",
    "datas":{
        "first-secret": {
            "path": "/path/to/store/in/vault",
            "content":"Y29uZmlnIGNvbnRleHQ="
        },
        "second-secret": {
            "path": "/path/to/store/in/vault",
            "content":"Y29uZmlnIGNvbnRleHQ="
        }
    }
}
```

字段含义：

* name：具体secret名字
* content：secret内容

**bcs application**数据结构

```json
"secrets": [
    {
        "secretName": "mySecret",
        "items": [
            {
                "type": "env",
                "dataKey": "abc",
                "keyOrPath": "SRECT_ENV"
            },
            {
                "type": "file",
                "dataKey": "abc",
                "keyOrPath": "/data/container/path/myfile.conf",
                "readOnly": false,
                "user": "user"
            }
        ]
    }
]
```

## 工作机制和流程

数据**流转流程**

* bcs-client create --type secret --name template-secret --clusterid saldfkaslkdfj
  * --path vault存储path，必填
  * --file key:/path/to/file.conf 指定文件，文件大小不能超过100k
  * --from-files dir 指定文件夹，根据文件名做key，总大小不能超过2M
  * --content key:content 直接指定具体内容
* bcs-apiserver（正常存入vault？）
  * DELETE: 根据规则删除secret
  * POST：存储vault中
  * GET: 查询secret信息
* bcs-route流程
  * 查询secret信息，并入taskgroup请求，转发给scheduler

**相关调整**工作

* apiserver增加secret接口
  * post
  * get
  * delete
* bcs-route数据解析
  * 捕获secret字段，提取secret具体内容
  * 转发给bcs-scheduler
* bcs-client增加secret子命令，参数如下
  * --path vault存储path，必填
  * --file key:/path/to/file.conf 指定文件，文件大小不能超过100k
  * --from-files dir 指定文件夹，根据文件名做key，总大小不能超过2M
  * --content key:content 直接指定具体内容
