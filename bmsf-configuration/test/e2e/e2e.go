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
	"io/ioutil"
	"net/http"

	"bk-bscp/pkg/common"
)

var (
	businessInterfaceV1            = "/v1/interface/businesses"
	businessListInterfaceV1        = "/v1/interface/businesslist"
	appInterfaceV1                 = "/v1/interface/apps"
	appListInterfaceV1             = "/v1/interface/applist"
	clusterInterfaceV1             = "/v1/interface/clusters"
	clusterListInterfaceV1         = "/v1/interface/clusterlist"
	zoneInterfaceV1                = "/v1/interface/zones"
	zoneListInterfaceV1            = "/v1/interface/zonelist"
	configsetInterfaceV1           = "/v1/interface/configsets"
	configsetListInterfaceV1       = "/v1/interface/configsetlist"
	configsetLockInterfaceV1       = "/v1/interface/configset-locks"
	commitInterfaceV1              = "/v1/interface/commits"
	commitHistoryInterfaceV1       = "/v1/interface/history-commits"
	commitCancelInterfaceV1        = "/v1/interface/cancel-commit"
	commitConfirmInterfaceV1       = "/v1/interface/confirm-commit"
	commitPreviewInterfaceV1       = "/v1/interface/preview-commit"
	configsInterfaceV1             = "/v1/interface/configs"
	configsListInterfaceV1         = "/v1/interface/configslist"
	releaseInterfaceV1             = "/v1/interface/releases"
	releaseHistoryInterfaceV1      = "/v1/interface/history-releases"
	releasePubInterfaceV1          = "/v1/interface/pub-release"
	releaseCancelInterfaceV1       = "/v1/interface/cancel-release"
	strategyInterfaceV1            = "/v1/interface/strategies"
	strategyListInterfaceV1        = "/v1/interface/strategylist"
	variableInterfaceV1            = "/v1/interface/variables"
	variableListInterfaceV1        = "/v1/interface/variablelist"
	templatesetInterfaceV1         = "/v1/interface/configtemplatesets"
	templatesetListInterfaceV1     = "/v1/interface/configtemplatesetlist"
	templateInterfaceV1            = "/v1/interface/configtemplates"
	templateListInterfaceV1        = "/v1/interface/configtemplatelist"
	templateversionInterfaceV1     = "/v1/interface/templateversions"
	templateversionListInterfaceV1 = "/v1/interface/templateversionlist"
	templatebindingInterfaceV1     = "/v1/interface/templatebindings"
	templatebindingListInterfaceV1 = "/v1/interface/templatebindinglist"
)

const (
	// TESTHOST is env var name of bscp e2e test gateway host.
	TESTHOST = "BSCP_E2E_TESTING_TESTHOST"

	// DEFAULTTESTHOST is default bscp e2e test gateway host.
	DEFAULTTESTHOST = "http://localhost:8080"
)

func testhost(path string) string {
	return common.GetenvCfg(TESTHOST, DEFAULTTESTHOST) + path
}

func respbody(resp *http.Response) (string, error) {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if len(data) == 0 {
		return "", errors.New("response body data empty")
	}
	return string(data), nil
}
