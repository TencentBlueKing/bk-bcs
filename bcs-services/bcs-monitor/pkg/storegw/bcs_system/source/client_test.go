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

package source

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	bkmonitor_client "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_monitor"
	bcstesting "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/testing"
)

func TestIsBKMonitorEnabled(t *testing.T) {
	ok, err := bkmonitor_client.IsBKMonitorEnabled(context.Background(), bcstesting.GetTestClusterId())
	assert.NoError(t, err)
	assert.Equal(t, ok, false)
}
