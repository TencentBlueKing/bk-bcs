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

package tspider

import "fmt"

const (
	// DtEventTimeKey human readable time key which created by bkbase
	DtEventTimeKey = "dtEventTime"
	// DtEventTimeStampKey time key with index which created by bkbase
	DtEventTimeStampKey = "dtEventTimeStamp"

	// AscendingFlag descending flag for query datas
	AscendingFlag = "ASC"
	// DescendingFlag descending flag for query datas
	DescendingFlag = "DESC"

	// SqlSelectAll SQL select all columns
	SqlSelectAll = "*"
	// SqlSelectCount SQL count all records
	SqlSelectCount = "COUNT(*)"
)

func ensureSortAscending(v int64) (string, error) {
	switch v {
	case 1:
		return AscendingFlag, nil
	case -1:
		return DescendingFlag, nil
	default:
		return "", fmt.Errorf("sort params must be 1 or -1")
	}
}
