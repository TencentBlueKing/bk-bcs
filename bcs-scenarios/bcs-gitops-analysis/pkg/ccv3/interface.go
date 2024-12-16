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

// Package ccv3 xxx
package ccv3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	searchBusinessApi     = "/api/c/compapi/v2/cc/search_business/"
	getStaffInfoApi       = "/component/compapi/tof/get_staff_info/"
	getParentDeptInfosApi = "/component/compapi/tof/get_parent_dept_infos/"
)

// Interface defines the interface to bkcc
type Interface interface {
	SearchBusiness(bkBizIds []int64) ([]CCBusiness, error)
	GetUserDeptInfo(user string) (*UserDeptInfo, error)
	GetBizDeptInfo(bkBizIDs []int64) (map[int64]*BusinessDeptInfo, error)
}

// UserDeptInfo user details
type UserDeptInfo struct {
	UserName    string `json:"userName"`
	ChineseName string `json:"chineseName"`

	Level0 string `json:"level0"`
	Level1 string `json:"level1"`
	Level2 string `json:"level2"`
	Level3 string `json:"level3"`
	Level4 string `json:"level4"`
	Level5 string `json:"level5"`
}

// BusinessDeptInfo defines the business dept info
type BusinessDeptInfo struct {
	BKBizID   int64  `json:"bk_biz_id"`
	BKBizName string `json:"bk_biz_name"`

	Level0 string `json:"level0"`
	Level1 string `json:"level1"`
	Level2 string `json:"level2"`
	Level3 string `json:"level3"`
	Level4 string `json:"level4"`
	Level5 string `json:"level5"`
}

var bkAuthFormat = `{"bk_app_code": "%s", "bk_app_secret": "%s", "bk_token": "%s", "bk_username": "%s"}`

func (h *handler) query(url string, method string, body interface{}, result interface{}) error {
	var req *http.Request
	var err error
	if body != nil {
		var bodyBytes []byte
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return errors.Wrapf(err, "marshal body failed")
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return errors.Wrapf(err, "create request failed")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Bkapi-Authorization", fmt.Sprintf(bkAuthFormat, h.op.Auth.AppCode,
		h.op.Auth.AppSecret, "", "admin"))
	c := &http.Client{
		Timeout: time.Second * 20,
	}
	httpResponse, err := c.Do(req)
	if err != nil {
		return errors.Wrapf(err, "do http request failed")
	}
	defer httpResponse.Body.Close()
	respBytes, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return errors.Wrapf(err, "read http response failed")
	}
	if err = json.Unmarshal(respBytes, result); err != nil {
		return errors.Wrapf(err, "unmarshal failed")
	}
	return nil
}

type deptInfoObject struct {
	Level0 string `json:"level0"`
	Level1 string `json:"level1"`
	Level2 string `json:"level2"`
	Level3 string `json:"level3"`
	Level4 string `json:"level4"`
	Level5 string `json:"level5"`
}

type getParentDeptInfosResponse struct {
	Code      string                    `json:"code"`
	Result    bool                      `json:"result"`
	RequestId string                    `json:"request_id"`
	Message   string                    `json:"message"`
	Data      []*getParentDeptInfosData `json:"data"`
}

type getParentDeptInfosData struct {
	Level string `json:"Level"`
	Name  string `json:"Name"`
}

func (h *handler) getDeptInfo(deptID int64, deptName string) (*deptInfoObject, error) {
	getDeptInfosUrl := fmt.Sprintf("%s%s?dept_id=%d&level=10", h.op.BKCCUrl, getParentDeptInfosApi, deptID)
	getDeptInfosResp := new(getParentDeptInfosResponse)
	if err := h.query(getDeptInfosUrl, http.MethodGet, nil, getDeptInfosResp); err != nil {
		return nil, errors.Wrapf(err, "get dept-info '%d' failed", deptID)
	}
	deptInfosData := getDeptInfosResp.Data
	if len(deptInfosData) == 0 {
		return nil, fmt.Errorf("get dept-infos groupID '%d' not found", deptID)
	}

	result := &deptInfoObject{}
	for _, deptInfo := range deptInfosData {
		switch deptInfo.Level {
		case "0":
			result.Level0 = deptInfo.Name
		case "1":
			result.Level1 = deptInfo.Name
		case "2":
			result.Level2 = deptInfo.Name
		case "3":
			result.Level3 = deptInfo.Name
		case "4":
			result.Level4 = deptInfo.Name
		case "5":
			result.Level5 = deptInfo.Name
		}
	}
	lastDept := deptInfosData[len(deptInfosData)-1]
	// 根据最后的 dept level 补齐最后一个组织架构信息
	switch lastDept.Level {
	case "0":
		result.Level1 = deptName
	case "1":
		result.Level2 = deptName
	case "2":
		result.Level3 = deptName
	case "3":
		result.Level4 = deptName
	case "4":
		result.Level5 = deptName
	}
	return result, nil
}
