## 描述

管理集群项目信息

## 接口列表和参数

可以参考 swagger 文档查看相关接口以及参数等信息：[bcsproject.swagger.json](https://github.com/TencentBlueKing/bk-bcs/blob/master/bcs-services/bcs-project-manager/proto/bcsproject/bcsproject.swagger.json)

## 接口调用示例

以 Project 相关的接口为例

- 查询项目信息
    ```bash
    curl -X GET \
    -H 'X-Bkapi-Authorization: {"bk_app_code": "bk_apigw_test", "bk_app_secret": "***"}' \
    'http://bcs-api.bkdomain/bcsapi/v4/bcsproject/v1/projects'
    ```
详情访问示例可以参考其他 bcsproject 相关的接口文档中的示例
