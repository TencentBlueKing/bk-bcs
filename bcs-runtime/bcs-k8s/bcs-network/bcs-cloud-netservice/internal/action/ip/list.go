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

package ip

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store/kube"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/utils"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/constant"
)

// ListAction action to list ips
type ListAction struct {
	req  *pb.ListIPsReq
	resp *pb.ListIPsResp

	ctx context.Context

	storeIf store.Interface

	ips []*types.IPObject
}

// NewListAction create list action
func NewListAction(ctx context.Context,
	req *pb.ListIPsReq, resp *pb.ListIPsResp,
	storeIf store.Interface) *ListAction {

	action := &ListAction{
		req:     req,
		resp:    resp,
		ctx:     ctx,
		storeIf: storeIf,
	}
	action.resp.Seq = req.Seq
	return action
}

// Err set err info
func (a *ListAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameters
func (a *ListAction) validate() error {
	if a.req.Offset < 0 || a.req.Limit < 0 {
		return fmt.Errorf("invalid offset or limit")
	}
	if a.req.Limit == 0 {
		a.req.Limit = constant.DefaultListLimit
	}
	return nil
}

// Input do something before Do function
func (a *ListAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *ListAction) Output() error {
	for _, ipObj := range a.ips {
		a.resp.Ips = append(a.resp.Ips, &pbcommon.IPObject{
			Address:      ipObj.Address,
			VpcID:        ipObj.VpcID,
			Region:       ipObj.Region,
			SubnetID:     ipObj.SubnetID,
			SubnetCidr:   ipObj.SubnetCidr,
			Cluster:      ipObj.Cluster,
			Namespace:    ipObj.Namespace,
			PodName:      ipObj.PodName,
			WorkloadName: ipObj.WorkloadName,
			WorkloadKind: ipObj.WorkloadKind,
			ContainerID:  ipObj.ContainerID,
			Host:         ipObj.Host,
			EniID:        ipObj.EniID,
			IsFixed:      ipObj.IsFixed,
			Status:       ipObj.Status,
		})
	}
	return nil
}

func (a *ListAction) listIPs() (pbcommon.ErrCode, string) {
	labelsMap := make(map[string]string)
	if len(a.req.VpcID) != 0 {
		labelsMap[kube.CrdNameLabelsVpcID] = a.req.VpcID
	}
	if len(a.req.Region) != 0 {
		labelsMap[kube.CrdNameLabelsRegion] = a.req.Region
	}
	if len(a.req.SubnetID) != 0 {
		labelsMap[kube.CrdNameLabelsSubnetID] = a.req.SubnetID
	}
	if len(a.req.Cluster) != 0 {
		labelsMap[kube.CrdNameLabelsCluster] = a.req.Cluster
	}
	if len(a.req.Namespace) != 0 {
		labelsMap[kube.CrdNameLabelsNamespace] = a.req.Namespace
	}
	if len(a.req.EniID) != 0 {
		labelsMap[kube.CrdNameLabelsEni] = a.req.EniID
	}
	if len(a.req.Host) != 0 {
		labelsMap[kube.CrdNameLabelsHost] = a.req.Host
	}
	if len(a.req.Status) != 0 {
		labelsMap[kube.CrdNameLabelsStatus] = a.req.Status
	}
	var ipsWithPodName []*types.IPObject
	var err error
	if len(a.req.PodName) != 0 {
		ipsWithPodName, err = a.storeIf.ListIPObjectByField(a.ctx, "spec.podName", utils.KeyToNamespacedKey(
			"bcs-system", a.req.PodName))
		if err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
		}
	}
	var retIPs []*types.IPObject
	tmpMap := make(map[string]*types.IPObject)
	for _, ip := range ipsWithPodName {
		tmpMap[ip.Address] = ip
	}
	ips, err := a.storeIf.ListIPObject(a.ctx, labelsMap)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	if len(a.req.PodName) != 0 {
		for _, ip := range ips {
			if _, ok := tmpMap[ip.Address]; ok {
				retIPs = append(retIPs, ip)
			}
		}
	} else {
		retIPs = ips
	}

	if a.req.Offset > int64(len(retIPs)) {
		return pbcommon.ErrCode_ERROR_OK, ""
	}
	if a.req.Offset+a.req.Limit >= int64(len(retIPs)) {
		a.ips = retIPs[a.req.Offset:]
	} else {
		a.ips = retIPs[a.req.Offset : a.req.Offset+a.req.Limit]
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do list action
func (a *ListAction) Do() error {
	if errCode, errMsg := a.listIPs(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
