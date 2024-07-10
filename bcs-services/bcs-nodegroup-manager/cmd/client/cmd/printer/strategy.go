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

// Package printer xxx
package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/pretty"

	nodegroupmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/proto"
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
			// no-op append call, 源码如此(append是否缺少参数,wide是否有意义)
			// nolint
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
				// no-op append call, 源码如此(append是否缺少参数,wide是否有意义)
				// nolint
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
		data, err := json.Marshal(strategy)
		if err != nil {
			fmt.Println(err.Error())
		}
		res, err := PrettyString(string(data))
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(string(pretty.Color(pretty.Pretty([]byte(res)), nil)))
		}
	}
}

// PrettyString 格式化
func PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", " "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}
