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

package cloudvpc

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cidrmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ListAction action for list online clusterVPC
type ListAction struct {
	ctx          context.Context
	model        store.ClusterManagerModel
	req          *cmproto.ListCloudVPCRequest
	resp         *cmproto.ListCloudVPCResponse
	cloudVPCList []*cmproto.CloudVPCResp
}

// NewListAction create list action for cluster vpc list
func NewListAction(model store.ClusterManagerModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

func (la *ListAction) listCloudVPC() error {
	condM := make(operator.M)
	//! we don't setting bson tag in proto file
	//! all fields are in lowcase
	if len(la.req.CloudID) != 0 {
		condM["cloudid"] = la.req.CloudID
	}
	if len(la.req.Region) != 0 {
		condM["region"] = la.req.Region
	}
	if len(la.req.VpcID) != 0 {
		condM["vpcid"] = la.req.VpcID
	}
	if len(la.req.NetworkType) != 0 {
		condM["networktype"] = la.req.NetworkType
	}

	cond := operator.NewLeafCondition(operator.Eq, condM)
	cloudVPCs, err := la.model.ListCloudVPC(la.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}

	var (
		barrier = utils.NewRoutinePool(5)
		lock    = sync.Mutex{}
	)
	defer barrier.Close()

	for i := range cloudVPCs {
		if cloudVPCs[i].Available == "false" {
			continue
		}

		if la.req.BusinessID != "" && cloudVPCs[i].BusinessID != "" &&
			!strings.EqualFold(la.req.BusinessID, cloudVPCs[i].BusinessID) {
			continue
		}

		barrier.Add(1)
		// query available vpc
		go func(vpc cmproto.CloudVPC) {
			defer func() {
				barrier.Done()
			}()

			var surPlusIPNum uint32
			ipNum, err := getAvailableIPNumByVpc(utils.GlobalRouter.String(), vpc.Region, vpc.VpcID)
			if err != nil {
				blog.Errorf("listCloudVPC getAvailableIPNumByVpc failed: %v", err)
			} else {
				blog.Infof("region[%s] vpc[%s] availableIPNum[%v]", vpc.Region, vpc.VpcID, ipNum)
			}

			if ipNum <= vpc.ReservedIPNum {
				surPlusIPNum = 0
			} else {
				surPlusIPNum = ipNum - vpc.ReservedIPNum
			}

			cloud := &cmproto.CloudVPCResp{
				CloudID:        vpc.CloudID,
				Region:         vpc.Region,
				RegionName:     vpc.RegionName,
				NetworkType:    vpc.NetworkType,
				VpcID:          vpc.VpcID,
				VpcName:        vpc.VpcName,
				Available:      vpc.Available,
				Extra:          vpc.Extra,
				ReservedIPNum:  vpc.ReservedIPNum,
				AvailableIPNum: surPlusIPNum,
			}
			lock.Lock()
			la.cloudVPCList = append(la.cloudVPCList, cloud)
			lock.Unlock()

			return
		}(cloudVPCs[i])
	}
	barrier.Wait()

	// sort cloudVpcResp by available IP num
	sort.Sort(utils.CloudVpcSlice(la.cloudVPCList))

	return nil
}

func getAvailableIPNumByVpc(networkType, region, vpc string) (uint32, error) {
	cidrCli, conClose, err := cidrmanager.GetCidrClient().GetCidrManagerClient()
	if err != nil {
		return 0, err
	}
	defer func() {
		if conClose != nil {
			conClose()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	req := &cidrmanager.GetVPCIPSurplusRequest{
		Region:   region,
		CidrType: networkType,
		VpcID:    vpc,
	}
	resp, err := cidrCli.GetVPCIPSurplus(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.Data.IPSurplus, nil
}

func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.cloudVPCList
}

// Handle list cluster vpc list
func (la *ListAction) Handle(
	ctx context.Context, req *cmproto.ListCloudVPCRequest, resp *cmproto.ListCloudVPCResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list clusterVPC failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listCloudVPC(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

// ListRegionAction action for list cloud regions
type ListRegionAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	req     *cmproto.ListCloudRegionsRequest
	resp    *cmproto.ListCloudRegionsResponse
	regions []*cmproto.CloudRegion
}

// RegionData for regionInfo
type RegionData struct {
	Region     string
	RegionName string
}

// NewListRegionsAction create list action for cluster vpc list
func NewListRegionsAction(model store.ClusterManagerModel) *ListRegionAction {
	return &ListRegionAction{
		model: model,
	}
}

func (la *ListRegionAction) listCloudRegions() error {
	condM := make(operator.M)
	condM["cloudid"] = la.req.CloudID

	cond := operator.NewLeafCondition(operator.Eq, condM)
	cloudVPCs, err := la.model.ListCloudVPC(la.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	regions := make(map[string]*RegionData)
	for i := range cloudVPCs {
		if cloudVPCs[i].Available == "false" {
			continue
		}

		if _, ok := regions[cloudVPCs[i].Region]; !ok {
			regions[cloudVPCs[i].Region] = &RegionData{
				Region:     cloudVPCs[i].Region,
				RegionName: cloudVPCs[i].RegionName,
			}
		}
	}

	for _, data := range regions {
		la.regions = append(la.regions, &cmproto.CloudRegion{
			CloudID:    la.req.CloudID,
			RegionName: data.RegionName,
			Region:     data.Region,
		})
	}

	return nil
}

func (la *ListRegionAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.regions
}

// Handle list cloud regions
func (la *ListRegionAction) Handle(
	ctx context.Context, req *cmproto.ListCloudRegionsRequest, resp *cmproto.ListCloudRegionsResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list cloud regions failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listCloudRegions(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
