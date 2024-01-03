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

package service

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
)

// BaseResp defines repo base response struct.
type BaseResp struct {
	Code    int32       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	TraceID string      `json:"traceId"`
}

// Err write err information to base repo, and then write base repo to ResponseWriter.
func (r *BaseResp) Err(w http.ResponseWriter, err error) {
	logs.Errorf("writer err to base response, err: %v", err)

	ef := errf.Error(err)
	r.Code = ef.Code
	r.Message = ef.Message

	marshal, err := jsoni.Marshal(r)
	if err != nil {
		logs.Errorf("marshal base response failed, err: %v", err)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write(marshal)
	return
}

// WriteResp write base repo to ResponseWriter with result data.
func (r *BaseResp) WriteResp(w http.ResponseWriter, data interface{}) {
	r.Data = data

	marshal, err := jsoni.Marshal(r)
	if err != nil {
		logs.Errorf("marshal base response failed, err: %v", err)
		return
	}

	w.Write(marshal)
	return
}

func unmarshal(body io.ReadCloser, data interface{}) error {
	bodyByte, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	if len(bodyByte) == 0 {
		return errf.New(errf.InvalidParameter, "request param is nil")
	}

	if err = jsoni.Unmarshal(bodyByte, data); err != nil {
		return err
	}

	return nil
}

// NodeMetadata define node metadata struct.
type NodeMetadata struct {
	BizID uint32   `json:"biz_id"`
	AppID []uint32 `json:"app_id"`
}

func parseMetadata(data string) (uint32, []uint32, error) {
	str, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return 0, nil, err
	}

	element := strings.Split(string(str), "&")
	if len(element) != 2 {
		return 0, nil, fmt.Errorf("metadata length should be 2")
	}

	bizEle := strings.Split(element[0], "=")
	if len(bizEle) != 2 {
		return 0, nil, fmt.Errorf("metadata %s biz info not right", element[0])
	}

	bizID, err := strconv.ParseInt(strings.TrimSpace(bizEle[1]), 10, 64)
	if err != nil {
		return 0, nil, fmt.Errorf("parse biz id failed, err: %v", err)
	}

	appEle := strings.Split(element[1], "=")
	if len(appEle) != 2 {
		return 0, nil, fmt.Errorf("metadata %s app info not right", element[0])
	}

	appIDs := make([]uint32, 0)
	if err = jsoni.Unmarshal([]byte(strings.TrimSpace(appEle[1])), &appIDs); err != nil {
		return 0, nil, err
	}

	return uint32(bizID), appIDs, nil
}

// UploadNodeRespData define upload node response data struct.
type UploadNodeRespData struct {
	CreateBy         string      `json:"createdBy"`
	CreateDate       string      `json:"createdDate"`
	LastModifiedBy   string      `json:"lastModifiedBy"`
	LastModifiedData string      `json:"lastModifiedDate"`
	Folder           string      `json:"folder"`
	Path             string      `json:"path"`
	Name             string      `json:"name"`
	FullPath         string      `json:"fullPath"`
	Size             uint64      `json:"size"`
	Sha256           uint64      `json:"sha256"`
	MD5              string      `json:"md5"`
	Metadata         interface{} `json:"metadata"`
	ProjectID        string      `json:"projectId"`
	RepoName         string      `json:"repoName"`
}
