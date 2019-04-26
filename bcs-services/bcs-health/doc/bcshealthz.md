# BCS 集群健康状态检查接口

## 说明
该接口用于查询整个bcs集群的健康状态。需要bcs-clusterkeeper有配置集群整个部署的静态信息。

## 详情
- 方法： GET
- 路径： /bcshealth/v1/bcshealthz
- 请求体： 无
- 请求示例：
```shell
curl http://127.0.0.1:8001/bcshealth/v1/bcshealthz
```

- 结果示例：
```json
{
    "code": 0,
    "ok": true,
    "message": "success",
    "data": {
        "healthy": "unknown",
        "platform_status": {
            "healthy": "unknown",
            "status": {
                "apiserver": {
                    "status": "unknown",
                    "message": [
                        "online endpoints[10.*.*.*] is different from the deployed"
                    ]
                },
                "clusterkeeper": {
                    "status": "healthy",
                    "message": null
                },
                "discovery": {
                    "status": "unhealthy",
                    "message": [
                        "lost 1 instance"
                    ]
                },
                "health": {
                    "status": "healthy",
                    "message": null
                },
                "metricservice": {
                    "status": "healthy",
                    "message": null
                },
                "netservice": {
                    "status": "healthy",
                    "message": null
                },
                "route": {
                    "status": "healthy",
                    "message": null
                },
                "storage": {
                    "status": "healthy",
                    "message": null
                }
            }
        },
        "clusters_status": [
            {
                "type": "mesos",
                "cluster_id": "BCS-MESOS-10032",
                "healthy": "unknown",
                "status": {}
            },
            {
                "type": "mesos",
                "cluster_id": "BCS-MESOSSELFTEST-10001",
                "healthy": "unknown",
                "status": {
                    "health": {
                        "status": "unknown",
                        "message": [
                            "retrieval status failed, err: zk: node does not exist"
                        ]
                    },
                    "mesosdatawatch": {
                        "status": "unknown",
                        "message": [
                            "retrieval status failed, err: zk: node does not exist"
                        ]
                    },
                    "scheduler": {
                        "status": "unknown",
                        "message": [
                            "retrieval status failed, err: zk: node does not exist"
                        ]
                    }
                }
            }
        ]
    }
}
```