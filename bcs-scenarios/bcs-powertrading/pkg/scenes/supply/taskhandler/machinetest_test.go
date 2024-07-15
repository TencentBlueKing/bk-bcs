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
 */

package taskhandler

import (
	"fmt"
	"testing"
)

func Test_getDefaultDateRange(t *testing.T) {
	fmt.Println(getDefaultDateRange())
}

func Test_11(t *testing.T) {
	// fmt.Println(6 / 10)
	//list := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	//for i := 0; i <= len(list)/10; i++ {
	//	test := make([]int, 0)
	//	fmt.Println(i)
	//	if i != len(list)/10 {
	//		test = list[i*10 : (i+1)*10]
	//		fmt.Println(test)
	//	} else {
	//		test = list[i*10:]
	//		fmt.Println(test)
	//	}
	//}
	sum := 261
	batch := sum / 20
	infos := make([]int, 261)
	fmt.Println(batch)
	for i := 0; i <= batch; i++ {
		if i != batch {
			fmt.Printf("start:%d, end:%d\n", i*20, i*20+19)
			choose := infos[i*20 : (i*20 + 20)]
			fmt.Println(len(choose))
		} else {
			fmt.Printf("start:%d, end:%d\n", i*20, sum)
			choose := infos[i*20:]
			fmt.Println(len(choose))
		}
	}
}
