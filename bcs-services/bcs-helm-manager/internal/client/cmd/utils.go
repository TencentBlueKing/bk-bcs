/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/client/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/client/pkg/client"

	"github.com/spf13/viper"
)

func newClientWithConfiguration() pkg.HelmClient {
	return client.New(client.Config{
		APIServer: viper.GetString("config.apiserver"),
		AuthToken: viper.GetString("config.bcs_token"),
		Operator:  viper.GetString("config.operator"),
	})
}

func getInputData() ([]byte, error) {
	if jsonData != "" {
		return []byte(jsonData), nil
	}

	if jsonFile != "" {
		data, err := ioutil.ReadFile(jsonFile)
		if err != nil {
			return nil, err
		}

		return data, nil
	}

	return nil, fmt.Errorf("empty param data")
}
