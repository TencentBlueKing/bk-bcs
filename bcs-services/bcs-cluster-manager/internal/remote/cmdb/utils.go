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

package cmdb

import (
	"github.com/kirito41dd/xslice"
)

func splitCountToPage(counts int, pageLimit int) []Page {
	var pages = make([]Page, 0)

	cntSlice := make([]int, 0)
	for i := 0; i < counts; i++ {
		cntSlice = append(cntSlice, i)
	}
	i := xslice.SplitToChunks(cntSlice, pageLimit)
	ss, ok := i.([][]int)
	if !ok {
		return nil
	}

	for _, s := range ss {
		if len(s) > 0 {
			pages = append(pages, Page{
				Start: s[0],
				Limit: pageLimit,
			})
		}
	}

	return pages
}
