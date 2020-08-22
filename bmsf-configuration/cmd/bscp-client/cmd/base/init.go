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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
)

var initCmd *cobra.Command

// init all resource create sub command.
func initRepoCmd() *cobra.Command {
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize the application configuration file repository",
		Long:  "Initialize the application configuration file repository and set some default parameters of the repository",
		Example: `
	bk-bscp-client init --business X-Game --app gameserver --operator guohu --token 123456
		`,
		RunE: handInitDir,
	}
	initCmd.Flags().StringP("app", "a", "", "Application Name to operate. Get parameter priority: command -> env -> .bscp/desc")
	return initCmd
}

func handInitDir(cmd *cobra.Command, args []string) error {
	busName, _ := cmd.Flags().GetString("business")
	appName, _ := cmd.Flags().GetString("app")
	opeName, _ := cmd.Flags().GetString("operator")
	token, _ := cmd.Flags().GetString("token")
	appDirDetailYaml, _ := yaml.Marshal(option.CurrentDirConf{
		App:      appName,
		Business: busName,
		Operator: opeName,
		Token:    token,
	})
	os.Mkdir("./.bscp", os.ModePerm)
	ioutil.WriteFile("./.bscp/desc", appDirDetailYaml, 0644)
	configMap := make(map[string]service.ConfigFile)
	configJson, _ := json.Marshal(configMap)
	configBase64 := base64.StdEncoding.EncodeToString(configJson)
	ioutil.WriteFile("./.bscp/record", []byte(configBase64), 0644)
	initDirPath, _ := os.Getwd()
	fmt.Printf("Initialize empty bscp operation directory successfully in %s/.bscp/\n", initDirPath)
	return nil
}
