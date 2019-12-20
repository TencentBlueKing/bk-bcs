# 证书说明

tls证书管理目录，需要根据需要自行生成。相关证书和名称建议如下：

* bcs-ca.crt
* bcs-server.key
* bcs-server.crt
* bcs-client.key
* bcs-client.crt

key文件为了安全，建议进行加密。完成加密后，针对加密的密码建议编译时注入。注入的环境变量参考`scripts/env.sh`。

该证书在制作以下镜像时默认注入：

* bcs-k8s-driver
* bcs-k8s-datawatch
* bcs-k8s-csi-tencentcloud
* bcs-k8s-custom-scheduler


