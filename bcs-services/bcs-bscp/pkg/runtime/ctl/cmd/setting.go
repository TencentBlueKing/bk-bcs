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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
)

// WithQueryRuntimeSetting init and returns the query runtime setting command for sidecar.
func WithQueryRuntimeSetting(settings interface{}) Cmd {
	cmd := &settingCmd{
		settings: settings,
	}

	cmd.cmd = &Command{
		Name: "query-setting",
		Usage: "query runtime setting, not only from the config file, but also from the runtime setting obtained " +
			"in other ways. setting data is sensitive information, so unnecessary query is prohibited.",
		Run: func(kt *kit.Kit, params map[string]interface{}) (interface{}, error) {
			setting := cmd.settings

			marshal, err := jsoni.Marshal(setting)
			if err != nil {
				return nil, errf.New(errf.Aborted, fmt.Sprintf("marshal runtime setting failed, %v", err))
			}

			logs.Infof("successfully query runtime setting, setting: %s, rid: %s", marshal, kt.Rid)

			// the JSON string after marshal is not returned directly, because the returned data will be returned to
			// the caller by marshal again, which will cause the marshal data to be more '\' before '"'.
			return setting, nil
		},
	}

	return cmd
}

// settingCmd setting related Cmd.
type settingCmd struct {
	cmd      *Command
	settings interface{}
}

// GetCommand get setting Command.
func (c *settingCmd) GetCommand() *Command {
	return c.cmd
}

// Validate setting related Command.
func (c *settingCmd) Validate() error {
	if c.settings == nil {
		return errors.New("setting is not set")
	}

	return c.cmd.Validate()
}
