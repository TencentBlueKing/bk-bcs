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

	"fmt"

	loadbalance "bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/network/v1"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/cloudlb"
)

//ElbClient client to manage aws elb
type ElbClient struct {
	awsAPI awsElbInterface
}

//NewElbClient construct function
func NewElbClient() cloudlb.Interface {
	return &ElbClient{
		awsAPI: NewAwsElbSdkAPI(),
	}
}

//LoadConfig implements cloud infrastructure interface
func (aws *ElbClient) LoadConfig() error {
	//TODO:
	return nil
}

//CreateLoadbalance implements cloud interface
//create a loadbalance
func (aws *ElbClient) CreateLoadbalance() (*loadbalance.CloudLoadBalancer, error) {
	//TODO:
	lbName := ""

	lbInfo, err := aws.DescribeLoadbalance(lbName)
	if err != nil {
		return nil, fmt.Errorf("DescribeLoadbalance failed, %s", err.Error())
	}

	//TODO:
	//call encapsulated api
	lbArn, err := aws.awsAPI.awsCreateLoadBalancer(nil, nil, lbName, "", "")
	if err != nil {
		return nil, err
	}
	lbInfo = &loadbalance.CloudLoadBalancer{
		ID: lbArn,
	}
	return lbInfo, err
}

//DescribeLoadbalance implements cloud infrastructure interface
func (aws *ElbClient) DescribeLoadbalance(name string) (*loadbalance.CloudLoadBalancer, error) {
	//validate name
	if len(name) == 0 {
		blog.Errorf("name is empty")
		return nil, fmt.Errorf("name is empty")
	}
	//do query
	lb, isExisted, err := aws.awsAPI.awsDescribeLoadBalancer(name)
	if err != nil {
		blog.Errorf("DescribeLoadbalance err: %s", err.Error())
		return nil, err
	}
	if !isExisted {
		blog.Infof("LoadBalancer %s no existed", name)
		return nil, nil
	}
	return lb, nil
}

//OnUpdate implements cloud infrastructure interface
//**cur** pointer to listener data, when success, will be added to cache directly
//if listener is not existed, we will create a new listener
func (aws *ElbClient) Update(old, cur *loadbalance.CloudListener) error {
	//describe listener to judge if listener existed
	_, isExisted, err := aws.awsAPI.awsDescribeListener(cur.Spec.LoadBalancerID, cur.Spec.ListenerID)
	if err != nil {
		return fmt.Errorf("awsDescribeListener failed, %s", err.Error())
	}
	//if not existed, we will create a new one
	if !isExisted {
		blog.Warnf("listener %s name %s does not exist, try to create a new listener", old.Spec.LoadBalancerID, old.GetName())
		err := aws.Add(cur)
		if err != nil {
			return fmt.Errorf("OnAdd failed, %s", err.Error())
		}
		return nil
	}
	//if cur equals to old, no need to update
	if old.IsEqual(cur) {
		blog.Infof("no need to update current listener is equal to old listener")
		return nil
	}
	//tcp listener use
	if old.Spec.Protocol == loadbalance.ClbListenerProtocolTCP {
		err := aws.update4LayerListener(old, cur)
		if err != nil {
			blog.Errorf("update 4 layer listener failed, %s", err)
			return fmt.Errorf("update 4 layer listener failed, %s", err)
		}
	} else if old.Spec.Protocol == loadbalance.ClbListenerProtocolHTTP || old.Spec.Protocol == loadbalance.ClbListenerProtocolHTTPS {
		err := aws.update7LayerListener(old, cur)
		if err != nil {
			blog.Errorf("update 7 layer listener failed, %s", err)
			return fmt.Errorf("update 7 layer listener failed, %s", err)
		}
	} else {
		blog.Errorf("error listener protocol %s", old.Spec.Protocol)
		return fmt.Errorf("error listener protocol %s", old.Spec.Protocol)
	}

	return nil
}

//OnAdd implements cloud interface
//when create listener, default action with targetgroup is required
func (aws *ElbClient) Add(ls *loadbalance.CloudListener) error {

	//default target group is necessary
	if ls.Spec.TargetGroup == nil {
		return fmt.Errorf("OnAdd error, default target group is necessary for aws lb protocol")
	}

	//create default target group
	targetGroupArn, err := aws.awsAPI.awsCreateTargetGroup(ls.Spec.TargetGroup.Name, "",
		ls.Spec.Protocol, ls.Spec.TargetGroup.HealthCheck.HTTPCheckPath, int64(ls.Spec.TargetGroup.Port))
	if err != nil {
		return fmt.Errorf("OnAdd create default target group failed, %s", err.Error())
	}
	//set default target group id
	ls.Spec.TargetGroup.ID = targetGroupArn

	//create listener
	//**Default Action** just default Action, no condition
	//when http or https request satisfied none of conditions, then default action will execute
	//we use the default action for 4 layer protocol listener
	listenerArn, err := aws.awsAPI.awsCreateListener(ls.Spec.TargetGroup.ID, ls.Spec.LoadBalancerID, ls.Spec.Protocol, int64(ls.Spec.ListenPort))
	if err != nil {
		blog.Errorf("create listener failed, %s", err.Error())
		return err
	}
	//set listener id
	ls.Spec.ListenerID = listenerArn

	//tcp protocol listener use default target to register backends
	if ls.Spec.Protocol == loadbalance.ClbListenerProtocolTCP {
		//register backend to default target group
		err := aws.awsAPI.awsRegisterTargets(ls.Spec.TargetGroup.ID, ls.Spec.TargetGroup.Backends)
		if err != nil {
			return fmt.Errorf("OnAdd call AwsRegisterTargets failed, %s", err.Error())
		}

	} else if ls.Spec.Protocol == loadbalance.ClbListenerProtocolHTTP || ls.Spec.Protocol == loadbalance.ClbListenerProtocolHTTPS {
		//create rule, create rule target group, register backends to rule target group
		err := aws.createRulesFor7LayerListener(ls)
		if err != nil {
			return fmt.Errorf("OnAdd call createRulesFor7LayerListener failed, %s", err.Error())
		}

	} else {
		return fmt.Errorf("OnAdd failed, unsupported listener protocol %s", ls.Spec.Protocol)
	}

	return nil
}

//OnDelete implements cloud interface
//when listener does not exist, just log error, do not return error
func (aws *ElbClient) Delete(ls *loadbalance.CloudListener) error {
	//check listener is existed or not
	_, isExisted, err := aws.awsAPI.awsDescribeListener(ls.Spec.LoadBalancerID, ls.Spec.ListenerID)
	if err != nil {
		return fmt.Errorf("awsDescribeListener failed, %s", err.Error())
	}
	if !isExisted {
		blog.Warnf("no need to delete, listener %s does not exist", ls.Spec.ListenerID)
		return nil
	}
	//delete it
	err = aws.awsAPI.awsDeleteListener(ls.Spec.ListenerID)
	if err != nil {
		return fmt.Errorf("awsDeleteListener failed, %s", err.Error())
	}
	return nil
}

// ListListeners list all listener on clb instance controlled by this controller
func (aws *ElbClient) ListListeners() ([]*loadbalance.CloudListener, error) {
	return nil, nil
}
