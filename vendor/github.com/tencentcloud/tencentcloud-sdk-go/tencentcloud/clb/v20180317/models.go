// Copyright (c) 2017-2018 THL A29 Limited, a Tencent company. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v20180317

import (
    "encoding/json"

    tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
)

type AutoRewriteRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// HTTPS:443监听器的ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// HTTPS:443监听器下需要重定向的域名
	Domains []*string `json:"Domains,omitempty" name:"Domains" list`
}

func (r *AutoRewriteRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *AutoRewriteRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type AutoRewriteResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *AutoRewriteResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *AutoRewriteResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type Backend struct {

	// 后端服务的类型，可取：CVM、ENI（即将支持）
	Type *string `json:"Type,omitempty" name:"Type"`

	// 后端服务的唯一 ID，如 ins-abcd1234
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 后端服务的监听端口
	Port *int64 `json:"Port,omitempty" name:"Port"`

	// 后端服务的转发权重，取值范围：[0, 100]，默认为 10。
	Weight *int64 `json:"Weight,omitempty" name:"Weight"`

	// 后端服务的外网 IP
	// 注意：此字段可能返回 null，表示取不到有效值。
	PublicIpAddresses []*string `json:"PublicIpAddresses,omitempty" name:"PublicIpAddresses" list`

	// 后端服务的内网 IP
	// 注意：此字段可能返回 null，表示取不到有效值。
	PrivateIpAddresses []*string `json:"PrivateIpAddresses,omitempty" name:"PrivateIpAddresses" list`

	// 后端服务的实例名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	InstanceName *string `json:"InstanceName,omitempty" name:"InstanceName"`

	// 后端服务被绑定的时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	RegisteredTime *string `json:"RegisteredTime,omitempty" name:"RegisteredTime"`

	// 弹性网卡唯一ID，如 eni-1234abcd
	// 注意：此字段可能返回 null，表示取不到有效值。
	EniId *string `json:"EniId,omitempty" name:"EniId"`
}

type BatchDeregisterTargetsRequest struct {
	*tchttp.BaseRequest

	// 负载均衡ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 解绑目标
	Targets []*BatchTarget `json:"Targets,omitempty" name:"Targets" list`
}

func (r *BatchDeregisterTargetsRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *BatchDeregisterTargetsRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type BatchDeregisterTargetsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 解绑失败的监听器ID
		FailListenerIdSet []*string `json:"FailListenerIdSet,omitempty" name:"FailListenerIdSet" list`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *BatchDeregisterTargetsResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *BatchDeregisterTargetsResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type BatchModifyTargetWeightRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 要批量修改权重的列表
	ModifyList []*RsWeightRule `json:"ModifyList,omitempty" name:"ModifyList" list`
}

func (r *BatchModifyTargetWeightRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *BatchModifyTargetWeightRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type BatchModifyTargetWeightResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *BatchModifyTargetWeightResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *BatchModifyTargetWeightResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type BatchRegisterTargetsRequest struct {
	*tchttp.BaseRequest

	// 负载均衡ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 绑定目标
	Targets []*BatchTarget `json:"Targets,omitempty" name:"Targets" list`
}

func (r *BatchRegisterTargetsRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *BatchRegisterTargetsRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type BatchRegisterTargetsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 绑定失败的监听器ID，如为空表示全部绑定成功。
	// 注意：此字段可能返回 null，表示取不到有效值。
		FailListenerIdSet []*string `json:"FailListenerIdSet,omitempty" name:"FailListenerIdSet" list`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *BatchRegisterTargetsResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *BatchRegisterTargetsResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type BatchTarget struct {

	// 监听器ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 绑定端口
	Port *int64 `json:"Port,omitempty" name:"Port"`

	// 子机ID
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 弹性网卡ip
	EniIp *string `json:"EniIp,omitempty" name:"EniIp"`

	// 子机权重，范围[0, 100]。绑定时如果不存在，则默认为10。
	Weight *int64 `json:"Weight,omitempty" name:"Weight"`

	// 七层规则ID
	LocationId *string `json:"LocationId,omitempty" name:"LocationId"`
}

type CertificateInput struct {

	// 认证类型，UNIDIRECTIONAL：单向认证，MUTUAL：双向认证
	SSLMode *string `json:"SSLMode,omitempty" name:"SSLMode"`

	// 服务端证书的 ID，如果不填写此项则必须上传证书，包括 CertContent，CertKey，CertName。
	CertId *string `json:"CertId,omitempty" name:"CertId"`

	// 客户端证书的 ID，当监听器采用双向认证，即 SSLMode=mutual 时，如果不填写此项则必须上传客户端证书，包括 CertCaContent，CertCaName。
	CertCaId *string `json:"CertCaId,omitempty" name:"CertCaId"`

	// 上传服务端证书的名称，如果没有 CertId，则此项必传。
	CertName *string `json:"CertName,omitempty" name:"CertName"`

	// 上传服务端证书的 key，如果没有 CertId，则此项必传。
	CertKey *string `json:"CertKey,omitempty" name:"CertKey"`

	// 上传服务端证书的内容，如果没有 CertId，则此项必传。
	CertContent *string `json:"CertContent,omitempty" name:"CertContent"`

	// 上传客户端 CA 证书的名称，如果 SSLMode=mutual，如果没有 CertCaId，则此项必传。
	CertCaName *string `json:"CertCaName,omitempty" name:"CertCaName"`

	// 上传客户端证书的内容，如果 SSLMode=mutual，如果没有 CertCaId，则此项必传。
	CertCaContent *string `json:"CertCaContent,omitempty" name:"CertCaContent"`
}

type CertificateOutput struct {

	// 认证类型，UNIDIRECTIONAL：单向认证，MUTUAL：双向认证
	SSLMode *string `json:"SSLMode,omitempty" name:"SSLMode"`

	// 服务端证书的 ID。
	CertId *string `json:"CertId,omitempty" name:"CertId"`

	// 客户端证书的 ID。
	// 注意：此字段可能返回 null，表示取不到有效值。
	CertCaId *string `json:"CertCaId,omitempty" name:"CertCaId"`
}

type ClassicalHealth struct {

	// 后端服务的内网 IP
	IP *string `json:"IP,omitempty" name:"IP"`

	// 后端服务的端口
	Port *int64 `json:"Port,omitempty" name:"Port"`

	// 负载均衡的监听端口
	ListenerPort *int64 `json:"ListenerPort,omitempty" name:"ListenerPort"`

	// 转发协议
	Protocol *string `json:"Protocol,omitempty" name:"Protocol"`

	// 健康检查结果，1 表示健康，0 表示不健康
	HealthStatus *int64 `json:"HealthStatus,omitempty" name:"HealthStatus"`
}

type ClassicalListener struct {

	// 负载均衡监听器ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 负载均衡监听器端口
	ListenerPort *int64 `json:"ListenerPort,omitempty" name:"ListenerPort"`

	// 监听器后端转发端口
	InstancePort *int64 `json:"InstancePort,omitempty" name:"InstancePort"`

	// 监听器名称
	ListenerName *string `json:"ListenerName,omitempty" name:"ListenerName"`

	// 监听器协议类型
	Protocol *string `json:"Protocol,omitempty" name:"Protocol"`

	// 会话保持时间
	SessionExpire *int64 `json:"SessionExpire,omitempty" name:"SessionExpire"`

	// 是否开启了健康检查：1（开启）、0（关闭）
	HealthSwitch *int64 `json:"HealthSwitch,omitempty" name:"HealthSwitch"`

	// 响应超时时间
	TimeOut *int64 `json:"TimeOut,omitempty" name:"TimeOut"`

	// 检查间隔
	IntervalTime *int64 `json:"IntervalTime,omitempty" name:"IntervalTime"`

	// 健康阈值
	HealthNum *int64 `json:"HealthNum,omitempty" name:"HealthNum"`

	// 不健康阈值
	UnhealthNum *int64 `json:"UnhealthNum,omitempty" name:"UnhealthNum"`

	// 传统型公网负载均衡的 HTTP、HTTPS 监听器的请求均衡方法。wrr 表示按权重轮询，ip_hash 表示根据访问的源 IP 进行一致性哈希方式来分发
	HttpHash *string `json:"HttpHash,omitempty" name:"HttpHash"`

	// 传统型公网负载均衡的 HTTP、HTTPS 监听器的健康检查返回码。具体可参考创建监听器中对该字段的解释
	HttpCode *int64 `json:"HttpCode,omitempty" name:"HttpCode"`

	// 传统型公网负载均衡的 HTTP、HTTPS 监听器的健康检查路径
	HttpCheckPath *string `json:"HttpCheckPath,omitempty" name:"HttpCheckPath"`

	// 传统型公网负载均衡的 HTTPS 监听器的认证方式
	SSLMode *string `json:"SSLMode,omitempty" name:"SSLMode"`

	// 传统型公网负载均衡的 HTTPS 监听器的服务端证书 ID
	CertId *string `json:"CertId,omitempty" name:"CertId"`

	// 传统型公网负载均衡的 HTTPS 监听器的客户端证书 ID
	CertCaId *string `json:"CertCaId,omitempty" name:"CertCaId"`

	// 监听器的状态，0 表示创建中，1 表示运行中
	Status *int64 `json:"Status,omitempty" name:"Status"`
}

type ClassicalLoadBalancerInfo struct {

	// 后端实例ID
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 负载均衡实例ID列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	LoadBalancerIds []*string `json:"LoadBalancerIds,omitempty" name:"LoadBalancerIds" list`
}

type ClassicalTarget struct {

	// 后端服务的类型，可取值：CVM、ENI（即将支持）
	Type *string `json:"Type,omitempty" name:"Type"`

	// 后端服务的唯一 ID，可通过 DescribeInstances 接口返回字段中的 unInstanceId 字段获取
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 后端服务的转发权重，取值范围：[0, 100]，默认为 10。
	Weight *int64 `json:"Weight,omitempty" name:"Weight"`

	// 后端服务的外网 IP
	// 注意：此字段可能返回 null，表示取不到有效值。
	PublicIpAddresses []*string `json:"PublicIpAddresses,omitempty" name:"PublicIpAddresses" list`

	// 后端服务的内网 IP
	// 注意：此字段可能返回 null，表示取不到有效值。
	PrivateIpAddresses []*string `json:"PrivateIpAddresses,omitempty" name:"PrivateIpAddresses" list`

	// 后端服务的实例名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	InstanceName *string `json:"InstanceName,omitempty" name:"InstanceName"`

	// 后端服务的状态
	// 1：故障，2：运行中，3：创建中，4：已关机，5：已退还，6：退还中， 7：重启中，8：开机中，9：关机中，10：密码重置中，11：格式化中，12：镜像制作中，13：带宽设置中，14：重装系统中，19：升级中，21：热迁移中
	// 注意：此字段可能返回 null，表示取不到有效值。
	RunFlag *int64 `json:"RunFlag,omitempty" name:"RunFlag"`
}

type ClassicalTargetInfo struct {

	// 后端实例ID
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 权重，取值范围 [0, 100]
	Weight *int64 `json:"Weight,omitempty" name:"Weight"`
}

type CreateListenerRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 要将监听器创建到哪些端口，每个端口对应一个新的监听器
	Ports []*int64 `json:"Ports,omitempty" name:"Ports" list`

	// 监听器协议： TCP | UDP | HTTP | HTTPS | TCP_SSL（TCP_SSL 正在内测中，如需使用请通过工单申请）
	Protocol *string `json:"Protocol,omitempty" name:"Protocol"`

	// 要创建的监听器名称列表，名称与Ports数组按序一一对应，如不需立即命名，则无需提供此参数
	ListenerNames []*string `json:"ListenerNames,omitempty" name:"ListenerNames" list`

	// 健康检查相关参数，此参数仅适用于TCP/UDP/TCP_SSL监听器
	HealthCheck *HealthCheck `json:"HealthCheck,omitempty" name:"HealthCheck"`

	// 证书相关信息，此参数仅适用于HTTPS/TCP_SSL监听器
	Certificate *CertificateInput `json:"Certificate,omitempty" name:"Certificate"`

	// 会话保持时间，单位：秒。可选值：30~3600，默认 0，表示不开启。此参数仅适用于TCP/UDP监听器。
	SessionExpireTime *int64 `json:"SessionExpireTime,omitempty" name:"SessionExpireTime"`

	// 监听器转发的方式。可选值：WRR、LEAST_CONN
	// 分别表示按权重轮询、最小连接数， 默认为 WRR。此参数仅适用于TCP/UDP/TCP_SSL监听器。
	Scheduler *string `json:"Scheduler,omitempty" name:"Scheduler"`

	// 是否开启SNI特性，此参数仅适用于HTTPS监听器。
	SniSwitch *int64 `json:"SniSwitch,omitempty" name:"SniSwitch"`
}

func (r *CreateListenerRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *CreateListenerRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type CreateListenerResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 创建的监听器的唯一标识数组
		ListenerIds []*string `json:"ListenerIds,omitempty" name:"ListenerIds" list`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreateListenerResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *CreateListenerResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type CreateLoadBalancerRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例的网络类型：
	// OPEN：公网属性， INTERNAL：内网属性。
	LoadBalancerType *string `json:"LoadBalancerType,omitempty" name:"LoadBalancerType"`

	// 负载均衡实例的类型。1：通用的负载均衡实例，目前只支持传入1
	Forward *int64 `json:"Forward,omitempty" name:"Forward"`

	// 负载均衡实例的名称，只在创建一个实例的时候才会生效。规则：1-50 个英文、汉字、数字、连接线“-”或下划线“_”。
	// 注意：如果名称与系统中已有负载均衡实例的名称相同，则系统将会自动生成此次创建的负载均衡实例的名称。
	LoadBalancerName *string `json:"LoadBalancerName,omitempty" name:"LoadBalancerName"`

	// 负载均衡后端目标设备所属的网络 ID，如vpc-12345678，可以通过 DescribeVpcEx 接口获取。 不传此参数则默认为基础网络（"0"）。
	VpcId *string `json:"VpcId,omitempty" name:"VpcId"`

	// 在私有网络内购买内网负载均衡实例的情况下，必须指定子网 ID，内网负载均衡实例的 VIP 将从这个子网中产生。其它情况不支持该参数。
	SubnetId *string `json:"SubnetId,omitempty" name:"SubnetId"`

	// 负载均衡实例所属的项目 ID，可以通过 DescribeProject 接口获取。不传此参数则视为默认项目。
	ProjectId *int64 `json:"ProjectId,omitempty" name:"ProjectId"`

	// 仅适用于公网负载均衡。IP版本，IPV4 | IPV6，默认值 IPV4。
	AddressIPVersion *string `json:"AddressIPVersion,omitempty" name:"AddressIPVersion"`

	// 创建负载均衡的个数，默认值 1。
	Number *uint64 `json:"Number,omitempty" name:"Number"`

	// 仅适用于公网负载均衡。设置跨可用区容灾时的主可用区ID，例如 100001 或 ap-guangzhou-1
	// 注：主可用区是需要承载流量的可用区，备可用区默认不承载流量，主可用区不可用时才使用备可用区，平台将为您自动选择最佳备可用区。可通过 DescribeMasterZones 接口查询一个地域的主可用区的列表。
	MasterZoneId *string `json:"MasterZoneId,omitempty" name:"MasterZoneId"`

	// 仅适用于公网负载均衡。可用区ID，指定可用区以创建负载均衡实例。如：ap-guangzhou-1
	ZoneId *string `json:"ZoneId,omitempty" name:"ZoneId"`

	// 仅适用于公网负载均衡。Anycast的发布域，可取 ZONE_A 或 ZONE_B。仅带宽非上移用户支持此参数。
	AnycastZone *string `json:"AnycastZone,omitempty" name:"AnycastZone"`

	// 仅适用于公网负载均衡。负载均衡的网络计费方式，此参数仅对带宽上移用户生效。
	InternetAccessible *InternetAccessible `json:"InternetAccessible,omitempty" name:"InternetAccessible"`

	// 购买负载均衡同时，给负载均衡打上标签
	Tags []*TagInfo `json:"Tags,omitempty" name:"Tags" list`
}

func (r *CreateLoadBalancerRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *CreateLoadBalancerRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type CreateLoadBalancerResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 由负载均衡实例唯一 ID 组成的数组。
		LoadBalancerIds []*string `json:"LoadBalancerIds,omitempty" name:"LoadBalancerIds" list`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreateLoadBalancerResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *CreateLoadBalancerResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type CreateRuleRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 监听器 ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 新建转发规则的信息
	Rules []*RuleInput `json:"Rules,omitempty" name:"Rules" list`
}

func (r *CreateRuleRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *CreateRuleRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type CreateRuleResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *CreateRuleResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *CreateRuleResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DeleteListenerRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 要删除的监听器 ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`
}

func (r *DeleteListenerRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DeleteListenerRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DeleteListenerResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeleteListenerResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DeleteListenerResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DeleteLoadBalancerRequest struct {
	*tchttp.BaseRequest

	// 要删除的负载均衡实例 ID数组，数组大小最大支持20
	LoadBalancerIds []*string `json:"LoadBalancerIds,omitempty" name:"LoadBalancerIds" list`
}

func (r *DeleteLoadBalancerRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DeleteLoadBalancerRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DeleteLoadBalancerResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeleteLoadBalancerResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DeleteLoadBalancerResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DeleteRewriteRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 源监听器ID
	SourceListenerId *string `json:"SourceListenerId,omitempty" name:"SourceListenerId"`

	// 目标监听器ID
	TargetListenerId *string `json:"TargetListenerId,omitempty" name:"TargetListenerId"`

	// 转发规则之间的重定向关系
	RewriteInfos []*RewriteLocationMap `json:"RewriteInfos,omitempty" name:"RewriteInfos" list`
}

func (r *DeleteRewriteRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DeleteRewriteRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DeleteRewriteResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeleteRewriteResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DeleteRewriteResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DeleteRuleRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 负载均衡监听器 ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 要删除的转发规则的ID组成的数组
	LocationIds []*string `json:"LocationIds,omitempty" name:"LocationIds" list`

	// 要删除的转发规则的域名，已提供LocationIds参数时本参数不生效
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// 要删除的转发规则的转发路径，已提供LocationIds参数时本参数不生效
	Url *string `json:"Url,omitempty" name:"Url"`
}

func (r *DeleteRuleRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DeleteRuleRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DeleteRuleResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeleteRuleResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DeleteRuleResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DeregisterTargetsFromClassicalLBRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 后端服务的实例ID列表
	InstanceIds []*string `json:"InstanceIds,omitempty" name:"InstanceIds" list`
}

func (r *DeregisterTargetsFromClassicalLBRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DeregisterTargetsFromClassicalLBRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DeregisterTargetsFromClassicalLBResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeregisterTargetsFromClassicalLBResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DeregisterTargetsFromClassicalLBResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DeregisterTargetsRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID，格式如 lb-12345678
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 监听器 ID，格式如 lbl-12345678
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 要解绑的后端服务列表，数组长度最大支持20
	Targets []*Target `json:"Targets,omitempty" name:"Targets" list`

	// 转发规则的ID，格式如 loc-12345678，当从七层转发规则解绑机器时，必须提供此参数或Domain+Url两者之一
	LocationId *string `json:"LocationId,omitempty" name:"LocationId"`

	// 目标规则的域名，提供LocationId参数时本参数不生效
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// 目标规则的URL，提供LocationId参数时本参数不生效
	Url *string `json:"Url,omitempty" name:"Url"`
}

func (r *DeregisterTargetsRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DeregisterTargetsRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DeregisterTargetsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DeregisterTargetsResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DeregisterTargetsResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeClassicalLBByInstanceIdRequest struct {
	*tchttp.BaseRequest

	// 后端实例ID列表
	InstanceIds []*string `json:"InstanceIds,omitempty" name:"InstanceIds" list`
}

func (r *DescribeClassicalLBByInstanceIdRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeClassicalLBByInstanceIdRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeClassicalLBByInstanceIdResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 负载均衡相关信息列表
		LoadBalancerInfoList []*ClassicalLoadBalancerInfo `json:"LoadBalancerInfoList,omitempty" name:"LoadBalancerInfoList" list`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClassicalLBByInstanceIdResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeClassicalLBByInstanceIdResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeClassicalLBHealthStatusRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 负载均衡监听器ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`
}

func (r *DescribeClassicalLBHealthStatusRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeClassicalLBHealthStatusRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeClassicalLBHealthStatusResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 后端健康状态列表
	// 注意：此字段可能返回 null，表示取不到有效值。
		HealthList []*ClassicalHealth `json:"HealthList,omitempty" name:"HealthList" list`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClassicalLBHealthStatusResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeClassicalLBHealthStatusResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeClassicalLBListenersRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 负载均衡监听器ID列表
	ListenerIds []*string `json:"ListenerIds,omitempty" name:"ListenerIds" list`

	// 负载均衡监听的协议, 'TCP', 'UDP', 'HTTP', 'HTTPS'
	Protocol *string `json:"Protocol,omitempty" name:"Protocol"`

	// 负载均衡监听端口， 范围[1-65535]
	ListenerPort *int64 `json:"ListenerPort,omitempty" name:"ListenerPort"`

	// 监听器的状态，0 表示创建中，1 表示运行中
	Status *int64 `json:"Status,omitempty" name:"Status"`
}

func (r *DescribeClassicalLBListenersRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeClassicalLBListenersRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeClassicalLBListenersResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 监听器列表
	// 注意：此字段可能返回 null，表示取不到有效值。
		Listeners []*ClassicalListener `json:"Listeners,omitempty" name:"Listeners" list`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClassicalLBListenersResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeClassicalLBListenersResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeClassicalLBTargetsRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`
}

func (r *DescribeClassicalLBTargetsRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeClassicalLBTargetsRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeClassicalLBTargetsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 后端服务列表
	// 注意：此字段可能返回 null，表示取不到有效值。
		Targets []*ClassicalTarget `json:"Targets,omitempty" name:"Targets" list`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeClassicalLBTargetsResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeClassicalLBTargetsResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeListenersRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 要查询的负载均衡监听器 ID数组
	ListenerIds []*string `json:"ListenerIds,omitempty" name:"ListenerIds" list`

	// 要查询的监听器协议类型，取值 TCP | UDP | HTTP | HTTPS | TCP_SSL
	Protocol *string `json:"Protocol,omitempty" name:"Protocol"`

	// 要查询的监听器的端口
	Port *int64 `json:"Port,omitempty" name:"Port"`
}

func (r *DescribeListenersRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeListenersRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeListenersResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 监听器列表
		Listeners []*Listener `json:"Listeners,omitempty" name:"Listeners" list`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeListenersResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeListenersResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeLoadBalancersRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID。
	LoadBalancerIds []*string `json:"LoadBalancerIds,omitempty" name:"LoadBalancerIds" list`

	// 负载均衡实例的网络类型：
	// OPEN：公网属性， INTERNAL：内网属性。
	LoadBalancerType *string `json:"LoadBalancerType,omitempty" name:"LoadBalancerType"`

	// 负载均衡实例的类型。1：通用的负载均衡实例，0：传统型负载均衡实例。如果不传此参数，则查询所有类型的负载均衡实例。
	Forward *int64 `json:"Forward,omitempty" name:"Forward"`

	// 负载均衡实例的名称。
	LoadBalancerName *string `json:"LoadBalancerName,omitempty" name:"LoadBalancerName"`

	// 腾讯云为负载均衡实例分配的域名，本参数仅对传统型公网负载均衡才有意义。
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// 负载均衡实例的 VIP 地址，支持多个。
	LoadBalancerVips []*string `json:"LoadBalancerVips,omitempty" name:"LoadBalancerVips" list`

	// 负载均衡绑定的后端服务的外网 IP。
	BackendPublicIps []*string `json:"BackendPublicIps,omitempty" name:"BackendPublicIps" list`

	// 负载均衡绑定的后端服务的内网 IP。
	BackendPrivateIps []*string `json:"BackendPrivateIps,omitempty" name:"BackendPrivateIps" list`

	// 数据偏移量，默认为 0。
	Offset *int64 `json:"Offset,omitempty" name:"Offset"`

	// 返回负载均衡实例的个数，默认为 20。
	Limit *int64 `json:"Limit,omitempty" name:"Limit"`

	// 排序参数，支持以下字段：LoadBalancerName，CreateTime，Domain，LoadBalancerType。
	OrderBy *string `json:"OrderBy,omitempty" name:"OrderBy"`

	// 1：倒序，0：顺序，默认按照创建时间倒序。
	OrderType *int64 `json:"OrderType,omitempty" name:"OrderType"`

	// 搜索字段，模糊匹配名称、域名、VIP。
	SearchKey *string `json:"SearchKey,omitempty" name:"SearchKey"`

	// 负载均衡实例所属的项目 ID，可以通过 DescribeProject 接口获取。
	ProjectId *int64 `json:"ProjectId,omitempty" name:"ProjectId"`

	// 负载均衡是否绑定后端服务，0：没有绑定后端服务，1：绑定后端服务，-1：查询全部。
	WithRs *int64 `json:"WithRs,omitempty" name:"WithRs"`

	// 负载均衡实例所属私有网络唯一ID，如 vpc-bhqkbhdx，
	// 基础网络可传入'0'。
	VpcId *string `json:"VpcId,omitempty" name:"VpcId"`

	// 安全组ID，如 sg-m1cc9123
	SecurityGroup *string `json:"SecurityGroup,omitempty" name:"SecurityGroup"`

	// 主可用区ID，如 ："100001" （对应的是广州一区）
	MasterZone *string `json:"MasterZone,omitempty" name:"MasterZone"`
}

func (r *DescribeLoadBalancersRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeLoadBalancersRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeLoadBalancersResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 满足过滤条件的负载均衡实例总数。此数值与入参中的Limit无关。
		TotalCount *uint64 `json:"TotalCount,omitempty" name:"TotalCount"`

		// 返回的负载均衡实例数组。
		LoadBalancerSet []*LoadBalancer `json:"LoadBalancerSet,omitempty" name:"LoadBalancerSet" list`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeLoadBalancersResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeLoadBalancersResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeRewriteRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 负载均衡监听器ID数组
	SourceListenerIds []*string `json:"SourceListenerIds,omitempty" name:"SourceListenerIds" list`

	// 负载均衡转发规则的ID数组
	SourceLocationIds []*string `json:"SourceLocationIds,omitempty" name:"SourceLocationIds" list`
}

func (r *DescribeRewriteRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeRewriteRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeRewriteResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 重定向转发规则构成的数组，若无重定向规则，则返回空数组
		RewriteSet []*RuleOutput `json:"RewriteSet,omitempty" name:"RewriteSet" list`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeRewriteResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeRewriteResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeTargetHealthRequest struct {
	*tchttp.BaseRequest

	// 要查询的负载均衡实例 ID列表
	LoadBalancerIds []*string `json:"LoadBalancerIds,omitempty" name:"LoadBalancerIds" list`
}

func (r *DescribeTargetHealthRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeTargetHealthRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeTargetHealthResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 负载均衡实例列表
	// 注意：此字段可能返回 null，表示取不到有效值。
		LoadBalancers []*LoadBalancerHealth `json:"LoadBalancers,omitempty" name:"LoadBalancers" list`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeTargetHealthResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeTargetHealthResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeTargetsRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 监听器 ID列表
	ListenerIds []*string `json:"ListenerIds,omitempty" name:"ListenerIds" list`

	// 监听器协议类型
	Protocol *string `json:"Protocol,omitempty" name:"Protocol"`

	// 监听器端口
	Port *int64 `json:"Port,omitempty" name:"Port"`
}

func (r *DescribeTargetsRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeTargetsRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeTargetsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 监听器后端绑定的机器信息
	// 注意：此字段可能返回 null，表示取不到有效值。
		Listeners []*ListenerBackend `json:"Listeners,omitempty" name:"Listeners" list`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeTargetsResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeTargetsResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeTaskStatusRequest struct {
	*tchttp.BaseRequest

	// 请求ID，即接口返回的 RequestId 参数
	TaskId *string `json:"TaskId,omitempty" name:"TaskId"`
}

func (r *DescribeTaskStatusRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeTaskStatusRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type DescribeTaskStatusResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 任务的当前状态。 0：成功，1：失败，2：进行中。
		Status *int64 `json:"Status,omitempty" name:"Status"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *DescribeTaskStatusResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *DescribeTaskStatusResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ExtraInfo struct {

	// 是否开通VIP直通
	// 注意：此字段可能返回 null，表示取不到有效值。
	ZhiTong *bool `json:"ZhiTong,omitempty" name:"ZhiTong"`

	// TgwGroup名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	TgwGroupName *string `json:"TgwGroupName,omitempty" name:"TgwGroupName"`
}

type HealthCheck struct {

	// 是否开启健康检查：1（开启）、0（关闭）。
	HealthSwitch *int64 `json:"HealthSwitch,omitempty" name:"HealthSwitch"`

	// 健康检查的响应超时时间（仅适用于四层监听器），可选值：2~60，默认值：2，单位：秒。响应超时时间要小于检查间隔时间。
	// 注意：此字段可能返回 null，表示取不到有效值。
	TimeOut *int64 `json:"TimeOut,omitempty" name:"TimeOut"`

	// 健康检查探测间隔时间，默认值：5，可选值：5~300，单位：秒。
	// 注意：此字段可能返回 null，表示取不到有效值。
	IntervalTime *int64 `json:"IntervalTime,omitempty" name:"IntervalTime"`

	// 健康阈值，默认值：3，表示当连续探测三次健康则表示该转发正常，可选值：2~10，单位：次。
	// 注意：此字段可能返回 null，表示取不到有效值。
	HealthNum *int64 `json:"HealthNum,omitempty" name:"HealthNum"`

	// 不健康阈值，默认值：3，表示当连续探测三次不健康则表示该转发异常，可选值：2~10，单位：次。
	// 注意：此字段可能返回 null，表示取不到有效值。
	UnHealthNum *int64 `json:"UnHealthNum,omitempty" name:"UnHealthNum"`

	// 健康检查状态码（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式）。可选值：1~31，默认 31。
	// 1 表示探测后返回值 1xx 代表健康，2 表示返回 2xx 代表健康，4 表示返回 3xx 代表健康，8 表示返回 4xx 代表健康，16 表示返回 5xx 代表健康。若希望多种返回码都可代表健康，则将相应的值相加。注意：TCP监听器的HTTP健康检查方式，只支持指定一种健康检查状态码。
	// 注意：此字段可能返回 null，表示取不到有效值。
	HttpCode *int64 `json:"HttpCode,omitempty" name:"HttpCode"`

	// 健康检查路径（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式）。
	// 注意：此字段可能返回 null，表示取不到有效值。
	HttpCheckPath *string `json:"HttpCheckPath,omitempty" name:"HttpCheckPath"`

	// 健康检查域名（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式）。
	// 注意：此字段可能返回 null，表示取不到有效值。
	HttpCheckDomain *string `json:"HttpCheckDomain,omitempty" name:"HttpCheckDomain"`

	// 健康检查方法（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式），默认值：HEAD，可选值HEAD或GET。
	// 注意：此字段可能返回 null，表示取不到有效值。
	HttpCheckMethod *string `json:"HttpCheckMethod,omitempty" name:"HttpCheckMethod"`

	// 自定义探测相关参数。健康检查端口，默认为后端服务的端口，除非您希望指定特定端口，否则建议留空。（仅适用于TCP/UDP监听器）。
	// 注意：此字段可能返回 null，表示取不到有效值。
	CheckPort *int64 `json:"CheckPort,omitempty" name:"CheckPort"`

	// 自定义探测相关参数。健康检查协议CheckType的值取CUSTOM时，必填此字段，代表健康检查的输入格式，可取值：HEX或TEXT；取值为HEX时，SendContext和RecvContext的字符只能在0123456789ABCDEF中选取且长度必须是偶数位。（仅适用于TCP/UDP监听器）
	// 注意：此字段可能返回 null，表示取不到有效值。
	ContextType *string `json:"ContextType,omitempty" name:"ContextType"`

	// 自定义探测相关参数。健康检查协议CheckType的值取CUSTOM时，必填此字段，代表健康检查发送的请求内容，只允许ASCII可见字符，最大长度限制500。（仅适用于TCP/UDP监听器）。
	// 注意：此字段可能返回 null，表示取不到有效值。
	SendContext *string `json:"SendContext,omitempty" name:"SendContext"`

	// 自定义探测相关参数。健康检查协议CheckType的值取CUSTOM时，必填此字段，代表健康检查返回的结果，只允许ASCII可见字符，最大长度限制500。（仅适用于TCP/UDP监听器）。
	// 注意：此字段可能返回 null，表示取不到有效值。
	RecvContext *string `json:"RecvContext,omitempty" name:"RecvContext"`

	// 自定义探测相关参数。健康检查使用的协议：TCP | HTTP | CUSTOM（仅适用于TCP/UDP监听器，其中UDP监听器只支持CUSTOM；如果使用自定义健康检查功能，则必传）。
	// 注意：此字段可能返回 null，表示取不到有效值。
	CheckType *string `json:"CheckType,omitempty" name:"CheckType"`

	// 自定义探测相关参数。健康检查协议CheckType的值取HTTP时，必传此字段，代表后端服务的HTTP版本：HTTP/1.0、HTTP/1.1；（仅适用于TCP监听器）
	// 注意：此字段可能返回 null，表示取不到有效值。
	HttpVersion *string `json:"HttpVersion,omitempty" name:"HttpVersion"`
}

type InternetAccessible struct {

	// TRAFFIC_POSTPAID_BY_HOUR 按流量按小时后计费 ; BANDWIDTH_POSTPAID_BY_HOUR 按带宽按小时后计费;
	// BANDWIDTH_PACKAGE 按带宽包计费（当前，只有指定运营商时才支持此种计费模式）
	InternetChargeType *string `json:"InternetChargeType,omitempty" name:"InternetChargeType"`

	// 最大出带宽，单位Mbps，范围支持0到2048，仅对公网属性的LB生效，默认值 10
	InternetMaxBandwidthOut *int64 `json:"InternetMaxBandwidthOut,omitempty" name:"InternetMaxBandwidthOut"`
}

type LBChargePrepaid struct {

	// 续费类型：AUTO_RENEW 自动续费，  MANUAL_RENEW 手动续费
	// 注意：此字段可能返回 null，表示取不到有效值。
	RenewFlag *string `json:"RenewFlag,omitempty" name:"RenewFlag"`

	// 周期，表示多少个月（保留字段）
	// 注意：此字段可能返回 null，表示取不到有效值。
	Period *int64 `json:"Period,omitempty" name:"Period"`
}

type Listener struct {

	// 负载均衡监听器 ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 监听器协议
	Protocol *string `json:"Protocol,omitempty" name:"Protocol"`

	// 监听器端口
	Port *int64 `json:"Port,omitempty" name:"Port"`

	// 监听器绑定的证书信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	Certificate *CertificateOutput `json:"Certificate,omitempty" name:"Certificate"`

	// 监听器的健康检查信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	HealthCheck *HealthCheck `json:"HealthCheck,omitempty" name:"HealthCheck"`

	// 请求的调度方式
	// 注意：此字段可能返回 null，表示取不到有效值。
	Scheduler *string `json:"Scheduler,omitempty" name:"Scheduler"`

	// 会话保持时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	SessionExpireTime *int64 `json:"SessionExpireTime,omitempty" name:"SessionExpireTime"`

	// 是否开启SNI特性（本参数仅对于HTTPS监听器有意义）
	// 注意：此字段可能返回 null，表示取不到有效值。
	SniSwitch *int64 `json:"SniSwitch,omitempty" name:"SniSwitch"`

	// 监听器下的全部转发规则（本参数仅对于HTTP/HTTPS监听器有意义）
	// 注意：此字段可能返回 null，表示取不到有效值。
	Rules []*RuleOutput `json:"Rules,omitempty" name:"Rules" list`

	// 监听器的名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	ListenerName *string `json:"ListenerName,omitempty" name:"ListenerName"`

	// 监听器的创建时间。
	// 注意：此字段可能返回 null，表示取不到有效值。
	CreateTime *string `json:"CreateTime,omitempty" name:"CreateTime"`

	// 端口段结束端口
	// 注意：此字段可能返回 null，表示取不到有效值。
	EndPort *int64 `json:"EndPort,omitempty" name:"EndPort"`
}

type ListenerBackend struct {

	// 监听器 ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 监听器的协议
	Protocol *string `json:"Protocol,omitempty" name:"Protocol"`

	// 监听器的端口
	Port *int64 `json:"Port,omitempty" name:"Port"`

	// 监听器下的规则信息（仅适用于HTTP/HTTPS监听器）
	// 注意：此字段可能返回 null，表示取不到有效值。
	Rules []*RuleTargets `json:"Rules,omitempty" name:"Rules" list`

	// 监听器上绑定的后端服务列表（仅适用于TCP/UDP/TCP_SSL监听器）
	// 注意：此字段可能返回 null，表示取不到有效值。
	Targets []*Backend `json:"Targets,omitempty" name:"Targets" list`
}

type ListenerHealth struct {

	// 监听器ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 监听器名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	ListenerName *string `json:"ListenerName,omitempty" name:"ListenerName"`

	// 监听器的协议
	Protocol *string `json:"Protocol,omitempty" name:"Protocol"`

	// 监听器的端口
	Port *int64 `json:"Port,omitempty" name:"Port"`

	// 监听器的转发规则列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	Rules []*RuleHealth `json:"Rules,omitempty" name:"Rules" list`
}

type LoadBalancer struct {

	// 负载均衡实例 ID。
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 负载均衡实例的名称。
	LoadBalancerName *string `json:"LoadBalancerName,omitempty" name:"LoadBalancerName"`

	// 负载均衡实例的网络类型：
	// OPEN：公网属性， INTERNAL：内网属性。
	LoadBalancerType *string `json:"LoadBalancerType,omitempty" name:"LoadBalancerType"`

	// 负载均衡类型标识，1：负载均衡，0：传统型负载均衡。
	Forward *uint64 `json:"Forward,omitempty" name:"Forward"`

	// 负载均衡实例的域名，仅公网传统型负载均衡实例才提供该字段
	// 注意：此字段可能返回 null，表示取不到有效值。
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// 负载均衡实例的 VIP 列表。
	// 注意：此字段可能返回 null，表示取不到有效值。
	LoadBalancerVips []*string `json:"LoadBalancerVips,omitempty" name:"LoadBalancerVips" list`

	// 负载均衡实例的状态，包括
	// 0：创建中，1：正常运行。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Status *uint64 `json:"Status,omitempty" name:"Status"`

	// 负载均衡实例的创建时间。
	// 注意：此字段可能返回 null，表示取不到有效值。
	CreateTime *string `json:"CreateTime,omitempty" name:"CreateTime"`

	// 负载均衡实例的上次状态转换时间。
	// 注意：此字段可能返回 null，表示取不到有效值。
	StatusTime *string `json:"StatusTime,omitempty" name:"StatusTime"`

	// 负载均衡实例所属的项目 ID， 0 表示默认项目。
	ProjectId *uint64 `json:"ProjectId,omitempty" name:"ProjectId"`

	// 私有网络的 ID
	// 注意：此字段可能返回 null，表示取不到有效值。
	VpcId *string `json:"VpcId,omitempty" name:"VpcId"`

	// 高防 LB 的标识，1：高防负载均衡 0：非高防负载均衡。
	// 注意：此字段可能返回 null，表示取不到有效值。
	OpenBgp *uint64 `json:"OpenBgp,omitempty" name:"OpenBgp"`

	// 在 2016 年 12 月份之前的传统型内网负载均衡都是开启了 snat 的。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Snat *bool `json:"Snat,omitempty" name:"Snat"`

	// 0：表示未被隔离，1：表示被隔离。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Isolation *uint64 `json:"Isolation,omitempty" name:"Isolation"`

	// 用户开启日志的信息，日志只有公网属性创建了 HTTP 、HTTPS 监听器的负载均衡才会有日志。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Log *string `json:"Log,omitempty" name:"Log"`

	// 负载均衡实例所在的子网（仅对内网VPC型LB有意义）
	// 注意：此字段可能返回 null，表示取不到有效值。
	SubnetId *string `json:"SubnetId,omitempty" name:"SubnetId"`

	// 负载均衡实例的标签信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	Tags []*TagInfo `json:"Tags,omitempty" name:"Tags" list`

	// 负载均衡实例的安全组
	// 注意：此字段可能返回 null，表示取不到有效值。
	SecureGroups []*string `json:"SecureGroups,omitempty" name:"SecureGroups" list`

	// 负载均衡实例绑定的后端设备的基本信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	TargetRegionInfo *TargetRegionInfo `json:"TargetRegionInfo,omitempty" name:"TargetRegionInfo"`

	// anycast负载均衡的发布域，对于非anycast的负载均衡，此字段返回为空字符串
	// 注意：此字段可能返回 null，表示取不到有效值。
	AnycastZone *string `json:"AnycastZone,omitempty" name:"AnycastZone"`

	// IP版本，ipv4 | ipv6
	// 注意：此字段可能返回 null，表示取不到有效值。
	AddressIPVersion *string `json:"AddressIPVersion,omitempty" name:"AddressIPVersion"`

	// 数值形式的私有网络 ID
	// 注意：此字段可能返回 null，表示取不到有效值。
	NumericalVpcId *uint64 `json:"NumericalVpcId,omitempty" name:"NumericalVpcId"`

	// 负载均衡IP地址所属的ISP
	// 注意：此字段可能返回 null，表示取不到有效值。
	VipIsp *string `json:"VipIsp,omitempty" name:"VipIsp"`

	// 主可用区
	// 注意：此字段可能返回 null，表示取不到有效值。
	MasterZone *ZoneInfo `json:"MasterZone,omitempty" name:"MasterZone"`

	// 备可用区
	// 注意：此字段可能返回 null，表示取不到有效值。
	BackupZoneSet []*ZoneInfo `json:"BackupZoneSet,omitempty" name:"BackupZoneSet" list`

	// 负载均衡实例被隔离的时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	IsolatedTime *string `json:"IsolatedTime,omitempty" name:"IsolatedTime"`

	// 负载均衡实例的过期时间，仅对预付费负载均衡生效
	// 注意：此字段可能返回 null，表示取不到有效值。
	ExpireTime *string `json:"ExpireTime,omitempty" name:"ExpireTime"`

	// 负载均衡实例的计费类型
	// 注意：此字段可能返回 null，表示取不到有效值。
	ChargeType *string `json:"ChargeType,omitempty" name:"ChargeType"`

	// 负载均衡实例的网络属性
	// 注意：此字段可能返回 null，表示取不到有效值。
	NetworkAttributes *InternetAccessible `json:"NetworkAttributes,omitempty" name:"NetworkAttributes"`

	// 负载均衡实例的预付费相关属性
	// 注意：此字段可能返回 null，表示取不到有效值。
	PrepaidAttributes *LBChargePrepaid `json:"PrepaidAttributes,omitempty" name:"PrepaidAttributes"`

	// 负载均衡日志服务(CLS)的日志集ID
	// 注意：此字段可能返回 null，表示取不到有效值。
	LogSetId *string `json:"LogSetId,omitempty" name:"LogSetId"`

	// 负载均衡日志服务(CLS)的日志主题ID
	// 注意：此字段可能返回 null，表示取不到有效值。
	LogTopicId *string `json:"LogTopicId,omitempty" name:"LogTopicId"`

	// 负载均衡实例的IPv6地址
	// 注意：此字段可能返回 null，表示取不到有效值。
	AddressIPv6 *string `json:"AddressIPv6,omitempty" name:"AddressIPv6"`

	// 暂做保留，一般用户无需关注。
	// 注意：此字段可能返回 null，表示取不到有效值。
	ExtraInfo *ExtraInfo `json:"ExtraInfo,omitempty" name:"ExtraInfo"`

	// 是否可绑定高防包
	// 注意：此字段可能返回 null，表示取不到有效值。
	IsDDos *bool `json:"IsDDos,omitempty" name:"IsDDos"`

	// 负载均衡维度的个性化配置ID
	// 注意：此字段可能返回 null，表示取不到有效值。
	ConfigId *string `json:"ConfigId,omitempty" name:"ConfigId"`
}

type LoadBalancerHealth struct {

	// 负载均衡实例ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 负载均衡实例名称
	// 注意：此字段可能返回 null，表示取不到有效值。
	LoadBalancerName *string `json:"LoadBalancerName,omitempty" name:"LoadBalancerName"`

	// 监听器列表
	// 注意：此字段可能返回 null，表示取不到有效值。
	Listeners []*ListenerHealth `json:"Listeners,omitempty" name:"Listeners" list`
}

type ManualRewriteRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 源监听器ID
	SourceListenerId *string `json:"SourceListenerId,omitempty" name:"SourceListenerId"`

	// 目标监听器ID
	TargetListenerId *string `json:"TargetListenerId,omitempty" name:"TargetListenerId"`

	// 转发规则之间的重定向关系
	RewriteInfos []*RewriteLocationMap `json:"RewriteInfos,omitempty" name:"RewriteInfos" list`
}

func (r *ManualRewriteRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ManualRewriteRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ManualRewriteResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ManualRewriteResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ManualRewriteResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ModifyDomainAttributesRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 负载均衡监听器 ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 域名（必须是已经创建的转发规则下的域名）
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// 要修改的新域名
	NewDomain *string `json:"NewDomain,omitempty" name:"NewDomain"`

	// 域名相关的证书信息，注意，仅对启用SNI的监听器适用。
	Certificate *CertificateInput `json:"Certificate,omitempty" name:"Certificate"`

	// 是否开启Http2，注意，只有HTTPS域名才能开启Http2。
	Http2 *bool `json:"Http2,omitempty" name:"Http2"`

	// 是否设为默认域名，注意，一个监听器下只能设置一个默认域名。
	DefaultServer *bool `json:"DefaultServer,omitempty" name:"DefaultServer"`
}

func (r *ModifyDomainAttributesRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ModifyDomainAttributesRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ModifyDomainAttributesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyDomainAttributesResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ModifyDomainAttributesResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ModifyDomainRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 负载均衡监听器 ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 监听器下的某个旧域名。
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// 新域名，	长度限制为：1-120。有三种使用格式：非正则表达式格式，通配符格式，正则表达式格式。非正则表达式格式只能使用字母、数字、‘-’、‘.’。通配符格式的使用 ‘*’ 只能在开头或者结尾。正则表达式以'~'开头。
	NewDomain *string `json:"NewDomain,omitempty" name:"NewDomain"`
}

func (r *ModifyDomainRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ModifyDomainRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ModifyDomainResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyDomainResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ModifyDomainResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ModifyListenerRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 负载均衡监听器 ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 新的监听器名称
	ListenerName *string `json:"ListenerName,omitempty" name:"ListenerName"`

	// 会话保持时间，单位：秒。可选值：30~3600，默认 0，表示不开启。此参数仅适用于TCP/UDP监听器。
	SessionExpireTime *int64 `json:"SessionExpireTime,omitempty" name:"SessionExpireTime"`

	// 健康检查相关参数，此参数仅适用于TCP/UDP/TCP_SSL监听器
	HealthCheck *HealthCheck `json:"HealthCheck,omitempty" name:"HealthCheck"`

	// 证书相关信息，此参数仅适用于HTTPS/TCP_SSL监听器
	Certificate *CertificateInput `json:"Certificate,omitempty" name:"Certificate"`

	// 监听器转发的方式。可选值：WRR、LEAST_CONN
	// 分别表示按权重轮询、最小连接数， 默认为 WRR。
	Scheduler *string `json:"Scheduler,omitempty" name:"Scheduler"`

	// 是否开启SNI特性，此参数仅适用于HTTPS监听器。注意：未开启SNI的监听器可以开启SNI；已开启SNI的监听器不能关闭SNI
	SniSwitch *int64 `json:"SniSwitch,omitempty" name:"SniSwitch"`
}

func (r *ModifyListenerRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ModifyListenerRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ModifyListenerResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyListenerResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ModifyListenerResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ModifyLoadBalancerAttributesRequest struct {
	*tchttp.BaseRequest

	// 负载均衡的唯一ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 负载均衡实例名称
	LoadBalancerName *string `json:"LoadBalancerName,omitempty" name:"LoadBalancerName"`

	// 负载均衡绑定的后端服务的地域信息
	TargetRegionInfo *TargetRegionInfo `json:"TargetRegionInfo,omitempty" name:"TargetRegionInfo"`

	// 网络计费相关参数，注意，目前只支持修改最大出带宽，不支持修改网络计费方式。
	InternetChargeInfo *InternetAccessible `json:"InternetChargeInfo,omitempty" name:"InternetChargeInfo"`
}

func (r *ModifyLoadBalancerAttributesRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ModifyLoadBalancerAttributesRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ModifyLoadBalancerAttributesResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyLoadBalancerAttributesResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ModifyLoadBalancerAttributesResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ModifyRuleRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 负载均衡监听器 ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 要修改的转发规则的 ID。
	LocationId *string `json:"LocationId,omitempty" name:"LocationId"`

	// 转发规则的新的转发路径，如不需修改Url，则不需提供此参数
	Url *string `json:"Url,omitempty" name:"Url"`

	// 健康检查信息
	HealthCheck *HealthCheck `json:"HealthCheck,omitempty" name:"HealthCheck"`

	// 规则的请求转发方式，可选值：WRR、LEAST_CONN、IP_HASH
	// 分别表示按权重轮询、最小连接数、按IP哈希， 默认为 WRR。
	Scheduler *string `json:"Scheduler,omitempty" name:"Scheduler"`

	// 会话保持时间
	SessionExpireTime *int64 `json:"SessionExpireTime,omitempty" name:"SessionExpireTime"`

	// 负载均衡实例与后端服务之间的转发协议，默认HTTP，可取值：HTTP、HTTPS
	ForwardType *string `json:"ForwardType,omitempty" name:"ForwardType"`
}

func (r *ModifyRuleRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ModifyRuleRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ModifyRuleResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyRuleResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ModifyRuleResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ModifyTargetPortRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 负载均衡监听器 ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 要修改端口的后端服务列表
	Targets []*Target `json:"Targets,omitempty" name:"Targets" list`

	// 后端服务绑定到监听器或转发规则的新端口
	NewPort *int64 `json:"NewPort,omitempty" name:"NewPort"`

	// 转发规则的ID，当后端服务绑定到七层转发规则时，必须提供此参数或Domain+Url两者之一
	LocationId *string `json:"LocationId,omitempty" name:"LocationId"`

	// 目标规则的域名，提供LocationId参数时本参数不生效
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// 目标规则的URL，提供LocationId参数时本参数不生效
	Url *string `json:"Url,omitempty" name:"Url"`
}

func (r *ModifyTargetPortRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ModifyTargetPortRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ModifyTargetPortResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyTargetPortResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ModifyTargetPortResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ModifyTargetWeightRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 负载均衡监听器 ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 转发规则的ID，当绑定机器到七层转发规则时，必须提供此参数或Domain+Url两者之一
	LocationId *string `json:"LocationId,omitempty" name:"LocationId"`

	// 目标规则的域名，提供LocationId参数时本参数不生效
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// 目标规则的URL，提供LocationId参数时本参数不生效
	Url *string `json:"Url,omitempty" name:"Url"`

	// 要修改权重的后端服务列表
	Targets []*Target `json:"Targets,omitempty" name:"Targets" list`

	// 后端服务新的转发权重，取值范围：0~100，默认值10。如果设置了 Targets.Weight 参数，则此参数不生效。
	Weight *int64 `json:"Weight,omitempty" name:"Weight"`
}

func (r *ModifyTargetWeightRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ModifyTargetWeightRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ModifyTargetWeightResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ModifyTargetWeightResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ModifyTargetWeightResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type RegisterTargetsRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 负载均衡监听器 ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 待绑定的后端服务列表，数组长度最大支持20
	Targets []*Target `json:"Targets,omitempty" name:"Targets" list`

	// 转发规则的ID，当绑定后端服务到七层转发规则时，必须提供此参数或Domain+Url两者之一
	LocationId *string `json:"LocationId,omitempty" name:"LocationId"`

	// 目标转发规则的域名，提供LocationId参数时本参数不生效
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// 目标转发规则的URL，提供LocationId参数时本参数不生效
	Url *string `json:"Url,omitempty" name:"Url"`
}

func (r *RegisterTargetsRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *RegisterTargetsRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type RegisterTargetsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *RegisterTargetsResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *RegisterTargetsResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type RegisterTargetsWithClassicalLBRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 后端服务信息
	Targets []*ClassicalTargetInfo `json:"Targets,omitempty" name:"Targets" list`
}

func (r *RegisterTargetsWithClassicalLBRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *RegisterTargetsWithClassicalLBRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type RegisterTargetsWithClassicalLBResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *RegisterTargetsWithClassicalLBResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *RegisterTargetsWithClassicalLBResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ReplaceCertForLoadBalancersRequest struct {
	*tchttp.BaseRequest

	// 需要被替换的证书的ID，可以是服务端证书或客户端证书
	OldCertificateId *string `json:"OldCertificateId,omitempty" name:"OldCertificateId"`

	// 新证书的内容等相关信息
	Certificate *CertificateInput `json:"Certificate,omitempty" name:"Certificate"`
}

func (r *ReplaceCertForLoadBalancersRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ReplaceCertForLoadBalancersRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type ReplaceCertForLoadBalancersResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *ReplaceCertForLoadBalancersResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *ReplaceCertForLoadBalancersResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type RewriteLocationMap struct {

	// 源转发规则ID
	SourceLocationId *string `json:"SourceLocationId,omitempty" name:"SourceLocationId"`

	// 重定向至的目标转发规则ID
	TargetLocationId *string `json:"TargetLocationId,omitempty" name:"TargetLocationId"`
}

type RewriteTarget struct {

	// 重定向目标的监听器ID
	// 注意：此字段可能返回 null，表示无重定向。
	// 注意：此字段可能返回 null，表示取不到有效值。
	TargetListenerId *string `json:"TargetListenerId,omitempty" name:"TargetListenerId"`

	// 重定向目标的转发规则ID
	// 注意：此字段可能返回 null，表示无重定向。
	// 注意：此字段可能返回 null，表示取不到有效值。
	TargetLocationId *string `json:"TargetLocationId,omitempty" name:"TargetLocationId"`
}

type RsWeightRule struct {

	// 负载均衡监听器 ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 要修改权重的后端机器列表
	Targets []*Target `json:"Targets,omitempty" name:"Targets" list`

	// 转发规则的ID，七层规则时需要此参数，4层规则不需要
	LocationId *string `json:"LocationId,omitempty" name:"LocationId"`

	// 目标规则的域名，提供LocationId参数时本参数不生效
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// 目标规则的URL，提供LocationId参数时本参数不生效
	Url *string `json:"Url,omitempty" name:"Url"`

	// 后端服务新的转发权重，取值范围：0~100。
	Weight *int64 `json:"Weight,omitempty" name:"Weight"`
}

type RuleHealth struct {

	// 转发规则ID
	LocationId *string `json:"LocationId,omitempty" name:"LocationId"`

	// 转发规则的域名
	// 注意：此字段可能返回 null，表示取不到有效值。
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// 转发规则的Url
	// 注意：此字段可能返回 null，表示取不到有效值。
	Url *string `json:"Url,omitempty" name:"Url"`

	// 本规则上绑定的后端的健康检查状态
	// 注意：此字段可能返回 null，表示取不到有效值。
	Targets []*TargetHealth `json:"Targets,omitempty" name:"Targets" list`
}

type RuleInput struct {

	// 转发规则的域名。长度限制为：1~80。
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// 转发规则的路径。长度限制为：1~200。
	Url *string `json:"Url,omitempty" name:"Url"`

	// 会话保持时间。设置为0表示关闭会话保持，开启会话保持可取值30~3600，单位：秒。
	SessionExpireTime *int64 `json:"SessionExpireTime,omitempty" name:"SessionExpireTime"`

	// 健康检查信息
	HealthCheck *HealthCheck `json:"HealthCheck,omitempty" name:"HealthCheck"`

	// 证书信息
	Certificate *CertificateInput `json:"Certificate,omitempty" name:"Certificate"`

	// 规则的请求转发方式，可选值：WRR、LEAST_CONN、IP_HASH
	// 分别表示按权重轮询、最小连接数、按IP哈希， 默认为 WRR。
	Scheduler *string `json:"Scheduler,omitempty" name:"Scheduler"`

	// 负载均衡与后端服务之间的转发协议，目前支持 HTTP
	ForwardType *string `json:"ForwardType,omitempty" name:"ForwardType"`

	// 是否将该域名设为默认域名，注意，一个监听器下只能设置一个默认域名。
	DefaultServer *bool `json:"DefaultServer,omitempty" name:"DefaultServer"`

	// 是否开启Http2，注意，只用HTTPS域名才能开启Http2。
	Http2 *bool `json:"Http2,omitempty" name:"Http2"`

	// 后端目标类型，NODE表示绑定普通节点，TARGETGROUP表示绑定目标组
	TargetType *string `json:"TargetType,omitempty" name:"TargetType"`
}

type RuleOutput struct {

	// 转发规则的 ID
	LocationId *string `json:"LocationId,omitempty" name:"LocationId"`

	// 转发规则的域名。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// 转发规则的路径。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Url *string `json:"Url,omitempty" name:"Url"`

	// 会话保持时间
	SessionExpireTime *int64 `json:"SessionExpireTime,omitempty" name:"SessionExpireTime"`

	// 健康检查信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	HealthCheck *HealthCheck `json:"HealthCheck,omitempty" name:"HealthCheck"`

	// 证书信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	Certificate *CertificateOutput `json:"Certificate,omitempty" name:"Certificate"`

	// 规则的请求转发方式
	Scheduler *string `json:"Scheduler,omitempty" name:"Scheduler"`

	// 转发规则所属的监听器 ID
	ListenerId *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 转发规则的重定向目标信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	RewriteTarget *RewriteTarget `json:"RewriteTarget,omitempty" name:"RewriteTarget"`

	// 是否开启gzip
	HttpGzip *bool `json:"HttpGzip,omitempty" name:"HttpGzip"`

	// 转发规则是否为自动创建
	BeAutoCreated *bool `json:"BeAutoCreated,omitempty" name:"BeAutoCreated"`

	// 是否作为默认域名
	DefaultServer *bool `json:"DefaultServer,omitempty" name:"DefaultServer"`

	// 是否开启Http2
	Http2 *bool `json:"Http2,omitempty" name:"Http2"`

	// 负载均衡与后端服务之间的转发协议
	ForwardType *string `json:"ForwardType,omitempty" name:"ForwardType"`

	// 转发规则的创建时间
	CreateTime *string `json:"CreateTime,omitempty" name:"CreateTime"`
}

type RuleTargets struct {

	// 转发规则的 ID
	LocationId *string `json:"LocationId,omitempty" name:"LocationId"`

	// 转发规则的域名
	Domain *string `json:"Domain,omitempty" name:"Domain"`

	// 转发规则的路径。
	Url *string `json:"Url,omitempty" name:"Url"`

	// 后端服务的信息
	// 注意：此字段可能返回 null，表示取不到有效值。
	Targets []*Backend `json:"Targets,omitempty" name:"Targets" list`
}

type SetLoadBalancerSecurityGroupsRequest struct {
	*tchttp.BaseRequest

	// 负载均衡实例 ID
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 安全组ID构成的数组，一个负载均衡实例最多可绑定50个安全组，如果要解绑所有安全组，可不传此参数，或传入空数组。
	SecurityGroups []*string `json:"SecurityGroups,omitempty" name:"SecurityGroups" list`
}

func (r *SetLoadBalancerSecurityGroupsRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *SetLoadBalancerSecurityGroupsRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type SetLoadBalancerSecurityGroupsResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *SetLoadBalancerSecurityGroupsResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *SetLoadBalancerSecurityGroupsResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type SetSecurityGroupForLoadbalancersRequest struct {
	*tchttp.BaseRequest

	// 安全组ID，如 sg-12345678
	SecurityGroup *string `json:"SecurityGroup,omitempty" name:"SecurityGroup"`

	// ADD 绑定安全组；
	// DEL 解绑安全组
	OperationType *string `json:"OperationType,omitempty" name:"OperationType"`

	// 负载均衡实例ID数组
	LoadBalancerIds []*string `json:"LoadBalancerIds,omitempty" name:"LoadBalancerIds" list`
}

func (r *SetSecurityGroupForLoadbalancersRequest) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *SetSecurityGroupForLoadbalancersRequest) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type SetSecurityGroupForLoadbalancersResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *SetSecurityGroupForLoadbalancersResponse) ToJsonString() string {
    b, _ := json.Marshal(r)
    return string(b)
}

func (r *SetSecurityGroupForLoadbalancersResponse) FromJsonString(s string) error {
    return json.Unmarshal([]byte(s), &r)
}

type TagInfo struct {

	// 标签的键
	TagKey *string `json:"TagKey,omitempty" name:"TagKey"`

	// 标签的值
	TagValue *string `json:"TagValue,omitempty" name:"TagValue"`
}

type Target struct {

	// 后端服务的监听端口
	// 注意：此字段可能返回 null，表示取不到有效值。
	Port *int64 `json:"Port,omitempty" name:"Port"`

	// 后端服务的类型，可取：CVM（云服务器）、ENI（弹性网卡）；作为入参时，目前本参数暂不生效。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Type *string `json:"Type,omitempty" name:"Type"`

	// 绑定CVM时需要传入此参数，代表CVM的唯一 ID，可通过 DescribeInstances 接口返回字段中的 InstanceId 字段获取。
	// 注意：参数 InstanceId 和 EniIp 只能传入一个且必须传入一个。
	// 注意：此字段可能返回 null，表示取不到有效值。
	InstanceId *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 后端服务的转发权重，取值范围：[0, 100]，默认为 10。
	Weight *int64 `json:"Weight,omitempty" name:"Weight"`

	// 绑定弹性网卡时需要传入此参数，代表弹性网卡的IP，弹性网卡必须先绑定至CVM，然后才能绑定到负载均衡实例。注意：参数 InstanceId 和 EniIp 只能传入一个且必须传入一个。注意：绑定弹性网卡需要先提交工单开白名单使用。
	// 注意：此字段可能返回 null，表示取不到有效值。
	EniIp *string `json:"EniIp,omitempty" name:"EniIp"`
}

type TargetHealth struct {

	// Target的内网IP
	IP *string `json:"IP,omitempty" name:"IP"`

	// Target绑定的端口
	Port *int64 `json:"Port,omitempty" name:"Port"`

	// 当前健康状态，true：健康，false：不健康（包括尚未开始探测、探测中、状态异常等几种状态）。只有处于健康状态（且权重大于0），负载均衡才会向其转发流量。
	HealthStatus *bool `json:"HealthStatus,omitempty" name:"HealthStatus"`

	// Target的实例ID，如 ins-12345678
	TargetId *string `json:"TargetId,omitempty" name:"TargetId"`

	// 当前健康状态的详细信息。如：Alive、Dead、Unknown。Alive状态为健康，Dead状态为异常，Unknown状态包括尚未开始探测、探测中、状态未知。
	HealthStatusDetial *string `json:"HealthStatusDetial,omitempty" name:"HealthStatusDetial"`
}

type TargetRegionInfo struct {

	// Target所属地域，如 ap-guangzhou
	Region *string `json:"Region,omitempty" name:"Region"`

	// Target所属网络，私有网络格式如 vpc-abcd1234，如果是基础网络，则为"0"
	VpcId *string `json:"VpcId,omitempty" name:"VpcId"`
}

type ZoneInfo struct {

	// 可用区数值形式的唯一ID，如：100001
	// 注意：此字段可能返回 null，表示取不到有效值。
	ZoneId *uint64 `json:"ZoneId,omitempty" name:"ZoneId"`

	// 可用区字符串形式的唯一ID，如：ap-guangzhou-1
	// 注意：此字段可能返回 null，表示取不到有效值。
	Zone *string `json:"Zone,omitempty" name:"Zone"`

	// 可用区名称，如：广州一区
	// 注意：此字段可能返回 null，表示取不到有效值。
	ZoneName *string `json:"ZoneName,omitempty" name:"ZoneName"`
}
