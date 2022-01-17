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
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/TencentBlueKing/iam-go-sdk"
	"github.com/TencentBlueKing/iam-go-sdk/logger"
	"github.com/TencentBlueKing/iam-go-sdk/metric"
	"github.com/sirupsen/logrus"
)

// PermClient interface for IAM backend client
type PermClient interface {
	IsAllowedWithoutResource(actionID string, request PermissionRequest, cache bool) (bool, error)
	IsAllowedWithResource(actionID string, request PermissionRequest, nodes []ResourceNode, cache bool) (bool, error)
	BatchResourceIsAllowed(actionID string, request PermissionRequest, nodes [][]ResourceNode) (map[string]bool, error)
	ResourceMultiActionsAllowed(actions []string, request PermissionRequest, nodes []ResourceNode) (map[string]bool, error)
	BatchResourceMultiActionsAllowed(actions []string, request PermissionRequest, nodes [][]ResourceNode) (map[string]map[string]bool, error)
	GetToken() (string, error)
	IsBasicAuthAllowed(user BkUser) error
	GetApplyURL(request ApplicationRequest, relatedResources []ApplicationAction, user BkUser) (string, error)
}

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
	// Debug
	Debug bool
}

func (opt *Options) validate() error {
	if opt == nil {
		return ErrServerNotInit
	}

	if opt.SystemID == "" || opt.AppCode == "" || opt.AppSecret == "" {
		return fmt.Errorf("systemID/AppCode/AppSecret required")
	}

	if !opt.External && opt.GateWayHost == "" {
		return fmt.Errorf("BKAPIGatewayHost required when UseGateway flag set to true")
	}
	if opt.External && (opt.BkiIAMHost == "" || opt.IAMHost == "") {
		return fmt.Errorf("BKIAMHost and BKPAASHost required when UseGateway flag set to false")
	}

	return nil
}

type iamClient struct {
	cli *iam.IAM
	opt *Options
}

func setIAMLogger(opt *Options) {
	defaultLogLevel := logrus.ErrorLevel
	if opt.Debug {
		defaultLogLevel = logrus.DebugLevel
	}

	log := &logrus.Logger{
		Out:          os.Stderr,
		Formatter:    new(logrus.TextFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        defaultLogLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}

	logger.SetLogger(log)
}

// NewIamClient create iam backend client
func NewIamClient(opt *Options) (PermClient, error) {
	err := opt.validate()
	if err != nil {
		return nil, fmt.Errorf("NewIamClient options invalid: %v", err)
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

	// init IAM logger
	setIAMLogger(opt)

	return client, nil
}

// BkUser user/token
type BkUser struct {
	BkToken    string
	BkUserName string
}

// IsAllowedWithoutResource query signal action permission without resource(cache use withoutResource or managerPerm)
func (ic *iamClient) IsAllowedWithoutResource(actionID string, request PermissionRequest, cache bool) (bool, error) {
	if ic == nil {
		return false, ErrServerNotInit
	}

	req := request.MakeRequestWithoutResources(actionID)
	if cache {
		return ic.cli.IsAllowedWithCache(req, defaultAllowTTL)
	}
	return ic.cli.IsAllowed(req)
}

// IsAllowedWithResource query signal action signal resource permission(cache use withoutResource or managerPerm)
func (ic *iamClient) IsAllowedWithResource(actionID string, request PermissionRequest, nodes []ResourceNode, cache bool) (bool, error) {
	if ic == nil {
		return false, ErrServerNotInit
	}

	req := request.MakeRequestWithResources(actionID, nodes)
	if cache {
		return ic.cli.IsAllowedWithCache(req, defaultAllowTTL)
	}
	return ic.cli.IsAllowed(req)
}

// BatchResourceIsAllowed batch resource check permission, signalAction multiResources
// resources []iam.ResourceNode: len=1, return node.ID; len > 1, node.Type:node.ID/node.Type:node.ID
func (ic *iamClient) BatchResourceIsAllowed(actionID string, request PermissionRequest, nodes [][]ResourceNode) (map[string]bool, error) {
	if ic == nil {
		return nil, ErrServerNotInit
	}

	req := request.MakeRequestWithoutResources(actionID)

	resourceList := make([]iam.Resources, 0)
	for _, nodeList := range nodes {
		iamNodes := make([]iam.ResourceNode, 0)
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

	resourceList := make([]iam.Resources, 0)
	for _, nodeList := range nodes {
		iamNodes := make([]iam.ResourceNode, 0)
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
func (ic *iamClient) GetApplyURL(request ApplicationRequest, relatedResources []ApplicationAction, user BkUser) (string, error) {
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
