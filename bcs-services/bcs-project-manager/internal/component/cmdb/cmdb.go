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

// Package cmdb xxx
package cmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bkuser"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

var (
	searchBusinessBatchSize = 200
	defaultTimeout          = 10
	defaultSupplierAccount  = "tencent"
	searchBizPath           = "/api/bk-cmdb/prod/api/v3/biz/search/default"
	getBizTopoPath          = "/api/bk-cmdb/prod/api/v3/find/topoinst/biz/%v"
	// CacheKeyBusinessPrefix cache key business prefix
	CacheKeyBusinessPrefix = "BUSINESS_%s"
)

// SearchBusinessResp cmdb search business resp
type SearchBusinessResp struct {
	Code      int                `json:"code"`
	Result    bool               `json:"result"`
	Message   string             `json:"message"`
	RequestID string             `json:"request_id"`
	Data      SearchBusinessData `json:"data"`
}

// GetBusinessTopologyResp cmdb get business topology resp
type GetBusinessTopologyResp struct {
	Code      int                    `json:"code"`
	Result    bool                   `json:"result"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"request_id"`
	Data      []BusinessTopologyData `json:"data"`
}

// SearchBusinessData cmdb search business resp data
type SearchBusinessData struct {
	Count int            `json:"count"`
	Info  []BusinessData `json:"info"`
}

// BusinessTopologyData cmdb get business topology resp data
type BusinessTopologyData struct {
	Default    int                    `json:"default"`
	BkObjID    string                 `json:"bk_obj_id"`
	BkObjName  string                 `json:"bk_obj_name"`
	BkInstID   int                    `json:"bk_inst_id"`
	BkInstName string                 `json:"bk_inst_name"`
	Child      []BusinessTopologyData `json:"child"`
}

// TransferToProto transfer cmdb data to proto
func (b *BusinessTopologyData) TransferToProto() *proto.TopologyData {
	protoData := &proto.TopologyData{
		Default:    uint32(b.Default),
		BkObjId:    b.BkObjID,
		BkObjName:  b.BkObjName,
		BkInstId:   uint32(b.BkInstID),
		BkInstName: b.BkInstName,
		Child:      []*proto.TopologyData{},
	}
	for _, child := range b.Child {
		protoData.Child = append(protoData.Child, child.TransferToProto())
	}
	return protoData
}

// BusinessData cmdb business data
type BusinessData struct {
	BS2NameID       int    `json:"bs2_name_id"`
	Default         int    `json:"default"`
	BKBizID         int64  `json:"bk_biz_id"`
	BKBizName       string `json:"bk_biz_name"`
	BKBizMaintainer string `json:"bk_biz_maintainer"`
}

// IsMaintainer 校验用户是否为指定业务的运维
func IsMaintainer(ctx context.Context, username string, bizID string) (bool, error) {
	searchData, err := SearchBusiness(ctx, username, bizID)
	if err != nil {
		return false, err
	}
	// 判断是否存在当前用户为业务运维角色的业务
	if searchData.Count > 0 {
		return true, nil
	}
	return false, errorx.NewNoMaintainerRoleErr()
}

// GetBusinessByID 通过业务ID获取业务信息
func GetBusinessByID(ctx context.Context, bizID string, useCache bool) (BusinessData, error) {
	// 先尝试从缓存中获取
	c := cache.GetCache()
	if useCache {
		if biz, exists := c.Get(fmt.Sprintf(CacheKeyBusinessPrefix, bizID)); exists {
			logging.Info("get business %s hit cache", bizID)
			data := biz.(BusinessData)
			return data, nil
		}
	}
	data, err := SearchBusiness(ctx, "", bizID)
	if err != nil {
		return BusinessData{}, err
	}
	if data.Count == 0 {
		return BusinessData{}, fmt.Errorf("business %s not exists", bizID)
	}
	_ = c.Add(fmt.Sprintf(CacheKeyBusinessPrefix, bizID), data.Info[0], time.Hour)
	return data.Info[0], nil
}

// SearchBusiness 通过用户和业务ID，查询业务
func SearchBusiness(ctx context.Context, username string, bizID string) (*SearchBusinessData, error) {
	// 获取超时时间
	timeout := defaultTimeout
	if config.GlobalConf.CMDB.Timeout != 0 {
		timeout = config.GlobalConf.CMDB.Timeout
	}
	// 获取开发商账户
	supplierAccount := defaultSupplierAccount
	if config.GlobalConf.CMDB.BKSupplierAccount != "" {
		supplierAccount = config.GlobalConf.CMDB.BKSupplierAccount
	}
	// 组装请求参数
	condition := map[string]interface{}{}
	if username != "" {
		condition["bk_biz_maintainer"] = username
	}
	if bizID != "" {
		bizIDInt, _ := strconv.Atoi(bizID)
		condition["bk_biz_id"] = bizIDInt
	}
	req := gorequest.SuperAgent{
		Url:    fmt.Sprintf("%s%s", config.GlobalConf.CMDB.Host, searchBizPath),
		Method: "POST",
		Data: map[string]interface{}{
			"condition":           condition,
			"bk_supplier_account": supplierAccount,
		},
		Debug: config.GlobalConf.CMDB.Debug,
	}

	headers, err := bkuser.GetAuthHeader(ctx)
	if err != nil {
		logging.Error("SearchBusiness get auth header failed, %s", err.Error())
		return nil, errorx.NewRequestCMDBErr(err.Error())
	}

	// 获取返回数据
	body, err := component.Request(req, timeout, config.GlobalConf.CMDB.Proxy, headers)
	if err != nil {
		return nil, errorx.NewRequestCMDBErr(err.Error())
	}
	// 解析返回的body
	var resp SearchBusinessResp
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		logging.Error("parse search biz body error, body: %v", body)
		return nil, err
	}
	if resp.Code != errorx.Success {
		return nil, errorx.NewRequestCMDBErr(resp.Message)
	}
	return &resp.Data, nil
}

// GetBusinessMaintainers get maintainers by bizID
func GetBusinessMaintainers(ctx context.Context, bizID string) ([]string, error) {
	business, err := GetBusinessByID(ctx, bizID, false)
	if err != nil {
		return nil, err
	}
	// 判断是否存在当前用户为业务运维角色的业务
	maintainers := stringx.SplitString(business.BKBizMaintainer)
	return maintainers, nil
}

// BatchSearchBusinessByBizIDs batch search business by bizIDs
// if any bizID not exists or error, warning will be logged and continue
func BatchSearchBusinessByBizIDs(ctx context.Context, bizIDs []string) map[string]BusinessData {
	if len(bizIDs) == 0 {
		return nil
	}
	bizIDs = stringx.RemoveDuplicateValues(bizIDs)
	result := make(map[string]BusinessData, len(bizIDs))
	// 先尝试从缓存中获取
	c := cache.GetCache()
	notHitBizIDs := make([]string, 0)
	for _, bizID := range bizIDs {
		if biz, exists := c.Get(fmt.Sprintf(CacheKeyBusinessPrefix, bizID)); exists {
			logging.Info("get business %s hit cache", bizID)
			result[bizID] = biz.(BusinessData)
		} else {
			notHitBizIDs = append(notHitBizIDs, bizID)
		}
	}
	if len(notHitBizIDs) == 0 {
		return result
	}
	// 如果部分业务没有命中缓存，再去cmdb中查询并更新缓存
	batchNum := len(notHitBizIDs) / searchBusinessBatchSize
	if len(notHitBizIDs)%searchBusinessBatchSize > 0 {
		batchNum++
	}
	for i := 0; i < batchNum; i++ {
		start := i * searchBusinessBatchSize
		end := (i + 1) * searchBusinessBatchSize
		if end > len(notHitBizIDs) {
			end = len(notHitBizIDs)
		}
		batchBizIDs := []int{}
		for _, bizID := range notHitBizIDs[start:end] {
			bizIDInt, err := strconv.Atoi(bizID)
			if err != nil {
				logging.Error("bizID %s is invalid", bizID)
				continue
			}
			batchBizIDs = append(batchBizIDs, bizIDInt)
		}
		condition := "AND"
		rules := []map[string]interface{}{
			{
				"field":    "bk_biz_id",
				"operator": "in",
				"value":    batchBizIDs,
			},
		}

		businesses, err := searchBusinessByIds(ctx, condition, rules)
		if err != nil {
			logging.Error("batch search business failed, err: %s", err.Error())
			continue
		}

		for _, biz := range businesses {
			result[strconv.Itoa(int(biz.BKBizID))] = biz
			_ = c.Add(fmt.Sprintf(CacheKeyBusinessPrefix, strconv.Itoa(int(biz.BKBizID))), biz, time.Hour)
		}
	}
	return result
}

// GetBusinessTopology get business topology by bizID
func GetBusinessTopology(ctx context.Context, bizID string) ([]BusinessTopologyData, error) {
	// 获取超时时间
	timeout := defaultTimeout
	if config.GlobalConf.CMDB.Timeout != 0 {
		timeout = config.GlobalConf.CMDB.Timeout
	}
	// 获取开发商账户
	supplierAccount := defaultSupplierAccount
	if config.GlobalConf.CMDB.BKSupplierAccount != "" {
		supplierAccount = config.GlobalConf.CMDB.BKSupplierAccount
	}
	// 组装请求参数
	req := gorequest.SuperAgent{
		Url:    fmt.Sprintf("%s%s", config.GlobalConf.CMDB.Host, fmt.Sprintf(getBizTopoPath, bizID)),
		Method: "POST",
		Data: map[string]interface{}{
			"bk_biz_id":           bizID,
			"bk_supplier_account": supplierAccount,
		},
		Debug: config.GlobalConf.CMDB.Debug,
	}

	headers, err := bkuser.GetAuthHeader(ctx)
	if err != nil {
		logging.Error("GetBusinessTopology get auth header failed, %s", err.Error())
		return nil, errorx.NewRequestCMDBErr(err.Error())
	}

	// 获取返回数据
	body, err := component.Request(req, timeout, config.GlobalConf.CMDB.Proxy, headers)
	if err != nil {
		return nil, errorx.NewRequestCMDBErr(err.Error())
	}
	// 解析返回的body
	var resp GetBusinessTopologyResp
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		logging.Error("parse search biz body error, body: %v", body)
		return nil, err
	}
	if resp.Code != errorx.Success {
		return nil, errorx.NewRequestCMDBErr(resp.Message)
	}
	return resp.Data, nil
}

func searchBusinessByIds(ctx context.Context, condition string,
	rules []map[string]interface{}) ([]BusinessData, error) {
	timeout := defaultTimeout
	if config.GlobalConf.CMDB.Timeout != 0 {
		timeout = config.GlobalConf.CMDB.Timeout
	}
	supplierAccount := defaultSupplierAccount
	if config.GlobalConf.CMDB.BKSupplierAccount != "" {
		supplierAccount = config.GlobalConf.CMDB.BKSupplierAccount
	}
	// 组装请求参数
	req := gorequest.SuperAgent{
		Url:    fmt.Sprintf("%s%s", config.GlobalConf.CMDB.Host, searchBizPath),
		Method: "POST",
		Data: map[string]interface{}{
			"biz_property_filter": map[string]interface{}{
				"condition": condition,
				"rules":     rules,
			},
			"bk_supplier_account": supplierAccount,
		},
		Debug: config.GlobalConf.CMDB.Debug,
	}

	headers, err := bkuser.GetAuthHeader(ctx)
	if err != nil {
		logging.Error("GetBusinessTopology get auth header failed, %s", err.Error())
		return nil, errorx.NewRequestCMDBErr(err.Error())
	}

	// 获取返回数据
	body, err := component.Request(req, timeout, config.GlobalConf.CMDB.Proxy, headers)
	if err != nil {
		logging.Error("request search business failed, err: %s", err.Error())
		return nil, fmt.Errorf("request search business failed, err: %s", err.Error())
	}
	// 解析返回的body
	var resp SearchBusinessResp
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return nil, fmt.Errorf("parse search biz body failed, body: %v, err: %v", body, err)
	}
	if resp.Code != errorx.Success {
		return nil, fmt.Errorf("search business failed, code: %d, message: %s", resp.Code, resp.Message)
	}
	return resp.Data.Info, nil
}
