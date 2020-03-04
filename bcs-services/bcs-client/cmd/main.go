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

package main

import (
	"fmt"
	"os"

	"bk-bcs/bcs-common/common/version"
	"bk-bcs/bcs-services/bcs-client/cmd/agent"
	"bk-bcs/bcs-services/bcs-client/cmd/application"
	"bk-bcs/bcs-services/bcs-client/cmd/available"
	"bk-bcs/bcs-services/bcs-client/cmd/batch"
	"bk-bcs/bcs-services/bcs-client/cmd/create"
	deletion "bk-bcs/bcs-services/bcs-client/cmd/delete"
	"bk-bcs/bcs-services/bcs-client/cmd/deployment"
	"bk-bcs/bcs-services/bcs-client/cmd/env"
	"bk-bcs/bcs-services/bcs-client/cmd/get"
	"bk-bcs/bcs-services/bcs-client/cmd/inspect"
	"bk-bcs/bcs-services/bcs-client/cmd/list"
	"bk-bcs/bcs-services/bcs-client/cmd/offer"
	"bk-bcs/bcs-services/bcs-client/cmd/template"
	"bk-bcs/bcs-services/bcs-client/cmd/update"
	"bk-bcs/bcs-services/bcs-client/cmd/utils"

	"github.com/urfave/cli"
)

func main() {
	bcsCli := cli.NewApp()
	bcsCli.Name = "bcs client"
	bcsCli.Usage = "command-line client for BlueKing Container Service"
	cliVersion := fmt.Sprintf("\n%s", version.GetVersion())
	bcsCli.Version = cliVersion

	bcsCli.Commands = []cli.Command{
		create.NewCreateCommand(),
		update.NewUpdateCommand(),
		deletion.NewDeleteCommand(),
		application.NewScaleCommand(),
		application.NewRollBackCommand(),
		list.NewListCommand(),
		inspect.NewInspectCommand(),
		get.NewGetCommand(),
		deployment.NewCancelCommand(),
		deployment.NewPauseCommand(),
		deployment.NewResumeCommand(),
		application.NewRescheduleCommand(),
		env.NewExportCommand(),
		env.NewEnvCommand(),
		available.NewEnableCommand(),
		available.NewDisableCommand(),
		offer.NewOfferCommand(),
		agent.NewAgentSettingCommand(),
		template.NewTemplateCommand(),
		batch.NewApplyCommand(),
		batch.NewCleanCommand(),
	}

	if err := utils.InitCfg(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// env init failed should not exit
	_ = utils.InitEnv()

	if err := bcsCli.Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
