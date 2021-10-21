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
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	loadbalance "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/network/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/cloudlb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/cloudlb/qcloud/qcloudif"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/cloudlb/qcloud/qcloudif/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/cloudlb/qcloud/qcloudif/sdk"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	clbBackendsAddMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "clb",
		Subsystem: "updater",
		Name:      "add_backends",
		Help:      "clb backend add",
	}, []string{"ip", "port"})
	clbBackendsDeleteMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "clb",
		Subsystem: "updater",
		Name:      "delete_backends",
		Help:      "clb backend change",
	}, []string{"ip", "port"})
)

func init() {
	prometheus.Register(clbBackendsAddMetric)
	prometheus.Register(clbBackendsDeleteMetric)
}

// ClbClient client for operating clb
type ClbClient struct {
	clbInfo    *loadbalance.CloudLoadBalancer
	clbCfg     *CLBConfig
	clbAdapter qcloudif.ClbAdapter
}

// NewClient construct new clb client
func NewClient(clbInfo *loadbalance.CloudLoadBalancer) (cloudlb.Interface, error) {
	if len(clbInfo.Name) < 1 || len(clbInfo.Name) > 50 {
		return nil, fmt.Errorf("clb name length %d invalid, valid range [1, 50]", len(clbInfo.Name))
	}
	return &ClbClient{
		clbInfo: clbInfo,
	}, nil
}

// LoadConfig load clb config from env
func (clb *ClbClient) LoadConfig() error {
	clbCfg := NewCLBCfg()
	err := clbCfg.LoadFromEnv()
	if err != nil {
		blog.Errorf("load clb config from env failed, err %s", err.Error())
		return fmt.Errorf("load clb config from env failed, err %s", err.Error())
	}
	if strings.ToLower(clb.clbInfo.NetworkType) == loadbalance.ClbNetworkTypePrivate &&
		len(clbCfg.SubnetID) == 0 {
		blog.Errorf("private clb instance must specified subnet id")
		return fmt.Errorf("private clb instance must specified subnet id")
	}
	clb.clbCfg = clbCfg
	if clb.clbCfg.ImplementMode == ConfigBcsClbImplementAPI {
		//create api client
		clb.clbAdapter = api.NewCloudClbAPI(
			clbCfg.ProjectID, clbCfg.Region, clbCfg.SubnetID,
			clbCfg.VpcID, clbCfg.SecretID, clbCfg.SecretKey, clbCfg.BackendMode,
			clbCfg.WaitPeriodLBDealing, clbCfg.WaitPeriodExceedLimit, clbCfg.MaxTimeout)
	} else {
		sdkConfig := clb.clbCfg.GenerateSdkConfig()
		clb.clbAdapter = sdk.NewClient(sdkConfig)
	}

	return nil
}

// CreateLoadbalance create lb, get lb id and vips
func (clb *ClbClient) CreateLoadbalance() (*loadbalance.CloudLoadBalancer, error) {

	lb, err := clb.DescribeLoadbalance(clb.clbInfo.Name)
	if err != nil {
		return nil, fmt.Errorf("describe loadbalance by name %s failed, err %s", clb.clbInfo.Name, err.Error())
	}
	if lb != nil {
		if lb.NetworkType != clb.clbInfo.NetworkType {
			blog.Errorf(
				"loadbalancer with name %s already existed with networktype %s, but want networktype %s, failed",
				lb.Name, lb.NetworkType, clb.clbInfo.NetworkType)
			return nil, fmt.Errorf(
				"loadbalancer with name %s already existed with networktype %s, but want networktype %s, failed",
				lb.Name, lb.NetworkType, clb.clbInfo.NetworkType)
		}
		blog.Infof("loadbalancer with name %s networktype %s already existed, take over it", lb.Name, lb.NetworkType)
		clb.clbInfo.ID = lb.ID
		clb.clbInfo.VIPS = lb.VIPS
		clb.clbInfo.NetworkType = lb.NetworkType
		return clb.clbInfo, nil
	}

	lbID, vips, err := clb.clbAdapter.CreateLoadBalance(clb.clbInfo)
	if err != nil {
		return nil, fmt.Errorf("clb adapter create loadbalance with %v failed, err %s", clb.clbInfo, err.Error())
	}
	clb.clbInfo.ID = lbID
	clb.clbInfo.VIPS = append(clb.clbInfo.VIPS, vips...)
	return clb.clbInfo, nil
}

// DescribeLoadbalance query lb by name
func (clb *ClbClient) DescribeLoadbalance(name string) (*loadbalance.CloudLoadBalancer, error) {

	lbInfo, isExisted, err := clb.clbAdapter.DescribeLoadBalance(name)
	if err != nil {
		return nil, fmt.Errorf("describe loadbalance with name %s failed, err %s", name, err.Error())
	}
	if !isExisted {
		return nil, nil
	}
	return lbInfo, nil
}

// Update update listener
// if listener does not existed, create one
func (clb *ClbClient) Update(old, cur *loadbalance.CloudListener) error {

	_, isExisted, err := clb.clbAdapter.DescribeListener(old.Spec.LoadBalancerID, old.Spec.ListenerID, -1)
	if err != nil {
		return fmt.Errorf("describe listener by lbid %s listener id %s failed, %s",
			old.Spec.LoadBalancerID, old.Spec.ListenerID, err.Error())
	}
	if !isExisted {
		blog.Warnf("listener %s name %s does not exist, try to create a new listener",
			old.Spec.LoadBalancerID, old.GetName())
		err := clb.addListener(cur)
		if err != nil {
			return fmt.Errorf("OnAdd failed, %s", err.Error())
		}
		return nil
	}

	// cur does not have listenerId, because it coverted from template data
	// we should set listenerId here
	cur.Spec.ListenerID = old.Spec.ListenerID
	if old.IsEqual(cur) {
		blog.Infof("no need to update current listener is equal to old listener")
		return nil
	}

	// tcp listener use
	if old.Spec.Protocol == loadbalance.ClbListenerProtocolTCP ||
		old.Spec.Protocol == loadbalance.ClbListenerProtocolUDP {
		err := clb.update4LayerListener(old, cur)
		if err != nil {
			blog.Errorf("update 4 layer listener failed, %s", err)
			return fmt.Errorf("update 4 layer listener failed, %s", err)
		}
	} else if old.Spec.Protocol == loadbalance.ClbListenerProtocolHTTP ||
		old.Spec.Protocol == loadbalance.ClbListenerProtocolHTTPS {
		err := clb.update7LayerListener(old, cur)
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

// Add add cloud listener
// first check if any listener with the same port is existed, if existed, delete it
func (clb *ClbClient) Add(ls *loadbalance.CloudListener) error {
	listener, isExisted, err := clb.clbAdapter.DescribeListener(ls.Spec.LoadBalancerID, "", int(ls.Spec.ListenPort))
	if err != nil {
		return fmt.Errorf("QCloudDescribeListener %d failed, err %s", ls.Spec.ListenPort, err.Error())
	}
	if isExisted {
		blog.Warnf("listener %s port %d is already in use, try to delete it",
			listener.Spec.ListenerID, listener.Spec.ListenPort)
		err := clb.clbAdapter.DeleteListener(listener.Spec.LoadBalancerID, listener.Spec.ListenerID)
		if err != nil {
			return fmt.Errorf("delete listener %s failed with port %d, err %s",
				listener.Spec.ListenerID, listener.Spec.ListenPort, err.Error())
		}
	}
	return clb.addListener(ls)
}

// Delete delete listener
// if listener does not existed, do nothing
func (clb *ClbClient) Delete(ls *loadbalance.CloudListener) error {

	_, isExisted, err := clb.clbAdapter.DescribeListener(ls.Spec.LoadBalancerID, ls.Spec.ListenerID, -1)
	if err != nil {
		return fmt.Errorf("describe listener failed, %s", err.Error())
	}
	if !isExisted {
		blog.Warnf("no need to delete, listener %s does not exist", ls.Spec.ListenerID)
		return nil
	}

	err = clb.clbAdapter.DeleteListener(ls.Spec.LoadBalancerID, ls.Spec.ListenerID)
	if err != nil {
		return fmt.Errorf("QCloudDeleteListener failed, %s", err.Error())
	}
	return nil
}

// ListListeners list listener
// list listeners in current clb instance
func (clb *ClbClient) ListListeners() ([]*loadbalance.CloudListener, error) {
	return clb.clbAdapter.ListListener(clb.clbInfo.ID)
}
