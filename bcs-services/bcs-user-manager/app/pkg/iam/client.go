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

package iam

import (
	"errors"
	"fmt"
	"github.com/TencentBlueKing/iam-go-sdk"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/TencentBlueKing/iam-go-sdk/metric"
)

// PermIAMClient is the interface of IAM backend client, for verify resource permission、get system token and apply for no permission url
type PermIAMClient interface {
	IsAllowedWithoutResource(actionID string, request PermissionRequest) (bool, error)
	IsAllowedWithResource(actionID string, request PermissionRequest, nodes []ResourceNode) (bool, error)
	BatchIsAllowed(actionID string, request PermissionRequest, nodes [][]ResourceNode) (map[string]bool, error)
	IsAllowedWithCache(actionID string, request PermissionRequest, nodes []ResourceNode, ttl time.Duration) (bool, error)
	ResourceMultiActionsAllowed(actions []string, request PermissionRequest, nodes []ResourceNode) (map[string]bool, error)
	BatchResourceMultiActionsAllowed(actions []string, request PermissionRequest, nodes [][]ResourceNode) (map[string]map[string]bool, error)
	GetToken() (string, error)
	IsBasicAuthAllowed(user BkUser) error
	GetApplyURL(request ApplicationRequest, relatedResources []iam.ApplicationRelatedResourceType, user BkUser) (string, error)
}

var (
	// ErrServerNotInit server not init
	ErrServerNotInit = errors.New("iam server not init")
)

// Options for init IAM client
type Options struct {
	// SystemID that bk_bcs used in auth center
	SystemID string
	// AppCode is code for authorize call iam
	AppCode string
	// AppSecret is secret for authorize call iam
	AppSecret string
	// External is false, use GateWayHost
	External bool
	// GateWay host
	GateWayHost string
	// IAM host
	IAMHost string
	// BkiIAM host
	BkiIAMHost string
	// Metrics
	Metric bool
}

func (opt *Options) validate() bool {
	if opt == nil {
		return false
	}

	if opt.SystemID == "" || opt.AppCode == "" || opt.AppSecret == "" {
		return false
	}

	return true
}

type iamClient struct {
	cli *iam.IAM
	opt *Options
}

// NewIamClient will create a iam backend client
func NewIamClient(opt *Options) (PermIAMClient, error) {
	ok := opt.validate()
	if !ok {
		return nil, errors.New("NewIamClient options invalid")
	}

	// register interface metric
	if opt.Metric {
		metric.RegisterMetrics()
	}

	client := &iamClient{
		opt: opt,
	}

	if opt.External {
		client.cli = iam.NewIAM(opt.SystemID, opt.AppCode, opt.AppSecret, opt.IAMHost, opt.BkiIAMHost)
	} else {
		client.cli = iam.NewAPIGatewayIAM(opt.SystemID, opt.AppCode, opt.AppSecret, opt.GateWayHost)
	}

	return client, nil
}

// AttributeKey for attr key type
type AttributeKey string

const (
	// BkIAMPath bk_iam_path
	BkIAMPath AttributeKey = "_bk_iam_path_"
)

// PathBuildIAM interface for build IAMPath
type PathBuildIAM interface {
	BuildIAMPath() string
}

// ClusterScopedResourcePath  build IAMPath for cluster resource
type ClusterScopedResourcePath struct {
	ProjectID string
}

// BuildIAMPath build IAMPath
func (rp ClusterScopedResourcePath) BuildIAMPath() string {
	return fmt.Sprintf("/project,%s/", rp.ProjectID)
}

// NamespaceResourcePath  build IAMPath for namespace resource
type NamespaceResourcePath struct {
	ProjectID     string
	ClusterID     string
	IsClusterPerm bool
}

// BuildIAMPath build IAMPath
func (rp NamespaceResourcePath) BuildIAMPath() string {
	// special case to handle create namespace resource
	if rp.IsClusterPerm {
		return fmt.Sprintf("/project,%s/", rp.ProjectID)
	}
	return fmt.Sprintf("/project,%s/cluster,%s/", rp.ProjectID, rp.ClusterID)
}

// NamespaceScopedResourcePath  build IAMPath for namespace resource
type NamespaceScopedResourcePath struct {
	ProjectID string
	ClusterID string
}

// BuildIAMPath build IAMPath
func (rp NamespaceScopedResourcePath) BuildIAMPath() string {
	return fmt.Sprintf("/project,%s/cluster,%s/", rp.ProjectID, rp.ClusterID)
}

// BkUser user/token
type BkUser struct {
	BkToken    string
	BkUserName string
}

// ResourceNode build resource node
type ResourceNode struct {
	System    string
	RType     string
	RInstance string
	Rp        PathBuildIAM
}

// BuildResourceNode build iam resourceNode
func (rn ResourceNode) BuildResourceNode() iam.ResourceNode {
	attr := map[string]interface{}{}
	bkPath := rn.Rp.BuildIAMPath()

	if bkPath != "" {
		attr[string(BkIAMPath)] = bkPath
	}

	resourceNode := iam.NewResourceNode(rn.System, rn.RType, rn.RInstance, attr)

	return resourceNode
}

// PermissionRequest xxx
type PermissionRequest struct {
	SystemID string
	UserName string
}

func (pr PermissionRequest) validate() bool {
	if pr.SystemID == "" || pr.UserName == "" {
		return false
	}

	return true
}

// MakeRequestWithoutResources make request for no resources
func (pr PermissionRequest) MakeRequestWithoutResources(actionID string) iam.Request {

	subject := iam.Subject{
		Type: "user",
		ID:   pr.UserName,
	}

	action := iam.Action{ID: actionID}
	return iam.NewRequest(pr.SystemID, subject, action, nil)
}

// MakeRequestWithResources make request for signal action signal resource
func (pr PermissionRequest) MakeRequestWithResources(actionID string, nodes []ResourceNode) iam.Request {
	subject := iam.Subject{
		Type: "user",
		ID:   pr.UserName,
	}

	action := iam.Action{ID: actionID}

	iamNodes := []iam.ResourceNode{}
	for i := range nodes {
		iamNodes = append(iamNodes, nodes[i].BuildResourceNode())
	}
	return iam.NewRequest(pr.SystemID, subject, action, iamNodes)
}

// MakeRequestMultiActionResources make request for multi actions and signal resource
func (pr PermissionRequest) MakeRequestMultiActionResources(actions []string, nodes []ResourceNode) iam.MultiActionRequest {
	subject := iam.Subject{
		Type: "user",
		ID:   pr.UserName,
	}

	multiAction := []iam.Action{}
	for i := range actions {
		multiAction = append(multiAction, iam.Action{ID: actions[i]})
	}

	iamNodes := []iam.ResourceNode{}
	for i := range nodes {
		iamNodes = append(iamNodes, nodes[i].BuildResourceNode())
	}

	return iam.NewMultiActionRequest(pr.SystemID, subject, multiAction, iamNodes)
}

// RelatedResourceNode xxx
type RelatedResourceNode struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// RelatedResourceLevel related resource level
type RelatedResourceLevel struct {
	Nodes []RelatedResourceNode
}

// BuildInstance build application resource instance
func (rrl RelatedResourceLevel) BuildInstance() iam.ApplicationResourceInstance {
	nodeList := []iam.ApplicationResourceNode{}

	for i := range rrl.Nodes {
		nodeList = append(nodeList, iam.ApplicationResourceNode{
			Type: rrl.Nodes[i].Type,
			ID:   rrl.Nodes[i].ID,
		})
	}
	return nodeList
}

// RelatedResourceType xxx
type RelatedResourceType struct {
	SystemID string // systemID
	RType    string // system resource type
}

// BuildRelatedResource build level instances
func (rrt RelatedResourceType) BuildRelatedResource(instances []iam.ApplicationResourceInstance) iam.ApplicationRelatedResourceType {
	return iam.ApplicationRelatedResourceType{
		SystemID: rrt.SystemID,
		Type:     rrt.RType,
		// 多种实例拓扑结构, 带层级的实例表示
		Instances: instances,
	}
}

// ApplicationRequest xxx
type ApplicationRequest struct {
	SystemID string
	ActionID string
}

// BuildApplication build application
func (ar ApplicationRequest) BuildApplication(relatedResources []iam.ApplicationRelatedResourceType) iam.Application {
	actions := []iam.ApplicationAction{
		iam.ApplicationAction{
			ID:                   ar.ActionID,
			RelatedResourceTypes: relatedResources,
		},
	}
	return iam.NewApplication(ar.SystemID, actions)
}

// IsAllowedWithoutResource query signal action permission without resource
func (ic *iamClient) IsAllowedWithoutResource(actionID string, request PermissionRequest) (bool, error) {
	if ic == nil {
		return false, ErrServerNotInit
	}

	req := request.MakeRequestWithoutResources(actionID)
	return ic.cli.IsAllowed(req)
}

// IsAllowedWithResource query signal action signal resource permission
func (ic *iamClient) IsAllowedWithResource(actionID string, request PermissionRequest, nodes []ResourceNode) (bool, error) {
	if ic == nil {
		return false, ErrServerNotInit
	}

	req := request.MakeRequestWithResources(actionID, nodes)
	fmt.Printf("%+v", req)
	return ic.cli.IsAllowed(req)
}

// IsAllowedWithCache for permission cache
func (ic *iamClient) IsAllowedWithCache(actionID string, request PermissionRequest, nodes []ResourceNode, ttl time.Duration) (bool, error) {
	if ic == nil {
		return false, ErrServerNotInit
	}

	var req iam.Request

	if len(nodes) == 0 {
		req = request.MakeRequestWithoutResources(actionID)
	} else {
		req = request.MakeRequestWithResources(actionID, nodes)
	}

	return ic.cli.IsAllowedWithCache(req, ttl)
}

// BatchIsAllowed batch resource check permission, signalAction multiResources
// resources []iam.ResourceNode: len=1, return node.ID; len > 1, node.Type:node.ID/node.Type:node.ID
func (ic *iamClient) BatchIsAllowed(actionID string, request PermissionRequest, nodes [][]ResourceNode) (map[string]bool, error) {
	if ic == nil {
		return nil, ErrServerNotInit
	}

	req := request.MakeRequestWithoutResources(actionID)

	resourceList := []iam.Resources{}
	for _, nodeList := range nodes {
		iamNodes := []iam.ResourceNode{}
		for i := range nodeList {
			iamNodes = append(iamNodes, nodeList[i].BuildResourceNode())
		}

		resourceList = append(resourceList, iamNodes)
	}

	return ic.cli.BatchIsAllowed(req, resourceList)
}

// ResourceMultiActionsAllowed for multiActions signalResource
func (ic *iamClient) ResourceMultiActionsAllowed(actions []string, request PermissionRequest, nodes []ResourceNode) (map[string]bool, error) {
	if ic == nil {
		return nil, ErrServerNotInit
	}

	req := request.MakeRequestMultiActionResources(actions, nodes)
	return ic.cli.ResourceMultiActionsAllowed(req)
}

// BatchResourceMultiActionsAllowed will check the permissions of batch-resource with multi-actions, multi actions and multi resource
// resource action isAllow
func (ic *iamClient) BatchResourceMultiActionsAllowed(actions []string, request PermissionRequest, nodes [][]ResourceNode) (map[string]map[string]bool, error) {
	if ic == nil {
		return nil, ErrServerNotInit
	}

	multiReq := request.MakeRequestMultiActionResources(actions, nil)

	resourceList := []iam.Resources{}
	for _, nodeList := range nodes {
		iamNodes := []iam.ResourceNode{}
		for i := range nodeList {
			iamNodes = append(iamNodes, nodeList[i].BuildResourceNode())
		}

		resourceList = append(resourceList, iamNodes)
	}

	return ic.cli.BatchResourceMultiActionsAllowed(multiReq, resourceList)
}

// GetToken will get the token of system
func (ic *iamClient) GetToken() (string, error) {
	if ic == nil {
		return "", ErrServerNotInit
	}

	return ic.cli.GetToken()
}

// check iam callback request auth
func (ic *iamClient) IsBasicAuthAllowed(user BkUser) error {
	if ic == nil {
		return ErrServerNotInit
	}

	return ic.cli.IsBasicAuthAllowed(user.BkUserName, user.BkToken)
}

// GetApplyURL will generate the application URL
func (ic *iamClient) GetApplyURL(request ApplicationRequest, relatedResources []iam.ApplicationRelatedResourceType, user BkUser) (string, error) {
	if ic == nil {
		return "", ErrServerNotInit
	}

	application := request.BuildApplication(relatedResources)

	url, err := ic.cli.GetApplyURL(application, user.BkToken, user.BkUserName)
	if err != nil {
		blog.Errorf("iam generate apply url failed: %s", err)
		return IamAppURL, nil
	}

	return url, nil
}
