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

// Package activity xxx
package activity

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"
	restful "github.com/emicklei/go-restful/v3"
	"github.com/gorilla/schema"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/errors"
	utils2 "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
)

// SearchActivitiesForm search activities form
type SearchActivitiesForm struct {
	ResourceType string `schema:"resource_type"`
	ActivityType string `schema:"activity_type"`
	Status       string `schema:"status"`
	StartTime    int64  `schema:"start_time"`
	EndTime      int64  `schema:"end_time"`
	Limit        int    `schema:"limit"`
	Offset       int    `schema:"offset"`
}

// SearchActivitiesResponse activity response
type SearchActivitiesResponse struct {
	ID           uint            `json:"id"`
	ProjectCode  string          `json:"project_code"`
	ResourceType string          `json:"resource_type"`
	ResourceName string          `json:"resource_name"`
	ResourceID   string          `json:"resource_id"`
	ActivityType string          `json:"activity_type"`
	Status       string          `json:"status"`
	Username     string          `json:"username"`
	CreatedAt    utils2.JSONTime `json:"created_at"`
	Description  string          `json:"description"`
	SourceIP     string          `json:"source_ip"`
	UserAgent    string          `json:"user_agent"`
	Extra        string          `json:"extra"`
}

var decoder = schema.NewDecoder()

// SearchActivities search activities
func SearchActivities(request *restful.Request, response *restful.Response) {
	project := utils.GetProjectFromAttribute(request)
	if project == nil {
		utils.ResponseParamsError(response, errors.ErrProjectNotFound)
		return
	}

	var form SearchActivitiesForm
	err := decoder.Decode(&form, request.Request.URL.Query())
	if err != nil {
		utils.ResponseParamsError(response, err)
		return
	}
	if form.Limit == 0 {
		form.Limit = 10
	}

	startTime := time.Unix(form.StartTime, 0)
	endTime := time.Unix(form.EndTime, 0)
	activities, count, err := sqlstore.SearchActivities(project.ProjectCode, form.ResourceType, form.ActivityType,
		models.GetStatus(form.Status), startTime, endTime, form.Offset, form.Limit)
	if err != nil {
		utils.ResponseDBError(response, err)
		return
	}
	results := make([]SearchActivitiesResponse, 0)
	for _, v := range activities {
		results = append(results, SearchActivitiesResponse{
			ID:           v.ID,
			ProjectCode:  v.ProjectCode,
			ResourceType: v.ResourceType,
			ResourceName: v.ResourceName,
			ResourceID:   v.ResourceID,
			ActivityType: v.ActivityType,
			Status:       v.Status.String(),
			Username:     v.Username,
			CreatedAt:    utils2.JSONTime{Time: v.CreatedAt},
			Description:  v.Description,
			Extra:        v.Extra,
			SourceIP:     v.SourceIP,
			UserAgent:    v.UserAgent,
		})
	}
	utils.ResponseOK(response, map[string]interface{}{
		"count": count,
		"items": results,
	})

}

// PushActivitiesForm push activities form
type PushActivitiesForm struct {
	Activities []PushActivitiesData `json:"activities" validate:"required"`
}

// PushActivitiesData push activities data
type PushActivitiesData struct {
	ProjectCode  string `json:"project_code" validate:"required"`
	ResourceType string `json:"resource_type" validate:"required"`
	ResourceName string `json:"resource_name" validate:"required"`
	ResourceID   string `json:"resource_id"`
	ActivityType string `json:"activity_type" validate:"required"`
	Status       string `json:"status"`
	Username     string `json:"username" validate:"required"`
	Description  string `json:"description"`
	SourceIP     string `json:"source_ip"`
	UserAgent    string `json:"user_agent"`
	Extra        string `json:"extra"`
}

// PushActivities push activities
func PushActivities(request *restful.Request, response *restful.Response) {
	form := PushActivitiesForm{}
	_ = request.ReadEntity(&form)
	err := utils.Validate.Struct(&form)
	if err != nil {
		utils.ResponseParamsError(response, err)
		return
	}

	activities := make([]*models.Activity, 0)
	for _, v := range form.Activities {
		var project *component.Project
		project, err = component.GetProjectWithCache(request.Request.Context(), v.ProjectCode)
		if err != nil {
			blog.Errorf("get project failed, err %s", err.Error())
			continue
		}
		activities = append(activities, &models.Activity{
			ProjectCode:  project.ProjectCode,
			ResourceType: v.ResourceType,
			ResourceName: v.ResourceName,
			ResourceID:   v.ResourceID,
			ActivityType: v.ActivityType,
			Status:       models.GetStatus(v.Status),
			Username:     v.Username,
			Description:  v.Description,
			Extra:        v.Extra,
			SourceIP:     v.SourceIP,
			UserAgent:    v.UserAgent,
		})
	}
	err = sqlstore.CreateActivity(activities)
	if err != nil {
		utils.ResponseDBError(response, err)
		return
	}
	utils.ResponseOK(response, nil)
}

var (
	resourceTypes = []string{
		"project",
		"cluster",
		"node",
		"node_group",
		"cloud_account",
		"namespace",
		"templateset",
		"variable",
		"k8s_resource",
		"helm",
		"addons",
		"chart",
		"web_console",
		"log_rule",
	}
)

// ResourceTypeResponse resource type response
type ResourceTypeResponse struct {
	ResourceType string `json:"resource_type"`
	Name         string `json:"name"`
}

// ResourceTypes resource types
func ResourceTypes(request *restful.Request, response *restful.Response) {
	items := make([]ResourceTypeResponse, 0)
	for _, v := range resourceTypes {
		items = append(items, ResourceTypeResponse{ResourceType: v, Name: i18n.T(request.Request.Context(), v)})
	}
	utils.ResponseOK(response, items)
}
