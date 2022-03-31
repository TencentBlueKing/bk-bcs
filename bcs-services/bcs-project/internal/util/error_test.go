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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestProjectError(t *testing.T) {
	// one message
	err := NewError(common.BcsProjectSuccess, common.BcsProjectSuccessMsg)
	assert.Equal(t, err.Error(), common.BcsProjectSuccessMsg)
	// some message
	err = NewError(common.BcsProjectParamErr, common.BcsProjectParamErrMsg, "some error")
	assert.Equal(t, int(err.Code()), int(common.BcsProjectParamErr))
	assert.Contains(t, err.Error(), common.BcsProjectParamErrMsg)
}
