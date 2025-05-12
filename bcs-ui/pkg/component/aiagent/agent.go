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

// Package aiagent xxx
package aiagent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/constants"
)

// AssistantResponse resp
type AssistantResponse struct {
	Result  bool                  `json:"result"`
	Message string                `json:"message"`
	TraceID string                `json:"trace_id"`
	Data    AssistantResponseData `json:"data"`
}

// AssistantResponseData xxx
type AssistantResponseData struct {
	Ouputs struct {
		Output string `json:"output"`
	} `json:"outputs"`
}

// Assistant ai assistant
func Assistant(ctx context.Context, bk_ticket, role, input, username string, chatHistory []map[string]interface{}) (
	interface{}, error) {
	if !config.G.BKAIAgent.Enable {
		return "", errors.New("assistant is not enabled")
	}
	for _, v := range config.G.BKAIAgent.Assistants {
		if v.Role == role {
			chatHistory = append([]map[string]interface{}{
				{"role": "user", "content": v.Prompt},
				// hunyuan 必须要有一对对话，因此这里加一条历史记录
				{"role": "assistant", "content": "ok"},
			}, chatHistory...)
			return BKAssistantStream(ctx, bk_ticket, input, username, chatHistory)
		}
	}
	return "", errors.New("assistant not found")
}

// BKAssistant ai assistant
func BKAssistant(ctx context.Context, bk_ticket, prompt, input, username string) (string, error) {

	out := &AssistantResponse{}

	url := fmt.Sprintf("%s/%s", strings.TrimRight(config.G.BKAIAgent.Host, "/"),
		strings.TrimLeft(config.G.BKAIAgent.Path, "/"))

	body := map[string]interface{}{
		"inputs": map[string]interface{}{
			"input": input,
			"chat_history": []map[string]interface{}{
				{"role": "user", "content": prompt},
				// hunyuan 必须要有一对对话，因此这里加一条历史记录
				{"role": "assistant", "content": "ok"},
			},
		},
		"context": map[string]interface{}{
			"executor": username,
		},
	}

	authHeader := fmt.Sprintf("{\"bk_ticket\": \"%s\", \"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\"}",
		bk_ticket, config.G.Base.AppCode, config.G.Base.AppSecret)
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authHeader).
		SetBody(body).
		Post(url)

	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(resp.Body(), out); err != nil {
		return "", err
	}
	if !out.Result {
		klog.Errorf("request bk_ai_agent failed, trace_id %s, message %s", out.TraceID, out.Message)
		return "", errors.New(out.Message)
	}
	return out.Data.Ouputs.Output, nil
}

// BKAssistantStream ai assistant stream
func BKAssistantStream(ctx context.Context, bk_ticket, input, username string, chatHistory []map[string]interface{}) (
	io.ReadCloser, error) {

	url := fmt.Sprintf("%s/%s", strings.TrimRight(config.G.BKAIAgent.Host, "/"),
		strings.TrimLeft(config.G.BKAIAgent.StreamPath, "/"))

	body := map[string]interface{}{
		"inputs": map[string]interface{}{
			"input":        input,
			"chat_history": chatHistory,
		},
		"context": map[string]interface{}{
			"executor": username,
		},
	}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	authHeader := fmt.Sprintf("{\"bk_ticket\": \"%s\", \"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\"}",
		bk_ticket, config.G.Base.AppCode, config.G.Base.AppSecret)
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Bkapi-Authorization", authHeader)
	req.Header.Set(constants.HeaderTenantId, ctx.Value(constants.TenantIdCtxKey).(string))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("X-Accel-Buffering", "no")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
