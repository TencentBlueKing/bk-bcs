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

package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	k8skube "k8s.io/client-go/kubernetes"
	k8srest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	joiner                  = "-"
	lockRecordAnnotationKey = "sync.bkbcs.tencent.com/lock"
)

// ConfigmapStore is store layer which use configmap to store lock record
type ConfigmapStore struct {
	prefix    string
	namespace string
	cmClient  corev1client.ConfigMapsGetter
}

// NewConfigmapStore create
func NewConfigmapStore(prefix, namespace, kubeconfig string) (*ConfigmapStore, error) {
	var restConfig *k8srest.Config
	var err error
	if len(kubeconfig) == 0 {
		restConfig, err = k8srest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("use InCluster restConfig failed, err %s", err.Error())
		}
	} else {
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("build restConfig by file %s failed, err %s", kubeconfig, err.Error())
		}
	}
	k8sClient, err := k8skube.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("build kube client failed, err %s", err.Error())
	}

	return &ConfigmapStore{
		prefix:    prefix,
		namespace: namespace,
		cmClient:  k8sClient.CoreV1(),
	}, nil
}

func (cs *ConfigmapStore) generateNameFromKey(key string) string {
	return cs.prefix + joiner + key
}

// Get get lock record
func (cs *ConfigmapStore) Get(ctx context.Context, key string) (*LockRecord, []byte, error) {
	var record LockRecord
	var err error
	cm, err := cs.cmClient.ConfigMaps(cs.namespace).Get(ctx, cs.generateNameFromKey(key), metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	if cm.Annotations == nil {
		return &record, nil, nil
	}
	recordBytes, found := cm.Annotations[lockRecordAnnotationKey]
	if !found {
		return &record, nil, nil
	}
	if err = json.Unmarshal([]byte(recordBytes), &record); err != nil {
		return nil, nil, err
	}
	record.ResourceVersion = cm.ResourceVersion
	return &record, []byte(recordBytes), nil
}

// Create create lock record
func (cs *ConfigmapStore) Create(ctx context.Context, key string, lr LockRecord) (
	*LockRecord, error) {
	recordBytes, err := json.Marshal(lr)
	if err != nil {
		return nil, err
	}
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cs.generateNameFromKey(key),
			Namespace: cs.namespace,
			Annotations: map[string]string{
				lockRecordAnnotationKey: string(recordBytes),
			},
		},
	}
	newCm, err := cs.cmClient.ConfigMaps(cs.namespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	newLr := &LockRecord{
		OwnerID:         lr.OwnerID,
		ExpireDuration:  lr.ExpireDuration,
		AcquireTime:     lr.AcquireTime,
		RenewTime:       lr.RenewTime,
		ResourceVersion: newCm.ResourceVersion,
	}
	return newLr, nil
}

// Update update lock record
func (cs *ConfigmapStore) Update(ctx context.Context, key string, lr LockRecord) (
	*LockRecord, error) {
	recordBytes, err := json.Marshal(lr)
	if err != nil {
		return nil, err
	}
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cs.generateNameFromKey(key),
			Namespace: cs.namespace,
			Annotations: map[string]string{
				lockRecordAnnotationKey: string(recordBytes),
			},
			ResourceVersion: lr.ResourceVersion,
		},
	}
	newCm, err := cs.cmClient.ConfigMaps(cs.namespace).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	newLr := &LockRecord{
		OwnerID:         lr.OwnerID,
		ExpireDuration:  lr.ExpireDuration,
		AcquireTime:     lr.AcquireTime,
		RenewTime:       lr.RenewTime,
		ResourceVersion: newCm.ResourceVersion,
	}
	return newLr, nil
}
