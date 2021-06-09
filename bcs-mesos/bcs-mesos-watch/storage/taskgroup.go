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

package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	schedtypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	containertypes "github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/container"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/util"
)

//TaskGroupHandler handle for taskgroup
type TaskGroupHandler struct {
	oper      DataOperator
	dataType  string
	ClusterID string

	DoCheckDirty bool
}

//GetType implementation
func (handler *TaskGroupHandler) GetType() string {
	return handler.dataType
}

//CheckDirty implementation
func (handler *TaskGroupHandler) CheckDirty() error {
	if handler.DoCheckDirty {
		blog.Info("check dirty data for type: %s", handler.dataType)
	} else {
		return nil
	}

	var (
		started       = time.Now()
		conditionData = &commtypes.BcsStorageDynamicBatchDeleteIf{
			UpdateTimeBegin: 0,
			UpdateTimeEnd:   time.Now().Unix() - 600,
		}
	)

	dataNode := fmt.Sprintf("/bcsstorage/v1/mesos/dynamic/all_resources/clusters/%s/%s",
		handler.ClusterID, handler.dataType)

	err := handler.oper.DeleteDCNodes(dataNode, conditionData, "DELETE")
	if err != nil {
		blog.Error("delete timeover node(%s) failed: %+v", dataNode, err)
		util.ReportStorageMetrics(handler.ClusterID, dataTypeTaskGroup, actionDelete, handlerAllClusterType, util.StatusFailure, started)
		return err
	}

	util.ReportStorageMetrics(handler.ClusterID, dataTypeTaskGroup, actionDelete, handlerAllClusterType, util.StatusSuccess, started)
	return nil
}

//Add handler to add
func (handler *TaskGroupHandler) Add(data interface{}) error {
	var (
		started  = time.Now()
		dataType = data.(*schedtypes.TaskGroup)
	)

	blog.Info("TaskGroup add event, ID: %s", dataType.ID)
	reportType, _ := handler.FormatConv(dataType)

	dataNode := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.RunAs + "/" + handler.dataType + "/" + dataType.ID

	err := handler.oper.CreateDCNode(dataNode, reportType, "PUT")
	if err != nil {
		blog.Error("TaskGroup add node(%s) failed: %+v", dataNode, err)
		util.ReportStorageMetrics(handler.ClusterID, dataTypeTaskGroup, actionPut, handlerClusterNamespaceTypeName, util.StatusFailure, started)
		return err
	}

	util.ReportStorageMetrics(handler.ClusterID, dataTypeTaskGroup, actionPut, handlerClusterNamespaceTypeName, util.StatusSuccess, started)
	return nil
}

//Delete delete info
func (handler *TaskGroupHandler) Delete(data interface{}) error {
	var (
		dataType = data.(*schedtypes.TaskGroup)
		started  = time.Now()
	)

	blog.Info("TaskGroup delete event, ID: %s", dataType.ID)
	dataNode := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.RunAs + "/" + handler.dataType + "/" + dataType.ID

	err := handler.oper.DeleteDCNode(dataNode, "DELETE")
	if err != nil {
		blog.Error("TaskGroup delete node(%s) failed: %+v", dataNode, err)
		util.ReportStorageMetrics(handler.ClusterID, dataTypeTaskGroup, actionDelete, handlerClusterNamespaceTypeName, util.StatusFailure, started)
		return err
	}

	util.ReportStorageMetrics(handler.ClusterID, dataTypeTaskGroup, actionDelete, handlerClusterNamespaceTypeName, util.StatusSuccess, started)
	return nil
}

//Update update in zookeeper
func (handler *TaskGroupHandler) Update(data interface{}) error {
	var (
		dataType = data.(*schedtypes.TaskGroup)
		started  = time.Now()
	)

	blog.V(3).Infof("TaskGroup update event, ID: %s", dataType.ID)
	reportType, _ := handler.FormatConv(dataType)

	dataNode := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.RunAs + "/" + handler.dataType + "/" + dataType.ID

	err := handler.oper.CreateDCNode(dataNode, reportType, "PUT")
	if err != nil {
		blog.Error("TaskGroup update node(%s) failed: %+v", dataNode, err)
		util.ReportStorageMetrics(handler.ClusterID, dataTypeTaskGroup, actionPut, handlerClusterNamespaceTypeName, util.StatusFailure, started)
		return err
	}

	util.ReportStorageMetrics(handler.ClusterID, dataTypeTaskGroup, actionPut, handlerClusterNamespaceTypeName, util.StatusSuccess, started)
	return nil
}

//FormatConv convert taskgroup to pod status for storage
func (handler *TaskGroupHandler) FormatConv(taskgroup *schedtypes.TaskGroup) (*commtypes.BcsPodStatus, error) {
	status := new(commtypes.BcsPodStatus)
	status.ObjectMeta = taskgroup.ObjectMeta
	status.ObjectMeta.Name = taskgroup.ID
	status.RcName = taskgroup.AppID
	status.Status = commtypes.PodStatus(taskgroup.Status)
	status.LastStatus = commtypes.PodStatus(taskgroup.LastStatus)
	status.HostName = taskgroup.HostName
	status.Message = taskgroup.Message
	status.KillPolicy = taskgroup.KillPolicy
	status.RestartPolicy = taskgroup.RestartPolicy
	status.StartTime = time.Unix(taskgroup.StartTime, 0)
	status.LastUpdateTime = time.Unix(taskgroup.UpdateTime, 0)
	status.ReportTime = time.Now()
	status.Kind = taskgroup.Kind
	by, _ := json.Marshal(taskgroup.BcsEventMsg)
	status.BcsMessage = string(by)

	for _, task := range taskgroup.Taskgroup {
		container := new(commtypes.BcsContainerStatus)
		container.RestartCount = int32(taskgroup.ReschededTimes)
		container.Status = commtypes.ContainerStatus(task.Status)
		container.LastStatus = commtypes.ContainerStatus(task.LastStatus)
		container.Image = task.Image
		container.LastUpdateTime = time.Unix(task.UpdateTime, 0)
		container.HealthCheckStatus = task.HealthCheckStatus
		container.Command = task.Command
		container.Args = task.Arguments
		container.Network = task.Network
		container.Labels = task.Labels
		status.ObjectMeta.Labels = task.Labels
		status.HostIP = task.AgentIPAddress

		bcsInfo := new(containertypes.BcsContainerInfo)
		if err := json.Unmarshal([]byte(task.StatusData), bcsInfo); err == nil {
			blog.V(3).Infof("get task(%s) data from Task.StatusData", task.ID)
			container.StartTime = bcsInfo.StartAt
			container.Name = bcsInfo.Name
			container.ContainerID = bcsInfo.ID
			container.Message = bcsInfo.Message
			container.FinishTime = bcsInfo.FinishAt

			/*if bcsInfo.NodeAddress != "" {
				status.HostIP = bcsInfo.NodeAddress
			}*/
			if bcsInfo.IPAddress != "" {
				status.PodIP = bcsInfo.IPAddress
			}
		}

		if container.Message == "" {
			container.Message = task.Message
		}

		container.Ports = make([]commtypes.ContainerPort, 0)

		for _, portMapping := range task.PortMappings {
			port := commtypes.ContainerPort{
				Name:          portMapping.Name,
				HostPort:      int(portMapping.HostPort),
				ContainerPort: int(portMapping.ContainerPort),
				Protocol:      portMapping.Protocol,
			}

			container.Ports = append(container.Ports, port)
		}

		container.Volumes = make([]commtypes.Volume, 0)

		for _, volume := range task.Volumes {
			v := commtypes.Volume{
				HostPath:  volume.HostPath,
				MountPath: volume.ContainerPath,
				ReadOnly:  true,
			}

			if volume.Mode == "RW" {
				v.ReadOnly = false
			}

			container.Volumes = append(container.Volumes, v)
		}

		if task.DataClass != nil && task.DataClass.Resources != nil {
			taskResource := task.DataClass.Resources
			if bcsInfo.Resource != nil && bcsInfo.Resource.Cpus > 0 {
				taskResource = bcsInfo.Resource
			}
			limits := commtypes.ResourceList{
				Cpu:     fmt.Sprintf("%f", taskResource.Cpus),
				Mem:     fmt.Sprintf("%f", taskResource.Mem),
				Storage: fmt.Sprintf("%f", taskResource.Disk),
			}
			container.Resources = commtypes.ResourceRequirements{
				Limits: limits,
			}
		}
		container.Env = make(map[string]string)
		for k, v := range task.Env {
			container.Env[k] = v
		}

		status.ContainerStatuses = append(status.ContainerStatuses, container)
	}

	blog.V(3).Infof("before post to CC, taskgroup format convert to: %+v", status)
	return status, nil
}
