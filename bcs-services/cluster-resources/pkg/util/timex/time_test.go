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

package timex_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/timex"
)

func TestCalcDuration(t *testing.T) {
	// 普通时间格式
	assert.Equal(t, "1s", timex.CalcDuration("2022-01-01 12:00:00", "2022-01-01 12:00:01"))
	assert.Equal(t, "24m29s", timex.CalcDuration("2022-01-01 12:35:30", "2022-01-01 12:59:59"))
	assert.Equal(t, "1h24m", timex.CalcDuration("2022-01-01 12:35:30", "2022-01-01 14:00:00"))
	assert.Equal(t, "2d1h", timex.CalcDuration("2022-01-01 12:35:30", "2022-01-03 14:00:00"))
	assert.Equal(t, "153d3h", timex.CalcDuration("2021-08-01 11:00:00", "2022-01-01 14:00:00"))
	assert.Equal(t, "275d3h", timex.CalcDuration("2021-04-01 11:00:00", "2022-01-01 14:00:00"))
	assert.Equal(t, "640d3h", timex.CalcDuration("2020-04-01 11:00:00", "2022-01-01 14:00:00"))

	// k8s manifest 时间格式
	assert.Equal(t, "1s", timex.CalcDuration("2022-01-01T14:00:00Z", "2022-01-01T14:00:01Z"))
	assert.Equal(t, "14m29s", timex.CalcDuration("2022-01-01T14:45:30Z", "2022-01-01T14:59:59Z"))
	assert.Equal(t, "275d3h", timex.CalcDuration("2021-04-01T11:00:00Z", "2022-01-01T14:00:00Z"))
}

func TestCalcAge(t *testing.T) {
	// 存在时间会随运行时间而变化，这里直接比较大于 1000 天的时间
	age := timex.CalcAge("2019-01-01 11:00:00")
	dayCnt, _ := strconv.Atoi(strings.Split(age, "d")[0])
	assert.True(t, dayCnt > 1000)
}

func TestNormalizeDatetime(t *testing.T) {
	ret, _ := timex.NormalizeDatetime("2022-01-01T12:00:00Z")
	assert.Equal(t, "2022-01-01 12:00:00", ret)

	ret, _ = timex.NormalizeDatetime("2022-01-02 14:00:00")
	assert.Equal(t, "2022-01-02 14:00:00", ret)

	_, err := timex.NormalizeDatetime("3/1/2021 12:00:00")
	assert.Error(t, err)
}

func TestCurrent(t *testing.T) {
	ret := timex.Current()
	assert.NotEqual(t, 0, len(ret))
}
