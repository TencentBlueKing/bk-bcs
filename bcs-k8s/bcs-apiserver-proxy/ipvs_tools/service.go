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

package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-apiserver-proxy/pkg/service"
)

var (
	// ErrLvsCareNotInited for lvsCare not inited
	ErrLvsCareNotInited = errors.New("LvsCare not inited")
	// ErrNotValidateOperation for invalid command
	ErrNotValidateOperation = errors.New("invalid operation command")
)

const (
	// Add command
	Add Operation = "add"
	// Delete command
	Delete Operation = "delete"
	// Invalid command
	Invalid Operation = "invalid operation"
)

// Operation for operation command
type Operation string

func (o Operation) validate() bool {
	return o == Add || o == Delete
}

func (o Operation) isAddCommand() bool {
	return o == Add
}

func (o Operation) isDeleteCommand() bool {
	return o == Delete
}

// NewLvsCare init lvsCare client
func NewLvsCare(opts options) (*LvsCare, error) {
	care := &LvsCare{
		command:       Operation(opts.command),
		virtualServer: opts.virtualServer,
		realServer:    opts.realServer,
		lvs:           service.NewLvsProxy(),
	}

	ok := care.validate()
	if !ok {
		infoMsg := fmt.Errorf("LvsCare validate failed")
		return nil, infoMsg
	}

	return care, nil
}

// LvsCare for create or delete vs
type LvsCare struct {
	command       Operation
	virtualServer string
	realServer    []string
	lvs           service.LvsProxy
}

func (lvs *LvsCare) validate() bool {
	if lvs == nil {
		return false
	}

	ok := lvs.command.validate()
	if !ok {
		log.Println("Command operation only support: add or delete virtual service operation")
		return false
	}

	if len(lvs.virtualServer) == 0 {
		log.Println("virtual server is empty")
		return false
	}

	if lvs.command.isAddCommand() {
		if len(lvs.realServer) == 0 {
			log.Println("real servers is empty")
			return false
		}
	}

	return true
}

// GetLvsCommand get operation command
func (lvs *LvsCare) GetLvsCommand() Operation {
	if lvs == nil {
		return Invalid
	}

	return lvs.command
}

// CreateVirtualService create vs
func (lvs *LvsCare) CreateVirtualService() error {
	if lvs == nil {
		return ErrLvsCareNotInited
	}

	var errs []string

	available := lvs.lvs.IsVirtualServerAvailable(lvs.virtualServer)
	if !available {
		err := lvs.lvs.CreateVirtualServer(lvs.virtualServer)
		if err != nil {
			log.Printf("CreateVirtualServer[%s] failed: %v", lvs.virtualServer, err)
			return err
		}
	}

	for _, r := range lvs.realServer {
		err := lvs.lvs.CreateRealServer(r)
		if err != nil {
			errs = append(errs, fmt.Sprintf("CreateRealServer[%s/%s] failed: %v", lvs.virtualServer, r, err))
		}
	}

	if len(errs) != 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

// DeleteVirtualService delete vs
func (lvs *LvsCare) DeleteVirtualService() error {
	if lvs == nil {
		return ErrLvsCareNotInited
	}

	exist := lvs.lvs.IsVirtualServerAvailable(lvs.virtualServer)
	if !exist {
		infoMsg := fmt.Errorf("deleteVirtualService[%s] failed: %s not exist", lvs.virtualServer, lvs.virtualServer)
		return infoMsg
	}

	err := lvs.lvs.DeleteVirtualServer(lvs.virtualServer)
	if err != nil {
		errMsg := fmt.Errorf("DeleteVirtualServer[%s] failed: %v", lvs.virtualServer, err)
		return errMsg
	}

	return nil
}
