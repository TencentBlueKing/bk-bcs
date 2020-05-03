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

package option

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	//GlobalOptions setting global options for all sub command
	GlobalOptions *Global
)

func init() {
	GlobalOptions = &Global{
		ConfigFile: "/etc/bscp/client.yaml",
		Business:   "",
		Operator:   "",
		Index:      0,
		Limit:      100,
	}
}

//Global all options shared in all sub commands
type Global struct {
	//ConfigFile for client connect to platform
	ConfigFile string
	//Business name for operation
	Business string
	//Operator user name for operation
	Operator string
	//Index for list
	Index int32
	//Limit for list
	Limit int32
}

//ParseGlobalOption parse global option for all commands
func ParseGlobalOption(cmd *cobra.Command, args []string) {
	//business enviroment
	business := os.Getenv("BSCP_BUSINESS")
	if len(business) != 0 {
		cmd.Flags().Set("business", business)
	}
	ops := os.Getenv("BSCP_OPERATOR")
	if len(ops) != 0 {
		cmd.Flags().Set("operator", ops)
	}
}
