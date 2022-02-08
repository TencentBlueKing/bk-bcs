/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package utils

// Barrier for limited go-routines
type Barrier struct {
	c chan struct{}
}

// NewBarrier creates new object and inits it
func NewBarrier(num int) *Barrier {
	b := &Barrier{}
	b.c = make(chan struct{}, num)
	for i := 0; i < num; i++ {
		b.c <- struct{}{}
	}
	return b
}

// Advance 1 step if there still is a unused go-routine
func (b *Barrier) Advance() {
	if b == nil {
		return
	}
	<-b.c
}

// Done means outside will release the go routine
func (b *Barrier) Done() {
	if b == nil {
		return
	}
	select {
	case b.c <- struct{}{}:
	default:
	}
}
