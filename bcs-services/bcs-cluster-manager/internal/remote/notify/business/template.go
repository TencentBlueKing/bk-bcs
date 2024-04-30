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

// Package business xxx
package business

import (
	"bytes"
	gotemplate "text/template"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

var (
	business  = "business"
	cluster   = "cluster"
	nodeGroup = "group"
	status    = "status"
)

// BuildClusterDimension build cluster dimensions
func BuildClusterDimension(clsId, bizId, state string) map[string]string {
	return map[string]string{
		business: bizId,
		cluster:  clsId,
		status:   state,
	}
}

// BuildNodeGroupDimension build nodeGroup dimensions
func BuildNodeGroupDimension(clsId, bizId, groupId, state string) map[string]string {
	return map[string]string{
		business:  bizId,
		cluster:   clsId,
		nodeGroup: groupId,
		status:    state,
	}
}

// 创建集群默认通知模板
const createClusterTemplate = `
用户 {{.Operator}}, 
在项目 {{ .ProjectID }} 下创建集群 {{ .ClusterID }} / {{ .ClusterName }} {{ .Result }}.
详细信息请参考, visit: {{.URL}}
`

// 删除集群默认通知模板
const deleteClusterTemplate = `
用户 {{.Operator}},
在项目 {{ .ProjectID }} 下删除集群 {{ .ClusterID }} / {{ .ClusterName }} {{ .Result }}.
详细信息请参考, visit: {{.URL}}
`

// 创建节点池默认通知模板
const createNodeGroupTemplate = `
用户 {{.Operator}},
在集群 {{ .ClusterID }} / {{ .ClusterName }} 创建节点池 {{ .NodeGroupID }} / {{ .NodeGroupName }}.
创建时间: {{ .OperatorTime }}.
操作结果: {{ .Result }}.
`

// 节点池扩容默认通知模板
const nodegroupScaleOutNodesTemplate = `
用户 {{.Operator}},
在集群 {{ .ClusterID }} / {{ .ClusterName }} 的节点池 {{ .NodeGroupID }} / {{ .NodeGroupName }} 下扩容 {{ .NodeNum }} 个节点.
扩容时间: {{ .OperatorTime }}.
扩容结果: {{ .Result }}.
操作节点列表: {{ .NodeIPList }}.
`

// 节点池缩容默认通知模板
const nodegroupScaleInNodesTemplate = `
用户 {{.Operator}},
在集群 {{ .ClusterID }} / {{ .ClusterName }} 的节点池 {{ .NodeGroupID }} / {{ .NodeGroupName }} 下缩容 {{ .NodeNum }} 个节点.
缩容时间: {{ .OperatorTime }}.
缩容结果: {{ .Result }}.
操作节点列表: {{ .NodeIPList }}.
`

var (
	// CreateCluster 创建集群默认通知模板
	CreateCluster = proto.NotifyData{
		Title:   "create_cluster",
		Content: createClusterTemplate,
	}
	// DeleteCluster 删除集群默认通知模板
	DeleteCluster = proto.NotifyData{
		Title:   "delete_cluster",
		Content: deleteClusterTemplate,
	}

	// CreateNodeGroup 节点池创建默认通知模板
	CreateNodeGroup = proto.NotifyData{
		Title:   "create_nodegroup",
		Content: createNodeGroupTemplate,
	}

	// NodeGroupScaleOutNodes 节点池扩容节点默认通知模板
	NodeGroupScaleOutNodes = proto.NotifyData{
		Title:   "group_scaleout_nodes",
		Content: nodegroupScaleOutNodesTemplate,
	}
	// NodeGroupScaleInNodes 节点池缩容默认通知模版
	NodeGroupScaleInNodes = proto.NotifyData{
		Title:   "group_scalein_nodes",
		Content: nodegroupScaleInNodesTemplate,
	}
)

// GetNotifyTemplateContent get notify template render result
func GetNotifyTemplateContent(cls *proto.Cluster, group *proto.NodeGroup,
	extra ExtraParas, script string) (string, error) {
	if !extra.Render {
		return script, nil
	}

	// render script
	render := NewNotifierTemplateVars(cls, group, extra)
	renderStr, err := render.RenderTemplateVars("", script)
	if err != nil {
		return script, err
	}

	return renderStr, nil
}

// ExtraParas paras
type ExtraParas struct {
	Render       bool
	NodeNum      int
	NodeIPList   string
	OperatorTime string
	Operator     string
	Result       string
}

// NotifierTemplateVars render template vars
type NotifierTemplateVars struct {
	ProjectID     string
	ClusterID     string
	ClusterName   string
	BusinessID    string
	Region        string
	NodeGroupID   string
	NodeGroupName string

	NodeNum      int
	NodeIPList   string
	OperatorTime string
	Operator     string
	Result       string
}

// NewNotifierTemplateVars init render struct
func NewNotifierTemplateVars(cluster *proto.Cluster, group *proto.NodeGroup, extra ExtraParas) NotifierTemplateVars {
	return NotifierTemplateVars{
		ProjectID:     cluster.ProjectID,
		ClusterID:     cluster.ClusterID,
		ClusterName:   cluster.ClusterName,
		BusinessID:    cluster.BusinessID,
		Region:        cluster.Region,
		NodeGroupID:   group.NodeGroupID,
		NodeGroupName: group.Name,
		NodeNum:       extra.NodeNum,
		NodeIPList:    extra.NodeIPList,
		OperatorTime:  extra.OperatorTime,
		Operator:      extra.Operator,
		Result:        extra.Result,
	}
}

// RenderTemplateVars render text by NotifierTemplateVars
func (rtv NotifierTemplateVars) RenderTemplateVars(name, text string) (string, error) {
	defaultRenderName := "notifier"
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
