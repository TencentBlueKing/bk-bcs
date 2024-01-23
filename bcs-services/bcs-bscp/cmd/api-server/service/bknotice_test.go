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

package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"text/template"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/api-server/options"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
)

// SysOpt is the system option
var SysOpt *options.Option

var configFileTmpl = `
esb:
  appCode: {{ .AppCode }}
  appSecret: {{ .AppSecret }}
bkNotice:
  enable: true
  host: {{ .BKNoticeHost }}
repository:
  # storageType: S3
  storageType: BKREPO
  bkRepo:
    endpoints:
      - http://example.bktencent.com
    project: bscp
    username: example
    password: example
  s3:
    endpoint: ""
    accessKeyID: ""
    secretAccessKey: ""
    useSSL: true
    bucketName: bscp-example
`

type MockHttpClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestGetCurrentAnnouncements(t *testing.T) {

	tempDir := os.TempDir()
	SysOpt.Sys.ConfigFiles = []string{path.Join(tempDir, "cc.yaml")}
	file, err := os.OpenFile(path.Join(tempDir, "cc.yaml"), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		t.Errorf("open file failed, err: %s", err.Error())
	}
	defer file.Close()
	tpl, err := template.New("cc.yaml").Parse(configFileTmpl)
	if err != nil {
		t.Errorf("parse template failed, err: %s", err.Error())
	}
	if e := tpl.Execute(file, map[string]string{
		"AppCode":      os.Getenv("APP_CODE"),
		"AppSecret":    os.Getenv("APP_SECRET"),
		"BKNoticeHost": os.Getenv("BK_NOTICE_HOST"),
	}); e != nil {
		t.Errorf("execute template failed, err: %s", e.Error())
	}

	if e := cc.LoadSettings(SysOpt.Sys); e != nil {
		t.Errorf("load settings from config files failed, err: %s", e.Error())
	}

	service, err := newBKNoticeService()
	if err != nil {
		t.Errorf("newBKNoticeService failed: %v", err)
	}

	recorder := httptest.NewRecorder()
	// Mock 一个http请求
	request, _ := http.NewRequest("GET", "/", nil)

	// 调用API
	service.GetCurrentAnnouncements(recorder, request)

	// 检查状态码
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// 检查响应体
	type expectedResp struct {
		Result bool        `json:"result"`
		Code   int         `json:"code"`
		Data   interface{} `json:"data"`
	}
	expected := &expectedResp{}
	if err := json.Unmarshal(recorder.Body.Bytes(), expected); err != nil {
		t.Errorf("handler returned response failed: %v", err)
	}

	if expected.Code != 0 {
		t.Errorf("handler returned response failed, code: %v", expected.Code)
	}

	if !expected.Result {
		t.Errorf("handler returned response failed, result: %v", expected.Result)
	}
}

func init() {

	SysOpt = options.InitOptions()

	cc.InitService(cc.APIServerName)
}
