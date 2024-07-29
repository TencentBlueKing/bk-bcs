/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package ctxutils xxx
package ctxutils

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
)

const (
	ctxAuditObject contextKey = "audit_object"
)

// ResourceType defines the resource type
type ResourceType string

// AuditAction defines the audit action
type AuditAction string

// RequestType defines the request type
type RequestType string

// AuditStatus defines the audit status
type AuditStatus string

// EmptyData defines the empty string
var EmptyData = ""

const (
	// AuditFailed defines the failed status
	AuditFailed AuditStatus = "Failed"
	// AuditSuccess defines the success status
	AuditSuccess AuditStatus = "Success"

	// ApplicationResource defines the application resource
	ApplicationResource ResourceType = "Application"
	// AppSetResource defines the appset resource
	AppSetResource ResourceType = "ApplicationSet"
	// RepoResource defines the repo resource
	RepoResource ResourceType = "Repository"
	// SecretResource defines the secret resource
	SecretResource ResourceType = "Secret"

	// GRPCRequest defines the grpc type
	GRPCRequest RequestType = "GRPC"
	// HTTPRequest defines the http type
	HTTPRequest RequestType = "HTTP"

	// None defines none action
	None AuditAction = "None"
	// ApplicationCreate defines the app create action
	ApplicationCreate AuditAction = "ApplicationCreate"
	// ApplicationUpdate defines the app update action
	ApplicationUpdate AuditAction = "ApplicationUpdate"
	// ApplicationSync defines the app sync action
	ApplicationSync AuditAction = "ApplicationSync"
	// ApplicationRollback defines the app rollback action
	ApplicationRollback AuditAction = "ApplicationRollback"
	// ApplicationDelete defines the app delete action
	ApplicationDelete AuditAction = "ApplicationDelete"
	// ApplicationSetCreateOrUpdate defines the appset create/update action
	ApplicationSetCreateOrUpdate AuditAction = "ApplicationSetCreateOrUpdate"
	// ApplicationSetDelete defines the appset delete action
	ApplicationSetDelete AuditAction = "ApplicationSetDelete"
	// RepoCreate defines the repo create action
	RepoCreate AuditAction = "RepoCreate"
	// RepoDelete defines the repo delete action
	RepoDelete AuditAction = "RepoDelete"
	// RepoUpdate defines the repo update action
	RepoUpdate AuditAction = "RepoUpdate"
	// SecretCreate defines the secret create action
	SecretCreate AuditAction = "SecretCreate"
	// SecretUpdate defines the secret update action
	SecretUpdate AuditAction = "SecretUpdate"
	// SecretDelete defines the secret delete action
	SecretDelete AuditAction = "SecretDelete"
	// SecretRollback defines the secret rollback action
	SecretRollback AuditAction = "SecretRollback"
)

// SetAuditMessage set audit message into context
func SetAuditMessage(r *http.Request, audit *dao.UserAudit) *http.Request {
	user := User(r.Context())
	audit.User = user.GetUser()
	audit.Client = user.ClientID
	audit.RequestID = RequestID(r.Context())
	audit.RequestURI = r.URL.RequestURI()
	if len(audit.RequestURI) > 256 {
		audit.RequestURI = audit.RequestURI[:256]
	}
	audit.RequestMethod = r.Method
	audit.SourceIP = getRequestSourceIP(r)
	audit.UserAgent = r.UserAgent()
	ctx := context.WithValue(r.Context(), ctxAuditObject, audit)
	return r.WithContext(ctx)
}

// AuditResp the response with audit
type AuditResp struct {
	StatusCode int
	ErrMsg     string
	Start      time.Time
	End        time.Time
}

// NeedAudit check the request whether need audit
func NeedAudit(ctx context.Context) bool {
	auditObj := ctx.Value(ctxAuditObject)
	if auditObj == nil {
		return false
	}
	_, ok := auditObj.(*dao.UserAudit)
	return ok
}

// SaveAuditMessage save the audit message
func SaveAuditMessage(ctx context.Context, auditResp *AuditResp) {
	auditObj := ctx.Value(ctxAuditObject)
	if auditObj == nil {
		return
	}
	userAudit, ok := auditObj.(*dao.UserAudit)
	if !ok {
		return
	}
	userAudit.ResponseStatus = auditResp.StatusCode
	if userAudit.ResponseStatus != http.StatusOK {
		userAudit.Status = string(AuditFailed)
	} else {
		userAudit.Status = string(AuditSuccess)
	}
	userAudit.ErrMsg = auditResp.ErrMsg
	// prevent gzip message
	if strings.Contains(userAudit.ErrMsg, "\x8B") {
		userAudit.ErrMsg = ""
	}
	userAudit.StartTime = auditResp.Start
	userAudit.EndTime = auditResp.End
	if err := dao.GlobalDB().SaveAuditMessage(userAudit); err != nil {
		blog.Errorf("RequestID[%s] save audit message failed: %s", RequestID(ctx), err.Error())
		return
	}
	saveBCSAudit(userAudit)
}

func saveBCSAudit(userAudit *dao.UserAudit) {
	auditCtx := audit.RecorderContext{
		Username:  userAudit.User,
		SourceIP:  userAudit.SourceIP,
		UserAgent: userAudit.UserAgent,
		RequestID: userAudit.RequestID,
		StartTime: userAudit.StartTime,
		EndTime:   userAudit.EndTime,
	}
	resourceData := userAudit.ResourceData
	if len(resourceData) > 1024 {
		resourceData = "too long no need save"
	}
	auditResource := audit.Resource{
		ProjectCode:  userAudit.Project,
		ResourceType: audit.ResourceTypeGitOps,
		ResourceID:   userAudit.ResourceType,
		ResourceName: userAudit.ResourceName,
		ResourceData: map[string]any{
			"data": resourceData,
		},
	}
	auditResult := audit.ActionResult{
		Status: audit.ActivityStatusSuccess,
		ExtraData: map[string]any{
			"Method": userAudit.RequestMethod,
			"Path":   userAudit.RequestURI,
		},
	}
	if userAudit.ErrMsg != "" {
		auditResult.Status = audit.ActivityStatusFailed
		auditResult.ResultCode = userAudit.ResponseStatus
		auditResult.ResultContent = userAudit.ErrMsg
	}
	auditAction := audit.Action{ActionID: userAudit.Action}
	switch userAudit.RequestMethod {
	case http.MethodDelete:
		auditAction.ActivityType = audit.ActivityTypeDelete
	case http.MethodPut:
		auditAction.ActivityType = audit.ActivityTypeUpdate
	case http.MethodPost:
		auditAction.ActivityType = audit.ActivityTypeCreate
	default:
		auditAction.ActivityType = audit.ActivityTypeUpdate
	}
	if err := component.GetAuditClient().R().
		SetContext(auditCtx).
		SetResource(auditResource).
		SetAction(auditAction).
		SetResult(auditResult).Do(); err != nil {
		blog.Errorf("save audit message failed: %s", err.Error())
	}
}

func getRequestSourceIP(req *http.Request) string {
	ip := req.Header.Get("X-Real-Ip")
	if ip == "" {
		ip = req.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = req.RemoteAddr
	}
	return ip
}
