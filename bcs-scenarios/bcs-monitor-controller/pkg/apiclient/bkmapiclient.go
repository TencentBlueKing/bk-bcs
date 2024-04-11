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

package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/option"
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
	envNameBKMFullAuthToken = "BKM_FULL_AUTH_TOKEN" // nolint
	envNameBKMAPIDomain     = "BKM_API_DOMAIN"
)

// IMonitorApiClient translate monitor crd to yaml file
type IMonitorApiClient interface {
	UploadConfig(bizID, bizToken, configPath, app string, overwrite bool) error
	DownloadConfig(bizID, bizToken string) error
}

// AsyncTaskResp response of upload config
type AsyncTaskResp struct {
	Result  bool     `json:"result,omitempty"`
	Code    int      `json:"code,omitempty"`
	Message string   `json:"message,omitempty"`
	Data    TaskData `json:"data,omitempty"`
}

// TaskData result for async task
type TaskData struct {
	TaskID string `json:"task_id,omitempty"`
}

// PollDownloadConfigTaskStatusResp response of PollDownloadTaskStatus
type PollDownloadConfigTaskStatusResp struct {
	Result  bool                             `json:"result,omitempty"`
	Code    int                              `json:"code,omitempty"`
	Message string                           `json:"message,omitempty"`
	Data    PollDownloadConfigTaskStatusData `json:"data,omitempty"`
}

// PollDownloadConfigTaskStatusData data struct of PollDownloadConfigTaskStatusResp
type PollDownloadConfigTaskStatusData struct {
	IsCompleted bool                   `json:"is_completed,omitempty"`
	Message     string                 `json:"message,omitempty"`
	State       string                 `json:"state,omitempty"`
	TaskID      string                 `json:"task_id,omitempty"`
	Traceback   string                 `json:"traceback,omitempty"`
	Data        DownloadConfigTaskData `json:"data,omitempty"`
}

// DownloadConfigTaskData data struct of DownloadConfigResp
type DownloadConfigTaskData struct {
	DownloadUrl string `json:"download_url"`
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
	httpCli       http.Client

	SubPath string
	Opts    *option.ControllerOption
}

// NewBkmApiClient return new bkm apli client
func NewBkmApiClient(subPath string, opts *option.ControllerOption) *BkmApiClient {
	return &BkmApiClient{
		httpCli:       http.Client{},
		FullAuthToken: os.Getenv(envNameBKMFullAuthToken),
		MonitorURL:    os.Getenv(envNameBKMAPIDomain),
		Opts:          opts,

		SubPath: subPath,
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
		return fmt.Errorf("open tar file'%s' failed, err: %w", configPath, err)
	}
	defer request.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(configPath))
	_, _ = io.Copy(part, request)

	_ = writer.WriteField(fieldBkBizID, bizID)
	_ = writer.WriteField(fieldApp, app)
	_ = writer.WriteField(fieldOverwrite, strconv.FormatBool(overwrite))
	writer.Close() // nolint not checked

	url := fmt.Sprintf("%s/rest/v2/as_code/import_config_file/", b.MonitorURL)
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", bizToken))
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("X-Async-Task", "True")

	resp, err := b.httpCli.Do(req)
	if err != nil {
		mf(StatusErr)
		return fmt.Errorf("do post request failed, req: %v, err: %w", req, err)
	}
	defer resp.Body.Close()

	// Read the response
	uploadConfigResp := &AsyncTaskResp{}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		mf(StatusErr)
		return fmt.Errorf("read UploadConfig from resp failed, err: %w", err)
	}
	err = json.Unmarshal(respBody, uploadConfigResp)
	if err != nil {
		mf(StatusErr)
		return fmt.Errorf("json marshal failed, raw value: %s, err: %w", string(respBody), err)
	}

	return b.pollTaskStatus(bizToken, bizID, uploadConfigResp.Data.TaskID, b.doPollUploadTaskStatus, mf)
}

// DownloadConfig download biz related bkm config to ../config_{biz_id}.tar.gz
func (b *BkmApiClient) DownloadConfig(bizID, bizToken string) error {
	if b.FullAuthToken != "" {
		bizToken = b.FullAuthToken
	}

	blog.Infof("download config req[bizID: %s]", bizID)
	startTime := time.Now()
	mf := func(ret string) {
		ReportAPIRequestMetric(HandlerBKM, "DownloadConfig", ret, startTime)
	}

	url := fmt.Sprintf("%s/rest/v2/as_code/export_config_file/", b.MonitorURL)
	reqParams := map[string]interface{}{
		"bk_biz_id":              bizID,
		"dashboard_for_external": false,
	}
	bts, _ := json.Marshal(reqParams)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bts))
	if err != nil {
		mf(StatusErr)
		return fmt.Errorf("http new request failed: %w", err)
	}
	defer req.Body.Close()
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bizToken))
	req.Header.Set("X-Async-Task", "True")
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	resp, err := b.httpCli.Do(req)
	if err != nil {
		mf(StatusErr)
		return fmt.Errorf("http post failed: %w", err)
	}
	defer resp.Body.Close()

	var downloadConfigResp AsyncTaskResp
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		mf(StatusErr)
		return fmt.Errorf("read resp failed")
	}
	err = json.Unmarshal(respBody, &downloadConfigResp)
	if err != nil {
		mf(StatusErr)
		return fmt.Errorf("json marshal failed, raw value: %s, err: %w", string(respBody), err)
	}
	if downloadConfigResp.Code != http.StatusOK || !downloadConfigResp.Result {
		mf(StatusErr)
		return fmt.Errorf("error status, resp: %s", string(respBody))
	}
	// print(string(respBody))
	err = b.pollTaskStatus(bizToken, bizID, downloadConfigResp.Data.TaskID, b.doPollDownloadTaskStatus, mf)
	if err != nil {
		mf(StatusErr)
		return fmt.Errorf("epollTaskStatus failed, bizID[%s], taskID[%s], err: %w", bizID,
			downloadConfigResp.Data.TaskID, err)
	}

	return nil
}

func (b *BkmApiClient) pollTaskStatus(token, bizID, taskID string,
	handleFunc func(string, *http.Response) (bool, error), metricFunc func(string)) error {
	blog.Infof("start poll task '%s/%s' status", bizID, taskID)

	startTime := time.Now()
	for {
		url := fmt.Sprintf("%s/rest/v2/commons/query_async_task_result/?bk_biz_id=%s&task_id=%s", b.MonitorURL, bizID,
			taskID)
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create GET request: %w", err)
		}

		req.Header.Add("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send GET request: %w", err)
		}

		success, err := handleFunc(bizID, resp)
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
func (b *BkmApiClient) doPollUploadTaskStatus(bizID string, resp *http.Response) (bool, error) {
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("get err status code: %d, message: %s", resp.StatusCode, bodyString)
	}

	var pollTaskStatusResp PollTaskStatusResp
	if err := json.Unmarshal(bodyBytes, &pollTaskStatusResp); err != nil {
		return false, fmt.Errorf("decode JSON response failed: %w", err)
	}

	blog.V(4).Infof("query task resp: %+v", pollTaskStatusResp)

	if !pollTaskStatusResp.Result {
		return false, fmt.Errorf("unknown result, resp: %+v", pollTaskStatusResp)
	}

	pollTaskStatusData := pollTaskStatusResp.Data
	if pollTaskStatusData.State == taskStateSuccess {
		if pollTaskStatusData.Data.Result {
			return true, nil
		}
		return false, fmt.Errorf("task failed, message: %s, err: %s", pollTaskStatusData.Data.Message,
			utils.ToJsonString(pollTaskStatusData.Data.Errors))
	} else if pollTaskStatusData.State == taskStateFailure {
		return false, fmt.Errorf("task failed, message: %s", pollTaskStatusData.Message)
	} else if pollTaskStatusResp.Data.IsCompleted {
		return false, fmt.Errorf("task unknown state, message: %v", pollTaskStatusData)
	}

	// Log data state and message
	blog.Infof("waiting task finish, state %s,message %s", pollTaskStatusData.State,
		pollTaskStatusData.Message)

	return false, nil
}

func (b *BkmApiClient) doPollDownloadTaskStatus(bizID string, resp *http.Response) (bool, error) {
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("get err status code: %d, message: %s", resp.StatusCode, bodyString)
	}

	var pollTaskStatusResp PollDownloadConfigTaskStatusResp
	if err := json.Unmarshal(bodyBytes, &pollTaskStatusResp); err != nil {
		return false, fmt.Errorf("decode JSON response failed: %w", err)
	}

	if !pollTaskStatusResp.Result {
		return false, fmt.Errorf("unknown result, resp: %+v", pollTaskStatusResp)
	}

	pollTaskStatusData := pollTaskStatusResp.Data
	if pollTaskStatusData.State != taskStateSuccess {
		if pollTaskStatusData.State == taskStateFailure {
			return false, fmt.Errorf("task failed, message: %s", pollTaskStatusData.Message)
		} else if pollTaskStatusResp.Data.IsCompleted {
			return false, fmt.Errorf("task unknown state, message: %v", pollTaskStatusData)
		}
		// blog.Infof("waiting task finish, state %s,message %s", pollTaskStatusData.State,
		// 	pollTaskStatusData.Message)
		return false, nil
	}

	downloadURL := strings.Replace(pollTaskStatusData.Data.DownloadUrl, "https", "http", 1)
	downloadResp, err := b.httpCli.Get(downloadURL)
	if err != nil {
		return false, fmt.Errorf("http get download file failed, err: %w", err)
	}
	defer downloadResp.Body.Close()

	if err = os.MkdirAll(filepath.Join(b.Opts.BKMDownloadConfigPath, b.SubPath), 0755); err != nil {
		return false, fmt.Errorf("mkdir failed, err: %w", err)
	}
	// 创建一个文件用于保存
	out, err := os.Create(utils.GenBkmConfigTarPath(b.Opts.BKMDownloadConfigPath, b.SubPath, bizID))
	if err != nil {
		return false, fmt.Errorf("create download file failed, err: %w", err)
	}
	defer out.Close()

	// 然后将响应流和文件流对接起来
	_, err = io.Copy(out, downloadResp.Body)
	if err != nil {
		return false, fmt.Errorf("copy download file failed, err: %w", err)
	}

	return true, nil
}
