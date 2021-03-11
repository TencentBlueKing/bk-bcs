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

//ConfigMapHandler handler for configmap synchronize to bcs-storage
type ConfigMapHandler struct {
	oper      DataOperator
	dataType  string
	ClusterID string
}

//GetType get datatype
func (handler *ConfigMapHandler) GetType() string {
	return handler.dataType
}

//CheckDirty clean remote storage dirty data
func (handler *ConfigMapHandler) CheckDirty() error {
	started := time.Now()
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
		util.ReportStorageMetrics(handler.ClusterID, dataTypeCfg, actionDelete, handlerAllClusterType, util.StatusFailure, started)
		return err
	}

	util.ReportStorageMetrics(handler.ClusterID, dataTypeCfg, actionDelete, handlerAllClusterType, util.StatusSuccess, started)
	return nil
}

//Add add new configmap to bcs-storage
func (handler *ConfigMapHandler) Add(data interface{}) error {
	var (
		dataType = data.(*commtypes.BcsConfigMap)
		started  = time.Now()
	)

	blog.Info("configmap add event, configmap: %s.%s", dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name)
	dataNode := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.ObjectMeta.NameSpace + "/" + handler.dataType + "/" + dataType.ObjectMeta.Name //nolint

	err := handler.oper.CreateDCNode(dataNode, data, "PUT")
	if err != nil {
		blog.V(3).Infof("configmap add node %s, err %+v", dataNode, err)
		util.ReportStorageMetrics(handler.ClusterID, dataTypeCfg, actionPut, handlerClusterNamespaceTypeName, util.StatusFailure, started)
		return err
	}

	util.ReportStorageMetrics(handler.ClusterID, dataTypeCfg, actionPut, handlerClusterNamespaceTypeName, util.StatusSuccess, started)
	return nil
}

//Delete delete configmap in storage
func (handler *ConfigMapHandler) Delete(data interface{}) error {
	var (
		dataType = data.(*commtypes.BcsConfigMap)
		started  = time.Now()
	)

	blog.Info("configmap delete event, configmap: %s.%s", dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name)
	dataNode := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.ObjectMeta.NameSpace + "/" + handler.dataType + "/" + dataType.ObjectMeta.Name

	err := handler.oper.DeleteDCNode(dataNode, "DELETE")
	if err != nil {
		blog.V(3).Infof("configmap delete node %s, err %+v", dataNode, err)
		util.ReportStorageMetrics(handler.ClusterID, dataTypeCfg, actionDelete, handlerClusterNamespaceTypeName, util.StatusFailure, started)
		return err
	}

	util.ReportStorageMetrics(handler.ClusterID, dataTypeCfg, actionDelete, handlerClusterNamespaceTypeName, util.StatusSuccess, started)
	return nil
}

//Update update configmap in storage
func (handler *ConfigMapHandler) Update(data interface{}) error {
	var (
		dataType = data.(*commtypes.BcsConfigMap)
		started  = time.Now()
	)

	dataNode := "/bcsstorage/v1/mesos/dynamic/namespace_resources/clusters/" + handler.ClusterID + "/namespaces/" + dataType.ObjectMeta.NameSpace + "/" + handler.dataType + "/" + dataType.ObjectMeta.Name

	err := handler.oper.CreateDCNode(dataNode, data, "PUT")
	if err != nil {
		blog.V(3).Infof("configmap update node %s, err %+v", dataNode, err)
		util.ReportStorageMetrics(handler.ClusterID, dataTypeCfg, actionPut, handlerClusterNamespaceTypeName, util.StatusFailure, started)
		return err
	}

	util.ReportStorageMetrics(handler.ClusterID, dataTypeCfg, actionPut, handlerClusterNamespaceTypeName, util.StatusSuccess, started)
	return nil
}
