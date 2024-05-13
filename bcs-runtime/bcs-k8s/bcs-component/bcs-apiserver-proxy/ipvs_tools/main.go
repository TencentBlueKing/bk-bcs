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

// Package main xxx
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	ipvsConfig "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy/pkg/ipvs/config"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy/pkg/utils"
)

const (
	// Init command
	Init Operation = "init"
	// Reload command
	Reload Operation = "reload"
	// Add command
	Add Operation = "add"
	// Delete command
	Delete Operation = "delete"
)

// Operation for operation command
type Operation string

func (o Operation) validate() bool { // nolint unused
	return o == Init || o == Add || o == Delete || o == Reload
}

func (o Operation) isInitCommand() bool { // nolint unused
	return o == Init
}

func (o Operation) isReloadCommand() bool { // nolint unused
	return o == Reload
}

func (o Operation) isAddCommand() bool { // nolint unused
	return o == Add
}

func (o Operation) isDeleteCommand() bool { // nolint unused
	return o == Delete
}

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
	command        string
	virtualServer  string
	realServer     sliceString
	scheduler      string
	ipvsPersistDir string
	toolPath       string
	healthScheme   string
	healthPath     string
}

var opts options

func main() {
	operation := Operation(opts.command)
	switch operation {
	case Init:
		initFunc()
	case Reload:
		reloadFunc()
	case Add:
		addFunc()
	case Delete:
		deleteFunc()
	default:
		log.Printf("invalid operation command")
	}
}

func initFunc() {
	if !validateInitOptions(opts) {
		log.Println("validate options failed, check your options")
		return
	}
	care, err := NewLvsCareFromFlag(opts)
	if err != nil {
		log.Printf("create lvsCare failed: %v", err)
	}
	err = care.CreateVirtualService()
	if err != nil {
		log.Printf("lvs[%s] init real servers %v failed: %v", opts.virtualServer, opts.realServer, err)
		return
	}
	scheduler, err := care.lvs.GetScheduler()
	if err != nil {
		log.Printf("lvs[%s] get scheduler failed: %v", opts.virtualServer, err)
		return
	}
	vs, err := care.lvs.GetVirtualServer()
	if err != nil {
		log.Printf("lvs[%s] get virtual server failed: %v", opts.virtualServer, err)
		return
	}
	rs, err := care.lvs.ListRealServer()
	if err != nil {
		log.Println("init failed")
	}
	config := ipvsConfig.IpvsConfig{
		Scheduler:     scheduler,
		VirtualServer: vs,
		RealServer:    rs,
	}
	err = ipvsConfig.WriteIpvsConfig(opts.ipvsPersistDir, config)
	if err != nil {
		return
	}
	err = utils.SetIpvsStartup(opts.ipvsPersistDir, opts.toolPath)
	if err != nil {
		log.Println("set ipvs startup failed")
		return
	}
	log.Printf("lvs[%s] init real servers %v successful", opts.virtualServer, opts.realServer)
}

func reloadFunc() {
	care, err := NewLvsCareFromConfig(opts)
	if err != nil {
		log.Printf("reload ipvs failed: %v", err)
	}
	err = care.CreateVirtualService()
	if err != nil {
		log.Printf("lvs[%s] reload real servers %v failed: %v", opts.virtualServer, opts.realServer, err)
		return
	}
}

func addFunc() {
	care, err := NewLvsCareFromFlag(opts)
	if err != nil {
		log.Printf("create lvsCare failed: %v", err)
	}
	err = care.CreateVirtualService()
	if err != nil {
		log.Printf("lvs[%s] add real servers %v failed: %v", opts.virtualServer, opts.realServer, err)
		return
	}

	log.Printf("lvs[%s] add real servers %v successful", opts.virtualServer, opts.realServer)
}

func deleteFunc() {
	care, err := NewLvsCareFromFlag(opts)
	if err != nil {
		log.Printf("create lvsCare failed: %v", err)
	}
	err = care.DeleteVirtualService()
	if err != nil {
		log.Printf("lvs[%s] delete failed: %v", opts.virtualServer, err)
		return
	}

	log.Printf("lvs[%s] delete successful", opts.virtualServer)
}

func validateInitOptions(opt options) bool {
	tool, err := os.Stat(opt.toolPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("error path, please set the valid absolute path for bcs-apiserver-proxy-tools")
			return false
		}
		log.Println("error path, please set the valid absolute path for bcs-apiserver-proxy-tools")
	}
	if tool.IsDir() {
		log.Println("error path, please set the valid absolute path for bcs-apiserver-proxy-tools")
		return false
	}
	return true
}

func init() {
	flag.StringVar(&opts.command, "cmd", "", "one of init|reload|add|delete")
	flag.StringVar(&opts.virtualServer, "vs", "127.0.0.1:6443", "virtual server")
	flag.StringVar(&opts.scheduler, "scheduler", "sh", "lvs scheduler, one of rr|wrr|lc|wlc|lblc|lblcr|dh|sh|sed|nq")
	flag.Var(&opts.realServer, "rs", "virtual server backend real server, for example: "+
		"-rs=127.0.0.1:6443 -rs=127.0.0.2:6443")
	flag.StringVar(&opts.ipvsPersistDir, "persistDir", "/root/.bcs", "persistent ipvs rules path")
	flag.StringVar(&opts.toolPath, "toolPath", "/root/bcs-apiserver-proxy-tools",
		"absolute path for bcs-apiserver-proxy-tools")
	flag.StringVar(&opts.healthScheme, "healthScheme", "https", "scheme for health check")
	flag.StringVar(&opts.healthPath, "healthPath", "/healthz", "path for health check")

	flag.Parse()
}
