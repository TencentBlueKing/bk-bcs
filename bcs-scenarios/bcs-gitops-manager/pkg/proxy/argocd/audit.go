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

package argocd

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/permitcheck"
)

// AuditPlugin plugin for user audit
type AuditPlugin struct {
	*mux.Router
	db            dao.Interface
	middleware    mw.MiddlewareInterface
	permitChecker permitcheck.PermissionInterface
}

// Init the user audit requests
func (plugin *AuditPlugin) Init() error {
	plugin.Path("/query").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.query))
	return nil
}

// QueryAuditsResponse defines the response for user audit query
type QueryAuditsResponse struct {
	Code      int64            `json:"code"`
	RequestID string           `json:"requestID"`
	Data      []*dao.UserAudit `json:"data"`
}

func (plugin *AuditPlugin) query(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projects := r.URL.Query()["projects"]
	if len(projects) == 0 {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, fmt.Errorf("query param 'projects' can not be empty"))
	}
	for i := range projects {
		project := projects[i]
		_, statusCode, err := plugin.permitChecker.CheckProjectPermission(r.Context(), project,
			permitcheck.ProjectViewRSAction)
		if err != nil {
			return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check permission for "+
				"project '%s' failed", project))
		}
	}
	userAuditQuery := &dao.UserAuditQuery{
		Projects:      projects,
		Users:         r.URL.Query()["users"],
		Actions:       r.URL.Query()["actions"],
		ResourceTypes: r.URL.Query()["rsTypes"],
		ResourceNames: r.URL.Query()["rsNames"],
		RequestIDs:    r.URL.Query()["reqids"],
		RequestURI:    strings.TrimSpace(r.URL.Query().Get("reqUri")),
		RequestType:   strings.TrimSpace(r.URL.Query().Get("reqType")),
		RequestMethod: strings.TrimSpace(r.URL.Query().Get("reqMethod")),
		StartTime:     strings.TrimSpace(r.URL.Query().Get("startTime")),
		EndTime:       strings.TrimSpace(r.URL.Query().Get("endTime")),
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.
				Wrapf(err, "query param 'limit' parse failed"))
		}
		userAuditQuery.Limit = limitInt
	} else {
		userAuditQuery.Limit = 50
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.
				Wrapf(err, "query param 'offset' parse failed"))
		}
		userAuditQuery.Offset = offsetInt
	}
	if userAuditQuery.StartTime != "" {
		if _, err := time.Parse(time.DateTime, userAuditQuery.StartTime); err != nil {
			return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "query param 'startTime' "+
				"with '%s' parse failed", userAuditQuery.StartTime))
		}
	}
	if userAuditQuery.EndTime != "" {
		if _, err := time.Parse(time.DateTime, userAuditQuery.EndTime); err != nil {
			return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "query param 'startTime' "+
				"with '%s' parse failed", userAuditQuery.EndTime))
		}
	}
	audits, err := plugin.db.QueryUserAudits(userAuditQuery)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError,
			errors.Wrapf(err, "query user audits failed"))
	}
	return r, mw.ReturnJSONResponse(&QueryAuditsResponse{
		Code:      0,
		RequestID: ctxutils.RequestID(r.Context()),
		Data:      audits,
	})
}
