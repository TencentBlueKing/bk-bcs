/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
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

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/client"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/iam/sys"
	"github.com/TencentBlueKing/iam-go-sdk"
	bkiam "github.com/TencentBlueKing/iam-go-sdk"
)

// bizIDAssembleSymbol used to assemble biz_id and resource id's symbol, used in app id generation.
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
	iamReq := bkiam.NewRequest(
		sys.SystemIDBSCP,
		dummyIAMUser,
		bkiam.NewAction(string(sys.BusinessViewResource)),
		[]iam.ResourceNode{
			{
				System:    sys.SystemIDCMDB,
				Type:      string(sys.Business),
				ID:        strconv.FormatUint(uint64(a.BizID), 10),
				Attribute: map[string]interface{}{},
			},
		},
	)

	switch a.Basic.Action {
	case meta.FindBusinessResource:
		// create app is related to cmdb business resource
		return &iamReq, nil
	default:
		return nil, fmt.Errorf("unsupported bscp action: %s", a.Basic.Action)
	}
}

func genBizIAMApplication(a *meta.ResourceAttribute) (*bkiam.Application, error) {
	actions := []bkiam.ApplicationAction{
		{
			ID: string(sys.BusinessViewResource),
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
		},
	}
	application := bkiam.NewApplication(sys.SystemIDBSCP, actions)

	switch a.Basic.Action {
	case meta.FindBusinessResource:
		// create app is related to cmdb business resource
		return &application, nil
	default:
		return nil, fmt.Errorf("unsupported bscp action: %s", a.Basic.Action)
	}
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
		ID: strconv.FormatUint(uint64(a.BizID), 10) + bizIDAssembleSymbol +
			strconv.FormatUint(uint64(a.ResourceID), 10),
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
	case meta.Find:
		// find app is related to cmdb business resource, using view biz action
		return sys.BusinessViewResource, []client.Resource{bizRes}, nil
	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp action: %s", a.Basic.Action))
	}
}

// genCommitResource generate commit related iam resource.
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
		return sys.ConfigItemPacking, []client.Resource{appRes}, nil
	case meta.Find:
		// find release is related to cmdb business resource, using view biz action
		return sys.BusinessViewResource, []client.Resource{bizRes}, nil
	default:
		return "", nil, errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported bscp action: %s", a.Basic.Action))
	}
}

// genReleasedCIRes generate released config item related iam resource.
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
		return sys.ConfigItemPublish, []client.Resource{appRes}, nil
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
