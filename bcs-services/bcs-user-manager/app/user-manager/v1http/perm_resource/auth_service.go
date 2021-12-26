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

package perm_resource

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/cmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/iam"

	"github.com/emicklei/go-restful"
	"github.com/patrickmn/go-cache"
)

const (
	// BKHTTPHeaderUser bk_user
	BKHTTPHeaderUser            = "BK_User"
	// IamRequestHeader request-id
	IamRequestHeader            = "X-Request-Id"
	// BCSHTTPCookieLanugageKey cookie language info
	BCSHTTPCookieLanugageKey    = "blueking_language"
	// BCSHTTPUserManagerRequestID usermanager request-id
	BCSHTTPUserManagerRequestID = "UserManager_Request_Id"
	// IAMTokenKey iam token key
	IAMTokenKey = "iam_token"
)

// Options for init authService parameter
type Options struct {
	IamPermCli iam.PermIAMClient
	ClusterCli *cmanager.ClusterManagerClient

	EnableAuth bool
}

// NewAuthService create authService client
func NewAuthService(opt *Options) *AuthService {
	registerResourceTypes := []iam.TypeID{iam.SysCluster}

	// register resource provider
	dispatcher := NewDispatcher()
	clusterProvider := NewClusterResourceProvider(opt.ClusterCli, iam.SysCluster)
	dispatcher.RegisterProvider(string(iam.SysCluster), clusterProvider)

	for _, rt := range registerResourceTypes {
		_, exist := dispatcher.GetProvider(string(rt))
		if exist {
			blog.Infof("register provider[%s] successful", string(rt))
		}
	}

	tokenCache := cache.New(time.Second*10, time.Second*30)

	return &AuthService{
		iam:        opt.IamPermCli,
		cm:         opt.ClusterCli,
		dispatcher: dispatcher,
		enableAuth: opt.EnableAuth,
		tokenCache: tokenCache,
	}
}

// AuthService for resource pull
type AuthService struct {
	iam iam.PermIAMClient
	cm  *cmanager.ClusterManagerClient

	dispatcher Dispatcher
	enableAuth bool
	tokenCache *cache.Cache
}

// PullResource pull different resource by request
func (as *AuthService) PullResource(req *restful.Request, response *restful.Response) {
	if as == nil {
		CreateResError(response, http.StatusInternalServerError, fmt.Sprintf("auth service not init"))
		return
	}

	data, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		CreateResError(response, http.StatusInternalServerError, fmt.Sprintf("read request body failed"))
		return
	}

	resourceReq := new(PullResourceReq)
	err = resourceReq.UnmarshalJSON(data)
	if err != nil {
		blog.Errorf("auth service Unmarshal PullResourceReq failed: %v", err)
		CreateResError(response, http.StatusInternalServerError, fmt.Sprintf("AuthService unmarshal PullResourceReq failed: %v", err))
		return
	}

	blog.V(5).Infof("PullResourceReq %+v", resourceReq)

	provider, exist := as.dispatcher.GetProvider(string(resourceReq.Type))
	if !exist {
		CreateResError(response, http.StatusNotFound, fmt.Sprintf("resourceType %s provider not found", resourceReq.Type))
		return
	}

	switch resourceReq.Method {
	case ListAttrMethod:
		attrResource, err := provider.ListAttr()
		if err != nil {
			CreateResError(response, http.StatusInternalServerError, err.Error())
			return
		}

		CreateResEntity(response, attrResource)
		return
	case ListAttrValueMethod:
		filter, err := ValidateListAttrValueRequest(resourceReq)
		if err != nil {
			CreateResError(response, http.StatusBadRequest, fmt.Sprintf("invalid filter parameter"))
			return
		}
		attrValue, err := provider.ListAttrValue(filter, resourceReq.Page)
		if err != nil {
			CreateResError(response, http.StatusBadRequest, err.Error())
			return
		}
		CreateResEntity(response, attrValue)
		return
	case ListInstanceMethod:
		filter, err := ValidateListInstanceRequest(resourceReq)
		if err != nil {
			CreateResError(response, http.StatusBadRequest, err.Error())
			return
		}
		instances, err := provider.ListInstance(filter, resourceReq.Page)
		if err != nil {
			CreateResError(response, http.StatusBadRequest, err.Error())
			return
		}
		CreateResEntity(response, instances)
		return
	case FetchInstanceInfoMethod:
		filter, err := ValidateFetchInstanceRequest(resourceReq)
		if err != nil {
			CreateResError(response, http.StatusBadRequest, fmt.Sprintf("invalid filter parameter"))
			return
		}
		instanceInfo, err := provider.FetchInstanceInfo(filter)
		if err != nil {
			CreateResError(response, http.StatusBadRequest, err.Error())
			return
		}
		CreateResEntity(response, instanceInfo)
		return
	case ListInstanceByPolicyMethod:
		filter, err := ValidateListInstanceByPolicyRequest(resourceReq)
		if err != nil {
			CreateResError(response, http.StatusBadRequest, fmt.Sprintf("invalid filter parameter"))
			return
		}

		instanceResult, err := provider.ListInstanceByPolicy(filter, resourceReq.Page)
		if err != nil {
			CreateResError(response, http.StatusBadRequest, err.Error())
			return
		}
		CreateResEntity(response, instanceResult)
		return
	case SearchInstanceMethod:
		filter, err := ValidateSearchInstanceRequest(resourceReq)
		if err != nil {
			CreateResError(response, http.StatusBadRequest, fmt.Sprintf("invalid filter parameter"))
			return
		}
		instance, err := provider.SearchInstance(filter, resourceReq.Page)
		if err != nil {
			CreateResError(response, http.StatusBadRequest, err.Error())
			return
		}

		CreateResEntity(response, instance)
		return
	default:
		rsp := iam.BaseResponse{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("method %s not found", resourceReq.Method),
		}
		_ = response.WriteAsJson(rsp)
		return
	}

	return
}

// Healthz for system health interface
func (as *AuthService) Healthz(request *restful.Request, response *restful.Response) {
	if as == nil {
		CreateResError(response, http.StatusInternalServerError, fmt.Sprintf("auth service not init"))
		return
	}

	CreateResEntity(response, "ok")
	return
}

// FilterRequestFromIAM check request authInfo
func (as *AuthService) FilterRequestFromIAM() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		if !as.enableAuth {
			chain.ProcessFilter(req, resp)
			return
		}

		isAuthorized, err := as.checkRequestAuthorization(req.Request)
		if err != nil {
			rsp := iam.BaseResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}

			_ = resp.WriteAsJson(rsp)
			return
		}
		if !isAuthorized {
			rsp := iam.BaseResponse{
				Code:    http.StatusUnauthorized,
				Message: "request not from iam",
			}
			_ = resp.WriteAsJson(rsp)
			return
		}

		// get iam request id
		rid := req.Request.Header.Get(IamRequestHeader)
		resp.Header().Set(IamRequestHeader, rid)

		if rid != "" {
			req.Request.Header.Set(BCSHTTPUserManagerRequestID, rid)
		} else if rid = GetUserManagerRequestID(req.Request.Header); rid == "" {
			rid = GenerateRequestID()
			req.Request.Header.Set(BCSHTTPUserManagerRequestID, rid)
		}
		resp.Header().Set(BCSHTTPUserManagerRequestID, rid)
		// use iam language as um language
		req.Request.Header.Set(BCSHTTPCookieLanugageKey, req.Request.Header.Get("Blueking-Language"))

		user := req.Request.Header.Get(BKHTTPHeaderUser)
		if len(user) == 0 {
			req.Request.Header.Set(BKHTTPHeaderUser, "auth")
		}

		chain.ProcessFilter(req, resp)
		return
	}
}

func (as *AuthService) checkRequestAuthorization(req *http.Request) (bool, error) {
	requestID := req.Header.Get(IamRequestHeader)
	name, pwd, ok := req.BasicAuth()
	if !ok || name != iam.SystemIDIAM {
		blog.Errorf("request have no basic authorization, rid: %s", requestID)
		return false, nil
	}

	// cache
	token, ok := as.tokenCache.Get(IAMTokenKey)
	if ok {
		t, ok1 := token.(string)
		if ok1 && t != "" && t == pwd {
			return true, nil
		}
	}

	token, err := as.iam.GetToken()
	if err != nil {
		blog.Errorf("check request system token failed: %v, requestID: %s", err, requestID)
		return false, err
	}

	// set cache
	as.tokenCache.Set(IAMTokenKey, token, 2*time.Minute)

	if pwd == token {
		return true, nil
	}

	blog.Errorf("request password not match system token, rid: %s", requestID)
	return false, nil
}
