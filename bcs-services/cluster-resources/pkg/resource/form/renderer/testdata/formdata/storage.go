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

// PVComplex ...
var PVComplex = model.PV{
	Metadata: model.Metadata{
		APIVersion: "v1",
		Kind:       resCsts.PV,
		Name:       "pv-complex-" + strings.ToLower(stringx.Rand(10, "")),
	},
	Spec: model.PVSpec{
		Type:        resCsts.PVTypeLocalVolume,
		SCName:      "local-path",
		StorageSize: 3,
		AccessModes: []string{"ReadOnlyMany", "ReadWriteOnce"},
		LocalPath:   "/data0",
	},
}

// PVCComplex ...
var PVCComplex = model.PVC{
	Metadata: model.Metadata{
		APIVersion: "v1",
		Kind:       resCsts.PVC,
		Name:       "pvc-complex-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
	},
	Spec: model.PVCSpec{
		ClaimType:   resCsts.PVCTypeUseExistPV,
		PVName:      "task-pv-volume",
		SCName:      "local-path",
		StorageSize: 5,
		AccessModes: []string{"ReadOnlyMany", "ReadWriteMany"},
	},
}

// SCComplex ...
var SCComplex = model.SC{
	Metadata: model.Metadata{
		APIVersion: "storage.k8s.io/v1",
		Kind:       resCsts.SC,
		Name:       "sc-complex-" + strings.ToLower(stringx.Rand(10, "")),
	},
	Spec: model.SCSpec{
		SetAsDefault:      true,
		Provisioner:       "k8s.io/minikube-hostpath",
		VolumeBindingMode: "Immediate",
		ReclaimPolicy:     "Delete",
		Params: []model.SCParam{
			{"type", "io1"},
		},
		MountOpts: []string{"ro", "soft"},
	},
}
