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

package refresh

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/urfave/cli"
)

//NewRefreshCommand sub command refresh registration
func NewRefreshCommand() cli.Command {
	return cli.Command{
		Name:  "refresh",
		Usage: "refresh usertoken",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "type, t",
				Usage: "Refresh type, value can be usertoken",
			},
			cli.StringFlag{
				Name:  "usertype",
				Usage: "user type, value can be saas/plain",
			},
			cli.StringFlag{
				Name:  "username",
				Usage: "user name",
			},
		},
		Action: func(c *cli.Context) error {
			return refresh(utils.NewClientContext(c))
		},
	}
}

func refresh(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionType); err != nil {
		return err
	}

	resourceType := c.String(utils.OptionType)

	switch resourceType {
	case "usertoken":
		return refreshUsertoken(c)
	default:
		return fmt.Errorf("invalid type: %s", resourceType)
	}
}
