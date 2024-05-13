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

package qcloud

func getMaxPrivateIPNumPerENI(cores, mem int) int {
	if cores <= 1 && mem <= 1 {
		return 2
	}
	if cores <= 1 && mem > 1 {
		return 6
	}
	if cores <= 2 {
		return 10
	}
	if cores <= 4 && mem < 16 {
		return 10
	}
	if cores <= 4 && mem >= 16 {
		return 20
	}
	if cores <= 8 {
		return 20
	}
	return 30
}

func getMaxENINumPerCVM(cores, mem int) int {
	if cores <= 1 {
		return 2
	}
	if cores <= 2 {
		return 4
	}
	if cores <= 4 {
		return 4
	}
	if cores <= 10 {
		return 6
	}
	return 8
}
