/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/config"
)

func TestLoadConfig(t *testing.T) {
	c, err := LoadConfig("../../" + config.DefaultConfigPath)
	if err != nil {
		t.Errorf("Load default config error: %v", err)
	}

	// etcd config
	expectedEtcdEP := "127.0.0.1:2379"
	assert.Equal(t, expectedEtcdEP, c.Etcd.EtcdEndpoints)

	// mongo config
	expectedMongoAddr := "127.0.0.1:27017"
	assert.Equal(t, expectedMongoAddr, c.Mongo.Address)

	// log config
	expectedLogFileName := "project.log"
	assert.Equal(t, expectedLogFileName, c.Log.Name)
}
