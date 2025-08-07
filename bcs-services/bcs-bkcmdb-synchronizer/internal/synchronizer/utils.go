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

package synchronizer

import (
	"strconv"
	"strings"
)

// ClusterList the cluster list
type ClusterList []string

// Len is a method that returns the length of the ClusterList.
func (s ClusterList) Len() int {
	return len(s)
}

// Swap is a method that swaps two elements in the ClusterList.
func (s ClusterList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less is a method that compares two elements in the ClusterList and
// returns true if the element at index i is less than the element at index j.
func (s ClusterList) Less(i, j int) bool {
	// Split the elements at index i and j by "-".
	is := strings.Split(s[i], "-")
	js := strings.Split(s[j], "-")

	// Convert the last part of the split elements to integers.
	idi, _ := strconv.Atoi(is[len(is)-1])
	idj, _ := strconv.Atoi(js[len(js)-1])

	// Return true if the element at index i is less than the element at index j.
	return idi < idj
}
