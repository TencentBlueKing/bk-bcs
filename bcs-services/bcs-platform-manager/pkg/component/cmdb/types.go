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

import bkcmdbkube "configcenter/src/kube/types"

// CommonReq defines the common request structure for BCS cluster operations.
type CommonReq struct {
	BKBizID int64           `json:"bk_biz_id"`
	Page    Page            `json:"page"`
	Fields  []string        `json:"fields"`
	Filter  *PropertyFilter `json:"filter"`
}

// CommonResp defines the common response structure for BCS cluster operations.
type CommonResp struct {
	Code      int64  `json:"code"`
	Result    bool   `json:"result"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// Page page
type Page struct {
	Start int    `json:"start"`
	Limit int    `json:"limit"`
	Sort  string `json:"sort"`
}

// PropertyFilter property filter
type PropertyFilter struct {
	Condition string `json:"condition"`
	Rules     []Rule `json:"rules"`
}

// Rule rule
type Rule struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// GetBcsPodReq defines the structure of the request for getting BCS pods.
type GetBcsPodReq struct {
	CommonReq
}

// GetBcsPodResp defines the structure of the response for getting BCS pods.
type GetBcsPodResp struct {
	CommonResp
	Data *GetBcsPodRespData `json:"data"`
}

// GetBcsPodRespData defines the structure of the response data for getting BCS pods.
type GetBcsPodRespData struct {
	Count int               `json:"count"`
	Info  *[]bkcmdbkube.Pod `json:"info"`
}

// DeleteBcsPodReq defines the request structure for deleting a BCS pod.
type DeleteBcsPodReq struct {
	Data *[]DeleteBcsPodReqData `json:"data"`
}

// DeleteBcsPodReqData defines the data structure in the DeleteBcsPodRequest.
type DeleteBcsPodReqData struct {
	BKBizID *int64   `json:"bk_biz_id"`
	IDs     *[]int64 `json:"ids"`
}

// DeleteBcsPodResp defines the response structure for deleting a BCS pod.
type DeleteBcsPodResp struct {
	CommonResp
	Data interface{} `json:"data"`
}

// GetBcsWorkloadReq represents a request for getting BCS workload
type GetBcsWorkloadReq struct {
	CommonReq
	Kind string `json:"kind"`
}

// GetBcsWorkloadResp represents a response for getting BCS workload
type GetBcsWorkloadResp struct {
	CommonResp
	Data *GetBcsWorkloadRespData `json:"data"`
}

// GetBcsWorkloadRespData represents the data structure of the response for getting BCS workload
type GetBcsWorkloadRespData struct {
	Count int64         `json:"count"`
	Info  []interface{} `json:"info"`
}

// DeleteBcsWorkloadReq defines the structure of the request for deleting a BCS workload.
type DeleteBcsWorkloadReq struct {
	BKBizID *int64   `json:"bk_biz_id"`
	Kind    *string  `json:"kind"`
	IDs     *[]int64 `json:"ids"`
}

// DeleteBcsWorkloadResp defines the structure of the response for deleting a BCS workload.
type DeleteBcsWorkloadResp struct {
	CommonResp
	Data interface{} `json:"data"`
}

// GetBcsNamespaceReq represents the request for getting a BCS namespace
type GetBcsNamespaceReq struct {
	CommonReq
}

// GetBcsNamespaceResp represents the response for getting a BCS namespace
type GetBcsNamespaceResp struct {
	CommonResp
	Data *GetBcsNamespaceRespData `json:"data"`
}

// GetBcsNamespaceRespData represents the data for getting a BCS namespace
type GetBcsNamespaceRespData struct {
	Count int64                   `json:"count"`
	Info  *[]bkcmdbkube.Namespace `json:"info"`
}

// DeleteBcsNamespaceReq represents the request for deleting a BCS namespace
type DeleteBcsNamespaceReq struct {
	BKBizID *int64   `json:"bk_biz_id"`
	IDs     *[]int64 `json:"ids"`
}

// DeleteBcsNamespaceResp represents the response for deleting a BCS namespace
type DeleteBcsNamespaceResp struct {
	CommonResp
	Data interface{} `json:"data"`
}

// GetBcsNodeReq defines the structure of the request for getting BCS nodes.
type GetBcsNodeReq struct {
	CommonReq
}

// DeleteBcsNodeReq defines the structure of the request for deleting BCS nodes.
type DeleteBcsNodeReq struct {
	BKBizID *int64   `json:"bk_biz_id"`
	IDs     *[]int64 `json:"ids"`
}

// GetBcsNodeResp defines the structure of the response for getting BCS nodes.
type GetBcsNodeResp struct {
	CommonResp
	Data *GetBcsNodeRespData `json:"data"`
}

// GetBcsNodeRespData defines the structure of the response data for getting BCS nodes.
type GetBcsNodeRespData struct {
	Count int64              `json:"count"`
	Info  *[]bkcmdbkube.Node `json:"info"`
}

// DeleteBcsNodeResponse defines the structure of the response for deleting BCS nodes.
type DeleteBcsNodeResp struct {
	CommonResp
	Data interface{} `json:"data"`
}

// GetBcsClusterReq defines the request structure for getting BCS cluster information.
type GetBcsClusterReq struct {
	CommonReq
}

// DeleteBcsClusterReq represents the request for deleting a BCS cluster
type DeleteBcsClusterReq struct {
	BKBizID *int64   `json:"bk_biz_id"`
	IDs     *[]int64 `json:"ids"`
}

// GetBcsClusterResp defines the response structure for getting BCS cluster information.
type GetBcsClusterResp struct {
	CommonResp
	Data GetBcsClusterRespData `json:"data"`
}

// GetBcsClusterRespData defines the data structure for getting BCS cluster information.
type GetBcsClusterRespData struct {
	Count int64                `json:"count"`
	Info  []bkcmdbkube.Cluster `json:"info"`
}

// DeleteBcsClusterResp represents the response for deleting a BCS cluster
type DeleteBcsClusterResp struct {
	CommonResp
	Data interface{} `json:"data"`
}
