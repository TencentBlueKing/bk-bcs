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

package cloudlb

import (
	cloudListenerType "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/network/v1"
)

const (
	// QCloudLB tencent cloud lb
	QCloudLB = "qcloudclb"
)

//Interface definition for cloud infrastructure
type Interface interface {
	LoadConfig() error                                                             //load all config item from Env (Subnet, security group)
	CreateLoadbalance() (*cloudListenerType.CloudLoadBalancer, error)              //create new loadbalancer if needed
	DescribeLoadbalance(name string) (*cloudListenerType.CloudLoadBalancer, error) //get loadbalancer by name, id or arn
	Update(old, cur *cloudListenerType.CloudListener) error                        //update event
	Add(ls *cloudListenerType.CloudListener) error                                 //new listener event
	Delete(ls *cloudListenerType.CloudListener) error                              //listener delete event
	ListListeners() ([]*cloudListenerType.CloudListener, error)                    // list all listener on clb instance controlled by this controller
}
