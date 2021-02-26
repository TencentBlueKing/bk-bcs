/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

import (
	"github.com/Tencent/bk-bcs/bcs-common/common"
)

const (
	// BcsErrClusterManagerSuccess success code
	BcsErrClusterManagerSuccess = 0
	// BcsErrClusterManagerSuccessStr success string
	BcsErrClusterManagerSuccessStr = "success"
	// BcsErrClusterManagerInvalidParameter invalid request parameter
	BcsErrClusterManagerInvalidParameter = common.AdditionErrorCode + 400
	// BcsErrClusterManagerStoreOperationFailed invalid request parameter
	BcsErrClusterManagerStoreOperationFailed = common.AdditionErrorCode + 401
	// BcsErrClusterManagerUnknown unknown error
	BcsErrClusterManagerUnknown = common.AdditionErrorCode + 402
	// BcsErrClusterManagerUnknownStr unknown error msg
	BcsErrClusterManagerUnknownStr = "unknown error"
	// BcsErrClusterManagerDBOperation db operation error
	BcsErrClusterManagerDBOperation = common.AdditionErrorCode + 403
	// BcsErrClusterManagerAllocateClusterInCreateQuota allocate cluster error
	BcsErrClusterManagerAllocateClusterInCreateQuota = common.AdditionErrorCode + 404
	// BcsErrClusterManagerK8SOpsFailed k8s operation failed
	BcsErrClusterManagerK8SOpsFailed = common.AdditionErrorCode + 405
	// BcsErrClusterManagerResourceDuplicated resource deplicated
	BcsErrClusterManagerResourceDuplicated = common.AdditionErrorCode + 406
	// BcsErrClusterManagerCommonErr common error
	BcsErrClusterManagerCommonErr = common.AdditionErrorCode + 407
)
