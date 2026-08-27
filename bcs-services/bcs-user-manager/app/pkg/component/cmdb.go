/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
    10| * limitations under the License.
*/

package component

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	pkgutils "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/utils"
	apputils "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
)

const (
	searchBizPath = "api/v3/biz/search/default"
)

// SearchBusinessData cmdb search business resp data
type SearchBusinessData struct {
	Count int            `json:"count"`
	Info  []BusinessData `json:"info"`
}

// BusinessData cmdb business data
type BusinessData struct {
	BKBizID         int64  `json:"bk_biz_id"`
	BKBizName       string `json:"bk_biz_name"`
	BKBizMaintainer string `json:"bk_biz_maintainer"`
}

// searchBusinessRequest 对齐 cluster-manager SearchBusinessRequest
type searchBusinessRequest struct {
	Fields    []string               `json:"fields"`
	Condition map[string]interface{} `json:"condition"`
	Page      searchBusinessPage     `json:"page"`
	UserName  string                 `json:"bk_username"`
	Operator  string                 `json:"operator"`
}

type searchBusinessPage struct {
	Start int    `json:"start"`
	Limit int    `json:"limit"`
	Sort  string `json:"sort"`
}

// GetBusinessByID 通过业务 ID 获取业务信息，对齐 cluster-manager 的 CMDB 查询
func GetBusinessByID(ctx context.Context, bizID string) (BusinessData, error) {
	data, err := SearchBusiness(ctx, "", bizID)
	if err != nil {
		return BusinessData{}, err
	}
	if data.Count == 0 || len(data.Info) == 0 {
		return BusinessData{}, fmt.Errorf("business %s not exists", bizID)
	}
	return data.Info[0], nil
}

// SearchBusiness 通过用户和业务 ID 查询业务
func SearchBusiness(ctx context.Context, username string, bizID string) (*SearchBusinessData, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil || !cfg.Cmdb.Enable || cfg.Cmdb.Host == "" {
		return nil, fmt.Errorf("cmdb config not found")
	}

	condition := map[string]interface{}{}
	if username != "" {
		condition["bk_biz_maintainer"] = username
	}
	if bizID != "" {
		bizIDInt, err := strconv.Atoi(bizID)
		if err != nil {
			return nil, fmt.Errorf("invalid business id %s", bizID)
		}
		condition["bk_biz_id"] = bizIDInt
	}

	authInfo, err := getCMDBAuthorization()
	if err != nil {
		blog.Errorf("SearchBusiness get auth header failed, %s", err.Error())
		return nil, err
	}

	req := GetClient().R().
		SetContext(ctx).
		SetHeaders(pkgutils.GetLaneIDByCtx(ctx)).
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetBody(&searchBusinessRequest{
			Fields:    []string{"bk_biz_id", "bk_biz_name", "bk_biz_maintainer"},
			Condition: condition,
		})
	if tenantID := apputils.GetTenantIDFromContext(ctx); tenantID != "" {
		req.SetHeader(apputils.HeaderTenantID, tenantID)
	}

	resp, err := req.Post(fmt.Sprintf("%s/%s", cfg.Cmdb.Host, searchBizPath))
	if err != nil {
		return nil, err
	}

	var data SearchBusinessData
	if err := UnmarshalBKResult(resp, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func getCMDBAuthorization() (string, error) {
	cfg := config.GetGlobalConfig()
	if cfg.Cmdb.AppCode != "" && cfg.Cmdb.AppSecret != "" {
		auth := &AuthInfo{
			BkAppCode:   cfg.Cmdb.AppCode,
			BkAppSecret: cfg.Cmdb.AppSecret,
			BkUserName:  cfg.Cmdb.BkUserName,
		}
		userAuth, err := json.Marshal(auth)
		if err != nil {
			return "", err
		}
		return string(userAuth), nil
	}
	return GetBKAPIAuthorization(cfg.Cmdb.BkUserName)
}
