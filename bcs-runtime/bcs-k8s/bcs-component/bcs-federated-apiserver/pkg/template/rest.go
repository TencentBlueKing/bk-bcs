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

package template

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"k8s.io/apiserver/pkg/endpoints/request"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-federated-apiserver/pkg/config"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-federated-apiserver/pkg/storage"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/registry/rest"
	clientgorest "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

const (
	CoreGroupPrefix  = "api"
	NamedGroupPrefix = "apis"

	// DefaultDeleteCollectionWorkers defines the default value for deleteCollectionWorkers
	DefaultDeleteCollectionWorkers = 2
)

// REST implements a RESTStorage for Shadow API
type REST struct {
	// name is the plural name of the resource.
	name string
	// shortNames is a list of suggested short names of the resource.
	shortNames []string
	// namespaced indicates if a resource is namespaced or not.
	namespaced bool
	// kind is the Kind for the resource (e.g. 'Foo' is the kind for a resource 'foo')
	kind string
	// group is the Group of the resource.
	group string
	// version is the Version of the resource.
	version string

	parameterCodec runtime.ParameterCodec

	dryRunClient clientgorest.Interface

	// deleteCollectionWorkers is the maximum number of workers in a single
	// DeleteCollection call. Delete requests for the items in a collection
	// are issued in parallel.
	deleteCollectionWorkers int

	config     *config.Config
	bcsStorage *storage.BcsStorage
}

// Create inserts a new item into Manifest according to the unique key from the object.
func (r *REST) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	return nil, fmt.Errorf("not support create object")
}

// Get retrieves the item from Manifest.
func (r *REST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	result := &unstructured.UnstructuredList{}
	result.SetAPIVersion("v1")
	result.SetKind("List")
	namespace := request.NamespaceValue(ctx)
	dataList, err := r.bcsStorage.ListResources(r.config.MemberCluster, namespace, name, r.kind, 0, 0)
	if err != nil {
		return nil, err
	}
	for _, data := range dataList {
		// set kind and apiversion
		resObj := map[string]interface{}{}
		err := json.Unmarshal(data.Data, &resObj)
		if err != nil {
			return nil, err
		}
		resObj["kind"] = r.kind
		resObj["apiVersion"] = r.version
		result.Items = append(result.Items, unstructured.Unstructured{Object: resObj})
	}

	return result, nil
}

// Update performs an atomic update and set of the object. Returns the result of the update
// or an error. If the registry allows create-on-update, the create flow will be executed.
// A bool is returned along with the object and any errors, to indicate object creation.
func (r *REST) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo,
	createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	// We are explicitly taking forceAllowCreate as false.
	return nil, false, nil
}

// Delete removes the item from storage.
// options can be mutated by rest.BeforeDelete due to a graceful deletion strategy.
func (r *REST) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	return nil, false, nil
}

// DeleteCollection removes all items returned by List with a given ListOptions from storage.
//
// DeleteCollection is currently NOT atomic. It can happen that only subset of objects
// will be deleted from storage, and then an error will be returned.
// In case of success, the list of deleted objects will be returned.
// Copied from k8s.io/apiserver/pkg/registry/generic/registry/store.go and modified.
func (r *REST) DeleteCollection(ctx context.Context, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions, listOptions *internalversion.ListOptions) (runtime.Object, error) {
	if listOptions == nil {
		listOptions = &internalversion.ListOptions{}
	} else {
		listOptions = listOptions.DeepCopy()
	}

	listObj, err := r.List(ctx, listOptions)
	if err != nil {
		return nil, err
	}
	items, err := meta.ExtractList(listObj)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		// Nothing to delete, return now
		return listObj, nil
	}
	// Spawn a number of goroutines, so that we can issue requests to storage
	// in parallel to speed up deletion.
	// It is proportional to the number of items to delete, up to
	// deleteCollectionWorkers (it doesn't make much sense to spawn 16
	// workers to delete 10 items).
	workersNumber := r.deleteCollectionWorkers
	if workersNumber > len(items) {
		workersNumber = len(items)
	}
	if workersNumber < 1 {
		workersNumber = 1
	}
	wg := sync.WaitGroup{}
	toProcess := make(chan int, 2*workersNumber)
	errs := make(chan error, workersNumber+1)

	go func() {
		defer utilruntime.HandleCrash(func(panicReason interface{}) {
			errs <- fmt.Errorf("DeleteCollection distributor panicked: %v", panicReason)
		})
		for i := 0; i < len(items); i++ {
			toProcess <- i
		}
		close(toProcess)
	}()

	wg.Add(workersNumber)
	for i := 0; i < workersNumber; i++ {
		go func() {
			// panics don't cross goroutine boundaries
			defer utilruntime.HandleCrash(func(panicReason interface{}) {
				errs <- fmt.Errorf("DeleteCollection goroutine panicked: %v", panicReason)
			})
			defer wg.Done()

			for index := range toProcess {
				accessor, err := meta.Accessor(items[index])
				if err != nil {
					errs <- err
					return
				}
				// DeepCopy the deletion options because individual graceful deleters communicate changes via a mutating
				// function in the delete strategy called in the delete method.  While that is always ugly, it works
				// when making a single call.  When making multiple calls via delete collection, the mutation applied to
				// pod/A can change the option ultimately used for pod/B.
				if _, _, err := r.Delete(ctx, accessor.GetName(), deleteValidation, options.DeepCopy()); err != nil && !errors.IsNotFound(err) {
					klog.V(4).InfoS("Delete object in DeleteCollection failed", "object", klog.KObj(accessor), "err", err)
					errs <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	select {
	case err := <-errs:
		return nil, err
	default:
		return listObj, nil
	}
}

// Watch makes a matcher for the given label and field.
func (r *REST) Watch(ctx context.Context, options *internalversion.ListOptions) (watch.Interface, error) {
	return nil, fmt.Errorf("not support watch object")
}

// List returns a list of items matching labels.
func (r *REST) List(ctx context.Context, options *internalversion.ListOptions) (runtime.Object, error) {
	klog.Infof("limit %d,continue %s", options.Limit, options.Continue)
	//将continue字段当做分页的起始位置
	var offset int64
	var err error
	if options.Continue != "" {
		offset, err = strconv.ParseInt(options.Continue, 10, 64)
		if err != nil {
			klog.Warningf("continue 字段必须为整数")
			return nil, err
		}
	}
	result := &unstructured.UnstructuredList{}
	result.SetAPIVersion("v1")
	result.SetKind("List")
	namespace := request.NamespaceValue(ctx)
	dataList, err := r.bcsStorage.ListResources(r.config.MemberCluster, namespace, "", r.kind, options.Limit, offset)
	if err != nil {
		return nil, err
	}
	for _, data := range dataList {
		// set kind and apiversion
		resObj := map[string]interface{}{}
		err := json.Unmarshal(data.Data, &resObj)
		if err != nil {
			return nil, err
		}
		resObj["kind"] = r.kind
		resObj["apiVersion"] = r.version
		result.Items = append(result.Items, unstructured.Unstructured{Object: resObj})
	}

	return result, nil
}

func (r *REST) NewList() runtime.Object {
	// Here the list GVK "meta.k8s.io/v1 List" is just a symbol,
	// since the real GVK will be set when List()
	newObj := &unstructured.UnstructuredList{}
	newObj.SetAPIVersion(metav1.SchemeGroupVersion.String())
	newObj.SetKind("List")
	return newObj
}

func (r *REST) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	tableConvertor := rest.NewDefaultTableConvertor(schema.GroupResource{Group: r.group, Resource: r.name})
	return tableConvertor.ConvertToTable(ctx, object, tableOptions)
}

func (r *REST) ShortNames() []string {
	return r.shortNames
}

func (r *REST) SetShortNames(ss []string) {
	r.shortNames = ss
}

func (r *REST) SetName(name string) {
	r.name = name
}

func (r *REST) NamespaceScoped() bool {
	return r.namespaced
}

func (r *REST) SetNamespaceScoped(namespaceScoped bool) {
	r.namespaced = namespaceScoped
}

func (r *REST) Categories() []string {
	//return []string{known.Category}
	return nil
}

func (r *REST) SetGroup(group string) {
	r.group = group
}

func (r *REST) SetVersion(version string) {
	r.version = version
}

func (r *REST) SetKind(kind string) {
	r.kind = kind
}

func (r *REST) New() runtime.Object {
	newObj := &unstructured.Unstructured{}
	orignalGVK := r.GroupVersionKind(schema.GroupVersion{})
	newObj.SetGroupVersionKind(orignalGVK)
	return newObj
}

func (r *REST) GroupVersionKind(_ schema.GroupVersion) schema.GroupVersionKind {
	// use original GVK
	return r.GroupVersion().WithKind(r.kind)
}

func (r *REST) GroupVersion() schema.GroupVersion {
	return schema.GroupVersion{
		Group:   r.group,
		Version: r.version,
	}
}

func (r *REST) normalizeRequest(req *clientgorest.Request, namespace string) *clientgorest.Request {
	if len(r.group) == 0 {
		req.Prefix(CoreGroupPrefix, r.version)
	} else {
		req.Prefix(NamedGroupPrefix, r.group, r.version)
	}
	if r.namespaced {
		req.Namespace(namespace)
	}
	return req
}

func (r *REST) getListKind() string {
	if strings.Contains(r.name, "/") {
		return r.kind
	}
	return fmt.Sprintf("%sList", r.kind)
}

// NewREST returns a RESTStorage object that will work against API services.
func NewREST(dryRunClient clientgorest.Interface, parameterCodec runtime.ParameterCodec,
	bcsStorage *storage.BcsStorage,
	config *config.Config) *REST {
	return &REST{
		dryRunClient:   dryRunClient,
		parameterCodec: parameterCodec,
		// currently we only set a default value for deleteCollectionWorkers
		// TODO: make it configurable?
		deleteCollectionWorkers: DefaultDeleteCollectionWorkers,
		bcsStorage:              bcsStorage,
		config:                  config,
	}
}

var _ rest.GroupVersionKindProvider = &REST{}
var _ rest.CategoriesProvider = &REST{}
var _ rest.ShortNamesProvider = &REST{}
var _ rest.StandardStorage = &REST{}

var supportedSubresources = sets.NewString(
	"scale",
)
