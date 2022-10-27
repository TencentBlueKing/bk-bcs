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

package feature

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/version"
)

func TestGenFeatureGates(t *testing.T) {
	ver := version.Info{Major: "1", Minor: "20"}
	gates := GenFeatureGates(&ver)
	assert.True(t, gates[ImmutableEphemeralVolumes])

	ver = version.Info{Major: "1", Minor: "19"}
	gates = GenFeatureGates(&ver)
	assert.True(t, gates[ImmutableEphemeralVolumes])

	ver = version.Info{Major: "1", Minor: "18"}
	gates = GenFeatureGates(&ver)
	assert.False(t, gates[ImmutableEphemeralVolumes])
}
