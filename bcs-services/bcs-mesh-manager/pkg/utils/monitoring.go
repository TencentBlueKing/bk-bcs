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

package utils

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/project"
)

const (
	// GrafanaPath Grafana路径
	GrafanaPath = "/grafana/dashboard"
	// HTTPSProtocol HTTPS协议
	HTTPSProtocol = "https://"
	// BizIDParam bizId参数名
	BizIDParam = "bizId"
	// DashNameParam dashName参数名
	DashNameParam = "dashName"
)

// GenerateMonitoringLink 生成监控链接
// 格式: https://xxx.com/grafana/dashboard?bizId=xxx&dashName=xxx
func GenerateMonitoringLink(ctx context.Context, projectCode string) string {
	// 从GlobalOptions获取监控配置
	if options.GlobalOptions == nil || options.GlobalOptions.Monitoring == nil {
		blog.Errorf("GenerateMonitoringLink: GlobalOptions or Monitoring config is nil")
		return ""
	}

	monitoringConfig := options.GlobalOptions.Monitoring
	if monitoringConfig.Domain == "" {
		blog.Errorf("GenerateMonitoringLink: monitoring domain is empty")
		return ""
	}

	projectInfo, err := project.GetProjectByCode(ctx, projectCode)
	if err != nil {
		blog.Errorf("GenerateMonitoringLink: failed to get project info for projectCode %s, error: %s",
			projectCode, err.Error())
		return ""
	}

	if projectInfo.BusinessID == "" {
		blog.Errorf("GenerateMonitoringLink: businessID is empty for projectCode %s", projectCode)
		return ""
	}

	baseURL := HTTPSProtocol + monitoringConfig.Domain + GrafanaPath
	queryParams := fmt.Sprintf("%s=%s", BizIDParam, projectInfo.BusinessID)

	if monitoringConfig.DashName != "" {
		encodedDashName := url.PathEscape(monitoringConfig.DashName)
		queryParams += fmt.Sprintf("&%s=%s", DashNameParam, encodedDashName)
	}

	monitoringURL := fmt.Sprintf("%s?%s", baseURL, queryParams)

	return monitoringURL
}
