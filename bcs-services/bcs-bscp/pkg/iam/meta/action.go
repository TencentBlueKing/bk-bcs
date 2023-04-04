/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package meta

// Action 表示 bscp 这一侧的资源类型， 对应的有 client.ActionID 表示 iam 一侧的资源类型
// 两者之间有映射关系，详情见 AdaptAuthOptions
type Action string

// String convert Action to string.
func (a Action) String() string {
	return string(a)
}

const (
	// FindBusinessResource
	FindBusinessResource Action = "find_business_resource"
	// Create operation's bscp auth action type
	Create Action = "create"
	// Update operation's bscp auth action type
	Update Action = "update"
	// Delete operation's bscp auth action type
	Delete Action = "delete"
	// Find operation's bscp auth action type
	Find Action = "find"
	// Publish operation's bscp auth action type
	Publish Action = "publish"
	// FinishPublish operation's bscp auth action type
	FinishPublish Action = "finish_publish"
	// Upload operation's bscp auth action type
	Upload Action = "upload"
	// Download operation's bscp auth action type
	Download Action = "download"
	// SkipAction means the operation do not need to do authentication, skip auth
	SkipAction Action = "skip"
	// Access means sidecar access the feed server action. and only for this scenario.
	Access Action = "access"
)
