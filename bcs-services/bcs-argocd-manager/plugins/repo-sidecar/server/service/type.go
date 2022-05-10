/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"bytes"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
)

// Message define the structure for client sending data to server.
type Message struct {
	Env     []string `json:"env"`
	Args    []string `json:"args"`
	Content string   `json:"content"`
}

// Request to repo-sidecar server for Result
func (m *Message) Request() (*Result, error) {
	var data []byte
	_ = codec.EncJson(m, &data)
	resp, err := http.Post("http://127.0.0.1:"+SvcPort, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	var r Result
	if err := codec.DecJsonReader(resp.Body, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

// GetEnv return env value by key
func (m *Message) GetEnv(key string) string {
	for _, content := range m.Env {
		for i := 0; i < len(content); i++ {
			if content[i] == '=' {
				if content[:i] != key {
					break
				}

				return content[i+1:]
			}
		}
	}

	return ""
}

// Result define the structure for server returning data back to client
type Result struct {
	// Code if 0 then mean success, or failure
	Code ResultErrorCode `json:"code"`

	// Message describe the error reasons
	Message string `json:"message"`

	Data []byte `json:"data"`
}

type ResultErrorCode int

const (
	ResultErrorCodeSuccess ResultErrorCode = iota
	ResultErrorCodeFailure
)
