# BCS Push Manager
基于Go-micro的微服务架构，实现推送事件管理、推送模版管理、推送通知功能。

## 项目结构

```
bcs-push-manager
├─ README.md
├─ bcs-push-manager.json
├─ cmd
│  ├─ options.go
│  └─ server.go
├─ go.mod
├─ go.sum
├─ internal
│  ├─ action
│  │  ├─ notification.go
│  │  ├─ pushevent.go
│  │  ├─ pushtemplate.go
│  │  └─ pushwhitelist.go
│  ├─ constant
│  │  └─ constant.go
│  ├─ handler
│  │  ├─ pushmanager.go
│  │  └─ utils.go
│  ├─ mq
│  │  ├─ message.go
│  │  ├─ mq.go
│  │  └─ rabbitmq
│  │     └─ rabbitmq.go
│  ├─ options
│  │  └─ options.go
│  ├─ requester
│  │  ├─ requester.go
│  │  └─ types.go
│  ├─ store
│  │  ├─ mongo
│  │  │  ├─ pushevent.go
│  │  │  ├─ pushtemplate.go
│  │  │  └─ pushwhitelist.go
│  │  ├─ store.go
│  │  └─ types
│  │     ├─ pushevent.go
│  │     ├─ pushtemplate.go
│  │     └─ pushwhitelist.go
│  └─ thirdparty
│     ├─ client.go
│     ├─ notification.go
│     └─ utils.go
├─ main.go
├─ pkg
│  └─ bcsapi
│     └─ thirdparty-service
│        ├─ bcs-thirdparty-service.pb.go
│        ├─ bcs-thirdparty-service.pb.gw.go
│        ├─ bcs-thirdparty-service.pb.micro.go
│        ├─ bcs-thirdparty-service.pb.validate.go
│        └─ bcs-thirdparty-service_grpc.pb.go
├─ proto
│  ├─ bcs-push-manager.pb.go
│  ├─ bcs-push-manager.pb.gw.go
│  ├─ bcs-push-manager.pb.micro.go
│  ├─ bcs-push-manager.pb.validate.go
│  ├─ bcs-push-manager.proto
│  ├─ bcs-push-manager.swagger.json
│  └─ bcs-push-manager_grpc.pb.go
└─ third_party
   ├─ google
   │  ├─ api
   │  │  ├─ annotations.proto
   │  │  ├─ field_behavior.proto
   │  │  ├─ http.proto
   │  │  └─ httpbody.proto
   │  └─ protobuf
   │     ├─ any.proto
   │     ├─ api.proto
   │     ├─ compiler
   │     │  └─ plugin.proto
   │     ├─ descriptor.proto
   │     ├─ duration.proto
   │     ├─ empty.proto
   │     ├─ field_mask.proto
   │     ├─ source_context.proto
   │     ├─ struct.proto
   │     ├─ timestamp.proto
   │     ├─ type.proto
   │     └─ wrappers.proto
   ├─ protoc-gen-openapiv2
   │  └─ options
   │     ├─ BUILD.bazel
   │     ├─ annotations.pb.go
   │     ├─ annotations.proto
   │     ├─ annotations.swagger.json
   │     ├─ openapiv2.pb.go
   │     ├─ openapiv2.proto
   │     └─ openapiv2.swagger.json
   ├─ protoc-gen-swagger
   │  └─ options
   │     ├─ annotations.proto
   │     └─ openapiv2.proto
   └─ validate
      └─ validate.proto

```

## 功能模块

### 1. 推送事件管理
- 创建推送事件 (CreatePushEvent)
- 删除推送事件 (DeletePushEvent)
- 获取推送事件 (GetPushEvent)
- 列出推送事件 (ListPushEvents)
- 更新推送事件 (UpdatePushEvent)

### 2. 推送白名单管理
- 创建推送白名单 (CreatePushWhitelist)
- 删除推送白名单 (DeletePushWhitelist)
- 获取推送白名单 (GetPushWhitelist)
- 列出推送白名单 (ListPushWhitelists)
- 更新推送白名单 (UpdatePushWhitelist)

### 3. 推送模板管理
- 创建推送模板 (CreatePushTemplate)
- 删除推送模板 (DeletePushTemplate)
- 获取推送模板 (GetPushTemplate)
- 列出推送模板 (ListPushTemplates)
- 更新推送模板 (UpdatePushTemplate)

## 数据库表结构

### push_events (推送事件)
- event_id: 事件ID
- domain: 域名
- rule_id: 规则ID
- event_detail: 事件详情
- push_level: 推送级别
- status: 状态
- notification_results: 通知结果
- dimension: 维度信息
- bk_biz_name: 业务名称
- metric_data: 指标数据
- created_at: 创建时间
- updated_at: 更新时间

### push_whitelists (推送白名单)
- whitelist_id: 白名单ID
- domain: 域名
- dimension: 维度信息
- reason: 申请原因
- applicant: 申请人
- approver: 审批人
- whitelist_status: 白名单状态
- approval_status: 审批状态
- start_time: 开始时间
- end_time: 结束时间
- approved_at: 审批时间
- created_at: 创建时间
- updated_at: 更新时间
- deleted_at: 删除时间（软删除）

### push_templates (推送模板)
- template_id: 模板ID
- domain: 域名
- template_type: 模板类型
- content: 模板内容（包含title、body、variables）
- creator: 创建者
- created_at: 创建时间
