/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package printer

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/pretty"
	"os"

	nodegroupmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/proto"
	"github.com/olekukonko/tablewriter"
)

// PrintStrategyInTable print strategy in table by default
func PrintStrategyInTable(wide bool, strategies []*nodegroupmanager.NodeGroupStrategy) {
	if strategies == nil {
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(func() []string {
		r := []string{"NAME", "RESOURCE_POOL", "LABELS"}
		if wide {
			r = append(r)
		}
		return r
	}())
	for _, strategy := range strategies {
		table.Append(func() []string {
			label, _ := json.Marshal(strategy.Labels)
			r := []string{
				strategy.Name, strategy.ResourcePool, string(label),
			}
			if wide {
				r = append(r)
			}
			return r
		}())
	}
	table.Render()
}

// PrintStrategyInJSON print strategy in json, -o json
func PrintStrategyInJSON(strategyList []*nodegroupmanager.NodeGroupStrategy) {
	if strategyList == nil {
		return
	}
	for _, strategy := range strategyList {
		var data []byte
		_ = encodeJSONWithIndent(4, strategy, &data)
		fmt.Println(string(pretty.Color(pretty.Pretty(data), nil)))
	}
}
