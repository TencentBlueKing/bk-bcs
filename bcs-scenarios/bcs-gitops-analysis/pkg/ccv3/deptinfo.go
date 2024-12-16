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

package ccv3

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
)

type getStaffInfoResponse struct {
	Code      string            `json:"code"`
	Result    bool              `json:"result"`
	RequestId string            `json:"request_id"`
	Message   string            `json:"message"`
	Data      *getStaffInfoData `json:"data"`
}

type getStaffInfoData struct {
	EnglishName string `json:"EnglishName"`
	ChineseName string `json:"ChineseName"`
	GroupName   string `json:"GroupName"`
	GroupId     string `json:"GroupId"`
}

// GetUserDeptInfo get the user info
func (h *handler) GetUserDeptInfo(user string) (*UserDeptInfo, error) {
	if h.op.IsExternal {
		return &UserDeptInfo{}, nil
	}
	getStaffUrl := fmt.Sprintf("%s%s?login_name=%s", h.op.BKCCUrl, getStaffInfoApi, user)
	getStaffResp := new(getStaffInfoResponse)
	if err := h.query(getStaffUrl, http.MethodGet, nil, getStaffResp); err != nil {
		return nil, errors.Wrapf(err, "get staff failed")
	}
	staffData := getStaffResp.Data
	if staffData == nil || staffData.GroupId == "" {
		return nil, fmt.Errorf("user '%s' not found", user)
	}
	groupID := staffData.GroupId
	group, err := strconv.Atoi(groupID)
	if err != nil {
		return nil, errors.Wrapf(err, "parse group id '%s' failed", groupID)
	}
	di, err := h.getDeptInfo(int64(group), staffData.GroupName)
	if err != nil {
		return nil, errors.Wrapf(err, "get dept info failed")
	}

	return &UserDeptInfo{
		UserName:    staffData.EnglishName,
		ChineseName: staffData.ChineseName,
		Level0:      di.Level0,
		Level1:      di.Level1,
		Level2:      di.Level2,
		Level3:      di.Level3,
		Level4:      di.Level4,
		Level5:      di.Level5,
	}, nil
}

// GetBizDeptInfo return business info with dept details
func (h *handler) GetBizDeptInfo(bkBizIDs []int64) (map[int64]*BusinessDeptInfo, error) {
	if h.op.IsExternal {
		return make(map[int64]*BusinessDeptInfo), nil
	}
	infos, err := h.SearchBusiness(bkBizIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "search business failed")
	}

	result := make(map[int64]*BusinessDeptInfo)
	for _, item := range infos {
		bkBizID := item.BkBizId
		groupID := item.GroupID
		di, err := h.getDeptInfo(groupID, item.GroupName)
		if err != nil {
			return nil, errors.Wrapf(err, "get dept info failed for group '%d'", groupID)
		}
		result[bkBizID] = &BusinessDeptInfo{
			BKBizID:   bkBizID,
			BKBizName: item.BkBizName,
			Level0:    di.Level0,
			Level1:    di.Level1,
			Level2:    di.Level2,
			Level3:    di.Level3,
			Level4:    di.Level4,
			Level5:    di.Level5,
		}
	}
	return result, nil
}
