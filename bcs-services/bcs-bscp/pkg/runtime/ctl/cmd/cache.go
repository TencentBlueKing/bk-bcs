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
	"errors"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/cache-service/service/cache/client"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// WithRefreshCache init and returns the refreshing cache command.
func WithRefreshCache(op client.Interface) Cmd {
	cmd := &cacheCmd{
		op: op,
		cmd: &Command{
			Name:  "refresh-cache",
			Usage: "force refresh cache of an app, including app meta & released group & released ci",
			Parameters: []Parameter{{
				Name:  "biz_id",
				Usage: "defines the biz id of the app to refresh",
				Value: new(uint32),
			}, {
				Name:  "app_id",
				Usage: "defines the app id to refresh",
				Value: new(uint32),
			}},
			FromURL: true,
			Run: func(kt *kit.Kit, params map[string]interface{}) (interface{}, error) {
				bizIDVal, exists := params["biz_id"]
				if !exists {
					return nil, errf.New(errf.InvalidParameter, "biz_id is not set")
				}
				bizID, ok := bizIDVal.(*uint32)
				if !ok {
					return nil, errf.New(errf.InvalidParameter, "biz_id is not integer")
				}

				appIDVal, exists := params["app_id"]
				if !exists {
					return nil, errf.New(errf.InvalidParameter, "app_id is not set")
				}
				appID, ok := appIDVal.(*uint32)
				if !ok {
					return nil, errf.New(errf.InvalidParameter, "app_id is not integer")
				}

				if err := op.RefreshAppCache(kt, *bizID, *appID); err != nil {
					logs.Errorf("refresh biz %d app %d cache failed, rid: %s", bizID, appID, kt.Rid)
					return nil, err
				}

				logs.Infof("successfully refreshed biz %d app %d cache, rid: %s", bizID, appID, kt.Rid)
				return nil, nil
			},
		},
	}

	return cmd
}

// cacheCmd cache related Cmd.
type cacheCmd struct {
	cmd *Command
	op  client.Interface
}

// GetCommand get disable/enable write Command.
func (c *cacheCmd) GetCommand() *Command {
	return c.cmd
}

// Validate write server related Command.
func (c *cacheCmd) Validate() error {
	if c.op == nil {
		return errors.New("cache client is not set")
	}

	return c.cmd.Validate()
}
