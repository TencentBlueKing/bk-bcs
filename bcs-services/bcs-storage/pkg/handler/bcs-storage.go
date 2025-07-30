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

// Package handler xxx
package handler

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/handler/internal/alarm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/handler/internal/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/handler/internal/clusterconfig"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/handler/internal/dynamic"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/handler/internal/dynamicquery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/handler/internal/dynamicwatch"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/handler/internal/events"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/handler/internal/hostconfig"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/handler/internal/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/handler/internal/metricwatch"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/handler/internal/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/handler/internal/watchk8smesos"
	storage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/util"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	v1httpclu "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/cluster"
	dynamic2 "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/dynamic"
	v1httppro "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/project"
)

// Storage storage struct, which include operations of alarm,
// cluster config, dynamic, dynamic-query, dynamic-watch, events,
// host config, metric, metric watch, watch k8s.
// operations include get, put, list, watch, delete,
// batch get, batch put, batch list...
type Storage struct {
}

// New new storage, return pointer of storage
func New() *Storage {
	return new(Storage)
}

// **** Alarm(告警) ****

// PostAlarm 创建告警 the function of create alarm
func (s *Storage) PostAlarm(ctx context.Context, req *storage.PostAlarmRequest, rsp *storage.PostAlarmResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostAlarm req: %v", util.PrettyStruct(req))

	if err := alarm.HandlerPostAlarm(ctx, req); err != nil {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStoragePutResourceFail
		rsp.Message = common.BcsErrStoragePutResourceFailStr
		blog.Errorf("PostAlarm | post alarm failed.err: %v", err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostAlarm rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// ListAlarm 查询告警 list all alarms
func (s *Storage) ListAlarm(ctx context.Context, req *storage.ListAlarmRequest, rsp *storage.ListAlarmResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("ListAlarm req: %v", util.PrettyStruct(req))

	data, err := alarm.HandlerListAlarm(ctx, req)
	if err != nil {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("ListAlarm | alarm query failed.err: %v", err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "ListAlarm"); err != nil {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("ListAlarm | data to json failed.err: %v", err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr
	rsp.Extra, _ = structpb.NewStruct(map[string]interface{}{"total": len(data)})

	// 打印响应体
	blog.Infof("ListAlarm rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PutProjectInfo 订阅项目数据
func (s *Storage) PutProjectInfo(ctx context.Context, req *storage.PutProjectInfoRequest,
	rsp *storage.PutProjectInfoResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PutProjectInfo req: %v", util.PrettyStruct(req))
	data, err := project.HandlerCreateProjectInfoReq(ctx, req)
	if err != nil {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStoragePutResourceFail
		rsp.Message = common.BcsErrStoragePutResourceFailStr
		blog.Errorf("PutProjectInfo | put project info failed.err: %v", err)
		return nil
	}

	v1httppro.PushCreateProjectInfoToQueue(data)

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PutProject rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PutClusterInfo 订阅集群数据
func (s *Storage) PutClusterInfo(ctx context.Context, req *storage.PutClusterInfoRequest,
	rsp *storage.PutClusterInfoResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PutClusterInfo req: %v", util.PrettyStruct(req))
	data, err := cluster.HandlerCreateClusterInfoReq(ctx, req)
	if err != nil {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStoragePutResourceFail
		rsp.Message = common.BcsErrStoragePutResourceFailStr
		blog.Errorf("PutClusterInfo | put cluster info failed.err: %v", err)
		return nil
	}

	v1httpclu.PushCreateClusterInfoToQueue(data)

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PutProject rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// DeleteClusterInfo 删除集群数据
func (s *Storage) DeleteClusterInfo(ctx context.Context, req *storage.DeleteClusterInfoRequest,
	rsp *storage.PutClusterInfoResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("DeleteClusterInfo req: %v", util.PrettyStruct(req))
	data, err := cluster.HandlerDeleteClusterInfoReq(ctx, req)
	if err != nil {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStorageDeleteResourceFail
		rsp.Message = common.BcsErrStorageDeleteResourceFailStr
		blog.Errorf("DeleteClusterInfo | delete cluster info failed.err: %v", err)
		return nil
	}

	if len(data) == 1 {
		v1httpclu.PushDeleteClusterInfoToQueue(data[0])
	}

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("DeleteClusterInfo rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// ****  Cluster Config(集群配置) ****

// GetClusterConfig 获取集群配置 get config of cluster
func (s *Storage) GetClusterConfig(ctx context.Context, req *storage.GetClusterConfigRequest,
	rsp *storage.GetClusterConfigResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetClusterConfig req: %v", util.PrettyStruct(req))

	data, err := clusterconfig.HandlerGetClusterConfig(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageGetResourceFail
		rsp.Message = common.BcsErrStorageGetResourceFailStr
		blog.Errorf("GetClusterConfig %s | query failed.err: %v", common.BcsErrStorageGetResourceFailStr, err)
		return nil
	}
	if err = util.StructToStruct(data, rsp.Data, "GetClusterConfig"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetClusterConfig %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetClusterConfig rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PutClusterConfig 保存集群配置 put config of cluster
func (s *Storage) PutClusterConfig(ctx context.Context, req *storage.PutClusterConfigRequest,
	rsp *storage.PutClusterConfigResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PutClusterConfig req: %v", util.PrettyStruct(req))
	clusterconfig.HandlerPutClusterConfig(ctx, req, rsp)

	// 打印响应体
	blog.Infof("PutClusterConfig rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// GetServiceConfig 获取集群配置 get config of service
func (s *Storage) GetServiceConfig(ctx context.Context, req *storage.GetServiceConfigRequest,
	rsp *storage.GetServiceConfigResponse) error {
	// 如果请求合法,则执行函数请求
	if (req.ClusterId == "" && req.ClusterIdNot == "") || req.Service == "" {
		errMessage := "ClusterId or Service cannot be nil"
		rsp.Result = false
		rsp.Message = errMessage
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetServiceConfig req: %v", util.PrettyStruct(req))

	data, err := clusterconfig.HandlerGetServiceConfig(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageGetResourceFail
		rsp.Message = common.BcsErrStorageGetResourceFailStr
		blog.Errorf("GetServiceConfig %s | query failed.err: %v", common.BcsErrStorageGetResourceFailStr, err)
		return nil
	}
	if err = util.StructToStruct(data, rsp.Data, "GetServiceConfig"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetServiceConfig %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetServiceConfig rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// GetStableVersion 获取稳定版本 get stable version
func (s *Storage) GetStableVersion(ctx context.Context, req *storage.GetStableVersionRequest,
	rsp *storage.GetStableVersionResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetStableVersion req: %v", util.PrettyStruct(req))

	data, err := clusterconfig.HandlerGetStableVersion(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageGetResourceFail
		rsp.Message = common.BcsErrStorageGetResourceFailStr
		blog.Errorf("GetStableVersion %s | query failed.err: %v", common.BcsErrStorageGetResourceFailStr, err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Data = data
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetStableVersion rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PutStableVersion 创建稳定版本 create stable version
func (s *Storage) PutStableVersion(ctx context.Context, req *storage.PutStableVersionRequest,
	rsp *storage.PutStableVersionResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PutStableVersion req: %v", util.PrettyStruct(req))

	if err := clusterconfig.HandlerPutStableVersion(ctx, req); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStoragePutResourceFail
		rsp.Message = common.BcsErrStoragePutResourceFailStr
		blog.Errorf("PutStableVersion %s | create failed.err: %v", common.BcsErrStoragePutResourceFailStr, err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PutStableVersion rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// **** dynamic(动态) *****
// k8s namespace resources

// GetK8SNamespaceResources 获取k8s命名空间资源 get k8s namespace resources
func (s *Storage) GetK8SNamespaceResources(ctx context.Context, req *storage.GetNamespaceResourcesRequest,
	rsp *storage.GetNamespaceResourcesResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetK8SNamespaceResources req: %v", util.PrettyStruct(req))

	data, err := dynamic.HandlerGetNsResourcesReq(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageGetResourceFail
		rsp.Message = common.BcsErrStorageGetResourceFailStr
		blog.Errorf("GetK8SNamespaceResources %s | query failed.err: %v", common.BcsErrStorageGetResourceFailStr, err)
		return nil
	}
	if len(data) == 0 {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageResourceNotExist
		rsp.Message = common.BcsErrStorageResourceNotExistStr
		blog.Errorf("GetK8SNamespaceResources %s | no date.err: %v", common.BcsErrStorageResourceNotExistStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetK8SNamespaceResources"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetK8SNamespaceResources %s | data to json failed.err: %v",
			common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetK8SNamespaceResources rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PutK8SNamespaceResources 创建k8s命名空间资源 create k8s namespace resources
func (s *Storage) PutK8SNamespaceResources(ctx context.Context, req *storage.PutNamespaceResourcesRequest,
	rsp *storage.PutNamespaceResourcesResponse) error {
	// 如果请求合法,则执行函数请求
	if req.ClusterId == "" || req.Namespace == "" || req.ResourceType == "" || req.ResourceName == "" {
		errMessage := "ClusterId、Namespace、ResourceType or ResourceName cannot be nil"
		rsp.Result = false
		rsp.Message = errMessage
		rsp.Code = common.AdditionErrorCode + 500

		return nil
	}
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PutK8SNamespaceResources req: %v", util.PrettyStruct(req))

	data, err := dynamic.HandlerPutNsResourcesReq(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStoragePutResourceFail
		rsp.Message = common.BcsErrStoragePutResourceFailStr
		blog.Errorf("PutK8SNamespaceResources %s | create failed.err: %v", common.BcsErrStoragePutResourceFailStr, err)
		return nil
	}
	dynamic2.PushCreateResourcesToQueue(data)
	_ = util.MapToStruct(data, rsp.Data, "PutK8SNamespaceResources")

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PutK8SNamespaceResources rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// DeleteK8SNamespaceResources 删除k8s命名空间资源 delete k8s namespace resources
func (s *Storage) DeleteK8SNamespaceResources(ctx context.Context, req *storage.DeleteNamespaceResourcesRequest,
	rsp *storage.DeleteNamespaceResourcesResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("DeleteK8SNamespaceResources req: %v", util.PrettyStruct(req))

	data, err := dynamic.HandlerDelNsResourcesReq(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageDeleteResourceFail
		rsp.Message = common.BcsErrStorageDeleteResourceFailStr
		blog.Errorf("DeleteK8SNamespaceResources %s | delete failed.err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		return nil
	}
	dynamic2.PushDeleteResourcesToQueue(data)
	_ = util.ListMapToListStruct(data, &rsp.Data, "DeleteK8SNamespaceResources")

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("DeleteK8SNamespaceResources rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// ListK8SNamespaceResources 批量查询k8s命名空间资源 batch list k8s namespaces resources
func (s *Storage) ListK8SNamespaceResources(ctx context.Context, req *storage.ListNamespaceResourcesRequest,
	rsp *storage.ListNamespaceResourcesResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("ListK8SNamespaceResources req: %v", util.PrettyStruct(req))

	data, err := dynamic.HandlerListNsResourcesReq(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("ListK8SNamespaceResources %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "ListK8SNamespaceResources"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("ListK8SNamespaceResources %s | data to json failed.err: %v",
			common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("ListK8SNamespaceResources rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// DeleteBatchK8SNamespaceResource 批量删除k8s命名空间资源 batch delete k8s namespaces
func (s *Storage) DeleteBatchK8SNamespaceResource(ctx context.Context, req *storage.DeleteBatchNamespaceResourceRequest,
	rsp *storage.DeleteBatchNamespaceResourceResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("DeleteBatchK8SNamespaceResource req: %v", util.PrettyStruct(req))

	data, err := dynamic.HandlerDelBatchNsResourceReq(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageDeleteResourceFail
		rsp.Message = common.BcsErrStorageDeleteResourceFailStr
		blog.Errorf("DeleteBatchK8SNamespaceResource %s | delete failed.err: %v",
			common.BcsErrStorageDeleteResourceFailStr, err)
		return nil
	}
	dynamic2.PushDeleteBatchResourceToQueue(data)
	_ = util.ListMapToListStruct(data, &rsp.Data, "DeleteBatchK8SNamespaceResource")

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("DeleteBatchK8SNamespaceResource rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// k8s cluster resources

// GetK8SClusterResources 获取k8s集群资源 get k8s cluster resources
func (s *Storage) GetK8SClusterResources(ctx context.Context, req *storage.GetClusterResourcesRequest,
	rsp *storage.GetClusterResourcesResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetK8SClusterResources req: %v", util.PrettyStruct(req))

	data, err := dynamic.HandlerGetClusterResourcesReq(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageGetResourceFail
		rsp.Message = common.BcsErrStorageGetResourceFailStr
		blog.Errorf("GetK8SClusterResources %s | query failed.err: %v", common.BcsErrStorageGetResourceFailStr, err)
		return nil
	}
	if len(data) == 0 {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageResourceNotExist
		rsp.Message = common.BcsErrStorageResourceNotExistStr
		blog.Errorf("GetK8SClusterResources %s | no date.err: %v", common.BcsErrStorageResourceNotExistStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetK8SClusterResources"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetK8SClusterResources %s | data to json failed.err: %v",
			common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetK8SClusterResources rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PutK8SClusterResources 创建k8s集群资源 create k8s cluster resources
func (s *Storage) PutK8SClusterResources(ctx context.Context, req *storage.PutClusterResourcesRequest,
	rsp *storage.PutClusterResourcesResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PutK8SClusterResources req: %v", util.PrettyStruct(req))

	data, err := dynamic.HandlerPutClusterResourcesReq(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStoragePutResourceFail
		rsp.Message = common.BcsErrStoragePutResourceFailStr
		blog.Errorf("PutK8SClusterResources %s | create failed.err: %v", common.BcsErrStoragePutResourceFailStr, err)
		return nil
	}
	dynamic2.PushCreateClusterToQueue(data)
	_ = util.MapToStruct(data, rsp.Data, "PutK8SClusterResources")

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PutK8SClusterResources rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// DeleteK8SClusterResources 删除k8s集群资源 delete k8s cluster resources
func (s *Storage) DeleteK8SClusterResources(ctx context.Context, req *storage.DeleteClusterResourcesRequest,
	rsp *storage.DeleteClusterResourcesResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("DeleteK8SClusterResources req: %v", util.PrettyStruct(req))

	data, err := dynamic.HandlerDelClusterResourcesReq(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageDeleteResourceFail
		rsp.Message = common.BcsErrStorageDeleteResourceFailStr
		blog.Errorf("DeleteK8SClusterResources %s | delete failed.err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		return nil
	}
	dynamic2.PushDeleteResourcesToQueue(data)
	_ = util.ListMapToListStruct(data, &rsp.Data, "DeleteK8SClusterResources")

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("DeleteK8SClusterResources rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// ListK8SClusterResources 批量查询k8s集群资源 list k8s cluster resources
func (s *Storage) ListK8SClusterResources(ctx context.Context, req *storage.ListClusterResourcesRequest,
	rsp *storage.ListClusterResourcesResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("ListK8SClusterResources req: %v", util.PrettyStruct(req))

	data, err := dynamic.HandlerListClusterResourcesReq(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("ListK8SClusterResources %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "ListK8SClusterResources"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("ListK8SClusterResources %s | data to json failed.err: %v",
			common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("ListK8SClusterResources rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// ListK8SClusterAllResources 批量查询k8s集群资源 list k8s cluster all resources
func (s *Storage) ListK8SClusterAllResources(ctx context.Context, req *storage.ListClusterResourcesRequest,
	rsp *storage.ListClusterResourcesResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("ListK8SClusterAllResources req: %v", util.PrettyStruct(req))

	data, err := dynamic.HandlerListClusterResourcesReq(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("ListK8SClusterAllResources %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "ListK8SClusterAllResources"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("ListK8SClusterAllResources %s | data to json failed.err: %v",
			common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("ListK8SClusterAllResources rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// DeleteBatchK8SClusterResource 批量删除k8s集群资源 batch delete k8s cluster resource
func (s *Storage) DeleteBatchK8SClusterResource(ctx context.Context, req *storage.DeleteBatchClusterResourceRequest,
	rsp *storage.DeleteBatchClusterResourceResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("DeleteBatchK8SClusterResource req: %v", util.PrettyStruct(req))

	data, err := dynamic.HandlerDelBatchClusterResourceReq(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageDeleteResourceFail
		rsp.Message = common.BcsErrStorageDeleteResourceFailStr
		blog.Errorf("DeleteBatchK8SClusterResource %s | delete failed.err: %v",
			common.BcsErrStorageDeleteResourceFailStr, err)
		return nil
	}
	dynamic2.PushDeleteBatchClusterToQueue(data)
	_ = util.ListMapToListStruct(data, &rsp.Data, "DeleteBatchK8SClusterResource")

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("DeleteBatchK8SClusterResource rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// DeleteBatchK8SClusterAllResource 批量删除k8s集群资源 batch delete k8s cluster all resource
func (s *Storage) DeleteBatchK8SClusterAllResource(ctx context.Context, req *storage.DeleteBatchClusterResourceRequest,
	rsp *storage.DeleteBatchClusterResourceResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("DeleteBatchK8SClusterAllResource req: %v", util.PrettyStruct(req))

	data, err := dynamic.HandlerDelBatchClusterResourceReq(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageDeleteResourceFail
		rsp.Message = common.BcsErrStorageDeleteResourceFailStr
		blog.Errorf("DeleteBatchK8SClusterAllResource %s | delete failed.err: %v",
			common.BcsErrStorageDeleteResourceFailStr, err)
		return nil
	}
	dynamic2.PushDeleteBatchClusterToQueue(data)
	_ = util.ListMapToListStruct(data, &rsp.Data, "DeleteBatchK8SClusterAllResource")

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("DeleteBatchK8SClusterAllResource rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// Custom resource
// Custom resources OPs

// GetCustomResources 查询自定义资源 get custom resources
func (s *Storage) GetCustomResources(ctx context.Context, req *storage.GetCustomResourcesRequest,
	rsp *storage.GetCustomResourcesResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetCustomResources req: %v", util.PrettyStruct(req))

	data, extra, err := dynamic.HandlerGetCustomResources(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageGetResourceFail
		rsp.Message = common.BcsErrStorageGetResourceFailStr
		blog.Errorf("GetCustomResources %s | query failed.err: %v", common.BcsErrStorageGetResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetCustomResources"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetCustomResources %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}
	rsp.Total = extra["total"].(int64)
	rsp.Offset = extra["offset"].(int64)
	rsp.PageSize = extra["pageSize"].(int64)

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetCustomResources rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// DeleteCustomResources 删除自定义资源 delete custom resources
func (s *Storage) DeleteCustomResources(ctx context.Context, req *storage.DeleteCustomResourcesRequest,
	rsp *storage.DeleteCustomResourcesResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("DeleteCustomResources req: %v", util.PrettyStruct(req))

	if err := dynamic.HandlerDelCustomResources(ctx, req); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageDeleteResourceFail
		rsp.Message = common.BcsErrStorageDeleteResourceFailStr
		blog.Errorf("DeleteCustomResources %s | delete failed.err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("DeleteCustomResources rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PutCustomResources 创建自定义资源 create custom resources
func (s *Storage) PutCustomResources(ctx context.Context, req *storage.PutCustomResourcesRequest,
	rsp *storage.PutCustomResourcesResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PutCustomResources req: %v", util.PrettyStruct(req))

	if err := dynamic.HandlerPutCustomResources(ctx, req); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStoragePutResourceFail
		rsp.Message = common.BcsErrStoragePutResourceFailStr
		blog.Errorf("PutCustomResources %s | create failed.err: %v", common.BcsErrStoragePutResourceFailStr, err)
		return nil
	}

	// 创建成功，并返回数据
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PutCustomResources rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// CreateCustomResourcesIndex 创建自定义资源索引 create custom resources index
func (s *Storage) CreateCustomResourcesIndex(ctx context.Context, req *storage.CreateCustomResourcesIndexRequest,
	rsp *storage.CreateCustomResourcesIndexResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("CreateCustomResourcesIndex req: %v", util.PrettyStruct(req))

	if err := dynamic.HandlerCreateCustomResourcesIndex(ctx, req); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStoragePutResourceFail
		rsp.Message = common.BcsErrStoragePutResourceFailStr
		blog.Errorf("CreateCustomResourcesIndex %s | create failed.err: %v", common.BcsErrStoragePutResourceFailStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("CreateCustomResourcesIndex rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// DeleteCustomResourcesIndex 删除自定义资源索引 delete custom resources index
func (s *Storage) DeleteCustomResourcesIndex(ctx context.Context, req *storage.DeleteCustomResourcesIndexRequest,
	rsp *storage.DeleteCustomResourcesIndexResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("DeleteCustomResourcesIndex req: %v", util.PrettyStruct(req))

	if err := dynamic.HandlerDelCustomResourcesIndex(ctx, req); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageDeleteResourceFail
		rsp.Message = common.BcsErrStorageDeleteResourceFailStr
		blog.Errorf("DeleteCustomResourcesIndex %s | delete failed.err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("DeleteCustomResourcesIndex rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// **** dynamic-query(动态查询) ****

// GetK8SIPPoolStatic 查询K8SIPPoolStatic get k8s IPPoolStatic
func (s *Storage) GetK8SIPPoolStatic(ctx context.Context, req *storage.IPPoolStaticRequest,
	rsp *storage.IPPoolStaticResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetK8SIPPoolStatic req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerIPPoolStaticRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("GetK8SIPPoolStatic %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetK8SIPPoolStatic"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetK8SIPPoolStatic %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetK8SIPPoolStatic rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostK8SIPPoolStatic 查询K8SIPPoolStatic get k8s IPPoolStatic
func (s *Storage) PostK8SIPPoolStatic(ctx context.Context, req *storage.IPPoolStaticRequest,
	rsp *storage.IPPoolStaticResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostK8SIPPoolStatic req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerIPPoolStaticRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("PostK8SIPPoolStatic %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "PostK8SIPPoolStatic"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("PostK8SIPPoolStatic %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostK8SIPPoolStatic rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// GetK8SIPPoolStaticDetail 查询K8SIPPoolStaticDetail get k8s IPPoolStatic
func (s *Storage) GetK8SIPPoolStaticDetail(ctx context.Context, req *storage.IPPoolStaticDetailRequest,
	rsp *storage.IPPoolStaticDetailResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetK8SIPPoolStaticDetail req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerIPPoolStaticDetailRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("GetK8SIPPoolStaticDetail %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetK8SIPPoolStaticDetail"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetK8SIPPoolStaticDetail %s | data to json failed.err: %v",
			common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetK8SIPPoolStaticDetail rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostK8SIPPoolStaticDetail 查询K8SIPPoolStaticDetail get k8s IPPoolStatic
func (s *Storage) PostK8SIPPoolStaticDetail(ctx context.Context, req *storage.IPPoolStaticDetailRequest,
	rsp *storage.IPPoolStaticDetailResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostK8SIPPoolStaticDetail req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerIPPoolStaticDetailRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("PostK8SIPPoolStaticDetail %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "PostK8SIPPoolStaticDetail"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("PostK8SIPPoolStaticDetail %s | data to json failed.err: %v",
			common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostK8SIPPoolStaticDetail rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// k8s

// GetPod 查询Pod get pod
func (s *Storage) GetPod(ctx context.Context, req *storage.PodRequest, rsp *storage.PodResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetPod req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerPodRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("GetPod %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetPod"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetPod %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetPod rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostPod 查询Pod get pod
func (s *Storage) PostPod(ctx context.Context, req *storage.PodRequest, rsp *storage.PodResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostPod req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerPodRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("PostPod %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "PostPod"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("PostPod %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostPod rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// GetReplicaSet 查询ReplicaSet get ReplicaSet
func (s *Storage) GetReplicaSet(ctx context.Context, req *storage.ReplicaSetRequest, rsp *storage.ReplicaSetResponse,
) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetReplicaSet req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerReplicaSetRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("GetReplicaSet %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetReplicaSet"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetReplicaSet %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetReplicaSet rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostReplicaSet 查询ReplicaSet get ReplicaSet
func (s *Storage) PostReplicaSet(ctx context.Context, req *storage.ReplicaSetRequest, rsp *storage.ReplicaSetResponse,
) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostReplicaSet req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerReplicaSetRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("PostReplicaSet %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "PostReplicaSet"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("PostReplicaSet %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostReplicaSet rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// GetDeploymentK8S 查询DeploymentK8S get Deployment
func (s *Storage) GetDeploymentK8S(ctx context.Context, req *storage.DeploymentK8SRequest,
	rsp *storage.DeploymentK8SResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetDeploymentK8S req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerDeploymentK8SRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("GetDeploymentK8S %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetDeploymentK8S"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetDeploymentK8S %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetDeploymentK8S rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostDeploymentK8S 查询DeploymentK8S get Deployment
func (s *Storage) PostDeploymentK8S(ctx context.Context, req *storage.DeploymentK8SRequest,
	rsp *storage.DeploymentK8SResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostDeploymentK8S req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerDeploymentK8SRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("PostDeploymentK8S %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "PostDeploymentK8S"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("PostDeploymentK8S %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostDeploymentK8S rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// GetServiceK8S 查询ServiceK8S get Service
func (s *Storage) GetServiceK8S(ctx context.Context, req *storage.ServiceK8SRequest, rsp *storage.ServiceK8SResponse,
) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetServiceK8S req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerServiceK8SRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("GetServiceK8S %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetServiceK8S"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetServiceK8S %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetServiceK8S rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostServiceK8S 查询ServiceK8S get Service
func (s *Storage) PostServiceK8S(ctx context.Context, req *storage.ServiceK8SRequest, rsp *storage.ServiceK8SResponse,
) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostServiceK8S req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerServiceK8SRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("PostServiceK8S %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "PostServiceK8S"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("PostServiceK8S %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostServiceK8S rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// GetConfigMapK8S 查询ConfigMapK8S get configmap
func (s *Storage) GetConfigMapK8S(ctx context.Context, req *storage.ConfigMapK8SRequest,
	rsp *storage.ConfigMapK8SResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetConfigMapK8S req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerConfigMapK8SRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("GetConfigMapK8S %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetConfigMapK8S"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetConfigMapK8S %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetConfigMapK8S rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostConfigMapK8S 查询ConfigMapK8S get configmap
func (s *Storage) PostConfigMapK8S(ctx context.Context, req *storage.ConfigMapK8SRequest,
	rsp *storage.ConfigMapK8SResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostConfigMapK8S req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerConfigMapK8SRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("PostConfigMapK8S %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "PostConfigMapK8S"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("PostConfigMapK8S %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostConfigMapK8S rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// GetSecretK8S 查询SecretK8S get secret
func (s *Storage) GetSecretK8S(ctx context.Context, req *storage.SecretK8SRequest, rsp *storage.SecretK8SResponse,
) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetSecretK8S req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerSecretK8SRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("GetSecretK8S %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetSecretK8S"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetSecretK8S %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetSecretK8S rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostSecretK8S 查询SecretK8S get secret
func (s *Storage) PostSecretK8S(ctx context.Context, req *storage.SecretK8SRequest, rsp *storage.SecretK8SResponse,
) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostSecretK8S req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerSecretK8SRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("PostSecretK8S %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "PostSecretK8S"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("PostSecretK8S %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostSecretK8S rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// GetEndpointsK8S 查询EndpointsK8S get secret
func (s *Storage) GetEndpointsK8S(ctx context.Context, req *storage.EndpointsK8SRequest,
	rsp *storage.EndpointsK8SResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetEndpointsK8S req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerEndpointsK8SRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("GetEndpointsK8S %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetEndpointsK8S"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetEndpointsK8S %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetEndpointsK8S rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostEndpointsK8S 查询EndpointsK8S get secret
func (s *Storage) PostEndpointsK8S(ctx context.Context, req *storage.EndpointsK8SRequest,
	rsp *storage.EndpointsK8SResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostEndpointsK8S req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerEndpointsK8SRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("PostEndpointsK8S %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "PostEndpointsK8S"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("PostEndpointsK8S %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostEndpointsK8S rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// GetIngress 查询Ingress get ingress
func (s *Storage) GetIngress(ctx context.Context, req *storage.IngressRequest, rsp *storage.IngressResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetIngress req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerIngressRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("GetIngress %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetIngress"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetIngress %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetIngress rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostIngress 查询Ingress get ingress
func (s *Storage) PostIngress(ctx context.Context, req *storage.IngressRequest, rsp *storage.IngressResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostIngress req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerIngressRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("PostIngress %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "PostIngress"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("PostIngress %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostIngress rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// GetNamespaceK8S 查询NamespaceK8S get namespace
func (s *Storage) GetNamespaceK8S(ctx context.Context, req *storage.NamespaceK8SRequest,
	rsp *storage.NamespaceK8SResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}

	blog.Infof("GetNamespaceK8S req: %v", util.PrettyStruct(req))
	dynamicquery.HandlerNamespaceK8SRequest(ctx, req, rsp)
	blog.Infof("GetNamespaceK8S rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostNamespaceK8S 查询NamespaceK8S get namespace
func (s *Storage) PostNamespaceK8S(ctx context.Context, req *storage.NamespaceK8SRequest,
	rsp *storage.NamespaceK8SResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}

	blog.Infof("PostNamespaceK8S req: %v", util.PrettyStruct(req))
	dynamicquery.HandlerNamespaceK8SRequest(ctx, req, rsp)
	blog.Infof("PostNamespaceK8S rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// GetNode 查询Node get node
func (s *Storage) GetNode(ctx context.Context, req *storage.NodeRequest, rsp *storage.NodeResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetNode req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerNodeRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("GetNode %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetNode"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetNode %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetNode rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostNode 查询Node get node
func (s *Storage) PostNode(ctx context.Context, req *storage.NodeRequest, rsp *storage.NodeResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostNode req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerNodeRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("PostNode %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "PostNode"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("PostNode %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostNode rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// GetDaemonSet 查询DaemonSet get daemonset
func (s *Storage) GetDaemonSet(ctx context.Context, req *storage.DaemonSetRequest, rsp *storage.DaemonSetResponse,
) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetDaemonSet req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerDaemonSetRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("GetDaemonSet %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetDaemonSet"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetDaemonSet %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetDaemonSet rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostDaemonSet 查询DaemonSet get daemonset
func (s *Storage) PostDaemonSet(ctx context.Context, req *storage.DaemonSetRequest, rsp *storage.DaemonSetResponse,
) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostDaemonSet req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerDaemonSetRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("PostDaemonSet %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "PostDaemonSet"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("PostDaemonSet %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostDaemonSet rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// GetJob 查询Job get job
func (s *Storage) GetJob(ctx context.Context, req *storage.JobRequest, rsp *storage.JobResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetJob req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerJobRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("GetJob %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetJob"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetJob %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetJob rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostJob 查询Job get job
func (s *Storage) PostJob(ctx context.Context, req *storage.JobRequest, rsp *storage.JobResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostJob req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerJobRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("PostJob %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "PostJob"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("PostJob %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostJob rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// GetStatefulSet 查询StatefulSet get statefulset
func (s *Storage) GetStatefulSet(ctx context.Context, req *storage.StatefulSetRequest,
	rsp *storage.StatefulSetResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetStatefulSet req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerStatefulSetRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("GetStatefulSet %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetStatefulSet"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetStatefulSet %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetStatefulSet rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostStatefulSet 查询StatefulSet get statefulset
func (s *Storage) PostStatefulSet(ctx context.Context, req *storage.StatefulSetRequest,
	rsp *storage.StatefulSetResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostStatefulSet req: %v", util.PrettyStruct(req))

	data, err := dynamicquery.HandlerStatefulSetRequest(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("PostStatefulSet %s | query failed.err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "PostStatefulSet"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("PostStatefulSet %s | data to json failed.err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostStatefulSet rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// **** dynamic-watch(watch) *****

// WatchDynamic watch dynamic data
func (s *Storage) WatchDynamic(ctx context.Context, req *storage.WatchDynamicRequest,
	stream storage.Storage_WatchDynamicStream) error {
	// nolint
	defer stream.Close()
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		_ = stream.SendMsg(
			&storage.WatchDynamicResponse{
				Type: -1,
			},
		)
		blog.Infof("%s", err.Error())
		return nil
	}
	// 打印请求体
	blog.Infof("WatchDynamic req: %v", util.PrettyStruct(req))

	event, err := dynamicwatch.HandlerWatchDynamic(ctx, req)
	if err != nil {
		_ = stream.SendMsg(
			&storage.WatchDynamicResponse{
				Type: -1,
			},
		)
		return errors.Wrapf(err, "watch failed. clusterId: '%s'", req.ClusterId)
	}

	for {
		select {
		case <-ctx.Done():
			blog.Infof("stop watch by server. clusterId: '%s'", req.ClusterId)
			return nil
		case e := <-event:
			if e.Type == lib.Brk {
				blog.Infof("stop watch by event break. clusterId: '%s'", req.ClusterId)
				return nil
			}
			v := &storage.WatchDynamicResponse{}
			if err = util.StructToStruct(e, v, "WatchDynamic"); err != nil {
				return errors.Wrapf(err, "event to json failed")
			}
			//  note: 无论是使用stream.SendMsg(v)还是使用stream.Send(v)，发送返回值时，都会遇到如下情况：
			//  当使用http调用v2 server时，由于使用grpc gateway 代理，故gateway会在收到grpc返回值后，会在原来的返回值上再包装一层，
			//  因此，会导致v1 server返回与v1 server返回值不一致。
			//  如： 假设grpc返回值为 { "msg" : "hell" }，gateway收到返回值后，会进行包装，如：{"result" : { "msg" : "hell" }}
			if err = stream.SendMsg(v); err != nil {
				return errors.Wrapf(err, "watchDynamic send message failed")
			}
			blog.Infof("WatchDynamic | event: %s", util.PrettyStruct(e))
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

// WatchContainer watch container data
func (s *Storage) WatchContainer(ctx context.Context, req *storage.WatchContainerRequest,
	stream storage.Storage_WatchContainerStream) error {
	// nolint
	defer stream.Close()
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		_ = stream.SendMsg(
			&storage.WatchContainerResponse{
				Type: -1,
			},
		)
		blog.Infof("%s", err.Error())
		return nil
	}
	// 打印请求体
	blog.Infof("WatchContainer req: %v", util.PrettyStruct(req))

	event, err := dynamicwatch.HandlerWatchContainer(ctx, req)
	if err != nil {
		_ = stream.SendMsg(
			&storage.WatchContainerResponse{
				Type: -1,
			},
		)
		return errors.Wrapf(err, "watch failed. clusterId: '%s'", req.ClusterId)
	}

	for {
		select {
		case <-ctx.Done():
			blog.Infof("stop watch by server. clusterId: '%s'", req.ClusterId)
			return nil
		case e := <-event:
			if e.Type == lib.Brk {
				blog.Infof("stop watch by event break. clusterId: '%s'", req.ClusterId)
				return nil
			}
			v := &storage.WatchContainerResponse{}
			if err = util.StructToStruct(e, v, "WatchContainer"); err != nil {
				return errors.Wrapf(err, "event to json failed")
			}
			//  note: 无论是使用stream.SendMsg(v)还是使用stream.Send(v)，发送返回值时，都会遇到如下情况：
			//  当使用http调用v2 server时，由于使用grpc gateway 代理，故gateway会在收到grpc返回值后，会在原来的返回值上再包装一层，
			//  因此，会导致v1 server返回与v1 server返回值不一致。
			//  如： 假设grpc返回值为 { "msg" : "hell" }，gateway收到返回值后，会进行包装，如：{"result" : { "msg" : "hell" }}
			if err = stream.SendMsg(v); err != nil {
				return errors.Wrapf(err, "watchContainer send message failed")
			}
			blog.Infof("WatchContainer | event: %s", util.PrettyStruct(e))
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

// **** events(事件) *****

// PutEvent 创建事件 create event
func (s *Storage) PutEvent(ctx context.Context, req *storage.PutEventRequest, rsp *storage.PutEventResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PutEvent req: %v", util.PrettyStruct(req))

	if err := events.HandlerPutEvent(ctx, req); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStoragePutResourceFail
		rsp.Message = common.BcsErrStoragePutResourceFailStr
		blog.Errorf("PutEvent %s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PutEvent rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// ListEvent 查询事件 list event
func (s *Storage) ListEvent(ctx context.Context, req *storage.ListEventRequest, rsp *storage.ListEventResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	if req.Env == "" {
		req.Env = "k8s"
	}
	// 打印请求体
	blog.Infof("ListEvent req: %v", util.PrettyStruct(req))

	data, total, err := events.HandlerListEvent(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("HandlerListEvent %s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "HandlerListEvent"); err != nil {
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("HandlerListEvent %s | err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr
	rsp.Extra, _ = structpb.NewStruct(map[string]interface{}{"total": total})
	blog.Infof("ListEvent rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// WatchEvent watch event
func (s *Storage) WatchEvent(ctx context.Context, req *storage.WatchEventRequest,
	stream storage.Storage_WatchEventStream) error {
	// nolint
	defer stream.Close()
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		_ = stream.SendMsg(
			&storage.WatchEventResponse{
				Type: -1,
			},
		)
		blog.Infof("%s", err.Error())
		return nil
	}
	// 打印请求体
	blog.Infof("WatchEvent req: %v", util.PrettyStruct(req))

	event, err := events.HandlerWatch(ctx, req)
	if err != nil {
		_ = stream.SendMsg(
			&storage.WatchEventResponse{
				Type: -1,
			},
		)
		return errors.Wrapf(err, "watch failed. clusterId: '%s'", req.ClusterId)
	}

	for {
		select {
		case <-ctx.Done():
			blog.Infof("stop watch by server. clusterId: '%s'", req.ClusterId)
			return nil
		case e := <-event:
			if e.Type == lib.Brk {
				blog.Infof("stop watch by event break. clusterId: '%s'", req.ClusterId)
				return nil
			}
			v := &storage.WatchEventResponse{}
			if err = util.StructToStruct(e, v, "WatchEvent"); err != nil {
				return errors.Wrapf(err, "event to json failed")
			}
			//  note: 无论是使用stream.SendMsg(v)还是使用stream.Send(v)，发送返回值时，都会遇到如下情况：
			//  当使用http调用v2 server时，由于使用grpc gateway 代理，故gateway会在收到grpc返回值后，会在原来的返回值上再包装一层，
			//  因此，会导致v1 server返回与v1 server返回值不一致。
			//  如： 假设grpc返回值为 { "msg" : "hell" }，gateway收到返回值后，会进行包装，如：{"result" : { "msg" : "hell" }}
			if err = stream.SendMsg(v); err != nil {
				return errors.Wrapf(err, "watchEvent send message failed")
			}
			blog.Infof("WatchEvent | event: %s", util.PrettyStruct(e))
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

// **** host config(主机配置) *****

// GetHost 获取主机配置 get host
func (s *Storage) GetHost(ctx context.Context, req *storage.GetHostRequest, rsp *storage.GetHostResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetHost req: %v", util.PrettyStruct(req))

	data, err := hostconfig.HandlerGetHost(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageGetResourceFail
		rsp.Message = common.BcsErrStorageGetResourceFailStr
		blog.Errorf("GetHost %s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetHost"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetHost %s | err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	// 查询成功，并返回数据
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetHost rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PutHost 创建主机配置 put host
func (s *Storage) PutHost(ctx context.Context, req *storage.PutHostRequest, rsp *storage.PutHostResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PutHost req: %v", util.PrettyStruct(req))

	if err := hostconfig.HandlerPutHost(ctx, req); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStoragePutResourceFail
		rsp.Message = common.BcsErrStoragePutResourceFailStr
		blog.Errorf("PutHost %s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PutHost rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// DeleteHost 删除主机配置 delete host
func (s *Storage) DeleteHost(ctx context.Context, req *storage.DeleteHostRequest, rsp *storage.DeleteHostResponse,
) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("DeleteHost req: %v", util.PrettyStruct(req))

	if err := hostconfig.HandlerDeleteHost(ctx, req); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageDeleteResourceFail
		rsp.Message = common.BcsErrStorageDeleteResourceFailStr
		blog.Errorf("DeleteHost %s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("DeleteHost rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// ListHost 批量查询主机配置 list host
func (s *Storage) ListHost(ctx context.Context, req *storage.ListHostRequest, rsp *storage.ListHostResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("ListHost req: %v", util.PrettyStruct(req))

	data, err := hostconfig.HandlerListHost(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("ListHost %s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "ListHost"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("ListHost %s | err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("ListHost rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PutClusterRelation 修改集群关系 update cluster relation
func (s *Storage) PutClusterRelation(ctx context.Context, req *storage.PutClusterRelationRequest,
	rsp *storage.PutClusterRelationResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PutClusterRelation req: %v", util.PrettyStruct(req))

	if err := hostconfig.HandlerPutClusterRelation(ctx, req); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStoragePutResourceFail
		rsp.Message = common.BcsErrStoragePutResourceFailStr
		blog.Errorf("PutClusterRelation %s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PutClusterRelation rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PostClusterRelation 创建集群关系 create cluster relation
func (s *Storage) PostClusterRelation(ctx context.Context, req *storage.PostClusterRelationRequest,
	rsp *storage.PostClusterRelationResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PostClusterRelation req: %v", util.PrettyStruct(req))

	if err := hostconfig.HandlerPostClusterRelation(ctx, req); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStoragePutResourceFail
		rsp.Message = common.BcsErrStoragePutResourceFailStr
		blog.Errorf("PostClusterRelation %s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PostClusterRelation rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// **** metric(指标)  ****

// GetMetric 查询警告指标 get metric
func (s *Storage) GetMetric(ctx context.Context, req *storage.GetMetricRequest, rsp *storage.GetMetricResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("GetMetric req: %v", util.PrettyStruct(req))

	data, err := metric.HandlerGetMetric(ctx, req)
	if err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageGetResourceFail
		rsp.Message = common.BcsErrStorageGetResourceFailStr
		blog.Errorf("GetMetric %s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		return nil
	}
	if len(data) == 0 {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageResourceNotExist
		rsp.Message = common.BcsErrStorageResourceNotExistStr
		blog.Errorf("GetMetric %s | err: %v", common.BcsErrStorageResourceNotExistStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "GetMetric"); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("GetMetric %s | err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("GetMetric rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// PutMetric 修改指标 update metric
func (s *Storage) PutMetric(ctx context.Context, req *storage.PutMetricRequest, rsp *storage.PutMetricResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("PutMetric req: %v", util.PrettyStruct(req))

	if err := metric.HandlerPutMetric(ctx, req); err != nil {
		rsp.Result = false
		rsp.Code = common.BcsErrStorageGetResourceFail
		rsp.Message = common.BcsErrStorageGetResourceFailStr
		blog.Errorf("PutMetric %s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		return nil
	}

	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("PutMetric rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// DeleteMetric 删除指标 delete metrics
func (s *Storage) DeleteMetric(ctx context.Context, req *storage.DeleteMetricRequest, rsp *storage.DeleteMetricResponse,
) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("DeleteMetric req: %v", util.PrettyStruct(req))

	if err := metric.HandlerDeleteMetric(ctx, req); err != nil {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStorageDeleteResourceFail
		rsp.Message = common.BcsErrStorageDeleteResourceFailStr
		blog.Errorf("DeleteMetric %s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("DeleteMetric rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// QueryMetric 查询指标 get metric
func (s *Storage) QueryMetric(ctx context.Context, req *storage.QueryMetricRequest, rsp *storage.QueryMetricResponse,
) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("QueryMetric req: %v", util.PrettyStruct(req))

	data, err := metric.HandlerQueryMetric(ctx, req)
	if err != nil {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("QueryMetric %s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}
	if len(data) == 0 {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStorageResourceNotExist
		rsp.Message = common.BcsErrStorageResourceNotExistStr
		blog.Errorf("QueryMetric %s | err: %v", common.BcsErrStorageResourceNotExistStr, err)
		return nil
	}
	if err = util.ListMapToListStruct(data, &rsp.Data, "QueryMetric"); err != nil {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStorageReturnDataIsNotJson
		rsp.Message = common.BcsErrStorageReturnDataIsNotJsonStr
		blog.Errorf("QueryMetric %s | err: %v", common.BcsErrStorageReturnDataIsNotJsonStr, err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("QueryMetric rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// ListMetricTables 查询警告表 list metric tables
func (s *Storage) ListMetricTables(ctx context.Context, req *storage.ListMetricTablesRequest,
	rsp *storage.ListMetricTablesResponse) error {
	// 打印请求体
	blog.Infof("ListMetricTables req: %v", util.PrettyStruct(req))

	data, err := metric.HandlerListMetricTables(ctx)
	if err != nil {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStorageDecodeListResourceFail
		rsp.Message = common.BcsErrStorageDecodeListResourceFailStr
		blog.Errorf("ListMetricTables %s | err: %v", common.BcsErrStorageDecodeListResourceFailStr, err)
		return nil
	}
	if len(data) == 0 {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStorageResourceNotExist
		rsp.Message = common.BcsErrStorageResourceNotExistStr
		blog.Errorf("ListMetricTables %s | err: %v", common.BcsErrStorageResourceNotExistStr, err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Data = data
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("ListMetricTables rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// **** metric watch ****

// WatchMetric watch metric
func (s *Storage) WatchMetric(ctx context.Context, req *storage.WatchMetricRequest,
	stream storage.Storage_WatchMetricStream) error {
	// nolint
	defer stream.Close()
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		_ = stream.SendMsg(
			&storage.WatchMetricResponse{
				Type: -1,
			},
		)
		blog.Infof("%s", err.Error())
		return nil
	}
	// 打印请求体
	blog.Infof("WatchMetric req: %v", util.PrettyStruct(req))

	event, err := metricwatch.HandlerWatch(ctx, req)
	if err != nil {
		_ = stream.SendMsg(
			&storage.WatchMetricResponse{
				Type: -1,
			},
		)
		return errors.Wrapf(err, "watch failed. clusterId: '%s'", req.ClusterId)
	}

	for {
		select {
		case <-ctx.Done():
			blog.Infof("stop watch by server. clusterId: '%s'", req.ClusterId)
			return nil
		case e := <-event:
			if e.Type == lib.Brk {
				blog.Infof("stop watch by event break. clusterId: '%s'", req.ClusterId)
				return nil
			}
			if e.Type != lib.Del && e.Value[constants.TypeTag] != req.Type {
				// 不符合查询条件，则继续watch
				continue
			}
			v := &storage.WatchMetricResponse{}
			if err = util.StructToStruct(e, v, "WatchMetric"); err != nil {
				return errors.Wrapf(err, "event to json failed")
			}
			//  note: 无论是使用stream.SendMsg(v)还是使用stream.Send(v)，发送返回值时，都会遇到如下情况：
			//  当使用http调用v2 server时，由于使用grpc gateway 代理，故gateway会在收到grpc返回值后，会在原来的返回值上再包装一层，
			//  因此，会导致v1 server返回与v1 server返回值不一致。
			//  如： 假设grpc返回值为 { "msg" : "hell" }，gateway收到返回值后，会进行包装，如：{"result" : { "msg" : "hell" }}
			if err = stream.SendMsg(v); err != nil {
				return errors.Wrapf(err, "watchMetric send message failed")
			}
			blog.Infof("WatchMetric | event: %s", util.PrettyStruct(e))
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

// **** watch k8s ****
// k8s

// K8SGetWatchResource 查询watch资源 get watch resources
func (s *Storage) K8SGetWatchResource(ctx context.Context, req *storage.K8SGetWatchResourceRequest,
	rsp *storage.K8SGetWatchResourceResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("K8SGetWatchResource req: %v", util.PrettyStruct(req))

	data, err := watchk8smesos.HandlerK8SGetWatchResource(ctx, req)
	if err != nil {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStorageGetResourceFail
		rsp.Message = common.BcsErrStorageGetResourceFailStr
		blog.Errorf("K8SGetWatchResource %s | err: %v", common.BcsErrStorageGetResourceFailStr, err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr
	rsp.Data, _ = structpb.NewStruct(util.StructToMap(data))

	// 打印响应体
	blog.Infof("K8SGetWatchResource rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// K8SPutWatchResource 修改watch资源 update watch resource
func (s *Storage) K8SPutWatchResource(ctx context.Context, req *storage.K8SPutWatchResourceRequest,
	rsp *storage.K8SPutWatchResourceResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("K8SPutWatchResource req: %v", util.PrettyStruct(req))

	if err := watchk8smesos.HandlerK8SPutWatchResource(ctx, req); err != nil {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStoragePutResourceFail
		rsp.Message = common.BcsErrStoragePutResourceFailStr
		blog.Errorf("K8SPutWatchResource %s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("K8SPutWatchResource rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// K8SDeleteWatchResource 删除watch资源 delete watch resource
func (s *Storage) K8SDeleteWatchResource(ctx context.Context, req *storage.K8SDeleteWatchResourceRequest,
	rsp *storage.K8SDeleteWatchResourceResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("K8SDeleteWatchResource req: %v", util.PrettyStruct(req))

	if err := watchk8smesos.HandlerK8SDeleteWatchResource(ctx, req); err != nil {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStorageDeleteResourceFail
		rsp.Message = common.BcsErrStorageDeleteResourceFailStr
		blog.Errorf("K8SDeleteWatchResource %s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("K8SDeleteWatchResource rsp: %v", util.PrettyStruct(rsp))
	return nil
}

// K8SListWatchResource 批量查询watch资源 batch get watch resource
func (s *Storage) K8SListWatchResource(ctx context.Context, req *storage.K8SListWatchResourceRequest,
	rsp *storage.K8SListWatchResourceResponse) error {
	// 如果请求合法,则执行函数请求
	if err := req.Validate(); err != nil {
		rsp.Result = false
		rsp.Message = err.Error()
		rsp.Code = common.AdditionErrorCode + 500
		return nil
	}
	// 打印请求体
	blog.Infof("K8SListWatchResource req: %v", util.PrettyStruct(req))

	data, err := watchk8smesos.HandlerK8SListWatchResource(ctx, req)
	if err != nil {
		// 如果处理过程中出错，将错误信息直接返回
		rsp.Result = false
		rsp.Code = common.BcsErrStorageListResourceFail
		rsp.Message = common.BcsErrStorageListResourceFailStr
		blog.Errorf("K8SListWatchResource %s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		return nil
	}

	// 将查询结果添加至响应体
	rsp.Data = data
	rsp.Result = true
	rsp.Code = common.BcsSuccess
	rsp.Message = common.BcsSuccessStr

	// 打印响应体
	blog.Infof("K8SListWatchResource rsp: %v", util.PrettyStruct(rsp))
	return nil
}
