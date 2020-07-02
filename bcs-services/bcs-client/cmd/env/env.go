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

package env

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"

	"github.com/urfave/cli"
)

func NewEnvCommand() cli.Command {
	return cli.Command{
		Name:  "env",
		Usage: "Show environmental variables",
		Flags: []cli.Flag{},
		Action: func(c *cli.Context) error {
			if err := env(c); err != nil {
				return err
			}
			return nil
		},
	}
}

func env(_ *cli.Context) error {
	utils.ShowEnv()
	return nil
}
