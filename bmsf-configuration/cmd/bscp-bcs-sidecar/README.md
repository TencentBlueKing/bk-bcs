# BSCP BCS SIDECAR
## 环境变量说明

|                  环境变量名                     |                                                                    示例                                                                           |                       备注                       |
| :---------------------------------------------- | -----------------------------------------------------------------------------------------------------------------------------------------------:  | :----------------------------------------------: |
| BSCP_BCSSIDECAR_PULL_CFG_INTERVAL               | 60s                                                                                                                                               | 自动同步最新配置版本间隔                         |
| BSCP_BCSSIDECAR_SYNC_CFGSETLIST_INTERVAL        | 10m                                                                                                                                               | 自动同步配置集合列表间隔                         |
| BSCP_BCSSIDECAR_REPORT_INFO_INTERVAL            | 10m                                                                                                                                               | 自动上报本地信息间隔                             |
| BSCP_BCSSIDECAR_ACCESS_INTERVAL                 | 3s                                                                                                                                                | 接入链接会话服务等待间隔                         |
| BSCP_BCSSIDECAR_SESSION_TIMEOUT                 | 21s                                                                                                                                               | 链接会话超时时间                                 |
| BSCP_BCSSIDECAR_SESSION_COEFFICIENT             | 2                                                                                                                                                 | 链接会话超时时间系数                             |
| BSCP_BCSSIDECAR_FILE_RELOAD_MODE                | true                                                                                                                                              | 文件通知reload模式                               |
| BSCP_BCSSIDECAR_FILE_RELOAD_FNAME               | my.reload.file                                                                                                                                    | 文件通知reload模式的通知文件名，默认BSCP.reload  |
| BSCP_BCSSIDECAR_CFGSETLIST_SIZE                 | 1000                                                                                                                                              | 拉取最大配置集合列表大小                         |
| BSCP_BCSSIDECAR_HANDLER_CH_SIZE                 | 10000                                                                                                                                             | main处理协程管道大小                             |
| BSCP_BCSSIDECAR_HANDLER_CH_TIMEOUT              | 1s                                                                                                                                                | main处理协程管道超时时间                         |
| BSCP_BCSSIDECAR_CFG_HANDLER_CH_SIZE             | 10000                                                                                                                                             | 配置处理协程管道大小                             |
| BSCP_BCSSIDECAR_CFG_HANDLER_CH_TIMEOUT          | 1s                                                                                                                                                | 配置处理协程管道超时时间                         |
| BSCP_BCSSIDECAR_CONNSERVER_HOSTNAME             | conn.bscp.bk.com                                                                                                                                  | 链接会话服务域名                                 |
| BSCP_BCSSIDECAR_CONNSERVER_PORT                 | 9516                                                                                                                                              | 链接会话服务端口                                 |
| BSCP_BCSSIDECAR_CONNSERVER_DIAL_TIMEOUT         | 3s                                                                                                                                                | 链接会话服务建立链接超时时间                     |
| BSCP_BCSSIDECAR_CONNSERVER_CALL_TIMEOUT         | 3s                                                                                                                                                | 链接会话服务请求超时时间                         |
| BSCP_BCSSIDECAR_APPINFO_IP_ETH                  | eth1                                                                                                                                              | 网卡名称，用于获取本地IP信息作为Sidecar身份标识  |
| BSCP_BCSSIDECAR_APPINFO_IP                      | 127.0.0.1                                                                                                                                         | IP地址, xxx.xxx.xxx.xxx                          |
| BSCP_BCSSIDECAR_APPINFO_MOD                     | [{"business":"mybusiness", "app":"myapp", "cluster":"cluster-01", "zone":"zone-01", "dc":"dc01", "labels":{"k1":"v1"}, "path":"/data/app/etc/"}]  | 当前Sidecar所属业务信息全量设置，支持多模块      |
| BSCP_BCSSIDECAR_APPCFG_PATH                     | ./app                                                                                                                                             | 应用配置路径                                     |
| BSCP_BCSSIDECAR_APPINFO_BUSINESS                | mybusiness                                                                                                                                        | 当前Sidecar所属业务名                            |
| BSCP_BCSSIDECAR_APPINFO_APP                     | myapp                                                                                                                                             | 当前Sidecar所属应用分组名                        |
| BSCP_BCSSIDECAR_APPINFO_CLUSTER                 | clustername                                                                                                                                       | 当前Sidecar所属应用分组的集群名                  |
| BSCP_BCSSIDECAR_APPINFO_ZONE                    | zonename                                                                                                                                          | 当前Sidecar所属应用分组的可用区名                |
| BSCP_BCSSIDECAR_APPINFO_DC                      | sz-idc-01                                                                                                                                         | 当前Sidecar所在物理机房表示，用作内网IP命名空间  |
| BSCP_BCSSIDECAR_APPINFO_LABELS                  | {"k1": "v1"}                                                                                                                                      | 当前Sidecar附带labels, json字符串KV格式          |
| BSCP_BCSSIDECAR_FILE_CACHE_PATH                 | ./cache/fcache/                                                                                                                                   | 生效信息文件缓存路径                             |
| BSCP_BCSSIDECAR_CONTENT_CACHE_PATH              | ./cache/ccache/                                                                                                                                   | 内容缓存路径                                     |
| BSCP_BCSSIDECAR_CONTENT_CACHE_EXPIRATION        | 168h                                                                                                                                              | 内容缓存过期时间                                 |
| BSCP_BCSSIDECAR_CONTENT_EXPCACHE_PATH           | /tmp/                                                                                                                                             | 过期内容缓存回收路径                             |
| BSCP_BCSSIDECAR_CONTENT_MCACHE_SIZE             | 1000                                                                                                                                              | 内存内容缓存大小                                 |
| BSCP_BCSSIDECAR_CONTENT_MCACHE_EXPIRATION       | 10m                                                                                                                                               | 内存内容缓存过期时间                             |
| BSCP_BCSSIDECAR_CONTENT_CACHE_PURGE_INTERVAL    | 30m                                                                                                                                               | 内容缓存清理间隔                                 |
| BSCP_BCSSIDECAR_LOG_DIR                         | ./log                                                                                                                                             | 日志目录                                         |
| BSCP_BCSSIDECAR_LOG_MAXSIZE                     | 200                                                                                                                                               | 日志单文件大小上限                               |
| BSCP_BCSSIDECAR_LOG_MAXNUM                      | 5                                                                                                                                                 | 日志文件个数上限                                 |
| BSCP_BCSSIDECAR_LOG_LEVEL                       | 5                                                                                                                                                 | 日志级别, 0-5                                    |
| BSCP_BCSSIDECAR_LOG_STDERR                      | 0                                                                                                                                                 | 是否将错误信息输出到标准出错而不是文件中         |
| BSCP_BCSSIDECAR_LOG_ALSOSTDERR                  | 0                                                                                                                                                 | 是否将错误信息同时输出到标准出错和文件中         |
| BSCP_BCSSIDECAR_LOG_STDERR_THRESHOLD            | 2                                                                                                                                                 | 达到或高于指定级别日志信息将输出到标准出错中     |

### 部署变量设置建议

必须设置的环境变量,

`单模块模式`, 通过对应环境变量设置单一模块名称,

* BSCP_BCSSIDECAR_APPCFG_PATH: 应用配置生效路径
* BSCP_BCSSIDECAR_APPINFO_BUSINESS：sidecar所属业务名
* BSCP_BCSSIDECAR_APPINFO_APP：sidecar所属应用名
* BSCP_BCSSIDECAR_APPINFO_CLUSTER: sidecar所属集群名称
* BSCP_BCSSIDECAR_APPINFO_ZONE: Sidecar所属大区名称
* BSCP_BCSSIDECAR_APPINFO_DC: Sidecar所在物理机房表示，用作内网IP命名空间

或

`多模块模式`，即所有信息在一个环境变量中用json格式标识:
* BSCP_BCSSIDECAR_APPINFO_MOD：sidecar所属业务信息全量设置


标签设置,

* BSCP_BCSSIDECAR_APPINFO_LABELS: 当前Sidecar附带labels, json字符串KV格式{"k1": "v1"}, 可用于策略控制

## 本地配置文件示例

```
sidecar:
    pullConfigInterval: 60s
    reportInfoInterval: 600s
    syncConfigsetListInterval: 60s
    sessionTimeout: 10s
    fileReloadMode: false
    fileReloadFName: my.reload.file

connserver:
    port: 9516
    dialtimeout: 3s
    calltimeout: 3s

appinfo:
    ipeth: eth1
    ip: 127.0.0.1

    mod:
        - business: mybusiness
          app: myapp
          cluster: cluster-01
          zone: zone-01
          dc: dc01
          path: /data/app/etc
          labels:
              k1: v1
              k2: v2

        - business: myAnotherBusiness
          app: myAnotherApp
          cluster: cluster-01
          zone: zone-01
          dc: dc01
          path: /data/app/etc
          labels:
              k1: v1
              k2: v2

logger:
    level: 3
    maxnum: 5
    maxsize: 200
```
