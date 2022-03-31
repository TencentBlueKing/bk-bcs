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

package copier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Src struct {
	A string
	B int
	C bool
}

type Dst struct {
	A string
	B int
	C bool
}

func TestCopyStruct(t *testing.T) {
	var d Dst
	s := Src{A: "a", B: 1}
	CopyStruct(&d, &s)
	assert.Equal(t, d.A, s.A)
	assert.Equal(t, d.B, s.B)
	assert.Equal(t, d.C, s.C)
}
