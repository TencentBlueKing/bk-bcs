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

package base

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/option"
)

var configCmd *cobra.Command

// init all resource create sub command.
func modifyConfigCmd() *cobra.Command {
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Set or Modify the default parameter of the configuration file repository",
		Long:  "Set or Modify the default parameter of the configuration file repository",
		Example: `
	bk-bscp-client config --local app/business/operator/token "gameServ"
		`,
		RunE: handModifyConfig,
	}
	configCmd.Flags().StringP("local", "", "", "use repository config file")
	return configCmd
}

func handModifyConfig(cmd *cobra.Command, args []string) error {
	if len(os.Args) != 5 {
		return fmt.Errorf("the number of input parameters does not meet the requirements")
	}
	name := os.Args[3]
	value := os.Args[4]
	localConfig, err := option.GetInitConfInfo()
	if err != nil {
		return err
	}
	switch name {
	case "app":
		localConfig.App = value
	case "business":
		localConfig.Business = value
	case "operator":
		localConfig.Operator = value
	case "token":
		localConfig.Token = value
	default:
		return fmt.Errorf("the input parameters do not meet the requirements")
	}
	appDirDetailYaml, _ := yaml.Marshal(localConfig)
	ioutil.WriteFile("./.bscp/desc", appDirDetailYaml, 0644)
	fmt.Printf("Set the default parameter successfully! %s: %s\n", name, value)
	return nil
}
