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
 *
 */

package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/component"
)

var (
	ApplicationCreate          = "application_create"
	ApplicationUpdate          = "application_update"
	ApplicationSync            = "application_sync"
	ApplicationRollback        = "application_rollback"
	ApplicationDelete          = "application_delete"
	ApplicationPatchResource   = "application_patch_resource"
	ApplicationDeleteResource  = "application_resource_delete"
	ApplicationClean           = "application_clean"
	ApplicationDeleteResources = "application_multiple_resources_delete"
	ApplicationGRPCOperate     = "application_grpc_operate"

	ApplicationSetCreateOrUpdate = "applicationset_create_or_update"
	ApplicationSetDelete         = "applicationset_delete"
	ApplicationSetGenerate       = "applicationset_generate"

	ProjectOpen = "project_open"

	RepositoryCreate = "repository_create"
	RepositoryDelete = "repository_delete"
	RepositoryUpdate = "repository_update"

	SecretCreate   = "secret_create"
	SecretUpdate   = "secret_update"
	SecretDelete   = "secret_delete"
	SecretRollback = "secret_rollback"

	WebhookTriggered = "webhook_triggered"
)

const (
	contextObject contextKey = "object"
	contextAction contextKey = "action"
)

// SetAuditMessage set audit message with context
func SetAuditMessage(r *http.Request, obj interface{}, action string) *http.Request {
	ctx := context.WithValue(r.Context(), contextObject, obj)
	ctx = context.WithValue(ctx, contextAction, action)
	return r.WithContext(ctx)
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

func handleAudit(req *http.Request, resp *HttpResponse, start, end time.Time) {
	method := req.Method
	if method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete {
		return
	}
	auditCtx := audit.RecorderContext{
		Username:  User(req.Context()).UserName,
		SourceIP:  getRequestSourceIP(req),
		UserAgent: req.UserAgent(),
		RequestID: RequestID(req.Context()),
		StartTime: start,
		EndTime:   end,
	}
	obj := req.Context().Value(contextObject)
	if obj == nil {
		return
	}
	action, ok := req.Context().Value(contextAction).(string)
	if !ok {
		return
	}
	var auditResource audit.Resource
	var auditAction = audit.Action{ActionID: action}
	switch action {
	case ApplicationCreate, ApplicationUpdate, ApplicationSync, ApplicationRollback,
		ApplicationDelete, ApplicationPatchResource, ApplicationDeleteResource,
		ApplicationClean, ApplicationGRPCOperate:
		var app *v1alpha1.Application
		app, ok = obj.(*v1alpha1.Application)
		if !ok {
			return
		}
		auditResource = audit.Resource{
			ProjectCode:  app.Spec.Project,
			ResourceType: audit.ResourceTypeGitOps,
			ResourceID:   "Application",
			ResourceName: app.Name,
			ResourceData: make(map[string]any),
		}
		auditAction.ActivityType = audit.ActivityTypeUpdate
		switch action {
		case ApplicationCreate:
			auditAction.ActivityType = audit.ActivityTypeCreate
		case ApplicationDelete, ApplicationDeleteResource:
			auditAction.ActivityType = audit.ActivityTypeDelete
		}
		if action == ApplicationCreate || action == ApplicationUpdate {
			auditResource.ResourceData["Data"] = app
		}
	case ApplicationSetCreateOrUpdate, ApplicationSetDelete, ApplicationSetGenerate:
		var appset *v1alpha1.ApplicationSet
		appset, ok = obj.(*v1alpha1.ApplicationSet)
		if !ok {
			return
		}
		auditResource = audit.Resource{
			ProjectCode:  appset.Spec.Template.Spec.Project,
			ResourceType: audit.ResourceTypeGitOps,
			ResourceID:   "ApplicationSet",
			ResourceName: appset.Name,
			ResourceData: make(map[string]any),
		}
		switch action {
		case ApplicationSetCreateOrUpdate:
			auditAction.ActivityType = audit.ActivityTypeUpdate
			auditResource.ResourceData["Data"] = appset
		case ApplicationSetDelete:
			auditAction.ActivityType = audit.ActivityTypeDelete
		case ApplicationSetGenerate:
			auditAction.ActivityType = audit.ActivityTypeUpdate
		}
	case RepositoryCreate, RepositoryUpdate, RepositoryDelete:
		var repository *v1alpha1.Repository
		repository, ok = obj.(*v1alpha1.Repository)
		if !ok {
			return
		}
		auditResource = audit.Resource{
			ProjectCode:  repository.Project,
			ResourceType: audit.ResourceTypeGitOps,
			ResourceID:   "Repository",
			ResourceName: repository.Name,
			ResourceData: make(map[string]any),
		}
	case SecretCreate, SecretUpdate, SecretDelete, SecretRollback:
		var project string
		project, ok = obj.(string)
		if !ok {
			return
		}
		auditResource = audit.Resource{
			ProjectCode:  project,
			ResourceType: audit.ResourceTypeGitOps,
			ResourceID:   "Secret",
			ResourceName: project,
			ResourceData: make(map[string]any),
		}
	}
	var auditResult = audit.ActionResult{
		Status: audit.ActivityStatusSuccess,
		ExtraData: map[string]any{
			"Method": req.Method,
			"Path":   req.RequestURI,
		},
	}
	if resp.respType == returnError || resp.respType == returnGrpcError {
		auditResult.Status = audit.ActivityStatusFailed
		auditResult.ResultCode = resp.statusCode
		auditResult.ResultContent = resp.err.Error()
	}
	if err := component.GetAuditClient().R().
		SetContext(auditCtx).
		SetResource(auditResource).
		SetAction(auditAction).
		SetResult(auditResult).
		Do(); err != nil {
		blog.Errorf("save audit message failed: %s", err.Error())
	}
}
