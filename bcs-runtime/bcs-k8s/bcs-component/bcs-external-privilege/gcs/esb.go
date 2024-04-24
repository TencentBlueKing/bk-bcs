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

package gcs

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/TencentBlueKing/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-external-privilege/common"
)

func PrivilegeRequest(opt *common.Option, env common.DBPrivEnv) (string, error) {
	payload := make(map[string]interface{})

	payload["app"] = env.AppName
	payload["target_ip"] = env.TargetDb
	payload["type"] = env.CallType
	payload["client_ip"] = opt.PrivilegeIP
	payload["call_user"] = env.CallUser
	payload["db_name"] = env.DbName
	payload["dbname"] = env.DbName
	if env.Operator != "" {
		payload["operator"] = env.Operator
	}
	blog.Infof("request payload: %v", payload)

	var privUrl string
	if env.UseCDP {
		privUrl = opt.CDPGCSUrl + "xxxxxx"
	} else {
		privUrl = opt.ESBUrl + "xxxxxxxx"
	}

	req := opt.RequestESB.PrepareRequest("POST", privUrl, payload)
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error doing request to privilege: %s", err.Error())
	}
	defer resp.Body.Close()

	// Parse body as JSON
	respBody, _ := ioutil.ReadAll(resp.Body)
	blog.Infof("resp body: %v", string(respBody))
	result, err := bodyToPrivAPIResponse(respBody)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("invalid status code %d", resp.StatusCode)
	}
	if err != nil {
		return "", fmt.Errorf("non-Json response(%s): %s", string(respBody), err.Error())
	}
	if !result.Result {
		return "", fmt.Errorf("failed to request the api to privilege, response code: %d, response message: %s", result.Code, result.Message)
	}

	blog.Info(result.Data.JobId)

	blog.Infof("request the privilege api successful, continue...")
	return result.Data.JobId, nil
}

func CheckFinalStatus(opt *common.Option, jobId string, failRetryLimit int) error {

	payload := make(map[string]interface{})
	payload["job_id"] = jobId
	blog.Infof("request payload: %v", payload)

	var req *http.Request
	checkUrl := opt.ESBUrl + "xxxxxxxxxxx"
	req = opt.RequestESB.PrepareRequest("POST", checkUrl, payload)
	httpClient := &http.Client{}

	var failRetry = 0

	for i := 0; i < 40 && failRetry < failRetryLimit; i++ {
		common.WaitForSeveralSeconds()
		data, err := checkStatus(httpClient, req)
		if err != nil {
			blog.Errorf("Check job status failed: %s, retry %d", err.Error(), failRetry)
			failRetry++
			continue
		}
		if data.Code == 0 {
			blog.Infof("job succeed")
			return nil
		} else if data.Code == 1 {
			blog.Infof("job is still doing, check the status again...")
			continue
		} else {
			blog.Errorf("job return error: %s, retry %d", data.Message, failRetry)
			failRetry++
			continue
		}
	}
	return fmt.Errorf("timeout or failed to check job status")
}

func checkStatus(httpClient *http.Client, req *http.Request) (*CheckStatusData, error) {
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request to check status: %s, err response: %+v", err.Error(), resp)
	}
	defer resp.Body.Close()

	// Parse body as JSON
	respBody, _ := ioutil.ReadAll(resp.Body)
	result, err := bodyToCheckAPIResponse(respBody)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code %d", resp.StatusCode)
	}
	if err != nil {
		return nil, fmt.Errorf("non-Json response(%s): %s", string(respBody), err.Error())
	}
	if !result.Result {
		return nil, fmt.Errorf("failed to request the api to check status, response code: %d, response: %s", result.Code, string(respBody))
	}

	return &result.Data, nil
}
