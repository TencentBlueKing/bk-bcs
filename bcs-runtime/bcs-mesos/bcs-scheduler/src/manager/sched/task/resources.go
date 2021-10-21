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

package task

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	//"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	//"github.com/golang/protobuf/proto"
)

func createScalarResource(name string, value float64) *mesos.Resource {
	return &mesos.Resource{
		Name:   &name,
		Type:   mesos.Value_SCALAR.Enum(),
		Scalar: &mesos.Value_Scalar{Value: &value},
	}
}

// BuildResources Build mesos resource format
func BuildResources(r *types.Resource) []*mesos.Resource {
	var resources = []*mesos.Resource{}

	if r.Cpus > 0 {
		resources = append(resources, createScalarResource("cpus", r.Cpus))
	}

	if r.Mem > 0 {
		resources = append(resources, createScalarResource("mem", r.Mem))
	}

	if r.Disk > 0 {
		resources = append(resources, createScalarResource("disk", r.Disk))
	}

	return resources
}
