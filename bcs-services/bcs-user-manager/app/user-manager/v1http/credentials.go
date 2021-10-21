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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/emicklei/go-restful"
)

// UpdateCredentialsForm update form for credential
type UpdateCredentialsForm struct {
	RegisterToken   string `json:"register_token" validate:"required"`
	ServerAddresses string `json:"server_addresses" validate:"required,apiserver_addresses"`
	CaCertData      string `json:"cacert_data" validate:"required"`
	UserToken       string `json:"user_token" validate:"required"`
}

// CredentialResp response
type CredentialResp struct {
	ServerAddresses string `json:"server_addresses"`
	CaCertData      string `json:"ca_cert_data"`
	UserToken       string `json:"user_token"`
	ClusterDomain   string `json:"cluster_domain"`
}

// UpdateCredentials updates the current cluster credentials, a valid registerToken is required to performe
// a credentials updating.
func UpdateCredentials(request *restful.Request, response *restful.Response) {
	start := time.Now()

	form := UpdateCredentialsForm{}
	_ = request.ReadEntity(&form)
	err := utils.Validate.Struct(&form)
	if err != nil {
		metrics.ReportRequestAPIMetrics("UpdateCredentials", request.Request.Method, metrics.ErrStatus, start)
		_ = response.WriteHeaderAndEntity(400, utils.FormatValidationError(err))
		return
	}

	clusterID := request.PathParameter("cluster_id")

	// validate if the registerToken is correct
	token := sqlstore.GetRegisterToken(clusterID)
	if token == nil {
		metrics.ReportRequestAPIMetrics("UpdateCredentials", request.Request.Method, metrics.ErrStatus, start)
		blog.Warnf("no valid register token found for cluster [%s]", clusterID)
		message := fmt.Sprintf("errcode: %d, no valid register token found for cluster", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}
	if token.Token != form.RegisterToken {
		metrics.ReportRequestAPIMetrics("UpdateCredentials", request.Request.Method, metrics.ErrStatus, start)
		blog.Warnf("register token [%s] is in valid", form.RegisterToken)
		message := fmt.Sprintf("errcode: %d, invalid register token given", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	err = sqlstore.SaveCredentials(clusterID, form.ServerAddresses, form.CaCertData, form.UserToken, "")
	if err != nil {
		metrics.ReportRequestAPIMetrics("UpdateCredentials", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to update cluster [%s] credential: %s", clusterID, err.Error())
		message := fmt.Sprintf("errcode: %d, can not update credentials, error: %s", common.BcsErrApiInternalDbError, err.Error())
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}
	data := utils.CreateResponeData(nil, "success", nil)
	response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("UpdateCredentials", request.Request.Method, metrics.SucStatus, start)
}

// GetCredentials get credential according cluster ID
func GetCredentials(request *restful.Request, response *restful.Response) {
	start := time.Now()

	clusterID := request.PathParameter("cluster_id")
	credential := sqlstore.GetCredentials(clusterID)
	if credential == nil {
		metrics.ReportRequestAPIMetrics("GetCredentials", request.Request.Method, metrics.ErrStatus, start)
		blog.Warnf("credentials not found for cluster [%s]", clusterID)
		message := fmt.Sprintf("errcode: %d, credentials not found", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	data := utils.CreateResponeData(nil, "success", credential)
	response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("GetCredentials", request.Request.Method, metrics.SucStatus, start)
}

// ListCredentials list all cluster credentials
func ListCredentials(request *restful.Request, response *restful.Response) {
	start := time.Now()

	credentials := make(map[string]CredentialResp)
	newCredentials := sqlstore.ListCredentials()
	for _, v := range newCredentials {
		credentials[v.ClusterId] = CredentialResp{
			ServerAddresses: v.ServerAddresses,
			CaCertData:      v.CaCertData,
			UserToken:       v.UserToken,
			ClusterDomain:   v.ClusterDomain,
		}
	}

	blog.Infof("client %s list all cluster credentials, num: %d", request.Request.RemoteAddr, len(credentials))
	data := utils.CreateResponeData(nil, "success", credentials)
	response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("ListCredentials", request.Request.Method, metrics.SucStatus, start)
}
