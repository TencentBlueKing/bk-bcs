/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tencentcloud

import (
	"errors"
	"fmt"
	"strings"
	"time"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/cloud"
)

// Clb client to operate clb instance
type Clb struct {
	sdkWrapper *SdkWrapper
	apiWrapper *APIWrapper
}

// NewClb create clb client
func NewClb() (*Clb, error) {
	sdkWrapper, err := NewSdkWrapper()
	if err != nil {
		return nil, err
	}
	apiWrapper, err := NewAPIWrapper()
	if err != nil {
		return nil, err
	}
	return &Clb{
		sdkWrapper: sdkWrapper,
		apiWrapper: apiWrapper,
	}, nil
}

// DescribeLoadBalancer get loadbalancer object by id
func (c *Clb) DescribeLoadBalancer(lbID string) (*cloud.LoadBalanceObject, error) {
	req := tclb.NewDescribeLoadBalancersRequest()
	req.LoadBalancerIds = tcommon.StringPtrs([]string{lbID})
	req.Forward = tcommon.Int64Ptr(1)

	ctime := time.Now()
	resp, err := c.sdkWrapper.DescribeLoadBalancers(req)
	if err != nil {
		blog.Errorf("DescribeLoadBalancers failed, err %s", err.Error())
		cloud.StatRequest("DescribeLoadBalancers", cloud.MetricAPIFailed, ctime, time.Now())
		return nil, fmt.Errorf("DescribeLoadBalancers failed, err %s", err.Error())
	}
	cloud.StatRequest("DescribeLoadBalancers", cloud.MetricAPISuccess, ctime, time.Now())

	if len(resp.Response.LoadBalancerSet) == 0 {
		return nil, cloud.ErrLoadbalancerNotFound
	}
	if len(resp.Response.LoadBalancerSet) > 1 {
		blog.Errorf("DescribeLoadBalancers response invalid, more than one lb, resp %s", resp.ToJsonString())
		return nil, fmt.Errorf("DescribeLoadBalancers response invalid, more than one lb, resp %s", resp.ToJsonString())
	}
	resplb := resp.Response.LoadBalancerSet[0]
	retlb := &cloud.LoadBalanceObject{}
	if resplb.LoadBalancerId != nil {
		retlb.LbID = *resplb.LoadBalancerId
	}
	if resplb.LoadBalancerType != nil {
		retlb.Type = *resplb.LoadBalancerType
	}
	if resplb.LoadBalancerName != nil {
		retlb.Name = *resplb.LoadBalancerName
	}
	retlb.IPs = tcommon.StringValues(resplb.LoadBalancerVips)
	return retlb, nil
}

// EnsureListener ensure listener to cloud, and get listener info
func (c *Clb) EnsureListener(listener *networkextensionv1.Listener) (string, error) {
	cloudListener, err := c.getListenerInfoByPort(listener.Spec.LoadbalancerID, listener.Spec.Port)
	if errors.Is(err, cloud.ErrListenerNotFound) {
		// to create listener
		listenerID, err := c.createListner(listener)
		if err != nil {
			return "", err
		}
		return listenerID, nil
	}

	blog.V(5).Infof("new listener %+v", listener)
	blog.V(5).Infof("cloud listener %+v", cloudListener)

	if strings.ToLower(listener.Spec.Protocol) != strings.ToLower(cloudListener.Spec.Protocol) {
		// delete listener
		err := c.deleteListener(listener.Spec.LoadbalancerID, listener.Spec.Port)
		if err != nil {
			return "", err
		}
		// create listener
		listenerID, err := c.createListner(listener)
		if err != nil {
			return "", err
		}
		return listenerID, nil
	}

	if err := c.updateListener(listener, cloudListener); err != nil {
		return "", err
	}
	return cloudListener.Status.ListenerID, nil
}

// DeleteListener delete listener by name
func (c *Clb) DeleteListener(listener *networkextensionv1.Listener) error {
	return c.deleteListener(listener.Spec.LoadbalancerID, listener.Spec.Port)
}

// EnsureSegmentListener ensure listener with port segment
func (c *Clb) EnsureSegmentListener(listener *networkextensionv1.Listener) (string, error) {
	cloudListener, err := c.getSegmentListenerInfoByPort(listener.Spec.LoadbalancerID, listener.Spec.Port)
	if errors.Is(err, cloud.ErrListenerNotFound) {
		// to create listener
		listenerID, err := c.createSegmentListener(listener)
		if err != nil {
			return "", err
		}
		return listenerID, nil
	}

	blog.V(5).Infof("new listener %+v", listener)
	blog.V(5).Infof("cloud listener %+v", cloudListener)

	if strings.ToLower(listener.Spec.Protocol) != strings.ToLower(cloudListener.Spec.Protocol) {
		// delete listener
		err := c.deleteSegmentListener(listener.Spec.LoadbalancerID, listener.Spec.Port)
		if err != nil {
			return "", err
		}
		// create listener
		listenerID, err := c.createSegmentListener(listener)
		if err != nil {
			return "", err
		}
		return listenerID, nil
	}

	if err := c.updateSegmentListener(listener, cloudListener); err != nil {
		return "", err
	}
	return cloudListener.Status.ListenerID, nil
}

// DeleteSegmentListener delete segment listener
func (c *Clb) DeleteSegmentListener(listener *networkextensionv1.Listener) error {
	return c.deleteSegmentListener(listener.Spec.LoadbalancerID, listener.Spec.Port)
}
