/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"errors"

	"bscp.io/cmd/sidecar/stream/types"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
)

// WithNotifyReconnect notify sidecar to reconnect upstream server.
func WithNotifyReconnect(nr func(signal types.ReconnectSignal)) Cmd {
	cmd := &watchCmd{
		notifyReConnect: nr,
	}

	cmd.cmd = &Command{
		Name:     "reconnect",
		Usage:    "notify sidecar to reconnect upstream server",
		IsHidden: true,
		Run: func(kt *kit.Kit, params map[string]interface{}) (interface{}, error) {
			cmd.notifyReConnect(types.ReconnectSignal{
				Reason: "call ctl tool to notify reconnect",
			})

			logs.Infof("successfully notify sidecar to reconnect upstream server, rid: %s", kt.Rid)

			return nil, nil
		},
	}

	return cmd
}

// watchCmd watch related Cmd.
type watchCmd struct {
	cmd             *Command
	notifyReConnect func(signal types.ReconnectSignal)
}

// GetCommand get disable/enable write Command.
func (c *watchCmd) GetCommand() *Command {
	return c.cmd
}

// Validate write server related Command.
func (c *watchCmd) Validate() error {
	if c.notifyReConnect == nil {
		return errors.New("notifyReConnect is not set")
	}

	return c.cmd.Validate()
}
