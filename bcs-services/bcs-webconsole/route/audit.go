package route

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	consoleAudit "github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/audit"
	"github.com/gin-gonic/gin"
	"time"
)

type resource struct {
	ClusterID string `json:"cluster_id" yaml:"cluster_id"`
	ProjectID string `json:"project_id" yaml:"project_id"`
}

// resource to map
func (r resource) toMap() map[string]any {
	result := make(map[string]any, 0)

	if r.ClusterID != "" {
		result["ClusterID"] = r.ClusterID
	}

	if r.ProjectID != "" {
		result["ProjectID"] = r.ProjectID
	}

	return result
}

var auditFuncMap = map[string]func(c *gin.Context) (audit.Resource, audit.Action){
	"GET./projects/:projectId/clusters/:clusterId/": func(c *gin.Context) (audit.Resource, audit.Action) {
		res := getResourceID(c)
		return audit.Resource{ResourceType: audit.ResourceTypeWebConsole, ResourceData: res.toMap()},
			audit.Action{ActionID: "web_console_start", ActivityType: audit.ActivityTypeStart}
	},
	"GET./projects/:projectId/mgr/": func(c *gin.Context) (audit.Resource, audit.Action) {
		res := getResourceID(c)
		return audit.Resource{ResourceType: audit.ResourceTypeWebConsole, ResourceData: res.toMap()},
			audit.Action{ActionID: "web_console_start", ActivityType: audit.ActivityTypeStart}
	},
}

// AuditHandler 操作记录中间件
func AuditHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		endTime := time.Now()

		addAudit(c, startTime, endTime)
	}
}

// 获取resourceData 的资源
func getResourceID(ctx *gin.Context) resource {
	projectId := ctx.Param("projectId")
	clusterId := ctx.Param("clusterId")
	return resource{
		ClusterID: clusterId,
		ProjectID: projectId,
	}
}

func addAudit(c *gin.Context, startTime, endTime time.Time) {
	method := c.Request.Method
	path := c.FullPath()
	s := method + "." + path
	// get method audit func
	fn, ok := auditFuncMap[s]
	if !ok {
		return
	}

	res, act := fn(c)

	authCtx := MustGetAuthContext(c)

	auditCtx := audit.RecorderContext{
		Username:  authCtx.Username,
		RequestID: authCtx.RequestId,
		StartTime: startTime,
		EndTime:   endTime,
	}
	resource := audit.Resource{
		ProjectCode:  authCtx.ProjectCode,
		ResourceType: res.ResourceType,
		ResourceID:   res.ResourceID,
		ResourceName: res.ResourceName,
		ResourceData: res.ResourceData,
	}
	action := audit.Action{
		ActionID:     act.ActionID,
		ActivityType: act.ActivityType,
	}

	result := audit.ActionResult{
		Status: audit.ActivityStatusSuccess,
	}

	result.ResultCode = c.Writer.Status()
	consoleAudit.GetAuditClient().R().
		SetContext(auditCtx).SetResource(resource).SetAction(action).SetResult(result).Do()
}
