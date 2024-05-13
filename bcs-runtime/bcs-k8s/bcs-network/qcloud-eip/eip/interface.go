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

package eip

import "github.com/containernetworking/cni/pkg/skel"

// Interface interface for qcloud eip plugin
type Interface interface {
	Init(configFilePath string, eniNum int, ipNum int)
	Recover(configFilePath string, eniNum int)
	Release(configFilePath string)
	Clean(configFilePath string)
	CNIAdd(args *skel.CmdArgs) error
	CNIDel(args *skel.CmdArgs) error
}
