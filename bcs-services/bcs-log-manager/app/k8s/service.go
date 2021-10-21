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

package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/api/proto/logmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/util"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/apis/bk-bcs/v1"
)

// HandleListLogCollectionTask deal with listing log collection task
func (m *LogManager) HandleListLogCollectionTask(ctx context.Context, filter *config.CollectionFilterConfig) map[string][]config.CollectionConfig {
	return m.getLogCollectionTaskByFilter(ctx, filter)
}

// HandleAddLogCollectionTask deal with adding log collection task
func (m *LogManager) HandleAddLogCollectionTask(ctx context.Context, conf *config.CollectionConfig) *proto.CollectionTaskCommonResp {
	return m.distributeAddTasks(ctx, m.getLogClients(), []config.CollectionConfig{*conf})
}

// HandleDeleteLogCollectionTask deal with deleting log collection task
func (m *LogManager) HandleDeleteLogCollectionTask(ctx context.Context, filter *config.CollectionFilterConfig) *proto.CollectionTaskCommonResp {
	return m.distributeDeleteTasks(ctx, filter)
}

// get bcslogconfigs from clusters
func (m *LogManager) getLogCollectionTaskByFilter(ctx context.Context, filter *config.CollectionFilterConfig) map[string][]config.CollectionConfig {
	var ret = make(map[string][]config.CollectionConfig)
	var wg sync.WaitGroup
	respCh := make(chan interface{}, 1)
	logClients := m.getLogClients()
	// get tasks from specified clusters
	if filter.ClusterIDs == "" {
		for _, ctl := range logClients {
			if ctx.Err() != nil {
				blog.Warnf("LogManager HandleListLogCollectionTask canceled: %s", ctx.Err().Error())
				break
			}
			wg.Add(1)
			go m.getTaskFromCluster(ctx, ctl, &wg, &RequestMessage{
				RespCh: respCh,
				Data:   filter,
			})
		}
	} else {
		clusters := strings.Split(filter.ClusterIDs, ",")
		for _, id := range clusters {
			if ctx.Err() != nil {
				blog.Warnf("LogManager HandleListLogCollectionTask canceled: %s", ctx.Err().Error())
				break
			}
			if client, ok := logClients[id]; !ok {
				blog.Warnf("No cluster id (%s)", id)
				continue
			} else {
				wg.Add(1)
				go m.getTaskFromCluster(ctx, client, &wg, &RequestMessage{
					RespCh: respCh,
					Data:   filter,
				})
			}
		}
	}
	// wait for get tasks finished
	go func() {
		wg.Wait()
		respCh <- "termination"
	}()
	// construct resp data
	for {
		select {
		case resp := <-respCh:
			switch data := resp.(type) {
			case string:
				if data == "termination" {
					close(respCh)
					return ret
				}
			case *[]config.CollectionConfig:
				if len(*data) > 0 {
					ret[(*data)[0].ClusterIDs] = *data
				}
			}
		}
	}
}

// distribute add task
func (m *LogManager) distributeAddTasks(ctx context.Context, newClusters map[string]*LogClient, confs []config.CollectionConfig) *proto.CollectionTaskCommonResp {
	blog.Infof("Start distribute log configs to clusters")
	blog.Infof("log config list: %+v", confs)
	var wg sync.WaitGroup
	ret := &proto.CollectionTaskCommonResp{
		ErrResult: make([]*proto.ClusterDimensionalResp, 0),
	}
	respCh := make(chan interface{}, 1)
	// add tasks to specified clusters
	for _, logconf := range confs {
		blog.Infof("distribute config : %+v", logconf)
		if logconf.ClusterIDs == "" {
			for _, client := range newClusters {
				if ctx.Err() != nil {
					blog.Warnf("LogManager HandleAddLogCollectionTask canceled: %s", ctx.Err().Error())
					goto distributeAddTasksExit
				}
				wg.Add(1)
				go m.addTaskToCluster(client, &wg, &RequestMessage{
					RespCh: respCh,
					Data:   logconf,
				})
				blog.Infof("Send logconf to cluster %s", client.ClusterInfo.ClusterID)
			}
			continue
		}
		clusters := strings.Split(strings.ToLower(logconf.ClusterIDs), ",")
		for _, clusterid := range clusters {
			if ctx.Err() != nil {
				blog.Warnf("LogManager HandleAddLogCollectionTask canceled: %s", ctx.Err().Error())
				goto distributeAddTasksExit
			}
			if _, ok := newClusters[clusterid]; !ok {
				blog.Errorf("Wrong cluster ID %s of collection config %+v", clusterid, logconf)
				ret.ErrResult = append(ret.ErrResult, &proto.ClusterDimensionalResp{
					ClusterID: clusterid,
					ErrCode:   int32(proto.ErrCode_ERROR_NO_SUCH_CLUSTER),
					ErrName:   proto.ErrCode_ERROR_NO_SUCH_CLUSTER,
					Message:   "No such cluster",
				})
				continue
			}
			client := newClusters[clusterid]
			wg.Add(1)
			go m.addTaskToCluster(client, &wg, &RequestMessage{
				RespCh: respCh,
				Data:   logconf,
			})
			blog.Infof("Send logconf to cluster %s", client.ClusterInfo.ClusterID)
		}
	}
distributeAddTasksExit:
	// wait for job finished
	go func() {
		wg.Wait()
		respCh <- "termination"
	}()
	// construct resp message
	for {
		select {
		case resp := <-respCh:
			switch data := resp.(type) {
			case string:
				if data == "termination" {
					close(respCh)
					return ret
				}
			case *proto.ClusterDimensionalResp:
				ret.ErrResult = append(ret.ErrResult, data)
			}
		}
	}
}

// distribute delete task
func (m *LogManager) distributeDeleteTasks(ctx context.Context, filter *config.CollectionFilterConfig) *proto.CollectionTaskCommonResp {
	ret := &proto.CollectionTaskCommonResp{
		ErrResult: make([]*proto.ClusterDimensionalResp, 0),
	}
	if filter.ClusterIDs == "" {
		ret.ErrCode = int32(proto.ErrCode_ERROR_CLUSTER_ID_REQUIRED)
		ret.ErrName = proto.ErrCode_ERROR_CLUSTER_ID_REQUIRED
		ret.Message = "Cluster ID is required in delete operation"
		return ret
	}
	var wg sync.WaitGroup
	respCh := make(chan interface{}, 1)
	logClients := m.getLogClients()
	clusters := strings.Split(filter.ClusterIDs, ",")
	// delete tasks from specified clusters
	for _, id := range clusters {
		if ctx.Err() != nil {
			blog.Warnf("LogManager HandleDeleteLogCollectionTask canceled: %s", ctx.Err().Error())
			break
		}
		if client, ok := logClients[id]; !ok {
			blog.Warnf("No cluster id (%s)", id)
			ret.ErrResult = append(ret.ErrResult, &proto.ClusterDimensionalResp{
				ClusterID: id,
				ErrCode:   int32(proto.ErrCode_ERROR_NO_SUCH_CLUSTER),
				ErrName:   proto.ErrCode_ERROR_NO_SUCH_CLUSTER,
				Message:   "No such cluster",
			})
			continue
		} else {
			wg.Add(1)
			go m.deleteTaskFromCluster(client, &wg, &RequestMessage{
				RespCh: respCh,
				Data:   filter,
			})
		}
	}
	// wait for jobs finishing
	go func() {
		wg.Wait()
		respCh <- "termination"
	}()
	// construct resp message
	for {
		select {
		case resp := <-respCh:
			switch data := resp.(type) {
			case string:
				if data == "termination" {
					close(respCh)
					return ret
				}
			case *proto.ClusterDimensionalResp:
				ret.ErrResult = append(ret.ErrResult, data)
			}
		}
	}
}

// msg.Data is *config.CollectionConfig, msg.RespCh is error return channel
func (m *LogManager) addTaskToCluster(client *LogClient, wg *sync.WaitGroup, msg *RequestMessage) {
	defer wg.Done()
	task, ok := msg.Data.(config.CollectionConfig)
	if !ok {
		blog.Errorf("addTaskToCluster convert msg.Data to *config.CollectionConfig failed. msg.Data: (%+v)", msg.Data)
		msg.RespCh <- &proto.ClusterDimensionalResp{
			ClusterID: client.ClusterInfo.ClusterID,
			ErrCode:   int32(proto.ErrCode_ERROR_LOG_MANAGER_INTERNAL_ERROR),
			ErrName:   proto.ErrCode_ERROR_LOG_MANAGER_INTERNAL_ERROR,
			Message:   "log manager internal error",
		}
		return
	}
	// construct BcsLogConfig
	logconf := &bcsv1.BcsLogConfig{}
	logconf.TypeMeta.Kind = LogConfigKind
	logconf.TypeMeta.APIVersion = LogConfigAPIVersion
	if task.ConfigName == "" {
		task.ConfigName = fmt.Sprintf("%s-%s-%d", LogConfigKind, client.ClusterInfo.ClusterID, util.GenerateID())
	}
	logconf.ObjectMeta.Name = task.ConfigName
	logconf.SetName(task.ConfigName)
	if task.ConfigNamespace == "" {
		task.ConfigNamespace = DefaultLogConfigNamespace
	}
	logconf.ObjectMeta.Namespace = task.ConfigNamespace
	task.ConfigSpec.ClusterId = client.ClusterInfo.ClusterID
	logconf.Spec = task.ConfigSpec
	// rest request
	err := client.Post().
		Resource("bcslogconfigs").
		Namespace(task.ConfigNamespace).
		Body(logconf).
		Do().
		Error()
	if err != nil {
		blog.Warnf("Create BcsLogConfig of Cluster %s failed: %s (config info: %+v)", client.ClusterInfo.ClusterID, err.Error(), logconf)
		msg.RespCh <- &proto.ClusterDimensionalResp{
			ClusterID: client.ClusterInfo.ClusterID,
			ErrCode:   int32(proto.ErrCode_ERROR_CLUSTER_OPERATION_ERROR),
			ErrName:   proto.ErrCode_ERROR_CLUSTER_OPERATION_ERROR,
			Message:   err.Error(),
		}
		return
	}
	blog.Infof("Create BcsLogConfig of Cluster %s success. (config info: %+v)", client.ClusterInfo.ClusterID, logconf)
}

// msg.Data is *config.CollectionFilterConfig, msg.RespCh is error return channel
func (m *LogManager) getTaskFromCluster(ctx context.Context, client *LogClient, wg *sync.WaitGroup, msg *RequestMessage) {
	defer wg.Done()
	filter, ok := msg.Data.(*config.CollectionFilterConfig)
	// TODO error return
	if !ok {
		blog.Errorf("getTaskFromCluster convert msg.Data to *config.CollectionFilterConfig failed. msg.Data: (%+v)", msg.Data)
		msg.RespCh <- client.ClusterInfo.ClusterID
		return
	}
	// rest request
	req := client.Get().Resource("bcslogconfigs").Namespace(filter.ConfigNamespace)
	if filter.ConfigName != "" {
		req = req.Name(filter.ConfigName)
	}
	result := req.Do()
	if result.Error() != nil {
		blog.Errorf("Get BcsLogConfig from Cluster %s failed: %s", client.ClusterInfo.ClusterID, result.Error().Error())
		msg.RespCh <- client.ClusterInfo.ClusterID
		return
	}
	raw, err := result.Raw()
	if err != nil {
		blog.Errorf("Get raw data from Cluster %s response failed: %s", client.ClusterInfo.ClusterID, err.Error())
		msg.RespCh <- client.ClusterInfo.ClusterID
		return
	}
	// parse result to BcsLogConfig slice
	var respSlice []bcsv1.BcsLogConfig
	if filter.ConfigName != "" {
		var conf bcsv1.BcsLogConfig
		err = json.Unmarshal(raw, &conf)
		if err != nil {
			blog.Errorf("Convert raw data to BcsLogConfig failed: %s, raw(%s), Cluster(%s)",
				client.ClusterInfo.ClusterID, err.Error(), string(raw), client.ClusterInfo.ClusterID)
			msg.RespCh <- client.ClusterInfo.ClusterID
			return
		}
		respSlice = append(respSlice, conf)
	} else {
		var conf bcsv1.BcsLogConfigList
		err = json.Unmarshal(raw, &conf)
		if err != nil {
			blog.Errorf("Convert raw data to BcsLogConfigList failed: %s, raw(%s), Cluster(%s)",
				client.ClusterInfo.ClusterID, err.Error(), string(raw), client.ClusterInfo.ClusterID)
			msg.RespCh <- client.ClusterInfo.ClusterID
			return
		}
		respSlice = conf.Items
	}
	msg.RespCh <- m.convert(respSlice)
	blog.Infof("Get BcsLogConfig from Cluster %s success.", client.ClusterInfo.ClusterID)
}

func (m *LogManager) deleteTaskFromCluster(client *LogClient, wg *sync.WaitGroup, msg *RequestMessage) {
	defer wg.Done()
	filter, ok := msg.Data.(*config.CollectionFilterConfig)
	// return error
	if !ok {
		blog.Errorf("getTaskFromCluster convert msg.Data to *config.CollectionFilterConfig failed. msg.Data: (%+v)", msg.Data)
		msg.RespCh <- &proto.ClusterDimensionalResp{
			ClusterID: client.ClusterInfo.ClusterID,
			ErrCode:   int32(proto.ErrCode_ERROR_LOG_MANAGER_INTERNAL_ERROR),
			ErrName:   proto.ErrCode_ERROR_LOG_MANAGER_INTERNAL_ERROR,
			Message:   "log manager internal error",
		}
		return
	}
	// rest request
	err := client.Delete().
		Resource("bcslogconfigs").
		Namespace(filter.ConfigNamespace).
		Name(filter.ConfigName).
		Do().
		Error()
	if err != nil {
		blog.Warnf("Delete BcsLogConfig(%s/%s) of Cluster %s failed: %s",
			filter.ConfigNamespace, filter.ConfigName, client.ClusterInfo.ClusterID, err.Error())
		msg.RespCh <- &proto.ClusterDimensionalResp{
			ClusterID: client.ClusterInfo.ClusterID,
			ErrCode:   int32(proto.ErrCode_ERROR_CLUSTER_OPERATION_ERROR),
			ErrName:   proto.ErrCode_ERROR_CLUSTER_OPERATION_ERROR,
			Message:   err.Error(),
		}
		return
	}
	blog.Infof("Delete BcsLogConfig(%s/%s) from Cluster %s success.", filter.ConfigNamespace, filter.ConfigName, client.ClusterInfo.ClusterID)
}

// convert converts BcsLogConfig CRD to CollectionConfig representation
func (m *LogManager) convert(in []bcsv1.BcsLogConfig) *[]config.CollectionConfig {
	ret := make([]config.CollectionConfig, len(in))
	for i := range in {
		ret[i].ClusterIDs = in[i].Spec.ClusterId
		ret[i].ConfigName = in[i].GetName()
		ret[i].ConfigNamespace = in[i].GetNamespace()
		ret[i].ConfigSpec = *in[i].Spec.DeepCopy()
	}
	return &ret
}
