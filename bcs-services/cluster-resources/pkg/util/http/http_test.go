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

package http_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/http"
)

func TestCustomHeaderMatcher(t *testing.T) {
	// 自定义头字段
	ret, _ := http.CustomHeaderMatcher("X-Request-Id")
	assert.Equal(t, "X-Request-Id", ret)

	// 标准头字段
	ret, _ = http.CustomHeaderMatcher("Content-Type")
	assert.Equal(t, "grpcgateway-Content-Type", ret)
}

func TestWSHeaderForwarder(t *testing.T) {
	// 允许转发的 Header Key
	assert.True(t, http.WSHeaderForwarder("Origin"))
	assert.True(t, http.WSHeaderForwarder("referer"))
	assert.True(t, http.WSHeaderForwarder("Authorization"))
	assert.True(t, http.WSHeaderForwarder("cookie"))

	// 暂不被转发的 Header Key
	assert.False(t, http.WSHeaderForwarder("Connection"))
	assert.False(t, http.WSHeaderForwarder("User-Agent"))
}
