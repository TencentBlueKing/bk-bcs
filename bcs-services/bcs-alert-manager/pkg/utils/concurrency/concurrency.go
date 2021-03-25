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

package concurrency

// Concurrency for limited resources
type Concurrency struct {
	c chan struct{}
}

// NewConcurrency create object and init it
func NewConcurrency(num int) *Concurrency {
	conNum := &Concurrency{}
	conNum.c = make(chan struct{}, num)

	for i := 0; i < num; i++ {
		conNum.c <- struct{}{}
	}

	return conNum
}

// Add allocate 1 if there is 1 un-used resources
func (con *Concurrency) Add() {
	if con == nil {
		return
	}

	<-con.c
}

// Done release 1 resource to resourcePool
func (con *Concurrency) Done() {
	if con == nil {
		return
	}

	select {
	case con.c <- struct{}{}:
	default:
	}
}
