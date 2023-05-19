/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package cmdb xxx
package cmdb

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"

	"github.com/parnurzeal/gorequest"
)

var (
	defaultTimeout         = 10
	defaultSupplierAccount = "tencent"
	searchBizPath          = "/api/c/compapi/v2/cc/search_business/"
	getBizTopoPath         = "/api/c/compapi/v2/cc/search_biz_inst_topo/"
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
func IsMaintainer(username string, bizID string) (bool, error) {
	searchData, err := SearchBusiness(username, bizID)
	if err != nil {
		return false, err
	}
	// 判断是否存在当前用户为业务运维角色的业务
	if searchData.Count > 0 {
		return true, nil
	}
	return false, errorx.NewNoMaintainerRoleErr()
}

// SearchBusiness 通过用户和业务ID，查询业务
func SearchBusiness(username string, bizID string) (*SearchBusinessData, error) {
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
	headers := map[string]string{"Content-Type": "application/json"}
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
			"bk_app_code":         config.GlobalConf.App.Code,
			"bk_app_secret":       config.GlobalConf.App.Secret,
			"bk_username":         config.GlobalConf.CMDB.BKUsername,
		},
		Debug: config.GlobalConf.CMDB.Debug,
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
func GetBusinessMaintainers(bizID string) ([]string, error) {
	searchData, err := SearchBusiness("", bizID)
	if err != nil {
		return nil, err
	}
	// 判断是否存在当前用户为业务运维角色的业务
	if searchData.Count == 0 {
		return nil, fmt.Errorf("get business by id %s failed", bizID)
	}
	business := searchData.Info[0]
	maintainers := stringx.SplitString(business.BKBizMaintainer)
	return maintainers, nil
}

// GetBusinessTopology get business topology by bizID
func GetBusinessTopology(bizID string) ([]BusinessTopologyData, error) {
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
	headers := map[string]string{"Content-Type": "application/json"}
	// 组装请求参数
	req := gorequest.SuperAgent{
		Url:    fmt.Sprintf("%s%s", config.GlobalConf.CMDB.Host, getBizTopoPath),
		Method: "POST",
		Data: map[string]interface{}{
			"bk_biz_id":           bizID,
			"bk_supplier_account": supplierAccount,
			"bk_app_code":         config.GlobalConf.App.Code,
			"bk_app_secret":       config.GlobalConf.App.Secret,
			"bk_username":         config.GlobalConf.CMDB.BKUsername,
		},
		Debug: config.GlobalConf.CMDB.Debug,
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
