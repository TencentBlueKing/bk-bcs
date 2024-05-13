# BCS 网关资源同步镜像

该镜像打包了 BCS 网关资源，并可以在部署时自动注册网关。

## 环境变量依赖


| 变量名                        | 说明                            | 样例                                    |
| ----------------------------- | ------------------------------- | --------------------------------------- |
| BKAPI_DESCRIPTION             | 网关描述                        |                                         |
| BK_API_URL_TMPL               | 网关 API 地址模板               | http://bkapi.example.com/api/{api_name} |
| BK_APP_CODE                   | 应用                            | bcs                                     |
| BK_APP_SECRET                 | 应用密钥                        |                                         |
| BKAPI_RELEASE_VERSION         | 网关版本号(推荐使用 appVersion) | 1.0.0                                   |
| BKAPI_RELEASE_COMMENT         | 网关版本说明                    |                                         |
| BKAPI_STAGE_HOST              | 网关部署环境地址                | http://bcs-app.svc                      |
| BK_API_GRANT_PERMISSIONS_APPS | 主动授权应用                    | bk_apigateway                           |

## Helm chart 集成

声明一个 Job，在部署时通过镜像，拉起容器，执行命令：`sync-apigateway`，进行网关注册操作。
**运行成功后，会在 /data/apigateway.pub 中生成网关公钥，可以验证网关请求。**

## apigw-manager 镜像构建

[apigw-manager](https://github.com/TencentBlueKing/bkpaas-python-sdk/tree/master/sdks/apigw-manager)
