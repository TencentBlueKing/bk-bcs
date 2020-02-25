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

package metastream

//Stream json stream interface for CLI reading multiple string lines
//	and parse them into specified json objects
type Stream interface {
	//Length get object number from stream
	Length() int
	//HasNext check if stream has Next JSON data
	HasNext() bool
	//GetResourceKind return apiVersion and Kind
	GetResourceKind() (string, string, error)
	//GetResourceKey return JSON object index: namespace & name
	GetResourceKey() (string, string, error)
	//GetRawJSON return  detail raw json string
	GetRawJSON() []byte
}
