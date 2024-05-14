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

// Package cachemanager xxx
package cachemanager

import (
	"context"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var (
	once         sync.Once
	cacheManager *CacheManager
)

// CacheInterface defines the interface of cache manager
type CacheInterface interface {
	Init() error
	Close()

	GetConfigMap(namespace, name string) (*corev1.ConfigMap, error)
	GetSecret(namespace, name string) (*corev1.Secret, error)
	UpdateSecret(ctx context.Context, secret *corev1.Secret) error
}

// CacheManager will make cache for some k8s resource
type CacheManager struct {
	client *kubernetes.Clientset
	stopCh chan struct{}

	informerFactory   informers.SharedInformerFactory
	configMapInformer corev1informers.ConfigMapInformer
	secretInformer    corev1informers.SecretInformer
}

// NewCacheManager create the instance of cache manager
func NewCacheManager(client *kubernetes.Clientset) CacheInterface {
	once.Do(func() {
		cacheManager = &CacheManager{
			stopCh: make(chan struct{}),
			client: client,
		}
	})
	return cacheManager
}

var (
	informerReSyncPeriod    = 30 * time.Minute
	waitInformerSyncTimeout = 180 * time.Second
)

// Init create and sync informers
func (m *CacheManager) Init() error {
	m.informerFactory = informers.NewSharedInformerFactory(m.client, informerReSyncPeriod)
	m.configMapInformer = m.informerFactory.Core().V1().ConfigMaps()
	m.configMapInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) {},
		UpdateFunc: func(oldObj, newObj interface{}) {},
		DeleteFunc: func(obj interface{}) {},
	})
	m.secretInformer = m.informerFactory.Core().V1().Secrets()
	m.secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) {},
		UpdateFunc: func(oldObj, newObj interface{}) {},
		DeleteFunc: func(obj interface{}) {},
	})
	m.informerFactory.Start(m.stopCh)

	blog.Infof("image acceleration cache manager waiting for informers sync...")
	waitCh := make(chan error)
	go func() {
		err := m.waitForSync()
		waitCh <- err
	}()
	waitTimeout := time.After(waitInformerSyncTimeout)
	select {
	case err := <-waitCh:
		if err != nil {
			return errors.Wrapf(err, "wait for informer sync failed")
		}
		blog.Infof("image acceleration cache manager sync success.")
		return nil
	case <-waitTimeout:
		close(m.stopCh)
		return errors.Errorf("CacheManager sync informers cache timeout")
	}
}

func (m *CacheManager) waitForSync() error {
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(m.stopCh, m.configMapInformer.Informer().HasSynced) {
		return errors.Errorf("failed to sync configmap informer")
	}
	return nil
}

// Close informers
func (m *CacheManager) Close() {
	close(m.stopCh)
}

// GetConfigMap get configmap from informer cache
func (m *CacheManager) GetConfigMap(namespace, name string) (*corev1.ConfigMap, error) {
	configMap, err := m.configMapInformer.Lister().ConfigMaps(namespace).Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, "get configmap '%s/%s' from informer failed", namespace, name)
	}
	return configMap, nil
}

// GetSecret get secret from informer cache
func (m *CacheManager) GetSecret(namespace, name string) (*corev1.Secret, error) {
	secret, err := m.secretInformer.Lister().Secrets(namespace).Get(name)
	if err != nil {
		return nil, errors.Wrapf(err, "get secret '%s/%s' from informer failed", namespace, name)
	}
	return secret, nil
}

// UpdateSecret update secret
func (m *CacheManager) UpdateSecret(ctx context.Context, secret *corev1.Secret) error {
	if _, err := m.client.CoreV1().Secrets(secret.Namespace).Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
		return errors.Wrapf(err, "update secret '%s/%s' failed", secret.Namespace, secret.Name)
	}
	return nil
}
