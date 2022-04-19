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

package formatter

import (
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
)

// Kind2FormatFuncMap 各资源类型对应 FormatFunc
var Kind2FormatFuncMap = map[string]func(manifest map[string]interface{}) map[string]interface{}{
	// namespace
	resource.NS: FormatNS,

	// workload
	resource.CJ:     FormatCJ,
	resource.DS:     FormatWorkloadRes,
	resource.Deploy: FormatWorkloadRes,
	resource.Job:    FormatJob,
	resource.Po:     FormatPo,
	resource.STS:    FormatWorkloadRes,

	// network
	resource.Ing: FormatIng,
	resource.SVC: FormatSVC,
	resource.EP:  FormatEP,

	// configuration
	resource.CM:     FormatConfigRes,
	resource.Secret: FormatConfigRes,

	// storage
	resource.PV:  FormatPV,
	resource.PVC: FormatPVC,
	resource.SC:  FormatStorageRes,

	// rbac
	resource.SA: FormatSA,

	// hpa
	resource.HPA: FormatHPA,

	// CustomResource
	resource.CRD:  FormatCRD,
	resource.CObj: FormatCObj,
}
