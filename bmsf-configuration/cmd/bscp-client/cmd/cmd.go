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

package cmd

import (
	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/cmd/base"
	"bk-bscp/cmd/bscp-client/cmd/cancel"
	"bk-bscp/cmd/bscp-client/cmd/commit"
	"bk-bscp/cmd/bscp-client/cmd/create"
	"bk-bscp/cmd/bscp-client/cmd/delete"
	"bk-bscp/cmd/bscp-client/cmd/get"
	"bk-bscp/cmd/bscp-client/cmd/list"
	"bk-bscp/cmd/bscp-client/cmd/lock"
	"bk-bscp/cmd/bscp-client/cmd/publish"
	"bk-bscp/cmd/bscp-client/cmd/release"
	"bk-bscp/cmd/bscp-client/cmd/reload"
	"bk-bscp/cmd/bscp-client/cmd/rollback"
	"bk-bscp/cmd/bscp-client/cmd/update"
)

var (
	// commandList for client sub commands
	commandList = []*cobra.Command{}
)

// init all subcommands
func init() {
	commandList = append(commandList, create.InitCommands()...)
	commandList = append(commandList, delete.InitCommands()...)
	commandList = append(commandList, update.InitCommands()...)
	commandList = append(commandList, get.InitCommands()...)
	commandList = append(commandList, list.InitCommands()...)
	commandList = append(commandList, lock.InitCommands()...)
	commandList = append(commandList, publish.InitCommands()...)
	commandList = append(commandList, release.InitCommands()...)
	commandList = append(commandList, commit.InitCommands()...)
	commandList = append(commandList, cancel.InitCommands()...)
	commandList = append(commandList, rollback.InitCommands()...)
	commandList = append(commandList, reload.InitCommands()...)
	commandList = append(commandList, base.InitCommands()...)
}

// GetCommandList interface for client to get all register sub commands
func GetCommandList() []*cobra.Command {
	return commandList
}
