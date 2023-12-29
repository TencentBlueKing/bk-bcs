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

// Package cmd NOTES
package cmd

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// Command is the control tools command definition.
type Command struct {
	// Name of this command, used to call the http handler of the command.
	Name string
	// Usage of this command, used in helper to inform the user how to use the command.
	Usage string
	// Parameters List of all needed parameters of the command, children will inherit these parameters.
	Parameters []Parameter
	// FromURL defines whether to read the parameters from url or from body, parameters should <= 5 if read from url.
	FromURL bool
	// IsHidden defines whether to show this command to the caller. if is true, caller only call this command, but not
	// get this command from help info.
	IsHidden bool
	// Run the command handler function.
	Run func(kt *kit.Kit, params map[string]interface{}) (interface{}, error)

	// commands is the mapping of child commands' name and command, **reserved for later use**.
	commands map[string]*Command
}

// Validate Command
func (c *Command) Validate() error {
	if c == nil {
		return errors.New("command is nil")
	}

	if len(c.Name) == 0 {
		return errors.New("command name is not set")
	}

	if len(c.Usage) == 0 {
		return errors.New("command usage is not set")
	}

	if c.Run == nil {
		return errors.New("command run function is nil")
	}

	if c.FromURL && len(c.Parameters) > 5 {
		return errors.New("parameters from url should be less than 5")
	}

	for _, param := range c.Parameters {
		if err := param.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Parameter is the control tools command parameter definition.
type Parameter struct {
	// Name of this parameter, used to pass to the command handler.
	Name string
	// Usage of this parameter, used in helper to inform the user how to use the parameter.
	Usage string
	// Default value of this parameter, this value is passed to the command handler if not set.
	Default interface{}
	// Value the **not nil pointer** with the value type of this parameter to decode param value into it.
	Value interface{}
}

// Validate Parameter
func (p *Parameter) Validate() error {
	if len(p.Name) == 0 {
		return errors.New("parameter name is not set")
	}

	if len(p.Usage) == 0 {
		return errors.New("parameter usage is not set")
	}

	if p.Value == nil {
		return errors.New("parameter value is not set")
	}

	if reflect.ValueOf(p.Value).Kind() != reflect.Ptr {
		return fmt.Errorf("parameter value is not a pointer")
	}

	return nil
}

// Cmd is an interface that represents a control tools command.
type Cmd interface {
	// GetCommand get Command.
	GetCommand() *Command
	// Validate Cmd.
	Validate() error
}

// defaultCmd default cmd implement with only one Command and its validation.
type defaultCmd struct {
	cmd *Command
}

// GetCommand get default Command.
func (c *defaultCmd) GetCommand() *Command {
	return c.cmd
}

// Validate default Command.
func (c *defaultCmd) Validate() error {
	return c.cmd.Validate()
}
