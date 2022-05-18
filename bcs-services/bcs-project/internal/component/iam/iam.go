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

package iam

import (
	"fmt"

	"github.com/parnurzeal/gorequest"

	bcsIAM "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/errorx"
)

var (
	grantActionPath = "/api/v1/open/authorization/resource_creator_action/"
	timeout         = 10
)

// GrantResourceCreatorActions grant create action perm for resource
func GrantResourceCreatorActions(username string, projectID string, projectName string) error {
	iamConf := config.GlobalConf.IAM
	// 使用网关访问
	reqUrl := fmt.Sprintf("%s%s", iamConf.GatewayHost, grantActionPath)
	headers := map[string]string{"Content-Type": "application/json"}
	req := gorequest.SuperAgent{
		Url:    reqUrl,
		Method: "POST",
		Data: map[string]interface{}{
			"bk_app_code":   config.GlobalConf.App.Code,
			"bk_app_secret": config.GlobalConf.App.Secret,
			"creator":       username,
			"system":        bcsIAM.SystemIDBKBCS,
			"type":          "project",
			"id":            projectID,
			"name":          projectName,
		},
	}
	// 请求API
	proxy := ""
	_, err := component.Request(req, timeout, proxy, headers)
	if err != nil {
		logging.Error("grant creator actions for project failed, %v", err)
		return errorx.NewRequestIAMErr(err)
	}
	return nil
}
