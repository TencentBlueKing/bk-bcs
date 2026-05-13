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

// Package bcs templateconfig操作
package bcs

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
)

// EnvCidrStep env cidr step
type EnvCidrStep struct {
	Env  string `json:"env"`
	Step uint32 `json:"step"`
}

// CloudNetworkTemplateConfig cloud network template config
type CloudNetworkTemplateConfig struct {
	CidrSteps         []*EnvCidrStep `json:"cidrSteps"`
	ServiceSteps      []uint32       `json:"serviceSteps"`
	PerNodePodNum     []uint32       `json:"perNodePodNum"`
	UnderlaySteps     []uint32       `json:"underlaySteps"`
	UnderlayAutoSteps []uint32       `json:"underlayAutoSteps"`
}

// CloudTemplateConfig cloud template config
type CloudTemplateConfig struct {
	CloudNetworkTemplateConfig *CloudNetworkTemplateConfig `json:"cloudNetworkTemplateConfig"`
}

// CreateTemplateConfigReq create template config req
type CreateTemplateConfigReq struct {
	BusinessID          string               `json:"businessID"`
	ProjectID           string               `json:"projectID"`
	ClusterID           string               `json:"clusterID"`
	Provider            string               `json:"provider"`
	ConfigType          string               `json:"configType"`
	CloudTemplateConfig *CloudTemplateConfig `json:"cloudTemplateConfig"`
}

// TemplateConfigInfo template config info
type TemplateConfigInfo struct {
	TemplateConfigID    string               `json:"templateConfigID"`
	BusinessID          string               `json:"businessID"`
	ProjectID           string               `json:"projectID"`
	ClusterID           string               `json:"clusterID"`
	Provider            string               `json:"provider"`
	ConfigType          string               `json:"configType"`
	CloudTemplateConfig *CloudTemplateConfig `json:"cloudTemplateConfig"`
	Creator             string               `json:"creator"`
	Updater             string               `json:"updater"`
	CreateTime          string               `json:"createTime"`
	UpdateTime          string               `json:"updateTime"`
}

// CreateTemplateConfig 创建template config
func CreateTemplateConfig(templateconfig *CreateTemplateConfigReq) (bool, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/templateconfigs", config.G.BCS.Host)

	resp, err := component.GetClient().R().
		SetAuthToken(config.G.BCS.Token).
		SetBody(templateconfig).
		Post(url)

	if err != nil {
		blog.Errorf("create template config error, %s", err.Error())
		return false, err
	}

	var result bool
	fmt.Printf("create template config response: %s", resp.String())
	if err = component.UnmarshalBKResult(resp, &result); err != nil {
		blog.Errorf("unmarshal template config error, %s", err.Error())
		return false, err
	}

	return result, nil
}

// ListTemplateConfig 列出template config
func ListTemplateConfig(businessID, projectID, clusterID, provider, configType string) ([]*TemplateConfigInfo, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/templateconfigs", config.G.BCS.Host)

	queryParams := make(map[string]string)
	if businessID != "" {
		queryParams["businessID"] = businessID
	}
	if projectID != "" {
		queryParams["projectID"] = projectID
	}
	if clusterID != "" {
		queryParams["clusterID"] = clusterID
	}
	if provider != "" {
		queryParams["provider"] = provider
	}
	if configType != "" {
		queryParams["configType"] = configType
	}
	resp, err := component.GetClient().R().
		SetAuthToken(config.G.BCS.Token).
		SetQueryParams(queryParams).
		Get(url)

	if err != nil {
		blog.Errorf("list template config error, %s", err.Error())
		return nil, err
	}

	var result []*TemplateConfigInfo
	fmt.Printf("list template config response: %s", resp.String())
	if err = component.UnmarshalBKData(resp, &result); err != nil {
		blog.Errorf("unmarshal template config error, %s", err.Error())
		return nil, err
	}

	return result, nil
}

// DeleteTemplateConfig 删除template config
func DeleteTemplateConfig(templateConfigID, businessID, projectID string) (bool, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/templateconfigs/%s", config.G.BCS.Host, templateConfigID)

	queryParams := make(map[string]string)
	if businessID != "" {
		queryParams["businessID"] = businessID
	}
	if projectID != "" {
		queryParams["projectID"] = projectID
	}
	resp, err := component.GetClient().R().
		SetAuthToken(config.G.BCS.Token).
		SetQueryParams(queryParams).
		Delete(url)

	if err != nil {
		blog.Errorf("delete template config error, %s", err.Error())
		return false, err
	}

	var result bool
	fmt.Printf("delete template config response: %s", resp.String())
	if err = component.UnmarshalBKResult(resp, &result); err != nil {
		blog.Errorf("unmarshal template config error, %s", err.Error())
		return false, err
	}

	return result, nil
}
