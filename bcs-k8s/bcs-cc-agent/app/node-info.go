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

package app

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	IdcId       = "bcs_idc_id"
	IdcName     = "bcs_idc_name"
	IdcUnitId   = "bcs_idc_unit_id"
	IdcUnitName = "bcs_idc_unit_name"
	Rack        = "bcs_rack"
	SvrTypeName = "bcs_svr_type_name"
	CvmRegion   = "bcs_cvm_region"
	CvmZone     = "bcs_cvm_zone"

	LabelOfIdcId       = "bkbcs.tencent.com/idc-id"
	LabelOfIdcName     = "bkbcs.tencent.com/idc-name"
	LabelOfIdcUnitId   = "bkbcs.tencent.com/idc-unit-id"
	LabelOfIdcUnitName = "bkbcs.tencent.com/idc-unit-name"
	LabelOfRack        = "bkbcs.tencent.com/rack"
	LabelOfSvrTypeName = "bkbcs.tencent.com/svr-type-name"
)

func updateK8sNodeInfo(clientset *kubernetes.Clientset, nodeName string, nodeInfo *NodeInfo) error {
	err := updateNodeInfoToFile(nodeInfo)
	if err != nil {
		return err
	}

	node, err := clientset.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("error get node from k8s: %s", err.Error())
		return err
	}

	node.Labels[LabelOfIdcId] = strconv.Itoa(nodeInfo.IdcId)
	node.Labels[LabelOfIdcUnitId] = strconv.Itoa(nodeInfo.IdcUnitId)
	node.Labels[LabelOfRack] = nodeInfo.Rack
	node.Labels[LabelOfSvrTypeName] = nodeInfo.SvrTypeName

	_, err = clientset.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
	if err != nil {
		blog.Errorf("error update node label: %s", err.Error())
		return err
	}
	blog.Info("succeed to update node label to k8s")

	return nil
}

func updateMesosNodeInfo(nodeInfo *NodeInfo) error {
	return updateNodeInfoToFile(nodeInfo)
}

func updateNodeInfoToFile(nodeInfo *NodeInfo) error {
	file := "/data/bcs/nodeinfo/node-info-env"
	bash := "#!/bin/bash\n"
	idcIdLine := "export " + IdcId + "=" + strconv.Itoa(nodeInfo.IdcId)
	idcNameLine := "export " + IdcName + "=" + nodeInfo.IdcName
	idcUnitIdLine := "export " + IdcUnitId + "=" + strconv.Itoa(nodeInfo.IdcUnitId)
	idcUnitNameLine := "export " + IdcUnitName + "=" + nodeInfo.IdcUnitName
	rackLine := "export " + Rack + "=" + nodeInfo.Rack
	svrTypeNameLine := "export " + SvrTypeName + "=" + nodeInfo.SvrTypeName

	infoSlice := []string{
		bash,
		idcIdLine,
		idcNameLine,
		idcUnitIdLine,
		idcUnitNameLine,
		rackLine,
		svrTypeNameLine,
	}
	if nodeInfo.CvmRegion != "" {
		region := "export " + CvmRegion + "=" + nodeInfo.CvmRegion
		infoSlice = append(infoSlice, region)
	}
	if nodeInfo.CvmZone != "" {
		zone := "export" + CvmZone + "=" + nodeInfo.CvmZone
		infoSlice = append(infoSlice, zone)
	}

	content := strings.Join(infoSlice, "\n")

	err := ioutil.WriteFile(file, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("error writing node info to file: %s", err.Error())
	}
	blog.Info("succeed to update node info to file")

	return nil
}
