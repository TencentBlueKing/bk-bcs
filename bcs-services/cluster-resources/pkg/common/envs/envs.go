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

// Package envs xxx
package envs

import (
	"path/filepath"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/envx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/path"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// 以下变量值可通过环境变量指定
var (
	// ExampleFileBaseDir Example 配置文件目录
	ExampleFileBaseDir = envx.Get(
		"EXAMPLE_FILE_BASE_DIR", filepath.Dir(filepath.Dir(path.GetCurPKGPath()))+"/resource/example",
	)
	// FormTmplFileBaseDir 表单化相关文件目录
	FormTmplFileBaseDir = envx.Get(
		"FORM_TMPL_FILE_BASE_DIR", filepath.Dir(filepath.Dir(path.GetCurPKGPath()))+"/resource/form/tmpl",
	)
	// LocalizeFilePath 国际化配置文件
	LocalizeFilePath = envx.Get(
		"LOCALIZE_FILE_PATH", filepath.Dir(filepath.Dir(path.GetCurPKGPath()))+"/i18n/locale/lc_msgs.yaml",
	)
	// BCSApiGWAuthToken 网关 Auth Token（仅挂载环境变量模式使用，来源为 Secret）
	BCSApiGWAuthToken = envx.Get("BCS_API_GW_AUTH_TOKEN", "")
	// LocalIP 服务运行 Pod IP，容器化部署时候指定
	LocalIP = envx.Get("LOCAL_IP", "")
	// ipv6本地网关地址
	IPv6Interface = envx.Get("IPV6_INTERFACE", "")
	// 双栈集群, 多个ip地址
	PodIPs = envx.Get("POD_IPs", "")
	// charts 环境变量配置
	BKAppCode        = envx.Get("BK_APP_CODE", "")
	BKAppSecret      = envx.Get("BK_APP_SECRET", "")
	BKPaaSHost       = envx.Get("BK_PAAS_HOST", "")
	BKIAMHost        = envx.Get("BK_IAM_HOST", "")
	BKIAMGatewayHost = envx.Get("BK_IAM_GATEWAY_HOST", "")
	BKIAMSystemID    = envx.Get("BK_IAM_SYSTEM_ID", "")
	RedisPassword    = envx.Get("REDIS_PASSWORD", "")
	MongoAddress     = envx.Get("MONGO_ADDRESS", "")
	MongoReplicaset  = envx.Get("MONGO_REPLICASET", "")
	MongoUsername    = envx.Get("MONGO_USERNAME", "")
	MongoPassword    = envx.Get("MONGO_PASSWORD", "")
)

// 以下变量值可通过环境变量指定（仅用于单元测试）
var (
	// AnonymousUsername 匿名用户
	AnonymousUsername = envx.Get("ANONYMOUS_USERNAME", "anonymous")
	// TestProjectID 单测指定的项目 ID
	TestProjectID = envx.Get("TEST_PROJECT_ID", stringx.Rand(32, ""))
	// TestProjectCode 单测指定的项目 Code
	TestProjectCode = envx.Get("TEST_PROJECT_CODE", "blueking")
	// TestClusterID 单测指定的集群 ID
	TestClusterID = envx.Get("TEST_CLUSTER_ID", "BCS-K8S-T"+stringx.Rand(5, "1234567890"))
	// TestNamespace 单测指定的命名空间
	TestNamespace = envx.Get("TEST_NAMESPACE", "default")
	// TestSharedClusterID 单测指定的共享集群 ID
	TestSharedClusterID = envx.Get("TEST_SHARED_CLUSTER_ID", "BCS-K8S-S99999")
	// TestSharedClusterNS 单测指定的共享集群中的命名空间
	TestSharedClusterNS = envx.Get("TEST_SHARED_CLUSTER_NS", "ieg-"+TestProjectCode+"-shared-t533")
	// TestNoPermClusterID 单测指定没有权限的集群 ID
	TestNoPermClusterID = envx.Get("TEST_NO_PERM_CLUSTER_ID", "BCS-K8S-NP8888")
)
