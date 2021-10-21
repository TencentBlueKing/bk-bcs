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

package v1http

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/external-cluster/tke"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"

	"github.com/emicklei/go-restful"
)

type AddTkeCidrForm struct {
	Vpc      string    `json:"vpc" validate:"required"`
	TkeCidrs []TkeCidr `json:"tke_cidrs" validate:"required"`
}

type TkeCidr struct {
	Cidr     string `json:"cidr" validate:"required"`
	IpNumber uint   `json:"ip_number" validate:"required"`
	Status   string `json:"status"`
}

type ApplyTkeCidrForm struct {
	Vpc      string `json:"vpc" validate:"required"`
	Cluster  string `json:"cluster" validate:"required"`
	IpNumber uint   `json:"ip_number" validate:"required"`
}

type ApplyTkeCidrResult struct {
	Vpc      string `json:"vpc" validate:"required"`
	Cidr     string `json:"cidr" validate:"required"`
	IpNumber uint   `json:"ip_number" validate:"required"`
	Status   string `json:"status"`
}

type ReleaseTkeCidrForm struct {
	Vpc     string `json:"vpc" validate:"required"`
	Cidr    string `json:"cidr" validate:"required"`
	Cluster string `json:"cluster" validate:"required"`
}

// AddTkeCidr init tke cidrs
func AddTkeCidr(request *restful.Request, response *restful.Response) {
	start := time.Now()

	blog.Info(fmt.Sprintf("Insert cidr"))
	form := AddTkeCidrForm{}
	_ = request.ReadEntity(&form)

	// validate the request data
	err := utils.Validate.Struct(&form)
	if err != nil {
		metrics.ReportRequestAPIMetrics("AddTkeCidr", request.Request.Method, metrics.ErrStatus, start)
		_ = response.WriteHeaderAndEntity(400, utils.FormatValidationError(err))
		return
	}

	for _, tkeCidr := range form.TkeCidrs {
		cidr := sqlstore.QueryTkeCidr(&models.TkeCidr{
			Vpc:      form.Vpc,
			Cidr:     tkeCidr.Cidr,
			IpNumber: tkeCidr.IpNumber,
		})
		if cidr != nil {
			blog.Warnf("Add Cidr failed, Cidr %s IpNumber %d in vpc %s already exists", tkeCidr.Cidr, tkeCidr.IpNumber, form.Vpc)
			continue
		}
		if tkeCidr.Status == "" {
			tkeCidr.Status = sqlstore.CidrStatusAvailable
		}
		err = sqlstore.SaveTkeCidr(form.Vpc, tkeCidr.Cidr, tkeCidr.IpNumber, tkeCidr.Status, "")
		if err != nil {
			metrics.ReportRequestAPIMetrics("AddTkeCidr", request.Request.Method, metrics.ErrStatus, start)
			blog.Errorf("add tke cidr failed, error: %s", err.Error())
			message := fmt.Sprintf("errcode: %d, add tke cidr failed, error: %s", common.BcsErrApiInternalDbError, err.Error())
			utils.WriteClientError(response, common.BcsErrApiInternalDbError, message)
			return
		}
	}

	data := utils.CreateResponeData(nil, "success", nil)
	response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("AddTkeCidr", request.Request.Method, metrics.SucStatus, start)
}

// ApplyTkeCidr assign an cidr to client
func ApplyTkeCidr(request *restful.Request, response *restful.Response) {
	start := time.Now()

	form := ApplyTkeCidrForm{}
	_ = request.ReadEntity(&form)

	// validate the request data
	err := utils.Validate.Struct(&form)
	if err != nil {
		metrics.ReportRequestAPIMetrics("ApplyTkeCidr", request.Request.Method, metrics.ErrStatus, start)
		_ = response.WriteHeaderAndEntity(400, utils.FormatValidationError(err))
		return
	}

	mutex.Lock()
	defer mutex.Unlock()
	// apply a available cidr
	tkeCidr := sqlstore.QueryTkeCidr(&models.TkeCidr{
		Vpc:      form.Vpc,
		IpNumber: form.IpNumber,
		Status:   sqlstore.CidrStatusAvailable,
	})
	// no more available cidr
	if tkeCidr == nil {
		metrics.ReportRequestAPIMetrics("ApplyTkeCidr", request.Request.Method, metrics.ErrStatus, start)
		blog.Warnf("Apply cidr ipNumber %d for cluster %s in vpc %s failed, no available cidr any more", form.IpNumber, form.Cluster, form.Vpc)
		message := fmt.Sprintf("errcode: %d, apply cidr failed, no available cidr any more", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	// update and save to db
	updatedTkeCidr := tkeCidr
	updatedTkeCidr.Status = sqlstore.CidrStatusUsed
	updatedTkeCidr.Cluster = &form.Cluster
	err = sqlstore.UpdateTkeCidr(tkeCidr, updatedTkeCidr)
	if err != nil {
		metrics.ReportRequestAPIMetrics("ApplyTkeCidr", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("apply tkeCidr failed, cidr %s, error: %s", tkeCidr.Cidr, err.Error())
		message := fmt.Sprintf("errcode: %d, apply tkeCidr failed: %s", common.BcsErrApiInternalDbError, err.Error())
		utils.WriteClientError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	blog.Info("assign an cidr successful, cidr: %s, ipNumber: %d， vpc: %s", tkeCidr.Cidr, tkeCidr.IpNumber, form.Vpc)
	cidr := &ApplyTkeCidrResult{
		Vpc:      tkeCidr.Vpc,
		Cidr:     tkeCidr.Cidr,
		IpNumber: tkeCidr.IpNumber,
		Status:   sqlstore.CidrStatusUsed,
	}

	data := utils.CreateResponeData(nil, "success", cidr)
	response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("ApplyTkeCidr", request.Request.Method, metrics.SucStatus, start)
}

// ReleaseTkeCidr release a cidr to be available
func ReleaseTkeCidr(request *restful.Request, response *restful.Response) {
	start := time.Now()

	form := ReleaseTkeCidrForm{}
	_ = request.ReadEntity(&form)

	// validate the request data
	err := utils.Validate.Struct(&form)
	if err != nil {
		metrics.ReportRequestAPIMetrics("ReleaseTkeCidr", request.Request.Method, metrics.ErrStatus, start)
		_ = response.WriteHeaderAndEntity(400, utils.FormatValidationError(err))
		return
	}

	// check if the cidr is valid
	mutex.Lock()
	defer mutex.Unlock()
	tkeCidr := sqlstore.QueryTkeCidr(&models.TkeCidr{
		Vpc:     form.Vpc,
		Cidr:    form.Cidr,
		Cluster: &form.Cluster,
	})
	if tkeCidr == nil || tkeCidr.Status == sqlstore.CidrStatusAvailable {
		metrics.ReportRequestAPIMetrics("ReleaseTkeCidr", request.Request.Method, metrics.ErrStatus, start)
		blog.Warnf("Release cidr %s in vpc %s failed, no such cidr to be released", form.Cidr, form.Vpc)
		message := fmt.Sprintf("errcode: %d, no such cidr to be released", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	// update and save to db
	updatedTkeCidr := tkeCidr
	updatedTkeCidr.Status = sqlstore.CidrStatusAvailable
	cluster := ""
	updatedTkeCidr.Cluster = &cluster
	err = sqlstore.UpdateTkeCidr(tkeCidr, updatedTkeCidr)
	if err != nil {
		metrics.ReportRequestAPIMetrics("ReleaseTkeCidr", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("release tkeCidr failed, cidr %s vpc %s, error: %s", tkeCidr.Cidr, tkeCidr.Vpc, err.Error())
		message := fmt.Sprintf("errcode: %d, release tkeCidr failed: %s", common.BcsErrApiInternalDbError, err.Error())
		utils.WriteClientError(response, common.BcsErrApiInternalDbError, message)
		return
	}

	blog.Info("release cidr successful, vpc %s, cidr: %s, ipNumber: %d, cluster: %s", tkeCidr.Vpc, tkeCidr.Cidr, tkeCidr.IpNumber, tkeCidr.Cluster)
	data := utils.CreateResponeData(nil, "success", nil)
	response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("ReleaseTkeCidr", request.Request.Method, metrics.SucStatus, start)
}

// SyncTkeClusterCredentials sync the tke cluster credentials from tke
func SyncTkeClusterCredentials(request *restful.Request, response *restful.Response) {
	start := time.Now()

	// whether this cluster is valid
	clusterId := request.PathParameter("cluster_id")
	cluster := sqlstore.GetCluster(clusterId)
	if cluster == nil {
		metrics.ReportRequestAPIMetrics("SyncTkeClusterCredentials", request.Request.Method, metrics.ErrStatus, start)
		blog.Warnf("cluster [%s] not exists", clusterId)
		message := fmt.Sprintf("errcode: %d, cluster not exists", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	if cluster.ClusterType != BcsTkeCluster {
		metrics.ReportRequestAPIMetrics("SyncTkeClusterCredentials", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, cluster %s is not tke cluster", common.BcsErrApiBadRequest, cluster.ID)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	tkeCluster := tke.NewTkeCluster(cluster.ID, cluster.TkeClusterId, cluster.TkeClusterRegion)

	// call tke api to sync credentials
	err := tkeCluster.SyncClusterCredentials()
	if err != nil {
		metrics.ReportRequestAPIMetrics("SyncTkeClusterCredentials", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("error when sync tke cluster [%s] credentials: %s", clusterId, err.Error())
		message := fmt.Sprintf("error when sync tke cluster [%s] credentials: %s", clusterId, err.Error())
		utils.WriteClientError(response, common.BcsErrApiInternalFail, message)
		return
	}
	data := utils.CreateResponeData(nil, "success", nil)
	response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("SyncTkeClusterCredentials", request.Request.Method, metrics.SucStatus, start)
}

// ListTkeCidr list cidr count group by vpc
func ListTkeCidr(request *restful.Request, response *restful.Response) {
	// support prometheus metrics

	cidrCounts := sqlstore.CountTkeCidr()
	response.WriteEntity(cidrCounts)

}
