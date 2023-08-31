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

package server

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	vpcommon "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/vaultplugin-server/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/vaultplugin-server/secret"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
)

const (
	gracefulexit = 5
)

// Options vaultplugin-server options
type Options struct {
	conf.LogConfig
	common.ServerConfig
	Secret secret.Options `json:"secret,omitempty"`
	GitOps GitOpsOption
}

// GitOpsOption were used by the vault and gitops system interaction
type GitOpsOption struct {
	Service        string `json:"service"`
	User           string `json:"user,omitempty"`
	Pass           string `json:"password,omitempty"`
	AdminNamespace string `json:"adminNamespace"`
}

// DefaultOptions vaultplugin-server default options
func DefaultOptions() *Options {
	return &Options{
		LogConfig: conf.LogConfig{
			LogDir:       "/data/bcs/logs/bcs",
			Verbosity:    3,
			AlsoToStdErr: true,
		},
		ServerConfig: common.ServerConfig{
			Address:  "0.0.0.0",
			HTTPPort: 8080,
		},
		Secret: secret.Options{
			Type:      vpcommon.GetSecretType(),
			Endpoints: vpcommon.GetVaultAddr(),
			Token:     vpcommon.GetVaultTokenForServer(),
		},
		GitOps: GitOpsOption{
			Service:        vpcommon.GetGitopsService(),
			AdminNamespace: vpcommon.GetGitopsAdminNamespace(),
			User:           vpcommon.GetGitopsUser(),
			Pass:           vpcommon.GetGitopsPassword(),
		},
	}
}

// Validate server options
func (o *Options) Validate() error {
	if o.Secret.Token == "" || o.Secret.Endpoints == "" {
		return fmt.Errorf("lost secret service token or endpoint")
	}
	if o.GitOps.Service == "" ||
		o.GitOps.AdminNamespace == "" ||
		o.GitOps.User == "" ||
		o.GitOps.Pass == "" {
		return fmt.Errorf("lost gitops service options")
	}
	return nil
}
