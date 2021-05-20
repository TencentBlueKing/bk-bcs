# bcs apiserver v4 http api for bcs ingress

## introduction

Mesos集群中clbingress是一种自定义资源，相关接口本质上就是api-scheduler.md中说明的**自定义资源(CustomResource)**接口。本文档在**自定义资源(CustomResource)**接口基础上，举例做进一步说明

## prerequisites

* 确保集群中已经安装好对应的clb-controller，并启动成功
* 确保bcs-mesos-driver版本在1.15以上

## 接口示例

### 创建ClbIngress资源

```shell
curl -X POST -H "BCS-ClusterID: {ClusterID}" -d @ingress.json http://{Bcs-Domain}/v4/scheduler/mesos/customresourcedefinitions/clb.bmsf.tencent.com/v1/namespaces/test-ns/clbingresses
```

ingress.json

```json
{
    "apiVersion": "clb.bmsf.tencent.com/v1",
    "kind": "ClbIngress",
    "metadata": {
        "labels": {
            "bmsf.tencent.com/clbname": "{{ clb实例名称 }}"
        },
        "name": "test-ingress",
        "namespace": "test-ns"
    },
    "spec": {
        "tcp": [
            {
                "serviceName": "tcptest", // 需要关联clb的service名字
                "namespace": "test", // 需要关联clb的service的命名空间
                "clbPort": 10080, // clb上面监听的端口
                "servicePort": 10081, // service里面的端口，servcie端口需要是tcp协议的
                "sessionTime": 30,
                "lbPolicy": {
                    "strategy": "wrr" // 可选[wrr, least_conn]
                },
                // 设置健康检查策略，默认开启，timeout默认2（单位：秒），intervalTime默认3（单位：秒），timeout必须小于intervalTime，healthNum默认3（单位：次），unhealthNum默认3（单位：次）
                "healthCheck": {
                    "enabled": true, // 是否开启健康检查， true（开启），false（关闭）
                    "timeout": 2, // 健康检查的响应超时时间，可选值：2~60，默认值：2，单位：秒。响应超时时间要小于检查间隔时间。
                    "intervalTime": 5, //  健康检查探测间隔时间，默认值：5，可选值：5~300，单位：秒。
                    "healthNum": 3, // 健康阈值，默认值：3，表示当连续探测三次健康则表示该转发正常，可选值：2~10，单位：次。
                    "unHealthNum": 3 // 不健康阈值，默认值：3，表示当连续探测三次不健康则表示该转发异常，可选值：2~10，单位：次。
                }
            }
        ],
        "udp": [],
        "http": [
            {
                "host": "www.qq.com", // http监听器转发规则的域名
                "path": "/", // http监听器转发规则的url
                "serviceName": "tcptest", // 需要关联clb的service名字
                "namespace": "test", // 需要关联clb的service的命名空间
                "clbPort": 10080, // clb上面监听的端口
                "servicePort": 10081, // service里面的端口，servcie端口需要是http协议的
                "sessionTime": 30,
                "lbPolicy": {
                    "strategy": "wrr" // 可选[wrr, least_conn]
                },
                // 设置健康检查策略，默认开启，timeout默认2（单位：秒），intervalTime默认3（单位：秒），timeout必须小于intervalTime，healthNum默认3（单位：次），unhealthNum默认3（单位：次）
                "healthCheck": {
                    "enabled": true, // 是否开启健康检查， true（开启），false（关闭）
                    "timeout": 2, // 健康检查的响应超时时间，可选值：2~60，默认值：2，单位：秒。响应超时时间要小于检查间隔时间。
                    "intervalTime": 5, //  健康检查探测间隔时间，默认值：5，可选值：5~300，单位：秒。
                    "healthNum": 3, // 健康阈值，默认值：3，表示当连续探测三次健康则表示该转发正常，可选值：2~10，单位：次。
                    "unHealthNum": 3, // 不健康阈值，默认值：3，表示当连续探测三次不健康则表示该转发异常，可选值：2~10，单位：次。
                    // 应用型负载均衡监听器转发规则的健康状态码。可选值：1~31，默认31。
                    // 1表示探测后返回值 1xx 表示健康，2表示返回 2xx 表示健康，4表示返回 3xx 表示健康，8表示返回 4xx 表示健康，16表示返回 5xx 表示健康。
                    // 若希望多种码都表示健康，则将相应的值相加。
                    "httpCode": 31
                }
            }
        ],
        "https": [
            {
                "host": "www.qq.com", // http监听器转发规则的域名
                "path": "/", // http监听器转发规则的url
                "serviceName": "tcptest", // 需要关联clb的service名字
                "namespace": "test", // 需要关联clb的service的命名空间
                "clbPort": 10080, // clb上面监听的端口
                "servicePort": 10081, // service里面的端口，servcie端口需要是http协议的
                "sessionTime": 30,
                "lbPolicy": {
                    "strategy": "wrr" // 可选[wrr, least_conn]
                },
                // 设置健康检查策略，默认开启，timeout默认2（单位：秒），intervalTime默认3（单位：秒），timeout必须小于intervalTime，healthNum默认3（单位：次），unhealthNum默认3（单位：次）
                "healthCheck": {
                    "enabled": true, // 是否开启健康检查， true（开启），false（关闭）
                    "timeout": 2, // 健康检查的响应超时时间，可选值：2~60，默认值：2，单位：秒。响应超时时间要小于检查间隔时间。
                    "intervalTime": 5, //  健康检查探测间隔时间，默认值：5，可选值：5~300，单位：秒。
                    "healthNum": 3, // 健康阈值，默认值：3，表示当连续探测三次健康则表示该转发正常，可选值：2~10，单位：次。
                    "unHealthNum": 3, // 不健康阈值，默认值：3，表示当连续探测三次不健康则表示该转发异常，可选值：2~10，单位：次。
                    // 应用型负载均衡监听器转发规则的健康状态码。可选值：1~31，默认31。
                    // 1表示探测后返回值 1xx 表示健康，2表示返回 2xx 表示健康，4表示返回 3xx 表示健康，8表示返回 4xx 表示健康，16表示返回 5xx 表示健康。
                    // 若希望多种码都表示健康，则将相应的值相加。
                    "httpCode": 31
                },
                "tls": {
                    // HTTPS 协议的认证类型，unidirectional：单向认证，mutual：双向认证
                    "mode": "unidirectional",
                    // 在腾讯云控制台创建的服务端证书的 ID，HTTPS 监听器如果不填写此项则必须上传证书，包括 certServerContent，certServerKey，certServerName。
                    "certId": "xxxxxx",
                    // 客户端证书的 ID，如果 mode=mutual，监听器如果不填写此项则必须上传客户端证书，包括 certClientCaName，certCilentCaContent
                    "certCaId": "xxxxx"
                }
            }
        ]
    }
}
```

### 更新ClbIngress

```shell
curl -X POST -H "Content-Type: application/json" -H "BCS-ClusterID: {ClusterID}" -d @ingress.json http://{Bcs-Domain}/v4/scheduler/mesos/customresources/clb.bmsf.tencent.com/v1/namespaces/test-ns/clbingresses/test-ingress
```

ingress.json 略

```json
...
```

### 查询ClbIngress

```shell
curl -X GET -H "Content-Type: application/json" -H "BCS-ClusterID: {ClusterID}" http://{Bcs-Domain}/v4/scheduler/mesos/customresources/clb.bmsf.tencent.com/v1/namespaces/test-ns/clbingresses/test-ingress
```

### 删除ClbIngress

```shell
curl -X GET -H "Content-Type: application/json" -H "BCS-ClusterID: {ClusterID}" http://{Bcs-Domain}/v4/scheduler/mesos/customresources/clb.bmsf.tencent.com/v1/namespaces/test-ns/clbingresses/test-ingress
```

### 查询ClbIngress列表

```shell
curl -X GET -H "Content-Type: application/json" -H "BCS-ClusterID: {ClusterID}" http://{Bcs-Domain}/v4/scheduler/mesos/customresources/clb.bmsf.tencent.com/v1/namespaces/test-ns/clbingresses
```
