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
	"bk-bcs/bcs-common/common/blog"
	commtypes "bk-bcs/bcs-common/common/types"
	schedulertypes "bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"fmt"
	"time"
)

type DeploymentHandler struct {
	oper      DataOperator
	dataType  string
	ClusterID string
}

func (handler *DeploymentHandler) GetType() string {
	return handler.dataType
}

func (handler *DeploymentHandler) CheckDirty() error {

	blog.Info("check dirty data for type: %s", handler.dataType)

	conditionData := &commtypes.BcsStorageDynamicBatchDeleteIf{
		UpdateTimeBegin: 0,
		UpdateTimeEnd:   time.Now().Unix() - 600,
	}

	dataNode := fmt.Sprintf("/bcsstorage/v1/mesos/dynamic/all_resources/clusters/%s/%s",
		handler.ClusterID, handler.dataType)
	err := handler.oper.DeleteDCNodes(dataNode, conditionData, "DELETE")
	if err != nil {
		blog.Error("delete timeover node(%s) failed: %+v", dataNode, err)
		return err
	}

	return nil
}

func (handler *DeploymentHandler) Add(data interface{}) error {
	dataType := data.(*schedulertypes.Deployment)
	blog.Info("deployment add event, deployment: %s.%s", dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name)

	dataNode := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.ObjectMeta.NameSpace + "/" + handler.dataType + "/" + dataType.ObjectMeta.Name
	err := handler.oper.CreateDCNode(dataNode, data, "PUT")
	if err != nil {
		blog.V(3).Infof("deployment add node %s, err %+v", dataNode, err)
		return err
	}

	return nil
}

func (handler *DeploymentHandler) Delete(data interface{}) error {
	dataType := data.(*schedulertypes.Deployment)
	blog.Info("deployment delete event, deployment: %s.%s", dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name)

	dataNode := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.ObjectMeta.NameSpace + "/" + handler.dataType + "/" + dataType.ObjectMeta.Name
	err := handler.oper.DeleteDCNode(dataNode, "DELETE")
	if err != nil {
		blog.V(3).Infof("deployment delete node %s, err %+v", dataNode, err)
	}
	return err
}

func (handler *DeploymentHandler) Update(data interface{}) error {
	dataType := data.(*schedulertypes.Deployment)

	dataNode := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.ObjectMeta.NameSpace + "/" + handler.dataType + "/" + dataType.ObjectMeta.Name
	err := handler.oper.CreateDCNode(dataNode, data, "PUT")
	if err != nil {
		blog.V(3).Infof("deployment update node %s, err %+v", dataNode, err)
	}

	return err
}
