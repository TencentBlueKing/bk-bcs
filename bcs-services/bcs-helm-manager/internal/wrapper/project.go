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

// Package wrapper xxx
package wrapper

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"go-micro.dev/v4/server"

	clusterClient "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component/clustermanager"
	projectClient "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
)

// ParseProjectIDWrapper parse projectID from req
func ParseProjectIDWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		body := req.Body()
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("ParseProjectIDWrapper error: %s", err)
		}

		type bodyStruct struct {
			ProjectCode string `json:"projectCode,omitempty"`
			ProjectID   string `json:"projectID,omitempty"`
			ClusterID   string `json:"clusterID,omitempty"`
		}
		project := &bodyStruct{}
		err = json.Unmarshal(b, project)
		if err != nil {
			return fmt.Errorf("ParseProjectIDWrapper error: %s", err)
		}
		if len(project.ProjectCode) == 0 {
			project.ProjectCode = project.ProjectID
		}

		if len(project.ProjectCode) == 0 {
			blog.Warn("ParseProjectIDWrapper error: projectCode is empty")
			return fn(ctx, req, rsp)
		}
		pj, err := projectClient.GetProjectByCode(project.ProjectCode)
		if err != nil {
			return fmt.Errorf("ParseProjectIDWrapper get projectID error, projectCode: %s, err: %s",
				project.ProjectCode, err.Error())
		}

		// check cluster
		if project.ClusterID != "" {
			cls, err := clusterClient.GetCluster(project.ClusterID)
			if err != nil {
				return fmt.Errorf("get cluster error, clusterID: %s, err: %s",
					project.ClusterID, err.Error())
			}
			if !cls.IsShared && cls.ProjectID != pj.ProjectID {
				return fmt.Errorf("cluster is invalid")
			}
		}

		ctx = context.WithValue(ctx, contextx.ProjectIDContextKey, pj.ProjectID)
		ctx = context.WithValue(ctx, contextx.ProjectCodeContextKey, pj.ProjectCode)
		return fn(ctx, req, rsp)
	}
}
