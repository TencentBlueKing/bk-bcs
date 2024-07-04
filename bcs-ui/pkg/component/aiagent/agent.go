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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
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
func Assistant(ctx context.Context, bk_ticket, role, input string) (string, error) {
	if !config.G.BKAIAgent.Enable {
		return "", errors.New("assistant is not enabled")
	}
	for _, v := range config.G.BKAIAgent.Assistants {
		if v.Role == role {
			return BKAssistant(ctx, bk_ticket, v.Prompt, input)
		}
	}
	return "", errors.New("assistant not found")
}

// BKAssistant ai assistant
func BKAssistant(ctx context.Context, bk_ticket, prompt, input string) (string, error) {

	out := &AssistantResponse{}

	url := fmt.Sprintf("%s/%s", strings.TrimRight(config.G.BKAIAgent.Host, "/"),
		strings.TrimLeft(config.G.BKAIAgent.Path, "/"))

	body := map[string]interface{}{
		"inputs": map[string]interface{}{
			"input": input,
			"chat_history": []map[string]interface{}{
				{"role": "user", "content": prompt},
			},
		},
		"context": map[string]interface{}{
			"executor": "bcs",
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
