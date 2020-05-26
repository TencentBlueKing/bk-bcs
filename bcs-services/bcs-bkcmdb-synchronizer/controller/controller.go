/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/discovery"
	"bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/taskinformer"
	"bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/taskmanager"
)

// Controller controller for controlling reconciler life cycle
type Controller struct {
	disc     *discovery.Client
	informer *taskinformer.Informer
	manager  *taskmanager.Manager
	isMaster bool
}

// NewController create controller
func NewController() *Controller {
	return &Controller{}
}

// Run run the controller
func (c *Controller) Run() {
	
}

// masterLoop
func (c *Controller) masterLoop() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			isMaster := c.disc.IsMaster()
			if isMaster && !c.isMaster {
				c.isMaster = isMaster
				blog.Infof("I become master, start task manager")
				go c.manager.Run()

			} else if !isMaster && c.isMaster {
				c.isMaster = isMaster
				blog.Infof("I become slave, stop task manager")
				c.manager.Stop()
			}
		}
	}
}
