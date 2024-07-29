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

package http

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/bcs"
)

func TestGetURL(t *testing.T) {
	// with namespace
	client1 := StorageClient{
		HTTPClientConfig: &bcs.HTTPClientConfig{
			URL: "http://www.test.com",
		},
		ClusterID:    "12121",
		Namespace:    "test",
		ResourceType: "Pod",
		ResourceName: "test-data-watch-pod-1",
	}

	url1, _ := client1.GetURL()
	if url1 != "http://www.test.com/bcsstorage/v1/k8s/dynamic/namespace_resources/clusters/12121/"+
		"namespaces/test/Pod/test-data-watch-pod-1" {
		t.Errorf("GetURL with Namespace not null fail, got: %s", url1)
	}

	// with no namespace
	client2 := StorageClient{
		HTTPClientConfig: &bcs.HTTPClientConfig{
			URL: "http://www.test.com",
		},
		ClusterID:    "12121",
		Namespace:    "",
		ResourceType: "Node",
		ResourceName: "test-data-watch-node-1",
	}

	url2, _ := client2.GetURL()
	if url2 != "http://www.test.com/bcsstorage/v1/k8s/dynamic/cluster_resources/"+
		"clusters/12121/Node/test-data-watch-node-1" {
		t.Errorf("GetURL with no Namespace not null fail, got: %s", url2)
	}

	// event
	client3 := StorageClient{
		HTTPClientConfig: &bcs.HTTPClientConfig{
			URL: "http://www.test.com",
		},
		ClusterID:    "12121",
		Namespace:    "test",
		ResourceType: "Event",
		ResourceName: "test-data-watch-event-1",
	}

	url3, _ := client3.GetURL()
	if url3 != "http://www.test.com/bcsstorage/v1/events" {
		t.Errorf("GetURL with type event fail, got: %s", url3)
	}

}
