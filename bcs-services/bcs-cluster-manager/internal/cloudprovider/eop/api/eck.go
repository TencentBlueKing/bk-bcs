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

package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/google/uuid"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

const (
	defaultTimeOut = time.Second * 60
	endingPoint    = "https://eck-global.ctapi.ctyun.cn"
)

// Credential auth credential
type credential struct {
	ak string
	sk string
}

// CTClient Ctyun client
type CTClient struct {
	credential
}

// NewCTClient create new Ctyun client
func NewCTClient(opt *cloudprovider.CommonOption) (*CTClient, error) {
	if opt.Account == nil {
		return nil, fmt.Errorf("create NewCTClient failed, empty Credential")
	}
	if opt.Account.SecretID == "" || opt.Account.SecretKey == "" {
		return nil, fmt.Errorf("create NewCTClient failed, empty AK or SK")
	}

	return &CTClient{
		credential{
			ak: opt.Account.SecretID,
			sk: opt.Account.SecretKey,
		},
	}, nil
}

// ListRegions lists regions
func (c *CTClient) ListRegions() ([]*Region, error) {
	reqPath := "/v2/cluster/listRegions"
	resp := &ListRegionResponse{}

	requestID, _ := uuid.NewUUID()
	singerDate := time.Now()
	eopDate := singerDate.Format("20060102T150405Z")
	eopDt := singerDate.Format("20060102")

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Get(endingPoint+reqPath).
		SetDebug(true).
		Set("Content-Type", "application/json").
		Set("User-Agent", "bcs-cluster-manager/v1.0").
		Set("Eop-date", eopDate).
		Set("ctyun-eop-request-id", requestID.String()).
		Set("Eop-Authorization", c.buildSignHeader(requestID.String(), eopDate, eopDt, "", "")).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call ListRegions API failed: %v", errs[0])
		return nil, errs[0]
	}

	if result.StatusCode != 200 {
		errMsg := fmt.Errorf("call ListRegions API error: code[%d], %s",
			result.StatusCode, string(body))
		return nil, errMsg
	}

	if resp.StatusCode != "ok" {
		blog.Errorf("ListRegions failed, %s", string(body))
		return nil, fmt.Errorf("ListRegions failed, %s", string(body))
	}

	if resp.ReturnObj == nil {
		return nil, fmt.Errorf("ListRegions got empty regions")
	}

	return resp.ReturnObj.Regions, nil
}

// GetCluster gets ECK cluster
func (c *CTClient) GetCluster(clusterID string) (*Cluster, error) {
	reqPath := "/v2/cluster/get"
	params := fmt.Sprintf("clusterId=%s", clusterID)
	resp := &GetClusterResponse{}

	requestID, _ := uuid.NewUUID()
	singerDate := time.Now()
	eopDate := singerDate.Format("20060102T150405Z")
	eopDt := singerDate.Format("20060102")

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Get(endingPoint+reqPath).
		Query(params).
		SetDebug(true).
		Set("Content-Type", "application/json").
		Set("User-Agent", "bcs-cluster-manager/v1.0").
		Set("Eop-date", eopDate).
		Set("ctyun-eop-request-id", requestID.String()).
		Set("Eop-Authorization", c.buildSignHeader(requestID.String(), eopDate, eopDt, "", params)).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call GetCluster API failed: %v", errs[0])
		return nil, errs[0]
	}

	if result.StatusCode != 200 {
		errMsg := fmt.Errorf("call GetCluster API error: code[%d], %s",
			result.StatusCode, string(body))
		return nil, errMsg
	}

	if resp.StatusCode != "ok" {
		blog.Errorf("GetCluster failed, %s", string(body))
		return nil, fmt.Errorf("GetCluster failed, %s", string(body))
	}

	if resp.ReturnObj == nil {
		blog.Errorf("GetCluster lost cluster info in response")
		return nil, fmt.Errorf("GetCluster lost cluster info in response")
	}

	return resp.ReturnObj.Cluster, nil
}

// CreateCluster creates ECK cluster
func (c *CTClient) CreateCluster(req *CreateClusterRequest) (*CreateClusterReObj, error) {
	reqPath := "/v2/cluster/create"
	resp := &CreateClusterResponse{}

	requestID, _ := uuid.NewUUID()
	signDate := time.Now()
	eopDate := signDate.Format("20060102T150405Z")
	eopDt := signDate.Format("20060102")

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	result, respBody, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(endingPoint+reqPath).
		SetDebug(true).
		Set("Content-Type", "application/json").
		Set("User-Agent", "bcs-cluster-manager/v1.0").
		Set("Eop-date", eopDate).
		Set("ctyun-eop-request-id", requestID.String()).
		Set("Eop-Authorization",
			c.buildSignHeader(requestID.String(), eopDate, eopDt, string(reqBody), "")).
		Send(string(reqBody)).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call CreateCluster API failed: %v", errs[0])
		return nil, errs[0]
	}

	if result.StatusCode != 200 {
		errMsg := fmt.Errorf("call CreateCluster API error: code[%d], %s",
			result.StatusCode, string(respBody))
		return nil, errMsg
	}

	if resp.StatusCode != "ok" {
		blog.Errorf("CreateCluster failed, %s", string(respBody))
		return nil, fmt.Errorf("CreateCluster failed, %s", string(respBody))
	}

	if resp.ReturnObj == nil {
		blog.Errorf("CreateCluster lost cluster info in response")
		return nil, fmt.Errorf("CreateCluster lost cluster info in response")
	}

	return resp.ReturnObj, nil
}

// DeleteCluster deletes ECK cluster
func (c *CTClient) DeleteCluster(req *DeleteClusterReq) (*DeleteClusterReObj, error) {
	reqPath := "/v2/cluster/delete"
	resp := &DeleteClusterResponse{}

	requestID, _ := uuid.NewUUID()
	signDate := time.Now()
	eopDate := signDate.Format("20060102T150405Z")
	eopDt := signDate.Format("20060102")

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	result, respBody, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(endingPoint+reqPath).
		SetDebug(true).
		Set("Content-Type", "application/json").
		Set("User-Agent", "bcs-cluster-manager/v1.0").
		Set("Eop-date", eopDate).
		Set("ctyun-eop-request-id", requestID.String()).
		Set("Eop-Authorization",
			c.buildSignHeader(requestID.String(), eopDate, eopDt, string(reqBody), "")).
		Send(string(reqBody)).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call DeleteCluster API failed: %v", errs[0])
		return nil, errs[0]
	}

	if result.StatusCode != 200 {
		errMsg := fmt.Errorf("call DeleteCluster API error: code[%d], %s",
			result.StatusCode, string(respBody))
		return nil, errMsg
	}

	if resp.StatusCode != "ok" {
		blog.Errorf("DeleteCluster failed, %s", string(respBody))
		return nil, fmt.Errorf("DeleteCluster failed, %s", string(respBody))
	}

	return resp.ReturnObj, nil
}

// GetKubeConfig gets kubeconfig
func (c *CTClient) GetKubeConfig(clusterId string) (*GetKubeConfigReObj, error) {
	reqPath := "/v2/cluster/getKubeConfig"
	params := fmt.Sprintf("clusterId=%s", clusterId)
	resp := &GetKubeConfigResponse{}

	requestID, _ := uuid.NewUUID()
	singerDate := time.Now()
	eopDate := singerDate.Format("20060102T150405Z")
	eopDt := singerDate.Format("20060102")

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Get(endingPoint+reqPath).
		Query(params).
		SetDebug(true).
		Set("Content-Type", "application/json").
		Set("User-Agent", "bcs-cluster-manager/v1.0").
		Set("Eop-date", eopDate).
		Set("ctyun-eop-request-id", requestID.String()).
		Set("Eop-Authorization", c.buildSignHeader(requestID.String(), eopDate, eopDt, "", params)).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call GetKubeConfig API failed: %v", errs[0])
		return nil, errs[0]
	}

	if result.StatusCode != 200 {
		errMsg := fmt.Errorf("call GetKubeConfig API error: code[%d], %s",
			result.StatusCode, string(body))
		return nil, errMsg
	}

	if resp.StatusCode != "ok" {
		blog.Errorf("GetKubeConfig failed, %s", string(body))
		return nil, fmt.Errorf("GetKubeConfig failed, %s", string(body))
	}

	if resp.ReturnObj == nil {
		blog.Errorf("GetKubeConfig lost kubeconfig in response")
		return nil, fmt.Errorf("GetKubeConfig lost kubeconfig in response")
	}

	return resp.ReturnObj, nil
}

// ListNodes lists ECK cluster nodes
func (c *CTClient) ListNodes(req *ListNodeReq) ([]*Node, error) {
	reqPath := "/v2/node/list"

	if req.ClusterID == "" {
		return nil, fmt.Errorf("ListNodes empty clusterId")
	}

	params := fmt.Sprintf("clusterId=%s", req.ClusterID)
	if req.NodeNames != "" {
		params = fmt.Sprintf("%s&nodeNames=%s", params, req.NodeNames)
	}
	if req.NodePoolId != "" {
		params = fmt.Sprintf("%s&nodePoolId=%s", params, req.NodePoolId)
	}
	if req.Page != 0 {
		params = fmt.Sprintf("%s&page=%d", params, req.Page)
	}
	if req.PerPage != 0 {
		params = fmt.Sprintf("%s&perPage=%d", params, req.PerPage)
	}

	resp := &ListNodeResponse{}

	requestID, _ := uuid.NewUUID()
	singerDate := time.Now()
	eopDate := singerDate.Format("20060102T150405Z")
	eopDt := singerDate.Format("20060102")

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Get(endingPoint+reqPath).
		Query(params).
		SetDebug(true).
		Set("Content-Type", "application/json").
		Set("User-Agent", "bcs-cluster-manager/v1.0").
		Set("Eop-date", eopDate).
		Set("ctyun-eop-request-id", requestID.String()).
		Set("Eop-Authorization", c.buildSignHeader(requestID.String(), eopDate, eopDt, "", params)).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call ListNodes API failed: %v", errs[0])
		return nil, errs[0]
	}

	if result.StatusCode != 200 {
		errMsg := fmt.Errorf("call ListNodes API error: code[%d], %s",
			result.StatusCode, string(body))
		return nil, errMsg
	}

	if resp.StatusCode != "ok" {
		blog.Errorf("ListNodes failed, %s", string(body))
		return nil, fmt.Errorf("ListNodes failed, %s", string(body))
	}

	if resp.ReturnObj == nil {
		blog.Errorf("ListNodes lost node info in response")
		return nil, fmt.Errorf("ListNodes lost node info in response")
	}

	return resp.ReturnObj.Nodes, nil
}

// GetNodePool gets nodepool
func (c *CTClient) GetNodePool(nodePoolId string) (*NodePool, error) {
	reqPath := "/v2/nodePool/get"
	params := fmt.Sprintf("nodePoolId=%s", nodePoolId)
	resp := &GetNodePoolResponse{}

	requestID, _ := uuid.NewUUID()
	singerDate := time.Now()
	eopDate := singerDate.Format("20060102T150405Z")
	eopDt := singerDate.Format("20060102")

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Get(endingPoint+reqPath).
		Query(params).
		SetDebug(true).
		Set("Content-Type", "application/json").
		Set("User-Agent", "bcs-cluster-manager/v1.0").
		Set("Eop-date", eopDate).
		Set("ctyun-eop-request-id", requestID.String()).
		Set("Eop-Authorization", c.buildSignHeader(requestID.String(), eopDate, eopDt, "", params)).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call GetNodePool API failed: %v", errs[0])
		return nil, errs[0]
	}

	if result.StatusCode != 200 {
		errMsg := fmt.Errorf("call GetNodePool API error: code[%d], %s",
			result.StatusCode, string(body))
		return nil, errMsg
	}

	if resp.StatusCode != "ok" {
		blog.Errorf("GetNodePool failed, %s", string(body))
		return nil, fmt.Errorf("GetNodePool failed, %s", string(body))
	}

	if resp.ReturnObj == nil {
		blog.Errorf("GetNodePool lost nodepool info in response")
		return nil, fmt.Errorf("GetNodePool lost nodepool info in response")
	}

	return resp.ReturnObj.NodePool, nil
}

// ListNodePool lists nodepool
func (c *CTClient) ListNodePool(req *ListNodePoolReq) ([]*NodePoolV2, error) {
	reqPath := "/v2/nodePool/list"

	if req.ClusterID == "" {
		return nil, fmt.Errorf("ListNodePool empty clusterId")
	}

	params := fmt.Sprintf("clusterId=%s", req.ClusterID)
	if req.EnableAutoScaling != "" {
		params = fmt.Sprintf("%s&enableAutoScaling=%s", params, req.EnableAutoScaling)
	}
	if req.NodePoolName != "" {
		params = fmt.Sprintf("%s&nodePoolName=%s", params, req.NodePoolName)
	}
	if req.Page != 0 {
		params = fmt.Sprintf("%s&page=%d", params, req.Page)
	}
	if req.PerPage != 0 {
		params = fmt.Sprintf("%s&perPage=%d", params, req.PerPage)
	}
	if req.RetainSystemNodePool {
		params = fmt.Sprintf("%s&retainSystemNodePool=%v", params, req.RetainSystemNodePool)
	}

	resp := &ListNodePoolResponse{}

	requestID, _ := uuid.NewUUID()
	singerDate := time.Now()
	eopDate := singerDate.Format("20060102T150405Z")
	eopDt := singerDate.Format("20060102")

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Get(endingPoint+reqPath).
		Query(params).
		SetDebug(true).
		Set("Content-Type", "application/json").
		Set("User-Agent", "bcs-cluster-manager/v1.0").
		Set("Eop-date", eopDate).
		Set("ctyun-eop-request-id", requestID.String()).
		Set("Eop-Authorization", c.buildSignHeader(requestID.String(), eopDate, eopDt, "", params)).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call ListNodePool API failed: %v", errs[0])
		return nil, errs[0]
	}

	if result.StatusCode != 200 {
		errMsg := fmt.Errorf("call ListNodePool API error: code[%d], %s",
			result.StatusCode, string(body))
		return nil, errMsg
	}

	if resp.StatusCode != "ok" {
		blog.Errorf("ListNodePool failed, %s", string(body))
		return nil, fmt.Errorf("ListNodePool failed, %s", string(body))
	}

	if resp.ReturnObj == nil {
		blog.Errorf("ListNodePool lost nodepool info in response")
		return nil, fmt.Errorf("ListNodePool lost nodepool info in response")
	}

	return resp.ReturnObj.NodePools, nil
}

// ListVpcs lists vpcs
func (c *CTClient) ListVpcs(location string) ([]*Vpc, error) {
	reqPath := "/v2/cluster/listVpcs"
	params := fmt.Sprintf("nodeCode=%s", location)
	resp := &ListVpcResponse{}

	requestID, _ := uuid.NewUUID()
	singerDate := time.Now()
	eopDate := singerDate.Format("20060102T150405Z")
	eopDt := singerDate.Format("20060102")

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Get(endingPoint+reqPath).
		Query(params).
		SetDebug(true).
		Set("Content-Type", "application/json").
		Set("User-Agent", "bcs-cluster-manager/v1.0").
		Set("Eop-date", eopDate).
		Set("ctyun-eop-request-id", requestID.String()).
		Set("Eop-Authorization", c.buildSignHeader(requestID.String(), eopDate, eopDt, "", params)).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call ListVpcs API failed: %v", errs[0])
		return nil, errs[0]
	}

	if result.StatusCode != 200 {
		errMsg := fmt.Errorf("call ListVpcs API error: code[%d], %s",
			result.StatusCode, string(body))
		return nil, errMsg
	}

	if resp.StatusCode != "ok" {
		blog.Errorf("ListVpcs failed, %s", string(body))
		return nil, fmt.Errorf("ListVpcs failed, %s", string(body))
	}

	if resp.ReturnObj == nil {
		blog.Errorf("ListVpcs lost Vpcs info in response")
		return nil, fmt.Errorf("ListVpcs lost Vpcs info in response")
	}

	return resp.ReturnObj.Vpcs, nil
}

func (c *CTClient) buildSignHeader(requestID, eopDate, eopDt, body, params string) string {
	headerStr := fmt.Sprintf("ctyun-eop-request-id:%s\neop-date:%s\n", requestID, eopDate)
	bodyDigest := fmt.Sprintf("%x", sha256.Sum256([]byte(body)))
	signatureStr := fmt.Sprintf("%s\n%s\n%s", headerStr, params, bodyDigest)

	kTime := hmacSha256(eopDate, c.sk)
	kAk := hmacSha256(c.ak, kTime)
	kDate := hmacSha256(eopDt, kAk)

	signatureSha := hmacSha256(signatureStr, kDate)
	signature := base64.StdEncoding.EncodeToString([]byte(signatureSha))

	signHeader := fmt.Sprintf("%s Headers=ctyun-eop-request-id;eop-date Signature=%s", c.ak, signature)
	return signHeader
}

func hmacSha256(message, key string) string {
	hmac256 := hmac.New(sha256.New, []byte(key))
	hmac256.Write([]byte(message))
	return fmt.Sprintf("%s", hmac256.Sum(nil))
}
