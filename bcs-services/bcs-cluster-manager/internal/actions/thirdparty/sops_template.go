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

package thirdparty

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"
	"github.com/avast/retry-go"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// ListTemplateListAction action for list sops templates
type ListTemplateListAction struct {
	ctx context.Context

	req          *cmproto.GetBkSopsTemplateListRequest
	resp         *cmproto.GetBkSopsTemplateListResponse
	templateList []*cmproto.TemplateInfo
}

// NewListTemplateListAction create list action for business templateList
func NewListTemplateListAction() *ListTemplateListAction {
	return &ListTemplateListAction{}
}

func (la *ListTemplateListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Data = la.templateList
}

func (la *ListTemplateListAction) getBusinessTemplateList() error {
	var (
		err          error
		templateList []*common.TemplateData
	)
	err = retry.Do(func() error {
		path := &common.TemplateListPathPara{
			BkBizID:  la.req.BusinessID,
			Operator: la.req.Operator,
		}
		req := &common.TemplateRequest{
			TemplateSource: la.req.TemplateSource,
			Scope:          common.Scope(la.req.Scope),
		}
		templateList, err = common.GetBKOpsClient().GetBusinessTemplateList(la.ctx, path, req)
		if err != nil {
			return err
		}

		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("getBusinessTemplateList failed: %v", err)
		return err
	}

	if la.templateList == nil {
		la.templateList = make([]*cmproto.TemplateInfo, 0)
	}
	for i := range templateList {
		la.templateList = append(la.templateList, &cmproto.TemplateInfo{
			TemplateName: templateList[i].Name,
			TemplateID:   fmt.Sprintf("%d", templateList[i].ID),
			BusinessID:   uint32(templateList[i].BkBizID),
			BusinessName: templateList[i].BkBizName,
			Creator:      templateList[i].Creator,
			Editor:       templateList[i].Editor,
		})
	}

	return nil
}

// Handle handle list sops template list
func (la *ListTemplateListAction) Handle(
	ctx context.Context, req *cmproto.GetBkSopsTemplateListRequest, resp *cmproto.GetBkSopsTemplateListResponse) {
	if req == nil || resp == nil {
		blog.Errorf("ListTemplateListAction failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(icommon.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.getBusinessTemplateList(); err != nil {
		la.setResp(icommon.BcsErrClusterManagerBkSopsInterfaceErr, err.Error())
		return
	}
	la.setResp(icommon.BcsErrClusterManagerSuccess, icommon.BcsErrClusterManagerSuccessStr)
}

// GetTemplateInfoAction action for list sops templates
type GetTemplateInfoAction struct {
	ctx context.Context

	req          *cmproto.GetBkSopsTemplateInfoRequest
	resp         *cmproto.GetBkSopsTemplateInfoResponse
	templateInfo *cmproto.TemplateDetailInfo
}

// NewGetTemplateInfoAction create list action for business templateList
func NewGetTemplateInfoAction() *GetTemplateInfoAction {
	return &GetTemplateInfoAction{}
}

func (la *GetTemplateInfoAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Data = la.templateInfo
}

func (la *GetTemplateInfoAction) getBusinessTemplateInfoValues() error {
	var (
		err            error
		constantValues []common.ConstantValue
		project        *common.ProjectInfo
	)
	err = retry.Do(func() error {
		path := &common.TemplateDetailPathPara{
			BkBizID:    la.req.BusinessID,
			TemplateID: la.req.TemplateID,
			Operator:   la.req.Operator,
		}
		req := &common.TemplateRequest{
			TemplateSource: la.req.TemplateSource,
			Scope:          common.Scope(la.req.Scope),
		}
		constantValues, err = common.GetBKOpsClient().GetBusinessTemplateInfo(la.ctx, path, req)
		if err != nil {
			return err
		}

		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("getBusinessTemplateInfoValues failed: %v", err)
		return err
	}

	// get bksops project info by bizID
	err = retry.Do(func() error {
		project, err = common.GetBKOpsClient().GetUserProjectDetailInfo(la.ctx, la.req.BusinessID)
		if err != nil {
			return err
		}
		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("getBusinessTemplateInfoValues failed: %v", err)
	}

	if la.templateInfo == nil {
		la.templateInfo = &cmproto.TemplateDetailInfo{
			TemplateUrl: getSopsTemplateUrl(project, la.req.TemplateID),
			Values:      make([]*cmproto.ConstantValue, 0),
		}
	}
	for i := range constantValues {
		newKey := extractValue(constantValues[i].Key)
		if newKey == "" {
			blog.Errorf("key[%s] not conform to rules[${key}]", constantValues[i].Key)
			continue
		}

		if la.templateInfo.Values == nil {
			la.templateInfo.Values = make([]*cmproto.ConstantValue, 0)
		}

		la.templateInfo.Values = append(la.templateInfo.Values, &cmproto.ConstantValue{
			Key:   newKey,
			Name:  constantValues[i].Name,
			Index: uint32(constantValues[i].Index),
			Desc:  constantValues[i].Desc,
		})
	}

	return nil
}

// Handle handle list sops template info
func (la *GetTemplateInfoAction) Handle(
	ctx context.Context, req *cmproto.GetBkSopsTemplateInfoRequest, resp *cmproto.GetBkSopsTemplateInfoResponse) {
	if req == nil || resp == nil {
		blog.Errorf("GetTemplateInfoAction failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(icommon.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.getBusinessTemplateInfoValues(); err != nil {
		la.setResp(icommon.BcsErrClusterManagerBkSopsInterfaceErr, err.Error())
		return
	}
	la.setResp(icommon.BcsErrClusterManagerSuccess, icommon.BcsErrClusterManagerSuccessStr)
}

// extract template para value
func extractValue(value string) string {
	if !strings.HasPrefix(value, "${") || !strings.HasSuffix(value, "}") {
		return ""
	}

	holderRex := regexp.MustCompile(`^\$\{(.*?)\}`)
	subMatch := holderRex.FindAllStringSubmatch(value, -1)
	if len(subMatch) > 0 && len(subMatch[0]) >= 1 {
		return subMatch[0][1]
	}

	return ""
}

// GetTemplateValuesAction action for list sops template values
type GetTemplateValuesAction struct {
	ctx context.Context

	model   store.ClusterManagerModel
	cluster *cmproto.Cluster

	req            *cmproto.GetInnerTemplateValuesRequest
	resp           *cmproto.GetInnerTemplateValuesResponse
	templateValues []*cmproto.TemplateValue
}

// NewGetTemplateValuesAction create list action for template values
func NewGetTemplateValuesAction(model store.ClusterManagerModel) *GetTemplateValuesAction {
	return &GetTemplateValuesAction{model: model}
}

func (la *GetTemplateValuesAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Data = la.templateValues
}

func (la *GetTemplateValuesAction) getInnerTemplateValues() error {
	if la.req.ClusterID != "" {
		cls, err := actions.GetClusterInfoByClusterID(la.model, la.req.ClusterID)
		if err != nil {
			blog.Errorf("GetTemplateValuesAction getInnerTemplateValues failed: %v", err)
			return err
		}

		la.cluster = cls
	}

	if la.templateValues == nil {
		la.templateValues = make([]*cmproto.TemplateValue, 0)
	}
	for i := range template.InnerTemplateVarsList {
		la.templateValues = append(la.templateValues, &cmproto.TemplateValue{
			Name:  template.InnerTemplateVarsList[i].VarName,
			Desc:  i18n.T(la.ctx, strings.ReplaceAll(template.InnerTemplateVarsList[i].ReferMethod, " ", "")),
			Refer: template.InnerTemplateVarsList[i].ReferMethod,
			Trans: template.InnerTemplateVarsList[i].TransMethod,
			Value: func() string {
				if la.cluster == nil {
					return ""
				}
				val, _ := template.GetInnerTemplateVarsByName(template.InnerTemplateVarsList[i].TransMethod,
					la.cluster, template.ExtraInfo{
						NodeOperator: la.req.Operator,
					})

				return val
			}(),
		})
	}

	return nil
}

// Handle handle list template values
func (la *GetTemplateValuesAction) Handle(
	ctx context.Context, req *cmproto.GetInnerTemplateValuesRequest, resp *cmproto.GetInnerTemplateValuesResponse) {
	if req == nil || resp == nil {
		blog.Errorf("GetTemplateValuesAction failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(icommon.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.getInnerTemplateValues(); err != nil {
		la.setResp(icommon.BcsErrClusterManagerBkSopsInterfaceErr, err.Error())
		return
	}

	la.setResp(icommon.BcsErrClusterManagerSuccess, icommon.BcsErrClusterManagerSuccessStr)
}

func getSopsTemplateUrl(project *common.ProjectInfo, moduleID string) string {
	switch {
	case options.GetEditionInfo().IsInnerEdition(), options.GetEditionInfo().IsCommunicationEdition():
		if options.GetGlobalCMOptions().BKOps.TemplateURL == "" || project == nil || project.ProjectId <= 0 {
			break
		}
		return fmt.Sprintf(options.GetGlobalCMOptions().BKOps.TemplateURL, project.ProjectId, moduleID)
	}

	return options.GetGlobalCMOptions().BKOps.FrontURL
}
