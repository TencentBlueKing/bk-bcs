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
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"

	"github.com/stretchr/testify/assert"
)

func TestCreateScaleResource(t *testing.T) {
	resource := createScalarResource("cpu", 0.1)
	assert.Equal(t, *resource.Name, "cpu")
	assert.Equal(t, *resource.Scalar.Value, 0.1)
}

func TestBuildResource(t *testing.T) {
	resources := BuildResources(&types.Resource{Cpus: 0.1, Mem: 16, Disk: 10})
	assert.Equal(t, *resources[0].Name, "cpus")
	assert.Equal(t, *resources[1].Name, "mem")
	assert.Equal(t, *resources[2].Name, "disk")
}
