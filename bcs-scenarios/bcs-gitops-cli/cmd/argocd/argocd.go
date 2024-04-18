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

// Package argocd defines the argocd command
package argocd

import (
	"fmt"

	"github.com/argoproj/argo-cd/v2/cmd/argocd/commands"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/options"
)

// NewArgoCmd create the argo command
func NewArgoCmd() *cobra.Command {
	op := options.GlobalOption()
	argo := commands.NewCommand()
	argo.Short = "Controls a Argo CD server"
	_ = argo.PersistentFlags().Set("grpc-web-root-path", op.ProxyPath)
	_ = argo.PersistentFlags().Set("header", "X-BCS-Client: bcs-gitops-manager")
	_ = argo.PersistentFlags().Set("header", "bkUserName: admin")

	server := options.GlobalOption().Server
	token := options.GlobalOption().Token
	if server == "" || token == "" {
		blog.Fatalf("Config file '%s' cannot miss param 'server' or 'token'", options.ConfigfilePath())
	}
	_ = argo.PersistentFlags().Set("server", server)
	_ = argo.PersistentFlags().Set("header", fmt.Sprintf("Authorization: Bearer %s", token))
	return argo
}
