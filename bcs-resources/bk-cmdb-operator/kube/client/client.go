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

package client

import (
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	extensionsV1beta1 "k8s.io/api/extensions/v1beta1"
	policyV1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func NewClient(c k8sClient.Client) Client {
	return client{
		Client: c,
	}
}

type Client interface {
	k8sClient.Client
	CreateOrUpdateDeploy(deploy *appsv1.Deployment) error
	CreateOrUpdateSts(sts *appsv1.StatefulSet) error
	CreateOrUpdateSecret(secret *v1.Secret) error
	CreateOrUpdateCm(cm *v1.ConfigMap) error
	CreateOrUpdateSa(sa *v1.ServiceAccount) error
	CreateOrUpdateService(svc *v1.Service) error
	CreateOrUpdatePdb(pdb *policyV1beta1.PodDisruptionBudget) error
	CreateOrUpdateIngress(ingress *extensionsV1beta1.Ingress) error
	CreateOrUpdateJob(job *batchV1.Job) error
}

type client struct {
	k8sClient.Client
}

// CreateOrUpdateDeploy create or update Deployment
func (c client) CreateOrUpdateDeploy(deploy *appsv1.Deployment) error {
	found := &appsv1.Deployment{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = c.Create(context.TODO(), deploy)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if !reflect.DeepEqual(deploy.Spec, found.Spec) {
		found.Spec = deploy.Spec
		_ = c.Update(context.TODO(), found)
	}
	return nil
}

// CreateOrUpdateSts create or update StatefulSet
func (c client) CreateOrUpdateSts(sts *appsv1.StatefulSet) error {
	found := &appsv1.StatefulSet{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: sts.Name, Namespace: sts.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = c.Create(context.TODO(), sts)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if !reflect.DeepEqual(sts.Spec, found.Spec) {
		found.Spec = sts.Spec
		_ = c.Update(context.TODO(), found)
	}
	return nil
}

// CreateOrUpdateSecret create or update Secret
func (c client) CreateOrUpdateSecret(secret *v1.Secret) error {
	found := &v1.Secret{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = c.Create(context.TODO(), secret)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if !reflect.DeepEqual(secret.Data, found.Data) {
		found.Data = secret.Data
		_ = c.Update(context.TODO(), found)
	}
	return nil
}

// CreateOrUpdateCm create or update ConfigMap
func (c client) CreateOrUpdateCm(cm *v1.ConfigMap) error {
	found := &v1.ConfigMap{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = c.Create(context.TODO(), cm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if !reflect.DeepEqual(cm.Data, found.Data) {
		found.Data = cm.Data
		_ = c.Update(context.TODO(), found)
	}
	return nil
}

// CreateOrUpdateSa create or update ServiceAccount
func (c client) CreateOrUpdateSa(sa *v1.ServiceAccount) error {
	found := &v1.ServiceAccount{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: sa.Name, Namespace: sa.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = c.Create(context.TODO(), sa)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// CreateOrUpdateService create or update Service
func (c client) CreateOrUpdateService(svc *v1.Service) error {
	found := &v1.Service{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = c.Create(context.TODO(), svc)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// CreateOrUpdatePdb create or update PodDisruptionBudget
func (c client) CreateOrUpdatePdb(pdb *policyV1beta1.PodDisruptionBudget) error {
	found := &policyV1beta1.PodDisruptionBudget{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: pdb.Name, Namespace: pdb.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = c.Create(context.TODO(), pdb)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if !reflect.DeepEqual(pdb.Spec, found.Spec) {
		found.Spec = pdb.Spec
		_ = c.Update(context.TODO(), found)
	}
	return nil
}

// CreateOrUpdateIngress create or update Ingress
func (c client) CreateOrUpdateIngress(ingress *extensionsV1beta1.Ingress) error {
	found := &extensionsV1beta1.Ingress{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: ingress.Name, Namespace: ingress.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = c.Create(context.TODO(), ingress)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if !reflect.DeepEqual(ingress.Spec, found.Spec) {
		found.Spec = ingress.Spec
		_ = c.Update(context.TODO(), found)
	}
	return nil
}

// CreateOrUpdateJob create or update Job
func (c client) CreateOrUpdateJob(job *batchV1.Job) error {
	found := &batchV1.Job{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = c.Create(context.TODO(), job)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if !reflect.DeepEqual(job.Spec, found.Spec) {
		found.Spec = job.Spec
		_ = c.Update(context.TODO(), found)
	}
	return nil
}
