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

package page

// Pagination 分页信息
type Pagination struct {
	Sort   map[string]int // {"createTime": -1}
	Offset int64          // 偏移
	Limit  int64          // 每页的数量
	All    bool           // 是否获取全量数据, 如果同时设置了 Limit 和 All, 则以 All 为准，拉取全量数据
}

// DefaultProjectLimit 默认项目数量
const DefaultProjectLimit = 20
