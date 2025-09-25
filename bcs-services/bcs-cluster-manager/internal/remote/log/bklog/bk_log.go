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

// Package bklog xxx
package bklog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	ioptions "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
)

const (
	defaultTimeOut = time.Second * 60
)

// EnableMonitorAuditResponse enable monitor audit response
type EnableMonitorAuditResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	RequestId string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// EnableMonitorAudit enable monitor audit
func EnableMonitorAudit(projectID, clusterID string) error { // nolint
	host := ioptions.GetGlobalCMOptions().ComponentDeploy.LogCollector.HttpServer
	rawURL := host + fmt.Sprintf("/bcsapi/v4/monitor/api/audit/projects/%s/clusters/%s/enable",
		projectID, clusterID)

	start := time.Now()
	respData := &EnableMonitorAuditResponse{}

	token := ioptions.GetGlobalCMOptions().ComponentDeploy.LogCollector.Token

	// api request empty json body
	emptyBody := map[string]interface{}{
		"": "",
	}

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Put(rawURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("Connection", "close").
		Set("Authorization", "Bearer "+token).
		Send(emptyBody).
		End()

	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("bklog", "EnableMonitorAudit", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api EnableMonitorAudit failed: %v", errs[0])
		return errs[0]
	}

	if result.StatusCode != http.StatusOK {
		metrics.ReportLibRequestMetric("bklog", "EnableMonitorAudit", "http", metrics.LibCallStatusErr, start)
		errMsg := fmt.Errorf("call EnableMonitorAudit API[%v] error: code[%v], body[%v]",
			rawURL, result.StatusCode, body)
		blog.Errorf("EnableMonitorAudit failed: %v", errMsg)
		return errMsg
	}

	if err := json.Unmarshal([]byte(body), respData); err != nil {
		metrics.ReportLibRequestMetric("bklog", "EnableMonitorAudit", "http", metrics.LibCallStatusErr, start)
		errMsg := fmt.Errorf("parse EnableMonitorAudit response failed: %v, raw body: %s", err, string(body))
		blog.Errorf("EnableMonitorAudit failed: %v", errMsg)
		return errMsg
	}

	if respData.Code != 0 {
		metrics.ReportLibRequestMetric("bklog", "EnableMonitorAudit", "http", metrics.LibCallStatusErr, start)
		errMsg := fmt.Errorf("call EnableMonitorAudit API failed: code[%v], message[%s]",
			respData.Code, respData.Message)
		blog.Errorf("EnableMonitorAudit failed: %v", errMsg)
		return errMsg
	}

	metrics.ReportLibRequestMetric("bklog", "EnableMonitorAudit", "http", metrics.LibCallStatusOK, start)
	blog.Infof("EnableMonitorAudit for project[%s] cluster[%s] successfully", projectID, clusterID)

	return nil
}
