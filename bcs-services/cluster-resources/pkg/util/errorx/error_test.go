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

package errorx_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

func TestNewError(t *testing.T) {
	err := errorx.New(errcode.General, "this is err msg: %s", "some error")
	assert.Equal(t, errcode.General, err.(*errorx.BaseError).Code())
	assert.Equal(t, "this is err msg: some error", err.(*errorx.BaseError).Error())

	err = errorx.New(errcode.NoPerm, "this is err msg")
	assert.Equal(t, errcode.NoPerm, err.(*errorx.BaseError).Code())
}
