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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeIntStr(t *testing.T) {
	val, unit := AnalyzeIntStr("25%")
	assert.Equal(t, int64(25), val)
	assert.Equal(t, UnitPercent, unit)

	val, unit = AnalyzeIntStr(int64(10))
	assert.Equal(t, int64(10), val)
	assert.Equal(t, UnitCnt, unit)
}

func TestConvertCPUUnit(t *testing.T) {
	assert.Equal(t, 1000, ConvertCPUUnit("1"))
	assert.Equal(t, 500, ConvertCPUUnit("0.5"))
	assert.Equal(t, 1200, ConvertCPUUnit("1200m"))
	assert.Equal(t, 100, ConvertCPUUnit("100m"))
}

func TestConvertMemoryUnit(t *testing.T) {
	assert.Equal(t, 100, ConvertMemoryUnit("100M"))
	assert.Equal(t, 120, ConvertMemoryUnit("120Mi"))
	assert.Equal(t, 1024, ConvertMemoryUnit("1G"))
	assert.Equal(t, 2048, ConvertMemoryUnit("2Gi"))
}
