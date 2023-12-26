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
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// WithLog init and returns the log command.
func WithLog() Cmd {
	cmd := &defaultCmd{
		cmd: &Command{
			Name:  "log",
			Usage: "change log level",
			Parameters: []Parameter{{
				Name:  "v",
				Usage: "defines the log level to be changed",
				Value: new(int32),
			}},
			FromURL: true,
			Run: func(kt *kit.Kit, params map[string]interface{}) (interface{}, error) {
				v, exists := params["v"]
				if !exists {
					return nil, errf.New(errf.InvalidParameter, "v is not set")
				}

				logs.SetV(*v.(*int32))

				logs.Infof("successfully changed log level to %d, rid: %s", logs.GetV(), kt.Rid)
				return nil, nil
			},
		},
	}

	return cmd
}
