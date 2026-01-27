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
// package xxx
package config

import (
	"fmt"
	"os"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/yaml"
)

// G is the global configuration
var G = &Configuration{}

// Init init config
func Init(name string) error {
	if name == "" {
		return fmt.Errorf("config file name is empty")
	}

	content, err := os.ReadFile(name)
	if err != nil {
		return err
	}

	ctrl.Log.WithName("config").Info(fmt.Sprintf("config content: %s", string(content)))

	if err := yaml.Unmarshal(content, G); err != nil {
		return err
	}

	return nil
}
