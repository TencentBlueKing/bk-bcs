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

// main ...
package main

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
)

const (
	initialized     = "Initialized"
	secretShares    = 5 // Number of key shares to split the generated root key into
	secretThreshold = 3 // Number of key shares required to reconstruct the root key
)

// #nosec G101
// NOCC:tosa/indent(ignore)
var secretTmpl = `global:
  vault:
    unsealKeys:
    {{- range .KeysB64 }}
      - {{ . }}
    {{- end }}
    rootToken: {{ .RootToken }}`

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "vault init operator",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(runVaultInitCmd())
		},
	}

	return cmd
}

// runVaultInitCmd 初始化
// 如果已经初始化, 返回 Initialized
// 如果未初始化, 返回结构化的 secret
// 其他返回错误信息
func runVaultInitCmd() string {
	c, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err.Error()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ok, err := c.Sys().InitStatusWithContext(ctx)
	if err != nil {
		return err.Error()
	}

	if ok {
		return initialized
	}

	resp, err := c.Sys().InitWithContext(ctx, &api.InitRequest{
		SecretShares:    secretShares,
		SecretThreshold: secretThreshold,
	})
	if err != nil {
		return err.Error()
	}

	tmpl, err := template.New("").Parse(secretTmpl)
	if err != nil {
		return err.Error()
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, resp); err != nil {
		return err.Error()
	}

	return tpl.String()
}
