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

// Package cmd init service options
package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/options"
)

// Parse parses the command-line flags and configuration file to create service options.
func Parse() (*options.ServiceOptions, error) {
	configPath := flag.String("f", "./bcs-push-manager.json", "Configuration file path")
	flag.Parse()

	opt := options.NewServiceOptions()
	if err := loadConfigFile(*configPath, opt); err != nil {
		return nil, fmt.Errorf("load config from %s failed: %w", *configPath, err)
	}
	return opt, nil
}

func loadConfigFile(fileName string, opt *options.ServiceOptions) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, opt)
}
