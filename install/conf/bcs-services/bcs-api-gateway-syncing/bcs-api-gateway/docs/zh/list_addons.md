### 描述

获取集群组件列表

### 路径参数

| 参数名称        | 参数类型   | 必选 | 描述     |
|-------------|--------|----|--------|
| projectCode | string | 是  | 项目英文名  |
| clusterID   | string | 是  | 目标集群ID |

### 调用示例

```sh
curl -X GET \
-H 'x-bkapi-authorization: {"access_token": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' \
--insecure https://bcs-api-gateway.apigw.com/prod/helmmanager/v1/projects/projecttest/clusters/clustertest/addons
```

### 响应示例

响应体中 data 为一个数组，有多个 addons，addons 的字段与**获取集群组件详情**(`get_addons_details`)接口中的响应相同，可以在该接口文档中找到 addons 的字段说明

```json
{
  "code": 0,
  "message": "success",
  "result": true,
  "data": [
    {
      "name": "bcs-k8s-watch",
      "chartName": "bcs-k8s-watch",
      "description": "监控、采集基于 k8s 的 bcs 集群中的资源数据",
      "logo": "",
      "docsLink": "",
      "version": "1.29.0",
      "currentVersion": "1.29.0",
      "namespace": "bcs-system",
      "defaultValues": "env:\n  BK_BCS_clusterId: \"{{ .BCS_SYS_CLUSTER_ID }}\"\n  BK_BCS_customStorage: \"https://<bcs-storage地址>\"\nsecret:\n  bcsCertsOverride: true\n  ca_crt: |\n  tls_crt: |\n  tls_key: |",
      "currentValues": "env:\n  BK_BCS_clusterId: \"{{ .BCS_SYS_CLUSTER_ID }}\"\n  BK_BCS_customStorage: \"https://<bcs-storage地址>\"\nsecret:\n  bcsCertsOverride: true\n  ca_crt: |\n  tls_crt: |\n  tls_key: |",
      "status": "failed-install",
      "message": "execute error, install bcs-system/bcs-k8s-watch in cluster BCS-K8S-00000 error, rendered manifests contain a resource that already exists. Unable to continue with install: ServiceAccount \"bcs-k8s-watch\" in namespace \"bcs-system\" exists and cannot be imported into the current release: invalid ownership metadata; annotation validation error: key \"meta.helm.sh/release-name\" must equal \"bcs-k8s-watch\": current value is \"bcs-services-stack\"",
      "supportedActions": [
        "install",
        "upgrade",
        "uninstall"
      ],
      "releaseName": "bcs-k8s-watch"
    }
  ],
  "requestID": "33cd860befbf71846f1c5bbec2c22007",
  "web_annotations": null
}
```