/*
Copyright (C) 2019 The BlueKing Authors. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
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

package adapter

import (
	"fmt"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient"
	cussvcclient "bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient/custom"
	kubesvcclient "bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient/kubernetes"
	mesossvcclient "bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient/mesos"

	"k8s.io/client-go/tools/cache"
)

//NewClient adapter for create different implemented service client
func NewClient(serviceRegistry, kubeconfig string, handler cache.ResourceEventHandler, syncPeriod int) (serviceclient.Client, error) {
	switch serviceRegistry {
	case serviceclient.ServiceRegistryKubernetes:
		k8sSvcClient, err := kubesvcclient.NewClient(kubeconfig, handler, (time.Duration(syncPeriod) * time.Second))
		if err != nil {
			defer k8sSvcClient.Close()
			blog.Errorf("create k8s service client with kubeconfig %s failed, err %s",
				kubeconfig, err.Error())
			return nil, fmt.Errorf("create k8s service client with kubeconfig %s failed, err %s",
				kubeconfig, err.Error())
		}
		return k8sSvcClient, nil
	case serviceclient.ServiceRegistryCustom:
		cusSvcClient, err := cussvcclient.NewClient(kubeconfig, handler, (time.Duration(syncPeriod) * time.Second))
		if err != nil {
			defer cusSvcClient.Close()
			blog.Errorf("create custom service client with kubeconfig %s failed, err %s",
				kubeconfig, err.Error())
			return nil, fmt.Errorf("create custom service client with kubeconfig %s failed, err %s",
				kubeconfig, err.Error())
		}
		return cusSvcClient, nil
	case serviceclient.ServiceRegistryMesos:
		mesosClient, err := mesossvcclient.NewClient(kubeconfig, handler, (time.Duration(syncPeriod) * time.Second))
		if err != nil {
			defer mesosClient.Close()
			blog.Errorf("create mesos service client with kubeconfig %s failed, %s", kubeconfig, err.Error())
			return nil, err
		}
		return mesosClient, nil
	default:
		return nil, fmt.Errorf("unknown registry type %s", serviceRegistry)
	}
}
