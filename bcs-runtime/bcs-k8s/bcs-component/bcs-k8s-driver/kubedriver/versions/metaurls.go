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

package versions

// all apiVersion supported list
var apiVersionMap = map[string][]string{
	"1.5":   apiSetV15,
	"1.6":   apiSetV16,
	"1.7":   apiSetV17,
	"1.8":   apiSetV18,
	"1.11":  apiSetV111,
	"1.12":  apiSetV112,
	"1.12+": apiSetV112,
	"1.13":  apiSetV112,
	"1.13+": apiSetV112,
	"1.14":  apiSetV112,
	"1.14+": apiSetV112,
	"1.15":  apiSetV112,
	"1.16":  apiSetV112,
	"1.17":  apiSetV112,
	"1.18":  apiSetV112,
	"1.19":  apiSetV112,
	"1.20":  apiSetV112,
}
