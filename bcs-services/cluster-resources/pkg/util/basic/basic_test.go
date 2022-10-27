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

package basic_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/basic"
)

func TestMaxInt(t *testing.T) {
	assert.Equal(t, 1, basic.MaxInt(1, 1))
	assert.Equal(t, 2, basic.MaxInt(1, 2))
	assert.Equal(t, 3, basic.MaxInt(3, 2))
	assert.Equal(t, 0, basic.MaxInt(0, 0))
	assert.Equal(t, 3, basic.MaxInt(-1, 3))
}

func TestMinInt(t *testing.T) {
	assert.Equal(t, 1, basic.MinInt(1, 1))
	assert.Equal(t, 1, basic.MinInt(1, 2))
	assert.Equal(t, 2, basic.MinInt(3, 2))
	assert.Equal(t, 0, basic.MinInt(0, 0))
	assert.Equal(t, -1, basic.MinInt(-1, 3))
}
