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

package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitMsgMap(t *testing.T) {
	assert.Nil(t, InitMsgMap())
	assert.Equal(t, i18nMsgMap["无指定操作权限"]["zh"], "无指定操作权限")
	assert.Equal(t, i18nMsgMap["无指定操作权限"]["en"], "no operate permission!")
}
