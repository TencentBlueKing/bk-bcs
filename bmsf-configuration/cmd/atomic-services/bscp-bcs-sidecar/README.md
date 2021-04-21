# BSCP BCS SIDECAR

## 环境变量说明

|                  环境变量名                              |                                                                    示例                                                                         |                       备注                           |
| :------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------: |
| BSCP_BCSSIDECAR_CONNSERVER_HOSTNAME                      | {nodeIP}                                                                                                                                        | 主机或Node节点IP                                     |
| BSCP_BCSSIDECAR_APPINFO_MOD                              | [{"biz_id":"biz01", "app_id":"XXXXXXXXXXXXXXX", "cloud_id":"0", "labels":{"k1":"v1"}, "path":"/data/app/etc/"}]                                 | 当前Sidecar所属业务信息全量设置，支持多模块          |
| BSCP_BCSSIDECAR_FILE_RELOAD_MODE                         | true                                                                                                                                            | 文件通知reload模式                                   |
| BSCP_BCSSIDECAR_FILE_RELOAD_NAME                         | my.reload.file                                                                                                                                  | 文件通知reload模式的通知文件名，默认BSCP.reload      |
| BSCP_BCSSIDECAR_ENABLE_DELETE_CONFIG                     | false                                                                                                                                           | 是否主动同步删除已经Delete状态的Config               |
| BSCP_BCSSIDECAR_READY_PULL_CONFIGS                       | true                                                                                                                                            | 是否立即同步配置而不等待本地Instance动态接口注入信息 |
| BSCP_BCSSIDECAR_INS_OPEN                                 | false                                                                                                                                           | 是否开启本地Instance动态注入信息                     |
| BSCP_BCSSIDECAR_INS_HTTP_ENDPOINT_PORT                   | 39610                                                                                                                                           | 本地Instance动态注入服务HTTP监听端口                 |
| BSCP_BCSSIDECAR_INS_GRPC_ENDPOINT_PORT                   | 39611                                                                                                                                           | 本地Instance动态注入服务gRPC监听端口                 |

### 部署变量设置建议

必须设置的环境变量,

* BSCP_BCSSIDECAR_APPINFO_MOD：sidecar所属业务信息全量设置

## 本地配置文件示例

```
sidecar:
    fileReloadMode: false
    fileReloadName: my.reload.file

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
