/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tkehandler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"github.com/emicklei/go-restful"
	"gopkg.in/go-playground/validator.v9"
)

// Validate local implementation
var Validate = validator.New()

// AddTkeCidrForm form of adding tke cidr
type AddTkeCidrForm struct {
	Vpc      string    `json:"vpc" validate:"required"`
	TkeCidrs []TkeCidr `json:"tke_cidrs" validate:"required"`
}

// TkeCidr tke cidr struct
type TkeCidr struct {
	Cidr     string `json:"cidr" validate:"required"`
	IPNumber uint64 `json:"ip_number" validate:"required"`
	Status   string `json:"status"`
}

// ApplyTkeCidrForm form of applying tke cidr
type ApplyTkeCidrForm struct {
	Vpc      string `json:"vpc" validate:"required"`
	Cluster  string `json:"cluster" validate:"required"`
	IPNumber uint64 `json:"ip_number" validate:"required"`
}

// ApplyTkeCidrResult result for applying tke cidr
type ApplyTkeCidrResult struct {
	Vpc      string `json:"vpc" validate:"required"`
	Cidr     string `json:"cidr" validate:"required"`
	IPNumber uint64 `json:"ip_number" validate:"required"`
	Status   string `json:"status"`
}

// ReleaseTkeCidrForm from of releasing tke cidr from
type ReleaseTkeCidrForm struct {
	Vpc     string `json:"vpc" validate:"required"`
	Cidr    string `json:"cidr" validate:"required"`
	Cluster string `json:"cluster" validate:"required"`
}

// TkeCidrCount tke cidr count
type TkeCidrCount struct {
	Vpc      string `json:"vpc"`
	IPNumber uint64 `json:"ip_number"`
	Count    uint64 `json:"count"`
	Status   string `json:"status"`
}

// Handler handler for Tke service
type Handler struct {
	model  store.ClusterManagerModel
	locker lock.DistributedLock
}

// NewTkeHandler create tke handler
func NewTkeHandler(model store.ClusterManagerModel, locker lock.DistributedLock) *Handler {
	return &Handler{
		model:  model,
		locker: locker,
	}
}

// AddTkeCidr init tke cidrs
func (h *Handler) AddTkeCidr(request *restful.Request, response *restful.Response) {
	blog.V(3).Infof("xreq %s, host %s, url %s, src %s",
		utils.GetXRequestIDFromHTTPRequest(request.Request),
		request.Request.Host,
		request.Request.URL,
		request.Request.RemoteAddr)
	start := time.Now()
	code := 200

	form := AddTkeCidrForm{}
	_ = request.ReadEntity(&form)

	// validate the request data
	err := Validate.Struct(&form)
	if err != nil {
		code = httpCodeClientError
		_ = response.WriteHeaderAndEntity(code, FormatValidationError(err))
		metrics.ReportAPIRequestMetric("AddTkeCidr", "http", strconv.Itoa(code), start)
		return
	}

	for _, tkeCidr := range form.TkeCidrs {
		cidr, err := h.model.GetTkeCidr(request.Request.Context(), form.Vpc, tkeCidr.Cidr)
		if err != nil {
			if !errors.Is(err, drivers.ErrTableRecordNotFound) {
				code = httpCodeClientError
				message := fmt.Sprintf("errcode: %d, add tke cidr failed, error: %s",
					types.BcsErrClusterManagerStoreOperationFailed, err.Error())
				WriteClientError(response, types.BcsErrClusterManagerStoreOperationFailed, message)
				metrics.ReportAPIRequestMetric("AddTkeCidr", "http", strconv.Itoa(code), start)
				return
			}
		}
		if cidr != nil {
			blog.Warnf("Add Cidr failed, Cidr %s IpNumber %d in vpc %s already exists",
				tkeCidr.Cidr, tkeCidr.IPNumber, form.Vpc)
			continue
		}
		if tkeCidr.Status == "" {
			tkeCidr.Status = types.TkeCidrStatusAvailable
		}
		now := time.Now()
		err = h.model.CreateTkeCidr(request.Request.Context(), &types.TkeCidr{
			Vpc:      form.Vpc,
			Cidr:     tkeCidr.Cidr,
			IPNumber: uint64(tkeCidr.IPNumber),
			Status:   tkeCidr.Status,
			Cluster:  "",
			CreateAt: now,
			UpdateAt: now,
		})
		if err != nil {
			code = httpCodeClientError
			message := fmt.Sprintf("errcode: %d, add tke cidr failed, err %s",
				types.BcsErrClusterManagerStoreOperationFailed, err.Error())
			blog.Warnf("add tke cidr failed, err %s", err.Error())
			WriteClientError(response, types.BcsErrClusterManagerStoreOperationFailed, message)
			metrics.ReportAPIRequestMetric("AddTkeCidr", "http", strconv.Itoa(code), start)
			return
		}
	}

	data := CreateResponeData(nil, "success", nil)
	response.Write([]byte(data))
	metrics.ReportAPIRequestMetric("AddTkeCidr", "http", strconv.Itoa(code), start)
}

func (h *Handler) getOneTkeCidr(ctx context.Context, vpc string, ipNumber uint64, status string) (
	*types.TkeCidr, error) {
	tkeCidrs, err := h.model.ListTkeCidr(ctx, operator.NewLeafCondition(operator.Eq, operator.M{
		"vpc":      vpc,
		"ipNumber": ipNumber,
		"status":   status,
	}), &storeopt.ListOption{
		Limit: 1,
	})
	if err != nil {
		blog.Warnf("get one tke cidr failed, err %s", err.Error())
		return nil, fmt.Errorf("get one tke cidr failed, err %s", err.Error())
	}
	if len(tkeCidrs) == 0 {
		blog.Warnf("get one tke cidr failed, no suitable cidr")
		return nil, fmt.Errorf("get one tke cidr failed, no suitable cidr")
	}
	if len(tkeCidrs) != 1 {
		blog.Warnf("get one tke cidr failed, returned more than one cidr, %+v", tkeCidrs)
		return nil, fmt.Errorf("get one tke cidr failed, returned more than one cidr")
	}
	return &tkeCidrs[0], nil
}

// ApplyTkeCidr assign an cidr to client
func (h *Handler) ApplyTkeCidr(request *restful.Request, response *restful.Response) {
	blog.V(3).Infof("xreq %s, host %s, url %s, src %s",
		utils.GetXRequestIDFromHTTPRequest(request.Request),
		request.Request.Host,
		request.Request.URL,
		request.Request.RemoteAddr)
	start := time.Now()
	code := 200

	form := ApplyTkeCidrForm{}
	_ = request.ReadEntity(&form)

	// validate the request data
	err := Validate.Struct(&form)
	if err != nil {
		code = httpCodeClientError
		_ = response.WriteHeaderAndEntity(code, FormatValidationError(err))
		metrics.ReportAPIRequestMetric("ApplyTkeCidr", "http", strconv.Itoa(code), start)
		return
	}

	h.locker.Lock(form.Vpc, []lock.LockOption{lock.LockTTL(5 * time.Second)}...)
	defer h.locker.Unlock(form.Vpc)

	// apply a available cidr
	tkeCidr, err := h.getOneTkeCidr(request.Request.Context(), form.Vpc, form.IPNumber, types.TkeCidrStatusAvailable)
	if err != nil {
		code = httpCodeClientError
		message := fmt.Sprintf("get one tke cidr failed, err %s", err.Error())
		blog.Warnf("get one tke cidr failed, err %s", err.Error())
		WriteClientError(response, types.BcsErrClusterManagerStoreOperationFailed, message)
		metrics.ReportAPIRequestMetric("ApplyTkeCidr", "http", strconv.Itoa(code), start)
		return
	}

	// update and save to db
	updatedTkeCidr := tkeCidr
	updatedTkeCidr.Status = types.TkeCidrStatusUsed
	updatedTkeCidr.Cluster = form.Cluster
	updatedTkeCidr.UpdateAt = time.Now()
	err = h.model.UpdateTkeCidr(request.Request.Context(), updatedTkeCidr)
	if err != nil {
		code = httpCodeClientError
		message := fmt.Sprintf("update tke cidr failed, err %s", err.Error())
		blog.Warnf("update tke cidr failed, err %s", err.Error())
		WriteClientError(response, types.BcsErrClusterManagerStoreOperationFailed, message)
		metrics.ReportAPIRequestMetric("ApplyTkeCidr", "http", strconv.Itoa(code), start)
		return
	}

	blog.Infof("assign a cidr successfully, cidr: %s, ipNumber: %d, vpc: %s", tkeCidr.Cidr, tkeCidr.IPNumber, form.Vpc)
	cidr := &ApplyTkeCidrResult{
		Vpc:      tkeCidr.Vpc,
		Cidr:     tkeCidr.Cidr,
		IPNumber: tkeCidr.IPNumber,
		Status:   types.TkeCidrStatusUsed,
	}
	data := CreateResponeData(nil, "success", cidr)
	response.Write([]byte(data))
	metrics.ReportAPIRequestMetric("ApplyTkeCidr", "http", strconv.Itoa(code), start)
}

// ReleaseTkeCidr release a cidr to be available
func (h *Handler) ReleaseTkeCidr(request *restful.Request, response *restful.Response) {
	blog.V(3).Infof("xreq %s, host %s, url %s, src %s",
		utils.GetXRequestIDFromHTTPRequest(request.Request),
		request.Request.Host,
		request.Request.URL,
		request.Request.RemoteAddr)
	start := time.Now()
	code := 200

	form := ReleaseTkeCidrForm{}
	_ = request.ReadEntity(&form)

	// validate the request data
	err := Validate.Struct(&form)
	if err != nil {
		code = httpCodeClientError
		_ = response.WriteHeaderAndEntity(code, FormatValidationError(err))
		metrics.ReportAPIRequestMetric("ReleaseTkeCidr", "http", strconv.Itoa(code), start)
		return
	}

	// check if cidr is valid
	h.locker.Lock(form.Vpc, []lock.LockOption{lock.LockTTL(5 * time.Second)}...)
	defer h.locker.Unlock(form.Vpc)
	tkeCidr, err := h.model.GetTkeCidr(request.Request.Context(), form.Vpc, form.Cidr)
	if err != nil {
		code = httpCodeClientError
		blog.Warnf("release cidr %s in vpc %s failed, err %s", form.Cidr, form.Vpc, err.Error())
		message := fmt.Sprintf("release cidr %s in vpc %s failed, err %s", form.Cidr, form.Vpc, err.Error())
		WriteClientError(response, types.BcsErrClusterManagerStoreOperationFailed, message)
		metrics.ReportAPIRequestMetric("ReleaseTkeCidr", "http", strconv.Itoa(code), start)
		return
	}
	if tkeCidr == nil || tkeCidr.Status == types.TkeCidrStatusAvailable {
		code = httpCodeClientError
		blog.Warnf("release cidr %s in vpc %s failed, no such cidr to be released", form.Cidr, form.Vpc)
		message := fmt.Sprintf("release cidr %s in vpc %s failed, no such cidr to be released", form.Cidr, form.Vpc)
		WriteClientError(response, types.BcsErrClusterManagerStoreOperationFailed, message)
		metrics.ReportAPIRequestMetric("ReleaseTkeCidr", "http", strconv.Itoa(code), start)
		return
	}

	// update and save to db
	cluster := tkeCidr.Cluster
	updatedTkeCidr := tkeCidr
	updatedTkeCidr.Status = types.TkeCidrStatusAvailable
	updatedTkeCidr.Cluster = ""
	updatedTkeCidr.UpdateAt = time.Now()
	err = h.model.UpdateTkeCidr(request.Request.Context(), updatedTkeCidr)
	if err != nil {
		code = httpCodeClientError
		message := fmt.Sprintf("release tke cidr failed, err %s", err.Error())
		blog.Warnf("release tke cidr failed, err %s", err.Error())
		WriteClientError(response, types.BcsErrClusterManagerStoreOperationFailed, message)
		metrics.ReportAPIRequestMetric("ReleaseTkeCidr", "http", strconv.Itoa(code), start)
		return
	}

	blog.Infof("release a cidr successfully, cidr: %s, ipNumber: %d, vpc: %s, cluster: %s",
		tkeCidr.Cidr, tkeCidr.IPNumber, form.Vpc, cluster)
	data := CreateResponeData(nil, "success", nil)
	response.Write([]byte(data))
	metrics.ReportAPIRequestMetric("ReleaseTkeCidr", "http", strconv.Itoa(code), start)
}

// ListTkeCidrCount list cidr count group by vpc
func (h *Handler) ListTkeCidrCount(request *restful.Request, response *restful.Response) {
	blog.V(3).Infof("xreq %s, host %s, url %s, src %s",
		utils.GetXRequestIDFromHTTPRequest(request.Request),
		request.Request.Host,
		request.Request.URL,
		request.Request.RemoteAddr)
	start := time.Now()
	code := 200

	storeTkeCidrCountList, err := h.model.ListTkeCidrCount(request.Request.Context(), &storeopt.ListOption{})
	if err != nil {
		code = httpCodeServerError
		message := fmt.Sprintf("list tke cidr count failed, err %s", err.Error())
		blog.Warnf("list tke cidr count failed, err %s", err.Error())
		WriteClientError(response, types.BcsErrClusterManagerStoreOperationFailed, message)
		metrics.ReportAPIRequestMetric("ListTkeCidrCount", "http", strconv.Itoa(code), start)
		return
	}
	var retTkeCidrCountList []TkeCidrCount
	for _, cidr := range storeTkeCidrCountList {
		retTkeCidrCountList = append(retTkeCidrCountList, TkeCidrCount{
			Vpc:      cidr.Vpc,
			IPNumber: cidr.IPNumber,
			Count:    cidr.Count,
			Status:   cidr.Status,
		})
	}
	blog.Infof("%+v", retTkeCidrCountList)
	// For forward compatibility, do not use code or message
	response.WriteEntity(retTkeCidrCountList)
	metrics.ReportAPIRequestMetric("ListTkeCidrCount", "http", strconv.Itoa(code), start)
}
