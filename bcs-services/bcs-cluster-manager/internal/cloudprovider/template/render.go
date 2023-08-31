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

package template

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	gotemplate "text/template"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// RenderVars render vars
type RenderVars struct {
	Cluster  *proto.Cluster
	IPList   string
	Operator string
	Render   bool
}

// GetNodeTemplateScript get script render result
func GetNodeTemplateScript(vars RenderVars, script string) (string, error) {
	if !vars.Render {
		return script, nil
	}

	decodeScript, err := base64.StdEncoding.DecodeString(script)
	if err != nil {
		return script, err
	}

	// render script
	render := NewRenderTemplateVars(vars.Cluster, vars.IPList, vars.Operator)
	renderStr, err := render.RenderTxtVars("", string(decodeScript))
	if err != nil {
		return script, err
	}

	return base64.StdEncoding.EncodeToString([]byte(renderStr)), nil
}

// RenderTemplateVars render template vars
type RenderTemplateVars struct {
	NodeIPList   string
	ProjectID    string
	ClusterID    string
	MasterIPList string
	Region       string
	ClusterVPC   string
	BusinessID   string
	NodeOperator string
}

// NewRenderTemplateVars init render struct
func NewRenderTemplateVars(cluster *proto.Cluster, ips, operator string) RenderTemplateVars {
	if len(ips) == 0 {
		ips = nodeIPList
	}
	return RenderTemplateVars{
		NodeIPList:   ips,
		ProjectID:    cluster.ProjectID,
		ClusterID:    cluster.ClusterID,
		MasterIPList: getClusterMasterIPs(cluster),
		Region:       cluster.Region,
		ClusterVPC:   cluster.VpcID,
		BusinessID:   cluster.BusinessID,
		NodeOperator: operator,
	}
}

// RenderTxtVars render text by RenderTemplateVars
func (rtv RenderTemplateVars) RenderTxtVars(name, text string) (string, error) {
	defaultRenderName := "renderVars"
	if name == "" {
		name = defaultRenderName
	}
	tmpl, err := gotemplate.New(name).Parse(text)
	if err != nil {
		blog.Errorf("template Parse[%s] failed: %v", text, err)
		return "", err
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, rtv); err != nil {
		blog.Errorf("template Execute[%s] failed: %v", text, err)
		return "", err
	}

	return buf.String(), nil
}

// VarTemplate xxx
type VarTemplate struct {
	VarName     string
	Desc        string
	ReferMethod string
	TransMethod string
}

// TransToReferMethod xxx
var TransToReferMethod = map[string]string{
	nodeIPList:       "{{ .NodeIPList }}",
	clusterProjectID: "{{ .ProjectID }}",
	clusterID:        "{{ .ClusterID }}",
	clusterMasterIPs: "{{ .MasterIPList }}",
	clusterRegion:    "{{ .Region }}",
	clusterVPC:       "{{ .ClusterVPC }}",
	clusterBizID:     "{{ .BusinessID }}",
	nodeOperator:     "{{ .NodeOperator }}",
}

// InnerTemplateVarsList inner template values list
var InnerTemplateVarsList = []VarTemplate{
	{
		VarName:     "node_list",
		Desc:        "集群上架节点列表(,隔开的字符串)",
		ReferMethod: "{{ .NodeIPList }}",
		TransMethod: nodeIPList,
	},
	{
		VarName:     "project_id",
		Desc:        "集群所属项目ID",
		ReferMethod: "{{ .ProjectID }}",
		TransMethod: clusterProjectID,
	},
	{
		VarName:     "cluster_id",
		Desc:        "集群ID",
		ReferMethod: "{{ .ClusterID }}",
		TransMethod: clusterID,
	},
	{
		VarName:     "master_ips",
		Desc:        "集群master节点IP列表(,隔开的字符串)",
		ReferMethod: "{{ .MasterIPList }}",
		TransMethod: clusterMasterIPs,
	},
	{
		VarName:     "region",
		Desc:        "集群地域信息",
		ReferMethod: "{{ .Region }}",
		TransMethod: clusterRegion,
	},
	{
		VarName:     "vpc",
		Desc:        "集群vpc信息",
		ReferMethod: "{{ .ClusterVPC }}",
		TransMethod: clusterVPC,
	},
	{
		VarName:     "business_id",
		Desc:        "集群所属业务ID",
		ReferMethod: "{{ .BusinessID }}",
		TransMethod: clusterBizID,
	},
	{
		VarName:     "node_operator",
		Desc:        "集群上下架节点操作人",
		ReferMethod: "{{ .NodeOperator }}",
		TransMethod: nodeOperator,
	},
}

// InnerTemplateVars inner template values referMethod To InnerTemplateVars
var InnerTemplateVars = map[string]VarTemplate{
	"{{ .NodeIPList }}": {
		VarName:     "node_list",
		Desc:        "集群上架节点列表(,隔开的字符串)",
		ReferMethod: "{{ .NodeIPList }}",
		TransMethod: nodeIPList,
	},
	"{{ .ProjectID }}": {
		VarName:     "project_id",
		Desc:        "集群所属项目ID",
		ReferMethod: "{{ .ProjectID }}",
		TransMethod: clusterProjectID,
	},
	"{{ .ClusterID }}": {
		VarName:     "cluster_id",
		Desc:        "集群ID",
		ReferMethod: "{{ .ClusterID }}",
		TransMethod: clusterID,
	},
	"{{ .MasterIPList }}": {
		VarName:     "master_ips",
		Desc:        "集群master节点IP列表(,隔开的字符串)",
		ReferMethod: "{{ .MasterIPList }}",
		TransMethod: clusterMasterIPs,
	},
	"{{ .Region }}": {
		VarName:     "region",
		Desc:        "集群地域信息",
		ReferMethod: "{{ .Region }}",
		TransMethod: clusterRegion,
	},
	"{{ .ClusterVPC }}": {
		VarName:     "vpc",
		Desc:        "集群vpc信息",
		ReferMethod: "{{ .ClusterVPC }}",
		TransMethod: clusterVPC,
	},
	"{{ .BusinessID }}": {
		VarName:     "business_id",
		Desc:        "集群所属业务ID",
		ReferMethod: "{{ .BusinessID }}",
		TransMethod: clusterBizID,
	},
	"{{ .NodeOperator }}": {
		VarName:     "node_operator",
		Desc:        "集群上下架节点操作人",
		ReferMethod: "{{ .NodeOperator }}",
		TransMethod: nodeOperator,
	},
}

// GetInnerTemplateVarsByName get template inner vars value
func GetInnerTemplateVarsByName(name string, cluster *proto.Cluster, extra ExtraInfo) (string, error) {
	if cluster == nil {
		errMsg := fmt.Errorf("cluster is empty when GetInnerTemplateVarsByName")
		return "", errMsg
	}

	switch name {
	case clusterProjectID:
		return cluster.GetProjectID(), nil
	case nodeIPList:
		return extra.NodeIPList, nil
	case clusterID:
		return cluster.GetClusterID(), nil
	case clusterMasterIPs:
		return getClusterMasterIPs(cluster), nil
	case clusterRegion:
		return cluster.GetRegion(), nil
	case clusterVPC:
		return cluster.GetVpcID(), nil
	case clusterBizID:
		if extra.BusinessID == "" {
			return cluster.GetBusinessID(), nil
		}
		return extra.BusinessID, nil
	case nodeOperator:
		return extra.NodeOperator, nil
	default:
	}

	return "", fmt.Errorf("GetInnerTemplateVarsByName unSupportType %s", name)
}

func renderTextContainTemplateVars(text string) bool {
	for renderVal := range InnerTemplateVars {
		if strings.Contains(text, renderVal) {
			return true
		}
	}

	return false
}
