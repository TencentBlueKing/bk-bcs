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

package get

import (
	"context"
	"fmt"
	"path"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"bk-bscp/internal/protocol/common"
)

func getConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configset",
		Aliases: []string{"cfgset"},
		Short:   "Get configSet details",
		Long:    "Get configSet detail information for application.",
		Example: `
	bscp-client get configset --path cfgsetPath --name cfgsetName
	bscp-client get configset --id xxxxxxx
		`,
		RunE: handleGetConfigSet,
	}
	// command line flags.
	cmd.Flags().StringP("app", "a", "", "the application name of configset")
	cmd.Flags().StringP("name", "n", "", "the name of configset")
	cmd.Flags().StringP("path", "p", "", "the path of configset")
	cmd.Flags().StringP("id", "i", "", "the id of configset")
	return cmd
}

//getCommitCmd: client get commit
func getMultiCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "commit",
		Aliases: []string{"ci"},
		Short:   "Get commit details",
		Long:    "Get commit detail information",
		//RunE: handleGetMultiCommit,
		RunE: handleGetCommitInfo,
	}
	// --Id is required
	cmd.Flags().StringP("id", "", "", "the id of commit")
	cmd.Flags().StringP("mid", "", "", "the id of commit module")
	return cmd
}

func handleGetCommitInfo(cmd *cobra.Command, args []string) error {
	multiCommitID, _ := cmd.Flags().GetString("id")
	commitID, _ := cmd.Flags().GetString("mid")
	if len(multiCommitID) == 0 && len(commitID) == 0 {
		return fmt.Errorf("id or mid is a required parameter")
	} else if len(multiCommitID) != 0 && len(commitID) != 0 {
		return fmt.Errorf("mid or id can only enter one as a parameter")
	}
	if len(multiCommitID) != 0 {
		handleGetMultiCommit(cmd, args)
	}
	if len(commitID) != 0 {
		handleGetCommit(cmd, args)
	}
	return nil
}

func handleGetCommit(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//check --Id option
	commitID, _ := cmd.Flags().GetString("mid")
	//create business and check result
	commit, err := operator.GetCommit(context.TODO(), commitID)
	if err != nil {
		return err
	}
	if commit == nil {
		cmd.Printf("Found no Commit resource.\n")
		return nil
	}
	//format output
	application, err := operator.QueryApplication(context.TODO(), &common.App{Bid: commit.Bid, Appid: commit.Appid})
	cfgset, err := operator.QueryConfigSet(context.TODO(), &common.ConfigSet{Bid: commit.Bid, Appid: commit.Appid, Cfgsetid: commit.Cfgsetid})
	utils.PrintDetailsCommit(commit, operator.Business, application.Name, path.Clean(cfgset.Fpath+"/"+cfgset.Name))
	cmd.Println()
	cmd.Printf("\t(use \"bk-bscp-client release --name <releaseName> --commitid <commitid> --strategy <strategyName> --memo \"this is a example\"\" to create release to publish)\n\n")
	return nil
}

func handleGetMultiCommit(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//check --Id option
	multiCommitID, _ := cmd.Flags().GetString("id")
	//create business and check result
	multiCommit, metadata, err := operator.GetMultiCommit(context.TODO(), multiCommitID)
	if err != nil {
		return err
	}
	if multiCommit == nil {
		cmd.Printf("Found no multi-commit resource.\n")
		return nil
	}
	application, err := operator.QueryApplication(context.TODO(), &common.App{Bid: multiCommit.Bid, Appid: multiCommit.Appid})
	utils.PrintDetailsMultiCommit(multiCommit, operator.Business, application.Name)
	utils.PrintCommitMetaData(operator, metadata, multiCommit.Bid, multiCommit.Appid)
	cmd.Println()
	cmd.Printf("\t(use \"bk-bscp-client get commit --mid <moduleId>\" to get commit module detail)\n")
	cmd.Printf("\t(use \"bk-bscp-client release --name <releaseName> --commitid <commitid> --strategy <strategyName> --memo \"this is a example\"\" to create release to publish)\n\n")
	return nil
}

func handleGetConfigSet(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	// Level 3 read appName
	option.SetGlobalVarByName(cmd, "app")
	cfgsetId, _ := cmd.Flags().GetString("id")
	cfgsetName, _ := cmd.Flags().GetString("name")
	if len(cfgsetId) == 0 && len(cfgsetName) == 0 {
		return fmt.Errorf("%s", option.ErrMsg_PARAM_NUM)
	}
	var cfgset *common.ConfigSet
	var err error
	// query by id
	if len(cfgsetId) != 0 {
		request := &common.ConfigSet{
			Cfgsetid: cfgsetId,
		}
		cfgset, err = operator.GetConfigSetById(context.TODO(), request)
	} else {
		// query by path-name
		appName, _ := cmd.Flags().GetString("app")
		if len(appName) == 0 {
			return fmt.Errorf("%s %s", option.ErrMsg_PARAM_MISS, "app")
		}
		cfgsetPath, _ := cmd.Flags().GetString("path")
		request := &common.ConfigSet{
			Name:  cfgsetName,
			Fpath: cfgsetPath,
		}
		cfgset, err = operator.GetConfigSet(context.TODO(), appName, request)
	}
	if err != nil {
		return err
	}
	if cfgset == nil {
		fmt.Printf("%s\n", option.SucMsg_DATA_NO_FOUNT)
		return nil
	}
	application, err := operator.QueryApplication(context.TODO(), &common.App{Bid: cfgset.Bid, Appid: cfgset.Appid})
	if err != nil {
		return err
	}
	utils.PrintDetailsCfgset(cfgset, operator.Business, application.Name)
	return nil
}
