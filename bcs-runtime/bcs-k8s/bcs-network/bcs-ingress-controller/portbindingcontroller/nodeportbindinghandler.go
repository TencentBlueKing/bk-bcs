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

package portbindingcontroller

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/pkg/errors"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/utils"
)

// NodePortBindingCache cache node portbinding info
type NodePortBindingCache struct {
	isInit    bool
	k8sClient client.Client

	Cache map[string]string
	sync.RWMutex
}

// NewNodePortBindingCache return new node port binding cache
func NewNodePortBindingCache(k8sClient client.Client) *NodePortBindingCache {
	return &NodePortBindingCache{
		isInit:    false,
		k8sClient: k8sClient,
		Cache:     make(map[string]string),
	}
}

// Init return err if init failed
func (nc *NodePortBindingCache) Init() error {
	if !nc.isInit {
		nc.Lock()
		defer nc.Unlock()
		if !nc.isInit {
			if err := nc.initCache(); err != nil {
				return err
			}
			nc.isInit = true
		}
	}
	return nil
}

// initCache use node list to build cache
func (nc *NodePortBindingCache) initCache() error {
	portBindingList := &networkextensionv1.PortBindingList{}
	if err := retry.OnError(retry.DefaultRetry, func(err error) bool {
		return true
	}, func() error {
		selector, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(k8slabels.Set(map[string]string{
			networkextensionv1.PortBindingTypeLabelKey: networkextensionv1.PortBindingTypeNode,
		})))
		if err != nil {
			return err
		}
		return nc.k8sClient.List(context.TODO(), portBindingList, &client.ListOptions{LabelSelector: selector})
	}); err != nil {
		blog.Errorf("get node list failed, err: %s", err.Error())
		return err
	}

	for _, portBinding := range portBindingList.Items {
		if err := nc.setCache(&portBinding); err != nil {
			return fmt.Errorf("update cache for portBinding[%s/%s]failed, err: %s", portBinding.GetNamespace(),
				portBinding.GetName(), err.Error())
		}
	}

	return nil
}

// UpdateCache update cache by portbinding
func (nc *NodePortBindingCache) UpdateCache(portBinding *networkextensionv1.PortBinding) error {
	nc.Lock()
	defer nc.Unlock()
	return nc.setCache(portBinding)
}

// set cache without lock
func (nc *NodePortBindingCache) setCache(portBinding *networkextensionv1.PortBinding) error {
	if portBinding == nil {
		return nil
	}
	if portBinding.Status.Status != constant.PortBindingStatusReady {
		delete(nc.Cache, portBinding.GetName())
		return nil
	}

	poolBindingAnno, err := json.Marshal(portBinding.Spec.PortBindingList)
	if err != nil {
		err = errors.Wrapf(err, "generate poolBinding annotation failed")
		blog.Errorf("%v", err)
		return err
	}
	nc.Cache[portBinding.GetName()] = string(poolBindingAnno)

	return nil
}

// GetCache return copy of cache
func (nc *NodePortBindingCache) GetCache() map[string]string {
	newMap := make(map[string]string)
	if err := nc.Init(); err != nil {
		return newMap
	}

	nc.RLock()
	defer nc.RUnlock()

	for k, v := range nc.Cache {
		newMap[k] = v
	}
	return newMap
}

type nodePortBindingHandler struct {
	node      *k8scorev1.Node
	bindCache *NodePortBindingCache

	*portBindingHandler
}

func newNodePortBindingHandler(ctx context.Context, k8sClient client.Client, eventer record.EventRecorder,
	node *k8scorev1.Node, bindCache *NodePortBindingCache) *nodePortBindingHandler {
	npbh := &nodePortBindingHandler{node: node, bindCache: bindCache}
	npbh.portBindingHandler = newPortBindingHandler(ctx, k8sClient, eventer)
	npbh.portBindingHandler.generateTargetGroup = npbh.generateTargetGroup
	npbh.portBindingHandler.postPortBindingUpdate = npbh.postPortBindingUpdate
	npbh.portBindingHandler.postPortBindingClean = npbh.postPortBindingClean
	npbh.portBindingHandler.portBindingType = networkextensionv1.PortBindingTypeNode

	return npbh
}

// generateTargetGroup use node internal ip as target group
func (n *nodePortBindingHandler) generateTargetGroup(item *networkextensionv1.PortBindingItem) *networkextensionv1.
	ListenerTargetGroup {
	if n.node == nil {
		blog.Warnf("generate target group for empty node")
		return nil
	}
	var nodeIP string
	for _, address := range n.node.Status.Addresses {
		if address.Type == k8scorev1.NodeInternalIP {
			nodeIP = address.Address
			break
		}
	}
	backend := networkextensionv1.ListenerBackend{
		IP:     nodeIP,
		Port:   item.RsStartPort,
		Weight: networkextensionv1.DefaultWeight,
	}
	return &networkextensionv1.ListenerTargetGroup{
		TargetGroupProtocol: item.Protocol,
		Backends:            []networkextensionv1.ListenerBackend{backend},
	}
}

// postPortBindingUpdate do after portbinding update
func (n *nodePortBindingHandler) postPortBindingUpdate(portBinding *networkextensionv1.PortBinding) error {
	if n.node == nil {
		err := errors.New("update portbinding for empty node")
		blog.Warnf("%v", err)
		return err
	}

	portBindingItemsBytes, err := json.Marshal(portBinding.Spec.PortBindingList)
	if err != nil {
		return fmt.Errorf("marshal node %s portbindingLisr failed, err:%s", n.node.GetName(), err.Error())
	}
	if err = utils.PatchNodeAnnotation(n.ctx, n.k8sClient, n.node, map[string]interface{}{
		constant.AnnotationForPortPoolBindingStatus: portBinding.Status.Status,
		constant.AnnotationForPortPoolBindings:      string(portBindingItemsBytes),
	}); err != nil {
		return fmt.Errorf("patch annotataion to node %s failed, err: %s", n.node.GetName(), err.Error())
	}

	if err = n.updateAllConfigMap(portBinding); err != nil {
		return err
	}

	n.recordEvent(portBinding, k8scorev1.EventTypeNormal, ReasonPortBindingUpdatePodSuccess,
		MsgPortBindingUpdatePodSuccess)
	return nil
}

// postPortBindingClean do after portbinding need clean
func (n *nodePortBindingHandler) postPortBindingClean(portBinding *networkextensionv1.PortBinding) error {
	// 优先清理configmap中的端口
	if err := n.updateAllConfigMap(portBinding); err != nil {
		return err
	}
	node := &k8scorev1.Node{}
	if err := n.k8sClient.Get(n.ctx, k8stypes.NamespacedName{Name: portBinding.GetName()}, node); err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Infof("node '%s' has been deleted, do not clean annotation", portBinding.GetName())
			return nil
		}
		blog.Warnf("get node '%s' failed, err %s", portBinding.GetName(),
			err.Error())
		return errors.Wrapf(err, "get node '%s' failed", portBinding.GetName())
	}

	delete(node.Annotations, constant.AnnotationForPortPoolBindings)
	delete(node.Annotations, constant.AnnotationForPortPoolBindingStatus)
	if err := n.k8sClient.Update(context.TODO(), node, &client.UpdateOptions{}); err != nil {
		blog.Warnf("remove annotation from node %s failed, err %s", portBinding.GetName(), err.Error())
		return errors.Wrapf(err, "remove annotation from node %s failed", portBinding.GetName())
	}

	return nil
}

// nolint unused
func (n *nodePortBindingHandler) patchNodeAnnotation(node *k8scorev1.Node, status string) error {
	rawPatch := client.RawPatch(k8stypes.MergePatchType, []byte(
		"{\"metadata\":{\"annotations\":{\""+constant.AnnotationForPortPoolBindingStatus+
			"\":\""+status+"\"}}}"))
	updatePod := &k8scorev1.Node{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      node.GetName(),
			Namespace: node.GetNamespace(),
		},
	}
	if err := n.k8sClient.Patch(context.Background(), updatePod, rawPatch, &client.PatchOptions{}); err != nil {
		blog.Errorf("patch node %s/%s annotation status failed, err %s", node.GetNamespace(), node.GetName(),
			err.Error())
		return fmt.Errorf("patch node %s/%s annotation status failed, err %s", node.GetNamespace(), node.GetName(),
			err.Error())
	}
	return nil
}

func (n *nodePortBindingHandler) updateAllConfigMap(portBinding *networkextensionv1.PortBinding) error {
	if err := n.bindCache.UpdateCache(portBinding); err != nil {
		return err
	}
	selector, lerr := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(k8slabels.Set(map[string]string{
		networkextensionv1.NodePortBindingConfigMapNsLabel: networkextensionv1.NodePortBindingConfigMapNsLabelValue,
	})))
	if lerr != nil {
		lerr = errors.Wrapf(lerr, "build config map label selector failed")
		blog.Warnf("%v", lerr)
		return lerr
	}
	nsList := &k8scorev1.NamespaceList{}
	if err := n.k8sClient.List(context.TODO(), nsList, client.MatchingLabelsSelector{Selector: selector}); err != nil {
		err = errors.Wrapf(err, "list namespace failed")
		blog.Warnf("%v", err)
		return err
	}
	for _, item := range nsList.Items {
		if err := n.updateConfigMap(item.Name); err != nil {
			return err
		}
	}

	return nil
}

func (n *nodePortBindingHandler) updateConfigMap(configMapNamespace string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		configMap := &k8scorev1.ConfigMap{}
		if err := n.k8sClient.Get(n.ctx, k8stypes.NamespacedName{Namespace: configMapNamespace,
			Name: networkextensionv1.NodePortBindingConfigMapName}, configMap); err != nil {
			if k8serrors.IsNotFound(err) {
				if cerr := n.createConfigMap(configMapNamespace); cerr != nil {
					return cerr
				}
			} else {
				err = errors.Wrapf(err, "get config map failed")
				blog.Errorf("%v", err)
				return err
			}
		}

		configMap.Data = n.bindCache.GetCache()

		if err := n.k8sClient.Update(n.ctx, configMap); err != nil {
			err = errors.Wrapf(err, "update node portbinding[''] config map failed")
			blog.Errorf("%v", err)
			return err
		}

		return nil
	})
}

func (n *nodePortBindingHandler) createConfigMap(configMapNamespace string) error {
	configMap := &k8scorev1.ConfigMap{}
	configMap.SetNamespace(configMapNamespace)
	configMap.SetName(networkextensionv1.NodePortBindingConfigMapName)

	configMap.Data = n.bindCache.GetCache()

	if err := n.k8sClient.Create(n.ctx, configMap); err != nil {
		err = errors.Wrapf(err, "update node portbinding[''] config map failed")
		blog.Errorf("%v", err)
		return err
	}

	return nil
}
