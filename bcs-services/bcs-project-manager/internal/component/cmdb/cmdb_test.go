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
package cmdb

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/headerkey"
	svcConfig "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/tenant"
)

var (
	username = constant.AnonymousUsername
	bizID    = "1"
)

func TestCheckMaintainer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code": 0, "data": {"count": 1, "info": [{"bk_biz_id": 1}]}}`))
	}))
	defer ts.Close()

	svcConfig.GlobalConf = &svcConfig.ProjectConfig{
		CMDB: svcConfig.CMDBConfig{Host: ts.URL, Timeout: 5, BKSupplierAccount: "tencent"},
		App:  svcConfig.AppConfig{Code: "c", Secret: "s", BkUsername: "admin"},
	}
	isMaintainer, err := IsMaintainer(context.Background(), username, bizID)
	assert.Nil(t, err)
	assert.True(t, isMaintainer)
}

func TestListAllBusinessesPaginated(t *testing.T) {
	orig := searchBusinessBatchSize
	searchBusinessBatchSize = 1
	defer func() { searchBusinessBatchSize = orig }()

	var gotPath, gotTenant string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotTenant = r.Header.Get(string(headerkey.TenantIdKey))
		var payload struct {
			Page struct {
				Start int `json:"start"`
			} `json:"page"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		pages := []string{
			`{"code":0,"data":{"count":2,"info":[{"bk_biz_id":1,"bk_biz_name":"a","bk_biz_productor":"p"}]}}`,
			`{"code":0,"data":{"count":2,"info":[{"bk_biz_id":2,"bk_biz_name":"b"}]}}`,
		}
		if payload.Page.Start >= len(pages) {
			_, _ = w.Write([]byte(`{"code":0,"data":{"count":2,"info":[]}}`))
			return
		}
		_, _ = w.Write([]byte(pages[payload.Page.Start]))
	}))
	defer ts.Close()

	svcConfig.GlobalConf = &svcConfig.ProjectConfig{
		CMDB: svcConfig.CMDBConfig{Host: ts.URL, Timeout: 5, BKSupplierAccount: "tencent"},
		App:  svcConfig.AppConfig{Code: "c", Secret: "s", BkUsername: "admin"},
	}
	list, err := ListAllBusinesses(tenant.WithTenantIdFromContext(context.Background(), "t1"))
	require.NoError(t, err)
	assert.Equal(t, "/api/v3/biz/search/tencent", gotPath)
	assert.Equal(t, "t1", gotTenant)
	require.Len(t, list, 2)
	assert.Equal(t, int64(1), list[0].BKBizID)
	assert.Equal(t, "p", list[0].BkBizProductor)
	assert.Equal(t, int64(2), list[1].BKBizID)
}

func TestListAllBusinessesError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":1,"message":"cmdb down"}`))
	}))
	defer ts.Close()
	svcConfig.GlobalConf = &svcConfig.ProjectConfig{
		CMDB: svcConfig.CMDBConfig{Host: ts.URL, Timeout: 5, BKSupplierAccount: "tencent"},
		App:  svcConfig.AppConfig{},
	}
	_, err := ListAllBusinesses(context.Background())
	assert.Error(t, err)
}
