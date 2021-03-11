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
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/util"
)

//ServiceHandler service event handler
type ServiceHandler struct {
	oper      DataOperator
	dataType  string
	ClusterID string
}

//GetType implementation
func (handler *ServiceHandler) GetType() string {
	return handler.dataType
}

//CheckDirty implementation
func (handler *ServiceHandler) CheckDirty() error {
	var (
		started       = time.Now()
		conditionData = &commtypes.BcsStorageDynamicBatchDeleteIf{
			UpdateTimeBegin: 0,
			UpdateTimeEnd:   time.Now().Unix() - 600,
		}
	)

	blog.Info("check dirty data for type: %s", handler.dataType)
	dataNode := fmt.Sprintf("/bcsstorage/v1/mesos/dynamic/all_resources/clusters/%s/%s",
		handler.ClusterID, handler.dataType)

	err := handler.oper.DeleteDCNodes(dataNode, conditionData, "DELETE")
	if err != nil {
		blog.Error("delete timeover node(%s) failed: %+v", dataNode, err)
		util.ReportStorageMetrics(handler.ClusterID, dataTypeSvr, actionDelete, handlerAllClusterType, util.StatusFailure, started)
		return err
	}

	util.ReportStorageMetrics(handler.ClusterID, dataTypeSvr, actionDelete, handlerAllClusterType, util.StatusSuccess, started)
	return nil
}

//Add event implementation
func (handler *ServiceHandler) Add(data interface{}) error {
	var (
		dataType = data.(*commtypes.BcsService)
		started  = time.Now()
	)

	blog.Info("service add event, service: %s.%s", dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name)
	dataNode := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.ObjectMeta.NameSpace + "/" + handler.dataType + "/" + dataType.ObjectMeta.Name //nolint

	err := handler.oper.CreateDCNode(dataNode, data, "PUT")
	if err != nil {
		blog.Errorf("service add node %s, err %+v", dataNode, err)
		util.ReportStorageMetrics(handler.ClusterID, dataTypeSvr, actionPut, handlerClusterNamespaceTypeName, util.StatusFailure, started)
		return err
	}

	util.ReportStorageMetrics(handler.ClusterID, dataTypeSvr, actionPut, handlerClusterNamespaceTypeName, util.StatusSuccess, started)
	return nil
}

//Delete event implementation
func (handler *ServiceHandler) Delete(data interface{}) error {
	var (
		dataType = data.(*commtypes.BcsService)
		started  = time.Now()
	)

	blog.Info("service delete event, service: %s.%s", dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name)
	dataNode := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.ObjectMeta.NameSpace + "/" + handler.dataType + "/" + dataType.ObjectMeta.Name //nolint

	err := handler.oper.DeleteDCNode(dataNode, "DELETE")
	if err != nil {
		blog.Errorf("service delete node %s, err %+v", dataNode, err)
		util.ReportStorageMetrics(handler.ClusterID, dataTypeSvr, actionDelete, handlerClusterNamespaceTypeName, util.StatusFailure, started)
		return err
	}

	util.ReportStorageMetrics(handler.ClusterID, dataTypeSvr, actionDelete, handlerClusterNamespaceTypeName, util.StatusSuccess, started)
	return nil
}

//Update event implementation
func (handler *ServiceHandler) Update(data interface{}) error {
	dataType := data.(*commtypes.BcsService)
	started := time.Now()
	dataNode := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.ObjectMeta.NameSpace + "/" + handler.dataType + "/" + dataType.ObjectMeta.Name //nolint

	err := handler.oper.CreateDCNode(dataNode, data, "PUT")
	if err != nil {
		blog.Errorf("service update node %s, err %+v", dataNode, err)
		util.ReportStorageMetrics(handler.ClusterID, dataTypeSvr, actionPut, handlerClusterNamespaceTypeName, util.StatusFailure, started)
		return err
	}

	util.ReportStorageMetrics(handler.ClusterID, dataTypeSvr, actionPut, handlerClusterNamespaceTypeName, util.StatusSuccess, started)
	return nil
}
