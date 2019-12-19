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
 *
 */

package mesos

import (
	commtypes "bk-bcs/bcs-common/common/types"
	v2 "bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v2"
	listers "bk-bcs/bcs-services/bcs-webhook-server/pkg/client/listers/bk-bcs/v2"
)

type DbPrivConfInject struct {
	BcsDbPrivConfigLister listers.BcsDbPrivConfigLister
}

func NewDbPrivConfInject(bcsDbPrivConfLister listers.BcsDbPrivConfigLister) MesosInject {
	mesosInject := &DbPrivConfInject{
		BcsDbPrivConfigLister: bcsDbPrivConfLister,
	}

	return mesosInject
}

func (dbPrivConf *DbPrivConfInject) InjectApplicationContent(application *commtypes.ReplicaController) (*commtypes.ReplicaController, error) {

	return nil, nil
}

func (dbPrivConf *DbPrivConfInject) InjectDeployContent(deploy *commtypes.BcsDeployment) (*commtypes.BcsDeployment, error) {

	return nil, nil
}

func checkSelector(d *v2.BcsDbPrivConfig, labels map[string]string) bool {
	for ks, vs := range d.Spec.PodSelector {
		vt, ok := labels[ks]
		if !ok {
			return false
		}
		if vs != vt {
			return false
		}
	}

	return true
}
