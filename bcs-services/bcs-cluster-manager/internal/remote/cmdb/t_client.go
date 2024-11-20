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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// TCmdbClient for tencent inner cmdb client
var TCmdbClient *TClient

// SetTCmdbClient set cmdb client
func SetTCmdbClient(options TOptions) error {
	// init client
	cli, err := NewTCmdbClient(options)
	if err != nil {
		return err
	}

	TCmdbClient = cli
	return nil
}

// GetTCmdbClient get tcmdb client
func GetTCmdbClient() *TClient {
	return TCmdbClient
}

// NewTCmdbClient create tcmdb client
func NewTCmdbClient(options TOptions) (*TClient, error) {
	c := &TClient{
		appId:  options.AppId,
		appKey: options.AppKey,
		server: options.Server,
		debug:  options.Debug,
	}

	// disable tcmdb client
	if !options.Enable {
		return nil, nil
	}

	return c, nil
}

// TOptions for tcmdb client
type TOptions struct {
	// Enable enable client
	Enable bool
	// AppId app id
	AppId string
	// AppKey app key
	AppKey string
	// Server server
	Server string
	// Debug debug
	Debug bool
}

// TClient for tencent cmdb
type TClient struct {
	appId  string
	appKey string
	server string
	debug  bool
}

// generateGateWayAuth generate gateway auth
func (t *TClient) sigAuthHeaders() map[string]string {
	timeObj := time.Now()
	timestamp := strconv.FormatInt(timeObj.Unix(), 10)
	s := timestamp + t.appKey

	h := sha256.New()
	h.Write([]byte(s))
	signature := hex.EncodeToString(h.Sum(nil))

	headers := map[string]string{
		"x-timestamp": timestamp,
		"x-signature": signature,
		"x-app-id":    t.appId,
	}

	return headers
}

// QueryBusinessLevel2DetailInfo query business level2 detail info
func (t *TClient) QueryBusinessLevel2DetailInfo(bizL2ID int) (*BusinessL2Info, error) {
	if t == nil {
		return nil, ErrServerNotInit
	}

	// get_query_info
	var (
		reqURL  = fmt.Sprintf("%s/cmdb-service-business-domain/queryBusinessLevel2DetailInfo", t.server)
		request = &QueryBusinessLeven2InfoReq{
			ResultColumn: fieldBizL2Info,
			Size:         defaultSize,
			ScrollId:     "0",
			Condition:    buildQueryCondition(businessLevel2Id, in.String(), []interface{}{bizL2ID}),
		}
		respData = &QueryBusinessL2InfoResp{}
	)

	start := time.Now()

	reqClient := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		SetDebug(t.debug)

	authHeaders := t.sigAuthHeaders()
	for k, v := range authHeaders {
		reqClient.Set(k, v)
	}

	_, _, errs := reqClient.Send(request).EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "QueryBusinessLevel2DetailInfo",
			"http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api queryBusinessLevel2DetailInfo failed: %v", errs[0])
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "QueryBusinessLevel2DetailInfo",
		"http", metrics.LibCallStatusOK, start)

	if respData.Code != "0" {
		blog.Errorf("call api queryBusinessLevel2DetailInfo[%s] failed: %v", respData.TraceID, respData.Message)
		return nil, errors.New(respData.Message)
	}
	// successfully request
	blog.Infof("call api queryBusinessLevel2DetailInfo with url(%s) successfully", reqURL)

	if len(respData.Data.List) == 0 {
		return nil, fmt.Errorf("call api queryBusinessLevel2DetailInfo[%s] resp null", respData.TraceID)
	}

	return &respData.Data.List[0], nil
}

// queryServerInfoByIps ips 最多一次性支持查询50个IP (分批处理)
func (t *TClient) queryServerInfoByIps(ips []string) ([]Server, error) {
	if t == nil {
		return nil, ErrServerNotInit
	}

	var (
		reqURL  = fmt.Sprintf("%s/cmdb-service-federal-query/queryAllServerByBaseCondition", t.server)
		request = &QueryServerInfoReq{
			ResultColumn: fieldServerInfo,
			Condition:    buildQueryCondition(serverIp, in.String(), utils.ConvertStringsToInterfaces(ips)),
		}
		respData = &QueryServerInfoResp{}
	)

	start := time.Now()

	reqClient := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		SetDebug(t.debug)
	authHeaders := t.sigAuthHeaders()
	for k, v := range authHeaders {
		reqClient.Set(k, v)
	}

	_, _, errs := reqClient.Send(request).EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "queryServerInfoByIps",
			"http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api queryServerInfoByIps failed: %v", errs[0])
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "queryServerInfoByIps",
		"http", metrics.LibCallStatusOK, start)

	if respData.Code != "0" {
		blog.Errorf("call api queryAllServerByBaseCondition[%s] failed: %v", respData.TraceID, respData.Message)
		return nil, errors.New(respData.Message)
	}
	// successfully request
	blog.Infof("call api queryAllServerByBaseCondition with url(%s) successfully", reqURL)

	if len(respData.Data.List) == 0 {
		return nil, fmt.Errorf("call api queryBusinessLevel2DetailInfo[%s] resp null", respData.TraceID)
	}

	return respData.Data.List, nil
}

// GetAssetIdsByIps get asset ids by ips
func (t *TClient) GetAssetIdsByIps(ips []string) ([]Server, error) {
	if t == nil {
		return nil, ErrServerNotInit
	}

	chunk := utils.SplitStringsChunks(ips, maxSize)

	servers := make([]Server, 0)
	for _, v := range chunk {
		data, err := t.queryServerInfoByIps(v)
		if err != nil {
			return nil, err
		}
		servers = append(servers, data...)
	}
	return servers, nil
}
