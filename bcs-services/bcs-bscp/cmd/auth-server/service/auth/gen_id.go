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

package auth

import (
	"fmt"
	"strconv"

	"github.com/TencentBlueKing/iam-go-sdk"
	bkiam "github.com/TencentBlueKing/iam-go-sdk"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sys"
)

// bizIDAssembleSymbol used to assemble biz_id and resource id's symbol, used in app id generation.
// nolint: unused
const bizIDAssembleSymbol = "-"

var (
	// dummyIAMUser
	dummyIAMUser = bkiam.NewSubject("user", "")
)

// genSkipResource generate iam resource for resource, using skip action.
func genSkipResource(_ *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return sys.Skip, make([]client.Resource, 0), nil
}

// genBizResource
func genBizResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Business,
		ID:     strconv.FormatUint(uint64(a.BizID), 10),
	}

	switch a.Basic.Action {
	case meta.FindBusinessResource:
		// create app is related to cmdb business resource
		return sys.BusinessViewResource, []client.Resource{bizRes}, nil
	default:
		return "", nil, fmt.Errorf("unsupported bscp action: %s", a.Basic.Action)
	}
}

func genBizIAMResource(a *meta.ResourceAttribute) (*bkiam.Request, error) {
	iamReq := bkiam.Request{
		System:  sys.SystemIDBSCP,
		Subject: dummyIAMUser,
		Resources: []iam.ResourceNode{
			{
				System:    sys.SystemIDCMDB,
				Type:      string(sys.Business),
				ID:        strconv.FormatUint(uint64(a.BizID), 10),
				Attribute: map[string]interface{}{},
			},
		},
	}

	switch a.Basic.Action {
	case meta.FindBusinessResource:
		// create app is related to cmdb business resource
		iamReq.Action = bkiam.NewAction(string(sys.BusinessViewResource))
	default:
		return nil, fmt.Errorf("unsupported bscp action: %s", a.Basic.Action)
	}
	return &iamReq, nil
}

func genAppIAMResource(a *meta.ResourceAttribute) (*bkiam.Request, error) {
	iamReq := bkiam.Request{
		System:  sys.SystemIDBSCP,
		Subject: dummyIAMUser,
		Resources: []iam.ResourceNode{
			{
				System:    sys.SystemIDBSCP,
				Type:      string(sys.Business),
				ID:        strconv.FormatUint(uint64(a.ResourceID), 10),
				Attribute: map[string]interface{}{},
			},
		},
	}

	switch a.Basic.Action {
	case meta.Create:
		iamReq.Action = bkiam.NewAction(string(sys.AppCreate))
	case meta.View:
		iamReq.Action = bkiam.NewAction(string(sys.AppView))
	case meta.Update:
		iamReq.Action = bkiam.NewAction(string(sys.AppEdit))
	case meta.Publish:
		iamReq.Action = bkiam.NewAction(string(sys.ReleasePublish))
	case meta.GenerateRelease:
		iamReq.Action = bkiam.NewAction(string(sys.ReleaseGenerate))
	case meta.Delete:
		iamReq.Action = bkiam.NewAction(string(sys.AppDelete))
	default:
		return nil, fmt.Errorf("unsupported bscp action: %s", a.Basic.Action)
	}
	return &iamReq, nil
}

func genCredIAMResource(a *meta.ResourceAttribute) (*bkiam.Request, error) {
	iamReq := bkiam.Request{
		System:  sys.SystemIDBSCP,
		Subject: dummyIAMUser,
		Resources: []iam.ResourceNode{
			{
				System:    sys.SystemIDCMDB,
				Type:      string(sys.Business),
				ID:        strconv.FormatUint(uint64(a.BizID), 10),
				Attribute: map[string]interface{}{},
			},
		},
	}

	switch a.Basic.Action {
	case meta.View:
		iamReq.Action = bkiam.NewAction(string(sys.CredentialView))
	case meta.Manage:
		iamReq.Action = bkiam.NewAction(string(sys.CredentialManage))
	default:
		return nil, fmt.Errorf("unsupported bscp action: %s", a.Basic.Action)
	}
	return &iamReq, nil
}

func genBizIAMApplication(a *meta.ResourceAttribute) (bkiam.ApplicationAction, error) {
	action := bkiam.ApplicationAction{
		RelatedResourceTypes: []bkiam.ApplicationRelatedResourceType{
			{
				SystemID: sys.SystemIDCMDB,
				Type:     string(sys.Business),
				Instances: []bkiam.ApplicationResourceInstance{
					[]iam.ApplicationResourceNode{
						{
							Type: string(sys.Business),
							ID:   strconv.FormatUint(uint64(a.BizID), 10),
						},
					},
				},
			},
		},
	}

	switch a.Basic.Action {
	case meta.FindBusinessResource:
		// create app is related to cmdb business resource
		action.ID = string(sys.BusinessViewResource)
	default:
		return action, fmt.Errorf("unsupported bscp action: %s", a.Basic.Action)
	}
	return action, nil
}

func genAppIAMApplication(a *meta.ResourceAttribute) (bkiam.ApplicationAction, error) {
	action := bkiam.ApplicationAction{}
	resourceNodes := []iam.ApplicationResourceNode{}
	resourceNodeBiz := iam.ApplicationResourceNode{
		Type: string(sys.Business),
		ID:   strconv.FormatUint(uint64(a.BizID), 10),
	}
	resourceNodeApp := iam.ApplicationResourceNode{
		Type: string(sys.Application),
		ID:   strconv.FormatUint(uint64(a.ResourceID), 10),
	}

	switch a.Basic.Action {
	case meta.Create:
		action.ID = string(sys.AppCreate)
		resourceNodes = append(resourceNodes, resourceNodeBiz)
		action.RelatedResourceTypes = []bkiam.ApplicationRelatedResourceType{
			{
				SystemID:  sys.SystemIDCMDB,
				Type:      string(sys.Business),
				Instances: []bkiam.ApplicationResourceInstance{resourceNodes},
			},
		}
		return action, nil
	case meta.View:
		action.ID = string(sys.AppView)
		resourceNodes = append(resourceNodes, resourceNodeBiz, resourceNodeApp)
	case meta.Update:
		action.ID = string(sys.AppEdit)
		resourceNodes = append(resourceNodes, resourceNodeBiz, resourceNodeApp)
	case meta.Delete:
		action.ID = string(sys.AppDelete)
		resourceNodes = append(resourceNodes, resourceNodeBiz, resourceNodeApp)
	case meta.Publish:
		action.ID = string(sys.ReleasePublish)
		resourceNodes = append(resourceNodes, resourceNodeBiz, resourceNodeApp)
	case meta.GenerateRelease:
		action.ID = string(sys.ReleaseGenerate)
		resourceNodes = append(resourceNodes, resourceNodeBiz, resourceNodeApp)
	default:
		return action, fmt.Errorf("unsupported bscp action: %s", a.Basic.Action)
	}
	action.RelatedResourceTypes = []bkiam.ApplicationRelatedResourceType{
		{
			SystemID:  sys.SystemIDBSCP,
			Type:      string(sys.Application),
			Instances: []bkiam.ApplicationResourceInstance{resourceNodes},
		},
	}
	return action, nil
}
func genCredIAMApplication(a *meta.ResourceAttribute) (bkiam.ApplicationAction, error) {
	action := bkiam.ApplicationAction{
		RelatedResourceTypes: []bkiam.ApplicationRelatedResourceType{{
			SystemID: sys.SystemIDCMDB,
			Type:     string(sys.Business),
			Instances: []bkiam.ApplicationResourceInstance{
				[]iam.ApplicationResourceNode{{
					Type: string(sys.Business),
					ID:   strconv.FormatUint(uint64(a.BizID), 10),
				}},
			},
		}},
	}

	switch a.Basic.Action {
	case meta.View:
		action.ID = string(sys.CredentialView)
	case meta.Manage:
		action.ID = string(sys.CredentialManage)
	default:
		return action, fmt.Errorf("unsupported bscp action: %s", a.Basic.Action)
	}
	return action, nil
}

// genAppResource generate application related iam resource.
func genAppResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Business,
		ID:     strconv.FormatUint(uint64(a.BizID), 10),
	}

	appRes := client.Resource{
		System: sys.SystemIDBSCP,
		Type:   sys.Application,
		ID:     strconv.FormatUint(uint64(a.ResourceID), 10),
		// can be authorized based on business
		Attribute: map[string]interface{}{
			client.IamPathKey: []string{fmt.Sprintf("/%s,%d/", sys.Business, a.BizID)},
		},
	}

	switch a.Basic.Action {
	case meta.Create:
		// create app is related to cmdb business resource
		return sys.AppCreate, []client.Resource{bizRes}, nil
	case meta.Update:
		// update app is related to bscp application resource
		return sys.AppEdit, []client.Resource{appRes}, nil
	case meta.Delete:
		// delete app is related to bscp application resource
		return sys.AppDelete, []client.Resource{appRes}, nil
	case meta.View:
		// view app is related to bscp application resource
		return sys.AppView, []client.Resource{appRes}, nil
	case meta.GenerateRelease:
		// generate release is related to bscp application resource
		return sys.ReleaseGenerate, []client.Resource{appRes}, nil
	case meta.Publish:
		// publish release is related to bscp application resource
		return sys.ReleasePublish, []client.Resource{appRes}, nil
	case meta.Find:
		// find app is related to cmdb business resource, using view biz action
		return sys.BusinessViewResource, []client.Resource{bizRes}, nil
	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp action: %s", a.Basic.Action))
	}
}

func genCredResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Business,
		ID:     strconv.FormatUint(uint64(a.BizID), 10),
	}

	switch a.Basic.Action {
	case meta.View:
		return sys.CredentialView, []client.Resource{bizRes}, nil
	case meta.Manage:
		return sys.CredentialManage, []client.Resource{bizRes}, nil
	default:
		return "", nil, fmt.Errorf("unsupported bscp action: %s", a.Basic.Action)
	}
}

// genCommitResource generate commit related iam resource.
// nolint: unused
func genCommitResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Business,
		ID:     strconv.FormatUint(uint64(a.BizID), 10),
	}

	appRes := client.Resource{
		System: sys.SystemIDBSCP,
		Type:   sys.Application,
		ID: strconv.FormatUint(uint64(a.BizID), 10) + bizIDAssembleSymbol +
			strconv.FormatUint(uint64(a.ResourceID), 10),
		// can be authorized based on business
		Attribute: map[string]interface{}{
			client.IamPathKey: []string{fmt.Sprintf("/%s,%d/", sys.Business, a.BizID)},
		},
	}

	switch a.Basic.Action {
	case meta.Create:
		// create commit is related to bscp application resource, using app edit action
		return sys.AppEdit, []client.Resource{appRes}, nil
	case meta.Find:
		// find commit is related to cmdb business resource, using view biz action
		return sys.BusinessViewResource, []client.Resource{bizRes}, nil
	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp action: %s", a.Basic.Action))
	}
}

// genConfigItemResource generate config item related iam resource.
// nolint: unused
func genConfigItemResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Business,
		ID:     strconv.FormatUint(uint64(a.BizID), 10),
	}

	appRes := client.Resource{
		System: sys.SystemIDBSCP,
		Type:   sys.Application,
		ID: strconv.FormatUint(uint64(a.BizID), 10) + bizIDAssembleSymbol +
			strconv.FormatUint(uint64(a.ResourceID), 10),
		// can be authorized based on business
		Attribute: map[string]interface{}{
			client.IamPathKey: []string{fmt.Sprintf("/%s,%d/", sys.Business, a.BizID)},
		},
	}

	switch a.Basic.Action {
	case meta.Create:
		// create config item is related to bscp application resource, using app edit action
		return sys.AppEdit, []client.Resource{appRes}, nil
	case meta.Update:
		// update config item is related to bscp application resource, using app edit action
		return sys.AppEdit, []client.Resource{appRes}, nil
	case meta.Delete:
		// delete config item is related to bscp application resource, using app edit action
		return sys.AppEdit, []client.Resource{appRes}, nil
	case meta.Find:
		// find config item is related to cmdb business resource, using view biz action
		return sys.BusinessViewResource, []client.Resource{bizRes}, nil
	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp action: %s", a.Basic.Action))
	}
}

// genContentResource generate content related iam resource.
// nolint: unused
func genContentResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Business,
		ID:     strconv.FormatUint(uint64(a.BizID), 10),
	}

	appRes := client.Resource{
		System: sys.SystemIDBSCP,
		Type:   sys.Application,
		ID: strconv.FormatUint(uint64(a.BizID), 10) + bizIDAssembleSymbol +
			strconv.FormatUint(uint64(a.ResourceID), 10),
		// can be authorized based on business
		Attribute: map[string]interface{}{
			client.IamPathKey: []string{fmt.Sprintf("/%s,%d/", sys.Business, a.BizID)},
		},
	}

	switch a.Basic.Action {
	case meta.Create:
		// create content is related to bscp application resource, using app edit action
		return sys.AppEdit, []client.Resource{appRes}, nil
	case meta.Find:
		// find content is related to cmdb business resource, using view biz action
		return sys.BusinessViewResource, []client.Resource{bizRes}, nil
	case meta.Upload:
		// upload content is related to bscp application resource, using app edit action
		return sys.AppEdit, []client.Resource{appRes}, nil
	case meta.Download:
		// download content is related to bscp application resource, using app edit action
		return sys.AppEdit, []client.Resource{appRes}, nil
	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp action: %s", a.Basic.Action))
	}
}

// genCRInstanceResource generate current released instance related iam resource.
// nolint: unused
func genCRInstanceResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Business,
		ID:     strconv.FormatUint(uint64(a.BizID), 10),
	}

	appRes := client.Resource{
		System: sys.SystemIDBSCP,
		Type:   sys.Application,
		ID: strconv.FormatUint(uint64(a.BizID), 10) + bizIDAssembleSymbol +
			strconv.FormatUint(uint64(a.ResourceID), 10),
		// can be authorized based on business
		Attribute: map[string]interface{}{
			client.IamPathKey: []string{fmt.Sprintf("/%s,%d/", sys.Business, a.BizID)},
		},
	}

	switch a.Basic.Action {
	case meta.Create:
		// create current released instance is related to bscp application resource, using app edit action
		return sys.AppEdit, []client.Resource{appRes}, nil
	case meta.Delete:
		// delete current released instance is related to bscp application resource, using app edit action
		return sys.AppEdit, []client.Resource{appRes}, nil
	case meta.Find:
		// find current released instance is related to cmdb business resource, using view biz action
		return sys.BusinessViewResource, []client.Resource{bizRes}, nil
	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp action: %s", a.Basic.Action))
	}
}

// genReleaseRes generate release related iam resource.
// nolint: unused
func genReleaseRes(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Business,
		ID:     strconv.FormatUint(uint64(a.BizID), 10),
	}

	appRes := client.Resource{
		System: sys.SystemIDBSCP,
		Type:   sys.Application,
		ID: strconv.FormatUint(uint64(a.BizID), 10) + bizIDAssembleSymbol +
			strconv.FormatUint(uint64(a.ResourceID), 10),
		// can be authorized based on business
		Attribute: map[string]interface{}{
			client.IamPathKey: []string{fmt.Sprintf("/%s,%d/", sys.Business, a.BizID)},
		},
	}

	switch a.Basic.Action {
	case meta.Create:
		// create current released instance is related to bscp application resource, using config item packing action
		return sys.ReleaseGenerate, []client.Resource{appRes}, nil
	case meta.Find:
		// find release is related to cmdb business resource, using view biz action
		return sys.BusinessViewResource, []client.Resource{bizRes}, nil
	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp action: %s", a.Basic.Action))
	}
}

// genReleasedCIRes generate released config item related iam resource.
// nolint: unused
func genReleasedCIRes(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Business,
		ID:     strconv.FormatUint(uint64(a.BizID), 10),
	}

	switch a.Basic.Action {
	case meta.Find:
		// find released config item is related to cmdb business resource, using view biz action
		return sys.BusinessViewResource, []client.Resource{bizRes}, nil
	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp action: %s", a.Basic.Action))
	}
}

// genStrategyRes generate strategy related iam resource.
// nolint: unused
func genStrategyRes(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Business,
		ID:     strconv.FormatUint(uint64(a.BizID), 10),
	}

	appRes := client.Resource{
		System: sys.SystemIDBSCP,
		Type:   sys.Application,
		ID: strconv.FormatUint(uint64(a.BizID), 10) + bizIDAssembleSymbol +
			strconv.FormatUint(uint64(a.ResourceID), 10),
		// can be authorized based on business
		Attribute: map[string]interface{}{
			client.IamPathKey: []string{fmt.Sprintf("/%s,%d/", sys.Business, a.BizID)},
		},
	}

	switch a.Basic.Action {
	case meta.Create:
		// create strategy is related to bscp application resource, using strategy create action
		return sys.StrategyCreate, []client.Resource{appRes}, nil
	case meta.Update:
		// update strategy is related to bscp application resource, using strategy edit action
		return sys.StrategyEdit, []client.Resource{appRes}, nil
	case meta.Delete:
		// delete strategy is related to bscp application resource, using strategy delete action
		return sys.StrategyDelete, []client.Resource{appRes}, nil
	case meta.Publish:
		// publish strategy is related to bscp application resource, using strategy publish action
		return sys.ReleasePublish, []client.Resource{appRes}, nil
	case meta.FinishPublish:
		// finish publish strategy is related to bscp application resource, using strategy finish publish action
		return sys.ConfigItemFinishPublish, []client.Resource{appRes}, nil
	case meta.Find:
		// find strategy is related to cmdb business resource, using view biz action
		return sys.BusinessViewResource, []client.Resource{bizRes}, nil
	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp action: %s", a.Basic.Action))
	}
}

// genStrategySetRes generate strategy set related iam resource.
// nolint: unused
func genStrategySetRes(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Business,
		ID:     strconv.FormatUint(uint64(a.BizID), 10),
	}

	appRes := client.Resource{
		System: sys.SystemIDBSCP,
		Type:   sys.Application,
		ID: strconv.FormatUint(uint64(a.BizID), 10) + bizIDAssembleSymbol +
			strconv.FormatUint(uint64(a.ResourceID), 10),
		// can be authorized based on business
		Attribute: map[string]interface{}{
			client.IamPathKey: []string{fmt.Sprintf("/%s,%d/", sys.Business, a.BizID)},
		},
	}

	switch a.Basic.Action {
	case meta.Create:
		// create strategy set is related to bscp application resource, using strategy set create action
		return sys.StrategySetCreate, []client.Resource{appRes}, nil
	case meta.Update:
		// update strategy set is related to bscp application resource, using strategy set edit action
		return sys.StrategySetEdit, []client.Resource{appRes}, nil
	case meta.Delete:
		// delete strategy set is related to bscp application resource, using strategy set delete action
		return sys.StrategySetDelete, []client.Resource{appRes}, nil
	case meta.Find:
		// find strategy set is related to cmdb business resource, using view biz action
		return sys.BusinessViewResource, []client.Resource{bizRes}, nil
	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp action: %s", a.Basic.Action))
	}
}

// genPSHRes generate published strategy history related iam resource.
// nolint: unused
func genPSHRes(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Business,
		ID:     strconv.FormatUint(uint64(a.BizID), 10),
	}

	switch a.Basic.Action {
	case meta.Find:
		// find published strategy history is related to cmdb business resource, using task history view action
		return sys.TaskHistoryView, []client.Resource{bizRes}, nil
	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp action: %s", a.Basic.Action))
	}
}

// genRepoRes generate repo related iam resource.
// nolint: unused
func genRepoRes(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Business,
		ID:     strconv.FormatUint(uint64(a.BizID), 10),
	}

	switch a.Basic.Action {
	case meta.Find:
		// find repo resource is related to cmdb business resource, using view biz action
		return sys.BusinessViewResource, []client.Resource{bizRes}, nil
	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp action: %s", a.Basic.Action))
	}
}

// genSidecarRes generate sidecar related iam resource.
// nolint: unused
func genSidecarRes(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Business,
		ID:     strconv.FormatUint(uint64(a.BizID), 10),
	}

	switch a.Basic.Action {
	case meta.Access:
		// request from sidecar is related to business view resource
		return sys.BusinessViewResource, []client.Resource{bizRes}, nil
	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp action: %s", a.Basic.Action))
	}
}

// genCredentialRes generate application credential related iam resource.
// nolint: unused
func genCredentialRes(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Business,
		ID:     strconv.FormatUint(uint64(a.BizID), 10),
	}

	switch a.Basic.Action {
	case meta.Access:
		// request from credential is related to business view resource
		return sys.CredentialView, []client.Resource{bizRes}, nil
	case meta.Manage:
		// manage credential is related to bscp application resource
		return sys.CredentialManage, []client.Resource{bizRes}, nil

	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp action: %s", a.Basic.Action))
	}
}
