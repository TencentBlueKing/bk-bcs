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
package cmdb

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/config"
	svcConfig "github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/config"
)

var (
	username = config.AnonymousUsername
	bizID    = "1"
)

func TestCheckMaintainer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"code": 0, "data": {"count": 1, "info": [{"bk_biz_id": 1}]}}`))
	}))
	defer ts.Close()

	// 需要加载配置，然后域名调整为mock的url
	svcConfig.LoadConfig("../../../" + config.DefaultConfigPath)
	svcConfig.GlobalConf.CMDB.Host = ts.URL
	isMaintainer, err := IsMaintainer(username, bizID)
	assert.Nil(t, err)
	assert.True(t, isMaintainer)
}
