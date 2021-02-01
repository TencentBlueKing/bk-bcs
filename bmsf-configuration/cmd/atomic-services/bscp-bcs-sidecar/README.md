# BSCP BCS SIDECAR
## 环境变量说明

|                  环境变量名                              |                                                                    示例                                                                         |                       备注                      |
| :------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------: |
| BSCP_BCSSIDECAR_FILE_RELOAD_MODE                         | true                                                                                                                                            | 文件通知reload模式                              |
| BSCP_BCSSIDECAR_FILE_RELOAD_NAME                         | my.reload.file                                                                                                                                  | 文件通知reload模式的通知文件名，默认BSCP.reload |
| BSCP_BCSSIDECAR_GW_HOSTNAME                              | gw.bkbscp.com                                                                                                                                   | 网关服务域名                                    |
| BSCP_BCSSIDECAR_GW_PORT                                  | 8080                                                                                                                                            | 网关服务端口                                    |
| BSCP_BCSSIDECAR_CONNSERVER_HOSTNAME                      | conn.bkbscp.com                                                                                                                                 | 链接会话服务域名                                |
| BSCP_BCSSIDECAR_CONNSERVER_PORT                          | 9516                                                                                                                                            | 链接会话服务端口                                |
| BSCP_BCSSIDECAR_CONNSERVER_DIAL_TIMEOUT                  | 3s                                                                                                                                              | 链接会话服务建立链接超时时间                    |
| BSCP_BCSSIDECAR_CONNSERVER_CALL_TIMEOUT                  | 3s                                                                                                                                              | 链接会话服务请求超时时间                        |
| BSCP_BCSSIDECAR_APPINFO_IP_ETH                           | eth1                                                                                                                                            | 网卡名称，用于获取本地IP信息作为Sidecar身份标识 |
| BSCP_BCSSIDECAR_APPINFO_IP                               | 127.0.0.1                                                                                                                                       | IP地址, xxx.xxx.xxx.xxx                         |
| BSCP_BCSSIDECAR_APPINFO_MOD                              | [{"biz_id":"biz01", "app_id":"XXXXXXXXXXXXXXX", "cloud_id":"0", "labels":{"k1":"v1"}, "path":"/data/app/etc/"}]                                 | 当前Sidecar所属业务信息全量设置，支持多模块     |
| BSCP_BCSSIDECAR_EFFECT_FILE_CACHE_PATH                   | ./bscp-cache/fcache/                                                                                                                            | 生效信息文件缓存路径                            |
| BSCP_BCSSIDECAR_CONTENT_CACHE_PATH                       | ./bscp-cache/ccache/                                                                                                                            | 内容缓存路径                                    |
| BSCP_BCSSIDECAR_LINK_CONTENT_CACHE_PATH                  | ./bscp-cache/lcache/                                                                                                                            | link内容缓存路径                                |
| BSCP_BCSSIDECAR_CONTENT_EXPIRED_CACHE_PATH               | /tmp/                                                                                                                                           | 过期内容缓存回收路径                            |
| BSCP_BCSSIDECAR_CONTENT_CACHE_MAX_DISK_USAGE_RATE        | 10                                                                                                                                              | 内容缓存最大磁盘占用比例                        |
| BSCP_BCSSIDECAR_CONTENT_CACHE_EXPIRATION                 | 168h                                                                                                                                            | 内容缓存过期时间                                |
| BSCP_BCSSIDECAR_CONTENT_CACHE_DISK_USAGE_CHECK_INTERVAL  | 15m                                                                                                                                             | 内容缓存磁盘占用检查间隔                        |
| BSCP_BCSSIDECAR_DOWNLOAD_PER_FILE_CONCURRENT             | 1                                                                                                                                               | 内容下载单文件最大并发数                        |
| BSCP_BCSSIDECAR_DOWNLOAD_PER_FILE_LIMIT_BYTES            | 1024*1024*1                                                                                                                                     | 内容下载单文件每秒限速, 单位:字节/秒，默认1MB/s |

### 部署变量设置建议

必须设置的环境变量,

* BSCP_BCSSIDECAR_APPINFO_MOD：sidecar所属业务信息全量设置

## 本地配置文件示例

```
sidecar:
    fileReloadMode: false
    fileReloadName: my.reload.file

connserver:
    port: 9516

appinfo:
    ipeth: eth1
    ip: 127.0.0.1

    mod:
        - biz_id: biz01
          app_id: XXXXXXXXXXXXXXX
          cloud_id: dev
          path: /data/app1/etc
          labels:
              k1: v1
              k2: v2

        - biz_id: biz02
          app_id: XXXXXXXXXXXXXXX
          cloud_id: dev
          path: /data/app2/etc
          labels:
              k1: v1
              k2: v2

logger:
    level: 3
    maxnum: 5
    maxsize: 200
```
