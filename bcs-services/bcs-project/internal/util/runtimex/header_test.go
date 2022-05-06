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

package runtimex

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/headerkey"
)

func TestCustomMatcher(t *testing.T) {
	ret, _ := CustomHeaderMatcher(headerkey.RequestIDKey)
	assert.Equal(t, headerkey.RequestIDKey, ret)

	ret, _ = CustomHeaderMatcher(headerkey.UsernameKey)
	assert.Equal(t, headerkey.UsernameKey, ret)

	ret, _ = CustomHeaderMatcher("Content-Type")
	assert.Equal(t, "grpcgateway-Content-Type", ret)
}
