# 权限模型注册 Migration

## 原理

使用 [migrate](github.com/golang-migrate/migrate) 框架，读取 migrations 目录中的 json 文件，根据数据库中当前处于的版本，
执行当前版本以上的 migrations。

## migrations 格式

文件格式: `{version}_{name}_up.json`，version 从 `0` 开始。

migration 文件支持 go 模板参数，可以在 migrations 文件中定义，并通过 `Migrate` `templateVar` 参数上传入，程序将自动渲染模板。

内容格式: 参考 [Migration 内容规范](https://bk.tencent.com/docs/document/7.0/236/55314)

```json
{
  "system_id": "{{ .BK_IAM_SYSTEM_ID }}",
  "enabled": true,
  "operations": [
    {
      "operation": "upsert_system",
      "data": {
        "id": "{{ .BK_IAM_SYSTEM_ID }}",
        "name": "容器管理平台",
        "name_en": "BlueKing Container Service",
        "description": "蓝鲸容器管理平台基于原生Kubernetes，提供给用户高度可扩展、灵活易用的容器管理服务",
        "description_en": "The BlueKing Container Management platform provides highly scalable, flexible and easy-to-use container management services base on native Kubernetes",
        "clients": "{{ .APP_CODE }},bk_bcs_monitor,bk_bcs,bk_devops,bk_harbor",
        "provider_config": {
          "host": "{{ .BCS_HOST }}",
          "auth": "basic"
        }
      }
    }
  ]
}
```

## 添加/修改模型

在 migrations 目录中增加 migration 文件。