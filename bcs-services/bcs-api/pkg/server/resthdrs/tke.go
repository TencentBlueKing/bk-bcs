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

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/metric"
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/server/external-cluster/tke"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/server/resthdrs/filters"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/server/resthdrs/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/server/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
	"github.com/emicklei/go-restful"
	"time"
)

var mutex sync.Mutex

type UpdateTkeLbForm struct {
	ClusterRegion string `json:"cluster_region" validate:"required"`
	SubnetId      string `json:"subnet_id" validate:"required"`
}

type AddTkeCidrForm struct {
	Vpc      string    `json:"vpc" validate:"required"`
	TkeCidrs []TkeCidr `json:"tke_cidrs" validate:"required"`
}

type ApplyTkeCidrForm struct {
	Vpc      string `json:"vpc" validate:"required"`
	Cluster  string `json:"cluster" validate:"required"`
	IpNumber uint   `json:"ip_number" validate:"required"`
}

type ReleaseTkeCidrForm struct {
	Vpc     string `json:"vpc" validate:"required"`
	Cidr    string `json:"cidr" validate:"required"`
	Cluster string `json:"cluster" validate:"required"`
}

type TkeCidr struct {
	Cidr     string `json:"cidr" validate:"required"`
	IpNumber uint   `json:"ip_number" validate:"required"`
	Status   string `json:"status"`
}

type ApplyTkeCidrResult struct {
	Vpc      string `json:"vpc" validate:"required"`
	Cidr     string `json:"cidr" validate:"required"`
	IpNumber uint   `json:"ip_number" validate:"required"`
	Status   string `json:"status"`
}

type LbStatus struct {
	ClusterId string `json:"cluster_id"`
	Status    string `json:"status"`
}

// BindLb bind a clb to tke cluster
func BindLb(request *restful.Request, response *restful.Response) {

	// support prometheus metrics
	start := time.Now()

	cluster := filters.GetCluster(request)

	// check if the cluster is valid
	externalClusterInfo := sqlstore.QueryBCSClusterInfo(&m.BCSClusterInfo{
		ClusterId: cluster.ID,
	})
	if externalClusterInfo == nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, external cluster info not exists", common.BcsErrApiBadRequest)
		WriteClientError(response, "EXTERNAL_CLUSTER_NOT_EXISTS", message)
		return
	}
	if externalClusterInfo.ClusterType != utils.BcsTkeCluster {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, cluster %s is not tke cluster", common.BcsErrApiBadRequest, cluster.ID)
		WriteClientError(response, "NOT_TKE_CLUSTER", message)
		return
	}

	// call tke api to bind cluster clb
	tkeCluster := tke.NewTkeCluster(cluster.ID, externalClusterInfo.TkeClusterId, externalClusterInfo.TkeClusterRegion)
	err := tkeCluster.BindClusterLb()
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Errorf("failed to bind cluster lb, cluster: %s, err: %s", cluster.ID, err.Error())
		message := err.Error()
		WriteClientError(response, "CANNOT_BIND_TKE_CLUSTER_LB", message)
		return
	}

	response.WriteEntity(types.EmptyResponse{})

	metric.RequestCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())

}

// GetLbStatus call the tke api to get clb status
func GetLbStatus(request *restful.Request, response *restful.Response) {

	// support prometheus metrics
	start := time.Now()

	cluster := filters.GetCluster(request)

	// check if the cluster is valid
	externalClusterInfo := sqlstore.QueryBCSClusterInfo(&m.BCSClusterInfo{
		ClusterId: cluster.ID,
	})
	if externalClusterInfo == nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, external cluster info not exists", common.BcsErrApiBadRequest)
		WriteClientError(response, "EXTERNAL_CLUSTER_NOT_EXISTS", message)
		return
	}
	if externalClusterInfo.ClusterType != utils.BcsTkeCluster {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, cluster %s is not tke cluster", common.BcsErrApiBadRequest, cluster.ID)
		WriteClientError(response, "NOT_TKE_CLUSTER", message)
		return
	}

	// call tke api to get lb status
	tkeCluster := tke.NewTkeCluster(cluster.ID, externalClusterInfo.TkeClusterId, externalClusterInfo.TkeClusterRegion)
	status, err := tkeCluster.GetMasterVip()
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Errorf("failed to get lb status, cluster: %s, err: %s", cluster.ID, err.Error())
		message := err.Error()
		WriteClientError(response, "GET_TKE_MASTER_VIP_FAILED", message)
		return
	}

	lbStatus := &LbStatus{
		ClusterId: cluster.ID,
		Status:    status,
	}

	response.WriteEntity(*lbStatus)

	metric.RequestCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
}

// UpdateTkeLbSubnet update a tke subnet
func UpdateTkeLbSubnet(request *restful.Request, response *restful.Response) {

	// support prometheus metrics
	start := time.Now()

	blog.Debug(fmt.Sprintf("Create or Update tke lb subnet"))
	form := UpdateTkeLbForm{}
	request.ReadEntity(&form)

	// validate request data
	err := validate.Struct(&form)
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		response.WriteEntity(FormatValidationError(err))
		return
	}

	// save to db
	err = sqlstore.SaveTkeLbSubnet(form.ClusterRegion, form.SubnetId)
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, can not update tke lb subnet, error: %s", common.BcsErrApiInternalDbError, err.Error())
		WriteClientError(response, "CANNOT_UPDATE_TKE_LB_SUBNET", message)
		return
	}

	response.WriteEntity(types.EmptyResponse{})

	metric.RequestCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
}

// AddTkeCidr init tke cidrs into bcs-api
func AddTkeCidr(request *restful.Request, response *restful.Response) {

	// support prometheus metrics
	start := time.Now()

	blog.Info(fmt.Sprintf("Insert cidr"))
	form := AddTkeCidrForm{}
	request.ReadEntity(&form)

	// validate request data
	err := validate.Struct(&form)
	if err != nil {
		response.WriteEntity(FormatValidationError(err))
		return
	}

	for _, tkeCidr := range form.TkeCidrs {
		cidr := sqlstore.QueryTkeCidr(&m.TkeCidr{
			Vpc:      form.Vpc,
			Cidr:     tkeCidr.Cidr,
			IpNumber: tkeCidr.IpNumber,
		})
		// this vpc cidr already exist, skip and continue
		if cidr != nil {
			blog.Warnf("Add Cidr failed, Cidr %s IpNumber %d in vpc %s already exists", tkeCidr.Cidr, tkeCidr.IpNumber, form.Vpc)
			continue
		}
		// set the cidr status to be "available"
		if tkeCidr.Status == "" {
			tkeCidr.Status = sqlstore.CidrStatusAvailable
		}
		// save to db
		err = sqlstore.SaveTkeCidr(form.Vpc, tkeCidr.Cidr, tkeCidr.IpNumber, tkeCidr.Status, "")
		if err != nil {
			metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
			metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
			blog.Errorf("add tke cidr failed, error: %s", err.Error())
			message := fmt.Sprintf("errcode: %d, add tke cidr failed, error: %s", common.BcsErrApiInternalDbError, err.Error())
			WriteClientError(response, "ADD_TKE_CIDR_FAILED", message)
			return
		}
	}

	response.WriteEntity(types.EmptyResponse{})

	metric.RequestCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
}

// ApplyTkeCidr assign an cidr to client
func ApplyTkeCidr(request *restful.Request, response *restful.Response) {

	// support prometheus metrics
	start := time.Now()

	form := ApplyTkeCidrForm{}
	request.ReadEntity(&form)

	// validate request data
	err := validate.Struct(&form)
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		response.WriteEntity(FormatValidationError(err))
		return
	}

	mutex.Lock()
	defer mutex.Unlock()
	// apply from db
	tkeCidr := sqlstore.QueryTkeCidr(&m.TkeCidr{
		Vpc:      form.Vpc,
		IpNumber: form.IpNumber,
		Status:   sqlstore.CidrStatusAvailable,
	})
	if tkeCidr == nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Warnf("Apply cidr ipNumber %d for cluster %s in vpc %s failed, no available cidr any more", form.IpNumber, form.Cluster, form.Vpc)
		message := fmt.Sprintf("errcode: %d, apply cidr failed, no available cidr any more", common.BcsErrApiInternalDbError)
		WriteClientError(response, "NO_AVAILABLE_CIDR", message)
		return
	}

	// update to db
	updatedTkeCidr := tkeCidr
	updatedTkeCidr.Status = sqlstore.CidrStatusUsed
	updatedTkeCidr.Cluster = &form.Cluster
	err = sqlstore.UpdateTkeCidr(tkeCidr, updatedTkeCidr)
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Errorf("apply tkeCidr failed, cidr %s, error: %s", tkeCidr.Cidr, err.Error())
		message := fmt.Sprintf("errcode: %d, apply tkeCidr failed: %s", common.BcsErrApiInternalDbError, err.Error())
		WriteClientError(response, "APPLY_TKE_CIDR_FAILED", message)
		return
	}

	blog.Info("assign an cidr successful, cidr: %s, ipNumber: %d， vpc: %s", tkeCidr.Cidr, tkeCidr.IpNumber, form.Vpc)
	cidr := &ApplyTkeCidrResult{
		Vpc:      tkeCidr.Vpc,
		Cidr:     tkeCidr.Cidr,
		IpNumber: tkeCidr.IpNumber,
		Status:   sqlstore.CidrStatusUsed,
	}
	response.WriteEntity(cidr)

	metric.RequestCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
}

// ReleaseTkeCidr release a cidr to be available
func ReleaseTkeCidr(request *restful.Request, response *restful.Response) {

	// support prometheus metrics
	start := time.Now()

	form := ReleaseTkeCidrForm{}
	request.ReadEntity(&form)

	// validate request data
	err := validate.Struct(&form)
	if err != nil {
		response.WriteEntity(FormatValidationError(err))
		return
	}

	mutex.Lock()
	defer mutex.Unlock()
	tkeCidr := sqlstore.QueryTkeCidr(&m.TkeCidr{
		Vpc:     form.Vpc,
		Cidr:    form.Cidr,
		Cluster: &form.Cluster,
	})
	if tkeCidr == nil || tkeCidr.Status == sqlstore.CidrStatusAvailable {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Warnf("Release cidr %s in vpc %s failed, no such cidr to be released", form.Cidr, form.Vpc)
		message := fmt.Sprintf("errcode: %d, no such cidr to be released", common.BcsErrApiBadRequest)
		WriteClientError(response, "NO_SUCH_CIDR", message)
		return
	}

	// update to db
	updatedTkeCidr := tkeCidr
	updatedTkeCidr.Status = sqlstore.CidrStatusAvailable
	cluster := ""
	updatedTkeCidr.Cluster = &cluster
	err = sqlstore.UpdateTkeCidr(tkeCidr, updatedTkeCidr)
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Errorf("release tkeCidr failed, cidr %s vpc %s, error: %s", tkeCidr.Cidr, tkeCidr.Vpc, err.Error())
		message := fmt.Sprintf("errcode: %d, release tkeCidr failed: %s", common.BcsErrApiInternalDbError, err.Error())
		WriteClientError(response, "RELEASE_TKE_CIDR_FAILED", message)
		return
	}

	blog.Info("release cidr successful, vpc %s, cidr: %s, ipNumber: %d, cluster: %s", tkeCidr.Vpc, tkeCidr.Cidr, tkeCidr.IpNumber, tkeCidr.Cluster)
	response.WriteEntity(types.EmptyResponse{})

	metric.RequestCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
}

// ListTkeCidr list cidr count group by vpc
func ListTkeCidr(request *restful.Request, response *restful.Response) {
	// support prometheus metrics
	start := time.Now()

	cidrCounts := sqlstore.CountTkeCidr()
	response.WriteEntity(cidrCounts)

	metric.RequestCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())

}
