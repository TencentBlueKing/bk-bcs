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

package grpcGateway

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/grpc"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
)

func getFormFile(r *http.Request, name string) ([]byte, error) {
	file, _, err := r.FormFile(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	buf := bytes.Buffer{}
	io.Copy(&buf, file)
	if err != nil {
		return nil, fmt.Errorf("error while reading form file %s", err.Error())
	}
	return buf.Bytes(), nil
}

func createRequestFromMultiPart(r *http.Request) (*http.Request, error) {
	data, err := getFormFile(r, "file")
	if err != nil {
		return nil, err
	}
	content := base64.StdEncoding.EncodeToString(data)
	force := r.FormValue("force")
	// 将字符串 "true" 转换为 bool 类型
	isForce, err := strconv.ParseBool(force)
	if err != nil {
		return nil, err
	}
	param := map[string]interface{}{
		"force": isForce,
		"file":  content,
	}
	// 转成json格式
	marshal, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(marshal)
	newR, err := http.NewRequest(r.Method, r.URL.String(), reader)
	if err != nil {
		return nil, err
	}
	return newR, nil
}

// GRPCHandlerFunc  grpc-gateway 处理 Content-Type 类型为 multipart/form-data
func GRPCHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
				newR, err := createRequestFromMultiPart(r)
				if err != nil {
					errorResponse := map[string]interface{}{
						"Code":      common.ErrHelmManagerUploadChartFailed,
						"Message":   err.Error(),
						"Result":    false,
						"Data":      nil,
						"requestID": "",
					}
					errorJSON, _ := json.Marshal(errorResponse)
					w.WriteHeader(http.StatusBadRequest)
					w.Header().Set("Content-Type", "application/json")
					w.Write(errorJSON)
					return
				}
				otherHandler.ServeHTTP(w, newR)
			} else {
				otherHandler.ServeHTTP(w, r)
			}
		}
	})
}
