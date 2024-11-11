## 描述

存储所有纳管、创建的 k8s 集群信息，包括 k8s 证书、token、node、nodegroup、云凭证等。

## cluster manager 接口列表和参数

可以参考 swagger 文档查看相关接口以及参数等信息：[clustermanager.swagger.json](https://github.com/TencentBlueKing/bk-bcs/blob/master/bcs-services/bcs-cluster-manager/api/clustermanager/clustermanager.swagger.json)

## 接口调用示例

以 Cluster 相关的接口为例

- 查询集群信息
    ```bash
    curl -X GET \
    -H 'X-Bkapi-Authorization: {"bk_app_code": "bk_apigw_test", "bk_app_secret": "***"}' \
    'http://bcs-api.bkdomain/bcsapi/v4/clustermanager/v1/cluster/BCS-K8S-00000'
    ```
- 删除集群
    ```bash
    curl -X DELETE \
    -H 'X-Bkapi-Authorization: {"bk_app_code": "bk_apigw_test", "bk_app_secret": "***"}' \
    'http://bcs-api.bkdomain/bcsapi/v4/clustermanager/v1/cluster/BCS-K8S-00000'
    ```
- 更新集群信息
    ```bash
    curl -X PUT \
    -H 'X-Bkapi-Authorization: {"bk_app_code": "bk_apigw_test", "bk_app_secret": "***"}' \
    -d '{}' \
    'http://bcs-api.bkdomain/bcsapi/v4/clustermanager/v1/cluster/BCS-K8S-00000'
    ```

详情访问示例可以参考其他 clustermanager 相关的接口文档中的示例
