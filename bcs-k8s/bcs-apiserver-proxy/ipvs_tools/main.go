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
	"flag"
	"fmt"
	"log"
)

type sliceString []string

// String xxx
func (f *sliceString) String() string {
	return fmt.Sprintf("%v", []string(*f))
}

// Set xxx
func (f *sliceString) Set(value string) error {
	*f = append(*f, value)
	return nil
}

type options struct {
	command       string
	virtualServer string
	realServer    sliceString
}

var opts options

func main() {
	care, err := NewLvsCare(opts)
	if err != nil {
		log.Printf("NewLvsCare failed: %v", err)
		return
	}

	switch care.GetLvsCommand() {
	case Add:
		err := care.CreateVirtualService()
		if err != nil {
			log.Printf("lvs[%s] add real servers %v failed: %v", opts.virtualServer, opts.realServer, err)
			return
		}

		log.Printf("lvs[%s] add real servers %v successful", opts.virtualServer, opts.realServer)
		return
	case Delete:
		err := care.DeleteVirtualService()
		if err != nil {
			log.Printf("lvs[%s] delete failed: %v", opts.virtualServer, err)
			return
		}

		log.Printf("lvs[%s] delete successful", opts.virtualServer)
		return
	default:
		log.Printf("invalid operation command, please input add or delete")
	}

	return
}

func init() {
	flag.StringVar(&opts.command, "cmd", "", "virtual server add or delete")
	flag.StringVar(&opts.virtualServer, "vs", "127.0.0.1:6443", "virtual server")
	flag.Var(&opts.realServer, "rs", "virtual server backend real server, for example: "+
		"-rs=127.0.0.1:6443 -rs=127.0.0.2:6443")

	flag.Parse()
}
