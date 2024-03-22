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

package bcs

import (
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/cli-runtime/pkg/resource"
)

// ManifestToK8sResources get k8s resources from manifest string
func ManifestToK8sResources(namespace, manifest string, restClientGetter resource.RESTClientGetter) (
	[]*resource.Info, error) {
	start := time.Now()
	builder := resource.NewBuilder(restClientGetter)
	infos, err := builder.
		Unstructured().
		ContinueOnError().
		Stream(strings.NewReader(manifest), "").
		Flatten().
		Do().Infos()
	if err != nil {
		blog.Errorf("get manifest err: %s", err.Error())
		return nil, err
	}
	blog.Debug("parse manifest took ", time.Since(start).String())

	wg := &sync.WaitGroup{}
	wg.Add(len(infos))
	for i, v := range infos {
		if len(v.Namespace) == 0 {
			v.Namespace = namespace
		}
		go func(i int) {
			defer wg.Done()
			if err := infos[i].Get(); err != nil {
				blog.Errorf("get k8s resource for %s in %s err: %s", infos[i].Name, infos[i].Namespace, err.Error())
			}
		}(i)
	}
	wg.Wait()
	blog.Debug("get k8s resource took ", time.Since(start).String())
	return infos, nil
}
