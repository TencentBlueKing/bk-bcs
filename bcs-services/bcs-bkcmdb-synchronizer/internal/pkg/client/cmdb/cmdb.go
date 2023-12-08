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

// Package cmdb define client for cmdb
package cmdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	bkcmdbkube "configcenter/src/kube/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	pmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	cmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/parnurzeal/gorequest"
	"google.golang.org/grpc"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client"
)

// NewCmdbClient create cmdb client
func NewCmdbClient(options *Options) *cmdbClient {
	c := &cmdbClient{
		config: &Options{
			AppCode:    options.AppCode,
			AppSecret:  options.AppSecret,
			BKUserName: options.BKUserName,
			Server:     options.Server,
			Debug:      options.Debug,
		},
	}

	auth, err := c.generateGateWayAuth()
	if err != nil {
		return nil
	}
	c.userAuth = auth
	return c
}

// Cmdb returns the CMDB client instance.
func (c *cmdbClient) Cmdb() client.CMDBClient {
	return c
}

var (
	defaultTimeOut = time.Second * 60
	// ErrServerNotInit server not init
	ErrServerNotInit = errors.New("server not inited")
)

// Options for cmdb client
type Options struct {
	AppCode    string
	AppSecret  string
	BKUserName string
	Server     string
	Debug      bool
}

// AuthInfo auth user
type AuthInfo struct {
	BkAppCode   string `json:"bk_app_code"`
	BkAppSecret string `json:"bk_app_secret"`
	BkUserName  string `json:"bk_username"`
}

// Client for cc
type cmdbClient struct {
	config   *Options
	userAuth string
}

// GetDataManagerConnWithURL returns a gRPC client connection with URL for data manager.
func (c *cmdbClient) GetDataManagerConnWithURL() (*grpc.ClientConn, error) {
	//implement me
	panic("implement me")
}

// GetDataManagerConn returns a gRPC client connection for data manager.
func (c *cmdbClient) GetDataManagerConn() (*grpc.ClientConn, error) {
	//implement me
	panic("implement me")
}

// GetClusterManagerConnWithURL returns a gRPC client connection with URL for cluster manager.
func (c *cmdbClient) GetClusterManagerConnWithURL() (*grpc.ClientConn, error) {
	//implement me
	panic("implement me")
}

// GetClusterManagerClient returns a cluster manager client instance.
func (c *cmdbClient) GetClusterManagerClient() (cmp.ClusterManagerClient, error) {
	//implement me
	panic("implement me")
}

// GetClusterManagerConn returns a gRPC client connection for cluster manager.
func (c *cmdbClient) GetClusterManagerConn() (*grpc.ClientConn, error) {
	//implement me
	panic("implement me")
}

// NewCMGrpcClientWithHeader creates a new cluster manager gRPC client with header.
func (c *cmdbClient) NewCMGrpcClientWithHeader(ctx context.Context,
	conn *grpc.ClientConn) *client.ClusterManagerClientWithHeader {
	//implement me
	panic("implement me")
}

// GetProjectManagerConnWithURL returns a gRPC client connection with URL for project manager.
func (c *cmdbClient) GetProjectManagerConnWithURL() (*grpc.ClientConn, error) {
	//implement me
	panic("implement me")
}

// GetProjectManagerClient returns a project manager client instance.
func (c *cmdbClient) GetProjectManagerClient() (pmp.BCSProjectClient, error) {
	//implement me
	panic("implement me")
}

// GetProjectManagerConn returns a gRPC client connection for project manager.
func (c *cmdbClient) GetProjectManagerConn() (*grpc.ClientConn, error) {
	//implement me
	panic("implement me")
}

// NewPMGrpcClientWithHeader creates a new project manager gRPC client with header.
func (c *cmdbClient) NewPMGrpcClientWithHeader(ctx context.Context,
	conn *grpc.ClientConn) *client.ProjectManagerClientWithHeader {
	//implement me
	panic("implement me")
}

// GetStorageClient returns a storage client instance.
func (c *cmdbClient) GetStorageClient() (bcsapi.Storage, error) {
	//implement me
	panic("implement me")
}

func (c *cmdbClient) generateGateWayAuth() (string, error) {
	if c == nil {
		return "", ErrServerNotInit
	}

	auth := &AuthInfo{
		BkAppCode:   c.config.AppCode,
		BkAppSecret: c.config.AppSecret,
		BkUserName:  c.config.BKUserName,
	}

	userAuth, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	return string(userAuth), nil
}

// GetBS2IDByBizID get bs2ID by bizID
func (c *cmdbClient) GetBS2IDByBizID(bizID int64) (int, error) {
	if c == nil {
		return 0, ErrServerNotInit
	}

	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/search_business/", c.config.Server)
		request = &client.SearchBusinessRequest{
			Fields: []string{client.FieldBS2NameID},
			Condition: map[string]interface{}{
				client.ConditionBkBizID: bizID,
			},
		}
		respData = &client.SearchBusinessResponse{}
	)

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api GetBS2IDByBizID failed: %v", errs[0])
		return 0, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api GetBS2IDByBizID failed: %v", respData.Message)
		return 0, fmt.Errorf(respData.Message)
	}
	//successfully request
	blog.Infof("call api GetBS2IDByBizID with url(%s) successfully", reqURL)

	if len(respData.Data.Info) > 0 {
		return respData.Data.Info[0].BS2NameID, nil
	}

	return 0, fmt.Errorf("call api GetBS2IDByBizID failed")
}

// GetBizInfo get biz Info
func (c *cmdbClient) GetBizInfo(bizID int64) (*client.Business, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	var (
		reqURL  = fmt.Sprintf("%s/component/compapi/cmdb/get_query_info/", c.config.Server)
		request = &client.QueryBusinessInfoReq{
			Method:    client.MethodBusinessRaw,
			ReqColumn: client.ReqColumns,
			KeyValues: map[string]interface{}{
				client.KeyBizID: bizID,
			},
		}
		respData = &client.QueryBusinessInfoResp{}
	)

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api GetBizInfo failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api GetBizInfo failed: %v", respData.Message)
		return nil, fmt.Errorf(respData.Message)
	}
	//successfully request
	blog.Infof("call api GetBizInfo with url(%s) successfully", reqURL)

	if len(respData.Data.Data) > 0 {
		return &respData.Data.Data[0], nil
	}

	return nil, fmt.Errorf("call api GetBizInfo failed")
}

// GetHostInfo get host Info
func (c *cmdbClient) GetHostInfo(hostIP []string) (*[]client.HostData, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	var hostData []client.HostData
	pageStart := 0

	for {

		from := pageStart * 100
		to := (pageStart + 1) * 100

		if len(hostIP) < to {
			to = len(hostIP)
		}

		var (
			reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/list_hosts_without_biz/", c.config.Server)
			request = &client.ListHostsWithoutBizRequest{
				Page: client.Page{
					Limit: 100,
					Start: pageStart,
				},
				Fields: []string{
					"bk_host_innerip",
					"svr_type_name",
					"bk_svr_device_cls_name",
					"bk_service_arr",
					"bk_svc_id_arr",
					"idc_city_id",
					"idc_city_name",
					"bk_host_id",
				},
				HostPropertyFilter: client.PropertyFilter{
					Condition: "AND",
					Rules: []client.Rule{
						{
							Field:    "bk_host_innerip",
							Operator: "in",
							Value:    hostIP[from:to],
						},
					},
				},
			}

			respData = &client.ListHostsWithoutBizResponse{}
		)

		_, _, errs := gorequest.New().
			Timeout(defaultTimeOut).
			Post(reqURL).
			Set("Content-Type", "application/json").
			Set("Accept", "application/json").
			Set("X-Bkapi-Authorization", c.userAuth).
			SetDebug(c.config.Debug).
			Send(request).
			EndStruct(&respData)
		if len(errs) > 0 {
			blog.Errorf("call api QueryHost failed: %v", errs[0])
			return nil, errs[0]
		}

		if !respData.Result {
			blog.Errorf("call api QueryHost failed: %v", respData.Message)
			return nil, fmt.Errorf(respData.Message)
		}
		//successfully request
		blog.Infof("call api QueryHost with url(%s) successfully", reqURL)

		hostData = append(hostData, respData.Data.Info...)

		if len(hostIP) == to {
			break
		}
		pageStart++
	}

	return &hostData, nil
}

// GetBcsCluster get bcs cluster
// /api/v3/kube/findmany/cluster/bk_biz_id/{bk_biz_id}
// /v2/cc/list_kube_cluster/
func (c *cmdbClient) GetBcsCluster(request *client.GetBcsClusterRequest) (*[]bkcmdbkube.Cluster, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/list_kube_cluster/", c.config.Server)
	respData := &client.GetBcsClusterResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_kube_cluster failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_kube_cluster failed: %v", respData.Message)
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_kube_cluster with url(%s) (%v) successfully", reqURL, request)

	return &respData.Data.Info, nil
}

// CreateBcsCluster create bcs cluster
// /api/v3/kube/create/cluster/bk_biz_id/{bk_biz_id}
// /v2/cc/create_kube_cluster/
func (c *cmdbClient) CreateBcsCluster(request *client.CreateBcsClusterRequest) (bkClusterID int64, err error) {
	if c == nil {
		return bkClusterID, ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/create_kube_cluster/", c.config.Server)
	respData := &client.CreateBcsClusterResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api create_kube_cluster failed: %v", errs[0])
		return bkClusterID, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api create_kube_cluster failed: %v", respData.Message)
		return bkClusterID, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api create_kube_cluster with url(%s) (%v)  successfully", reqURL, request)

	bkClusterID = respData.Data.ID

	return bkClusterID, nil
}

// UpdateBcsCluster update bcs cluster
// /api/v3/kube/updatemany/cluster/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_update_kube_cluster/
func (c *cmdbClient) UpdateBcsCluster(request *client.UpdateBcsClusterRequest) error {
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/batch_update_kube_cluster/", c.config.Server)
	respData := &client.UpdateBcsClusterResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_update_kube_cluster failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_update_kube_cluster failed: %v", respData.Message)
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_update_kube_cluster with url(%s) (%v) successfully", reqURL, request)

	return nil
}

// UpdateBcsClusterType update bcs cluster type
// /api/v3/update/kube/cluster/type
// /v2/cc/update_kube_cluster_type/
func (c *cmdbClient) UpdateBcsClusterType(request *client.UpdateBcsClusterTypeRequest) error {
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/update_kube_cluster_type/", c.config.Server)
	respData := &client.UpdateBcsClusterTypeResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api update_kube_cluster_type failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api update_kube_cluster_type failed: %v", respData.Message)
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api update_kube_cluster_type with url(%s) (%v) successfully", reqURL, request)

	return nil
}

// DeleteBcsCluster delete bcs cluster
// /api/v3/kube/delete/cluster/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_delete_kube_cluster/
func (c *cmdbClient) DeleteBcsCluster(request *client.DeleteBcsClusterRequest) error {
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/batch_delete_kube_cluster/", c.config.Server)
	respData := &client.DeleteBcsClusterResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_kube_cluster failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_delete_kube_cluster failed: %v", respData.Message)
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_kube_cluster with url(%s) (%v) successfully", reqURL, request)

	return nil
}

// GetBcsNamespace get bcs namespace
// /api/v3/kube/findmany/namespace/bk_biz_id/{bk_biz_id}
// /v2/cc/list_namespace/
func (c *cmdbClient) GetBcsNamespace(request *client.GetBcsNamespaceRequest) (*[]bkcmdbkube.Namespace, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/list_kube_namespace/", c.config.Server)

	respData := &client.GetBcsNamespaceResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_namespace failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_namespace failed: %v", respData.Message)
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_namespace with url(%s) (%v) successfully", reqURL, request)

	return respData.Data.Info, nil
}

// CreateBcsNamespace create bcs namespace
// /api/v3/kube/createmany/namespace/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_create_namespace/
func (c *cmdbClient) CreateBcsNamespace(request *client.CreateBcsNamespaceRequest) (*[]int64, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/batch_create_kube_namespace/", c.config.Server)
	respData := &client.CreateBcsNamespaceResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api create_kube_namespace failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api create_kube_namespace failed: %v", respData.Message)
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api create_kube_namespace with url(%s) (%v) successfully", reqURL, request)

	return &respData.Data.IDs, nil
}

// UpdateBcsNamespace update bcs namespace
// /api/v3/kube/updatemany/namespace/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_update_namespace/
func (c *cmdbClient) UpdateBcsNamespace(request *client.UpdateBcsNamespaceRequest) error {
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/batch_update_kube_namespace/", c.config.Server)
	respData := &client.UpdateBcsNamespaceResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_update_namespace failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_update_namespace failed: %v", respData.Message)
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_update_namespace with url(%s) (%v) successfully", reqURL, request)

	return nil
}

// DeleteBcsNamespace delete bcs namespace
// /api/v3/kube/deletemany/namespace/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_delete_namespace/
func (c *cmdbClient) DeleteBcsNamespace(request *client.DeleteBcsNamespaceRequest) error {
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/batch_delete_kube_namespace/", c.config.Server)
	respData := &client.DeleteBcsNamespaceResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_namespace failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_delete_namespace failed: %v", respData.Message)
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_namespace with url(%s) (%v) successfully", reqURL, request)

	return nil
}

// GetBcsWorkload get bcs workload
// /api/v3/kube/findmany/workload/{kind}/{bk_biz_id}
// /v2/cc/list_workload/
func (c *cmdbClient) GetBcsWorkload(request *client.GetBcsWorkloadRequest) (*[]interface{}, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/list_kube_workload/", c.config.Server)
	respData := &client.GetBcsWorkloadResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_workload failed: %v, rid: %s", errs[0], respData.RequestID)
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_workload failed: %v, rid: %s", respData.Message, respData.RequestID)
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_workload with url(%s) (%v) successfully", reqURL, request)

	return &respData.Data.Info, nil
}

// CreateBcsWorkload create bcs workload
// /api/v3/kube/createmany/workload/{kind}/{bk_biz_id}
// /v2/cc/batch_create_workload/
func (c *cmdbClient) CreateBcsWorkload(request *client.CreateBcsWorkloadRequest) (*[]int64, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/batch_create_kube_workload/", c.config.Server)
	respData := &client.CreateBcsWorkloadResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_create_workload failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_create_workload failed: %v", respData.Message)
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_create_workload with url(%s) (%v) successfully", reqURL, request)

	return &respData.Data.IDs, nil
}

// UpdateBcsWorkload update bcs workload
// /api/v3/kube/updatemany/workload/{kind}/{bk_biz_id}
// /v2/cc/batch_update_workload/
func (c *cmdbClient) UpdateBcsWorkload(request *client.UpdateBcsWorkloadRequest) error {
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/batch_update_kube_workload/", c.config.Server)
	respData := &client.UpdateBcsWorkloadResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_update_workload failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_update_workload failed: %v", respData.Message)
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_update_workload with url(%s) (%v) successfully", reqURL, request)

	return nil
}

// DeleteBcsWorkload delete bcs workload
// /api/v3/kube/deletemany/workload/{kind}/{bk_biz_id}
// /v2/cc/batch_delete_workload/
func (c *cmdbClient) DeleteBcsWorkload(request *client.DeleteBcsWorkloadRequest) error {
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/batch_delete_kube_workload/", c.config.Server)
	respData := &client.DeleteBcsWorkloadResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_workload failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_delete_workload failed: %v", respData.Message)
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_workload with url(%s) (%v) successfully", reqURL, request)

	return nil
}

// GetBcsNode get bcs node
// /api/v3/kube/findmany/node/bk_biz_id/{bk_biz_id}
// /v2/cc/list_kube_node/
func (c *cmdbClient) GetBcsNode(request *client.GetBcsNodeRequest) (*[]bkcmdbkube.Node, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/list_kube_node/", c.config.Server)
	respData := &client.GetBcsNodeResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_kube_node failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_kube_node failed: %v", respData.Message)
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_kube_node with url(%s) (%v) successfully", reqURL, request)

	return respData.Data.Info, nil
}

// CreateBcsNode create bcs node
// /api/v3/kube/createmany/node/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_create_kube_node/
func (c *cmdbClient) CreateBcsNode(request *client.CreateBcsNodeRequest) (*[]int64, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/batch_create_kube_node/", c.config.Server)
	respData := &client.CreateBcsNodeResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_create_kube_node failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_create_kube_node failed: %v", respData.Message)
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_create_kube_node with url(%s) (%v) successfully", reqURL, request)

	return &respData.Data.IDs, nil
}

// UpdateBcsNode update bcs node
// /api/v3/kube/updatemany/node/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_update_kube_node/
func (c *cmdbClient) UpdateBcsNode(request *client.UpdateBcsNodeRequest) error {
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/batch_update_kube_node/", c.config.Server)
	respData := &client.UpdateBcsNodeResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_update_kube_node failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_update_kube_node failed: %v", respData.Message)
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_update_kube_node with url(%s) (%v) successfully", reqURL, request)

	return nil
}

// DeleteBcsNode delete bcs node
// /api/v3/kube/deletemany/node/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_delete_kube_node/
func (c *cmdbClient) DeleteBcsNode(request *client.DeleteBcsNodeRequest) error {
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/batch_delete_kube_node/", c.config.Server)
	respData := &client.DeleteBcsNodeResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_kube_node failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_delete_kube_node failed: %v", respData.Message)
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_kube_node with url(%s) (%v) successfully", reqURL, request)

	return nil
}

// GetBcsPod get bcs pod
// /api/v3/kube/findmany/pod/bk_biz_id/{bk_biz_id}
// /v2/cc/list_pod/
func (c *cmdbClient) GetBcsPod(request *client.GetBcsPodRequest) (*[]bkcmdbkube.Pod, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/list_kube_pod/", c.config.Server)
	respData := &client.GetBcsPodResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_pod failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_pod failed: %v", respData.Message)
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_pod with url(%s) (%v) successfully", reqURL, request)

	return respData.Data.Info, nil
}

// CreateBcsPod create bcs pod
// /api/v3/kube/createmany/pod/
// /v2/cc/batch_create_kube_pod/
func (c *cmdbClient) CreateBcsPod(request *client.CreateBcsPodRequest) (*[]int64, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/batch_create_kube_pod/", c.config.Server)
	respData := &client.CreateBcsPodResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_create_kube_pod failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_create_kube_pod failed: %v", respData.Message)
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_create_kube_pod with url(%s) (%v) successfully", reqURL, request)

	return &respData.Data.IDs, nil
}

// DeleteBcsPod delete bcs pod
// /api/v3/kube/deletemany/pod/
// /v2/cc/batch_delete_kube_pod/
func (c *cmdbClient) DeleteBcsPod(request *client.DeleteBcsPodRequest) error {
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/batch_delete_kube_pod/", c.config.Server)
	respData := &client.DeleteBcsPodResponse{}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_kube_pod failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_delete_kube_pod failed: %v", respData.Message)
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_kube_pod with url(%s) (%v) successfully", reqURL, request)

	return nil
}

// GetCMDBClient get cmdb client
func (c *cmdbClient) GetCMDBClient() (client.CMDBClient, error) {
	cli := NewCmdbClient(c.config)
	return cli.Cmdb(), nil
}
