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

package api

import "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/suite/cases"

func genNormalNameForCreateTest() []string {
	return []string{
		// to test: Only Chinese, English, numbers, underscores, and underscores are allowed
		cases.TZhEnNumUnderHyphen,
		// to test: maximum length is 128
		cases.RandString(128),
		// to test: start and end with number
		cases.TNumber,
		// to test: start and end with English
		cases.TEnglish,
		// to test: start and end with Chinese
		cases.TChinese,
	}
}

func genNormalNameForUpdateTest() []string {
	return []string{
		// to test: Only Chinese, English, numbers, underline, hyphen and dot are allowed
		cases.TZhEnNumUnderHyphen,
		// to test: maximum length is 128
		cases.RandString(128),
		// to test: start and end with number
		cases.TNumber + "6789",
		// to test: start and end with English
		cases.TEnglish + "_name",
		// to test: start and end with Chinese
		cases.TChinese + "名字",
	}
}

func genAbnormalNameForTest() []string {
	return []string{
		// to test: Only Chinese, English, numbers, underline, and hyphen are allowed
		cases.WCharacter,
		// to test: maximum length is 128
		cases.RandString(129),
		// to test: Must start with Chinese, English, numbers
		cases.WPrefix,
		// to test: Must end with Chinese, English, numbers
		cases.WTail,
	}
}

func genNormalMemoForTest() []string {
	return []string{
		// to test: Only Chinese, English, numbers, underline, hyphen, spaces are allowed
		cases.TZhEnNumUnderHyphenSpace,
		// to test: maximum length is 256
		cases.RandString(256),
		// to test: start and end with number
		cases.TNumber,
		// to test: start and end with English
		cases.TEnglish,
		// to test: start and end with Chinese
		cases.TChinese,
	}
}

func genAbnormalMemoForTest() []string {
	return []string{
		// to test: Only Chinese, English, numbers, underline, and hyphen are allowed
		cases.WCharacter,
		// to test: maximum length is 256
		cases.RandString(257),
		// to test: Must start with Chinese, English, numbers
		cases.WPrefix,
		// to test: Must end with Chinese, English, numbers
		cases.WTail,
	}
}
