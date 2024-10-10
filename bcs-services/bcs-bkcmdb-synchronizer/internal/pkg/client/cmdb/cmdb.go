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

	bkcmdbkube "configcenter/src/kube/types" // nolint
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	pmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	cmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/parnurzeal/gorequest"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/store/model"
)

// 定义常量表示不同的Kubernetes资源类型
const (
	// DaemonSet 表示守护进程集
	DaemonSet = "daemonset"
	// GameStatefulSet 表示游戏有状态集
	GameStatefulSet = "gameStatefulSet"
	// Pods 表示Pod集合
	Pods = "pods"
	// GameDeployment 表示游戏部署
	GameDeployment = "gameDeployment"
	// StatefulSet 表示有状态集
	StatefulSet = "state7efulSet"
	// Deployment 表示部署
	Deployment = "deployment"
	// And 表示逻辑与操作
	And = "AND"
	// OR 表示逻辑或操作
	OR = "OR"
)

// NewCmdbClient create cmdb client
// nolint
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
	// implement me
	panic("implement me")
}

// GetDataManagerConn returns a gRPC client connection for data manager.
func (c *cmdbClient) GetDataManagerConn() (*grpc.ClientConn, error) {
	// implement me
	panic("implement me")
}

// GetClusterManagerConnWithURL returns a gRPC client connection with URL for cluster manager.
func (c *cmdbClient) GetClusterManagerConnWithURL() (*grpc.ClientConn, error) {
	// implement me
	panic("implement me")
}

// GetClusterManagerClient returns a cluster manager client instance.
func (c *cmdbClient) GetClusterManagerClient() (cmp.ClusterManagerClient, error) {
	// implement me
	panic("implement me")
}

// GetClusterManagerConn returns a gRPC client connection for cluster manager.
func (c *cmdbClient) GetClusterManagerConn() (*grpc.ClientConn, error) {
	// implement me
	panic("implement me")
}

// NewCMGrpcClientWithHeader creates a new cluster manager gRPC client with header.
func (c *cmdbClient) NewCMGrpcClientWithHeader(ctx context.Context,
	conn *grpc.ClientConn) *client.ClusterManagerClientWithHeader {
	// implement me
	panic("implement me")
}

// GetProjectManagerConnWithURL returns a gRPC client connection with URL for project manager.
func (c *cmdbClient) GetProjectManagerConnWithURL() (*grpc.ClientConn, error) {
	// implement me
	panic("implement me")
}

// GetProjectManagerClient returns a project manager client instance.
func (c *cmdbClient) GetProjectManagerClient() (pmp.BCSProjectClient, error) {
	// implement me
	panic("implement me")
}

// GetProjectManagerConn returns a gRPC client connection for project manager.
func (c *cmdbClient) GetProjectManagerConn() (*grpc.ClientConn, error) {
	// implement me
	panic("implement me")
}

// NewPMGrpcClientWithHeader creates a new project manager gRPC client with header.
func (c *cmdbClient) NewPMGrpcClientWithHeader(ctx context.Context,
	conn *grpc.ClientConn) *client.ProjectManagerClientWithHeader {
	// implement me
	panic("implement me")
}

// GetStorageClient returns a storage client instance.
func (c *cmdbClient) GetStorageClient() (bcsapi.Storage, error) {
	// implement me
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
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api GetBS2IDByBizID failed: %v", errs[0])
		return 0, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api GetBS2IDByBizID failed: %v, rid: %s", respData.Message, respData.RequestID)
		return 0, fmt.Errorf(respData.Message)
	}
	// successfully request
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
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api GetBizInfo failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api GetBizInfo failed: %v, rid: %s", respData.Message, respData.RequestID)
		return nil, fmt.Errorf(respData.Message)
	}
	// successfully request
	blog.Infof("call api GetBizInfo with url(%s) successfully", reqURL)

	if len(respData.Data.Data) > 0 {
		return &respData.Data.Data[0], nil
	}

	return nil, fmt.Errorf("call api GetBizInfo failed")
}

// GetHostInfo 根据提供的IP地址列表获取主机信息
// nolint
func (c *cmdbClient) GetHostInfo(hostIP []string) (*[]client.HostData, error) {
	// 检查客户端是否已初始化
	if c == nil {
		return nil, ErrServerNotInit
	}

	// 初始化主机数据切片
	var hostData []client.HostData
	// 初始化分页起始位置
	pageStart := 0

	// 循环处理所有IP地址
	for {

		// 计算当前分页的起始和结束索引
		from := pageStart * 100
		to := (pageStart + 1) * 100

		// 如果剩余IP地址不足100个，则调整结束索引
		if len(hostIP) < to {
			to = len(hostIP)
		}

		// 构建请求URL和请求体
		var (
			reqURL  = fmt.Sprintf("%s/api/v3/hosts/list_hosts_without_app", c.config.Server)
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
					Condition: And,
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

		// 发送HTTP请求并处理响应
		resp, _, errs := gorequest.New().
			Timeout(defaultTimeOut).
			Post(reqURL).
			Set("Content-Type", "application/json").
			Set("Accept", "application/json").
			Set("X-Bkapi-Authorization", c.userAuth).
			SetDebug(c.config.Debug).
			Send(request).
			Retry(3, 3*time.Second, 429).
			EndStruct(&respData)
		// 检查是否有错误发生
		if len(errs) > 0 {
			blog.Errorf("call api QueryHost failed: %v", errs[0])
			return nil, errs[0]
		}

		// 检查API响应是否成功
		if !respData.Result {
			blog.Errorf("call api QueryHost failed: %v, rid: %s", respData.Message, resp.Header.Get("X-Request-Id"))
			return nil, fmt.Errorf(respData.Message)
		}
		// 请求成功，记录日志
		blog.Infof("call api QueryHost with url(%s) successfully, X-Request-Id: %s", reqURL, resp.Header.Get("X-Request-Id"))

		// 将响应数据添加到主机数据切片中
		hostData = append(hostData, respData.Data.Info...)

		// 如果所有IP地址都已处理完毕，则退出循环
		if len(hostIP) == to {
			break
		}
		// 更新分页起始位置
		pageStart++
	}

	// 返回收集到的主机数据
	return &hostData, nil
}

// GetHostsByBiz 根据业务ID和主机IP列表获取主机数据
// nolint
func (c *cmdbClient) GetHostsByBiz(bkBizID int64, hostIP []string) (*[]client.HostData, error) {
	// 检查客户端是否初始化
	if c == nil {
		return nil, ErrServerNotInit
	}

	var hostData []client.HostData // 存储获取到的主机数据
	pageStart := 0                 // 分页起始位置

	// 循环获取所有分页数据
	for {

		from := pageStart * 100     // 当前分页的起始索引
		to := (pageStart + 1) * 100 // 当前分页的结束索引

		// 如果剩余IP数量小于当前分页大小，则调整结束索引
		if len(hostIP) < to {
			to = len(hostIP)
		}

		// 构建请求URL和请求体
		var (
			reqURL  = fmt.Sprintf("%s/api/v3/hosts/app/%d/list_hosts", c.config.Server, bkBizID)
			request = &client.ListHostsByBizRequest{
				Page: client.Page{
					Limit: 100,       // 每页限制100条记录
					Start: pageStart, // 分页起始位置
				},
				Fields: []string{
					// 请求返回的字段列表
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
					Condition: And, // 条件为AND
					Rules: []client.Rule{
						{
							Field:    "bk_host_innerip", // 过滤字段为主机内网IP
							Operator: "in",              // 操作符为IN
							Value:    hostIP[from:to],   // 过滤值为当前分页的IP列表
						},
					},
				},
			}

			respData = &client.ListHostsWithoutBizResponse{} // 响应数据结构体
		)

		// 发送HTTP请求并解析响应
		resp, _, errs := gorequest.New().
			Timeout(defaultTimeOut).
			Post(reqURL).
			Set("Content-Type", "application/json").
			Set("Accept", "application/json").
			Set("X-Bkapi-Authorization", c.userAuth).
			SetDebug(c.config.Debug).
			Send(request).
			Retry(3, 3*time.Second, 429).
			EndStruct(&respData)

		// 检查请求是否有错误
		if len(errs) > 0 {
			blog.Errorf("call api QueryHost failed: %v", errs[0])
			return nil, errs[0]
		}

		// 检查响应结果是否成功
		if !respData.Result {
			blog.Errorf("call api QueryHost failed: %v, rid: %s",
				respData.Message, resp.Header.Get("X-Request-Id"))
			return nil, fmt.Errorf(respData.Message)
		}
		// 成功获取数据，记录日志
		blog.Infof("call api QueryHost with url(%s) successfully, X-Request-Id: %s",
			reqURL, resp.Header.Get("X-Request-Id"))

		// 将当前分页的数据添加到总数据列表中
		hostData = append(hostData, respData.Data.Info...)

		// 如果所有IP都已处理完毕，则退出循环
		if len(hostIP) == to {
			break
		}
		pageStart++ // 更新分页起始位置
	}

	return &hostData, nil // 返回获取到的主机数据列表
}

// GetBcsCluster 根据请求获取BCS集群信息，可以选择是否通过数据库进行查询和处理
// /api/v3/kube/findmany/cluster/bk_biz_id/{bk_biz_id}
// /v2/cc/list_kube_cluster/
// nolint
func (c *cmdbClient) GetBcsCluster(
	request *client.GetBcsClusterRequest, db *gorm.DB, withDB bool) (*[]bkcmdbkube.Cluster, error) {
	// 检查客户端是否初始化
	if c == nil {
		return nil, ErrServerNotInit
	}

	// 如果指定了使用数据库并且db对象不为空，则通过数据库查询集群信息
	if withDB && db != nil {
		// 创建一个新的数据库会话
		query := db.Session(&gorm.Session{NewDB: true})
		// 遍历请求中的过滤规则，并根据条件拼接查询语句
		for _, rule := range request.Filter.Rules {
			if request.Filter.Condition == And {
				query = query.Where(fmt.Sprintf("%s %s ?", rule.Field, rule.Operator), rule.Value)
			}
			if request.Filter.Condition == OR {
				query = query.Or(query.Where(fmt.Sprintf("%s %s ?", rule.Field, rule.Operator), rule.Value))
			}
		}

		// 执行查询并将结果存储在cluster变量中
		var cluster []model.Cluster
		if err := query.Find(&cluster).Error; err != nil {
			blog.Errorf("query cluster withDB failed: %v", err)
		} else {
			// 将查询结果序列化为JSON
			if clusterMarshal, err := json.Marshal(cluster); err != nil {
				blog.Errorf("marshal cluster failed: %v", err)
			} else {
				// 将JSON反序列化为bkcmdbkube.Cluster类型的切片
				var bkCluster []bkcmdbkube.Cluster
				errM := json.Unmarshal(clusterMarshal, &bkCluster)
				if errM == nil {
					blog.Infof("GetBcsCluster clusterWithDB get")
					return &bkCluster, nil
				}
				blog.Errorf("unmarshal cluster failed: %v", errM)
			}
		}
	}

	// 如果没有通过数据库查询，则通过API调用获取集群信息
	reqURL := fmt.Sprintf("%s/api/v3/findmany/kube/cluster", c.config.Server)
	respData := &client.GetBcsClusterResponse{}
	// 使用gorequest库发送POST请求，并处理响应
	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	// 检查是否有错误发生
	if len(errs) > 0 {
		blog.Errorf("call api list_kube_cluster failed: %v", errs[0])
		return nil, errs[0]
	}
	// 检查API响应是否成功
	if !respData.Result {
		blog.Errorf("call api list_kube_cluster failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}
	blog.Infof("call api list_kube_cluster with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	// 如果db不为空，则将API获取的集群信息同步到数据库
	if db != nil {
		if clusterMarshal, err := json.Marshal(respData.Data.Info); err != nil {
			blog.Errorf("marshal cluster failed: %v", err)
		} else {
			var cluster []model.Cluster
			if err := json.Unmarshal(clusterMarshal, &cluster); err != nil {
				blog.Errorf("unmarshal cluster failed: %v", err)
			} else {
				// 遍历集群信息，检查数据库中是否存在，不存在则创建，存在则更新
				for _, cl := range cluster {
					var existingCl model.Cluster
					query := db.Session(&gorm.Session{NewDB: true})
					if errC := query.Where("id = ?", cl.ID).First(&existingCl).Error; errC != nil {
						if errors.Is(errC, gorm.ErrRecordNotFound) {
							if errCC := db.Create(&cl).Error; errCC != nil {
								blog.Errorf("create cluster failed: %v", errCC)
							}
							blog.Infof("GetBcsCluster clusterWithDB create")
						} else {
							blog.Errorf("get cluster failed: %v", errC)
						}
					} else {
						if errCS := query.Save(&cl).Error; errCS != nil {
							blog.Errorf("update cluster failed: %v", errCS)
						} else {
							blog.Infof("GetBcsCluster clusterWithDB update: %s", cl.Name)
						}
					}
				}
			}
		}
	}
	// 返回获取到的集群信息
	return &respData.Data.Info, nil
}

// CreateBcsCluster create bcs cluster
// /api/v3/kube/create/cluster/bk_biz_id/{bk_biz_id}
// /v2/cc/create_kube_cluster/
// CreateBcsCluster 创建BCS集群
func (c *cmdbClient) CreateBcsCluster(
	request *client.CreateBcsClusterRequest, db *gorm.DB) (bkClusterID int64, err error) {
	// 检查客户端是否初始化，未初始化则返回错误
	if c == nil {
		return bkClusterID, ErrServerNotInit
	}

	// 构造请求URL
	reqURL := fmt.Sprintf("%s/api/v3/create/kube/cluster", c.config.Server)
	// 初始化响应数据结构
	respData := &client.CreateBcsClusterResponse{}

	// 使用gorequest库发送POST请求
	resp, _, errs := gorequest.New().
		// 设置请求超时时间
		Timeout(defaultTimeOut).
		// 指定请求方法和URL
		Post(reqURL).
		// 设置请求头，包括Content-Type, Accept和X-Bkapi-Authorization
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		// 根据配置决定是否开启调试模式
		SetDebug(c.config.Debug).
		// 发送请求体
		Send(request).
		// 设置重试策略
		Retry(3, 3*time.Second, 429).
		// 解析响应到respData结构体
		EndStruct(&respData)
	// 检查是否有错误发生
	if len(errs) > 0 {
		blog.Errorf("call api create_kube_cluster failed: %v", errs[0])
		return bkClusterID, errs[0]
	}

	// 检查响应结果是否成功
	if !respData.Result {
		blog.Errorf("call api create_kube_cluster failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return bkClusterID, fmt.Errorf(respData.Message)
	}

	// 记录成功的日志信息
	blog.Infof("call api create_kube_cluster with url(%s) (%v)  successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	// 获取创建的集群ID
	bkClusterID = respData.Data.ID

	// 根据集群ID查询集群信息，验证创建是否成功
	_, err = c.GetBcsCluster(&client.GetBcsClusterRequest{
		CommonRequest: client.CommonRequest{
			BKBizID: *request.BKBizID,
			Page: client.Page{
				Limit: 100,
				Start: 0,
			},
			Filter: &client.PropertyFilter{
				Condition: And,
				Rules: []client.Rule{
					{
						Field:    "id",
						Operator: "in",
						Value:    []int64{bkClusterID},
					},
				},
			},
		},
	}, db, false)
	// 如果查询失败，记录错误日志
	if err != nil {
		blog.Errorf("get cluster failed: %v", err)
	}

	// 返回创建的集群ID和可能的错误
	return bkClusterID, nil
}

// UpdateBcsCluster update bcs cluster
// /api/v3/kube/updatemany/cluster/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_update_kube_cluster/
// UpdateBcsCluster 更新BCS集群信息
func (c *cmdbClient) UpdateBcsCluster(request *client.UpdateBcsClusterRequest, db *gorm.DB) error {
	// 检查客户端是否初始化，未初始化则返回错误
	if c == nil {
		return ErrServerNotInit
	}

	// 构造请求URL
	reqURL := fmt.Sprintf("%s/api/v3/updatemany/kube/cluster", c.config.Server)
	// 初始化响应数据结构
	respData := &client.UpdateBcsClusterResponse{}

	// 使用gorequest库发送PUT请求
	resp, _, errs := gorequest.New().
		// 设置请求超时时间
		Timeout(defaultTimeOut).
		// 指定HTTP方法和请求URL
		Put(reqURL).
		// 设置请求头部信息
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		// 开启调试模式（如果配置中开启）
		SetDebug(c.config.Debug).
		// 发送请求体数据
		Send(request).
		// 设置重试机制
		Retry(3, 3*time.Second, 429).
		// 解析响应到respData结构体
		EndStruct(&respData)
	// 检查是否有请求错误发生
	if len(errs) > 0 {
		blog.Errorf("call api batch_update_kube_cluster failed: %v", errs[0])
		return errs[0]
	}

	// 如果响应结果为false，记录错误日志并返回错误信息
	if !respData.Result {
		blog.Errorf("call api batch_update_kube_cluster failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	// 记录成功调用API的日志信息
	blog.Infof("call api batch_update_kube_cluster with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	// 调用GetBcsCluster方法获取更新后的集群信息
	_, err := c.GetBcsCluster(&client.GetBcsClusterRequest{
		CommonRequest: client.CommonRequest{
			BKBizID: *request.BKBizID,
			Page: client.Page{
				Limit: 100,
				Start: 0,
			},
			Filter: &client.PropertyFilter{
				Condition: And,
				Rules: []client.Rule{
					{
						Field:    "id",
						Operator: "in",
						Value:    *request.IDs,
					},
				},
			},
		},
	}, db, false)
	// 如果获取集群信息失败，记录错误日志
	if err != nil {
		blog.Errorf("get cluster failed: %v", err)
	}

	// 返回nil表示函数执行成功
	return nil
}

// UpdateBcsClusterType update bcs cluster type
// /api/v3/update/kube/cluster/type
// /v2/cc/update_kube_cluster_type/
// UpdateBcsClusterType 更新BCS集群类型的函数
func (c *cmdbClient) UpdateBcsClusterType(request *client.UpdateBcsClusterTypeRequest, db *gorm.DB) error {
	// 检查cmdbClient是否已初始化
	if c == nil {
		return ErrServerNotInit
	}

	// 构造请求URL
	reqURL := fmt.Sprintf("%s/api/v3/update/kube/cluster/type", c.config.Server)
	// 初始化响应数据结构
	respData := &client.UpdateBcsClusterTypeResponse{}

	// 使用gorequest库发送PUT请求
	resp, _, errs := gorequest.New().
		// 设置请求超时时间
		Timeout(defaultTimeOut).
		// 指定HTTP方法为PUT
		Put(reqURL).
		// 设置请求头Content-Type为application/json
		Set("Content-Type", "application/json").
		// 设置请求头Accept为application/json
		Set("Accept", "application/json").
		// 设置自定义认证头
		Set("X-Bkapi-Authorization", c.userAuth).
		// 根据配置决定是否开启调试模式
		SetDebug(c.config.Debug).
		// 发送请求体
		Send(request).
		// 设置重试策略：最多重试3次，每次间隔3秒，遇到429状态码也重试
		Retry(3, 3*time.Second, 429).
		// 将响应体解析到respData结构体中
		EndStruct(&respData)
	// 检查是否有错误发生
	if len(errs) > 0 {
		blog.Errorf("call api update_kube_cluster_type failed: %v", errs[0])
		return errs[0]
	}

	// 检查响应数据中的Result字段，如果不为true则表示API调用失败
	if !respData.Result {
		blog.Errorf("call api update_kube_cluster_type failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	// 记录API调用成功的日志，包括请求URL、请求体和请求ID
	blog.Infof("call api update_kube_cluster_type with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	// 调用GetBcsCluster函数获取更新后的集群信息
	_, err := c.GetBcsCluster(&client.GetBcsClusterRequest{
		CommonRequest: client.CommonRequest{
			BKBizID: *request.BKBizID,
			Page: client.Page{
				Limit: 100,
				Start: 0,
			},
			Filter: &client.PropertyFilter{
				Condition: And,
				Rules: []client.Rule{
					{
						Field:    "id",
						Operator: "in",
						Value:    []int64{*request.ID},
					},
				},
			},
		},
	}, db, false)
	// 如果获取集群信息失败，则记录错误日志
	if err != nil {
		blog.Errorf("get cluster failed: %v", err)
	}

	// 返回nil表示函数执行成功
	return nil
}

// DeleteBcsCluster delete bcs cluster
// /api/v3/kube/delete/cluster/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_delete_kube_cluster/
// DeleteBcsCluster 方法用于删除指定的 Kubernetes 集群。
// 它首先检查客户端是否已初始化，然后构造请求 URL 并发送 HTTP DELETE 请求到该 URL。
// 如果请求成功，它还会尝试从数据库中删除对应的集群记录。
func (c *cmdbClient) DeleteBcsCluster(request *client.DeleteBcsClusterRequest, db *gorm.DB) error {
	// 检查客户端是否已初始化，如果未初始化则返回错误
	if c == nil {
		return ErrServerNotInit
	}

	// 构造请求的 URL
	reqURL := fmt.Sprintf("%s/api/v3/delete/kube/cluster", c.config.Server)
	// 初始化响应数据结构
	respData := &client.DeleteBcsClusterResponse{}

	// 使用 gorequest 库发送 HTTP DELETE 请求
	// 设置请求超时时间、内容类型、接受类型、授权头等信息
	// 发送请求体并尝试重试，最多重试3次，每次间隔3秒，如果遇到429状态码也会重试
	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)

	// 如果请求过程中出现错误，则记录错误并返回
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_kube_cluster failed: %v", errs[0])
		return errs[0]
	}

	// 如果响应结果指示操作失败，则记录错误信息并返回错误
	if !respData.Result {
		blog.Errorf("call api batch_delete_kube_cluster failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	// 如果操作成功，则记录成功的日志信息
	blog.Infof("call api batch_delete_kube_cluster with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	// 如果提供了数据库连接，则尝试从数据库中删除对应的集群记录
	if db != nil {
		// 使用 GORM 删除数据库中的集群记录
		if err := db.Where("id in (?)", request.IDs).Delete(&model.Cluster{}).Error; err != nil {
			// 如果删除操作失败，则记录错误信息
			blog.Errorf("delete cluster failed: %v", err)
		} else {
			// 如果删除成功，则记录成功的日志信息
			blog.Infof("DeleteBcsCluster clusterWithDB delete")
		}
	}

	// 返回 nil 表示操作成功
	return nil
}

// GetBcsNamespace get bcs namespace
// /api/v3/kube/findmany/namespace/bk_biz_id/{bk_biz_id}
// /v2/cc/list_namespace/
func (c *cmdbClient) GetBcsNamespace(
	request *client.GetBcsNamespaceRequest, db *gorm.DB, withDB bool) (*[]bkcmdbkube.Namespace, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}
	if withDB && db != nil {
		query := db.Session(&gorm.Session{NewDB: true, Logger: db.Logger.LogMode(logger.Info)})
		for _, rule := range request.Filter.Rules {
			if request.Filter.Condition == And {
				query = query.Where(fmt.Sprintf("%s %s ?", rule.Field, rule.Operator), rule.Value)
			}
			if request.Filter.Condition == OR {
				query = query.Or(query.Where(fmt.Sprintf("%s %s ?", rule.Field, rule.Operator), rule.Value))
			}
		}
		var ns []model.Namespace
		if err := query.Debug().Find(&ns).Error; err != nil {
			blog.Errorf("query namespace withDB failed: %v", err)
		} else {
			if namespaceMarshal, err := json.Marshal(ns); err != nil {
				blog.Errorf("marshal namespace failed: %v", err)
			} else {
				var bkNamespace []bkcmdbkube.Namespace
				errM := json.Unmarshal(namespaceMarshal, &bkNamespace)
				if errM == nil {
					blog.Infof("GetBcsNamespace namespaceWithDB get: %d", len(bkNamespace))
					return &bkNamespace, nil
				}
				blog.Errorf("unmarshal namespace failed: %v", errM)
			}
		}
	}
	reqURL := fmt.Sprintf("%s/api/v3/findmany/kube/namespace", c.config.Server)
	respData := &client.GetBcsNamespaceResponse{}
	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_namespace failed: %v", errs[0])
		return nil, errs[0]
	}
	if !respData.Result {
		blog.Errorf("call api list_namespace failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}
	blog.Infof("call api list_namespace with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))
	if db != nil {
		if namespaceMarshal, err := json.Marshal(respData.Data.Info); err != nil {
			blog.Errorf("marshal namespace failed: %v", err)
		} else {
			var ns []model.Namespace
			if errM := json.Unmarshal(namespaceMarshal, &ns); errM != nil {
				blog.Errorf("unmarshal namespace failed: %v", errM)
			} else {
				for _, n := range ns {
					var existingNamespace model.Namespace
					query := db.Session(&gorm.Session{NewDB: true})
					if errN := query.Where("id = ?", n.ID).First(&existingNamespace).Error; errN != nil {
						if errors.Is(errN, gorm.ErrRecordNotFound) {
							if errNC := query.Create(&n).Error; errNC != nil {
								blog.Errorf("create namespace failed: %v", errNC)
							}
							blog.Infof("GetBcsNamespace nodeWithDB create: %s.%s", n.ClusterUID, n.Name)
						} else {
							blog.Errorf("query namespace failed: %v", errN)
						}
					} else {
						if errNS := query.Save(&n).Error; errNS != nil {
							blog.Errorf("update namespace failed: %v", errNS)
						} else {
							blog.Infof("GetBcsNamespace nodeWithDB update: %s.%s", n.ClusterUID, n.Name)
						}
					}
				}
			}
		}
	}
	return respData.Data.Info, nil
}

// CreateBcsNamespace create bcs namespace
// /api/v3/kube/createmany/namespace/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_create_namespace/
func (c *cmdbClient) CreateBcsNamespace(request *client.CreateBcsNamespaceRequest, db *gorm.DB) (*[]int64, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/v3/createmany/kube/namespace", c.config.Server)
	respData := &client.CreateBcsNamespaceResponse{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api create_kube_namespace failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api create_kube_namespace failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		if respData.Code == 1199014 && db != nil {
			for _, d := range *(request.Data) {
				if _, err := c.GetBcsNamespace(&client.GetBcsNamespaceRequest{
					CommonRequest: client.CommonRequest{
						BKBizID: *request.BKBizID,
						Page: client.Page{
							Limit: 100,
							Start: 0,
						},
						Filter: &client.PropertyFilter{
							Condition: And,
							Rules: []client.Rule{
								{
									Field:    "name",
									Operator: "in",
									Value:    []string{d.Name},
								},
								{
									Field:    "bk_cluster_id",
									Operator: "in",
									Value:    []int64{d.ClusterID},
								},
							},
						},
					},
				}, db, false); err != nil {
					blog.Errorf("GetBcsNamespace failed: %v", err)
				}
			}
		}
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api create_kube_namespace with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	if _, err := c.GetBcsNamespace(&client.GetBcsNamespaceRequest{
		CommonRequest: client.CommonRequest{
			BKBizID: *request.BKBizID,
			Page: client.Page{
				Limit: 100,
				Start: 0,
			},
			Filter: &client.PropertyFilter{
				Condition: And,
				Rules: []client.Rule{
					{
						Field:    "id",
						Operator: "in",
						Value:    respData.Data.IDs,
					},
				},
			},
		},
	}, db, false); err != nil {
		blog.Errorf("GetBcsNamespace failed: %v", err)
	}

	return &respData.Data.IDs, nil
}

// UpdateBcsNamespace update bcs namespace
// /api/v3/kube/updatemany/namespace/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_update_namespace/
func (c *cmdbClient) UpdateBcsNamespace(request *client.UpdateBcsNamespaceRequest, db *gorm.DB) error {
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/v3/updatemany/kube/namespace", c.config.Server)
	respData := &client.UpdateBcsNamespaceResponse{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Put(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_update_namespace failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_update_namespace failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_update_namespace with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	if _, err := c.GetBcsNamespace(&client.GetBcsNamespaceRequest{
		CommonRequest: client.CommonRequest{
			BKBizID: *request.BKBizID,
			Page: client.Page{
				Limit: 100,
				Start: 0,
			},
			Filter: &client.PropertyFilter{
				Condition: And,
				Rules: []client.Rule{
					{
						Field:    "id",
						Operator: "in",
						Value:    *request.IDs,
					},
				},
			},
		},
	}, db, false); err != nil {
		blog.Errorf("GetBcsNamespace failed: %v", err)
	}

	return nil
}

// DeleteBcsNamespace delete bcs namespace
// /api/v3/kube/deletemany/namespace/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_delete_namespace/
func (c *cmdbClient) DeleteBcsNamespace(request *client.DeleteBcsNamespaceRequest, db *gorm.DB) error {
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/v3/deletemany/kube/namespace", c.config.Server)
	respData := &client.DeleteBcsNamespaceResponse{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_namespace failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_delete_namespace failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_namespace with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	if db != nil {
		query := db.Session(&gorm.Session{NewDB: true})
		if err := query.Where("id in (?)", *request.IDs).Delete(&model.Namespace{}).Error; err != nil {
			blog.Errorf("delete bcs namespace failed: %v", err)
		} else {
			blog.Infof("DeleteBcsNamespace namespaceWithDB delete: %v", *request.IDs)
		}
	}

	return nil
}

// GetBcsWorkload get bcs workload
// /api/v3/kube/findmany/workload/{kind}/{bk_biz_id}
// /v2/cc/list_workload/
// nolint funlen
func (c *cmdbClient) GetBcsWorkload(
	request *client.GetBcsWorkloadRequest, db *gorm.DB, withDB bool) (*[]interface{}, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	if withDB && db != nil {
		query := db.Session(&gorm.Session{NewDB: true, Logger: db.Logger.LogMode(logger.Info)})
		for _, rule := range request.Filter.Rules {
			if request.Filter.Condition == And {
				query = query.Where(fmt.Sprintf("%s %s ?", rule.Field, rule.Operator), rule.Value)
			}
			if request.Filter.Condition == OR {
				query = query.Or(query.Where(fmt.Sprintf("%s %s ?", rule.Field, rule.Operator), rule.Value))
			}
		}

		switch request.Kind {
		case Deployment:
			var deployment []model.Deployment
			if err := query.Debug().Find(&deployment).Error; err != nil {
				blog.Errorf("query deployment withDB failed: %v", err)
			} else {
				if deploymentMarshal, errM := json.Marshal(deployment); errM != nil {
					blog.Errorf("marshal deployment failed: %v", errM)
				} else {
					var bkDeployment []interface{}
					errUM := json.Unmarshal(deploymentMarshal, &bkDeployment)
					if errUM == nil {
						blog.Infof("GetBcsWorkload deploymentWithDB get: %d", len(bkDeployment))
						return &bkDeployment, nil
					}
					blog.Errorf("unmarshal deployment failed: %v", errUM)
				}
			}
		case StatefulSet:
			var statefulSet []model.StatefulSet
			if err := query.Debug().Find(&statefulSet).Error; err != nil {
				blog.Errorf("query statefulSet withDB failed: %v", err)
			} else {
				if statefulSetMarshal, errM := json.Marshal(statefulSet); errM != nil {
					blog.Errorf("marshal statefulSet failed: %v", errM)
				} else {
					var bkStatefulSet []interface{}
					errUM := json.Unmarshal(statefulSetMarshal, &bkStatefulSet)
					if errUM == nil {
						blog.Infof("GetBcsWorkload statefulSetWithDB get: %d", len(bkStatefulSet))
						return &bkStatefulSet, nil
					}
					blog.Errorf("unmarshal statefulSet failed: %v", errUM)
				}
			}
		case DaemonSet:
			var daemonSet []model.DaemonSet
			if err := query.Debug().Find(&daemonSet).Error; err != nil {
				blog.Errorf("query daemonSet withDB failed: %v", err)
			} else {
				if daemonSetMarshal, errM := json.Marshal(daemonSet); errM != nil {
					blog.Errorf("marshal daemonSet failed: %v", errM)
				} else {
					var bkDaemonSet []interface{}
					errUM := json.Unmarshal(daemonSetMarshal, &bkDaemonSet)
					if errUM == nil {
						blog.Infof("GetBcsWorkload daemonSetWithDB get: %d", len(bkDaemonSet))
						return &bkDaemonSet, nil
					}
					blog.Errorf("unmarshal daemonSet failed: %v", errUM)
				}
			}
		case GameDeployment:
			var gameDeployment []model.GameDeployment
			if err := query.Debug().Find(&gameDeployment).Error; err != nil {
				blog.Errorf("query gameDeployment withDB failed: %v", err)
			} else {
				if gameDeploymentMarshal, errM := json.Marshal(gameDeployment); errM != nil {
					blog.Errorf("marshal gameDeployment failed: %v", errM)
				} else {
					var bkGameDeployment []interface{}
					errUM := json.Unmarshal(gameDeploymentMarshal, &bkGameDeployment)
					if errUM == nil {
						blog.Infof("GetBcsWorkload gameDeploymentWithDB get: %d", len(bkGameDeployment))
						return &bkGameDeployment, nil
					}
					blog.Errorf("unmarshal gameDeployment failed: %v", errUM)
				}
			}
		case GameStatefulSet:
			var gameStatefulSet []model.GameStatefulSet
			if err := query.Debug().Find(&gameStatefulSet).Error; err != nil {
				blog.Errorf("query gameStatefulSet withDB failed: %v", err)
			} else {
				if gameStatefulSetMarshal, errM := json.Marshal(gameStatefulSet); errM != nil {
					blog.Errorf("marshal gameStatefulSet failed: %v", errM)
				} else {
					var bkGameStatefulSet []interface{}
					errUM := json.Unmarshal(gameStatefulSetMarshal, &bkGameStatefulSet)
					if errUM == nil {
						blog.Infof("GetBcsWorkload gameStatefulSetWithDB get: %d", len(bkGameStatefulSet))
						return &bkGameStatefulSet, nil
					}
					blog.Errorf("unmarshal gameStatefulSet failed: %v", errUM)
				}
			}
		case Pods:
			var podsWorkload []model.PodsWorkload
			if err := query.Debug().Find(&podsWorkload).Error; err != nil {
				blog.Errorf("query podsWorkload withDB failed: %v", err)
			} else {
				if podsWorkloadMarshal, errM := json.Marshal(podsWorkload); errM != nil {
					blog.Errorf("marshal podsWorkload failed: %v", errM)
				} else {
					var bkPodsWorkload []interface{}
					errUM := json.Unmarshal(podsWorkloadMarshal, &bkPodsWorkload)
					if errUM == nil {
						blog.Infof("GetBcsWorkload podsWorkloadWithDB get: %d", len(bkPodsWorkload))
						return &bkPodsWorkload, nil
					}
					blog.Errorf("unmarshal podsWorkload failed: %v", errUM)
				}
			}
		}
	}

	reqURL := fmt.Sprintf("%s/api/v3/findmany/kube/workload/%s", c.config.Server, request.Kind)
	respData := &client.GetBcsWorkloadResponse{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_workload failed: %v, rid: %s", errs[0], respData.RequestID)
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_workload failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_workload with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	if db != nil {
		if workloadMarshal, err := json.Marshal(respData.Data.Info); err != nil {
			blog.Errorf("marshal workload failed: %v", err)
		} else {
			switch request.Kind {
			case Deployment:
				var deployment []model.Deployment
				if errM := json.Unmarshal(workloadMarshal, &deployment); errM != nil {
					blog.Errorf("unmarshal deployment failed: %v", errM)
				} else {
					for _, d := range deployment {
						var existingDeploument model.Deployment
						query := db.Session(&gorm.Session{NewDB: true})
						if errD := query.Where("id = ?", d.ID).First(&existingDeploument).Error; errD != nil {
							if errors.Is(errD, gorm.ErrRecordNotFound) {
								if errDC := query.Create(&d).Error; errDC != nil {
									blog.Errorf("create deployment failed: %v", errDC)
								}
								blog.Infof("GetBcsWorkload deploymentWithDB create: %s.%s",
									d.ClusterUID, d.Name)
							} else {
								blog.Errorf("query deployment failed: %v", errD)
							}
						} else {
							if errDS := query.Save(&d).Error; errDS != nil {
								blog.Errorf("update deployment failed: %v", errDS)
							} else {
								blog.Infof("GetBcsWorkload deploymentWithDB update: %s.%s.%s",
									d.ClusterUID, d.Namespace, d.Name)
							}
						}
					}
				}
			case StatefulSet:
				var statefulSet []model.StatefulSet
				if errM := json.Unmarshal(workloadMarshal, &statefulSet); errM != nil {
					blog.Errorf("unmarshal statefulSet failed: %v", err)
				} else {
					for _, s := range statefulSet {
						var existingStatefulSet model.StatefulSet
						query := db.Session(&gorm.Session{NewDB: true})
						if errS := query.Where("id = ?", s.ID).First(&existingStatefulSet).Error; errS != nil {
							if errors.Is(errS, gorm.ErrRecordNotFound) {
								if errSC := query.Create(&s).Error; errSC != nil {
									blog.Errorf("create statefulSet failed: %v", errSC)
								}
								blog.Infof("GetBcsWorkload statefulSetWithDB create: %s.%s",
									s.ClusterUID, s.Name)
							} else {
								blog.Errorf("query statefulSet failed: %v", errS)
							}
						} else {
							if errSS := query.Save(&s).Error; errSS != nil {
								blog.Errorf("update statefulSet failed: %v", errSS)
							} else {
								blog.Infof("GetBcsWorkload statefulSetWithDB update: %s.%s.%s",
									s.ClusterUID, s.Namespace, s.Name)
							}
						}
					}
				}
			case DaemonSet:
				var daemonSet []model.DaemonSet
				if errM := json.Unmarshal(workloadMarshal, &daemonSet); errM != nil {
					blog.Errorf("unmarshal daemonSet failed: %v", errM)
				} else {
					for _, d := range daemonSet {
						var existingDaemonSet model.DaemonSet
						query := db.Session(&gorm.Session{NewDB: true})
						if errD := query.Where("id = ?", d.ID).First(&existingDaemonSet).Error; errD != nil {
							if errors.Is(errD, gorm.ErrRecordNotFound) {
								if errDC := query.Create(&d).Error; errDC != nil {
									blog.Errorf("create daemonSet failed: %v", errDC)
								}
								blog.Infof("GetBcsWorkload daemonSetWithDB create: %s.%s", d.ClusterUID, d.Name)
							} else {
								blog.Errorf("query daemonSet failed: %v", errD)
							}
						} else {
							if errDS := query.Save(&d).Error; errDS != nil {
								blog.Errorf("update daemonSet failed: %v", errDS)
							} else {
								blog.Infof("GetBcsWorkload daemonSetWithDB update: %s.%s.%s",
									d.ClusterUID, d.Namespace, d.Name)
							}
						}
					}
				}
			case GameDeployment:
				var gameDeployment []model.GameDeployment
				if errM := json.Unmarshal(workloadMarshal, &gameDeployment); errM != nil {
					blog.Errorf("unmarshal gameDeployment failed: %v", errM)
				} else {
					for _, g := range gameDeployment {
						var existingGameDeployment model.GameDeployment
						query := db.Session(&gorm.Session{NewDB: true})
						if errG := query.Where("id = ?", g.ID).First(&existingGameDeployment).Error; errG != nil {
							if errors.Is(errG, gorm.ErrRecordNotFound) {
								if errGC := query.Create(&g).Error; errGC != nil {
									blog.Errorf("create gameDeployment failed: %v", errGC)
								}
								blog.Infof("GetBcsWorkload gameDeploymentWithDB create: %s.%s",
									g.ClusterUID, g.Name)
							} else {
								blog.Errorf("query gameDeployment failed: %v", errG)
							}
						} else {
							if errGS := query.Save(&g).Error; errGS != nil {
								blog.Errorf("update gameDeployment failed: %v", errGS)
							} else {
								blog.Infof("GetBcsWorkload gameDeploymentWithDB update: %s.%s.%s",
									g.ClusterUID, g.Namespace, g.Name)
							}
						}
					}
				}
			case GameStatefulSet:
				var gameStatefulSet []model.GameStatefulSet
				if errM := json.Unmarshal(workloadMarshal, &gameStatefulSet); errM != nil {
					blog.Errorf("unmarshal gameStatefulSet failed: %v", errM)
				} else {
					for _, g := range gameStatefulSet {
						var existingGameStatefulSet model.GameStatefulSet
						query := db.Session(&gorm.Session{NewDB: true})
						if errG := query.Where("id = ?", g.ID).First(&existingGameStatefulSet).Error; errG != nil {
							if errors.Is(errG, gorm.ErrRecordNotFound) {
								if errGC := query.Create(&g).Error; errGC != nil {
									blog.Errorf("create gameStatefulSet failed: %v", errGC)
								}
								blog.Infof("GetBcsWorkload gameStatefulSetWithDB create: %s.%s",
									g.ClusterUID, g.Name)
							} else {
								blog.Errorf("query gameStatefulSet failed: %v", errG)
							}
						} else {
							if errGS := query.Save(&g).Error; errGS != nil {
								blog.Errorf("update gameStatefulSet failed: %v", errGS)
							} else {
								blog.Infof("GetBcsWorkload gameStatefulSetWithDB update: %s.%s.%s",
									g.ClusterUID, g.Namespace, g.Name)
							}
						}
					}
				}
			case Pods:
				var podsWorkload []model.PodsWorkload
				if errM := json.Unmarshal(workloadMarshal, &podsWorkload); errM != nil {
					blog.Errorf("unmarshal podsWordload failed: %v", errM)
				} else {
					for _, p := range podsWorkload {
						var existingPodsWorkload model.PodsWorkload
						query := db.Session(&gorm.Session{NewDB: true})
						if errP := query.Where("id = ?", p.ID).First(&existingPodsWorkload).Error; errP != nil {
							if errors.Is(errP, gorm.ErrRecordNotFound) {
								if errPC := query.Create(&p).Error; errPC != nil {
									blog.Errorf("create podsWordload failed: %v", errPC)
								}
								blog.Infof("GetBcsWorkload podsWordloadWithDB create: %s.%s",
									p.ClusterUID, p.Name)
							} else {
								blog.Errorf("query podsWordload failed: %v", errP)
							}
						} else {
							if errPS := query.Save(&p).Error; errPS != nil {
								blog.Errorf("update podsWordload failed: %v", errPS)
							} else {
								blog.Infof("GetBcsWorkload podsWordloadWithDB update: %s.%s",
									p.ClusterUID, p.Name)
							}
						}
					}
				}
			}
		}
	}

	return &respData.Data.Info, nil
}

// CreateBcsWorkload create bcs workload
// /api/v3/kube/createmany/workload/{kind}/{bk_biz_id}
// /v2/cc/batch_create_workload/
func (c *cmdbClient) CreateBcsWorkload(request *client.CreateBcsWorkloadRequest, db *gorm.DB) (*[]int64, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/v3/createmany/kube/workload/%s", c.config.Server, *request.Kind)
	respData := &client.CreateBcsWorkloadResponse{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_create_workload failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_create_workload failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		if respData.Code == 1199014 && db != nil {
			for _, d := range *(request.Data) {
				if _, err := c.GetBcsWorkload(&client.GetBcsWorkloadRequest{
					Kind: *(request.Kind),
					CommonRequest: client.CommonRequest{
						BKBizID: *request.BKBizID,
						Page: client.Page{
							Limit: 100,
							Start: 0,
						},
						Filter: &client.PropertyFilter{
							Condition: And,
							Rules: []client.Rule{
								{
									Field:    "name",
									Operator: "in",
									Value:    []string{*d.Name},
								},
								{
									Field:    "bk_namespace_id",
									Operator: "in",
									Value:    []int64{*d.NamespaceID},
								},
							},
						},
					},
				}, db, false); err != nil {
					blog.Errorf("get bcs workload failed: %v", err)
				}
			}
		}
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_create_workload with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	if _, err := c.GetBcsWorkload(&client.GetBcsWorkloadRequest{
		Kind: *(request.Kind),
		CommonRequest: client.CommonRequest{
			BKBizID: *request.BKBizID,
			Page: client.Page{
				Limit: 100,
				Start: 0,
			},
			Filter: &client.PropertyFilter{
				Condition: And,
				Rules: []client.Rule{
					{
						Field:    "id",
						Operator: "in",
						Value:    respData.Data.IDs,
					},
				},
			},
		},
	}, db, false); err != nil {
		blog.Errorf("get bcs workload failed: %v", err)
	}

	return &respData.Data.IDs, nil
}

// UpdateBcsWorkload update bcs workload
// /api/v3/kube/updatemany/workload/{kind}/{bk_biz_id}
// /v2/cc/batch_update_workload/
func (c *cmdbClient) UpdateBcsWorkload(request *client.UpdateBcsWorkloadRequest, db *gorm.DB) error {
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/v3/updatemany/kube/workload/%s", c.config.Server, *request.Kind)
	respData := &client.UpdateBcsWorkloadResponse{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Put(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_update_workload failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_update_workload failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_update_workload with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	if _, err := c.GetBcsWorkload(&client.GetBcsWorkloadRequest{
		Kind: *request.Kind,
		CommonRequest: client.CommonRequest{
			BKBizID: *request.BKBizID,
			Page: client.Page{
				Limit: 100,
				Start: 0,
			},
			Filter: &client.PropertyFilter{
				Condition: And,
				Rules: []client.Rule{
					{
						Field:    "id",
						Operator: "in",
						Value:    *request.IDs,
					},
				},
			},
		},
	}, db, false); err != nil {
		blog.Errorf("get bcs workload failed: %v", err)
	}

	return nil
}

// DeleteBcsWorkload delete bcs workload
// /api/v3/kube/deletemany/workload/{kind}/{bk_biz_id}
// /v2/cc/batch_delete_workload/
func (c *cmdbClient) DeleteBcsWorkload(request *client.DeleteBcsWorkloadRequest, db *gorm.DB) error { // nolint: cyclop
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/v3/deletemany/kube/workload/%s", c.config.Server, *request.Kind)
	respData := &client.DeleteBcsWorkloadResponse{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_workload failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf(
			"call api batch_delete_workload failed: %v, rid: %s, request: bkbizid: %d, kind: %s, ids: %v",
			respData.Message, resp.Header.Get("X-Request-Id"), *request.BKBizID, *request.Kind, *request.IDs)
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_workload with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	deleteBcsWorkloadDB(request, db)

	return nil
}

func deleteBcsWorkloadDB(request *client.DeleteBcsWorkloadRequest, db *gorm.DB) {
	switch *request.Kind {
	case Deployment:
		if db != nil {
			query := db.Session(&gorm.Session{NewDB: true})
			if err := query.Where("id in (?)", *request.IDs).Delete(&model.Deployment{}).Error; err != nil {
				blog.Errorf("delete bcs deployment failed: %v", err)
			} else {
				blog.Infof("DeleteBcsWorkload deploymentWithDB delete: %v", *request.IDs)
			}
		}
	case StatefulSet:
		if db != nil {
			query := db.Session(&gorm.Session{NewDB: true})
			if err := query.Where("id in (?)", *request.IDs).Delete(&model.StatefulSet{}).Error; err != nil {
				blog.Errorf("delete bcs statefulset failed: %v", err)
			} else {
				blog.Infof("DeleteBcsWorkload statefulSetWithDB delete: %v", *request.IDs)
			}
		}
	case DaemonSet:
		if db != nil {
			query := db.Session(&gorm.Session{NewDB: true})
			if err := query.Where("id in (?)", *request.IDs).Delete(&model.DaemonSet{}).Error; err != nil {
				blog.Errorf("delete bcs daemonSet failed: %v", err)
			} else {
				blog.Infof("DeleteBcsWorkload daemonSetWithDB delete: %v", *request.IDs)
			}
		}
	case GameDeployment:
		if db != nil {
			query := db.Session(&gorm.Session{NewDB: true})
			if err := query.Where("id in (?)", *request.IDs).Delete(&model.GameDeployment{}).Error; err != nil {
				blog.Errorf("delete bcs gameDeployment failed: %v", err)
			} else {
				blog.Infof("DeleteBcsWorkload gameDeploymentWithDB delete: %v", *request.IDs)
			}
		}
	case GameStatefulSet:
		if db != nil {
			query := db.Session(&gorm.Session{NewDB: true})
			if err := query.Where("id in (?)", *request.IDs).Delete(&model.GameStatefulSet{}).Error; err != nil {
				blog.Errorf("delete bcs gameStatefulSet failed: %v", err)
			} else {
				blog.Infof("DeleteBcsWorkload gameStatefulSetWithDB delete: %v", *request.IDs)
			}
		}
	case Pods:
		if db != nil {
			query := db.Session(&gorm.Session{NewDB: true})
			if err := query.Where("id in (?)", *request.IDs).Delete(&model.PodsWorkload{}).Error; err != nil {
				blog.Errorf("delete bcs podsWordload failed: %v", err)
			} else {
				blog.Infof("DeleteBcsWorkload podsWordloadWithDB delete: %v", *request.IDs)
			}
		}
	}
}

// GetBcsNode get bcs node
// /api/v3/kube/findmany/node/bk_biz_id/{bk_biz_id}
// /v2/cc/list_kube_node/
func (c *cmdbClient) GetBcsNode(
	request *client.GetBcsNodeRequest, db *gorm.DB, withDB bool) (*[]bkcmdbkube.Node, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}
	if withDB && db != nil {
		query := db.Session(&gorm.Session{NewDB: true, Logger: db.Logger.LogMode(logger.Info)})
		for _, rule := range request.Filter.Rules {
			if request.Filter.Condition == And {
				query = query.Where(fmt.Sprintf("%s %s ?", rule.Field, rule.Operator), rule.Value)
			}
			if request.Filter.Condition == OR {
				query = query.Or(query.Where(fmt.Sprintf("%s %s ?", rule.Field, rule.Operator), rule.Value))
			}
		}
		var node []model.Node
		if err := query.Debug().Find(&node).Error; err != nil {
			blog.Errorf("query node withDB failed: %v", err)
		} else {
			if nodeMarshal, errM := json.Marshal(node); errM != nil {
				blog.Errorf("marshal node failed: %v", errM)
			} else {
				var bkNode []bkcmdbkube.Node
				errUM := json.Unmarshal(nodeMarshal, &bkNode)
				if errUM == nil {
					blog.Infof("GetBcsNode nodeWithDB get: %d", len(bkNode))
					return &bkNode, nil
				}
				blog.Errorf("unmarshal node failed: %v", errUM)
			}
		}
	}
	reqURL := fmt.Sprintf("%s/api/v3/findmany/kube/node", c.config.Server)
	respData := &client.GetBcsNodeResponse{}
	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_kube_node failed: %v", errs[0])
		return nil, errs[0]
	}
	if !respData.Result {
		blog.Errorf("call api list_kube_node failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}
	blog.Infof("call api list_kube_node with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))
	if db != nil {
		if nodeMarshal, err := json.Marshal(respData.Data.Info); err != nil {
			blog.Errorf("marshal node failed: %v", err)
		} else {
			var node []model.Node
			if errM := json.Unmarshal(nodeMarshal, &node); errM != nil {
				blog.Errorf("unmarshal node failed: %v", errM)
			} else {
				for _, n := range node {
					var existingNode model.Node
					query := db.Session(&gorm.Session{NewDB: true})
					if errN := query.Where("id = ?", n.ID).First(&existingNode).Error; errN != nil {
						if errors.Is(errN, gorm.ErrRecordNotFound) {
							if errNC := query.Create(&n).Error; errNC != nil {
								blog.Errorf("create node failed: %v", errNC)
							}
							blog.Infof("GetBcsNode nodeWithDB create: %s.%s", n.ClusterUID, n.Name)
						} else {
							blog.Errorf("query node failed: %v", errN)
						}
					} else {
						if errNS := query.Save(&n).Error; errNS != nil {
							blog.Errorf("update node failed: %v", errNS)
						} else {
							blog.Infof("GetBcsNode nodeWithDB update: %s.%s", n.ClusterUID, n.Name)
						}
					}
				}
			}
		}
	}
	return respData.Data.Info, nil
}

// CreateBcsNode create bcs node
// /api/v3/kube/createmany/node/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_create_kube_node/
func (c *cmdbClient) CreateBcsNode(request *client.CreateBcsNodeRequest, db *gorm.DB) (*[]int64, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}
	blog.Infof("request: biz %d, data %v", *request.BKBizID, *request.Data)
	reqURL := fmt.Sprintf("%s/api/v3/createmany/kube/node", c.config.Server)
	respData := &client.CreateBcsNodeResponse{}
	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_create_kube_node failed: %v", errs[0])
		return nil, errs[0]
	}
	if !respData.Result {
		blog.Errorf(
			"call api batch_create_kube_node failed: %v, bk_biz_id: %d, bk_host_id: %d, host_ip: %s, rid: %s",
			respData.Message, *request.BKBizID, *(*request.Data)[0].HostID,
			*(*request.Data)[0].InternalIP, resp.Header.Get("X-Request-Id"))
		if respData.Code == 1199014 && db != nil {
			for _, d := range *(request.Data) {
				_, err := c.GetBcsNode(&client.GetBcsNodeRequest{
					CommonRequest: client.CommonRequest{
						BKBizID: *request.BKBizID,
						Page: client.Page{
							Limit: 100,
							Start: 0,
						},
						Filter: &client.PropertyFilter{
							Condition: And,
							Rules: []client.Rule{
								{
									Field:    "name",
									Operator: "in",
									Value:    []string{*d.Name},
								},
								{
									Field:    "bk_cluster_id",
									Operator: "in",
									Value:    []int64{*d.ClusterID},
								},
							},
						},
					},
				}, db, false)

				if err != nil {
					blog.Errorf("get bcs node failed, err: %s", err.Error())
				}
			}
		}
		return nil, fmt.Errorf(respData.Message)
	}
	blog.Infof("call api batch_create_kube_node with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))
	if db != nil {
		_, err := c.GetBcsNode(&client.GetBcsNodeRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: *request.BKBizID,
				Page: client.Page{
					Limit: 100,
					Start: 0,
				},
				Filter: &client.PropertyFilter{
					Condition: And,
					Rules: []client.Rule{
						{
							Field:    "id",
							Operator: "in",
							Value:    respData.Data.IDs,
						},
					},
				},
			},
		}, db, false)

		if err != nil {
			blog.Errorf("get bcs node failed, err: %s", err.Error())
		}
	}
	return &respData.Data.IDs, nil
}

// UpdateBcsNode update bcs node
// /api/v3/kube/updatemany/node/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_update_kube_node/
func (c *cmdbClient) UpdateBcsNode(request *client.UpdateBcsNodeRequest, db *gorm.DB) error {
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/v3/updatemany/kube/node", c.config.Server)
	respData := &client.UpdateBcsNodeResponse{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Put(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_update_kube_node failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_update_kube_node failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_update_kube_node with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	_, err := c.GetBcsNode(&client.GetBcsNodeRequest{
		CommonRequest: client.CommonRequest{
			BKBizID: *request.BKBizID,
			Page: client.Page{
				Limit: 100,
				Start: 0,
			},
			Filter: &client.PropertyFilter{
				Condition: And,
				Rules: []client.Rule{
					{
						Field:    "id",
						Operator: "in",
						Value:    *request.IDs,
					},
				},
			},
		},
	}, db, false)

	if err != nil {
		blog.Errorf("get bcs node failed, err: %s", err.Error())
	}

	return nil
}

// DeleteBcsNode delete bcs node
// /api/v3/kube/deletemany/node/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_delete_kube_node/
func (c *cmdbClient) DeleteBcsNode(request *client.DeleteBcsNodeRequest, db *gorm.DB) error {
	if c == nil {
		return ErrServerNotInit
	}

	if db != nil {
		query := db.Session(&gorm.Session{NewDB: true})
		if err := query.Where("id in (?)", *request.IDs).Delete(&model.Node{}).Error; err != nil {
			blog.Errorf("delete bcs node failed: %v", err)
		} else {
			blog.Infof("DeleteBcsNode nodeWithDB delete: %v", *request.IDs)
		}
	}

	reqURL := fmt.Sprintf("%s/api/v3/deletemany/kube/node", c.config.Server)
	respData := &client.DeleteBcsNodeResponse{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_kube_node failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_delete_kube_node failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_kube_node with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	return nil
}

// GetBcsPod get bcs pod
// /api/v3/kube/findmany/pod/bk_biz_id/{bk_biz_id}
// /v2/cc/list_pod/
func (c *cmdbClient) GetBcsPod(request *client.GetBcsPodRequest, db *gorm.DB, withDB bool) (*[]bkcmdbkube.Pod, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	if withDB && db != nil {
		query := db.Session(&gorm.Session{NewDB: true, Logger: db.Logger.LogMode(logger.Info)})
		for _, rule := range request.Filter.Rules {
			if request.Filter.Condition == And {
				query = query.Where(fmt.Sprintf("%s %s ?", rule.Field, rule.Operator), rule.Value)
			}
			if request.Filter.Condition == OR {
				query = query.Or(query.Where(fmt.Sprintf("%s %s ?", rule.Field, rule.Operator), rule.Value))
			}
		}

		var pod []model.Pod

		if err := query.Debug().Find(&pod).Error; err != nil {
			blog.Errorf("query pod withDB failed: %v", err)
		} else {
			if podMarshal, errM := json.Marshal(pod); errM != nil {
				blog.Errorf("marshal pod failed: %v", errM)
			} else {
				var bkPod []bkcmdbkube.Pod
				errUM := json.Unmarshal(podMarshal, &bkPod)
				if errUM == nil {
					blog.Infof("GetBcsPod podWithDB get: %d", len(bkPod))
					return &bkPod, nil
				}
				blog.Errorf("unmarshal pod failed: %v", errUM)
			}
		}
	}

	reqURL := fmt.Sprintf("%s/api/v3/findmany/kube/pod", c.config.Server)
	respData := &client.GetBcsPodResponse{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_pod failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_pod failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_pod with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	if db != nil {
		if podMarshal, err := json.Marshal(respData.Data.Info); err != nil {
			blog.Errorf("marshal pod failed: %v", err)
		} else {
			var pod []model.Pod
			if errM := json.Unmarshal(podMarshal, &pod); errM != nil {
				blog.Errorf("unmarshal pod failed: %v", errM)
			} else {
				for _, p := range pod {
					var existingPod model.Pod
					query := db.Session(&gorm.Session{NewDB: true})
					if errP := query.Where("id = ?", p.ID).First(&existingPod).Error; errP != nil {
						if errors.Is(errP, gorm.ErrRecordNotFound) {
							if errPC := query.Create(&p).Error; errPC != nil {
								blog.Errorf("create pod failed: %v", errPC)
							}
							blog.Infof("GetBcsPod podWithDB create: %s.%s.%s",
								p.ClusterUID, p.NameSpace, p.Name)
						} else {
							blog.Errorf("query pod failed: %v", errP)
						}
					}
				}
			}
		}
	}

	return respData.Data.Info, nil
}

// GetBcsContainer get bcs container
// /v2/cc/list_kube_container/
func (c *cmdbClient) GetBcsContainer(
	request *client.GetBcsContainerRequest, db *gorm.DB, withDB bool) (*[]client.Container, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	if withDB && db != nil {
		query := db.Session(&gorm.Session{NewDB: true, Logger: db.Logger.LogMode(logger.Info)})
		for _, rule := range request.Filter.Rules {
			if request.Filter.Condition == And {
				query = query.Where(fmt.Sprintf("%s %s ?", rule.Field, rule.Operator), rule.Value)
			}
			if request.Filter.Condition == OR {
				query = query.Or(query.Where(fmt.Sprintf("%s %s ?", rule.Field, rule.Operator), rule.Value))
			}
		}

		var container []model.Container

		if err := query.Debug().Find(&container).Error; err != nil {
			blog.Errorf("query container withDB failed: %v", err)
		} else {
			if containerMarshal, errM := json.Marshal(container); errM != nil {
				blog.Errorf("marshal container failed: %v", errM)
			} else {
				var bkContainer []client.Container
				errUM := json.Unmarshal(containerMarshal, &bkContainer)
				if errUM == nil {
					blog.Infof("GetBcsContainer containerWithDB get: %d", len(bkContainer))
					return &bkContainer, nil
				}
				blog.Errorf("unmarshal container failed: %v", errUM)
			}
		}
	}

	reqURL := fmt.Sprintf("%s/api/v3/findmany/kube/container", c.config.Server)
	respData := &client.GetBcsContainerResponse{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_kube_container failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_kube_container failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_kube_container with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	if db != nil {
		if containerMarshal, err := json.Marshal(respData.Data.Info); err != nil {
			blog.Errorf("marshal container failed: %v", err)
		} else {
			var container []model.Container
			if errM := json.Unmarshal(containerMarshal, &container); errM != nil {
				blog.Errorf("unmarshal container failed: %v", errM)
			} else {
				for _, ct := range container {
					var existingContainer model.Container
					query := db.Session(&gorm.Session{NewDB: true})
					if errC := query.Where("id = ?", ct.ID).First(&existingContainer).Error; errC != nil {
						if errors.Is(errC, gorm.ErrRecordNotFound) {
							if errCC := query.Create(&ct).Error; errCC != nil {
								blog.Errorf("create container failed: %v", errCC)
							}
							blog.Infof("GetBcsContainer containerWithDB create: %d.%d.%s",
								ct.ClusterID, ct.PodID, ct.Name)
						} else {
							blog.Errorf("query container failed: %v", errC)
						}
					}
				}
			}
		}
	}

	return respData.Data.Info, nil
}

// CreateBcsPod create bcs pod
// /api/v3/kube/createmany/pod/
// /v2/cc/batch_create_kube_pod/
// nolint funlen
func (c *cmdbClient) CreateBcsPod(request *client.CreateBcsPodRequest, db *gorm.DB) (*[]int64, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/v3/createmany/kube/pod", c.config.Server)
	respData := &client.CreateBcsPodResponse{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_create_kube_pod failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_create_kube_pod failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		if respData.Code == 1199014 && db != nil {
			data := (*(request.Data))[0]
			for _, d := range *data.Pods {
				_, err := c.GetBcsPod(&client.GetBcsPodRequest{
					CommonRequest: client.CommonRequest{
						BKBizID: *data.BizID,
						Page: client.Page{
							Limit: 100,
							Start: 0,
						},
						Filter: &client.PropertyFilter{
							Condition: And,
							Rules: []client.Rule{
								{
									Field:    "name",
									Operator: "in",
									Value:    []string{*d.Name},
								},
								{
									Field:    "bk_cluster_id",
									Operator: "in",
									Value:    []int64{*d.Spec.ClusterID},
								},
								{
									Field:    "bk_namespace_id",
									Operator: "in",
									Value:    []int64{*d.Spec.NameSpaceID},
								},
							},
						},
					},
				}, db, false)
				if err != nil {
					blog.Errorf("get bcs pod failed, err: %s", err.Error())
				}
			}
		}
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_create_kube_pod with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	if db != nil {
		_, err := c.GetBcsPod(&client.GetBcsPodRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: *(*request.Data)[0].BizID,
				Page: client.Page{
					Limit: 100,
					Start: 0,
				},
				Filter: &client.PropertyFilter{
					Condition: And,
					Rules: []client.Rule{
						{
							Field:    "id",
							Operator: "in",
							Value:    respData.Data.IDs,
						},
					},
				},
			},
		}, db, false)

		if err != nil {
			blog.Errorf("get bcs pod failed, err: %s", err.Error())
		}

		_, err = c.GetBcsContainer(&client.GetBcsContainerRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: *(*request.Data)[0].BizID,
				Page: client.Page{
					Limit: 100,
					Start: 0,
				},
				Filter: &client.PropertyFilter{
					Condition: And,
					Rules: []client.Rule{
						{
							Field:    "bk_pod_id",
							Operator: "in",
							Value:    respData.Data.IDs,
						},
					},
				},
			},
		}, db, false)

		if err != nil {
			blog.Errorf("get bcs container failed, err: %s", err.Error())
		}
	}

	return &respData.Data.IDs, nil
}

// DeleteBcsPod delete bcs pod
// /api/v3/kube/deletemany/pod/
// /v2/cc/batch_delete_kube_pod/
func (c *cmdbClient) DeleteBcsPod(request *client.DeleteBcsPodRequest, db *gorm.DB) error {
	if c == nil {
		return ErrServerNotInit
	}

	reqURL := fmt.Sprintf("%s/api/v3/deletemany/kube/pod", c.config.Server)
	respData := &client.DeleteBcsPodResponse{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(request).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_kube_pod failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_delete_kube_pod failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_kube_pod with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, request, resp.Header.Get("X-Request-Id"))

	if db != nil {
		query := db.Session(&gorm.Session{NewDB: true})
		if err := query.Where("id in (?)", *(*request.Data)[0].IDs).Delete(&model.Pod{}).Error; err != nil {
			blog.Errorf("delete bcs pod failed: %v", err)
		} else {
			blog.Infof("DeleteBcsPod podWithDB delete: %v", *(*request.Data)[0].IDs)
		}
		if err := query.Where("bk_pod_id in (?)", *(*request.Data)[0].IDs).Delete(&model.Container{}).Error; err != nil {
			blog.Errorf("delete bcs container failed: %v", err)
		} else {
			blog.Infof("DeleteBcsPod containerWithDB delete: %v", *(*request.Data)[0].IDs)
		}
	}

	return nil
}

// GetCMDBClient get cmdb client
func (c *cmdbClient) GetCMDBClient() (client.CMDBClient, error) {
	cli := NewCmdbClient(c.config)
	return cli.Cmdb(), nil
}

// DeleteBcsClusterAll delete bcs cluster
// /api/v3/kube/delete/cluster/bk_biz_id/{bk_biz_id}
// /v2/cc/batch_delete_kube_cluster/
func (c *cmdbClient) DeleteBcsClusterAll(request *client.DeleteBcsClusterAllRequest, db *gorm.DB) error {
	if c == nil {
		return ErrServerNotInit
	}

	return nil
}
