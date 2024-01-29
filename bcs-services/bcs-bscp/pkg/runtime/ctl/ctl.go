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

// Package ctl NOTES
package ctl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/ctl/cmd"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
)

// Ctl control tools structure
type Ctl struct {
	commands map[string]*cmd.Command
}

var httpHandler http.Handler

// Handler returns the http handler of control tools.
func Handler() http.Handler {
	return httpHandler
}

// LoadCtl load control tools with http handler.
func LoadCtl(commands ...cmd.Cmd) error {
	if len(commands) == 0 {
		return errors.New("control tools commands are not set")
	}

	// set up control tools supported commands.
	ctl := &Ctl{
		commands: make(map[string]*cmd.Command),
	}

	for _, command := range commands {
		if command == nil {
			return errors.New("control tools command is nil")
		}

		if err := command.Validate(); err != nil {
			return err
		}

		c := command.GetCommand()
		ctl.commands[c.Name] = c
	}

	// set up control tools http handler
	httpHandler = http.HandlerFunc(ctl.httpHandler)

	return nil
}

// WithBasics init and returns the basic commands(register & deregister & log) that all servers needed.
func WithBasics(sd serviced.Service) []cmd.Cmd {
	return []cmd.Cmd{cmd.WithLog(), cmd.WithRegister(sd), cmd.WithDeregister(sd), cmd.WithEnableMasterSlave(sd),
		cmd.WithDisableMasterSlave(sd)}
}

func (b *Ctl) httpHandler(w http.ResponseWriter, req *http.Request) {

	// get the command to run from url parameters, returns help message if not set or is 'help'
	urlParams := req.URL.Query()
	cmdName := urlParams.Get("cmd")
	if len(cmdName) == 0 || cmdName == "help" {
		rest.WriteResp(w, &rest.Response{Code: errf.OK, Data: cmd.CtlHelp(b.commands)})
		return
	}

	command, exists := b.commands[cmdName]
	if !exists {
		rest.WriteResp(w, &rest.Response{Code: errf.InvalidParameter, Data: cmd.CtlHelp(b.commands),
			Message: "'cmd' not set or not supported, please check your command, or refer to help command!"})
		return
	}

	if urlParams.Get("sub-cmd") == "help" {
		rest.WriteResp(w, &rest.Response{Code: errf.OK, Data: command.Help()})
		return
	}

	// parse params into map[name]rawValue, get from url if command has set FromURL, otherwise, get from body.
	rawParams := make(map[string]json.RawMessage)
	if command.FromURL {
		for key, value := range urlParams {
			if len(value) == 0 {
				continue
			}
			rawParams[key] = json.RawMessage(value[0])
		}
	} else {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			rest.WriteResp(w, rest.NewBaseResp(errf.InvalidParameter, "read body failed, err: "+err.Error()))
			return
		}

		if len(body) > 0 {
			if err := json.Unmarshal(body, &rawParams); err != nil {
				rest.WriteResp(w, rest.NewBaseResp(errf.InvalidParameter, "parse body failed, err: "+err.Error()))
				return
			}
		}
	}

	// parse raw params into map[paramName]paramValue for command parameters.
	params := make(map[string]interface{})
	for _, param := range command.Parameters {
		rawValue, exists := rawParams[param.Name]
		if !exists {
			if param.Default != nil {
				params[param.Name] = param.Default
			}
			continue
		}

		value := param.Value
		err := json.Unmarshal(rawValue, value)
		if err != nil {
			rest.WriteResp(w, rest.NewBaseResp(errf.InvalidParameter,
				fmt.Sprintf("parse param %s %s failed, err: %v", param.Name, rawValue, err)))
			return
		}

		params[param.Name] = value
	}

	// run command
	kt := kit.New()
	w.Header().Set(constant.RidKey, kt.Rid)

	data, err := command.Run(kt, params)
	if err != nil {
		parsed := errf.Error(err)
		rest.WriteResp(w, rest.NewBaseResp(parsed.Code, parsed.Message))
		return
	}

	rest.WriteResp(w, &rest.Response{Code: errf.OK, Data: data})
}
