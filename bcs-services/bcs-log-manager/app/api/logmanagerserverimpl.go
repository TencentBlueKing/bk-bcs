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
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bkdata "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/apigateway/bkdata"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/api/proto/logmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/apis/bk-bcs/v1"
)

const timeout = time.Second * 10

// LogManagerServerImpl is grpc Server
type LogManagerServerImpl struct {
	logManager          k8s.LogManagerInterface
	apiHost             string
	bkdataClientCreator bkdata.ClientCreatorInterface
}

// ObtainDataID ObtainDataid
func (l *LogManagerServerImpl) ObtainDataID(ctx context.Context, req *proto.ObtainDataidReq, resp *proto.ObtainDataidResp) error {
	// bkdata api esb client
	client := l.bkdataClientCreator.NewClientFromConfig(bkdata.BKDataClientConfig{
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
		resp.ErrCode = int32(proto.ErrCode_ERROR_LOG_MANAGER_FAILED)
		resp.ErrName = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
		resp.Message = err.Error()
		return err
	}
	// create default data clean strategy
	strategy := bkdata.NewDefaultCleanStrategy()
	strategy.RawDataID = int(dataid)
	strategy.BkBizID = int(req.BizID)
	tableName := fmt.Sprintf("container_log_clean_strategy_%d", dataid)
	strategy.ResultTableName = tableName
	strategy.ResultTableNameAlias = tableName
	err = client.SetCleanStrategy(strategy)
	// return
	resp.ErrCode = int32(proto.ErrCode_ERROR_OK)
	resp.ErrName = proto.ErrCode_ERROR_OK
	resp.DataID = int32(dataid)
	if err != nil {
		resp.Message = fmt.Sprintf("Create default clean strategy failed: %s", err.Error())
	}
	return nil
}

// CreateCleanStrategy CreateCleanStrategy
func (l *LogManagerServerImpl) CreateCleanStrategy(ctx context.Context, req *proto.CreateCleanStrategyReq, resp *proto.CommonResp) error {
	client := l.bkdataClientCreator.NewClientFromConfig(bkdata.BKDataClientConfig{
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
		resp.ErrCode = int32(proto.ErrCode_ERROR_LOG_MANAGER_FAILED)
		resp.ErrName = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
		resp.Message = err.Error()
		return err
	}
	resp.ErrCode = int32(proto.ErrCode_ERROR_OK)
	resp.ErrName = proto.ErrCode_ERROR_OK
	return nil
}

// ListLogCollectionTask ListLogCollectionTask
func (l *LogManagerServerImpl) ListLogCollectionTask(ctx context.Context, req *proto.ListLogCollectionTaskReq, resp *proto.ListLogCollectionTaskResp) error {
	blog.Infof("reqest: %+v", req)
	filter := &config.CollectionFilterConfig{
		ClusterIDs:      strings.ToLower(req.ClusterIDs),
		ConfigName:      req.ConfigName,
		ConfigNamespace: req.ConfigNamespace,
	}
	resultCh := make(chan map[string][]config.CollectionConfig)
	go func() {
		// send message to manager
		result := l.logManager.HandleListLogCollectionTask(ctx, filter)
		resultCh <- result
		close(resultCh)
	}()
	var result map[string][]config.CollectionConfig
	select {
	// cancel
	case <-ctx.Done():
		blog.Warnf("LogManagerServerImpl ListLogCollectionTask canceled: %s", ctx.Err().Error())
		resp.ErrCode = int32(proto.ErrCode_ERROR_LOG_MANAGER_FAILED)
		resp.ErrName = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
		resp.Message = ctx.Err().Error()
		return ctx.Err()
	// timeout
	case <-time.After(timeout):
		blog.Warnf("LogManagerServerImpl ListLogCollectionTask timeout")
		resp.ErrCode = int32(proto.ErrCode_ERROR_LOG_MANAGER_FAILED)
		resp.ErrName = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
		resp.Message = "LogManagerServerImpl ListLogCollectionTask timeout"
		return fmt.Errorf("LogManagerServerImpl ListLogCollectionTask timeout")
	// get result
	case result = <-resultCh:
		if result == nil {
			resp.ErrCode = int32(proto.ErrCode_ERROR_LOG_MANAGER_FAILED)
			resp.ErrName = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
			resp.Message = "Log Mnagaer internal error"
			return fmt.Errorf("Log Manager internal error")
		}
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
		resp.ErrCode = int32(proto.ErrCode_ERROR_OK)
		resp.ErrName = proto.ErrCode_ERROR_OK
		return nil
	}
}

// CreateLogCollectionTask CreateLogCollectionTask
func (l *LogManagerServerImpl) CreateLogCollectionTask(ctx context.Context, req *proto.CreateLogCollectionTaskReq, resp *proto.CollectionTaskCommonResp) error {
	blog.Infof("reqest: %+v", req)
	config := l.buildLogCollectionTaskConfig(req)
	if config == nil {
		err := fmt.Errorf("Error in CreateLogCollectionTask: no LogCollectionConfig specified")
		resp.ErrCode = int32(proto.ErrCode_ERROR_LOG_MANAGER_FAILED)
		resp.ErrName = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
		resp.Message = err.Error()
		return err
	}
	resultCh := make(chan *proto.CollectionTaskCommonResp)
	go func() {
		// send message to manager
		result := l.logManager.HandleAddLogCollectionTask(ctx, config)
		resultCh <- result
		close(resultCh)
	}()
	var result *proto.CollectionTaskCommonResp
	select {
	// cancel
	case <-ctx.Done():
		blog.Warnf("LogManagerServerImpl CreateLogCollectionTask canceled: %s", ctx.Err().Error())
		resp.ErrCode = int32(proto.ErrCode_ERROR_LOG_MANAGER_FAILED)
		resp.ErrName = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
		resp.Message = ctx.Err().Error()
		return ctx.Err()
	// timeout
	case <-time.After(timeout):
		blog.Warnf("LogManagerServerImpl CreateLogCollectionTask timeout")
		resp.ErrCode = int32(proto.ErrCode_ERROR_LOG_MANAGER_FAILED)
		resp.ErrName = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
		resp.Message = "LogManagerServerImpl CreateLogCollectionTask timeout"
		return fmt.Errorf("LogManagerServerImpl CreateLogCollectionTask timeout")
	// get result
	case result = <-resultCh:
		if result == nil {
			resp.ErrCode = int32(proto.ErrCode_ERROR_LOG_MANAGER_FAILED)
			resp.ErrName = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
			resp.Message = "Log Mnagaer internal error"
			return fmt.Errorf("Log Manager internal error")
		}
		*resp = *result
		if len(resp.ErrResult) != 0 {
			resp.ErrCode = int32(proto.ErrCode_ERROR_LOG_MANAGER_PARTIALLY_FAILED)
			resp.ErrName = proto.ErrCode_ERROR_LOG_MANAGER_PARTIALLY_FAILED
			resp.Message = "log collection task operation failed partially"
		} else {
			resp.ErrCode = int32(proto.ErrCode_ERROR_OK)
			resp.ErrName = proto.ErrCode_ERROR_OK
		}
		return nil
	}
}

// DeleteLogCollectionTask DeleteLogCollectionTask
func (l *LogManagerServerImpl) DeleteLogCollectionTask(ctx context.Context, req *proto.DeleteLogCollectionTaskReq, resp *proto.CollectionTaskCommonResp) error {
	blog.Infof("reqest: %+v", req)
	filter := &config.CollectionFilterConfig{
		ClusterIDs:      strings.ToLower(req.ClusterIDs),
		ConfigName:      req.ConfigName,
		ConfigNamespace: req.ConfigNamespace,
	}
	resultCh := make(chan *proto.CollectionTaskCommonResp)
	go func() {
		// send message to manager
		result := l.logManager.HandleDeleteLogCollectionTask(ctx, filter)
		resultCh <- result
		close(resultCh)
	}()
	var result *proto.CollectionTaskCommonResp
	select {
	// cancel
	case <-ctx.Done():
		blog.Warnf("LogManagerServerImpl DeleteLogCollectionTask canceled: %s", ctx.Err().Error())
		resp.ErrCode = int32(proto.ErrCode_ERROR_LOG_MANAGER_FAILED)
		resp.ErrName = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
		resp.Message = ctx.Err().Error()
		return ctx.Err()
	// timeout
	case <-time.After(timeout):
		blog.Warnf("LogManagerServerImpl DeleteLogCollectionTask timeout")
		resp.ErrCode = int32(proto.ErrCode_ERROR_LOG_MANAGER_FAILED)
		resp.ErrName = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
		resp.Message = "LogManagerServerImpl DeleteLogCollectionTask timeout"
		return fmt.Errorf("LogManagerServerImpl DeleteLogCollectionTask timeout")
	// get result
	case result = <-resultCh:
		if result == nil {
			resp.ErrCode = int32(proto.ErrCode_ERROR_LOG_MANAGER_FAILED)
			resp.ErrName = proto.ErrCode_ERROR_LOG_MANAGER_FAILED
			resp.Message = "Log Mnagaer internal error"
			return fmt.Errorf("Log Manager internal error")
		}
		*resp = *result
		if len(resp.ErrResult) != 0 {
			resp.ErrCode = int32(proto.ErrCode_ERROR_LOG_MANAGER_PARTIALLY_FAILED)
			resp.ErrName = proto.ErrCode_ERROR_LOG_MANAGER_PARTIALLY_FAILED
			resp.Message = "log collection task operation failed partially"
		} else {
			resp.ErrCode = int32(proto.ErrCode_ERROR_OK)
			resp.ErrName = proto.ErrCode_ERROR_OK
		}
		return nil
	}
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
			Selector: &proto.PodSelector{
				MatchLabels: conf.ConfigSpec.Selector.MatchLabels,
			},
		},
	}
	// resolve containerconf
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
	// resolve podselector match expression field
	matchExpressions := make([]*proto.SelectorExpression, len(conf.ConfigSpec.Selector.MatchExpressions))
	for i, v := range conf.ConfigSpec.Selector.MatchExpressions {
		matchExpressions[i] = &proto.SelectorExpression{
			Key:      v.Key,
			Operator: v.Operator,
		}
		var newValues []string
		if len(v.Values) != 0 {
			newValues = make([]string, len(v.Values))
			copy(newValues, v.Values)
			matchExpressions[i].Values = newValues
		}
	}
	ret.Config.Selector.MatchExpressions = matchExpressions
	return ret
}

// buildLogCollectionTaskConfig convert log config message from protobuf struct type to CollectionConfig
func (l *LogManagerServerImpl) buildLogCollectionTaskConfig(conf *proto.CreateLogCollectionTaskReq) *config.CollectionConfig {
	if conf.Config == nil || conf.Config.Config == nil {
		return nil
	}
	ret := &config.CollectionConfig{
		ClusterIDs:      strings.ToLower(conf.ClusterIDs),
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
	// resolve containerconf
	if conf.Config.Config.ContainerConfs != nil {
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
	}
	// resolve podselector match expression field
	if conf.Config.Config.Selector != nil {
		matchExpressions := make([]bcsv1.SelectorExpression, len(conf.Config.Config.Selector.MatchExpressions))
		for i, v := range conf.Config.Config.Selector.MatchExpressions {
			matchExpressions[i] = bcsv1.SelectorExpression{
				Key:      v.Key,
				Operator: v.Operator,
			}
			var newValues []string
			if len(v.Values) != 0 {
				newValues = make([]string, len(v.Values))
				copy(newValues, v.Values)
				matchExpressions[i].Values = newValues
			}
		}
		ret.ConfigSpec.Selector.MatchLabels = conf.Config.Config.Selector.MatchLabels
		ret.ConfigSpec.Selector.MatchExpressions = matchExpressions
	}
	return ret
}
