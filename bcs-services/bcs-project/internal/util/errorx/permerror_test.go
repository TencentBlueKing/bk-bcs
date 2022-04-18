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

package errorx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	applyUrl = "http://iam.example.com"
	actionID = "projectView"
	hasPerm  = false
)

func TestPermissionDeniedError(t *testing.T) {
	// one message
	err := NewPermDeniedErr(applyUrl, actionID, hasPerm)
	assert.Equal(t, err.ApplyUrl(), applyUrl)
	assert.Equal(t, err.HasPerm(), hasPerm)
	assert.Equal(t, err.ActionID(), actionID)
}
