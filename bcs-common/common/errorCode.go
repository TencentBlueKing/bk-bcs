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

package common

const AdditionErrorCode = 1405000

//bcs container service
//Bcs error number defined in this file
//Errno name is composed of the following format BcsErr{Module}{Type}
//all error code range 1401 001 ~ 14001 999
const (
	BcsSuccess    = 0
	BcsSuccessStr = "success"

	/*Common error code 1401 001~1401 049
	All common errno name is as a beginning to BcsErrComm*/

	//zookeeper
	BcsErrCommZkConnectFail       = AdditionErrorCode + 1
	BcsErrCommZkConnectFailStr    = "connect to zookeeper failed"
	BcsErrCommCreateZkNodeFail    = AdditionErrorCode + 2
	BcsErrCommCreateZkNodeFailStr = "cteate zookeeper node failed"
	BcsErrCommDeleteZkNodeFail    = AdditionErrorCode + 3
	BcsErrCommDeleteZkNodeFailStr = "delete zookeeper node failed"
	BcsErrCommGetZkNodeFail       = AdditionErrorCode + 4
	BcsErrCommGetZkNodeFailStr    = "get zookeeper node failed"
	BcsErrCommListZkNodeFail      = AdditionErrorCode + 5
	BcsErrCommListZkNodeFailStr   = "list zookeeper node failed"

	// http
	BcsErrCommHttpReadBodyFail        = AdditionErrorCode + 6
	BcsErrCommHttpReadBodyFailStr     = "read http body failed"
	BcsErrCommHttpParametersFailed    = AdditionErrorCode + 7
	BcsErrCommHttpParametersFailedStr = "http parameters is invalid"
	BcsErrCommRequestDataErr          = AdditionErrorCode + 8
	BcsErrCommRequestDataErrStr       = "request data not correct"
	BcsErrCommHttpDo                  = AdditionErrorCode + 9
	BcsErrCommHttpDoStr               = "do http request failed!"
	BcsErrCommJsonDecode              = AdditionErrorCode + 10
	BcsErrCommJsonDecodeStr           = "json decode failed!"
	BcsErrCommJsonEncode              = AdditionErrorCode + 11
	BcsErrCommJsonEncodeStr           = "json encode failed!"
	BcsErrCommHttpReadReqBody         = AdditionErrorCode + 12
	BcsErrCommHttpReadReqBodyStr      = "read http request body failed!"
	BcsErrCommHttpReadRsp             = AdditionErrorCode + 13
	BcsErrCommHttpReadRspStr          = "read http response failed!"
	BcsErrCommHttpNewRequest          = AdditionErrorCode + 14
	BcsErrCommHttpNewRequestStr       = "http new request failed!"

	/*Common error code 1401 050~1401 079
	bcs api module errno name is as a beginning to BcsErrApi*/
	//BcsErrApi
	BcsErrApiInternalFail            = AdditionErrorCode + 50
	BcsErrApiInternalFailStr         = "bcs apiserver internal error"
	BcsErrApiGetMesosApiFail         = AdditionErrorCode + 51
	BcsErrApiGetMesosApiFailStr      = "bcs mesos apiserver not found"
	BcsErrApiGetStorageFail          = AdditionErrorCode + 52
	BcsErrApiGetStorageFailStr       = "bcs storage not found"
	BcsErrApiRequestMesosApiFail     = AdditionErrorCode + 53
	BcsErrApiRequestMesosApiFailStr  = "request mesos apiserver failed"
	BcsErrApiGetNetserviceFail       = AdditionErrorCode + 54
	BcsErrApiGetNetserviceFailStr    = "bcs netservice not found"
	BcsErrApiGetMetricsFail          = AdditionErrorCode + 55
	BcsErrApiGetMetricsFailStr       = "bcs metrics not found"
	BcsErrApiGetK8sApiFail           = AdditionErrorCode + 56
	BcsErrApiGetK8sApiFailStr        = "bcs k8s apiserver not found"
	BcsErrApiAuthCheckFail           = AdditionErrorCode + 57
	BcsErrApiAuthCheckFailStr        = "bcs auth check error"
	BcsErrApiAuthCheckNoAuthority    = AdditionErrorCode + 58
	BcsErrApiAuthCheckNoAuthorityStr = "bcs auth check no authority"
	BcsErrApiUnauthorized            = AdditionErrorCode + 59
	BcsErrApiK8sClusterNotFound      = AdditionErrorCode + 60
	BcsErrApiBadRequest              = AdditionErrorCode + 61
	BcsErrApiInternalDbError         = AdditionErrorCode + 62
	BcsErrApiK8sInternalError        = AdditionErrorCode + 63
	BcsErrApiWebConsoleFailedCode    = AdditionErrorCode + 64
	BcsErrApiMediaTypeError          = AdditionErrorCode + 65
	BcsErrApiMediaTypeErrorStr       = "request yaml convert to json error"

	/*Common error code 1401 080~1401 109
	bcs storage module errno name is as a beginning to BcsErrStorage*/
	//BcsErrStorage
	BcsErrStorageRestRequestDataIsNotJson    = AdditionErrorCode + 80
	BcsErrStorageRestRequestDataIsNotJsonStr = "request data unmarshal failed."
	BcsErrStorageReturnDataIsNotJson         = AdditionErrorCode + 81
	BcsErrStorageReturnDataIsNotJsonStr      = "returned data marshal failed."
	BcsErrStoragePutResourceFail             = AdditionErrorCode + 82
	BcsErrStoragePutResourceFailStr          = "put resource failed."
	BcsErrStorageGetResourceFail             = AdditionErrorCode + 83
	BcsErrStorageGetResourceFailStr          = "get resource failed."
	BcsErrStorageDeleteResourceFail          = AdditionErrorCode + 84
	BcsErrStorageDeleteResourceFailStr       = "delete resource failed."
	BcsErrStorageListResourceFail            = AdditionErrorCode + 85
	BcsErrStorageListResourceFailStr         = "list resource failed."
	BcsErrStorageDecodeListResourceFail      = AdditionErrorCode + 86
	BcsErrStorageDecodeListResourceFailStr   = "base64 decode failed."
	BcsErrStorageResourceNotExist            = AdditionErrorCode + 87
	BcsErrStorageResourceNotExistStr         = "resource does not exist."
	BcsErrStorageStatusNotReady              = AdditionErrorCode + 88
	BcsErrStorageStatusNotReadyStr           = "status not ready"

	/*Common error code 1401 110~1401 139
	bcs metric service module errno name is as a beginning to BcsErrMetric*/
	//BcsErrMetric
	BcsErrMetricSetMetricFailed              = AdditionErrorCode + 110
	BcsErrMetricSetMetricFailedStr           = "failed to set metric"
	BcsErrMetricSubscriptionUnknown          = AdditionErrorCode + 111
	BcsErrMetricSubscriptionUnknownStr       = "unknown the subscription type"
	BcsErrMetricSetApplicationFailed         = AdditionErrorCode + 112
	BcsErrMetricSetApplicationFailedStr      = "failed to set application"
	BcsErrMetricSetCollectorFailed           = AdditionErrorCode + 113
	BcsErrMetricSetCollectorFailedStr        = "failed to set collector config"
	BcsErrMetricClusterTypeIsInvalid         = AdditionErrorCode + 114
	BcsErrMetricClusterTypeIsInvalidStr      = "the cluster type is invalid"
	BcsErrMetricSubscriptionIsInvalid        = AdditionErrorCode + 115
	BcsErrMetricSubscriptionIsInvalidStr     = "the subscription type is invalid"
	BcsErrMetricCreateSubscriptionFailed     = AdditionErrorCode + 116
	BcsErrMetricCreateSubscriptionFailedStr  = "failed to create the subscription"
	BcsErrMetricDestorySubscriptionFailed    = AdditionErrorCode + 117
	BcsErrMetricDestorySubscriptionFailedStr = "failed to destroy the subscription"
	BcsErrMetricGetCollectorFailed           = AdditionErrorCode + 118
	BcsErrMetricGetCollectorFailedStr        = "failed to get collector config"
	BcsErrMetricDeleteMetricFailed           = AdditionErrorCode + 119
	BcsErrMetricDeleteMetricFailedStr        = "failed to delete metric"
	BcsErrMetricStatusNotReady               = AdditionErrorCode + 120
	BcsErrMetricStatusNotReadyStr            = "status not ready"
	BcsErrMetricGetMetricFailed              = AdditionErrorCode + 121
	BcsErrMetricGetMetricFailedStr           = "failed to get metric"
	BcsErrMetricSetMetricTaskFailed          = AdditionErrorCode + 122
	BcsErrMetricSetMetricTaskFailedStr       = "failed to set metric task"
	BcsErrMetricDeleteMetricTaskFailed       = AdditionErrorCode + 123
	BcsErrMetricDeleteMetricTaskFailedStr    = "failed to delete metric task"
	BcsErrMetricGetMetricTaskFailed          = AdditionErrorCode + 124
	BcsErrMetricGetMetricTaskFailedStr       = "failed to get metric task"
	BcsErrMetricMetricTaskNotExist           = AdditionErrorCode + 125
	BcsErrMetricMetricTaskNotExistStr        = "metric task not exist"
	BcsErrMetricListMetricTaskFailed         = AdditionErrorCode + 126
	BcsErrMetricListMetricTaskFailedStr      = "failed to list metric task"
	BcsErrMetricStorageNoFound               = AdditionErrorCode + 127
	BcsErrMetricStorageNoFoundStr            = "storage not found"
	BcsErrMetricInvalidData                  = AdditionErrorCode + 128
	BcsErrMetricInvalidDataStr               = "invalid metric data"
	BcsErrMetricUnknownClusterType           = AdditionErrorCode + 129
	BcsErrMetricUnknownClusterTypeStr        = "unknown cluster type"
	BcsErrMetricResourceFileNotExist         = AdditionErrorCode + 130
	BcsErrMetricResourceFileNotExistStr      = "resource file not exist"

	/*Common error code 1401 140~1401 169
	bcs health service module errno name is as a beginning to BcsErrHealth*/
	//BcsErrHealth
	BcsErrHealthGetHealthzInfoErr    = AdditionErrorCode + 140
	BcsErrHealthGetHealthzInfoErrStr = "get healthz info error"

	/*Common error code 1401 170~1401 199
	bcs mesos api service module errno name is as a beginning to BcsErrMesosApi*/
	//BcsErrMesosApi

	/*Common error code 1401 200~1401 229
	bcs mesos scheduler service module errno name is as a beginning to BcsErrMesosSched*/
	//BcsErrMesosSched
	BcsErrMesosSchedCommon           = AdditionErrorCode + 200
	BcsErrMesosSchedCommonStr        = "scheduler common error"
	BcsErrMesosSchedResourceExist    = AdditionErrorCode + 201
	BcsErrMesosSchedResourceExistStr = "resource already exist"
	BcsErrMesosSchedNotFound         = AdditionErrorCode + 202
	BcsErrMesosSchedNotFoundStr      = "404 not found"

	/*Common error code 1401 230~1401 259
	bcs mesos driver module errno name is as a beginning to BcsErrMesosDriver*/
	//BcsErrMesosDriver
	BcsErrMesosDriverCommon               = AdditionErrorCode + 230
	BcsErrMesosDriverCommonStr            = "mesos driver common error"
	BcsErrMesosDriverParameterErr         = AdditionErrorCode + 231
	BcsErrMesosDriverParameterErrStr      = "there is some error in the parameters"
	BcsErrMesosDriverNoVersionId          = AdditionErrorCode + 232
	BcsErrMesosDriverNoVersionIdStr       = "no version id"
	BcsErrMesosDriverSendMsgUnknowType    = AdditionErrorCode + 233
	BcsErrMesosDriverSendMsgUnknowTypeStr = "unkown the message type"
	BcsErrMesosDriverHttpFilterFailed     = AdditionErrorCode + 234
	BcsErrMesosDriverHttpFilterFailedStr  = "bcs auth check no authority"

	/*Common error code 1401 260~1401 289
	bcs process daemon module errno name is as a beginning to BcsErrDaemon*/
	BcsErrDaemonCreateProcessFailed     = AdditionErrorCode + 260
	BcsErrDaemonCreateProcessFailedStr  = "failed to create process"
	BcsErrDaemonInspectProcessFailed    = AdditionErrorCode + 261
	BcsErrDaemonInspectProcessFailedStr = "failed to inspect process"
	BcsErrDaemonStopProcessFailed       = AdditionErrorCode + 262
	BcsErrDaemonStopProcessFailedStr    = "failed to stop process"
	BcsErrDaemonDeleteProcessFailed     = AdditionErrorCode + 263
	BcsErrDaemonDeleteProcessFailedStr  = "failed to delete process"

	/*Common error code 1401 290~1401 319*/
	//bcs-netservice error code
	//Bcs_Err_NETSERVICE_PARTIAL_ERR     = AdditionErrorCode + 290
	BcsErrNetservicePartialErr = AdditionErrorCode + 290
	//Bcs_Err_NETSERVICE_PARTIAL_ERR_STR = "partial error for requst"
	BcsErrNetservicePartialErrStr = "partial error for requst"
	//Bcs_ERR_NETSERVICE_FAILED          = AdditionErrorCode + 291
	BcsErrNetserviceFailed = AdditionErrorCode + 291
	//Bcs_ERR_NETSERVICE_PARAMETER_ERR   = AdditionErrorCode + 292
	BcsErrNetserviceParameterErr = AdditionErrorCode + 292
	//Bcs_ERR_NETSERVICE_THIRDPARTY_ERR  = AdditionErrorCode + 293
	BcsErrNetserviceThirdpartyErr = AdditionErrorCode + 293

	/*Common error code 1401 320~1401 349*/
	//bcs-ipservice error code
	//Bcs_ERR_IPSERVICE_PARTIAL_ERR    = AdditionErrorCode + 320
	BcsErrIPServicePartialErr = AdditionErrorCode + 320
	//Bcs_ERR_IPSERVICE_FAILED         = AdditionErrorCode + 321
	BcsErrIPServiceFailed = AdditionErrorCode + 321
	//Bcs_ERR_IPSERVICE_PARAMETER_ERR  = AdditionErrorCode + 322
	BcsErrIPServiceParameterErr = AdditionErrorCode + 322
	//Bcs_ERR_IPSERVICE_THIRDPARTY_ERR = AdditionErrorCode + 323
	BcsErrIPServiceThirdPartyErr = AdditionErrorCode + 323
)
