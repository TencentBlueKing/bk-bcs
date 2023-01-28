/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package cases

import "math"

// normal generic test cases. T means true, for normal test.
const (
	TBizID uint32 = math.MaxUint32
	// TZhEnNumUnderHyphenSpace to test: Only Chinese, English, number, underline, hyphen, space are allowed.
	TZhEnNumUnderHyphenSpace string = "中文_English- 12345"
	// TZhEnNumUnderHyphenDot to test: Only Chinese, English, number, underline, hyphen, point are allowed.
	TZhEnNumUnderHyphenDot string = "中文_English-12345."
	// TEnNumUnderHyphenDot to test: Only English, number, underline, hyphen, point are allowed.
	TEnNumUnderHyphenDot string = "English_12.3-4"
	// TZhEnNumUnderHyphen to test: Only Chinese, English, number, underline, hyphen are allowed.
	TZhEnNumUnderHyphen string = "中文_English-12345"
	// TNumber to test: string should start and end with number.
	TNumber string = "12345"
	// TEnglish to test: string should start and end with English.
	TEnglish string = "English"
	// TChinese to test: string should start and end with Chinese.
	TChinese string = "中文"
)

// abnormal generic test cases. W means wrong, for abnormal test.
const (
	WID uint32 = 0
	// WCharacter to test: Only Chinese, English, number, underline, hyphen, space, dot are allowed.
	WCharacter string = "¥@%&"
	// WPrefix to test: string should start with Chinese, English, number.
	WPrefix string = "_underline_prefix"
	// WTail to test: string should end with Chinese, English, number.
	WTail string = "underline_tail_"
	// WEnum If the field value is a limited range value, like file type and so on, use this test case.
	WEnum string = "wrong_enum"
)
