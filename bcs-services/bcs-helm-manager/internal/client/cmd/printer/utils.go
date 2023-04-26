/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package printer

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/pretty"
)

func cut(s string, length int) string {
	if len(s) <= length {
		return s
	}

	return s[:length] + "..."
}

// PrintResultInJSON print data in json format
func PrintResultInJSON(result interface{}) {
	data, _ := json.Marshal(result)
	fmt.Print(string(pretty.Color(pretty.Pretty(data), nil)))
}
