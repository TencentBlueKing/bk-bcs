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

package apiclient

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/apiclient/xrequests"
)

var (
	httpMethodMep = map[string]xrequests.SendFunc{
		http.MethodGet:    xrequests.Get,
		http.MethodPost:   xrequests.Post,
		http.MethodDelete: xrequests.Delete,
		http.MethodPut:    xrequests.Put,
	}
)

// BaseRequest bk base request
type BaseRequest struct {
	BkBizId     int64  `json:"bk_biz_id"`
	BkAppCode   string `json:"bk_app_code"`
	BkAppSecret string `json:"bk_app_secret"`
	AccessToken string `json:"access_token"`
}

// ApiResponse contains the common information and data
type ApiResponse struct {
	*BaseResponse
	Data any `json:"data"`
}

// NewBaseRequest return new base request
func NewBaseRequest(bkBizId int64, bkAppCode, bkAppSecret, accessToken string) *BaseRequest {
	return &BaseRequest{
		BkBizId:     bkBizId,
		BkAppCode:   bkAppCode,
		BkAppSecret: bkAppSecret,
		AccessToken: accessToken,
	}
}

// BaseResponse bk base response
type BaseResponse struct {
	Result     bool   `json:"result"`
	Code       any    `json:"code"`
	Message    string `json:"message"`
	Permission any    `json:"permission"`
	RequestId  string `json:"request_id"`
}

// UptimeCheckTask def pf uptime check
type UptimeCheckTask struct {
	ID       int64    `json:"id,omitempty"`
	Config   Config   `json:"config,omitempty"`
	Location Location `json:"location,omitempty"`
	Nodes    []Node   `json:"nodes,omitempty"`
	Groups   []Group  `json:"groups,omitempty"`
	// Available     float64   `json:"available,omitempty"`
	// TaskDuration  float64   `json:"task_duration,omitempty"`
	URL []string `json:"url,omitempty"`
	// CreateTime    time.Time `json:"create_time,omitempty"`
	// UpdateTime    time.Time `json:"update_time,omitempty"`
	// CreateUser    string    `json:"create_user,omitempty"`
	// UpdateUser    string    `json:"update_user,omitempty"`
	IsDeleted     bool   `json:"is_deleted,omitempty"`
	BkBizID       int64  `json:"bk_biz_id,omitempty"`
	Name          string `json:"name,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
	CheckInterval int    `json:"check_interval,omitempty"`
	Status        string `json:"status,omitempty"`
}

// Config uptime check task config
type Config struct {
	Method            string     `json:"method,omitempty"` // GET POST PUT DELETE PATCH
	Authorize         *Authorize `json:"authorize,omitempty"`
	Body              *Body      `json:"body,omitempty"`
	QueryParams       []*Params  `json:"query_params,omitempty"`
	Headers           []*Params  `json:"headers,omitempty"`
	ResponseCode      string     `json:"response_code,omitempty"`
	IPList            []string   `json:"ip_list,omitempty"`
	OutputFields      []string   `json:"output_fields,omitempty"`
	TargetIPType      int64      `json:"target_ip_type,omitempty"`
	DNSCheckMode      string     `json:"dns_check_mode,omitempty"`
	Period            int64      `json:"period,omitempty"`
	ResponseFormat    string     `json:"response_format"` // nin / raw|in /
	Response          string     `json:"response"`
	Timeout           int64      `json:"timeout,omitempty"`
	Urls              string     `json:"urls,omitempty"`     // 单个URL时使用这个字段
	URLList           []string   `json:"url_list,omitempty"` // 多个URL时使用这个字段
	Port              string     `json:"port,omitempty"`
	Request           string     `json:"request"`
	RequestFormat     string     `json:"request_format"`
	WaitEmptyResponse bool       `json:"wait_empty_response,omitempty"`
}

// Params http check params
type Params struct {
	IsEnabled bool   `json:"is_enabled"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Desc      string `json:"desc"`
}

// Authorize auth config
type Authorize struct {
	AuthType           string      `json:"auth_type,omitempty"` // basic_auth / bearer_token
	AuthConfig         *AuthConfig `json:"auth_config,omitempty"`
	InsecureSkipVerify bool        `json:"insecure_skip_verify,omitempty"`
}

// AuthConfig auth config
type AuthConfig struct {
	Token    string `json:"token,omitempty"`    // set when auth_type = 'bearer_token'
	UserName string `json:"username,omitempty"` // set when auth_type = 'basic_auth'
	PassWord string `json:"password,omitempty"` // set when auth_type = 'basic_auth'
}

// Body http body
type Body struct {
	DataType    string    `json:"data_type,omitempty"`
	Params      []*Params `json:"params,omitempty"`
	Content     string    `json:"content,omitempty"`
	ContentType string    `json:"content_type,omitempty"`
}

// Location bk location
type Location struct {
	BkStateName    string `json:"bk_state_name"`
	BkProvinceName string `json:"bk_province_name"`
}

// Node info
type Node struct {
	ID         int64  `json:"id,omitempty"`
	CreateUser string `json:"create_user,omitempty"`
	UpdateUser string `json:"update_user,omitempty"`
	IsDeleted  bool   `json:"is_deleted,omitempty"`
	BkBizID    int64  `json:"bk_biz_id,omitempty"`
	IsCommon   bool   `json:"is_common,omitempty"`
	// BizScope        string       `json:"biz_scope,omitempty"`
	IPType          int           `json:"ip_type,omitempty"`
	Name            string        `json:"name,omitempty"`
	IP              string        `json:"ip,omitempty"`
	BkHostID        int64         `json:"bk_host_id,omitempty"`
	BkCloudID       int           `json:"bk_cloud_id,omitempty"`
	Location        *NodeLocation `json:"location,omitempty"`
	Carrieroperator string        `json:"carrieroperator,omitempty"`
}

// NodeLocation node location
type NodeLocation struct {
	Country string `json:"country,omitempty"`
	City    string `json:"city,omitempty"`
}

// Group task group
type Group struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// -----------------------------------------------

// ListUptimeCheckRequest list uptime check request
type ListUptimeCheckRequest struct {
	BkBizID         int64 `json:"bkBizID,omitempty"`
	GetAvailable    bool  `json:"getAvailable,omitempty"`
	GetTaskDuration bool  `json:"getTaskDuration,omitempty"`
	GetGroups       bool  `json:"getGroups,omitempty"`
}

// ListUptimeCheckResponse list uptime check resp
type ListUptimeCheckResponse struct {
	*BaseResponse
	Data []*UptimeCheckTask `json:"data"`
}

// -----------------------------------------------

// ListNodeRequest list node req
type ListNodeRequest struct {
	BkBizID int64 `json:"bkBizID,omitempty"`
}

// ListNodeResponse list node resp
type ListNodeResponse struct {
	*BaseResponse
	Data []*Node `json:"data"`
}

// -----------------------------------------------

// CreateOrUpdateUptimeCheckTaskResponse create or update uptime check response
type CreateOrUpdateUptimeCheckTaskResponse struct {
	*BaseResponse
	Data *UptimeCheckTask `json:"data"`
}

// CreateOrUpdateUptimeCheckTaskRequest create or update check request
type CreateOrUpdateUptimeCheckTaskRequest struct {
	// BKBizID     int64    `json:"bk_biz_id"`
	TaskID      int64    `json:"task_id,omitempty"` // set when update
	Protocol    string   `json:"protocol"`
	NodeIDList  []int64  `json:"node_id_list"` // 拨测节点ID
	Config      Config   `json:"config"`
	Location    Location `json:"location"`
	Name        string   `json:"name"`
	GroupIDList []int64  `json:"group_id_list"` // 拨测任务组ID
}

// -----------------------------------------------

// DeleteUptimeCheckRequest delete uptime check
type DeleteUptimeCheckRequest struct {
	// BkBizID int64 `json:"bk_biz_id"`
	TaskID int64 `json:"task_id"`
}

// -----------------------------------------------

// DeployUptimeCheckRequest deploy uptime check
type DeployUptimeCheckRequest struct {
	// BkBizID int64 `json:"bk_biz_id"`
	TaskID int64 `json:"task_id"`
}

// -----------------------------------------------

// GetApigwApiUrl apigw api
func GetApigwApiUrl(serviceName string, urlPrefix string, uri string) string {
	apigwHost := os.Getenv(envNameApiGwHost)
	return fmt.Sprintf("%s://%s.%s%s%s", apigwApiScheme, serviceName, apigwHost, urlPrefix, uri)
}
