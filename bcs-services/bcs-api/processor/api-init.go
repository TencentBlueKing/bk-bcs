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

package processor

import (
	//import v4 http clusterkeeper actions
	_ "bk-bcs/bcs-services/bcs-api/processor/http/actions/v4http/clusterkeeper"
	//import v4 http k8s actions
	_ "bk-bcs/bcs-services/bcs-api/processor/http/actions/v4http/k8s"
	//import v4 http mesos actions
	_ "bk-bcs/bcs-services/bcs-api/processor/http/actions/v4http/mesos"
	//import v4 http metrics actions
	_ "bk-bcs/bcs-services/bcs-api/processor/http/actions/v4http/metrics"
	//import v4 http netservice actions
	_ "bk-bcs/bcs-services/bcs-api/processor/http/actions/v4http/netservice"
	//import v4 http storage actions
	_ "bk-bcs/bcs-services/bcs-api/processor/http/actions/v4http/storage"
	//import v4 http detection actions
	_ "bk-bcs/bcs-services/bcs-api/processor/http/actions/v4http/detection"
)
