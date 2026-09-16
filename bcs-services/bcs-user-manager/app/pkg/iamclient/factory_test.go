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

package iamclient

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testIAMFactoryOptions() iam.Options {
	return iam.Options{
		SystemID:    iam.SystemIDBKBCS,
		AppCode:     "test-app",
		AppSecret:   "test-secret",
		External:    false,
		GateWayHost: "http://127.0.0.1",
		Metric:      true,
		Debug:       false,
	}
}

func TestNewFactoryReuseAndMultiTenant(t *testing.T) {
	factory := NewFactory(testIAMFactoryOptions())

	defaultCli := factory("")
	require.NotNil(t, defaultCli)
	assert.Same(t, defaultCli, factory(""))
	assert.Same(t, defaultCli, factory(iam.DefaultTenantId))

	systemCli := factory("system")
	require.NotNil(t, systemCli)
	assert.Same(t, systemCli, factory("system"))
	assert.NotSame(t, defaultCli, systemCli)
}

func TestNewAuthFactoryV4(t *testing.T) {
	factory := NewAuthFactory(testIAMFactoryOptions(), true, "http://127.0.0.1:8080")
	cli := factory("")
	require.NotNil(t, cli)
	assert.Same(t, cli, factory(""))
	assert.NotSame(t, cli, factory("system"))
}

func TestNewAuthFactoryV4RequiresHost(t *testing.T) {
	factory := NewAuthFactory(testIAMFactoryOptions(), true, "")
	assert.Panics(t, func() { _ = factory("") })
}

func TestNewV3MigrateClient(t *testing.T) {
	opt := testIAMFactoryOptions()
	iam.ApplyV4Config(&opt, true, "http://127.0.0.1:8080")
	cli, err := NewV3MigrateClient(opt)
	require.NoError(t, err)
	require.NotNil(t, cli)
}
