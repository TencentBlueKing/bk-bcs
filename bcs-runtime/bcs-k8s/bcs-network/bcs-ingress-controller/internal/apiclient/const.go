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

package apiclient

const (
	serviceName = "bkmonitorv3"
	urlPrefix   = "/prod"

	// SystemNameInMetricBlueKingMonitor system name in metric for bkmonitor
	SystemNameInMetricBlueKingMonitor = "bkmonitor"
	// HandlerNameInMetricBkmAPI handler name in metric for tencent cloud api
	HandlerNameInMetricBkmAPI = "api"
)

const (
	httpMethodGet    = "GET"
	httpMethodPost   = "POST"
	httpMethodPut    = "PUT"
	httpMethodDelete = "DELETE"
	httpMethodPatch  = "PATCH"
)

const (
	apigwApiScheme   = "https"
	envNameApiGwHost = "API_GW_HOST"
)
