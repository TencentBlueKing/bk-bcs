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
 *
 */

package api

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/esb/apigateway/bkdata"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/api/proto/logmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v1"
)

// LogManagerServerImpl is grpc Server
type LogManagerServerImpl struct {
	logManager *k8s.LogManager
	apiHost    string
}

// ObtainDataID ObtainDataid
func (l *LogManagerServerImpl) ObtainDataID(ctx context.Context, req *proto.ObtainDataidReq, resp *proto.ObtainDataidResp) error {
	// bkdata api esb client
	client := bkdata.NewClientInterface(bkdata.BKDataClientConfig{
		BkAppCode:                  req.AppCode,
		BkUsername:                 req.UserName,
		BkAppSecret:                req.AppSecret,
		BkdataAuthenticationMethod: "user",
		Host:                       l.apiHost,
	})
	config := bkdata.NewDefaultAccessDeployPlanConfig()
	config.BkBizID = int(req.BizID)
	config.AccessRawData.RawDataName = req.DataName
	config.AccessRawData.RawDataAlias = req.DataName
	if req.Maintainers != "" {
		config.AccessRawData.Maintainer = req.Maintainers
	} else {
		config.AccessRawData.Maintainer = req.UserName
	}
	// request
	dataid, err := client.ObtainDataID(config)
	if err != nil {
		blog.Errorf("Obtain dataid failed: %s, req info: %+v", err.Error(), *req)
		resp.ErrCode = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
		resp.Message = err.Error()
		return err
	}
	// return
	resp.ErrCode = proto.ErrCode_ERROR_OK
	resp.DataID = int32(dataid)
	return nil
}

// CreateCleanStrategy CreateCleanStrategy
func (l *LogManagerServerImpl) CreateCleanStrategy(ctx context.Context, req *proto.CreateCleanStrategyReq, resp *proto.CommonResp) error {
	client := bkdata.NewClientInterface(bkdata.BKDataClientConfig{
		BkAppCode:                  req.AppCode,
		BkUsername:                 req.UserName,
		BkAppSecret:                req.AppSecret,
		BkdataAuthenticationMethod: "user",
		Host:                       l.apiHost,
	})
	var config bkdata.DataCleanStrategy
	// whether to use default clean strategy
	if req.Default {
		config = bkdata.NewDefaultCleanStrategy()
		config.RawDataID = int(req.DataID)
		config.BkBizID = int(req.BizID)
		config.ResultTableName = req.ResultTableName
		config.ResultTableNameAlias = req.ResultTableName
	} else {
		config = bkdata.DataCleanStrategy{}
		config.BkBizID = int(req.BizID)
		config.RawDataID = int(req.DataID)
		config.CleanConfigName = req.StrategyName
		config.ResultTableName = req.ResultTableName
		config.ResultTableNameAlias = req.ResultTableName
		config.JSONConfig = req.JSONConfig
		for _, v := range req.Fields {
			field := bkdata.Fields{
				FieldName:   v.FieldName,
				FieldAlias:  v.FieldAlias,
				FieldType:   v.FieldType,
				IsDimension: v.IsDimension,
				FieldIndex:  int(v.FieldIndex),
			}
			if v.FieldAlias == "" {
				field.FieldAlias = v.FieldName
			}
			config.Fields = append(config.Fields, field)
		}
	}
	// request
	err := client.SetCleanStrategy(config)
	if err != nil {
		blog.Errorf("Create clean strategy failed: %s, req info: %+v", err.Error(), *req)
		resp.ErrCode = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
		resp.Message = err.Error()
		return err
	}
	resp.ErrCode = proto.ErrCode_ERROR_OK
	return nil
}

// ListLogCollectionTask ListLogCollectionTask
func (l *LogManagerServerImpl) ListLogCollectionTask(ctx context.Context, req *proto.ListLogCollectionTaskReq, resp *proto.ListLogCollectionTaskResp) error {
	blog.Infof("reqest: %+v", req)
	filter := &config.CollectionFilterConfig{
		ClusterIDs:      req.ClusterIDs,
		ConfigName:      req.ConfigName,
		ConfigNamespace: req.ConfigNamespace,
	}
	// send message to manager
	recvCh := make(chan interface{})
	defer close(recvCh)
	message := k8s.RequestMessage{
		Data:   filter,
		RespCh: recvCh,
	}
	l.logManager.GetLogCollectionTask <- &message
	result := make(map[string][]config.CollectionConfig)
	// receive data from manager
	for {
		select {
		case conf, ok := <-recvCh:
			if !ok {
				resp.ErrCode = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
				err := fmt.Errorf("error receiving response data from log manager")
				resp.Message = err.Error()
				return err
			}
			switch v := conf.(type) {
			case error:
				resp.ErrCode = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
				resp.Message = v.Error()
				return v
			case string:
				if v == "termination" {
					goto exit
				}
				blog.Warnf("Unexpected string value (%s) while listing log collection configs", v)
			case config.CollectionConfig:
				result[v.ConfigSpec.ClusterId] = append(result[v.ConfigSpec.ClusterId], v)
			default:
				blog.Warnf("Unrecognized type of value (%+v) while receving data from manager", v)
			}
		}
	}
exit:
	// PARSE []logconfigs TO clusterid => []logconfigs
	for k, v := range result {
		respItem := proto.ListLogCollectionTaskRespItem{
			ClusterID: k,
			Configs:   make([]*proto.LogCollectionTaskConfig, len(v)),
		}
		for i, conf := range v {
			respItem.Configs[i] = l.buildProtoLogCollectionTaskConfig(conf)
		}
		resp.Data = append(resp.Data, &respItem)
	}
	resp.ErrCode = proto.ErrCode_ERROR_OK
	return nil
}

// CreateLogCollectionTask CreateLogCollectionTask
func (l *LogManagerServerImpl) CreateLogCollectionTask(ctx context.Context, req *proto.CreateLogCollectionTaskReq, resp *proto.CommonResp) error {
	blog.Infof("reqest: %+v", req)
	config := l.buildLogCollectionTaskConfig(req)
	// send message to manager
	recvCh := make(chan interface{})
	defer close(recvCh)
	message := k8s.RequestMessage{
		Data:   config,
		RespCh: recvCh,
	}
	l.logManager.AddLogCollectionTask <- &message
	// receive data from manager
	data, ok := <-recvCh
	if !ok {
		resp.ErrCode = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
		err := fmt.Errorf("error receiving response data from log manager")
		resp.Message = err.Error()
		return err
	}
	switch v := data.(type) {
	case error:
		resp.ErrCode = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
		resp.Message = v.Error()
		return v
	case string:
		if v == "termination" {
			break
		}
		blog.Warnf("Unexpected string value (%s) while listing log collection configs", v)
	default:
		blog.Warnf("Unrecognized type of value (%+v) while receving data from manager", v)
	}
	resp.ErrCode = proto.ErrCode_ERROR_OK
	return nil
}

// DeleteLogCollectionTask DeleteLogCollectionTask
func (l *LogManagerServerImpl) DeleteLogCollectionTask(ctx context.Context, req *proto.DeleteLogCollectionTaskReq, resp *proto.CommonResp) error {
	blog.Infof("reqest: %+v", req)
	filter := &config.CollectionFilterConfig{
		ClusterIDs:      req.ClusterIDs,
		ConfigName:      req.ConfigName,
		ConfigNamespace: req.ConfigNamespace,
	}
	// send message to manager
	recvCh := make(chan interface{})
	defer close(recvCh)
	message := k8s.RequestMessage{
		Data:   filter,
		RespCh: recvCh,
	}
	l.logManager.DeleteLogCollectionTask <- &message
	// receive data from manager
	data, ok := <-recvCh
	if !ok {
		resp.ErrCode = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
		err := fmt.Errorf("error receiving response data from log manager")
		resp.Message = err.Error()
		return err
	}
	switch v := data.(type) {
	case error:
		resp.ErrCode = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
		resp.Message = v.Error()
		return v
	case string:
		if v == "termination" {
			break
		}
		blog.Warnf("Unexpected string value (%s) while listing log collection configs", v)
	default:
		blog.Warnf("Unrecognized type of value (%+v) while receving data from manager", v)
	}
	resp.ErrCode = proto.ErrCode_ERROR_OK
	return nil
}

// buildProtoLogCollectionTaskConfig convert log config message from CollectionConfig to protobuf struct type
func (l *LogManagerServerImpl) buildProtoLogCollectionTaskConfig(conf config.CollectionConfig) *proto.LogCollectionTaskConfig {
	ret := &proto.LogCollectionTaskConfig{
		ConfigName:      conf.ConfigName,
		ConfigNamespace: conf.ConfigNamespace,
		Config: &proto.LogCollectionTaskConfigSpec{
			ConfigType:        conf.ConfigSpec.ConfigType,
			AppId:             conf.ConfigSpec.AppId,
			ClusterId:         conf.ConfigSpec.ClusterId,
			Stdout:            conf.ConfigSpec.Stdout,
			StdDataId:         conf.ConfigSpec.StdDataId,
			NonStdDataId:      conf.ConfigSpec.NonStdDataId,
			LogPaths:          conf.ConfigSpec.LogPaths,
			LogTags:           conf.ConfigSpec.LogTags,
			WorkloadType:      conf.ConfigSpec.WorkloadType,
			WorkloadName:      conf.ConfigSpec.WorkloadName,
			WorkloadNamespace: conf.ConfigSpec.WorkloadNamespace,
			PodLabels:         conf.ConfigSpec.PodLabels,
		},
	}
	containerconf := make([]*proto.ContainerConf, len(conf.ConfigSpec.ContainerConfs))
	for i, v := range conf.ConfigSpec.ContainerConfs {
		containerconf[i] = &proto.ContainerConf{
			ContainerName: v.ContainerName,
			Stdout:        v.Stdout,
			StdDataId:     v.StdDataId,
			NonStdDataId:  v.NonStdDataId,
			LogPaths:      v.LogPaths,
			LogTags:       v.LogTags,
		}
	}
	ret.Config.ContainerConfs = containerconf
	return ret
}

// buildLogCollectionTaskConfig convert log config message from protobuf struct type to CollectionConfig
func (l *LogManagerServerImpl) buildLogCollectionTaskConfig(conf *proto.CreateLogCollectionTaskReq) *config.CollectionConfig {
	ret := &config.CollectionConfig{
		ClusterIDs:      conf.ClusterIDs,
		ConfigName:      conf.Config.ConfigName,
		ConfigNamespace: conf.Config.ConfigNamespace,
		ConfigSpec: bcsv1.BcsLogConfigSpec{
			ConfigType:        conf.Config.Config.ConfigType,
			AppId:             conf.Config.Config.AppId,
			ClusterId:         conf.Config.Config.ClusterId,
			Stdout:            conf.Config.Config.Stdout,
			StdDataId:         conf.Config.Config.StdDataId,
			NonStdDataId:      conf.Config.Config.NonStdDataId,
			LogPaths:          conf.Config.Config.LogPaths,
			LogTags:           conf.Config.Config.LogTags,
			WorkloadType:      conf.Config.Config.WorkloadType,
			WorkloadName:      conf.Config.Config.WorkloadName,
			WorkloadNamespace: conf.Config.Config.WorkloadNamespace,
			PodLabels:         conf.Config.Config.PodLabels,
		},
	}
	containerconf := make([]bcsv1.ContainerConf, len(conf.Config.Config.ContainerConfs))
	for i, v := range conf.Config.Config.ContainerConfs {
		containerconf[i] = bcsv1.ContainerConf{
			ContainerName: v.ContainerName,
			Stdout:        v.Stdout,
			StdDataId:     v.StdDataId,
			NonStdDataId:  v.NonStdDataId,
			LogPaths:      v.LogPaths,
			LogTags:       v.LogTags,
		}
	}
	ret.ConfigSpec.ContainerConfs = containerconf
	return ret
}
