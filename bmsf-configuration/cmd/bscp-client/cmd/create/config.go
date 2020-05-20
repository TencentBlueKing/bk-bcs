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

package create

import (
	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"bk-bscp/internal/protocol/accessserver"
	"context"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

// createConfigSetCmd: client create configset.
func createConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configset",
		Aliases: []string{"cfgset"},
		Short:   "create configset",
		Long:    "create ConfigSet for application",
		Example: `
	bscp-client create configset --business somegame --app gamesvr --name cfgName --operator nobody
	bscp-client create cfgset --business somegame -a gamesvr -n cfgName --operator nobody
		`,
		RunE: handleCreateConfigSet,
	}
	// command line flags.
	cmd.Flags().StringP("name", "n", "", "settings new ConfigSet name")
	cmd.Flags().StringP("app", "a", "", "settings app that ConfigSet belongs to.")
	cmd.Flags().StringP("path", "p", "", "settings sub file path of ConfigSet.")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("app")
	return cmd
}

//createCommitCmd: client create commit
func createCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "commit",
		Aliases: []string{"ci"},
		Short:   "create commit",
		Long:    "create commit for application",
		Example: `
	bscp-client create commit --app gamesvr --cfgset cfgName --config-file ./somefile.json
	bscp-client create ci -a gamesvr -c cfgName -f ./somefile.json
		`,
		RunE: handleCreateCommit,
	}
	//command line flags
	cmd.Flags().StringP("app", "a", "", "settings application name that commit belongs to")
	cmd.Flags().StringP("cfgset", "c", "", "settings ConfigSet that commit belongs to")
	cmd.Flags().StringP("config-file", "f", "", "configuration detail information that this commit works for.")
	cmd.Flags().StringP("template", "t", "", "configuration template that this commit use for generating configuration.")
	cmd.MarkFlagRequired("app")
	cmd.MarkFlagRequired("cfgset")
	return cmd
}

func handleCreateCommit(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}
	templateFile, err := cmd.Flags().GetString("template")
	if err != nil {
		return err
	}
	if len(configFile) == 0 && len(templateFile) == 0 {
		return fmt.Errorf("--config-file or --template is required")
	}
	//reading all details from config-file
	var cfgContent []byte
	if len(configFile) != 0 {
		cfgBytes, cfgErr := ioutil.ReadFile(configFile)
		if cfgErr != nil {
			return err
		}
		cfgContent = cfgBytes
	}
	appName, _ := cmd.Flags().GetString("app")
	cfgsetName, _ := cmd.Flags().GetString("cfgset")
	//construct createRequest
	request := &service.CreateCommitOption{
		AppName:       appName,
		ConfigSetName: cfgsetName,
		Content:       cfgContent,
	}
	//create Commit and check result
	commitID, err := operator.CreateCommit(context.TODO(), request)
	if err != nil {
		return err
	}
	cmd.Printf("Create Commit successfully: %s\n", commitID)
	return nil
}

func handleCreateConfigSet(cmd *cobra.Command, args []string) error {
	// get global command info and create business operator.
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}

	// check all flags.
	appName, err := cmd.Flags().GetString("app")
	if err != nil {
		return err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	fpath, err := cmd.Flags().GetString("path")
	if err != nil {
		return err
	}
	business, app, err := utils.GetBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return err
	}

	// construct createRequest.
	request := &accessserver.CreateConfigSetReq{
		Bid:     business.Bid,
		Appid:   app.Appid,
		Name:    name,
		Fpath:   fpath,
		Creator: operator.User,
	}

	// create configset and check result.
	cfgSetID, err := operator.CreateConfigSet(context.TODO(), request)
	if err != nil {
		return err
	}

	cmd.Printf("Create ConfigSet successfully: %s\n", cfgSetID)
	return nil
}
