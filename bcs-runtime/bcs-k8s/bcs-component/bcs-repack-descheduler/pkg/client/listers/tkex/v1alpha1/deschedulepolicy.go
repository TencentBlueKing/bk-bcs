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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis/tkex/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// DeschedulePolicyLister helps list DeschedulePolicies.
// All objects returned here must be treated as read-only.
type DeschedulePolicyLister interface {
	// List lists all DeschedulePolicies in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.DeschedulePolicy, err error)
	// DeschedulePolicies returns an object that can list and get DeschedulePolicies.
	DeschedulePolicies(namespace string) DeschedulePolicyNamespaceLister
	DeschedulePolicyListerExpansion
}

// deschedulePolicyLister implements the DeschedulePolicyLister interface.
type deschedulePolicyLister struct {
	indexer cache.Indexer
}

// NewDeschedulePolicyLister returns a new DeschedulePolicyLister.
func NewDeschedulePolicyLister(indexer cache.Indexer) DeschedulePolicyLister {
	return &deschedulePolicyLister{indexer: indexer}
}

// List lists all DeschedulePolicies in the indexer.
func (s *deschedulePolicyLister) List(selector labels.Selector) (ret []*v1alpha1.DeschedulePolicy, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.DeschedulePolicy))
	})
	return ret, err
}

// DeschedulePolicies returns an object that can list and get DeschedulePolicies.
func (s *deschedulePolicyLister) DeschedulePolicies(namespace string) DeschedulePolicyNamespaceLister {
	return deschedulePolicyNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// DeschedulePolicyNamespaceLister helps list and get DeschedulePolicies.
// All objects returned here must be treated as read-only.
type DeschedulePolicyNamespaceLister interface {
	// List lists all DeschedulePolicies in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.DeschedulePolicy, err error)
	// Get retrieves the DeschedulePolicy from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.DeschedulePolicy, error)
	DeschedulePolicyNamespaceListerExpansion
}

// deschedulePolicyNamespaceLister implements the DeschedulePolicyNamespaceLister
// interface.
type deschedulePolicyNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all DeschedulePolicies in the indexer for a given namespace.
func (s deschedulePolicyNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.DeschedulePolicy, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.DeschedulePolicy))
	})
	return ret, err
}

// Get retrieves the DeschedulePolicy from the indexer for a given namespace and name.
func (s deschedulePolicyNamespaceLister) Get(name string) (*v1alpha1.DeschedulePolicy, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("deschedulepolicy"), name)
	}
	return obj.(*v1alpha1.DeschedulePolicy), nil
}