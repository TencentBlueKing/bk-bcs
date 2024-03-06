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

package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy/pkg/health"
	ipvsConfig "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy/pkg/ipvs/config"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy/pkg/service"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy/pkg/utils"
)

var (
	// ErrLvsCareNotInited for lvsCare not inited
	ErrLvsCareNotInited = errors.New("LvsCare not inited")
)

// Scheduler xxx
type Scheduler string

func (s Scheduler) validate() bool {
	return s == "rr" || s == "wrr" || s == "lc" || s == "wlc" ||
		s == "lblc" || s == "lblcr" || s == "dh" || s == "sh" ||
		s == "sed" || s == "nq"
}

// NewLvsCareFromFlag init lvsCare client
func NewLvsCareFromFlag(opts options) (*LvsCare, error) {
	care := &LvsCare{
		virtualServer: opts.virtualServer,
		realServer:    opts.realServer,
		lvs:           service.NewLvsProxy(opts.scheduler),
	}

	schedulerOk := Scheduler(opts.scheduler).validate()
	if !schedulerOk {
		infoMsg := fmt.Errorf("LvsCare validate failed, invalid scheduler")
		return nil, infoMsg
	}

	return care, nil
}

// NewLvsCareFromConfig init lvsCare client
func NewLvsCareFromConfig(opts options) (*LvsCare, error) {
	config, err := ipvsConfig.ReadIpvsConfig(opts.ipvsPersistDir)
	if err != nil {
		log.Printf("read ipvs config failed: %v", err)
		return nil, nil
	}
	care := &LvsCare{
		virtualServer: config.VirtualServer,
		realServer:    config.RealServer,
		lvs:           service.NewLvsProxy(opts.scheduler),
	}
	schedulerOk := Scheduler(opts.scheduler).validate()
	if !schedulerOk {
		infoMsg := fmt.Errorf("LvsCare validate failed, invalid scheduler")
		return nil, infoMsg
	}
	return care, nil
}

// LvsCare for create or delete vs
type LvsCare struct {
	virtualServer string
	realServer    []string
	scheduler     Scheduler // nolint unused
	lvs           service.LvsProxy
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
		healthCheck, err := health.NewHealthConfig(opts.healthScheme, opts.healthPath)
		if err != nil {
			log.Printf("build health check client faild: %v", err)
		}
		ip, port := utils.SplitServer(r)
		if !healthCheck.IsHTTPAPIHealth(ip, port) {
			log.Printf("create rs [%s] failed, it is not health", r)
			continue
		}
		err = lvs.lvs.CreateRealServer(r)
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
