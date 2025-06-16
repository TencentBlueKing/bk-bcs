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

package validator

import resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"

// FormSupportedResAPIVersion 支持表单化的资源版本
var FormSupportedResAPIVersion = map[string][]string{
	// 工作负载类
	resCsts.Deploy: {"apps/v1", "extensions/v1", "extensions/v1beta1", "apps/v1beta2"},
	resCsts.DS:     {"apps/v1", "extensions/v1", "extensions/v1beta1", "apps/v1beta2"},
	resCsts.STS:    {"apps/v1", "apps/v1beta2"},
	resCsts.CJ:     {"batch/v1", "batch/v1beta1"},
	resCsts.Job:    {"batch/v1", "apps/v1beta2"},
	resCsts.Po:     {"v1"},
	resCsts.HPA:    {"autoscaling/v2", "autoscaling/v2beta2"},
	// 网络类
	resCsts.Ing: {"networking.k8s.io/v1", "networking.k8s.io/v1beta1", "extensions/v1beta1"},
	resCsts.SVC: {"v1"},
	resCsts.EP:  {"v1"},
	// 存储类
	resCsts.PV:  {"v1"},
	resCsts.PVC: {"v1"},
	resCsts.SC:  {"storage.k8s.io/v1"},
	// 配置类
	resCsts.CM:         {"v1"},
	resCsts.Secret:     {"v1"},
	resCsts.BscpConfig: {"bk.tencent.com/v1alpha1"},
	// 自定义资源
	resCsts.GDeploy:  {"tkex.tencent.com/v1alpha1"},
	resCsts.HookTmpl: {"tkex.tencent.com/v1alpha1"},
	resCsts.GSTS:     {"tkex.tencent.com/v1alpha1"},
}

// FormSupportedCObjKinds 支持表单化的自定义资源
var FormSupportedCObjKinds = []string{resCsts.GDeploy, resCsts.HookTmpl, resCsts.GSTS, resCsts.BscpConfig}
