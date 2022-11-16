### 描述
sidecar Reload文件协议。

#### reload 文件内容示例
```json
{
    "version": "v1", // reload文件协议版本，不同版本的 sidecar 的 reload文件协议可能不同
    "timestamp": "2022-07-29 10:52:16", // 通知reload的时间
    "app_id": 1,  // 应用ID
    "release_id": 1, // 版本ID
    "root_directory": "/data/bscp/workspace/bk-bscp/fileReleaseV1/2/1/1/configItems", // 存放配置文件的根目录，需要注意不同版本的配置文件的根目录是不同的
    "config_item": [  // 配置相对于 root_directory 的子路径，子路径是通过 config_item path + name 来生成的
        "/etc/mysql.yaml",
        "/etc/redis.yaml"
    ]
}
```
