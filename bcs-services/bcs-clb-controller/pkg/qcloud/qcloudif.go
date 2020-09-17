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

package qcloud

import (
	"encoding/json"
	"fmt"
)

//APIInterface qcloud api interface
type APIInterface interface {
	DescribeLoadBalanceTaskResult(input *DescribeLoadBalancersTaskResultInput) (*DescribeLoadBalancersTaskResultOutput, error)
	//clb v2
	CreateLoadBalance(input *CreateLBInput) (*CreateLBOutput, error)
	DescribeLoadBalance(input *DescribeLBInput) (*DescribeLBOutput, error)
	Create7LayerListener(input *CreateSeventhLayerListenerInput) (*CreateSeventhLayerListenerOutput, error)
	Create4LayerListener(input *CreateForwardLBFourthLayerListenersInput) (*CreateForwardLBFourthLayerListenersOutput, error)
	DescribeListener(input *DescribeListenerInput) (*DescribeListenerOutput, error)
	DescribeForwardLBListeners(input *DescribeForwardLBListenersInput) (*DescribeForwardLBListenersOutput, error)
	DescribeForwardLBBackends(input *DescribeForwardLBBackendsInput) (*DescribeForwardLBBackendsOutput, error)
	DeleteListener(input *DeleteForwardLBListenerInput) (*DeleteForwardLBListenerOutput, error)
	Modify7LayerListener(input *ModifyForwardLBSeventhListenerInput) (*ModifyForwardLBSeventhListenerOutput, error)
	Modify4LayerListener(input *ModifyForwardLBFourthListenerInput) (*ModifyForwardLBFourthListenerOutput, error)

	RegInstancesWith4LayerListener(input *RegisterInstancesWithForwardLBFourthListenerInput) (*RegisterInstancesWithForwardLBFourthListenerOutput, error)
	DeRegInstancesWith4LayerListener(input *DeregisterInstancesFromForwardLBFourthListenerInput) (*DeregisterInstancesFromForwardLBFourthListenerOutput, error)
	RegInstancesWith7LayerListener(input *RegisterInstancesWithForwardLBSeventhListenerInput) (*RegisterInstancesWithForwardLBSeventhListenerOutput, error)
	DeRegInstancesWith7LayerListener(input *DeregisterInstancesFromForwardLBSeventhListenerInput) (*DeregisterInstancesFromForwardLBSeventhListenerOutput, error)
	ModifyForward4LayerBackendsWeight(input *ModifyForwardFourthBackendsInput) (*ModifyForwardFourthBackendsOutput, error)
	ModifyForward7LayerBackendsWeight(input *ModifyForwardSeventhBackendsInput) (*ModifyForwardSeventhBackendsOutput, error)
	CreateRules(input *CreateForwardLBListenerRulesInput) (*CreateForwardLBListenerRulesOutput, error)
	DeleteRules(input *DeleteForwardLBListenerRulesInput) (*DeleteForwardLBListenerRulesOutput, error)
	ModifyRuleDomain(input *ModifyForwardLBRulesDomainInput) (*ModifyForwardLBRulesDomainOutput, error)
	ModifyRuleProbe(input *ModifyLoadBalancerRulesProbeInput) (*ModifyLoadBalancerRulesProbeOutput, error)
	//cvm
	DescribeCVMInstance(input *DescribeCVMInstanceInput) (*DescribeCVMInstanceOutput, error)
	DescribeCVMInstanceV3(input *DescribeCVMInstanceV3Input) (*DescribeCVMInstanceV3Output, error)
}

//API api implements APIInterface
type API struct {
	LBClient  *Client
	CVMClient *Client
}

//NewAPI new api
func NewAPI(lbClient *Client, cvmClient *Client) APIInterface {
	return &API{
		LBClient:  lbClient,
		CVMClient: cvmClient,
	}
}

//DescribeLoadBalanceTaskResult describe lb task result
func (api *API) DescribeLoadBalanceTaskResult(input *DescribeLoadBalancersTaskResultInput) (*DescribeLoadBalancersTaskResultOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &DescribeLoadBalancersTaskResultOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//CreateLoadBalance create lb
func (api *API) CreateLoadBalance(input *CreateLBInput) (*CreateLBOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &CreateLBOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//DescribeLoadBalance describe lb
func (api *API) DescribeLoadBalance(input *DescribeLBInput) (*DescribeLBOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &DescribeLBOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//Create7LayerListener create http or https listener
func (api *API) Create7LayerListener(input *CreateSeventhLayerListenerInput) (*CreateSeventhLayerListenerOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &CreateSeventhLayerListenerOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//Create4LayerListener create tcp or udp listener
func (api *API) Create4LayerListener(input *CreateForwardLBFourthLayerListenersInput) (*CreateForwardLBFourthLayerListenersOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &CreateForwardLBFourthLayerListenersOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//DescribeListener describe listener info
func (api *API) DescribeListener(input *DescribeListenerInput) (*DescribeListenerOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &DescribeListenerOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

// DescribeForwardLBListeners describe forward loadbalancer listener info
func (api *API) DescribeForwardLBListeners(input *DescribeForwardLBListenersInput) (
	*DescribeForwardLBListenersOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &DescribeForwardLBListenersOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

// DescribeForwardLBBackends describe forward loadbalancer listener backends
func (api *API) DescribeForwardLBBackends(input *DescribeForwardLBBackendsInput) (
	*DescribeForwardLBBackendsOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &DescribeForwardLBBackendsOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//DeleteListener delete listener
func (api *API) DeleteListener(input *DeleteForwardLBListenerInput) (*DeleteForwardLBListenerOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &DeleteForwardLBListenerOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

// Modify7LayerListener modify 7 layer listener ssl config
func (api *API) Modify7LayerListener(input *ModifyForwardLBSeventhListenerInput) (*ModifyForwardLBSeventhListenerOutput, error) {
	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &ModifyForwardLBSeventhListenerOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

// Modify4LayerListener modify 4 layer listener session config and health check config
func (api *API) Modify4LayerListener(input *ModifyForwardLBFourthListenerInput) (*ModifyForwardLBFourthListenerOutput, error) {
	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &ModifyForwardLBFourthListenerOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//RegInstancesWith4LayerListener register instance with 4 layer listener
func (api *API) RegInstancesWith4LayerListener(input *RegisterInstancesWithForwardLBFourthListenerInput) (*RegisterInstancesWithForwardLBFourthListenerOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &RegisterInstancesWithForwardLBFourthListenerOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//DeRegInstancesWith4LayerListener deregister instance with 4 layer listener
func (api *API) DeRegInstancesWith4LayerListener(input *DeregisterInstancesFromForwardLBFourthListenerInput) (*DeregisterInstancesFromForwardLBFourthListenerOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &DeregisterInstancesFromForwardLBFourthListenerOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//RegInstancesWith7LayerListener register instance with 7 layer listener
func (api *API) RegInstancesWith7LayerListener(input *RegisterInstancesWithForwardLBSeventhListenerInput) (*RegisterInstancesWithForwardLBSeventhListenerOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &RegisterInstancesWithForwardLBSeventhListenerOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//DeRegInstancesWith7LayerListener deregister instance with 7 layer listener
func (api *API) DeRegInstancesWith7LayerListener(input *DeregisterInstancesFromForwardLBSeventhListenerInput) (*DeregisterInstancesFromForwardLBSeventhListenerOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &DeregisterInstancesFromForwardLBSeventhListenerOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//ModifyForward4LayerBackendsWeight modify 4 layer listener backends weight
func (api *API) ModifyForward4LayerBackendsWeight(input *ModifyForwardFourthBackendsInput) (*ModifyForwardFourthBackendsOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &ModifyForwardFourthBackendsOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//ModifyForward7LayerBackendsWeight modify 7 layer listener backends weight
func (api *API) ModifyForward7LayerBackendsWeight(input *ModifyForwardSeventhBackendsInput) (*ModifyForwardSeventhBackendsOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &ModifyForwardSeventhBackendsOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//CreateRules create lb listener rule
func (api *API) CreateRules(input *CreateForwardLBListenerRulesInput) (*CreateForwardLBListenerRulesOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &CreateForwardLBListenerRulesOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//DeleteRules delete lb listener rule
func (api *API) DeleteRules(input *DeleteForwardLBListenerRulesInput) (*DeleteForwardLBListenerRulesOutput, error) {

	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &DeleteForwardLBListenerRulesOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

// ModifyRuleDomain modify rule domain
func (api *API) ModifyRuleDomain(input *ModifyForwardLBRulesDomainInput) (*ModifyForwardLBRulesDomainOutput, error) {
	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &ModifyForwardLBRulesDomainOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

// ModifyRuleProbe modify rule health check config
func (api *API) ModifyRuleProbe(input *ModifyLoadBalancerRulesProbeInput) (*ModifyLoadBalancerRulesProbeOutput, error) {
	dataBytes, err := api.LBClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &ModifyLoadBalancerRulesProbeOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//DescribeCVMInstance describe cvm instance info
func (api *API) DescribeCVMInstance(input *DescribeCVMInstanceInput) (*DescribeCVMInstanceOutput, error) {

	dataBytes, err := api.CVMClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &DescribeCVMInstanceOutput{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}

//DescribeCVMInstanceV3 describe cvm instance api v3
func (api *API) DescribeCVMInstanceV3(input *DescribeCVMInstanceV3Input) (*DescribeCVMInstanceV3Output, error) {

	dataBytes, err := api.CVMClient.GetRequest(input)
	if err != nil {
		return nil, fmt.Errorf("GetRequest failed, err %s", err.Error())
	}
	output := &DescribeCVMInstanceV3Output{}
	if err := json.Unmarshal(dataBytes, output); err != nil {
		return nil, fmt.Errorf("parse response failed, err %s", err.Error())
	}
	return output, nil
}
