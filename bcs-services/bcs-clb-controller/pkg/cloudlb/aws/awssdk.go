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

package aws

import (
	"bk-bcs/bcs-common/common/blog"
	loadbalance "bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/network/v1"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

type awsElbSdkAPI struct{}

func NewAwsElbSdkAPI() *awsElbSdkAPI {
	return &awsElbSdkAPI{}
}

//awsCreateLoadBalancer use elb sdk to CreateLoadBalancer
func (a *awsElbSdkAPI) awsCreateLoadBalancer(subNets, securityGroups []string, name, networkType, lbType string) (string, error) {
	var reqSubNets []*string
	for _, sn := range subNets {
		reqSubNets = append(reqSubNets, aws.String(sn))
	}
	var reqSecurityGroups []*string
	for _, sg := range securityGroups {
		reqSecurityGroups = append(reqSecurityGroups, aws.String(sg))
	}

	networkType, err := networkTypeBcs2Aws(networkType)
	if err != nil {
		return "", fmt.Errorf("networkTypeBcs2Aws failed, %s", err.Error())
	}
	// lbType, err = lbTypeBcs2Aws(lbType)
	// if err != nil {
	// 	return "", fmt.Errorf("lbTypeBcs2Aws failed, %s", err.Error())
	// }

	//the default is internet-facing
	elbScheme := elbv2.LoadBalancerSchemeEnumInternetFacing
	if strings.ToLower(networkType) == AWS_LOADBALANCE_NETWORK_PRIVATE {
		//internal network
		elbScheme = elbv2.LoadBalancerSchemeEnumInternal
	}
	//the default is appliction type lb
	elbType := elbv2.LoadBalancerTypeEnumApplication
	// if strings.ToLower(lbType) == AWS_LOADBALANCE_TYPE_NETWORK {
	// 	//lb with tcp protocol is network type lb
	// 	elbType = elbv2.LoadBalancerTypeEnumNetwork
	// 	//appliction type lb can't bind with securityGroup
	// 	securityGroups = nil
	// }

	awsClient := elbv2.New(session.New())
	reqCreateLB := &elbv2.CreateLoadBalancerInput{
		Name:           aws.String(name),
		Scheme:         aws.String(elbScheme),
		Type:           aws.String(elbType),
		Subnets:        reqSubNets,
		SecurityGroups: reqSecurityGroups,
	}
	blog.Infof("CreateLoadBalancerInput: %v", reqCreateLB)

	respCreateLB, err := awsClient.CreateLoadBalancer(reqCreateLB)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case elbv2.ErrCodeDuplicateLoadBalancerNameException:
				blog.Errorf("CreateLoadBalancer failed:%v, %s", elbv2.ErrCodeDuplicateLoadBalancerNameException, awsErr.Error())
			case elbv2.ErrCodeTooManyLoadBalancersException:
				blog.Errorf("CreateLoadBalancer failed:%v, %s", elbv2.ErrCodeTooManyLoadBalancersException, awsErr.Error())
			case elbv2.ErrCodeInvalidConfigurationRequestException:
				blog.Errorf("CreateLoadBalancer failed:%v, %s", elbv2.ErrCodeInvalidConfigurationRequestException, awsErr.Error())
			case elbv2.ErrCodeSubnetNotFoundException:
				blog.Errorf("CreateLoadBalancer failed:%v, %s", elbv2.ErrCodeSubnetNotFoundException, awsErr.Error())
			case elbv2.ErrCodeInvalidSubnetException:
				blog.Errorf("CreateLoadBalancer failed:%v, %s", elbv2.ErrCodeInvalidSubnetException, awsErr.Error())
			case elbv2.ErrCodeInvalidSecurityGroupException:
				blog.Errorf("CreateLoadBalancer failed:%v, %s", elbv2.ErrCodeInvalidSecurityGroupException, awsErr.Error())
			case elbv2.ErrCodeInvalidSchemeException:
				blog.Errorf("CreateLoadBalancer failed:%v, %s", elbv2.ErrCodeInvalidSchemeException, awsErr.Error())
			case elbv2.ErrCodeTooManyTagsException:
				blog.Errorf("CreateLoadBalancer failed:%v, %s", elbv2.ErrCodeTooManyTagsException, awsErr.Error())
			case elbv2.ErrCodeDuplicateTagKeysException:
				blog.Errorf("CreateLoadBalancer failed:%v, %s", elbv2.ErrCodeDuplicateTagKeysException, awsErr.Error())
			case elbv2.ErrCodeResourceInUseException:
				blog.Errorf("CreateLoadBalancer failed:%v, %s", elbv2.ErrCodeResourceInUseException, awsErr.Error())
			case elbv2.ErrCodeAllocationIdNotFoundException:
				blog.Errorf("CreateLoadBalancer failed:%v, %s", elbv2.ErrCodeAllocationIdNotFoundException, awsErr.Error())
			case elbv2.ErrCodeAvailabilityZoneNotSupportedException:
				blog.Errorf("CreateLoadBalancer failed:%v, %s", elbv2.ErrCodeAvailabilityZoneNotSupportedException, awsErr.Error())
			default:
				blog.Errorf("CreateLoadBalancer failed: %s", awsErr.Error())
			}
		} else {
			blog.Errorf("CreateLoadBalancer cast err to awserr.Error failed %s", awsErr.Error())
		}
		return "", err
	}
	blog.Infof("CreateLoadBalancer response: %v", respCreateLB)

	if len(respCreateLB.LoadBalancers) == 0 {
		blog.Infof("no lb information in response")
		return "", fmt.Errorf("no lb information in response")
	}

	lbArn := respCreateLB.LoadBalancers[0].LoadBalancerArn

	return aws.StringValue(lbArn), nil
}

//awsDescribeLoadBalancer query loadbalancer instance information with certain name
func (a *awsElbSdkAPI) awsDescribeLoadBalancer(nameDescribe string) (*loadbalance.CloudLoadBalancer, bool, error) {
	input := &elbv2.DescribeLoadBalancersInput{
		Names: []*string{
			aws.String(nameDescribe),
		},
	}
	blog.Infof("DescribeLoadBalancer input: %v", input)
	svc := elbv2.New(session.New())
	result, err := svc.DescribeLoadBalancers(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeLoadBalancerNotFoundException:
				blog.Errorf("DescribeLoadBalancer failed: %v, %s", elbv2.ErrCodeLoadBalancerNotFoundException, aerr.Error())
				return nil, false, nil
			default:
				blog.Errorf("DescribeLoadBalancer failed: %s", aerr.Error())
			}
		} else {
			blog.Errorf("DescribeLoadBalancer: %s", err.Error())
		}
		return nil, false, err
	}
	blog.Infof("DescribeLoadBalancer result: %v", result)

	if len(result.LoadBalancers) == 0 {
		blog.Errorf("no lb information in result")
		return nil, false, fmt.Errorf("no lb information in result")
	}

	awslb := result.LoadBalancers[0]
	if err != nil {
		blog.Errorf("lbTypeAws2Bcs error, %s", err.Error())
		return nil, true, fmt.Errorf("lbTypeAws2Bcs error, %s", err.Error())
	}
	networkType, err := networkTypeAws2Bcs(aws.StringValue(awslb.Scheme))
	if err != nil {
		blog.Errorf("networkTypeAws2Bcs error, %s", err.Error())
		return nil, true, fmt.Errorf("networkTypeAws2Bcs error, %s", err.Error())
	}

	lb := &loadbalance.CloudLoadBalancer{
		ID:          aws.StringValue(awslb.LoadBalancerArn),
		NetworkType: networkType,
		Name:        aws.StringValue(awslb.LoadBalancerName),
	}
	return lb, true, nil
}

//awsCreateListener use aws api to create listener
//protocol field must be one of [TCP, HTTP, HTTPS, TLS](uppercase)
//it is unneccessary to spcify listener name
//return created listener arn
func (a *awsElbSdkAPI) awsCreateListener(targetGroup, loadBalancerArn, protocol string, port int64) (string, error) {
	protocol, err := protocolTypeBcs2Aws(protocol)
	if err != nil {
		return "", fmt.Errorf("protocolTypeBcs2Aws failed, %s", err.Error())
	}
	awsClient := elbv2.New(session.New())
	reqCreateListener := &elbv2.CreateListenerInput{
		DefaultActions: []*elbv2.Action{
			{
				TargetGroupArn: aws.String(targetGroup),
				Type:           aws.String("forward"),
			},
		},
		LoadBalancerArn: aws.String(loadBalancerArn),
		Port:            aws.Int64(port),
		Protocol:        aws.String(strings.ToUpper(protocol)),
	}
	blog.Infof("CreateListenerInput: %v", reqCreateListener)

	respCreateListener, err := awsClient.CreateListener(reqCreateListener)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeDuplicateListenerException:
				blog.Errorf("CreateListener failed:%v, %s", elbv2.ErrCodeDuplicateListenerException, aerr.Error())
			case elbv2.ErrCodeTooManyListenersException:
				blog.Errorf("CreateListener failed:%v, %s", elbv2.ErrCodeTooManyListenersException, aerr.Error())
			case elbv2.ErrCodeTooManyCertificatesException:
				blog.Errorf("CreateListener failed:%v, %s", elbv2.ErrCodeTooManyCertificatesException, aerr.Error())
			case elbv2.ErrCodeLoadBalancerNotFoundException:
				blog.Errorf("CreateListener failed:%v, %s", elbv2.ErrCodeLoadBalancerNotFoundException, aerr.Error())
			case elbv2.ErrCodeTargetGroupNotFoundException:
				blog.Errorf("CreateListener failed:%v, %s", elbv2.ErrCodeTargetGroupNotFoundException, aerr.Error())
			case elbv2.ErrCodeTargetGroupAssociationLimitException:
				blog.Errorf("CreateListener failed:%v, %s", elbv2.ErrCodeTargetGroupAssociationLimitException, aerr.Error())
			case elbv2.ErrCodeInvalidConfigurationRequestException:
				blog.Errorf("CreateListener failed:%v, %s", elbv2.ErrCodeInvalidConfigurationRequestException, aerr.Error())
			case elbv2.ErrCodeIncompatibleProtocolsException:
				blog.Errorf("CreateListener failed:%v, %s", elbv2.ErrCodeIncompatibleProtocolsException, aerr.Error())
			case elbv2.ErrCodeSSLPolicyNotFoundException:
				blog.Errorf("CreateListener failed:%v, %s", elbv2.ErrCodeSSLPolicyNotFoundException, aerr.Error())
			case elbv2.ErrCodeCertificateNotFoundException:
				blog.Errorf("CreateListener failed:%v, %s", elbv2.ErrCodeCertificateNotFoundException, aerr.Error())
			case elbv2.ErrCodeUnsupportedProtocolException:
				blog.Errorf("CreateListener failed:%v, %s", elbv2.ErrCodeUnsupportedProtocolException, aerr.Error())
			case elbv2.ErrCodeTooManyRegistrationsForTargetIdException:
				blog.Errorf("CreateListener failed:%v, %s", elbv2.ErrCodeTooManyRegistrationsForTargetIdException, aerr.Error())
			case elbv2.ErrCodeTooManyTargetsException:
				blog.Errorf("CreateListener failed:%v, %s", elbv2.ErrCodeTooManyTargetsException, aerr.Error())
			default:
				blog.Errorf("CreateListener failed:%s", aerr.Error())
			}
		} else {
			blog.Errorf("CreateListener cast err to awserr.Error failed %s", err.Error())
		}
		return "", err
	}

	blog.Infof("CreateListener response: %v", respCreateListener)
	listenerArn := aws.StringValue(respCreateListener.Listeners[0].ListenerArn)
	return listenerArn, nil
}

//awsDeleteListener delete listener
func (a *awsElbSdkAPI) awsDeleteListener(listenerArn string) error {
	input := &elbv2.DeleteListenerInput{
		ListenerArn: aws.String(listenerArn),
	}
	svc := elbv2.New(session.New())
	result, err := svc.DeleteListener(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeTargetGroupNotFoundException:
				blog.Errorf("DeleteListener failed:%v, %s", elbv2.ErrCodeTargetGroupNotFoundException, aerr.Error())
			case elbv2.ErrCodeTooManyTargetsException:
				blog.Errorf("DeleteListener failed:%v, %s", elbv2.ErrCodeTooManyTargetsException, aerr.Error())
			case elbv2.ErrCodeInvalidTargetException:
				blog.Errorf("DeleteListener failed:%v, %s", elbv2.ErrCodeInvalidTargetException, aerr.Error())
			case elbv2.ErrCodeTooManyRegistrationsForTargetIdException:
				blog.Errorf("DeleteListener failed:%v, %s", elbv2.ErrCodeTooManyRegistrationsForTargetIdException, aerr.Error())
			default:
				blog.Errorf("DeleteListener failed:%s", aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			blog.Errorf("DeleteListener:%s", err.Error())
		}
		return err
	}

	blog.Infof("DeleteListener result %v", result)
	return nil
}

//awsUpdateRule update listener port
func (a *awsElbSdkAPI) awsUpdateListener(listenerID string, port int64) error {
	svc := elbv2.New(session.New())
	input := &elbv2.ModifyListenerInput{
		ListenerArn: aws.String(listenerID),
		Port:        aws.Int64(port),
	}
	blog.Infof("awsUpdateListener input: %v", input)
	result, err := svc.ModifyListener(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeDuplicateListenerException:
				blog.Errorf("%v %s", elbv2.ErrCodeDuplicateListenerException, aerr.Error())
			case elbv2.ErrCodeTooManyListenersException:
				blog.Errorf("%v %s", elbv2.ErrCodeTooManyListenersException, aerr.Error())
			case elbv2.ErrCodeTooManyCertificatesException:
				blog.Errorf("%v %s", elbv2.ErrCodeTooManyCertificatesException, aerr.Error())
			case elbv2.ErrCodeListenerNotFoundException:
				blog.Errorf("%v %s", elbv2.ErrCodeListenerNotFoundException, aerr.Error())
			case elbv2.ErrCodeTargetGroupNotFoundException:
				blog.Errorf("%v %s", elbv2.ErrCodeTargetGroupNotFoundException, aerr.Error())
			case elbv2.ErrCodeTargetGroupAssociationLimitException:
				blog.Errorf("%v %s", elbv2.ErrCodeTargetGroupAssociationLimitException, aerr.Error())
			case elbv2.ErrCodeIncompatibleProtocolsException:
				blog.Errorf("%v %s", elbv2.ErrCodeIncompatibleProtocolsException, aerr.Error())
			case elbv2.ErrCodeSSLPolicyNotFoundException:
				blog.Errorf("%v %s", elbv2.ErrCodeSSLPolicyNotFoundException, aerr.Error())
			case elbv2.ErrCodeCertificateNotFoundException:
				blog.Errorf("%v %s", elbv2.ErrCodeCertificateNotFoundException, aerr.Error())
			case elbv2.ErrCodeInvalidConfigurationRequestException:
				blog.Errorf("%v %s", elbv2.ErrCodeInvalidConfigurationRequestException, aerr.Error())
			case elbv2.ErrCodeUnsupportedProtocolException:
				blog.Errorf("%v %s", elbv2.ErrCodeUnsupportedProtocolException, aerr.Error())
			case elbv2.ErrCodeTooManyRegistrationsForTargetIdException:
				blog.Errorf("%v %s", elbv2.ErrCodeTooManyRegistrationsForTargetIdException, aerr.Error())
			case elbv2.ErrCodeTooManyTargetsException:
				blog.Errorf("%v %s", elbv2.ErrCodeTooManyTargetsException, aerr.Error())
			default:
				blog.Errorf("%s", aerr.Error())
			}
		} else {
			blog.Errorf("%s", err.Error())
		}
		return err
	}

	blog.Infof("awsUpdateListener result: %v", result)
	return nil
}

//awsDescribeListener describe a listener
func (a *awsElbSdkAPI) awsDescribeListener(lbArn, listenerArn string) (*loadbalance.CloudListener, bool, error) {
	svc := elbv2.New(session.New())
	input := &elbv2.DescribeListenersInput{
		ListenerArns: []*string{
			aws.String(listenerArn),
		},
		LoadBalancerArn: aws.String(lbArn),
	}
	blog.Infof("DescribeListeners input: %v", input)
	result, err := svc.DescribeListeners(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeListenerNotFoundException:
				//if no found, do not return error
				blog.Errorf("%v %s", elbv2.ErrCodeListenerNotFoundException, aerr.Error())
				return nil, false, nil
			case elbv2.ErrCodeLoadBalancerNotFoundException:
				blog.Errorf("%v %s", elbv2.ErrCodeLoadBalancerNotFoundException, aerr.Error())
			case elbv2.ErrCodeUnsupportedProtocolException:
				blog.Errorf("%v %s", elbv2.ErrCodeUnsupportedProtocolException, aerr.Error())
			default:
				blog.Errorf("%s", aerr.Error())
			}
		} else {
			blog.Errorf("%s", err.Error())
		}
		return nil, false, err
	}
	blog.Infof("DescribeListeners result: %v", result)
	if len(result.Listeners) == 0 {
		blog.Errorf("DescribeListeners return no listener info")
		return nil, false, fmt.Errorf("DescribeListeners return no listener info")
	}

	listenerOutput := result.Listeners[0]
	listener := &loadbalance.CloudListener{
		Spec: loadbalance.CloudListenerSpec{
			ListenerID:     aws.StringValue(listenerOutput.ListenerArn),
			LoadBalancerID: aws.StringValue(listenerOutput.LoadBalancerArn),
			Protocol:       aws.StringValue(listenerOutput.Protocol),
			ListenPort:     int(aws.Int64Value(listenerOutput.Port)),
		},
	}

	return listener, true, nil
}

//AwsCreateRule create a rule in a listener for a target group
//**priority** is neccessary, a listener can't have muliple rules with the same priority
func (a *awsElbSdkAPI) awsCreateRule(targetGroup, listenerArn, domain, path string, priority int64) (string, error) {
	svc := elbv2.New(session.New())
	input := &elbv2.CreateRuleInput{
		Actions: []*elbv2.Action{
			{
				TargetGroupArn: aws.String(targetGroup),
				Type:           aws.String("forward"),
			},
		},
		Conditions: []*elbv2.RuleCondition{
			{
				Field: aws.String("host-header"),
				Values: []*string{
					aws.String(domain),
				},
			},
			{
				Field: aws.String("path-pattern"),
				Values: []*string{
					aws.String(path),
				},
			},
		},
		ListenerArn: aws.String(listenerArn),
		Priority:    aws.Int64(priority),
	}
	blog.Infof("CreateRuleInput: %v", input)
	result, err := svc.CreateRule(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodePriorityInUseException:
				blog.Errorf("CreateRule failed: %v, %s", elbv2.ErrCodePriorityInUseException, aerr.Error())
			case elbv2.ErrCodeTooManyTargetGroupsException:
				blog.Errorf("CreateRule failed: %v, %s", elbv2.ErrCodeTooManyTargetGroupsException, aerr.Error())
			case elbv2.ErrCodeTooManyRulesException:
				blog.Errorf("CreateRule failed: %v, %s", elbv2.ErrCodeTooManyRulesException, aerr.Error())
			case elbv2.ErrCodeTargetGroupAssociationLimitException:
				blog.Errorf("CreateRule failed: %v, %s", elbv2.ErrCodeTargetGroupAssociationLimitException, aerr.Error())
			case elbv2.ErrCodeIncompatibleProtocolsException:
				blog.Errorf("CreateRule failed: %v, %s", elbv2.ErrCodeIncompatibleProtocolsException, aerr.Error())
			case elbv2.ErrCodeListenerNotFoundException:
				blog.Errorf("CreateRule failed: %v, %s", elbv2.ErrCodeListenerNotFoundException, aerr.Error())
			case elbv2.ErrCodeTargetGroupNotFoundException:
				blog.Errorf("CreateRule failed: %v, %s", elbv2.ErrCodeTargetGroupNotFoundException, aerr.Error())
			case elbv2.ErrCodeInvalidConfigurationRequestException:
				blog.Errorf("CreateRule failed: %v, %s", elbv2.ErrCodeInvalidConfigurationRequestException, aerr.Error())
			case elbv2.ErrCodeTooManyRegistrationsForTargetIdException:
				blog.Errorf("CreateRule failed: %v, %s", elbv2.ErrCodeTooManyRegistrationsForTargetIdException, aerr.Error())
			case elbv2.ErrCodeTooManyTargetsException:
				blog.Errorf("CreateRule failed: %v, %s", elbv2.ErrCodeTooManyTargetsException, aerr.Error())
			default:
				blog.Errorf("CreateRule failed: %s", aerr.Error())
			}
		} else {
			blog.Errorf("CreateRule failed: %s", err.Error())
		}
		return "", err
	}

	if len(result.Rules) == 0 {
		blog.Errorf("CreateRule result invalid %v", result)
		return "", fmt.Errorf("CreateRule result invalid %v", result)
	}

	blog.Infof("CreateRule result: %v", result)
	ruleArn := aws.StringValue(result.Rules[0].RuleArn)
	return ruleArn, nil
}

//AwsUpdateRule update rule
func (a *awsElbSdkAPI) awsUpdateRule(ruleArn, targetGroup, domain, path string) error {
	svc := elbv2.New(session.New())
	input := &elbv2.ModifyRuleInput{
		Actions: []*elbv2.Action{
			{
				TargetGroupArn: aws.String(targetGroup),
				Type:           aws.String("forward"),
			},
		},
		Conditions: []*elbv2.RuleCondition{
			{
				Field: aws.String("host-header"),
				Values: []*string{
					aws.String(domain),
				},
			},
			{
				Field: aws.String("path-pattern"),
				Values: []*string{
					aws.String(path),
				},
			},
		},
		RuleArn: aws.String(ruleArn),
	}
	blog.Infof("UpdateRuleInput: %v", input)
	result, err := svc.ModifyRule(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeIncompatibleProtocolsException:
				blog.Errorf("UpdateRule failed: %v, %s", elbv2.ErrCodeIncompatibleProtocolsException, aerr.Error())
			case elbv2.ErrCodeOperationNotPermittedException:
				blog.Errorf("UpdateRule failed: %v, %s", elbv2.ErrCodeOperationNotPermittedException, aerr.Error())
			case elbv2.ErrCodeRuleNotFoundException:
				blog.Errorf("UpdateRule failed: %v, %s", elbv2.ErrCodeRuleNotFoundException, aerr.Error())
			case elbv2.ErrCodeTargetGroupAssociationLimitException:
				blog.Errorf("UpdateRule failed: %v, %s", elbv2.ErrCodeTargetGroupAssociationLimitException, aerr.Error())
			case elbv2.ErrCodeTargetGroupNotFoundException:
				blog.Errorf("UpdateRule failed: %v, %s", elbv2.ErrCodeTargetGroupNotFoundException, aerr.Error())
			case elbv2.ErrCodeTooManyRegistrationsForTargetIdException:
				blog.Errorf("UpdateRule failed: %v, %s", elbv2.ErrCodeTooManyRegistrationsForTargetIdException, aerr.Error())
			case elbv2.ErrCodeTooManyTargetsException:
				blog.Errorf("UpdateRule failed: %v, %s", elbv2.ErrCodeTooManyTargetsException, aerr.Error())
			case elbv2.ErrCodeUnsupportedProtocolException:
				blog.Errorf("UpdateRule failed: %v, %s", elbv2.ErrCodeUnsupportedProtocolException, aerr.Error())
			default:
				blog.Errorf("UpdateRule failed: %s", aerr.Error())
			}
		} else {
			blog.Errorf("UpdateRule failed: %s", err.Error())
		}
		return err
	}

	if len(result.Rules) == 0 {
		blog.Errorf("UpdateRule result invalid %v", result)
		return fmt.Errorf("UpdateRule result invalid %v", result)
	}
	blog.Infof("UpdateRule result: %v", result)
	return nil
}

//AwsDeleteRule detele rule by arn
func (a *awsElbSdkAPI) awsDeleteRule(ruleArn string) error {
	svc := elbv2.New(session.New())
	input := &elbv2.DeleteRuleInput{
		RuleArn: aws.String(ruleArn),
	}
	blog.Infof("DeleteRule input: %v", input)

	result, err := svc.DeleteRule(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeOperationNotPermittedException:
				blog.Errorf("DeleteRule failed: %v, %s", elbv2.ErrCodeOperationNotPermittedException, aerr.Error())
			case elbv2.ErrCodeRuleNotFoundException:
				blog.Errorf("DeleteRule failed: %v, %s", elbv2.ErrCodeRuleNotFoundException, aerr.Error())
			default:
				blog.Errorf("DeleteRule failed: %s", aerr.Error())
			}
			return err
		}
	}
	blog.Infof("DeleteRule result: %v", result)
	return nil
}

//awsDescribeRule query rule data
func (a *awsElbSdkAPI) awsDescribeRule(listenerArn, ruleArn string) (*loadbalance.Rule, bool, error) {
	svc := elbv2.New(session.New())
	input := &elbv2.DescribeRulesInput{
		ListenerArn: aws.String(listenerArn),
		RuleArns: []*string{
			aws.String(ruleArn),
		},
	}
	blog.Infof("DescribeRules input: %v", input)
	result, err := svc.DescribeRules(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeListenerNotFoundException:
				blog.Errorf("%v %s", elbv2.ErrCodeListenerNotFoundException, aerr.Error())
			case elbv2.ErrCodeRuleNotFoundException:
				blog.Errorf("%v %s", elbv2.ErrCodeRuleNotFoundException, aerr.Error())
				return nil, false, nil
			case elbv2.ErrCodeUnsupportedProtocolException:
				blog.Errorf("%v %s", elbv2.ErrCodeUnsupportedProtocolException, aerr.Error())
			default:
				blog.Errorf("%s", aerr.Error())
			}
		} else {
			blog.Errorf("%s", err.Error())
		}
		return nil, false, err
	}
	blog.Infof("DescribeRules result: %v", result)
	if len(result.Rules) == 0 {
		blog.Errorf("DescribeRules return no rule info")
		return nil, false, fmt.Errorf("DescribeRules return no rule info")
	}

	ruleOutput := result.Rules[0]
	//HACK: it is not elegant to get rule's target group, url and domain
	domain := ""
	url := ""
	if len(ruleOutput.Conditions) == 2 {
		if len(ruleOutput.Conditions[0].Values) != 0 {
			domain = aws.StringValue(ruleOutput.Conditions[0].Values[0])
		}
		if len(ruleOutput.Conditions[1].Values) != 0 {
			url = aws.StringValue(ruleOutput.Conditions[1].Values[0])
		}
	}
	rule := &loadbalance.Rule{
		ID:     aws.StringValue(ruleOutput.RuleArn),
		Domain: domain,
		URL:    url,
	}

	return rule, true, nil
}

//AwsCreateTargetGroup use aws api to create target group
func (a *awsElbSdkAPI) awsCreateTargetGroup(targetGroupName, vpcID, protocol, healthCheckPath string,
	targetGroupPort int64) (string, error) {
	//translate protocol type
	protocol, err := protocolTypeBcs2Aws(protocol)
	if err != nil {
		return "", fmt.Errorf("protocolTypeBcs2Aws failed, %s", err.Error())
	}

	svc := elbv2.New(session.New())
	input := &elbv2.CreateTargetGroupInput{
		Name:       aws.String(targetGroupName),
		Port:       aws.Int64(targetGroupPort),
		Protocol:   aws.String(protocol),
		VpcId:      aws.String(vpcID),
		TargetType: aws.String(elbv2.TargetTypeEnumIp),
	}
	if strings.ToUpper(protocol) == AWS_LOADBALANCE_PROTOCOL_HTTP || strings.ToUpper(protocol) == AWS_LOADBALANCE_PROTOCOL_HTTPS {
		input.HealthCheckPath = aws.String(healthCheckPath)
	}
	blog.Infof("CreateTargetGroupInput: %v", input)

	result, err := svc.CreateTargetGroup(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeDuplicateTargetGroupNameException:
				blog.Errorf("CreateTargetGroup failed: %v, %s", elbv2.ErrCodeDuplicateTargetGroupNameException, aerr.Error())
			case elbv2.ErrCodeTooManyTargetGroupsException:
				blog.Errorf("CreateTargetGroup failed: %v, %s", elbv2.ErrCodeTooManyTargetGroupsException, aerr.Error())
			case elbv2.ErrCodeInvalidConfigurationRequestException:
				blog.Errorf("CreateTargetGroup failed: %v, %s", elbv2.ErrCodeInvalidConfigurationRequestException, aerr.Error())
			default:
				blog.Errorf("CreateTargetGroup failed: %s", aerr.Error())
			}
		} else {
			blog.Errorf("CreateTargetGroup:%s", err.Error())
		}
		return "", err
	}

	blog.Infof("CreateTargetGroup result: %v", result)
	if len(result.TargetGroups) == 0 {
		blog.Errorf("no target group in result")
		return "", fmt.Errorf("no target group in result")
	}

	targetGroupArn := aws.StringValue(result.TargetGroups[0].TargetGroupArn)
	return targetGroupArn, nil
}

//AwsDeleteTargetGroup use aws api delete target group
func (a *awsElbSdkAPI) awsDeleteTargetGroup(targetGroupArn string) error {
	input := &elbv2.DeleteTargetGroupInput{
		TargetGroupArn: aws.String(targetGroupArn),
	}
	svc := elbv2.New(session.New())
	result, err := svc.DeleteTargetGroup(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeTargetGroupNotFoundException:
				blog.Errorf("DeleteTargetGroup failed: %v, %s", elbv2.ErrCodeTargetGroupNotFoundException, aerr.Error())
			case elbv2.ErrCodeTooManyTargetsException:
				blog.Errorf("DeleteTargetGroup failed: %v, %s", elbv2.ErrCodeTooManyTargetsException, aerr.Error())
			case elbv2.ErrCodeInvalidTargetException:
				blog.Errorf("DeleteTargetGroup failed: %v, %s", elbv2.ErrCodeInvalidTargetException, aerr.Error())
			case elbv2.ErrCodeTooManyRegistrationsForTargetIdException:
				blog.Errorf("DeleteTargetGroup failed: %v, %s", elbv2.ErrCodeTooManyRegistrationsForTargetIdException, aerr.Error())
			default:
				blog.Errorf("DeleteTargetGroup failed: %s", aerr.Error())
			}
		} else {
			blog.Errorf("DeleteTargetGroup: %s", err.Error())
		}
		return err
	}

	blog.Infof("DeRegisterTargets result %v", result)
	return nil
}

//AwsRegisterTargets register backend instance into target group
func (a *awsElbSdkAPI) awsRegisterTargets(targetGroupArn string, backends loadbalance.BackendList) error {
	var targetDescriptions []*elbv2.TargetDescription
	for _, backend := range backends {
		t := &elbv2.TargetDescription{
			Id:   aws.String(backend.IP),
			Port: aws.Int64(int64(backend.Port)),
		}
		targetDescriptions = append(targetDescriptions, t)
	}
	svc := elbv2.New(session.New())
	input := &elbv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(targetGroupArn),
		Targets:        targetDescriptions,
	}
	blog.Infof("RegisterTargetsInput: %v", input)
	result, err := svc.RegisterTargets(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeTargetGroupNotFoundException:
				blog.Errorf("RegisterTargets failed: %v, %s", elbv2.ErrCodeTargetGroupNotFoundException, aerr.Error())
			case elbv2.ErrCodeTooManyTargetsException:
				blog.Errorf("RegisterTargets failed: %v, %s", elbv2.ErrCodeTooManyTargetsException, aerr.Error())
			case elbv2.ErrCodeInvalidTargetException:
				blog.Errorf("RegisterTargets failed: %v, %s", elbv2.ErrCodeInvalidTargetException, aerr.Error())
			case elbv2.ErrCodeTooManyRegistrationsForTargetIdException:
				blog.Errorf("RegisterTargets failed: %v, %s", elbv2.ErrCodeTooManyRegistrationsForTargetIdException, aerr.Error())
			default:
				blog.Errorf("RegisterTargets failed: %s", aerr.Error())
			}
		} else {
			blog.Errorf("RegisterTargets: %s", err.Error())
		}
		return err
	}
	blog.Infof("RegisterTargets result: %v", result)
	return nil
}

//AwsDeRegisterTargets deregister targets from target group
func (a *awsElbSdkAPI) awsDeRegisterTargets(targetGroupArn string, backends loadbalance.BackendList) error {
	//construct target descriptions data list
	var targetDescriptions []*elbv2.TargetDescription
	for _, backend := range backends {
		newTargetDescription := &elbv2.TargetDescription{
			Id:   aws.String(backend.IP),
			Port: aws.Int64(int64(backend.Port)),
		}
		targetDescriptions = append(targetDescriptions, newTargetDescription)
	}

	//request aws
	svc := elbv2.New(session.New())
	input := &elbv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(targetGroupArn),
		Targets:        targetDescriptions,
	}
	blog.Infof("DeregisterTargetsInput: %v", input)
	result, err := svc.DeregisterTargets(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeTargetGroupNotFoundException:
				blog.Errorf("DeregisterTargets failed: %v, %s", elbv2.ErrCodeTargetGroupNotFoundException, aerr.Error())
			case elbv2.ErrCodeTooManyTargetsException:
				blog.Errorf("DeregisterTargets failed: %v, %s", elbv2.ErrCodeTooManyTargetsException, aerr.Error())
			case elbv2.ErrCodeInvalidTargetException:
				blog.Errorf("DeregisterTargets failed: %v, %s", elbv2.ErrCodeInvalidTargetException, aerr.Error())
			case elbv2.ErrCodeTooManyRegistrationsForTargetIdException:
				blog.Errorf("DeregisterTargets failed: %v, %s", elbv2.ErrCodeTooManyRegistrationsForTargetIdException, aerr.Error())
			default:
				blog.Errorf("DeregisterTargets failed: %s", aerr.Error())
			}
		} else {
			blog.Errorf("DeregisterTargets:%s", err.Error())
		}
		return err
	}

	blog.Infof("DeregisterTargets result: %v", result)
	return nil
}

//AwsAuthorizeSecurityGroupIngress add ingress port permissions to security group
//support udp 2018/02/27
func (a *awsElbSdkAPI) awsAuthorizeSecurityGroupIngress(securityGroupID *string, startPort, endPort int64) error {
	svc := ec2.New(session.New())
	// Add permissions to the security group
	rst, err := svc.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: securityGroupID,
		IpPermissions: []*ec2.IpPermission{
			// Can use setters to simplify seting multiple values without the
			// needing to use aws.String or associated helper utilities.
			(&ec2.IpPermission{}).
				SetIpProtocol("tcp").
				SetFromPort(startPort).
				SetToPort(endPort).
				SetIpRanges([]*ec2.IpRange{
					{CidrIp: aws.String("0.0.0.0/0")},
				}),
			//add at 2018/02/27
			(&ec2.IpPermission{}).
				SetIpProtocol("udp").
				SetFromPort(startPort).
				SetToPort(endPort).
				SetIpRanges([]*ec2.IpRange{
					{CidrIp: aws.String("0.0.0.0/0")},
				}),
		},
	})
	if err != nil {
		blog.Errorf("AuthorizeSecurityGroupIngress failed:%s", err.Error())
		return err
	}
	blog.Infof("AuthorizeSecurityGroupIngress result:%v", rst)
	return nil
}

//AwsModifyInstanceSecurityGroups modify backend instance security group
func (a *awsElbSdkAPI) awsModifyInstanceSecurityGroups(securityGroupID, hostIP string) error {
	//describe all relevant instances
	svc := ec2.New(session.New())
	output, err := svc.DescribeInstances(nil)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeRuleNotFoundException:
				blog.Errorf("DescribeInstances failed %v, %s", elbv2.ErrCodeRuleNotFoundException, aerr.Error())
			case elbv2.ErrCodeOperationNotPermittedException:
				blog.Errorf("DescribeInstances failed %v, %s", elbv2.ErrCodeOperationNotPermittedException, aerr.Error())
			default:
				blog.Errorf("DescribeInstances failed %s", aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			blog.Errorf("DescribeInstances failed %s", err.Error())
		}
		return err
	}
	//find instance with certain privateIp and modify its security group
	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			if string(*instance.PrivateIpAddress) == hostIP {
				blog.Infof("modify instance: %v", instance)
				var groups []*string
				for _, relatedSecurityGroup := range instance.SecurityGroups {
					groups = append(groups, relatedSecurityGroup.GroupId)
				}
				input := &ec2.ModifyInstanceAttributeInput{
					InstanceId: instance.InstanceId,
					Groups:     groups,
				}
				out, err := svc.ModifyInstanceAttribute(input)
				if err != nil {
					blog.Errorf("ModifyInstanceAttribute failed: %v", err)
					return err
				}
				blog.Infof("ModifyInstanceAttribute success: %v", out)
				return nil
			}
		}
	}

	return nil
}

//AwsDescribeSecurityGroup
func (a *awsElbSdkAPI) awsDescribeSecurityGroup(securityGroupID string) (*ec2.DescribeSecurityGroupsOutput, error) {
	svc := ec2.New(session.New())
	//call describe security group
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: []*string{
			aws.String(securityGroupID),
		},
	}
	blog.Infof("AwsDescribeSecurityGroup input: %v", input)

	result, err := svc.DescribeSecurityGroups(input)
	if err != nil {
		blog.Errorf("AwsDescribeSecurityGroup failed: %s", err.Error())
		return result, err
	}
	blog.Infof("AwsDescribeSecurityGroup result: %v", result)
	if len(result.SecurityGroups) == 0 {
		blog.Errorf("AwsDescribeSecurityGroup failed, no security group in result")
		return result, fmt.Errorf("AwsDescribeSecurityGroup failed, no security group in result")
	}

	return result, nil
}
