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

package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/utils"
)

const (
	fieldBkBizID   = "bk_biz_id"
	fieldApp       = "app"
	fieldOverwrite = "overwrite"

	taskStateSuccess = "SUCCESS"
	taskStateFailure = "FAILURE"
)

const (
	envNameBKMFullAuthToken = "BKM_FULL_AUTH_TOKEN"
	envNameBKMAPIDomain     = "BKM_API_DOMAIN"
)

// IMonitorApiClient translate monitor crd to yaml file
type IMonitorApiClient interface {
	UploadConfig(bizID, bizToken, configPath, app string, overwrite bool) error
}

// UploadConfigResp response of upload config
type UploadConfigResp struct {
	Result  bool                 `json:"result,omitempty"`
	Code    int                  `json:"code,omitempty"`
	Message string               `json:"message,omitempty"`
	Data    UploadConfigTaskData `json:"data,omitempty"`
}

// UploadConfigTaskData data struct of UploadConfigResp
type UploadConfigTaskData struct {
	TaskID string `json:"task_id,omitempty"`
}

// PollTaskStatusResp response of PollTaskStatus
type PollTaskStatusResp struct {
	Result  bool               `json:"result,omitempty"`
	Code    int                `json:"code,omitempty"`
	Message string             `json:"message,omitempty"`
	Data    PollTaskStatusData `json:"data,omitempty"`
}

// PollTaskStatusData data struct of PollTaskStatusResp
type PollTaskStatusData struct {
	IsCompleted bool               `json:"is_completed,omitempty"`
	Message     string             `json:"message,omitempty"`
	State       string             `json:"state,omitempty"`
	TaskID      string             `json:"task_id,omitempty"`
	Traceback   string             `json:"traceback,omitempty"`
	Data        PollTaskResultData `json:"data,omitempty"`
}

// PollTaskResultData data struct of PollTaskStatusData
type PollTaskResultData struct {
	Result  bool              `json:"result,omitempty"`
	Data    interface{}       `json:"data,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
	Message string            `json:"message,omitempty"`
}

// BkmApiClient api client to call bk monitor
type BkmApiClient struct {
	MonitorURL    string
	FullAuthToken string
}

// NewBkmApiClient return new bkm apli client
func NewBkmApiClient() *BkmApiClient {
	return &BkmApiClient{
		FullAuthToken: os.Getenv(envNameBKMFullAuthToken),
		MonitorURL:    os.Getenv(envNameBKMAPIDomain),
	}
}

// UploadConfig upload monitor config to bkm
func (b *BkmApiClient) UploadConfig(bizID, bizToken, configPath, app string, overwrite bool) error {
	if b.FullAuthToken != "" {
		bizToken = b.FullAuthToken
	}
	blog.Infof("upload config req[bizID: %s, configPath: %s, app: %s]", bizID, configPath, app)
	startTime := time.Now()
	mf := func(ret string) {
		ReportAPIRequestMetric(HandlerBKM, "UploadConfig", ret, startTime)
	}

	request, err := os.Open(configPath)
	if err != nil {
		mf(StatusErr)
		blog.Errorf("open tar file'%s' failed, err: %s", configPath, err.Error())
		return err
	}
	defer request.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(configPath))
	_, _ = io.Copy(part, request)

	_ = writer.WriteField(fieldBkBizID, bizID)
	_ = writer.WriteField(fieldApp, app)
	_ = writer.WriteField(fieldOverwrite, strconv.FormatBool(overwrite))
	writer.Close()

	url := fmt.Sprintf("%s/rest/v2/as_code/import_config_file/", b.MonitorURL)
	req, err := http.NewRequest("POST", url, body)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", bizToken))
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("X-Async-Task", "True")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		mf(StatusErr)
		blog.Errorf("do post request failed, req: %v, err: %s", req, err.Error())
		return err
	}
	defer resp.Body.Close()

	// Read the response
	uploadConfigResp := &UploadConfigResp{}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		mf(StatusErr)
		blog.Errorf("read UploadConfig from resp failed, err: %s", err.Error())
		return err
	}
	err = json.Unmarshal(respBody, uploadConfigResp)
	if err != nil {
		mf(StatusErr)
		blog.Errorf("json marshal failed, raw value: %s, err: %s", string(respBody), err.Error())
		return err
	}

	return b.pollTaskStatus(bizToken, bizID, uploadConfigResp.Data.TaskID, mf)
}

// TaskData task response
type TaskData struct {
	State       string `json:"state,omitempty"`
	Message     string `json:"message,omitempty"`
	IsCompleted bool   `json:"is_completed,omitempty"`
}

func (b *BkmApiClient) pollTaskStatus(token, bizID, taskID string, metricFunc func(string)) error {
	blog.Infof("start poll task '%s/%s' status", bizID, taskID)

	startTime := time.Now()
	for {
		success, err := b.doPollTaskStatus(token, bizID, taskID)
		if err != nil {
			metricFunc(StatusErr)
			return err
		}
		if success {
			metricFunc(StatusOK)
			blog.Infof("task '%s/%s' success", bizID, taskID)
			return nil
		}
		// Check if polling time exceeds 10 minutes.
		if time.Since(startTime).Minutes() >= 10 {
			metricFunc(StatusTimeout)
			blog.Warnf("task'%s' timeout", taskID)
			return fmt.Errorf("task'%s' timeout", taskID)
		}

		// Sleep for 1 seconds before the next polling.
		time.Sleep(time.Second)
	}
}

// return true if task success
func (b *BkmApiClient) doPollTaskStatus(token, bizID, taskID string) (bool, error) {
	url := fmt.Sprintf("%s/rest/v2/commons/query_async_task_result/?bk_biz_id=%s&task_id=%s", b.MonitorURL, bizID,
		taskID)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		blog.Errorf("failed to create GET request: %s", err.Error())
		return false, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		blog.Errorf("failed to send GET request: %s", err.Error())
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		blog.Errorf("get err status code: %d, message: %s", resp.StatusCode, bodyString)
		return false, errors.Errorf("get err status code: %d, message: %s", resp.StatusCode, bodyString)
	}

	var pollTaskStatusResp PollTaskStatusResp
	if err = json.NewDecoder(resp.Body).Decode(&pollTaskStatusResp); err != nil {
		blog.Errorf("decode JSON response failed: %s", err.Error())
		return false, err
	}

	blog.V(4).Infof("query task'%s' resp: %+v", taskID, pollTaskStatusResp)

	if !pollTaskStatusResp.Result {
		blog.Errorf("unknown result, resp: %+v", pollTaskStatusResp)
		return false, fmt.Errorf("unknown result, resp: %+v", pollTaskStatusResp)
	}

	pollTaskStatusData := pollTaskStatusResp.Data
	if pollTaskStatusData.State == taskStateSuccess {
		if pollTaskStatusData.Data.Result {
			blog.Infof("task '%s' success", taskID)
			return true, nil
		}
		blog.Warnf("task '%s' failed, message: %s, err: %s", taskID, pollTaskStatusData.Data.Message,
			utils.ToJsonString(pollTaskStatusData.Data.Errors))
		return false, fmt.Errorf("task '%s' failed, message: %s, err: %s", taskID, pollTaskStatusData.Data.Message,
			utils.ToJsonString(pollTaskStatusData.Data.Errors))
	} else if pollTaskStatusData.State == taskStateFailure {
		blog.Errorf("task '%s' failed, message: %s", taskID, pollTaskStatusData.Message)
		return false, fmt.Errorf("task '%s' failed, message: %s", taskID, pollTaskStatusData.Message)
	} else if pollTaskStatusResp.Data.IsCompleted {
		blog.Errorf("task '%s' unknown state, resp: %v", taskID, pollTaskStatusData)
		return false, fmt.Errorf("task '%s' unknown state, message: %v", taskID, pollTaskStatusData)
	}

	// Log data state and message
	blog.Infof("waiting task'%s' finish, state %s,message %s", taskID, pollTaskStatusData.State,
		pollTaskStatusData.Message)

	return false, nil
}
