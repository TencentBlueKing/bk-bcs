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
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/auth-server/options"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// WithDisableWriteAuthAccess init and returns the disabling write auth access command, it blocks write operations
// on the authorization phase.
func WithDisableWriteAuthAccess(opt *options.DisableWriteOption) Cmd {
	cmd := &writeCmd{
		opt: opt,
		cmd: &Command{
			Name:  "disable-write-auth-access",
			Usage: "disable write operation for system publishing scenario that affects db",
			Parameters: []Parameter{{
				Name:  "biz_id",
				Usage: "defines the biz ids to disable write operations, separated by ','. eg. \"1\" or \"1,3,5\"",
				Value: new(string),
			}, {
				Name:  "is_all",
				Usage: "defines if all write operations need to be disabled, can not be set when biz_id is set",
				Value: new(bool),
			}},
			FromURL: true,
			Run: func(kt *kit.Kit, params map[string]interface{}) (interface{}, error) {
				isAllVal, isAllExists := params["is_all"]
				isAll := false
				if isAllExists {
					isAll = *isAllVal.(*bool)
				}

				bizIDVal, bizIDExists := params["biz_id"]
				bizIDStr := ""
				if bizIDExists {
					bizIDStr = strings.TrimSpace(*bizIDVal.(*string))
				}

				opt.IsDisabled = true

				// only one of biz_id and is_all needs to be set
				if isAll {
					if len(bizIDStr) > 0 {
						return nil, errf.New(errf.InvalidParameter, "biz_id and is_all can not both be set")
					}

					opt.IsAll = true
					logs.Infof("successfully disabled all write operations, rid: %s", kt.Rid)
					return nil, nil
				}

				if len(bizIDStr) == 0 {
					return nil, errf.New(errf.InvalidParameter, "one of biz_id and is_all must be set")
				}

				// split biz_id into separate ids, parse the ids into integer
				bizIDStrArr := strings.Split(bizIDStr, ",")
				for _, bizIDElement := range bizIDStrArr {
					if len(bizIDElement) == 1 {
						bizID, err := strconv.ParseUint(bizIDElement, 10, 64)
						if err != nil {
							logs.Errorf("parse biz id %s failed, err: %v, rid: %s", bizIDElement, err, kt.Rid)
							return nil, err
						}

						opt.BizIDMap.Store(uint32(bizID), struct{}{})
						continue
					}

					return nil, errf.New(errf.InvalidParameter, "parse biz_id element failed")
				}

				logs.Infof("successfully disabled write operations with %+v, rid: %s", opt, kt.Rid)
				return nil, nil
			},
		},
	}

	return cmd
}

// WithEnableWriteAuthAccess init and returns the enabling write command.
func WithEnableWriteAuthAccess(opt *options.DisableWriteOption) Cmd {
	cmd := &writeCmd{
		opt: opt,
		cmd: &Command{
			Name:    "enable-write-auth-access",
			Usage:   "enable all write operations that has been disabled",
			FromURL: true,
			Run: func(kt *kit.Kit, params map[string]interface{}) (interface{}, error) {
				opt.IsDisabled = false
				opt.IsAll = false
				opt.BizIDMap = sync.Map{}
				logs.Infof("successfully enabled all write operations, rid: %s", kt.Rid)
				return nil, nil
			},
		},
	}

	return cmd
}

// WithGetDisableWriteAuthAccess init and returns the getting disable write option command.
func WithGetDisableWriteAuthAccess(opt *options.DisableWriteOption) Cmd {
	cmd := &writeCmd{
		opt: opt,
		cmd: &Command{
			Name:    "get-write-auth-access",
			Usage:   "get write operations config",
			FromURL: true,
			Run: func(kt *kit.Kit, params map[string]interface{}) (interface{}, error) {
				var err error
				bizIDs := make([]uint32, 0)

				opt.BizIDMap.Range(func(key, value interface{}) bool {
					bizID, ok := key.(uint32)
					if !ok {
						err = fmt.Errorf("biz id %+v is not integer", key)
						return false
					}
					bizIDs = append(bizIDs, bizID)
					return true
				})

				if err != nil {
					return nil, err
				}

				return map[string]interface{}{
					"is_disabled": opt.IsDisabled,
					"is_all":      opt.IsAll,
					"biz_ids":     bizIDs,
				}, nil
			},
		},
	}

	return cmd
}

// writeCmd disable/enable write Cmd.
type writeCmd struct {
	cmd *Command
	opt *options.DisableWriteOption
}

// GetCommand get disable/enable write Command.
func (c *writeCmd) GetCommand() *Command {
	return c.cmd
}

// Validate write related Command.
func (c *writeCmd) Validate() error {
	if c.opt == nil {
		return errors.New("disable write option is not set")
	}

	return c.cmd.Validate()
}

// WithWrites init and returns the disable/enable/get write operations commands that auth server needed.
func WithWrites(opt *options.DisableWriteOption) []Cmd {
	return []Cmd{WithDisableWriteAuthAccess(opt), WithEnableWriteAuthAccess(opt), WithGetDisableWriteAuthAccess(opt)}
}
