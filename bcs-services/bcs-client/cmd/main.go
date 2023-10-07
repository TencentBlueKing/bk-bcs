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

// Package main xxx
package main

import (
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/urfave/cli"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/add"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/agent"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/application"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/available"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/batch"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/create"
	deletion "github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/delete"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/deployment"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/env"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/exec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/get"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/inspect"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/list"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/offer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/permission"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/refresh"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/template"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/transaction"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/update"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
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
		refresh.NewRefreshCommand(),
		add.NewAddCommand(),
		exec.NewExecCommand(),
		permission.NewPermissionCommand(),
		transaction.NewTransactionCommand(),
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
