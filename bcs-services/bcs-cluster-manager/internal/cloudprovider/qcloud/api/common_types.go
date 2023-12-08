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

package api

import (
	"encoding/json"

	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
)

// ClusterExternalConfig xxx
type ClusterExternalConfig struct {
	// 集群网络插件类型，支持：Flannel、CiliumBGP、CiliumVXLan
	NetworkType *string `json:"NetworkType,omitempty" name:"NetworkType"`

	// 子网ID
	SubnetId *string `json:"SubnetId,omitempty" name:"SubnetId"`

	// Pod CIDR
	// 注意：此字段可能返回 null，表示取不到有效值。
	ClusterCIDR *string `json:"ClusterCIDR,omitempty" name:"ClusterCIDR"`

	// 是否开启第三方节点池支持
	Enabled *bool `json:"Enabled,omitempty" name:"Enabled"`
}

// EnableExternalNodeSupportRequestParams Predefined struct for user
type EnableExternalNodeSupportRequestParams struct {
	// 集群Id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 开启第三方节点池支持配置信息
	ClusterExternalConfig *ClusterExternalConfig `json:"ClusterExternalConfig,omitempty" name:"ClusterExternalConfig"`
}

// EnableExternalNodeSupportRequest xxx
type EnableExternalNodeSupportRequest struct {
	*tchttp.BaseRequest

	// 集群Id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 开启第三方节点池支持配置信息
	ClusterExternalConfig *ClusterExternalConfig `json:"ClusterExternalConfig,omitempty" name:"ClusterExternalConfig"`
}

// ToJsonString xxx
func (r *EnableExternalNodeSupportRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString xxx
func (r *EnableExternalNodeSupportRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "ClusterExternalConfig")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError",
			"EnableExternalNodeSupportRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

// EnableExternalNodeSupportResponseParams Predefined struct for user
type EnableExternalNodeSupportResponseParams struct {
	// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
}

// EnableExternalNodeSupportResponse xxx
type EnableExternalNodeSupportResponse struct {
	*tchttp.BaseResponse
	Response *EnableExternalNodeSupportResponseParams `json:"Response"`
}

// ToJsonString xxx
func (r *EnableExternalNodeSupportResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString xxx
func (r *EnableExternalNodeSupportResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

// DescribeExternalNodeScriptRequestParams Predefined struct for user
type DescribeExternalNodeScriptRequestParams struct {
	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池ID
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 网卡名
	Interface *string `json:"Interface,omitempty" name:"Interface"`

	// 节点名称
	Name *string `json:"Name,omitempty" name:"Name"`
}

// DescribeExternalNodeScriptRequest xxx
type DescribeExternalNodeScriptRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池ID
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 网卡名
	Interface *string `json:"Interface,omitempty" name:"Interface"`

	// 节点名称
	Name *string `json:"Name,omitempty" name:"Name"`
}

// ToJsonString xxx
func (r *DescribeExternalNodeScriptRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString xxx
func (r *DescribeExternalNodeScriptRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "NodePoolId")
	delete(f, "Interface")
	delete(f, "Name")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError",
			"DescribeExternalNodeScriptRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

// DescribeExternalNodeScriptResponseParams Predefined struct for user
type DescribeExternalNodeScriptResponseParams struct {
	// 添加脚本cos下载链接
	Link *string `json:"Link,omitempty" name:"Link"`

	// cos临时密钥
	Token *string `json:"Token,omitempty" name:"Token"`

	// 添加脚本下载命令
	Command *string `json:"Command,omitempty" name:"Command"`

	// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
}

// DescribeExternalNodeScriptResponse xxx
type DescribeExternalNodeScriptResponse struct {
	*tchttp.BaseResponse
	Response *DescribeExternalNodeScriptResponseParams `json:"Response"`
}

// ToJsonString xxx
func (r *DescribeExternalNodeScriptResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString xxx
func (r *DescribeExternalNodeScriptResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

// DeleteExternalNodeRequestParams Predefined struct for user
type DeleteExternalNodeRequestParams struct {
	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 第三方节点列表
	Names []*string `json:"Names,omitempty" name:"Names"`

	// 是否强制删除：如果第三方节点上有运行中Pod，则非强制删除状态下不会进行删除
	Force *bool `json:"Force,omitempty" name:"Force"`
}

// DeleteExternalNodeRequest xxx
type DeleteExternalNodeRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 第三方节点列表
	Names []*string `json:"Names,omitempty" name:"Names"`

	// 是否强制删除：如果第三方节点上有运行中Pod，则非强制删除状态下不会进行删除
	Force *bool `json:"Force,omitempty" name:"Force"`
}

// ToJsonString xxx
func (r *DeleteExternalNodeRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString xxx
func (r *DeleteExternalNodeRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "Names")
	delete(f, "Force")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DeleteExternalNodeRequest has unknown keys!",
			"")
	}
	return json.Unmarshal([]byte(s), &r)
}

// DeleteExternalNodeResponseParams Predefined struct for user
type DeleteExternalNodeResponseParams struct {
	// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
}

// DeleteExternalNodeResponse xxx
type DeleteExternalNodeResponse struct {
	*tchttp.BaseResponse
	Response *DeleteExternalNodeResponseParams `json:"Response"`
}

// ToJsonString xxx
func (r *DeleteExternalNodeResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteExternalNodeResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

// DrainExternalNodeRequestParams Predefined struct for user
type DrainExternalNodeRequestParams struct {
	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点名
	Name *string `json:"Name,omitempty" name:"Name"`

	// 是否只是拉取列表
	DryRun *bool `json:"DryRun,omitempty" name:"DryRun"`
}

// DrainExternalNodeRequest xxx
type DrainExternalNodeRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点名
	Name *string `json:"Name,omitempty" name:"Name"`

	// 是否只是拉取列表
	DryRun *bool `json:"DryRun,omitempty" name:"DryRun"`
}

// ToJsonString xxx
func (r *DrainExternalNodeRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DrainExternalNodeRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "Name")
	delete(f, "DryRun")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DrainExternalNodeRequest has unknown keys!",
			"")
	}
	return json.Unmarshal([]byte(s), &r)
}

// DrainExternalNodeResponseParams Predefined struct for user
type DrainExternalNodeResponseParams struct {
	// pod信息集合
	// 注意：此字段可能返回 null，表示取不到有效值。
	Pods []*SimplePodInfo `json:"Pods,omitempty" name:"Pods"`

	// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
}

// DrainExternalNodeResponse xxx
type DrainExternalNodeResponse struct {
	*tchttp.BaseResponse
	Response *DrainExternalNodeResponseParams `json:"Response"`
}

// ToJsonString xxx
func (r *DrainExternalNodeResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DrainExternalNodeResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

// SimplePodInfo xxx
type SimplePodInfo struct {
	// pod 名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	Name *string `json:"Name,omitempty" name:"Name"`

	// pod 命名空间
	// 注意：此字段可能返回 null，表示取不到有效值。
	Namespace *string `json:"Namespace,omitempty" name:"Namespace"`

	// pod所在节点的ip
	// 注意：此字段可能返回 null，表示取不到有效值。
	NodeIp *string `json:"NodeIp,omitempty" name:"NodeIp"`
}

// DeleteExternalNodePoolRequestParams Predefined struct for user
type DeleteExternalNodePoolRequestParams struct {
	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 第三方节点池ID列表
	NodePoolIds []*string `json:"NodePoolIds,omitempty" name:"NodePoolIds"`

	// 是否强制删除，在第三方节点上有pod的情况下，如果选择非强制删除，则删除会失败
	Force *bool `json:"Force,omitempty" name:"Force"`
}

// DeleteExternalNodePoolRequest xxx
type DeleteExternalNodePoolRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 第三方节点池ID列表
	NodePoolIds []*string `json:"NodePoolIds,omitempty" name:"NodePoolIds"`

	// 是否强制删除，在第三方节点上有pod的情况下，如果选择非强制删除，则删除会失败
	Force *bool `json:"Force,omitempty" name:"Force"`
}

// ToJsonString xxx
func (r *DeleteExternalNodePoolRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteExternalNodePoolRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "NodePoolIds")
	delete(f, "Force")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError",
			"DeleteExternalNodePoolRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

// DeleteExternalNodePoolResponseParams Predefined struct for user
type DeleteExternalNodePoolResponseParams struct {
	// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
}

// DeleteExternalNodePoolResponse xxx
type DeleteExternalNodePoolResponse struct {
	*tchttp.BaseResponse
	Response *DeleteExternalNodePoolResponseParams `json:"Response"`
}

// ToJsonString xxx
func (r *DeleteExternalNodePoolResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteExternalNodePoolResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

// DescribeExternalNodeSupportConfigRequestParams Predefined struct for user
type DescribeExternalNodeSupportConfigRequestParams struct {
	// 集群Id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

// DescribeExternalNodeSupportConfigRequest xxx
type DescribeExternalNodeSupportConfigRequest struct {
	*tchttp.BaseRequest

	// 集群Id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

// ToJsonString xxx
func (r *DescribeExternalNodeSupportConfigRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeExternalNodeSupportConfigRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError",
			"DescribeExternalNodeSupportConfigRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

// DescribeExternalNodeSupportConfigResponseParams Predefined struct for user
type DescribeExternalNodeSupportConfigResponseParams struct {
	// 用于分配集群容器和服务 IP 的 CIDR，不得与 VPC CIDR 冲突，也不得与同 VPC 内其他集群 CIDR 冲突。且网段范围必须在内网网段内，
	// 例如:10.1.0.0/14, 192.168.0.1/18,172.16.0.0/16。
	// 注意：此字段可能返回 null，表示取不到有效值。
	ClusterCIDR *string `json:"ClusterCIDR,omitempty" name:"ClusterCIDR"`

	// 集群网络插件类型，支持：CiliumBGP、CiliumVXLan
	// 注意：此字段可能返回 null，表示取不到有效值。
	NetworkType *string `json:"NetworkType,omitempty" name:"NetworkType"`

	// 子网ID
	// 注意：此字段可能返回 null，表示取不到有效值。
	SubnetId *string `json:"SubnetId,omitempty" name:"SubnetId"`

	// 是否开启第三方节点支持
	// 注意：此字段可能返回 null，表示取不到有效值。
	Enabled *bool `json:"Enabled,omitempty" name:"Enabled"`

	// 节点所属交换机的BGP AS 号
	// 注意：此字段可能返回 null，表示取不到有效值。
	AS *string `json:"AS,omitempty" name:"AS"`

	// 节点所属交换机的交换机 IP
	// 注意：此字段可能返回 null，表示取不到有效值。
	SwitchIP *string `json:"SwitchIP,omitempty" name:"SwitchIP"`

	// 开启第三方节电池状态
	Status *string `json:"Status,omitempty" name:"Status"`

	// 如果开启失败原因
	// 注意：此字段可能返回 null，表示取不到有效值。
	FailedReason *string `json:"FailedReason,omitempty" name:"FailedReason"`

	// 内网访问地址
	// 注意：此字段可能返回 null，表示取不到有效值。
	Master *string `json:"Master,omitempty" name:"Master"`

	// 镜像仓库代理地址
	// 注意：此字段可能返回 null，表示取不到有效值。
	Proxy *string `json:"Proxy,omitempty" name:"Proxy"`

	// 用于记录开启第三方节点的过程进行到哪一步了
	// 注意：此字段可能返回 null，表示取不到有效值。
	Progress []*Step `json:"Progress,omitempty" name:"Progress"`

	// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
}

// DescribeExternalNodeSupportConfigResponse xxx
type DescribeExternalNodeSupportConfigResponse struct {
	*tchttp.BaseResponse
	Response *DescribeExternalNodeSupportConfigResponseParams `json:"Response"`
}

// ToJsonString xxx
func (r *DescribeExternalNodeSupportConfigResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeExternalNodeSupportConfigResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

// DescribeFlowIdStatusRequestParams Predefined struct for user
type DescribeFlowIdStatusRequestParams struct {
	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 开启集群外网访问端口的任务ID
	RequestFlowId *int64 `json:"RequestFlowId,omitempty" name:"RequestFlowId"`
}

// Step xxx
type Step struct {
	// 名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 开始时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	StartAt *string `json:"StartAt,omitempty" name:"StartAt"`

	// 结束时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	EndAt *string `json:"EndAt,omitempty" name:"EndAt"`

	// 当前状态
	// 注意：此字段可能返回 null，表示取不到有效值。
	Status *string `json:"Status,omitempty" name:"Status"`

	// 执行信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	Message *string `json:"Message,omitempty" name:"Message"`
}

// CreateExternalNodePoolRequestParams Predefined struct for user
type CreateExternalNodePoolRequestParams struct {
	// 集群Id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 运行时
	ContainerRuntime *string `json:"ContainerRuntime,omitempty" name:"ContainerRuntime"`

	// 运行时版本
	RuntimeVersion *string `json:"RuntimeVersion,omitempty" name:"RuntimeVersion"`

	// 第三方节点label
	Labels []*Label `json:"Labels,omitempty" name:"Labels"`

	// 第三方节点taint
	Taints []*Taint `json:"Taints,omitempty" name:"Taints"`

	// 第三方节点高级设置
	InstanceAdvancedSettings *InstanceAdvancedSettings `json:"InstanceAdvancedSettings,omitempty" name:"InstanceAdvancedSettings"` // nolint

	// 第三方节点池机器的CPU架构
	Arch *string `json:"Arch,omitempty" name:"Arch"`
}

// CreateExternalNodePoolRequest xxx
type CreateExternalNodePoolRequest struct {
	*tchttp.BaseRequest

	// 集群Id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 运行时
	ContainerRuntime *string `json:"ContainerRuntime,omitempty" name:"ContainerRuntime"`

	// 运行时版本
	RuntimeVersion *string `json:"RuntimeVersion,omitempty" name:"RuntimeVersion"`

	// 第三方节点label
	Labels []*Label `json:"Labels,omitempty" name:"Labels"`

	// 第三方节点taint
	Taints []*Taint `json:"Taints,omitempty" name:"Taints"`

	// 第三方节点高级设置
	InstanceAdvancedSettings *tke.InstanceAdvancedSettings `json:"InstanceAdvancedSettings,omitempty" name:"InstanceAdvancedSettings"` // nolint

	// 第三方节点池机器的CPU架构
	Arch *string `json:"Arch,omitempty" name:"Arch"`
}

// ToJsonString xxx
func (r *CreateExternalNodePoolRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateExternalNodePoolRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "Name")
	delete(f, "ContainerRuntime")
	delete(f, "RuntimeVersion")
	delete(f, "Labels")
	delete(f, "Taints")
	delete(f, "InstanceAdvancedSettings")
	delete(f, "Arch")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError",
			"CreateExternalNodePoolRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

// CreateExternalNodePoolResponseParams Predefined struct for user
type CreateExternalNodePoolResponseParams struct {
	// 节点池ID
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
}

// CreateExternalNodePoolResponse xxx
type CreateExternalNodePoolResponse struct {
	*tchttp.BaseResponse
	Response *CreateExternalNodePoolResponseParams `json:"Response"`
}

// ToJsonString xxx
func (r *CreateExternalNodePoolResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateExternalNodePoolResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

// ModifyExternalNodePoolRequestParams Predefined struct for user
type ModifyExternalNodePoolRequestParams struct {
	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池ID
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 节点池名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 第三方节点label
	Labels []*Label `json:"Labels,omitempty" name:"Labels"`

	// 第三方节点taint
	Taints []*Taint `json:"Taints,omitempty" name:"Taints"`
}

// ModifyExternalNodePoolRequest xxx
type ModifyExternalNodePoolRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池ID
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 节点池名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 第三方节点label
	Labels []*Label `json:"Labels,omitempty" name:"Labels"`

	// 第三方节点taint
	Taints []*Taint `json:"Taints,omitempty" name:"Taints"`
}

// ToJsonString xxx
func (r *ModifyExternalNodePoolRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyExternalNodePoolRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "NodePoolId")
	delete(f, "Name")
	delete(f, "Labels")
	delete(f, "Taints")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError",
			"ModifyExternalNodePoolRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

// ModifyExternalNodePoolResponseParams Predefined struct for user
type ModifyExternalNodePoolResponseParams struct {
	// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
}

// ModifyExternalNodePoolResponse xxx
type ModifyExternalNodePoolResponse struct {
	*tchttp.BaseResponse
	Response *ModifyExternalNodePoolResponseParams `json:"Response"`
}

// ToJsonString xxx
func (r *ModifyExternalNodePoolResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyExternalNodePoolResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

// DescribeExternalNodeRequestParams Predefined struct for user
type DescribeExternalNodeRequestParams struct {
	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池ID
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 节点名称
	Names []*string `json:"Names,omitempty" name:"Names"`
}

// DescribeExternalNodeRequest xxx
type DescribeExternalNodeRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池ID
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 节点名称
	Names []*string `json:"Names,omitempty" name:"Names"`
}

// ToJsonString xxx
func (r *DescribeExternalNodeRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeExternalNodeRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	delete(f, "NodePoolId")
	delete(f, "Names")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeExternalNodeRequest has unknown keys!",
			"")
	}
	return json.Unmarshal([]byte(s), &r)
}

// DescribeExternalNodeResponseParams Predefined struct for user
type DescribeExternalNodeResponseParams struct {
	// 节点列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	Nodes []*ExternalNode `json:"Nodes,omitempty" name:"Nodes"`

	// 节点总数
	// 注意：此字段可能返回 null，表示取不到有效值。
	TotalCount *uint64 `json:"TotalCount,omitempty" name:"TotalCount"`

	// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
}

// DescribeExternalNodeResponse xxx
type DescribeExternalNodeResponse struct {
	*tchttp.BaseResponse
	Response *DescribeExternalNodeResponseParams `json:"Response"`
}

// ToJsonString xxx
func (r *DescribeExternalNodeResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeExternalNodeResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

// ExternalNode xxx
type ExternalNode struct {
	// 第三方节点名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 第三方节点所属节点池
	// 注意：此字段可能返回 null，表示取不到有效值。
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 第三方IP地址
	IP *string `json:"IP,omitempty" name:"IP"`

	// 第三方地域
	Location *string `json:"Location,omitempty" name:"Location"`

	// 第三方节点状态
	Status *string `json:"Status,omitempty" name:"Status"`

	// 创建时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	CreatedTime *string `json:"CreatedTime,omitempty" name:"CreatedTime"`

	// 异常原因
	// 注意：此字段可能返回 null，表示取不到有效值。
	Reason *string `json:"Reason,omitempty" name:"Reason"`

	// 是否封锁。true表示已封锁，false表示未封锁
	// 注意：此字段可能返回 null，表示取不到有效值。
	Unschedulable *bool `json:"Unschedulable,omitempty" name:"Unschedulable"`
}

// DescribeExternalNodePoolsRequestParams Predefined struct for user
type DescribeExternalNodePoolsRequestParams struct {
	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

// DescribeExternalNodePoolsRequest xxx
type DescribeExternalNodePoolsRequest struct {
	*tchttp.BaseRequest

	// 集群ID
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`
}

// ToJsonString xxx
func (r *DescribeExternalNodePoolsRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeExternalNodePoolsRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "ClusterId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError",
			"DescribeExternalNodePoolsRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

// DescribeExternalNodePoolsResponseParams Predefined struct for user
type DescribeExternalNodePoolsResponseParams struct {
	// 节点池总数
	// 注意：此字段可能返回 null，表示取不到有效值。
	TotalCount *uint64 `json:"TotalCount,omitempty" name:"TotalCount"`

	// 第三方节点池列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	NodePoolSet []*ExternalNodePool `json:"NodePoolSet,omitempty" name:"NodePoolSet"`

	// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
}

// DescribeExternalNodePoolsResponse xxx
type DescribeExternalNodePoolsResponse struct {
	*tchttp.BaseResponse
	Response *DescribeExternalNodePoolsResponseParams `json:"Response"`
}

// ToJsonString xxx
func (r *DescribeExternalNodePoolsResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeExternalNodePoolsResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

// ExternalNodePool xxx
type ExternalNodePool struct {
	// 第三方节点池ID
	NodePoolId *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 第三方节点池名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 节点池生命周期
	LifeState *string `json:"LifeState,omitempty" name:"LifeState"`

	// 集群CIDR
	ClusterCIDR *string `json:"ClusterCIDR,omitempty" name:"ClusterCIDR"`

	// 集群网络插件类型
	NetworkType *string `json:"NetworkType,omitempty" name:"NetworkType"`

	// 第三方节点Runtime配置
	RuntimeConfig *RuntimeConfig `json:"RuntimeConfig,omitempty" name:"RuntimeConfig"`

	// 第三方节点label
	// 注意：此字段可能返回 null，表示取不到有效值。
	Labels []*tke.Label `json:"Labels,omitempty" name:"Labels"`

	// 第三方节点taint
	// 注意：此字段可能返回 null，表示取不到有效值。
	Taints []*tke.Taint `json:"Taints,omitempty" name:"Taints"`

	// 第三方节点高级设置
	// 注意：此字段可能返回 null，表示取不到有效值。
	InstanceAdvancedSettings *tke.InstanceAdvancedSettings `json:"InstanceAdvancedSettings,omitempty" name:"InstanceAdvancedSettings"` // nolint
}

// RuntimeConfig xxx
type RuntimeConfig struct {
	// 运行时类型
	// 注意：此字段可能返回 null，表示取不到有效值。
	RuntimeType *string `json:"RuntimeType,omitempty" name:"RuntimeType"`

	// 运行时版本
	// 注意：此字段可能返回 null，表示取不到有效值。
	RuntimeVersion *string `json:"RuntimeVersion,omitempty" name:"RuntimeVersion"`
}

// NewDescribeOSImagesRequest xxx
func NewDescribeOSImagesRequest() (request *DescribeOSImagesRequest) {
	request = &DescribeOSImagesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeOSImages")

	return
}

// NewDescribeOSImagesResponse xxx
func NewDescribeOSImagesResponse() (response *DescribeOSImagesResponse) {
	response = &DescribeOSImagesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeOSImagesRequestParams Predefined struct for user
type DescribeOSImagesRequestParams struct {
}

// DescribeOSImagesRequest xxx
type DescribeOSImagesRequest struct {
	*tchttp.BaseRequest
}

// ToJsonString xxx
func (r *DescribeOSImagesRequest) ToJsonString() string {
	b, _ := json.Marshal(r) // nolint
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeOSImagesRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}

	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeOSImagesRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

// DescribeOSImagesResponseParams Predefined struct for user
type DescribeOSImagesResponseParams struct {
	// 镜像信息列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	OSImageSeriesSet []*OSImage `json:"OSImageSeriesSet,omitempty" name:"OSImageSeriesSet"`

	// 镜像数量
	// 注意：此字段可能返回 null，表示取不到有效值。
	TotalCount *uint64 `json:"TotalCount,omitempty" name:"TotalCount"`

	// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
}

// DescribeOSImagesResponse xxx
type DescribeOSImagesResponse struct {
	*tchttp.BaseResponse
	Response *DescribeOSImagesResponseParams `json:"Response"`
}

// ToJsonString xxx
func (r *DescribeOSImagesResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeOSImagesResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

// OSImage xxx
type OSImage struct {
	// os聚合名称
	SeriesName *string `json:"SeriesName,omitempty" name:"SeriesName"`

	// os别名
	Alias *string `json:"Alias,omitempty" name:"Alias"`

	// os架构
	Arch *string `json:"Arch,omitempty" name:"Arch"`

	// os名称
	OsName *string `json:"OsName,omitempty" name:"OsName"`

	// 操作系统类型(分为定制和非定制，取值分别为:DOCKER_CUSTOMIZE、GENERAL)
	OsCustomizeType *string `json:"OsCustomizeType,omitempty" name:"OsCustomizeType"`

	// os是否下线(online表示在线,offline表示下线)
	Status *string `json:"Status,omitempty" name:"Status"`

	// 镜像id
	ImageId *string `json:"ImageId,omitempty" name:"ImageId"`
}

// DescribeInstanceCreateProgressRequestParams Predefined struct for user
type DescribeInstanceCreateProgressRequestParams struct {
	// 集群Id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点实例Id
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`
}

// DescribeInstanceCreateProgressRequest request
type DescribeInstanceCreateProgressRequest struct {
	*tchttp.BaseRequest

	// 集群Id
	ClusterId *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点实例Id
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`
}

// ToJsonString json string
func (r *DescribeInstanceCreateProgressRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// DescribeInstanceCreateProgressResponseParams Predefined struct for user
type DescribeInstanceCreateProgressResponseParams struct {
	// 创建进度
	// 注意：此字段可能返回 null，表示取不到有效值。
	Progress []*Step `json:"Progress,omitempty" name:"Progress"`

	// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
}

// DescribeInstanceCreateProgressResponse response
type DescribeInstanceCreateProgressResponse struct {
	*tchttp.BaseResponse
	Response *DescribeInstanceCreateProgressResponseParams `json:"Response"`
}

// ToJsonString to string
func (r *DescribeInstanceCreateProgressResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}
