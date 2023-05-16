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

package formdata

import (
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/util"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// GDeployComplex ...
var GDeployComplex = model.GDeploy{
	Metadata: model.Metadata{
		APIVersion: "tkex.tencent.com/v1alpha1",
		Kind:       resCsts.GDeploy,
		Name:       "gdeploy-complex-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
			{"label-key-2", "label-val-2"},
		},
		Annotations: []model.Annotation{
			{"anno-key-1", "anno-val-1"},
			{"anno-key-2", "anno-val-2"},
		},
	},
	Spec: model.GDeploySpec{
		Replicas: model.GDeployReplicas{
			Cnt:             2,
			UpdateStrategy:  resCsts.DefaultUpdateStrategy,
			MaxSurge:        0,
			MSUnit:          util.UnitCnt,
			MaxUnavailable:  20,
			MUAUnit:         util.UnitPercent,
			MinReadySecs:    0,
			Partition:       1,
			GracePeriodSecs: 30,
		},
		GracefulManage: model.GWorkloadGracefulManage{
			PreDeleteHook: model.GWorkloadHookSpec{
				Enabled:  true,
				TmplName: "hook-tmpl-1",
				Args:     []model.HookCallArg{{"test-key-1", "test-val-1"}},
			},
			PreInplaceHook: model.GWorkloadHookSpec{
				Enabled:  true,
				TmplName: "hook-tmpl-2",
				Args:     []model.HookCallArg{{"test-key-2", "test-val-2"}},
			},
		},
		DeletionProtect: model.GWorkloadDeletionProtect{
			Policy: resCsts.DeletionProtectPolicyNotAllow,
		},
		NodeSelect: nodeSelect,
		Affinity:   affinity,
		Toleration: toleration,
		Networking: networking,
		Security:   security,
		Other:      specOther,
	},
	ContainerGroup: model.ContainerGroup{
		InitContainers: initContainers,
		Containers:     containers,
	},
	Volume: volume,
}

// GDeploySimple ...
var GDeploySimple = model.GDeploy{
	Metadata: model.Metadata{
		APIVersion: "tkex.tencent.com/v1alpha1",
		Kind:       resCsts.GDeploy,
		Name:       "gdeploy-simple-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
		},
	},
	Spec: model.GDeploySpec{
		Replicas: model.GDeployReplicas{
			Cnt:            2,
			UpdateStrategy: resCsts.DefaultUpdateStrategy,
			MaxSurge:       1,
			MSUnit:         util.UnitCnt,
		},
		DeletionProtect: model.GWorkloadDeletionProtect{
			Policy: resCsts.DeletionProtectPolicyAlways,
		},
	},
	ContainerGroup: model.ContainerGroup{
		Containers: []model.Container{
			{
				Basic: model.ContainerBasic{
					Name:       "busybox",
					Image:      "busybox:latest",
					PullPolicy: "IfNotPresent",
				},
			},
		},
	},
}

// GSTSComplex ...
var GSTSComplex = model.GSTS{
	Metadata: model.Metadata{
		APIVersion: "tkex.tencent.com/v1alpha1",
		Kind:       resCsts.GSTS,
		Name:       "gsts-complex-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
			{"label-key-2", "label-val-2"},
		},
		Annotations: []model.Annotation{
			{"anno-key-1", "anno-val-1"},
			{"anno-key-2", "anno-val-2"},
		},
	},
	Spec: model.GSTSSpec{
		Replicas: model.GSTSReplicas{
			Cnt:             2,
			SVCName:         "svc-complex-y3xk1r9vg9",
			UpdateStrategy:  resCsts.DefaultUpdateStrategy,
			PodManPolicy:    "OrderedReady",
			Partition:       3,
			MaxSurge:        2,
			MSUnit:          util.UnitCnt,
			MaxUnavailable:  10,
			MUAUnit:         util.UnitPercent,
			GracePeriodSecs: 30,
		},
		GracefulManage: model.GWorkloadGracefulManage{
			PreDeleteHook: model.GWorkloadHookSpec{
				Enabled:  true,
				TmplName: "hook-tmpl-1",
				Args:     []model.HookCallArg{{"test-key-1", "test-val-1"}},
			},
			PostInplaceHook: model.GWorkloadHookSpec{
				Enabled:  true,
				TmplName: "hook-tmpl-3",
				Args:     []model.HookCallArg{{"test-key-3", "test-val-3"}},
			},
		},
		DeletionProtect: model.GWorkloadDeletionProtect{
			Policy: resCsts.DeletionProtectPolicyCascading,
		},
		NodeSelect: nodeSelect,
		Affinity:   affinity,
		Toleration: toleration,
		Networking: networking,
		Security:   security,
		Other:      specOther,
	},
	ContainerGroup: model.ContainerGroup{
		InitContainers: initContainers,
		Containers:     containers,
	},
	Volume: volume,
}

// GSTSSimple ...
var GSTSSimple = model.GSTS{
	Metadata: model.Metadata{
		APIVersion: "tkex.tencent.com/v1alpha1",
		Kind:       resCsts.GSTS,
		Name:       "gsts-complex-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
		},
	},
	Spec: model.GSTSSpec{
		Replicas: model.GSTSReplicas{
			Cnt:             2,
			SVCName:         "svc-complex-y3xk1r9vg9",
			UpdateStrategy:  resCsts.DefaultUpdateStrategy,
			PodManPolicy:    "OnDelete",
			Partition:       3,
			MaxSurge:        2,
			MSUnit:          util.UnitCnt,
			MaxUnavailable:  10,
			MUAUnit:         util.UnitPercent,
			GracePeriodSecs: 30,
		},
		DeletionProtect: model.GWorkloadDeletionProtect{
			Policy: resCsts.DeletionProtectPolicyCascading,
		},
	},
	ContainerGroup: model.ContainerGroup{
		Containers: []model.Container{
			{
				Basic: model.ContainerBasic{
					Name:       "busybox",
					Image:      "busybox:latest",
					PullPolicy: "IfNotPresent",
				},
			},
		},
	},
}

// HookTmplComplex ...
var HookTmplComplex = model.HookTmpl{
	Metadata: model.Metadata{
		APIVersion: "tkex.tencent.com/v1alpha1",
		Kind:       resCsts.HookTmpl,
		Name:       "hook-tmpl-complex-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
	},
	Spec: model.HookTmplSpec{
		Args: []model.HookTmplArg{
			{
				Key:   "customArg1",
				Value: "value1",
			},
			{
				Key:   "customArg2",
				Value: "",
			},
		},
		ExecPolicy:            "Ordered",
		DeletionProtectPolicy: resCsts.DeletionProtectPolicyAlways,
		Metrics: []model.HookTmplMetric{
			{
				Name:             "web",
				HookType:         resCsts.HookTmplMetricTypeWeb,
				URL:              "http://1.1.1.1:80",
				JSONPath:         "{$.result}",
				TimeoutSecs:      10,
				Count:            2,
				Interval:         1,
				SuccessCondition: "asInt(result) == 1",
				SuccessPolicy:    resCsts.HookTmplConsecutiveSuccessfulLimit,
				SuccessCnt:       1,
			},
			{
				Name:             "prom",
				HookType:         resCsts.HookTmplMetricTypeProm,
				Address:          "http://prometheus.com",
				Query:            "query_test",
				Count:            0,
				Interval:         1,
				SuccessCondition: "asInt(result) == 2",
				SuccessPolicy:    resCsts.HookTmplSuccessfulLimit,
				SuccessCnt:       1,
			},
			{
				Name:     "k8s",
				HookType: resCsts.HookTmplMetricTypeK8S,
				Function: "patch",
				Fields: []model.HookTmplField{
					{
						Key:   "metadata.name",
						Value: "resName-xx",
					},
				},
				Count:         0,
				Interval:      2,
				SuccessPolicy: resCsts.HookTmplConsecutiveSuccessfulLimit,
				SuccessCnt:    3,
			},
		},
	},
}
