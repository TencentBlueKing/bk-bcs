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

// CreateScalingRule invokes the ess.CreateScalingRule API synchronously
// api document: https://help.aliyun.com/api/ess/createscalingrule.html
func (client *Client) CreateScalingRule(request *CreateScalingRuleRequest) (response *CreateScalingRuleResponse, err error) {
	response = CreateCreateScalingRuleResponse()
	err = client.DoAction(request, response)
	return
}

// CreateScalingRuleWithChan invokes the ess.CreateScalingRule API asynchronously
// api document: https://help.aliyun.com/api/ess/createscalingrule.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) CreateScalingRuleWithChan(request *CreateScalingRuleRequest) (<-chan *CreateScalingRuleResponse, <-chan error) {
	responseChan := make(chan *CreateScalingRuleResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.CreateScalingRule(request)
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

// CreateScalingRuleWithCallback invokes the ess.CreateScalingRule API asynchronously
// api document: https://help.aliyun.com/api/ess/createscalingrule.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) CreateScalingRuleWithCallback(request *CreateScalingRuleRequest, callback func(response *CreateScalingRuleResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *CreateScalingRuleResponse
		var err error
		defer close(result)
		response, err = client.CreateScalingRule(request)
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

// CreateScalingRuleRequest is the request struct for api CreateScalingRule
type CreateScalingRuleRequest struct {
	*requests.RpcRequest
	ScalingRuleName      string           `position:"Query" name:"ScalingRuleName"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	AdjustmentValue      requests.Integer `position:"Query" name:"AdjustmentValue"`
	ScalingGroupId       string           `position:"Query" name:"ScalingGroupId"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	Cooldown             requests.Integer `position:"Query" name:"Cooldown"`
	AdjustmentType       string           `position:"Query" name:"AdjustmentType"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
}

// CreateScalingRuleResponse is the response struct for api CreateScalingRule
type CreateScalingRuleResponse struct {
	*responses.BaseResponse
	ScalingRuleId  string `json:"ScalingRuleId" xml:"ScalingRuleId"`
	ScalingRuleAri string `json:"ScalingRuleAri" xml:"ScalingRuleAri"`
	RequestId      string `json:"RequestId" xml:"RequestId"`
}

// CreateCreateScalingRuleRequest creates a request to invoke CreateScalingRule API
func CreateCreateScalingRuleRequest() (request *CreateScalingRuleRequest) {
	request = &CreateScalingRuleRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Ess", "2014-08-28", "CreateScalingRule", "ess", "openAPI")
	return
}

// CreateCreateScalingRuleResponse creates a response to parse from CreateScalingRule response
func CreateCreateScalingRuleResponse() (response *CreateScalingRuleResponse) {
	response = &CreateScalingRuleResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
