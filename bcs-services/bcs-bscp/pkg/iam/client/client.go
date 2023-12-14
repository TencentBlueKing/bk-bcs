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

// Package client NOTES
package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sdk/operator"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/rest/client"
)

const (
	codeNotFound = 1901404
)

var (
	// ErrNotFound is iam return not exist resource's error.
	ErrNotFound = errors.New("iam not found")
)

// Client is auth center client.
type Client struct {
	config *Config
	// http client instance
	client rest.ClientInterface
	// http header info
	basicHeader http.Header
}

// NewClient new iam client.
func NewClient(cfg *Config, reg prometheus.Registerer) (*Client, error) {
	restCli, err := client.NewClient(cfg.TLS)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: restCli,
		Discover: &iamDiscovery{
			servers: cfg.Address,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	header.Set("Accept", "application/json")
	header.Set(bkapiAuthHeader, fmt.Sprintf("{\"bk_app_code\":\"%s\", \"bk_app_secret\":\"%s\"}",
		cfg.AppCode, cfg.AppSecret))

	cli := &Client{
		config:      cfg,
		client:      rest.NewClient(c, "/"),
		basicHeader: header,
	}
	return cli, nil
}

// RegisterSystem register a system in IAM
func (c *Client) RegisterSystem(ctx context.Context, sys System) error {
	resp := new(BaseResponse)
	result := c.client.Post().
		SubResourcef("/api/v1/model/systems").
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(sys).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("register system failed, code: %d, msg:%s", resp.Code, resp.Message),
		}
	}

	return nil
}

// GetSystemInfo get a system info from IAM
// if fields is empty, find all system info
func (c *Client) GetSystemInfo(ctx context.Context, fields []SystemQueryField) (*SystemResp, error) {
	resp := new(SystemResp)
	fieldsStr := ""
	if len(fields) > 0 {
		fieldArr := make([]string, len(fields))
		for idx, field := range fields {
			fieldArr[idx] = string(field)
		}
		fieldsStr = strings.Join(fieldArr, ",")
	}

	result := c.client.Get().
		SubResourcef("/api/v1/model/systems/%s/query", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		WithParam("fields", fieldsStr).
		Body(nil).Do()
	err := result.Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		if resp.Code == codeNotFound {
			return resp, ErrNotFound
		}
		return nil, &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("get system info failed, code: %d, msg:%s", resp.Code, resp.Message),
		}
	}

	return resp, nil
}

// UpdateSystemConfig update system config in IAM
// Note: can only update provider_config.host field.
func (c *Client) UpdateSystemConfig(ctx context.Context, config *SysConfig) error {
	sys := new(System)
	config.Auth = "basic"
	sys.ProviderConfig = config
	resp := new(BaseResponse)
	result := c.client.Put().
		SubResourcef("/api/v1/model/systems/%s", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(sys).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("update system config failed, code: %d, msg:%s", resp.Code, resp.Message),
		}
	}

	return nil
}

// RegisterResourcesTypes register resource types in IAM
func (c *Client) RegisterResourcesTypes(ctx context.Context, resTypes []ResourceType) error {
	resp := new(BaseResponse)
	result := c.client.Post().
		SubResourcef("/api/v1/model/systems/%s/resource-types", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(resTypes).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("register system failed, code: %d, msg:%s", resp.Code, resp.Message),
		}
	}

	return nil

}

// UpdateResourcesType update resource type in IAM
func (c *Client) UpdateResourcesType(ctx context.Context, resType ResourceType) error {
	resp := new(BaseResponse)
	result := c.client.Put().
		SubResourcef("/api/v1/model/systems/%s/resource-types/%s", c.config.SystemID, resType.ID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(resType).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("udpate resource type %s failed, code: %d, msg:%s", resType.ID, resp.Code, resp.Message),
		}
	}

	return nil
}

// DeleteResourcesTypes delete resource types in IAM
func (c *Client) DeleteResourcesTypes(ctx context.Context, resTypeIDs []TypeID) error {

	ids := make([]struct {
		ID TypeID `json:"id"`
	}, len(resTypeIDs))
	for idx := range resTypeIDs {
		ids[idx].ID = resTypeIDs[idx]
	}

	resp := new(BaseResponse)
	result := c.client.Delete().
		SubResourcef("/api/v1/model/systems/%s/resource-types", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(ids).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("delete resource type %v failed, code: %d, msg:%s", resTypeIDs, resp.Code, resp.Message),
		}
	}

	return nil
}

// RegisterActions register actions in IAM
func (c *Client) RegisterActions(ctx context.Context, actions []ResourceAction) error {

	resp := new(BaseResponse)
	result := c.client.Post().
		SubResourcef("/api/v1/model/systems/%s/actions", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(actions).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("add resource actions %v failed, code: %d, msg:%s", actions, resp.Code, resp.Message),
		}
	}

	return nil
}

// UpdateAction update action in IAM
func (c *Client) UpdateAction(ctx context.Context, action ResourceAction) error {

	resp := new(BaseResponse)
	result := c.client.Put().
		SubResourcef("/api/v1/model/systems/%s/actions/%s", c.config.SystemID, action.ID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(action).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("udpate resource action %v failed, code: %d, msg:%s", action, resp.Code, resp.Message),
		}
	}

	return nil
}

// DeleteActions delete actions in IAM
func (c *Client) DeleteActions(ctx context.Context, actionIDs []ActionID) error {
	ids := make([]struct {
		ID ActionID `json:"id"`
	}, len(actionIDs))
	for idx := range actionIDs {
		ids[idx].ID = actionIDs[idx]
	}

	resp := new(BaseResponse)
	result := c.client.Delete().
		SubResourcef("/api/v1/model/systems/%s/actions", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(ids).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("delete resource actions %v failed, code: %d, msg:%s", actionIDs, resp.Code, resp.Message),
		}
	}

	return nil
}

// RegisterActionGroups register action groups in IAM
func (c *Client) RegisterActionGroups(ctx context.Context, actionGroups []ActionGroup) error {

	resp := new(BaseResponse)
	result := c.client.Post().
		SubResourcef("/api/v1/model/systems/%s/configs/action_groups", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(actionGroups).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("register action groups %v failed, code: %d, msg:%s", actionGroups, resp.Code, resp.Message),
		}
	}

	return nil
}

// UpdateActionGroups update action groups in IAM
func (c *Client) UpdateActionGroups(ctx context.Context, actionGroups []ActionGroup) error {

	resp := new(BaseResponse)
	result := c.client.Put().
		SubResourcef("/api/v1/model/systems/%s/configs/action_groups", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(actionGroups).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("update action groups %v failed, code: %d, msg:%s", actionGroups, resp.Code, resp.Message),
		}
	}

	return nil
}

// RegisterInstanceSelections register instance selections.
func (c *Client) RegisterInstanceSelections(ctx context.Context, instanceSelections []InstanceSelection) error {

	resp := new(BaseResponse)
	result := c.client.Post().
		SubResourcef("/api/v1/model/systems/%s/instance-selections", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(instanceSelections).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("add instance selections %v failed, code: %d, msg:%s", instanceSelections,
				resp.Code, resp.Message),
		}
	}

	return nil
}

// UpdateInstanceSelection update instance selection in IAM
func (c *Client) UpdateInstanceSelection(ctx context.Context, instanceSelection InstanceSelection) error {

	resp := new(BaseResponse)
	result := c.client.Put().
		SubResourcef("/api/v1/model/systems/%s/instance-selections/%s", c.config.SystemID, instanceSelection.ID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(instanceSelection).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("udpate instance selections %v failed, code: %d, msg:%s", instanceSelection,
				resp.Code, resp.Message),
		}
	}

	return nil
}

// DeleteInstanceSelections delete instance selections in IAM
func (c *Client) DeleteInstanceSelections(ctx context.Context, instanceSelectionIDs []InstanceSelectionID) error {
	ids := make([]struct {
		ID InstanceSelectionID `json:"id"`
	}, len(instanceSelectionIDs))
	for idx := range instanceSelectionIDs {
		ids[idx].ID = instanceSelectionIDs[idx]
	}

	resp := new(BaseResponse)
	result := c.client.Delete().
		SubResourcef("/api/v1/model/systems/%s/instance-selections", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(ids).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("delete instance selections %v failed, code: %d, msg:%s", instanceSelectionIDs,
				resp.Code, resp.Message),
		}
	}

	return nil
}

// RegisterResourceCreatorActions regitser resource creator actions in IAM
func (c *Client) RegisterResourceCreatorActions(ctx context.Context, resourceCreatorActions ResourceCreatorActions,
) error {

	resp := new(BaseResponse)
	result := c.client.Post().
		SubResourcef("/api/v1/model/systems/%s/configs/resource_creator_actions", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(resourceCreatorActions).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("register resource creator actions %v failed, code: %d, msg:%s",
				resourceCreatorActions, resp.Code, resp.Message),
		}
	}

	return nil
}

// UpdateResourceCreatorActions update resource creator actions in IAM
func (c *Client) UpdateResourceCreatorActions(ctx context.Context, resourceCreatorActions ResourceCreatorActions,
) error {

	resp := new(BaseResponse)
	result := c.client.Put().
		SubResourcef("/api/v1/model/systems/%s/configs/resource_creator_actions", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(resourceCreatorActions).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("update resource creator actions %v failed, code: %d, msg:%s",
				resourceCreatorActions, resp.Code, resp.Message),
		}
	}

	return nil
}

// RegisterCommonActions register common actions in IAM
func (c *Client) RegisterCommonActions(ctx context.Context, commonActions []CommonAction) error {
	resp := new(BaseResponse)
	result := c.client.Post().
		SubResourcef("/api/v1/model/systems/%s/configs/common_actions", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(commonActions).Do()

	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("register common actions %v failed, code: %d, msg: %s", commonActions, resp.Code,
				resp.Message),
		}
	}

	return nil
}

// UpdateCommonActions update common actions in IAM
func (c *Client) UpdateCommonActions(ctx context.Context, commonActions []CommonAction) error {
	resp := new(BaseResponse)
	result := c.client.Put().
		SubResourcef("/api/v1/model/systems/%s/configs/common_actions", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(commonActions).Do()

	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("update common actions %v failed, code: %d, msg: %s", commonActions, resp.Code,
				resp.Message),
		}
	}

	return nil
}

// DeleteActionPolicies delete action policies in IAM
func (c *Client) DeleteActionPolicies(ctx context.Context, actionID ActionID) error {
	resp := new(BaseResponse)
	result := c.client.Delete().
		SubResourcef("/api/v1/model/systems/%s/actions/%s/policies", c.config.SystemID, actionID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("delete action %s policies failed, code: %d, msg: %s", actionID, resp.Code, resp.Message),
		}
	}

	return nil
}

// ListPolicies list iam policies
func (c *Client) ListPolicies(ctx context.Context, params *ListPoliciesParams) (*ListPoliciesData, error) {
	parsedParams := map[string]string{"action_id": string(params.ActionID)}
	if params.Page != 0 {
		parsedParams["page"] = strconv.FormatInt(params.Page, 10)
	}
	if params.PageSize != 0 {
		parsedParams["page_size"] = strconv.FormatInt(params.PageSize, 10)
	}
	if params.Timestamp != 0 {
		parsedParams["timestamp"] = strconv.FormatInt(params.Timestamp, 10)
	}

	resp := new(ListPoliciesResp)
	result := c.client.Get().
		SubResourcef("/api/v1/systems/%s/policies", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		WithParams(parsedParams).
		Body(nil).Do()

	err := result.Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("get system info failed, code: %d, msg:%s", resp.Code, resp.Message),
		}
	}

	return resp.Data, nil
}

// GetSystemToken get system token from iam, used to validate if request is from iam
func (c *Client) GetSystemToken(ctx context.Context) (string, error) {
	resp := new(struct {
		BaseResponse
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	})
	result := c.client.Get().
		SubResourcef("/api/v1/model/systems/%s/token", c.config.SystemID).
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(nil).Do()
	err := result.Into(resp)
	if err != nil {
		return "", err
	}

	if resp.Code != 0 {
		return "", &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("get system info failed, code: %d, msg:%s", resp.Code, resp.Message),
		}
	}

	return resp.Data.Token, nil
}

// GetUserPolicy get a user's policy with a action and resources
func (c *Client) GetUserPolicy(ctx context.Context, opt *GetPolicyOption) (*operator.Policy, error) {
	resp := new(GetPolicyResp)

	// iam requires resources to be set
	if opt.Resources == nil {
		opt.Resources = make([]Resource, 0)
	}

	result := c.client.Post().
		SubResourcef("/api/v1/policy/query").
		WithContext(ctx).
		WithHeaders(c.cloneHeader(ctx)).
		Body(opt).
		Do()

	err := result.Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("get system info failed, code: %d, msg:%s", resp.Code, resp.Message),
		}
	}

	return resp.Data, nil
}

// ListUserPolicies get a user's policy with multiple actions and resources
func (c *Client) ListUserPolicies(ctx context.Context, opts *ListPolicyOptions) (
	[]*ActionPolicy, error) {

	resp := new(ListPolicyResp)

	// iam requires resources to be set
	if opts.Resources == nil {
		opts.Resources = make([]Resource, 0)
	}

	result := c.client.Post().
		SubResourcef("/api/v1/policy/query_by_actions").
		WithContext(ctx).
		WithHeaders(c.cloneHeader(ctx)).
		Body(opts).
		Do()

	err := result.Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("list user policies failed, code: %d, msg: %s", resp.Code, resp.Message),
		}
	}

	return resp.Data, nil
}

// GetUserPolicyByExtRes get a user's policy by external resource.
func (c *Client) GetUserPolicyByExtRes(ctx context.Context, opts *GetPolicyByExtResOption) (*GetPolicyByExtResResult,
	error) {

	resp := new(GetPolicyByExtResResp)

	// iam requires resources to be set
	if opts.Resources == nil {
		opts.Resources = make([]Resource, 0)
	}

	result := c.client.Post().
		SubResourcef("/api/v1/policy/query_by_ext_resources").
		WithContext(ctx).
		WithHeaders(c.cloneHeader(ctx)).
		Body(opts).
		Do()

	err := result.Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("get policy by external resource failed, code: %d, msg: %s", resp.Code, resp.Message),
		}
	}

	return resp.Data, nil
}

func (c *Client) cloneHeader(ctx context.Context) http.Header {
	h := http.Header{}
	rid, ok := ctx.Value(constant.RidKey).(string)
	if ok {
		h.Set(RequestIDHeader, rid)
	}

	for key := range c.basicHeader {
		h.Set(key, c.basicHeader.Get(key))
	}
	return h
}

// GrantResourceCreatorAction grant resource creator action in IAM
func (c *Client) GrantResourceCreatorAction(ctx context.Context, opt GrantResourceCreatorActionOption) error {
	resp := new(BaseResponse)
	result := c.client.Post().
		SubResourcef("api/v1/open/authorization/resource_creator_action").
		WithContext(ctx).
		WithHeaders(c.basicHeader).
		Body(opt).Do()

	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("grant resource creator action %v failed, code: %d, msg: %s",
				opt, resp.Code, resp.Message),
		}
	}

	return nil
}
