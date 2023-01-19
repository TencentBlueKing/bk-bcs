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
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// HPAComplex 单元测试用 HPA 表单数据(全量)
var HPAComplex = model.HPA{
	Metadata: model.Metadata{
		APIVersion: "autoscaling/v2beta2",
		Kind:       resCsts.HPA,
		Name:       "hpa-complex-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
		},
		Annotations: []model.Annotation{
			{"anno-key-1", "anno-val-1"},
		},
	},
	Spec: model.HPASpec{
		Ref: model.HPATargetRef{
			APIVersion:  "apps/v1",
			Kind:        "Deployment",
			ResName:     "deployment-xxx1",
			MinReplicas: 3,
			MaxReplicas: 8,
		},
		Resource: model.ResourceMetric{
			Items: []model.ResourceMetricItem{
				{
					Name:   resCsts.MetricResCPU,
					Type:   resCsts.HPATargetTypeAverageValue,
					CPUVal: 500,
				},
				{
					Name:    "gpu",
					Type:    resCsts.HPATargetTypeUtilization,
					Percent: 50,
				},
				{
					Name:   resCsts.MetricResMem,
					Type:   resCsts.HPATargetTypeAverageValue,
					MEMVal: 512,
				},
			},
		},
		External: model.ExternalMetric{
			Items: []model.ExternalMetricItem{
				{
					Name:  "ext1",
					Type:  resCsts.HPATargetTypeValue,
					Value: "10",
					Selector: model.MetricSelector{
						Expressions: []model.ExpSelector{
							{
								Key:    "exp1",
								Op:     "In",
								Values: "value1",
							},
							{
								Key:    "exp2",
								Op:     "NotIn",
								Values: "value2,value2-1",
							},
							{
								Key: "exp3",
								Op:  "Exists",
							},
						},
						Labels: []model.LabelSelector{
							{
								Key:   "key1",
								Value: "val1",
							},
							{
								Key:   "key2",
								Value: "val2",
							},
						},
					},
				},
			},
		},
		Object: model.ObjectMetric{
			Items: []model.ObjectMetricItem{
				{
					Name:       "object1",
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					ResName:    "deploy-aaa",
					Type:       resCsts.HPATargetTypeAverageValue,
					Value:      "10",
					Selector: model.MetricSelector{
						Expressions: []model.ExpSelector{
							{
								Key:    "exp1",
								Op:     "In",
								Values: "val1,val2",
							},
							{
								Key: "exp2",
								Op:  "Exists",
							},
						},
						Labels: []model.LabelSelector{
							{
								Key:   "key1",
								Value: "val1",
							},
						},
					},
				},
				{
					Name:       "object2",
					APIVersion: "tkex.tencent.com/v1alpha1",
					Kind:       "GameDeployment",
					ResName:    "gdeploy-xx",
					Type:       resCsts.HPATargetTypeValue,
					Value:      "20",
					Selector: model.MetricSelector{
						Expressions: []model.ExpSelector{
							{
								Key:    "exp1",
								Op:     "NotIn",
								Values: "val1,val2",
							},
						},
						Labels: []model.LabelSelector{
							{
								Key:   "key2",
								Value: "val2",
							},
						},
					},
				},
			},
		},
		Pod: model.PodMetric{
			Items: []model.PodMetricItem{
				{
					Name:  "pod1",
					Type:  resCsts.HPATargetTypeAverageValue,
					Value: "10",
					Selector: model.MetricSelector{
						Expressions: []model.ExpSelector{
							{
								Key: "exp1",
								Op:  "Exists",
							},
							{
								Key:    "exp2",
								Op:     "In",
								Values: "val1,val2",
							},
						},
						Labels: []model.LabelSelector{
							{
								Key:   "key11",
								Value: "val22",
							},
						},
					},
				},
			},
		},
	},
}

// HPASimple 单元测试用 HPA 表单数据(最简单版本)
var HPASimple = model.HPA{
	Metadata: model.Metadata{
		APIVersion: "autoscaling/v2beta2",
		Kind:       resCsts.HPA,
		Name:       "hpa-simple-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
		},
		Annotations: []model.Annotation{
			{"anno-key-1", "anno-val-1"},
		},
	},
	Spec: model.HPASpec{
		Ref: model.HPATargetRef{
			APIVersion:  "apps/v1",
			Kind:        "Deployment",
			ResName:     "deployment-6byc8q0oyc",
			MinReplicas: 1,
			MaxReplicas: 3,
		},
		Resource: model.ResourceMetric{
			Items: []model.ResourceMetricItem{
				{
					Name:   resCsts.MetricResCPU,
					Type:   resCsts.HPATargetTypeAverageValue,
					CPUVal: 500,
				},
			},
		},
	},
}
