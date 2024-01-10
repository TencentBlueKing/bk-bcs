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

package cmd

import (
	"errors"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
)

// WithRegister init and returns the register command.
func WithRegister(sd serviced.Service) Cmd {
	cmd := &servicedCmd{
		cmd: &Command{
			Name:  "register",
			Usage: "register current service to service discovery",
			Run: func(kt *kit.Kit, params map[string]interface{}) (interface{}, error) {
				if err := sd.Register(); err != nil {
					logs.Errorf("register current service failed, err: %v, rid: %s", err, kt.Rid)
					return nil, err
				}

				logs.Infof("successfully registered current service, rid: %s", kt.Rid)
				return nil, nil
			},
		},
		sd: sd,
	}

	return cmd
}

// WithDeregister init and returns the deregister command.
func WithDeregister(sd serviced.Service) Cmd {
	cmd := &servicedCmd{
		cmd: &Command{
			Name:  "deregister",
			Usage: "deregister current service from service discovery",
			Run: func(kt *kit.Kit, params map[string]interface{}) (interface{}, error) {
				if err := sd.Deregister(); err != nil {
					logs.Errorf("deregister current service failed, err: %v, rid: %s", err, kt.Rid)
					return nil, err
				}

				logs.Infof("successfully de-registered current service, rid: %s", kt.Rid)
				return nil, nil
			},
		},
		sd: sd,
	}

	return cmd
}

// WithDisableMasterSlave init and returns the disabling master-slave command.
func WithDisableMasterSlave(sd serviced.Service) Cmd {
	cmd := &servicedCmd{
		cmd: &Command{
			Name:  "disable-master-slave",
			Usage: "disable master-slave service discovery, this service is treated as slave",
			Run: func(kt *kit.Kit, params map[string]interface{}) (interface{}, error) {
				sd.DisableMasterSlave(true)

				logs.Infof("successfully disabled master-slave service discovery, rid: %s", kt.Rid)
				return nil, nil
			},
		},
		sd: sd,
	}

	return cmd
}

// WithEnableMasterSlave init and returns the enabling master-slave command.
func WithEnableMasterSlave(sd serviced.Service) Cmd {
	cmd := &servicedCmd{
		cmd: &Command{
			Name:  "enable-master-slave",
			Usage: "enable master-slave service discovery",
			Run: func(kt *kit.Kit, params map[string]interface{}) (interface{}, error) {
				sd.DisableMasterSlave(false)

				logs.Infof("successfully enabled master-slave service discovery, rid: %s", kt.Rid)
				return nil, nil
			},
		},
		sd: sd,
	}

	return cmd
}

// servicedCmd serviced related Cmd.
type servicedCmd struct {
	cmd *Command
	sd  serviced.Service
}

// GetCommand get serviced related Command.
func (c *servicedCmd) GetCommand() *Command {
	return c.cmd
}

// Validate serviced related Command.
func (c *servicedCmd) Validate() error {
	if c.sd == nil {
		return errors.New("sd is not set")
	}

	return c.cmd.Validate()
}
