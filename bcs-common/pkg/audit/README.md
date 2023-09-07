# bcs-audit

BCS 审计和操作记录 SDK，提供接入手段。

使用示例：
```go
// 生成 Client
// import "github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
auditClient := audit.NewClient("bcs_host", "bcs_token", klog.V(4))

// 字段中有 `validate:"required"` 代表必填
// 用户名相关信息
auditCtx := audit.RecorderContext{}
// 操作的资源
resource := audit.Resource{}
// 如何操作
action := audit.Action{}
// 操作结果
result := audit.ActionResult{}
err := auditClient.R().SetContext(auditCtx).SetResource(resource).SetAction(action).SetResult(result).Do()


// 关闭审计
auditClient.R().DisableAudit()
// 关闭操作记录
auditClient.R().DisableActivity()
// 程序停止后，关闭操作记录数据通道
auditClient.Close()
```