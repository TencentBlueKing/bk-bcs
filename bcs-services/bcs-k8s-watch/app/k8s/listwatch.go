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

package k8s

import (
	"fmt"
	"strings"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

// MultiNamespaceListerWatcher takes allowed and denied namespaces and a
// cache.ListerWatcher generator func and returns a single cache.ListerWatcher
// capable of operating on multiple namespaces.
//
// Allowed namespaces and denied namespaces are mutually exclusive.
// If allowed namespaces contain multiple items, the given denied namespaces have no effect.
// If the allowed namespaces includes exactly one entry with the value v1.NamespaceAll (empty string),
// the given denied namespaces are applied.
func MultiNamespaceListerWatcher(allowedNamespaces, deniedNamespaces map[string]struct{}, f func(string) cache.ListerWatcher) cache.ListerWatcher {
	// If there is only one namespace then there is no need to create a
	// multi lister watcher proxy.
	if IsAllNamespaces(allowedNamespaces) {
		return newDenylistListerWatcher(deniedNamespaces, f(v1.NamespaceAll))
	}
	if len(allowedNamespaces) == 1 {
		for n := range allowedNamespaces {
			return f(n)
		}
	}

	var lws []cache.ListerWatcher
	for n := range allowedNamespaces {
		lws = append(lws, f(n))
	}
	return multiListerWatcher(lws)
}

// multiListerWatcher abstracts several cache.ListerWatchers, allowing them
// to be treated as a single cache.ListerWatcher.
type multiListerWatcher []cache.ListerWatcher

// List implements the ListerWatcher interface.
// It combines the output of the List method of every ListerWatcher into
// a single result.
func (mlw multiListerWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	l := metav1.List{}
	resourceVersions := sets.NewString()
	for _, lw := range mlw {
		list, err := lw.List(options)
		if err != nil {
			return nil, err
		}
		items, err := meta.ExtractList(list)
		if err != nil {
			return nil, err
		}
		metaObj, err := meta.ListAccessor(list)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			l.Items = append(l.Items, runtime.RawExtension{Object: item.DeepCopyObject()})
		}
		if !resourceVersions.Has(metaObj.GetResourceVersion()) {
			resourceVersions.Insert(metaObj.GetResourceVersion())
		}
	}
	// Combine the resource versions so that the composite Watch method can
	// distribute appropriate versions to each underlying Watch func.
	l.ListMeta.ResourceVersion = strings.Join(resourceVersions.List(), "/")
	return &l, nil
}

// Watch implements the ListerWatcher interface.
// It returns a watch.Interface that combines the output from the
// watch.Interface of every cache.ListerWatcher into a single result chan.
func (mlw multiListerWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	var resourceVersions string
	// Allow resource versions to be "".
	if options.ResourceVersion != "" {
		rvs := strings.Split(options.ResourceVersion, "/")
		if len(rvs) > 1 {
			return nil, fmt.Errorf("expected resource version to have 1 part, got %d", len(rvs))
		}
		resourceVersions = options.ResourceVersion
	}
	return newMultiWatch(mlw, resourceVersions, options)
}

// multiWatch abstracts multiple watch.Interface's, allowing them
// to be treated as a single watch.Interface.
type multiWatch struct {
	result   chan watch.Event
	stopped  chan struct{}
	stoppers []func()
}

// newMultiWatch returns a new multiWatch or an error if one of the underlying
// Watch funcs errored.
func newMultiWatch(lws []cache.ListerWatcher, resourceVersions string, options metav1.ListOptions) (*multiWatch, error) {
	var (
		result   = make(chan watch.Event)
		stopped  = make(chan struct{})
		stoppers []func()
		wg       sync.WaitGroup
	)

	wg.Add(len(lws))

	for _, lw := range lws {
		o := options.DeepCopy()
		o.ResourceVersion = resourceVersions
		w, err := lw.Watch(*o)
		if err != nil {
			return nil, err
		}

		go func() {
			defer wg.Done()

			for {
				event, ok := <-w.ResultChan()
				if !ok {
					return
				}

				select {
				case result <- event:
				case <-stopped:
					return
				}
			}
		}()
		stoppers = append(stoppers, w.Stop)
	}

	// result chan must be closed,
	// once all event sender goroutines exited.
	go func() {
		wg.Wait()
		close(result)
	}()

	return &multiWatch{
		result:   result,
		stoppers: stoppers,
		stopped:  stopped,
	}, nil
}

// ResultChan implements the watch.Interface interface.
func (mw *multiWatch) ResultChan() <-chan watch.Event {
	return mw.result
}

// Stop implements the watch.Interface interface.
// It stops all of the underlying watch.Interfaces and closes the backing chan.
// Can safely be called more than once.
func (mw *multiWatch) Stop() {
	select {
	case <-mw.stopped:
		// nothing to do, we are already stopped
	default:
		for _, stop := range mw.stoppers {
			stop()
		}
		close(mw.stopped)
	}
	return
}

// IsAllNamespaces checks if the given map of namespaces
// contains only v1.NamespaceAll.
func IsAllNamespaces(namespaces map[string]struct{}) bool {
	_, ok := namespaces[v1.NamespaceAll]
	return ok && len(namespaces) == 1
}
