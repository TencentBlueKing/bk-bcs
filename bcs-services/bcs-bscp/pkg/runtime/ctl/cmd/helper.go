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
	"encoding/json"
	"fmt"
)

// Help returns the help of the command if user set sub-cmd=help.
func (c *Command) Help() *Help {
	h := &Help{
		Usage: c.Usage,
	}

	h.AvailableCommands = map[string]string{
		"help": "returns the help of the command if url parameter 'sub-cmd=help' is set",
	}
	if len(c.commands) > 0 {
		for _, cmd := range c.commands {
			if cmd.IsHidden {
				continue
			}

			h.AvailableCommands[cmd.Name] = cmd.Usage
		}
	}

	if len(c.Parameters) == 0 {
		h.Example = `curl -XPOST http://127.0.0.1/ctl?cmd=[command]`
		return h
	}

	h.Parameters = make(map[string]string)
	for _, param := range c.Parameters {
		val, _ := json.Marshal(param.Value)
		h.Parameters[param.Name] = fmt.Sprintf("%s, value example: %s", param.Usage, val)

		if param.Default != nil {
			def, _ := json.Marshal(param.Default)
			h.Parameters[param.Name] += fmt.Sprintf(", default: %s", def)
		}
	}

	if c.FromURL {
		h.Example = `curl -XPOST http://127.0.0.1/ctl?cmd=[command][&parameter=value]'`
	} else {
		h.Example = `curl -XPOST http://127.0.0.1/ctl?cmd=[command] -d '{[parameter:value]}'`
	}

	return h
}

// CtlHelp returns the help of control tool if user set no cmd or cmd=help.
func CtlHelp(commands map[string]*Command) *Help {
	rootCommand := &Command{
		Usage:    "tools to manage service status in runtime",
		commands: commands,
	}

	return rootCommand.Help()
}

// Help cmd help info.
type Help struct {
	Usage             string            `json:"usage"`
	Example           string            `json:"example"`
	AvailableCommands map[string]string `json:"available_commands,omitempty"`
	Parameters        map[string]string `json:"parameters,omitempty"`
}
