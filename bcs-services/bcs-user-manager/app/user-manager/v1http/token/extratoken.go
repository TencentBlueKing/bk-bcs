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

package token

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	cmdb "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/cmdbv3"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/cmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/emicklei/go-restful"
)

// ExtraTokenHandler handles extra token for third-party system
type ExtraTokenHandler struct {
	tokenStore    sqlstore.TokenStore
	notifyStore   sqlstore.TokenNotifyStore
	cache         cache.Cache
	jwtClient     jwt.BCSJWTAuthentication
	clusterClient *cmanager.ClusterManagerClient
	cmdbClient    *cmdb.Client
	encryptPriKey string
}

// NewExtraTokenHandler creates a new ExtraTokenHandler
func NewExtraTokenHandler(tokenStore sqlstore.TokenStore, notifyStore sqlstore.TokenNotifyStore, cache cache.Cache,
	jwtClient jwt.BCSJWTAuthentication, clusterClient *cmanager.ClusterManagerClient, cmdbClient *cmdb.Client) *ExtraTokenHandler {
	return &ExtraTokenHandler{
		tokenStore:    tokenStore,
		notifyStore:   notifyStore,
		cache:         cache,
		jwtClient:     jwtClient,
		clusterClient: clusterClient,
		cmdbClient:    cmdbClient,
		encryptPriKey: os.Getenv("ENCRYPT_PRI_KEY"),
	}
}

// ExtraTokenResponse is the response of extra token
type ExtraTokenResponse struct {
	UserName  string       `json:"username"`
	Token     string       `json:"token"`
	Status    *TokenStatus `json:"status,omitempty"`
	ExpiredAt *time.Time   `json:"expired_at"` // nil means never expired
}

// GetTokenByUserAndClusterID get token by user and cluster id
func (t *ExtraTokenHandler) GetTokenByUserAndClusterID(request *restful.Request, response *restful.Response) {
	username := request.Request.URL.Query().Get("username")
	clusterID := request.Request.URL.Query().Get("cluster_id")
	businessID := request.Request.URL.Query().Get("business_id")
	start := time.Now()
	if len(username) == 0 || len(clusterID) == 0 || len(businessID) == 0 {
		blog.Errorf("param from %s is invalid", request.Request.RemoteAddr)
		metrics.ReportRequestAPIMetrics("GetTokenByUserAndClusterID", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, "param is invalid")
		return
	}

	respBusinessID, err := t.clusterClient.GetBusinessIDByClusterID(clusterID)
	if err != nil {
		blog.Errorf("GetBusinessIDByClusterID failed, err: %v", err.Error())
		metrics.ReportRequestAPIMetrics("GetTokenByUserAndClusterID", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, err.Error())
		return
	}
	if len(respBusinessID) == 0 {
		blog.Errorf("GetBusinessIDByClusterID failed, cluster %s not found", clusterID)
		metrics.ReportRequestAPIMetrics("GetTokenByUserAndClusterID", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteServerError(response, common.BcsErrApiInternalDbError, fmt.Sprintf("cluster %s not found", clusterID))
		return
	}

	// check businessID
	if businessID != respBusinessID {
		message := fmt.Sprintf("cluster %s is not belong to business %s", clusterID, businessID)
		blog.Error(message)
		metrics.ReportRequestAPIMetrics("GetTokenByUserAndClusterID", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteForbiddenError(response, 400, message)
		return
	}
	// check user is maintainer
	intBizID, _ := strconv.Atoi(businessID)
	bizResult, err := t.cmdbClient.ESBSearchBusiness("", map[string]interface{}{
		"bk_biz_id": intBizID,
	})
	if bizResult == nil || bizResult.Data == nil || len(bizResult.Data.Info) == 0 {
		message := fmt.Sprintf("business %s is not found", businessID)
		blog.Error(message)
		metrics.ReportRequestAPIMetrics("GetTokenByUserAndClusterID", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteForbiddenError(response, 400, message)
		return
	}
	if !utils.StringInSlice(username, strings.Split(bizResult.Data.Info[0].BkBizMaintainer, ",")) {
		message := fmt.Sprintf("user %s is not maintainer in business %s", username, businessID)
		blog.Error(message)
		metrics.ReportRequestAPIMetrics("GetTokenByUserAndClusterID", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteUnauthorizedError(response, 401, message)
		return
	}

	tokenInDB := t.tokenStore.GetTokenByCondition(&models.BcsUser{Name: username})
	if tokenInDB == nil {
		blog.Errorf("can't find user %s token", username)
		metrics.ReportRequestAPIMetrics("GetTokenByUserAndClusterID", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteForbiddenError(response, 400, fmt.Sprintf("can't find user %s token", username))
		return
	}
	status := TokenStatusActive
	if tokenInDB.HasExpired() {
		status = TokenStatusExpired
	}
	expiresAt := &tokenInDB.ExpiresAt
	// transfer never expired
	if expiresAt.After(NeverExpired) {
		expiresAt = nil
	}
	// encrypt token
	encryptToken, err := encrypt.DesEncryptToBase([]byte(tokenInDB.UserToken), t.encryptPriKey)
	if err != nil {
		blog.Errorf("encrypt token failed, err: %s", err.Error())
		metrics.ReportRequestAPIMetrics("GetTokenByUserAndClusterID", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteServerError(response, 500, "encrypt token failed")
		return
	}
	respToken := &ExtraTokenResponse{
		UserName:  username,
		Token:     string(encryptToken),
		Status:    &status,
		ExpiredAt: expiresAt,
	}
	data := utils.CreateResponseData(nil, "success", respToken)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("GetTokenByUserAndClusterID", request.Request.Method, metrics.SucStatus, start)
}
