/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ess

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/requests"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/responses"
)

// ModifyScalingRule invokes the ess.ModifyScalingRule API synchronously
// api document: https://help.aliyun.com/api/ess/modifyscalingrule.html
func (client *Client) ModifyScalingRule(request *ModifyScalingRuleRequest) (response *ModifyScalingRuleResponse, err error) {
	response = CreateModifyScalingRuleResponse()
	err = client.DoAction(request, response)
	return
}

// ModifyScalingRuleWithChan invokes the ess.ModifyScalingRule API asynchronously
// api document: https://help.aliyun.com/api/ess/modifyscalingrule.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ModifyScalingRuleWithChan(request *ModifyScalingRuleRequest) (<-chan *ModifyScalingRuleResponse, <-chan error) {
	responseChan := make(chan *ModifyScalingRuleResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.ModifyScalingRule(request)
		if err != nil {
			errChan <- err
		} else {
			responseChan <- response
		}
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

// ModifyScalingRuleWithCallback invokes the ess.ModifyScalingRule API asynchronously
// api document: https://help.aliyun.com/api/ess/modifyscalingrule.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ModifyScalingRuleWithCallback(request *ModifyScalingRuleRequest, callback func(response *ModifyScalingRuleResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *ModifyScalingRuleResponse
		var err error
		defer close(result)
		response, err = client.ModifyScalingRule(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

// ModifyScalingRuleRequest is the request struct for api ModifyScalingRule
type ModifyScalingRuleRequest struct {
	*requests.RpcRequest
	ScalingRuleName      string           `position:"Query" name:"ScalingRuleName"`
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	AdjustmentValue      requests.Integer `position:"Query" name:"AdjustmentValue"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	Cooldown             requests.Integer `position:"Query" name:"Cooldown"`
	AdjustmentType       string           `position:"Query" name:"AdjustmentType"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	ScalingRuleId        string           `position:"Query" name:"ScalingRuleId"`
}

// ModifyScalingRuleResponse is the response struct for api ModifyScalingRule
type ModifyScalingRuleResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateModifyScalingRuleRequest creates a request to invoke ModifyScalingRule API
func CreateModifyScalingRuleRequest() (request *ModifyScalingRuleRequest) {
	request = &ModifyScalingRuleRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Ess", "2014-08-28", "ModifyScalingRule", "ess", "openAPI")
	return
}

// CreateModifyScalingRuleResponse creates a response to parse from ModifyScalingRule response
func CreateModifyScalingRuleResponse() (response *ModifyScalingRuleResponse) {
	response = &ModifyScalingRuleResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
