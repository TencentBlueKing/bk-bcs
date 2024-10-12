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

// Package job xxx
package job

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/utils"
)

// jobClient global blueking job client
var jobClient *Client

// SetJobClient set job client
func SetJobClient(options Options) error {
	cli, err := NewJobClient(options)
	if err != nil {
		return err
	}

	jobClient = cli
	return nil
}

// GetJobClient get job client
func GetJobClient() *Client {
	return jobClient
}

// Client for job
type Client struct {
	utils.CommonClient
	bkUserName string
	userAuth   string
}

// NewJobClient create job client
func NewJobClient(options Options) (*Client, error) {
	c := &Client{
		CommonClient: utils.CommonClient{
			AppCode:   options.AppCode,
			AppSecret: options.AppSecret,
			Server:    options.Server,
			Debug:     options.Debug,
		},
		bkUserName: options.BKUserName,
	}

	auth, err := utils.BuildGateWayAuth(&utils.AuthInfo{
		BkAppUser: utils.BkAppUser{
			BkAppCode:   options.AppCode,
			BkAppSecret: options.AppSecret,
		},
		BkUserName: options.BKUserName,
	}, "")
	if err != nil {
		return nil, err
	}
	c.userAuth = auth

	return c, nil
}

// ExecuteScript xxx
func (c *Client) ExecuteScript(ctx context.Context, paras ExecuteScriptParas) (uint64, error) {
	var (
		_    = "ExecuteScript"
		path = fmt.Sprintf("%s/api/v3/fast_execute_script/", c.Server)
	)

	if options.GetEditionInfo().IsCommunicationEdition() {
		path = fmt.Sprintf("%s/api/c/compapi/v2/jobv3/fast_execute_script/", c.Server)
	}

	req := transToBkJobExecuteScriptReq(paras)
	resp := &FastExecuteScriptRsp{}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(utils.DefaultTimeOut).
		Post(path).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.Debug).
		Send(req).
		EndStruct(&resp)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("job", "ExecuteScript", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api blueking Job ExecuteScript failed: %v", errs[0])
		return 0, errs[0]
	}
	metrics.ReportLibRequestMetric("job", "ExecuteScript", "http", metrics.LibCallStatusOK, start)

	if !resp.Result || resp.Code != 0 {
		blog.Errorf("call api blueking Job ExecuteScript failed: %v", resp)
		return 0, fmt.Errorf("call api blueking Job ExecuteScript failed: %v", resp.Message)
	}

	blog.Infof("call api blueking Job ExecuteScript with url(%s) successfully", path)

	return resp.Data.JobInstanceID, nil
}

// GetJobStatus get job status
func (c *Client) GetJobStatus(ctx context.Context, job JobInfo) (int, error) {
	var (
		_    = "GetJobStatus"
		path = fmt.Sprintf("%s/api/v3/get_job_instance_status/", c.Server)
	)

	if options.GetEditionInfo().IsCommunicationEdition() {
		path = fmt.Sprintf("%s/api/c/compapi/v2/jobv3/get_job_instance_status/", c.Server)
	}

	resp := &GetJobInstanceStatusRsp{}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(utils.DefaultTimeOut).
		Get(path).
		Query(fmt.Sprintf("bk_scope_type=%s", string(Biz))).
		Query(fmt.Sprintf("bk_scope_id=%s", job.BizID)).
		Query(fmt.Sprintf("job_instance_id=%v", job.JobID)).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.Debug).
		EndStruct(&resp)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("job", "GetJobStatus", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api Job GetJobStatus failed: %v", errs[0])
		return UnKnownStatus, errs[0]
	}
	metrics.ReportLibRequestMetric("job", "GetJobStatus", "http", metrics.LibCallStatusOK, start)

	if !resp.Result || resp.Code != 0 {
		blog.Errorf("call api Job GetJobStatus failed: %v", resp)
		return UnKnownStatus, fmt.Errorf("call api Job GetJobStatus failed: %v", resp.Message)
	}
	blog.Infof("call api Job GetJobStatus with url(%s) successfully", path)

	return transJobStatus(resp.Data.JobInstance.Status), nil
}

// GetJobTaskLink xxx
func GetJobTaskLink(instanceId uint64) string {
	return fmt.Sprintf(options.GetGlobalCMOptions().Job.JobTaskLink, instanceId)
}
