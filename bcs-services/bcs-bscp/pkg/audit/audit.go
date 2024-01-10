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

	"k8s.io/klog/v2"

	ad "github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/audit/routematch"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sys"
)

var (
	auditClient *ad.Client
	auditOnce   sync.Once

	routeMatcher *routematch.RouteMatcher
	routeOnce    sync.Once
)

// GetAuditClient 获取审计客户端
func GetAuditClient() *ad.Client {
	if auditClient == nil {
		auditOnce.Do(func() {
			auditClient = ad.NewClient("", "", nil)
		})
	}
	return auditClient
}

// GetRouteMatcher 获取路由匹配器
func GetRouteMatcher() *routematch.RouteMatcher {
	if routeMatcher == nil {
		routeOnce.Do(func() {
			routeMatcher = routematch.NewRouteMatcher(getRoutes())
			klog.Infof("construct route map: %#v", routeMatcher.RouteMap())
		})
	}
	return routeMatcher
}

func getRoutes() []routematch.Route {
	rs := make([]routematch.Route, 0)
	for k := range auditFuncMap {
		if i := strings.Index(k, "."); i > 0 {
			rs = append(rs, routematch.Route{Method: k[:i], Pattern: k[i+1:]})
		}
	}
	return rs
}

var auditFuncMap = map[string]func() (ad.Resource, ad.Action){
	// config server api, see pkg/protocol/config-server/config_service.proto
	"POST./api/v1/config/create/app/app/biz_id/{biz_id}": func() (ad.Resource, ad.Action) {
		return ad.Resource{ResourceType: ad.ResourceType(sys.Application)},
			ad.Action{ActionID: "create_app", ActivityType: ad.ActivityTypeCreate}
	},
	"PUT./api/v1/config/update/app/app/app_id/{id}/biz_id/{biz_id}": func() (ad.Resource,
		ad.Action) {
		return ad.Resource{ResourceType: ad.ResourceType(sys.Application)},
			ad.Action{ActionID: "update_app", ActivityType: ad.ActivityTypeUpdate}
	},
	"DELETE./api/v1/config/delete/app/app/app_id/{id}/biz_id/{biz_id}": func() (ad.Resource,
		ad.Action) {
		return ad.Resource{ResourceType: ad.ResourceType(sys.Application)},
			ad.Action{ActionID: "delete_app", ActivityType: ad.ActivityTypeDelete}
	},
	"GET./api/v1/config/biz/{biz_id}/apps/{app_id}": func() (ad.Resource, ad.Action) {
		return ad.Resource{ResourceType: ad.ResourceType(sys.Application)},
			ad.Action{ActionID: "get_app", ActivityType: ad.ActivityTypeView}
	},
	"GET./api/v1/config/biz/{biz_id}/apps/query/name/{app_name}": func() (ad.Resource, ad.Action) {
		return ad.Resource{ResourceType: ad.ResourceType(sys.Application)},
			ad.Action{ActionID: "get_app_by_name", ActivityType: ad.ActivityTypeView}
	},
	"GET./api/v1/config/biz/{biz_id}/apps": func() (ad.Resource, ad.Action) {
		return ad.Resource{ResourceType: ad.ResourceType(sys.Application)},
			ad.Action{ActionID: "list_apps_rest", ActivityType: ad.ActivityTypeView}
	},
	"GET./api/v1/config/list/app/app/biz_id/{biz_id}": func() (ad.Resource, ad.Action) {
		return ad.Resource{ResourceType: ad.ResourceType(sys.Application)},
			ad.Action{ActionID: "list_apps_by_space_rest", ActivityType: ad.ActivityTypeView}
	},
	"POST./api/v1/config/create/config_item/config_item/app_id/{app_id}/biz_id/{biz_id}": func() (
		ad.Resource, ad.Action) {
		return ad.Resource{ResourceType: ad.ResourceType(sys.Application)},
			ad.Action{ActionID: "create_config_item", ActivityType: ad.ActivityTypeCreate}
	},
}
