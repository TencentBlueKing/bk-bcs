/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"bk-bscp/pkg/common"
)

const (
	// e2eTestBizID is e2e biz id.
	e2eTestBizID = "e2e"

	// e2eTestOperator is e2e operator name.
	e2eTestOperator = "e2e"
)

var (
	createAppAPIV2 = "/api/v2/config/biz/%s/app"
	queryAppAPIV2  = "/api/v2/config/biz/%s/app"
	listAppAPIV2   = "/api/v2/config/list/biz/%s/app"
	updateAppAPIV2 = "/api/v2/config/biz/%s/app/%s"
	deleteAppAPIV2 = "/api/v2/config/biz/%s/app/%s"

	createConfigAPIV2 = "/api/v2/config/biz/%s/app/%s/config"
	queryConfigAPIV2  = "/api/v2/config/biz/%s/app/%s/config"
	listConfigAPIV2   = "/api/v2/config/list/biz/%s/app/%s/config"
	updateConfigAPIV2 = "/api/v2/config/biz/%s/app/%s/config/%s"
	deleteConfigAPIV2 = "/api/v2/config/biz/%s/app/%s/config/%s"

	createCommitAPIV2  = "/api/v2/config/biz/%s/app/%s/commit"
	queryCommitAPIV2   = "/api/v2/config/biz/%s/app/%s/commit/%s"
	listCommitAPIV2    = "/api/v2/config/list/biz/%s/app/%s/commit"
	updateCommitAPIV2  = "/api/v2/config/biz/%s/app/%s/commit/%s"
	cancelCommitAPIV2  = "/api/v2/config/cancel/biz/%s/app/%s/commit/%s"
	confirmCommitAPIV2 = "/api/v2/config/confirm/biz/%s/app/%s/commit/%s"

	createReleaseAPIV2   = "/api/v2/config/biz/%s/app/%s/release"
	queryReleaseAPIV2    = "/api/v2/config/biz/%s/app/%s/release/%s"
	updateReleaseAPIV2   = "/api/v2/config/biz/%s/app/%s/release/%s"
	cancelReleaseAPIV2   = "/api/v2/config/cancel/biz/%s/app/%s/release/%s"
	publishReleaseAPIV2  = "/api/v2/config/publish/biz/%s/app/%s/release/%s"
	rollbackReleaseAPIV2 = "/api/v2/config/rollback/biz/%s/app/%s/release/%s"
	listReleaseAPIV2     = "/api/v2/config/list/biz/%s/app/%s/release"

	createStrategyAPIV2 = "/api/v2/config/biz/%s/app/%s/strategy"
	queryStrategyAPIV2  = "/api/v2/config/biz/%s/app/%s/strategy"
	listStrategyAPIV2   = "/api/v2/config/list/biz/%s/app/%s/strategy"
	deleteStrategyAPIV2 = "/api/v2/config/biz/%s/app/%s/strategy/%s"
)

const (
	// TESTHOST is env var name of bscp e2e test gateway host.
	TESTHOST = "BSCP_E2E_TESTING_TESTHOST"

	// DEFAULTTESTHOST is default bscp e2e test gateway host.
	DEFAULTTESTHOST = "http://localhost:8080"
)

func testHost(path string) string {
	return common.GetenvCfg(TESTHOST, DEFAULTTESTHOST) + path
}

func respBody(resp *http.Response) (string, error) {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if len(data) == 0 {
		return "", errors.New("response body data empty")
	}
	return string(data), nil
}

func httpRequest(method, url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add(common.RidHeaderKey, common.Sequence())
	req.Header.Add(common.UserHeaderKey, e2eTestOperator)
	req.Header.Add(common.AppCodeHeaderKey, e2eTestBizID)

	return http.DefaultClient.Do(req)
}
