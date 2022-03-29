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

package constant

const (
	//CurrentUserAttr user header
	CurrentUserAttr = "current-user"

	// DefaultTokenLength user token default length
	// token is consisted of digital and alphabet(case sensetive)
	// we can refer to http://coolaf.com/tool/rd when testing
	DefaultTokenLength = 32
	// TokenKeyPrefix is the redis key for token
	TokenKeyPrefix = "bcs_auth:token:"
	// TokenLimits for token
	TokenLimits = 1
)
