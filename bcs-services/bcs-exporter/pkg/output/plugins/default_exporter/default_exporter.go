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

package main

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/protocol"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-exporter/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-exporter/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-exporter/pkg/output"
)

//Init plugin entrance
func Init(_ *config.Config) (output.PluginIf, error) { //nolint
	return &fileExporter{}, nil
}

type fileExporter struct {
}

func (fe *fileExporter) Key() output.PluginKey {
	return protocol.DefaultExporterPlugin
}

func (fe *fileExporter) Name() output.PluginName {
	return output.PluginName("default_exporter")
}

func (fe *fileExporter) SetCfg(_ *config.Config) error {
	return nil
}

func (fe *fileExporter) AddData(mapStr pkg.MapStr) error {
	data, ok := mapStr["data"]
	if !ok {
		return fmt.Errorf("data not found")
	}

	byteData, ok := data.([]byte)
	if !ok {
		return fmt.Errorf("data is not byte")
	}

	blog.Info("exporter get data: %s", string(byteData))

	return nil
}
