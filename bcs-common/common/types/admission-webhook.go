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

package types

type AdmissionWebhookConfiguration struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata"`
	//resources ref
	ResourcesRef *ResourcesRef
	//adminssion webhook info
	AdmissionWebhooks []*AdmissionWebhook `json:"admissionWebhooks"`
}

type ResourcesRef struct {
	//admission operation, Http method: POST=Create; PUT=Update
	Operation AdmissionOperation
	//resources kind, mesos resources json: Deployment,Application...
	Kind AdmissionResourcesKind
}

type AdmissionWebhook struct {
	//admission webhook name
	Name string
	//failurePolicy, if communication with webhook service failed,
	//according to FailurePolicy to decide continue or return fail
	FailurePolicy WebhookFailurePolicyKind
	//webhook http client config
	ClientConfig *WebhookClientConfig
	//webhook server list, examples: ["https://127.0.0.1:31000","https://127.0.0.1:31001",...]
	WebhookServers []string `json:"-"`
}

type AdmissionOperation string

const (
	AdmissionOperationCreate  = "Create"
	AdmissionOperationUpdate  = "Update"
	AdmissionOperationUnknown = "unknown"
)

type AdmissionResourcesKind string

const (
	AdmissionResourcesApplication = "application"
	AdmissionResourcesDeployment  = "deployment"
)

type WebhookFailurePolicyKind string

const (
	WebhookFailurePolicyIgnore = "Ignore"
	WebhookFailurePolicyFail   = "Fail"
)

type WebhookClientConfig struct {
	//pem encoded ca cert that signs the server cert used by the webhook
	CaBundle string
	//webhook service namespace
	Namespace string
	//webhook service name
	Name string
}
