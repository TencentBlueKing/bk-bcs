/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package wrapper

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/micro/go-micro/v2/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
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
		}
		project := &bodyStruct{}
		err = json.Unmarshal(b, project)
		if err != nil {
			return fmt.Errorf("ParseProjectIDWrapper error: %s", err)
		}
		if len(project.ProjectCode) == 0 {
			project.ProjectCode = project.ProjectID
		}

		username := auth.GetUserFromCtx(ctx)
		if len(username) == 0 || len(project.ProjectCode) == 0 {
			blog.Warn("ParseProjectIDWrapper error: username or projectCode is invalid")
			return fn(ctx, req, rsp)
		}
		projectID, err := projectClient.GetProjectIDByCode(auth.GetUserFromCtx(ctx), project.ProjectCode)
		if err != nil {
			return fmt.Errorf("ParseProjectIDWrapper get projectID error, projectCode: %s, err: %s",
				project.ProjectCode, err.Error())
		}

		ctx = context.WithValue(ctx, contextx.ProjectIDContextKey, projectID)
		return fn(ctx, req, rsp)
	}
}
