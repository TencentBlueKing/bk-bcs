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
 */

package pkg

import (
	"context"
	"fmt"
	"log"
	"strconv"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"golang.org/x/sync/errgroup"
	appv1 "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/tencentcloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
)

// Handler for validate listener
type Handler struct {
	ctx context.Context

	opts        *ControllerOption
	k8sCli      client.Client
	sdkWrapper  *tencentcloud.SdkWrapper
	lbRegionMap map[string]string
}

// NewHandler for validate listener
func NewHandler(ctx context.Context, k8sCli client.Client, opts *ControllerOption) (*Handler, error) {
	deploy := &appv1.Deployment{}
	if err1 := k8sCli.Get(ctx, k8stypes.NamespacedName{
		Namespace: opts.IngressControllerNamespace,
		Name:      opts.IngresControllerWorkloadName,
	}, deploy); err1 != nil {
		return nil, fmt.Errorf("获取Bcs Ingress Controller deployment[%s/%s]失败, "+
			"请确认初始化参数中的[namespace]/[workloadname] 是否正确。 err: %s", opts.IngressControllerNamespace,
			opts.IngresControllerWorkloadName, err1.Error())
	}

	for _, env := range deploy.Spec.Template.Spec.Containers[0].Env {
		switch env.Name {
		case constant.EnvNameBkBCSClusterID:
			opts.BcsClusterID = env.Value
		case constant.EnvNameIsTCPUDPPortReuse:
			opts.TcpUdpPortReuse, _ = strconv.ParseBool(env.Value)
		case tencentcloud.EnvNameTencentCloudAccessKeyID:
			opts.StoreCloudSecretName = env.ValueFrom.SecretKeyRef.Name
			opts.KeyStoreCloudSecretID = env.ValueFrom.SecretKeyRef.Key
		case tencentcloud.EnvNameTencentCloudAccessKey:
			opts.StoreCloudSecretName = env.ValueFrom.SecretKeyRef.Name
			opts.KeyStoreCloudSecretKey = env.ValueFrom.SecretKeyRef.Key
		case tencentcloud.EnvNameTencentCloudClbDomain:
			opts.CloudDomain = env.Value
		}
	}
	secret := &k8scorev1.Secret{}
	if err := k8sCli.Get(ctx, k8stypes.NamespacedName{
		Namespace: opts.IngressControllerNamespace,
		Name:      opts.StoreCloudSecretName,
	}, secret); err != nil {
		return nil, fmt.Errorf("获取Secret[%s]失败, err: %s", opts.StoreCloudSecretName, err.Error())
	}

	opts.CloudSecretID = string(secret.Data[opts.KeyStoreCloudSecretID])
	opts.CloudSecretKey = string(secret.Data[opts.KeyStoreCloudSecretKey])

	sdkWrapper, err := tencentcloud.NewSdkWrapperWithParams(opts.CloudSecretID, opts.CloudSecretKey, opts.CloudDomain)
	if err != nil {
		return nil, fmt.Errorf("初始化SDK失败, err %s", err.Error())
	}

	return &Handler{
		ctx:    ctx,
		k8sCli: k8sCli,
		opts:   opts,

		sdkWrapper:  sdkWrapper,
		lbRegionMap: map[string]string{},
	}, nil
}

// LoadListener load listeners from informer
func (h *Handler) LoadListener() (map[string][]networkextensionv1.Listener, []networkextensionv1.Listener, error) {
	listenerList := &networkextensionv1.ListenerList{}
	if err := h.k8sCli.List(h.ctx, listenerList); err != nil {
		return nil, nil, fmt.Errorf("获取集群内监听器列表失败, err: %s", err.Error())
	}
	lbListenerMap := make(map[string][]networkextensionv1.Listener)
	notReadyListeners := make([]networkextensionv1.Listener, 0)
	for _, li := range listenerList.Items {
		if li.Status.Status != networkextensionv1.ListenerStatusSynced || li.Status.ListenerID == "" {
			notReadyListeners = append(notReadyListeners, li)
			continue
		}

		lbListenerMap[li.Spec.LoadbalancerID] = append(lbListenerMap[li.Spec.LoadbalancerID], li)

		region, ok := li.Labels[networkextensionv1.LabelKeyForLoadbalanceRegion]
		if !ok {
			return nil, nil, fmt.Errorf("listener [%s/%s] 未设置标签%s", li.GetNamespace(), li.GetName(),
				networkextensionv1.LabelKeyForLoadbalanceRegion)
		}
		h.lbRegionMap[li.Spec.LoadbalancerID] = region
	}
	return lbListenerMap, notReadyListeners, nil
}

// CheckListenerName return invalid listeners
func (h *Handler) CheckListenerName(lbID string, liList []networkextensionv1.Listener) ([]*tclb.Listener, error) {
	invalidListeners := make([]*tclb.Listener, 0)
	if len(liList) == 0 {
		return invalidListeners, nil
	}

	listenerIdList := make([]string, 0, len(liList))
	for _, li := range liList {
		listenerIdList = append(listenerIdList, li.Status.ListenerID)
	}

	region, ok := h.lbRegionMap[lbID]
	if !ok {
		return nil, fmt.Errorf("获取负载均衡[%s]对应的地域信息失败", lbID)
	}

	req := tclb.NewDescribeListenersRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.ListenerIds = tcommon.StringPtrs(listenerIdList)
	resp, err1 := h.sdkWrapper.DescribeListeners(region, req)
	if err1 != nil {
		return nil, fmt.Errorf("DescribeListeners for lb[%s] failed, err: %s", lbID, err1.Error())
	}

	for _, li := range resp.Response.Listeners {
		if li.ListenerName == nil {
			return nil, fmt.Errorf("listeners[%s/%d] 获取云上监听器名称异常", lbID, *li.Port)
		}

		if !tencentcloud.ValidateListenerName(h.opts.ListenerNameValidateMode, lbID, h.opts.BcsClusterID, li) {
			invalidListeners = append(invalidListeners, li)
		}
	}

	return invalidListeners, nil
}

// BatchUpdateListenerName  batch update listener name with ListenerValidateMode
func (h *Handler) BatchUpdateListenerName(invalidListenerMap map[string][]*tclb.Listener) error {
	group := errgroup.Group{}
	group.SetLimit(h.opts.MaxCloudUpdateConcurrent)
	for lbID, liList := range invalidListenerMap {
		lbID := lbID
		liList := liList
		group.Go(func() error {
			return h.UpdateListenerName(lbID, liList)
		})
	}

	return group.Wait()
}

// UpdateListenerName update listener name with ListenerValidateMode
func (h *Handler) UpdateListenerName(lbID string, liList []*tclb.Listener) error {
	log.Printf("update for lb[%s] listeners, num[%d]\n", lbID, len(liList))
	region, ok := h.lbRegionMap[lbID]
	if !ok {
		return fmt.Errorf("未找到负载均衡[%s]对应的地域信息", lbID)
	}

	for idx, li := range liList {
		if idx%5 == 0 {
			log.Printf("lb[%s] update listener[%d/%d]\n", lbID, idx+1, len(liList))
		}
		if err := h.doUpdateListenerName(region, lbID, li); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) doUpdateListenerName(region, lbID string, li *tclb.Listener) error {
	var port int64 = *li.Port
	var endPort int64 = 0
	if li.EndPort != nil {
		endPort = *li.EndPort
	}
	newListenerName := common.GetListenerNameWithProtocol(lbID, *li.Protocol, int(port), int(endPort))
	if h.opts.ListenerNameValidateMode == constant.ListenerNameValidateModeStrict {
		newListenerName = h.opts.BcsClusterID + "-" + newListenerName
	}
	req := tclb.NewModifyListenerRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.ListenerId = li.ListenerId
	req.ListenerName = tcommon.StringPtr(newListenerName)

	if err := h.sdkWrapper.ModifyListener(region, req); err != nil {
		return fmt.Errorf("更新监听器[%s-%s-%d]失败, err: %s", lbID, *li.Protocol, port, err.Error())
	}
	return nil
}
