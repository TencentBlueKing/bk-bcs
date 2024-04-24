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
 *
 */

package scr

import (
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-external-privilege/common"
)

func getKeyFromDBPrivEnv(env common.DBPrivEnv) string {
	return fmt.Sprintf("%s,%s", env.AppName, env.Operator)
}

func PrivilegeRequest(opt *common.Option, env common.DBPrivEnv) (string, error) {
	if env.TableName == "" {
		env.TableName = "*"
	}
	req := &ApplyRequest{
		App:      env.AppName,
		User:     env.Operator,
		Describe: "mysql privilege by bcs",
		ApplyInfo: []ApplyInfo{
			{
				ClientVersion:       "5",
				DbUser:              env.CallUser,
				DbPassword:          env.Password,
				DbName:              env.DbName,
				TbName:              env.TableName,
				Grants:              env.Grants,
				SourceIPInput:       opt.PrivilegeIP,
				TargetInstanceInput: env.TargetDb,
			},
		},
	}
	resp := &SCRResponse{}

	header := make(http.Header)
	header.Add("Authorization", fmt.Sprintf("Bearer %s", opt.SCRToken))
	header.Add("Content-Type", "application/json")
	err := client.NewRESTClient().Post().
		WithHeaders(header).
		WithEndpoints([]string{opt.SCRURL}).
		WithBasePath("/xxxxxx/xxxxxxx/apply").
		WithJSON(req).Do().Into(resp)
	if err != nil {
		return "", fmt.Errorf("error doing request to privilege using scr api: %s", err.Error())
	}
	blog.Infof("Response: %+v", *resp)
	if resp.Code != 0 {
		return "", fmt.Errorf("error response from scr privilege api: %s", resp.Msg)
	}
	return resp.Jobid, nil
}

func CheckFinalStatus(opt *common.Option, jobId string, failRetryLimit int) error {

	header := make(http.Header)
	header.Add("Authorization", fmt.Sprintf("Bearer %s", opt.SCRToken))
	header.Add("Content-Type", "application/json")
	req := client.NewRESTClient().Get().
		WithHeaders(header).
		WithEndpoints([]string{opt.SCRURL}).
		WithBasePath("/xxxxxx/xxxxxxx/").
		SubPathf("%s", jobId)

	var failRetry = 0
	resp := &SCRResponse{}

	for i := 0; i < 40 && failRetry < failRetryLimit; i++ {
		common.WaitForSeveralSeconds()
		err := req.Do().Into(resp)
		if err != nil {
			blog.Errorf("Check job status failed: %s, retry %d", err.Error(), failRetry)
			failRetry++
			continue
		}
		if resp.Code == 0 {
			blog.Infof("job succeed")
			return nil
		} else if resp.Code == 1 {
			blog.Infof("job is still doing, check the status again...")
			continue
		} else {
			blog.Errorf("job return error: %s, retry %d", resp.Msg, failRetry)
			failRetry++
			continue
		}
	}
	return fmt.Errorf("timeout or failed to check job status")
}
