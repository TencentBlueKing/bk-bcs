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

package formatter

import (
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
)

// Kind2FormatFuncMap 各资源类型对应 FormatFunc
var Kind2FormatFuncMap = map[string]func(manifest map[string]interface{}) map[string]interface{}{
	// namespace
	resCsts.NS: FormatNS,

	// workload
	resCsts.CJ:     FormatCJ,
	resCsts.DS:     FormatWorkloadRes,
	resCsts.Deploy: FormatDeploy,
	resCsts.RS:     FormatRS,
	resCsts.Job:    FormatJob,
	resCsts.Po:     FormatPo,
	resCsts.STS:    FormatSTS,

	// network
	resCsts.Ing: FormatIng,
	resCsts.SVC: FormatSVC,
	resCsts.EP:  FormatEP,

	// configuration
	resCsts.CM:     FormatConfigRes,
	resCsts.Secret: FormatConfigRes,

	// storage
	resCsts.PV:  FormatPV,
	resCsts.PVC: FormatPVC,
	resCsts.SC:  FormatSC,

	// rbac
	resCsts.SA: FormatSA,

	// hpa
	resCsts.HPA: FormatHPA,

	// CustomResource
	resCsts.CRD:     FormatCRD,
	resCsts.GDeploy: FormatGDeploy,
	resCsts.GSTS:    FormatGSTS,
	resCsts.CObj:    FormatCObj,
}

// Kind2PruneFuncMap 各资源类型对应 PruneFunc
var Kind2PruneFuncMap = map[string]func(manifest map[string]interface{}) map[string]interface{}{
	resCsts.Deploy:   PruneDeploy,
	resCsts.STS:      PruneSTS,
	resCsts.DS:       PruneDS,
	resCsts.Job:      PruneJob,
	resCsts.CJ:       PruneCJ,
	resCsts.Po:       PrunePod,
	resCsts.Ing:      PruneIng,
	resCsts.SVC:      PruneSVC,
	resCsts.EP:       PruneEP,
	resCsts.CM:       PruneConfig,
	resCsts.Secret:   PruneConfig,
	resCsts.PV:       PrunePV,
	resCsts.PVC:      PrunePVC,
	resCsts.SC:       PruneSC,
	resCsts.SA:       PruneSA,
	resCsts.HPA:      PruneHPA,
	resCsts.GDeploy:  PruneGDeploy,
	resCsts.GSTS:     PruneGSTS,
	resCsts.HookTmpl: PruneHookTmpl,
	resCsts.CObj:     PruneCObj,
	resCsts.CRD:      PruneCRD,
}

const (
	// WorkloadStatusNormal 正常状态
	WorkloadStatusNormal = "normal"

	// WorkloadStatusCreating 创建中
	WorkloadStatusCreating = "creating"

	// WorkloadStatusUpdating 更新中
	WorkloadStatusUpdating = "updating"

	// WorkloadStatusRestarting 重启中
	WorkloadStatusRestarting = "restarting"

	// WorkloadStatusDeleting 删除中
	WorkloadStatusDeleting = "deleting"
)

const (
	// WorkloadRestartAnnotationKey 重启标识
	WorkloadRestartAnnotationKey = "kubectl.kubernetes.io/restartedAt"
	// WorkloadRestartVersionAnnotationKey 重启版本标识
	WorkloadRestartVersionAnnotationKey = "kubectl.kubernetes.io/restartedVersion"
)
