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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
)

// execute path
const (
	executePath = "/prod/v4/apigw-app/projects/%s/build_start?pipelineId=%s"
)

// header key
const (
	headerDevopsUID           = "X-DEVOPS-UID"
	headerAuthorization       = "X-Bkapi-Authorization"
	headerAuthorizationFormat = `{"bk_app_code":"%s","bk_app_secret":"%s","bk_username":"%s"}`
)

// request params key
const (
	tokenKey       = "token"
	bizIdKey       = "biz_Id"
	enableGroupKey = "enable_group"
	collectionKey  = "collection"
)

// PipelineConfig 全局配置结构体
type PipelineConfig struct {
	BKDevOpsUrl     string
	AppCode         string
	AppSecret       string
	DevopsProjectID string
	DevopsUID       string
	BkUsername      string
	DevOpsToken     string
	BizID           int64
	Collection      string
	EnableGroup     bool
	PipelineID      string
	Enable          bool
}

// 全局变量
var (
	pipelineConfig *PipelineConfig
	configOnce     sync.Once
	configMutex    sync.RWMutex
)

// InitPipelineConfig 初始化全局配置，只初始化一次
func InitPipelineConfig(config *PipelineConfig) {
	configOnce.Do(func() {
		configMutex.Lock()
		defer configMutex.Unlock()
		pipelineConfig = config
	})
}

// GetPipelineConfig 获取全局配置
func GetPipelineConfig() *PipelineConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return pipelineConfig
}

// executeResp execute workflow response
type executeResp struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID           string `json:"id"`
		ExecuteCount int    `json:"executeCount"`
		ProjectID    string `json:"projectId"`
		PipelineID   string `json:"pipelineId"`
		Num          int64  `json:"num"`
	} `json:"data"`
}

// HTTPRequest defines the http request
type HTTPRequest struct {
	Url         string
	Method      string
	QueryParams map[string]string
	Body        interface{}
	Header      map[string]string
}

// Send the http request
func Send(ctx context.Context, hr *HTTPRequest) ([]byte, error) {
	var req *http.Request
	var err error

	if hr.Body != nil {
		var body []byte
		body, err = json.Marshal(hr.Body)
		if err != nil {
			return nil, errors.Wrapf(err, "marshal body failed")
		}
		req, err = http.NewRequestWithContext(ctx, hr.Method, hr.Url, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, hr.Method, hr.Url, nil)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "create http request failed")
	}

	for k, v := range hr.Header {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")

	if hr.QueryParams != nil {
		query := req.URL.Query()
		for k, v := range hr.QueryParams {
			query.Set(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "http request failed")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body failed")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("http response code not 200 but %d, resp: %s",
			resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// ExecutePipeline 执行pipeline
func ExecutePipeline(ctx context.Context) error {
	config := GetPipelineConfig()
	if config == nil || !config.Enable {
		return nil
	}

	// 构建请求URL
	url := config.BKDevOpsUrl + fmt.Sprintf(executePath, config.DevopsProjectID, config.PipelineID)

	// 构建请求头
	headers := map[string]string{
		headerDevopsUID: config.DevopsUID,
		headerAuthorization: fmt.Sprintf(headerAuthorizationFormat,
			config.AppCode, config.AppSecret, config.BkUsername),
	}

	// 构建请求参数
	requestParams := map[string]interface{}{
		tokenKey:       config.DevOpsToken,
		bizIdKey:       config.BizID,
		enableGroupKey: config.EnableGroup,
		collectionKey:  config.Collection,
	}

	// 发送HTTP请求
	respBody, err := Send(ctx, &HTTPRequest{
		Url:    url,
		Method: http.MethodPost,
		Header: headers,
		Body:   requestParams,
	})
	if err != nil {
		blog.Errorf("failed to execute pipeline (project: %s, pipeline: %s), err: %s",
			config.DevopsProjectID, config.PipelineID, err.Error())
		return fmt.Errorf("failed to execute pipeline (project: %s, pipeline: %s), err: %s",
			config.DevopsProjectID, config.PipelineID, err.Error())
	}

	// 解析响应
	resp := new(executeResp)
	if err = json.Unmarshal(respBody, resp); err != nil {
		blog.Errorf("failed to unmarshal pipeline response (project: %s, pipeline: %s), err: %s",
			config.DevopsProjectID, config.PipelineID, err.Error())
		return fmt.Errorf("failed to unmarshal pipeline response (project: %s, pipeline: %s), err: %s",
			config.DevopsProjectID, config.PipelineID, err.Error())
	}

	// 检查响应状态
	if resp.Status != 0 {
		blog.Errorf("pipeline execution failed (project: %s, pipeline: %s) with status %d: %s",
			config.DevopsProjectID, config.PipelineID, resp.Status, resp.Message)
		return fmt.Errorf("pipeline execution failed (project: %s, pipeline: %s) with status %d: %s",
			config.DevopsProjectID, config.PipelineID, resp.Status, resp.Message)
	}

	return nil
}
