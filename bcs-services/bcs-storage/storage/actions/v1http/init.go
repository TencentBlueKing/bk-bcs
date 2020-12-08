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

package v1http

import (
	// trigger all package init to register handlers to actions
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/alarms"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/clusterconfig"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/dynamic"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/dynamicquery"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/dynamicwatch"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/events"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/hostconfig"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/metric"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/metricwatch"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/watchk8smesos"
)
