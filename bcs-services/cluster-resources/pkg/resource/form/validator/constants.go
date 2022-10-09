/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package validator

import (
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
)

// FormSupportedResAPIVersion 支持表单化的资源版本
var FormSupportedResAPIVersion = map[string][]string{
	// 工作负载类
	res.Deploy: {"apps/v1", "extensions/v1", "extensions/v1beta1"},
	res.DS:     {"apps/v1", "extensions/v1", "extensions/v1beta1"},
	res.STS:    {"apps/v1"},
	res.CJ:     {"batch/v1", "batch/v1beta1"},
	res.Job:    {"batch/v1"},
	res.Po:     {"v1"},
	res.HPA:    {"autoscaling/v2beta2", "autoscaling/v2"},
	// 网络类
	res.Ing: {"networking.k8s.io/v1", "networking.k8s.io/v1beta1", "extensions/v1beta1"},
	res.SVC: {"v1"},
	res.EP:  {"v1"},
	// 存储类
	res.PV:  {"v1"},
	res.PVC: {"v1"},
	res.SC:  {"storage.k8s.io/v1"},
	// 配置类
	res.CM:     {"v1"},
	res.Secret: {"v1"},
	// 自定义资源
	res.GDeploy:  {"tkex.tencent.com/v1alpha1"},
	res.HookTmpl: {"tkex.tencent.com/v1alpha1"},
	res.GSTS:     {"tkex.tencent.com/v1alpha1"},
}

// FormSupportedCObjKinds 支持表单化的自定义资源
var FormSupportedCObjKinds = []string{res.GDeploy, res.HookTmpl, res.GSTS}
