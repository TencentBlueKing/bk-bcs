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

package bcscc

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/config"
	svcConfig "github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/config"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store/project"
)

var project = &pm.Project{
	CreateTime:  "2022-05-09T16:33:53Z",
	UpdateTime:  "2022-05-09T16:33:53Z",
	Creator:     "admin",
	Updater:     "admin",
	ProjectID:   "test",
	Name:        "test",
	ProjectCode: "test",
	UseBKRes:    false,
	Description: "test",
	IsOffline:   false,
	Kind:        "k8s",
	BusinessID:  "1",
	IsSecret:    false,
	ProjectType: 1,
	DeployType:  1,
	BGID:        "1",
	BGName:      "test",
	DeptID:      "1",
	DeptName:    "test",
	CenterID:    "1",
	CenterName:  "test",
}

type FakeProject struct {
	Name        string `json:"project_name"`
	EnglishName string `json:"english_name"`
	Creator     string `json:"creator"`
	Updator     string `json:"updator"`
	Description string `json:"desc"`
	ProjectType uint   `json:"project_type"`
	IsOfflined  bool   `json:"is_offlined"`
	ProjectID   string `json:"project_id"`
	UseBK       bool   `json:"use_bk"`
	CCAppID     uint   `json:"cc_app_id"`
	Kind        int    `json:"kind"`
	DeployType  []int  `json:"deploy_type"`
	BGID        uint   `json:"bg_id"`
	BGName      string `json:"bg_name"`
	DeptID      uint   `json:"dept_id"`
	DeptName    string `json:"dept_name"`
	CenterID    uint   `json:"center_id"`
	CenterName  string `json:"center_name"`
	DataID      uint   `json:"data_id"`
	IsSecrecy   bool   `json:"is_secrecy"`
}

func TestCreateProject(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 校验获取到的数据
		body, _ := ioutil.ReadAll(r.Body)
		var p FakeProject
		err := json.Unmarshal(body, &p)
		assert.Nil(t, err)
		assert.Equal(t, p.Name, "test")
		assert.Equal(t, p.Kind, 1)
		assert.Equal(t, p.Creator, "admin")
		// 设置返回的数据
		w.Write([]byte(`{"code": 0, "request_id": "requestid", "message":"success"}`))
	}))
	defer ts.Close()

	svcConfig.LoadConfig("../../../" + config.DefaultConfigPath)
	svcConfig.GlobalConf.BCSCC.Host = ts.URL
	err := CreateProject(project)
	assert.Nil(t, err)
}

func TestUpdateProject(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 校验获取到的数据
		body, _ := ioutil.ReadAll(r.Body)
		var p FakeProject
		err := json.Unmarshal(body, &p)
		assert.Nil(t, err)
		assert.Equal(t, p.Name, "test")
		assert.Equal(t, p.Kind, 1)
		assert.Equal(t, p.Updator, "admin")
		// 设置返回的数据
		w.Write([]byte(`{"code": 0, "request_id": "requestid", "message":"success"}`))
	}))
	defer ts.Close()

	svcConfig.LoadConfig("../../../" + config.DefaultConfigPath)
	svcConfig.GlobalConf.BCSCC.Host = ts.URL
	err := UpdateProject(project)
	assert.Nil(t, err)
}
