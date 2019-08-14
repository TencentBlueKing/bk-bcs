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

package resthdrs

import (
	"fmt"
	"sync"

	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	m "bk-bcs/bcs-services/bcs-api/pkg/models"
	"bk-bcs/bcs-services/bcs-api/pkg/server/external-cluster/tke"
	"bk-bcs/bcs-services/bcs-api/pkg/server/resthdrs/filters"
	"bk-bcs/bcs-services/bcs-api/pkg/server/resthdrs/utils"
	"bk-bcs/bcs-services/bcs-api/pkg/server/types"
	"bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
	"github.com/emicklei/go-restful"
)

var mutex sync.Mutex

type UpdateTkeLbForm struct {
	ClusterRegion string `json:"cluster_region" validate:"required"`
	SubnetId      string `json:"subnet_id" validate:"required"`
}

type AddTkeCidrForm struct {
	TkeCidrs []TkeCidr `json:"tke_cidrs" validate:"required"`
}

type ApplyTkeCidrForm struct {
	Cluster  string `json:"cluster" validate:"required"`
	IpNumber uint   `json:"ip_number" validate:"required"`
}

type ReleaseTkeCidrForm struct {
	Cidr    string `json:"cidr" validate:"required"`
	Cluster string `json:"cluster" validate:"required"`
}

type TkeCidr struct {
	Cidr     string `json:"cidr" validate:"required"`
	IpNumber uint   `json:"ip_number" validate:"required"`
	Status   string `json:"status"`
}

type LbStatus struct {
	ClusterId string `json:"cluster_id"`
	Status    string `json:"status"`
}

func BindLb(request *restful.Request, response *restful.Response) {
	cluster := filters.GetCluster(request)

	externalClusterInfo := sqlstore.QueryBCSClusterInfo(&m.BCSClusterInfo{
		ClusterId: cluster.ID,
	})
	if externalClusterInfo == nil {
		message := fmt.Sprintf("errcode: %d, external cluster info not exists", common.BcsErrApiBadRequest)
		WriteClientError(response, "EXTERNAL_CLUSTER_NOT_EXISTS", message)
		return
	}
	if externalClusterInfo.ClusterType != utils.BcsTkeCluster {
		message := fmt.Sprintf("errcode: %d, cluster %s is not tke cluster", common.BcsErrApiBadRequest, cluster.ID)
		WriteClientError(response, "NOT_TKE_CLUSTER", message)
		return
	}

	tkeCluster := tke.NewTkeCluster(cluster.ID, externalClusterInfo.TkeClusterId, externalClusterInfo.TkeClusterRegion)
	err := tkeCluster.BindClusterLb()
	if err != nil {
		message := err.Error()
		WriteServerError(response, "CANNOT_BIND_TKE_CLUSTER_LB", message)
		return
	}

	response.WriteEntity(types.EmptyResponse{})
}

func GetLbStatus(request *restful.Request, response *restful.Response) {
	cluster := filters.GetCluster(request)

	externalClusterInfo := sqlstore.QueryBCSClusterInfo(&m.BCSClusterInfo{
		ClusterId: cluster.ID,
	})
	if externalClusterInfo == nil {
		message := fmt.Sprintf("errcode: %d, external cluster info not exists", common.BcsErrApiBadRequest)
		WriteClientError(response, "EXTERNAL_CLUSTER_NOT_EXISTS", message)
		return
	}
	if externalClusterInfo.ClusterType != utils.BcsTkeCluster {
		message := fmt.Sprintf("errcode: %d, cluster %s is not tke cluster", common.BcsErrApiBadRequest, cluster.ID)
		WriteClientError(response, "NOT_TKE_CLUSTER", message)
		return
	}

	tkeCluster := tke.NewTkeCluster(cluster.ID, externalClusterInfo.TkeClusterId, externalClusterInfo.TkeClusterRegion)
	status, err := tkeCluster.GetMasterVip()
	if err != nil {
		message := err.Error()
		WriteServerError(response, "GET_TKE_MASTER_VIP_FAILED", message)
		return
	}

	lbStatus := &LbStatus{
		ClusterId: cluster.ID,
		Status:    status,
	}

	response.WriteEntity(*lbStatus)
}

func UpdateTkeLbSubnet(request *restful.Request, response *restful.Response) {
	blog.Debug(fmt.Sprintf("Create or Update tke lb subnet"))
	form := UpdateTkeLbForm{}
	request.ReadEntity(&form)

	err := validate.Struct(&form)
	if err != nil {
		response.WriteEntity(FormatValidationError(err))
		return
	}

	err = sqlstore.SaveTkeLbSubnet(form.ClusterRegion, form.SubnetId)
	if err != nil {
		message := fmt.Sprintf("errcode: %d, can not update tke lb subnet, error: %s", common.BcsErrApiInternalDbError, err.Error())
		WriteClientError(response, "CANNOT_UPDATE_TKE_LB_SUBNET", message)
		return
	}

	response.WriteEntity(types.EmptyResponse{})
}

// AddTkeCidr init tke cidrs into bcs-api
func AddTkeCidr(request *restful.Request, response *restful.Response) {
	blog.Info(fmt.Sprintf("Insert cidr"))
	form := AddTkeCidrForm{}
	request.ReadEntity(&form)

	err := validate.Struct(&form)
	if err != nil {
		response.WriteEntity(FormatValidationError(err))
		return
	}

	for _, tkeCidr := range form.TkeCidrs {
		cidr := sqlstore.QueryTkeCidr(&m.TkeCidr{
			Cidr:     tkeCidr.Cidr,
			IpNumber: tkeCidr.IpNumber,
		})
		if cidr != nil {
			blog.Warnf("Add Cidr failed, Cidr %s IpNumber %d already exists", tkeCidr.Cidr, tkeCidr.IpNumber)
			continue
		}
		if tkeCidr.Status == "" {
			tkeCidr.Status = sqlstore.CidrStatusAvailable
		}
		err = sqlstore.SaveTkeCidr(tkeCidr.Cidr, tkeCidr.IpNumber, tkeCidr.Status, "")
		if err != nil {
			blog.Errorf("add tke cidr failed, error: %s", err.Error())
			message := fmt.Sprintf("errcode: %d, add tke cidr failed, error: %s", common.BcsErrApiInternalDbError, err.Error())
			WriteClientError(response, "ADD_TKE_CIDR_FAILED", message)
			return
		}
	}

	response.WriteEntity(types.EmptyResponse{})
}

// ApplyTkeCidr assign an cidr to client
func ApplyTkeCidr(request *restful.Request, response *restful.Response) {
	form := ApplyTkeCidrForm{}
	request.ReadEntity(&form)

	err := validate.Struct(&form)
	if err != nil {
		response.WriteEntity(FormatValidationError(err))
		return
	}

	mutex.Lock()
	tkeCidr := sqlstore.QueryTkeCidr(&m.TkeCidr{
		IpNumber: form.IpNumber,
		Status:   sqlstore.CidrStatusAvailable,
	})
	if tkeCidr == nil {
		blog.Warnf("Apply cidr ipNumber %d for cluster %s failed, no available cidr any more", form.IpNumber, form.Cluster)
		message := fmt.Sprintf("errcode: %d, apply cidr failed, no available cidr any more", common.BcsErrApiInternalDbError)
		WriteClientError(response, "NO_AVAILABLE_CIDR", message)
		return
	}

	updatedTkeCidr := tkeCidr
	updatedTkeCidr.Status = sqlstore.CidrStatusUsed
	updatedTkeCidr.Cluster = &form.Cluster
	err = sqlstore.UpdateTkeCidr(tkeCidr, updatedTkeCidr)
	if err != nil {
		blog.Errorf("apply tkeCidr failed, cidr %s, error: %s", tkeCidr.Cidr, err.Error())
		message := fmt.Sprintf("errcode: %d, apply tkeCidr failed: %s", common.BcsErrApiInternalDbError, err.Error())
		WriteClientError(response, "APPLY_TKE_CIDR_FAILED", message)
		return
	}
	mutex.Unlock()

	blog.Info("assign an cidr successful, cidr: %s, ipNumber: %d", tkeCidr.Cidr, tkeCidr.IpNumber)
	cidr := &TkeCidr{
		Cidr:     tkeCidr.Cidr,
		IpNumber: tkeCidr.IpNumber,
		Status:   sqlstore.CidrStatusUsed,
	}
	response.WriteEntity(cidr)
}

// ReleaseTkeCidr release a cidr to be available
func ReleaseTkeCidr(request *restful.Request, response *restful.Response) {
	form := ReleaseTkeCidrForm{}
	request.ReadEntity(&form)

	err := validate.Struct(&form)
	if err != nil {
		response.WriteEntity(FormatValidationError(err))
		return
	}

	mutex.Lock()
	tkeCidr := sqlstore.QueryTkeCidr(&m.TkeCidr{
		Cidr:    form.Cidr,
		Cluster: &form.Cluster,
	})
	if tkeCidr == nil || tkeCidr.Status == sqlstore.CidrStatusAvailable {
		blog.Warnf("Release cidr %s failed, no such cidr to be released", form.Cidr)
		message := fmt.Sprintf("errcode: %d, no such cidr to be released", common.BcsErrApiBadRequest)
		WriteClientError(response, "NO_SUCH_CIDR", message)
		return
	}

	updatedTkeCidr := tkeCidr
	updatedTkeCidr.Status = sqlstore.CidrStatusAvailable
	cluster := ""
	updatedTkeCidr.Cluster = &cluster
	err = sqlstore.UpdateTkeCidr(tkeCidr, updatedTkeCidr)
	if err != nil {
		blog.Errorf("release tkeCidr failed, cidr %s, error: %s", tkeCidr.Cidr, err.Error())
		message := fmt.Sprintf("errcode: %d, release tkeCidr failed: %s", common.BcsErrApiInternalDbError, err.Error())
		WriteClientError(response, "RELEASE_TKE_CIDR_FAILED", message)
		return
	}
	mutex.Unlock()

	blog.Info("release cidr successful, cidr: %s, ipNumber: %d, cluster: %s", tkeCidr.Cidr, tkeCidr.IpNumber, tkeCidr.Cluster)
	response.WriteEntity(types.EmptyResponse{})
}
