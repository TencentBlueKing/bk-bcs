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

package backend

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"strconv"
	"strings"
)

//TaskSorter bia name of []TaskGroup
type TaskSorter []*types.TaskGroup

func (s TaskSorter) Len() int      { return len(s) }
func (s TaskSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s TaskSorter) Less(i, j int) bool {

	// the time for building taskgroup
	a, _ := strconv.Atoi(strings.Split(s[i].ID, ".")[4])
	b, _ := strconv.Atoi(strings.Split(s[j].ID, ".")[4])

	return a < b
}
