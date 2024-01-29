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

// Package audit is for audit operation
package audit

import (
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	"k8s.io/klog/v2"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/audit/routematch"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sys"
)

var (
	auditClient *audit.Client
	auditOnce   sync.Once

	routeMatcher *routematch.RouteMatcher
	routeOnce    sync.Once
)

// GetAuditClient 获取审计客户端
func GetAuditClient() *audit.Client {
	if auditClient == nil {
		auditOnce.Do(func() {
			auditClient = audit.NewClient("", "", nil)
		})
	}
	return auditClient
}

// GetRouteMatcher 获取路由匹配器
func GetRouteMatcher() *routematch.RouteMatcher {
	if routeMatcher == nil {
		routeOnce.Do(func() {
			routeMatcher = routematch.NewRouteMatcher(getRoutes())
		})
	}
	return routeMatcher
}

func getRoutes() []routematch.Route {
	rs := make([]routematch.Route, 0)
	for k := range auditHttpMap {
		if i := strings.Index(k, "."); i > 0 {
			rs = append(rs, routematch.Route{Method: k[:i], Pattern: k[i+1:]})
		}
	}
	return rs
}

type auditParam struct {
	Username     string
	SourceIP     string
	UserAgent    string
	Rid          string
	Resource     audit.Resource
	Action       audit.Action
	StartTime    time.Time
	ResultStatus int
	ResultMsg    string
}

// ignoredField means ignored fields for audit event , eg: all fields for activity
const ignoredField string = "ignored"

func addAudit(p auditParam) {
	auditCtx := audit.RecorderContext{
		Username:  p.Username,
		SourceIP:  p.SourceIP,
		UserAgent: p.UserAgent,
		RequestID: p.Rid,
		StartTime: p.StartTime,
		EndTime:   time.Now(),
	}
	resource := audit.Resource{
		ProjectCode:  ignoredField,
		ResourceID:   ignoredField,
		ResourceName: ignoredField,
		ResourceType: p.Resource.ResourceType,
		ResourceData: p.Resource.ResourceData,
	}
	action := audit.Action{
		ActionID:     p.Action.ActionID,
		ActivityType: audit.ActivityType(ignoredField),
	}

	result := audit.ActionResult{
		Status:        audit.ActivityStatus(ignoredField),
		ResultCode:    p.ResultStatus,
		ResultContent: p.ResultMsg,
	}

	if err := GetAuditClient().R().DisableActivity().SetContext(auditCtx).SetResource(resource).SetAction(action).
		SetResult(result).Do(); err != nil {
		klog.Errorf("add audit err: %v", err)
	}
}

var auditHttpMap = map[string]func() (audit.Resource, audit.Action){
	// api-server api
	"PUT./api/v1/biz/{biz_id}/content/upload": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "upload_file"}
	},
	"GET./api/v1/biz/{biz_id}/content/download": func() (audit.Resource,
		audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "download_file"}
	},
	"GET./api/v1/biz/{biz_id}/content/metadata": func() (audit.Resource,
		audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "get_file_metadata"}
	},
}

var auditGrpcMap = map[string]func() (audit.Resource, audit.Action){
	// config-server api, see pkg/protocol/config-server/config_service.proto
	"/pbcs.Config/CreateApp": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "create_app"}
	},
	"/pbcs.Config/UpdateApp": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "update_app"}
	},
	"/pbcs.Config/DeleteApp": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "delete_app"}
	},
	"/pbcs.Config/GetApp": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "get_app"}
	},
	"/pbcs.Config/GetAppByName": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "get_app_by_name"}
	},
	"/pbcs.Config/ListAppsRest": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "list_apps_rest"}
	},
	"/pbcs.Config/ListAppsBySpaceRest": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "list_apps_by_space_rest"}
	},
	"/pbcs.Config/CreateConfigItem": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "create_config_item"}
	},
	"/pbcs.Config/BatchUpsertConfigItems": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "batch_upsert_config_items"}
	},
	"/pbcs.Config/DeleteConfigItem": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "delete_config_item"}
	},
	"/pbcs.Config/GetConfigItem": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "get_config_item"}
	},
	"/pbcs.Config/GetReleasedConfigItem": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "get_released_config_item"}
	},
	"/pbcs.Config/ListConfigItems": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "list_config_items"}
	},
	"/pbcs.Config/ListReleasedConfigItems": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "list_released_config_items"}
	},
	"/pbcs.Config/ListConfigItemCount": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "list_config_item_count"}
	},
	"/pbcs.Config/ListConfigItemByTuple": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "list_config_item_by_tuple"}
	},
	"/pbcs.Config/GetReleasedKv": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "get_released_kv"}
	},
	"/pbcs.Config/ListReleasedKvs": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "list_released_kvs"}
	},
	"/pbcs.Config/UpdateConfigHook": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "update_config_hook"}
	},
	"/pbcs.Config/CreateRelease": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "create_release"}
	},
	"/pbcs.Config/ListReleases": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "list_releases"}
	},
	"/pbcs.Config/GetReleaseByName": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "get_release_by_name"}
	},
	"/pbcs.Config/DeprecateRelease": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "deprecate_release"}
	},
	"/pbcs.Config/UnDeprecateRelease": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "undeprecate_release"}
	},
	"/pbcs.Config/DeleteRelease": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "delete_release"}
	},
	"/pbcs.Config/CreateHook": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "create_hook"}
	},
	"/pbcs.Config/DeleteHook": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "delete_hook"}
	},
	"/pbcs.Config/ListHooks": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_hooks"}
	},
	"/pbcs.Config/ListHookTags": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_hook_tags"}
	},
	"/pbcs.Config/GetHook": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "get_hook"}
	},
	"/pbcs.Config/CreateHookRevision": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "create_hook_revision"}
	},
	"/pbcs.Config/ListHookRevisions": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_hook_revisions"}
	},
	"/pbcs.Config/DeleteHookRevision": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "delete_hook_revision"}
	},
	"/pbcs.Config/PublishHookRevision": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "publish_hook_revision"}
	},
	"/pbcs.Config/GetHookRevision": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "get_hook_revision"}
	},
	"/pbcs.Config/UpdateHookRevision": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "update_hook_revision"}
	},
	"/pbcs.Config/ListHookReferences": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_hook_references"}
	},
	"/pbcs.Config/ListHookRevisionReferences": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_hook_revision_references"}
	},
	"/pbcs.Config/GetReleaseHook": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "get_release_hook"}
	},
	"/pbcs.Config/CreateTemplateSpace": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "create_template_space"}
	},
	"/pbcs.Config/DeleteTemplateSpace": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "delete_template_space"}
	},
	"/pbcs.Config/UpdateTemplateSpace": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "update_template_space"}
	},
	"/pbcs.Config/ListTemplateSpaces": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_template_spaces"}
	},
	"/pbcs.Config/GetAllBizsOfTmplSpaces": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "get_all_bizs_of_tmpl_spaces"}
	},
	"/pbcs.Config/CreateDefaultTmplSpace": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "create_default_tmpl_space"}
	},
	"/pbcs.Config/ListTmplSpacesByIDs": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_tmpl_spaces_by_ids"}
	},
	"/pbcs.Config/CreateTemplate": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "create_template"}
	},
	"/pbcs.Config/DeleteTemplate": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "delete_template"}
	},
	"/pbcs.Config/BatchDeleteTemplate": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "batch_delete_template"}
	},
	"/pbcs.Config/UpdateTemplate": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "update_template"}
	},
	"/pbcs.Config/ListTemplates": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_templates"}
	},
	"/pbcs.Config/BatchUpsertTemplates": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "batch_upsert_templates"}
	},
	"/pbcs.Config/AddTmplsToTmplSets": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "add_tmpls_to_tmpl_sets"}
	},
	"/pbcs.Config/DeleteTmplsFromTmplSets": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "delete_tmpls_from_tmpl_sets"}
	},
	"/pbcs.Config/ListTemplatesByIDs": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_templates_by_ids"}
	},
	"/pbcs.Config/ListTemplatesNotBound": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_templates_not_bound"}
	},
	"/pbcs.Config/ListTemplateByTuple": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_template_by_tuple"}
	},
	"/pbcs.Config/ListTmplsOfTmplSet": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_tmpls_of_tmpl_set"}
	},
	"/pbcs.Config/CreateTemplateRevision": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "create_template_revision"}
	},
	"/pbcs.Config/ListTemplateRevisions": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_template_revisions"}
	},
	"/pbcs.Config/ListTemplateRevisionsByIDs": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_template_revisions_by_ids"}
	},
	"/pbcs.Config/ListTmplRevisionNamesByTmplIDs": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_tmpl_revision_names_by_tmpl_ids"}
	},
	"/pbcs.Config/CreateTemplateSet": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "create_template_set"}
	},
	"/pbcs.Config/DeleteTemplateSet": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "delete_template_set"}
	},
	"/pbcs.Config/UpdateTemplateSet": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "update_template_set"}
	},
	"/pbcs.Config/ListTemplateSets": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_template_sets"}
	},
	"/pbcs.Config/ListAppTemplateSets": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "list_app_template_sets"}
	},
	"/pbcs.Config/ListTemplateSetsByIDs": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_template_sets_by_ids"}
	},
	"/pbcs.Config/ListTmplSetsOfBiz": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_tmpl_sets_of_biz"}
	},
	"/pbcs.Config/CreateAppTemplateBinding": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "create_app_template_binding"}
	},
	"/pbcs.Config/DeleteAppTemplateBinding": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "delete_app_template_binding"}
	},
	"/pbcs.Config/UpdateAppTemplateBinding": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "update_app_template_binding"}
	},
	"/pbcs.Config/ListAppTemplateBindings": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "list_app_template_bindings"}
	},
	"/pbcs.Config/ListAppBoundTmplRevisions": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "list_app_bound_tmpl_revisions"}
	},
	"/pbcs.Config/ListReleasedAppBoundTmplRevisions": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "list_released_app_bound_tmpl_revisions"}
	},
	"/pbcs.Config/GetReleasedAppBoundTmplRevision": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "get_released_app_bound_tmpl_revision"}
	},
	"/pbcs.Config/UpdateAppBoundTmplRevisions": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "update_app_bound_tmpl_revisions"}
	},
	"/pbcs.Config/DeleteAppBoundTmplSets": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "delete_app_bound_tmpl_sets"}
	},
	"/pbcs.Config/CheckAppTemplateBinding": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "check_app_template_binding"}
	},
	"/pbcs.Config/ListTmplBoundCounts": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_tmpl_bound_counts"}
	},
	"/pbcs.Config/ListTmplRevisionBoundCounts": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_tmpl_revision_bound_counts"}
	},
	"/pbcs.Config/ListTmplSetBoundCounts": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_tmpl_set_bound_counts"}
	},
	"/pbcs.Config/ListTmplBoundUnnamedApps": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_tmpl_bound_unnamed_apps"}
	},
	"/pbcs.Config/ListTmplBoundNamedApps": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_tmpl_bound_named_apps"}
	},
	"/pbcs.Config/ListTmplBoundTmplSets": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_tmpl_bound_tmpl_sets"}
	},
	"/pbcs.Config/ListMultiTmplBoundTmplSets": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_multi_tmpl_bound_tmpl_sets"}
	},
	"/pbcs.Config/ListTmplRevisionBoundUnnamedApps": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_tmpl_revision_bound_unnamed_apps"}
	},
	"/pbcs.Config/ListTmplRevisionBoundNamedApps": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_tmpl_revision_bound_named_apps"}
	},
	"/pbcs.Config/ListTmplSetBoundUnnamedApps": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_tmpl_set_bound_unnamed_apps"}
	},
	"/pbcs.Config/ListMultiTmplSetBoundUnnamedApps": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_multi_tmpl_set_bound_unnamed_apps"}
	},
	"/pbcs.Config/ListTmplSetBoundNamedApps": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_tmpl_set_bound_named_apps"}
	},
	"/pbcs.Config/ListLatestTmplBoundUnnamedApps": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_latest_tmpl_bound_unnamed_apps"}
	},
	"/pbcs.Config/CreateTemplateVariable": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "create_template_variable"}
	},
	"/pbcs.Config/DeleteTemplateVariable": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "delete_template_variable"}
	},
	"/pbcs.Config/UpdateTemplateVariable": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "update_template_variable"}
	},
	"/pbcs.Config/ListTemplateVariables": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_template_variables"}
	},
	"/pbcs.Config/ImportTemplateVariables": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "import_template_variables"}
	},
	"/pbcs.Config/ExtractAppTmplVariables": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "extract_app_tmpl_variables"}
	},
	"/pbcs.Config/GetAppTmplVariableRefs": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "get_app_tmpl_variable_refs"}
	},
	"/pbcs.Config/GetReleasedAppTmplVariableRefs": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "get_released_app_tmpl_variable_refs"}
	},
	"/pbcs.Config/UpdateAppTmplVariables": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "update_app_tmpl_variables"}
	},
	"/pbcs.Config/ListAppTmplVariables": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "list_app_tmpl_variables"}
	},
	"/pbcs.Config/ListReleasedAppTmplVariables": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "list_released_app_tmpl_variables"}
	},
	"/pbcs.Config/CreateGroup": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "create_group"}
	},
	"/pbcs.Config/DeleteGroup": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "delete_group"}
	},
	"/pbcs.Config/UpdateGroup": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "update_group"}
	},
	"/pbcs.Config/ListAllGroups": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_all_groups"}
	},
	"/pbcs.Config/ListAppGroups": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_app_groups"}
	},
	"/pbcs.Config/ListGroupReleasedApps": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "list_group_released_apps"}
	},
	"/pbcs.Config/GetGroupByName": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Business)},
			audit.Action{ActionID: "get_group_by_name"}
	},
	"/pbcs.Config/Publish": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "publish"}
	},
	"/pbcs.Config/GenerateReleaseAndPublish": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "generate_release_and_publish"}
	},
	"/pbcs.Config/CreateCredentials": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.AppCredential)},
			audit.Action{ActionID: "create_credentials"}
	},
	"/pbcs.Config/ListCredentials": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.AppCredential)},
			audit.Action{ActionID: "list_credentials"}
	},
	"/pbcs.Config/DeleteCredential": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.AppCredential)},
			audit.Action{ActionID: "delete_credential"}
	},
	"/pbcs.Config/UpdateCredential": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.AppCredential)},
			audit.Action{ActionID: "update_credential"}
	},
	"/pbcs.Config/ListCredentialScopes": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.AppCredential)},
			audit.Action{ActionID: "list_credential_scopes"}
	},
	"/pbcs.Config/UpdateCredentialScope": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.AppCredential)},
			audit.Action{ActionID: "update_credential_scope"}
	},
	"/pbcs.Config/CreateKv": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "create_kv"}
	},
	"/pbcs.Config/UpdateKv": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "update_kv"}
	},
	"/pbcs.Config/ListKvs": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "list_kvs"}
	},
	"/pbcs.Config/DeleteKv": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "delete_kv"}
	},
	"/pbcs.Config/BatchUpsertKvs": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "batch_upsert_kvs"}
	},
	"/pbcs.Config/UnDeleteKv": func() (audit.Resource, audit.Action) {
		return audit.Resource{ResourceType: audit.ResourceType(sys.Application)},
			audit.Action{ActionID: "undelete_kv"}
	},
}
