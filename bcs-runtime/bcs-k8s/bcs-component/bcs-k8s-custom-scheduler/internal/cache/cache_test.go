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

package cache

import (
	"reflect"
	"sort"
	"testing"
)

// TestGetEmpty test get empty
func TestGetEmpty(t *testing.T) {
	testCache := NewResourceCache()
	res := testCache.GetResource("tmpkey")
	if res != nil {
		t.Fatalf("expect nil but get %v", res)
	}
}

// TestResourceCacheUpdate test update function
func TestResourceCacheUpdate(t *testing.T) {
	testCases := []struct {
		title    string
		resource *Resource
		hasErr   bool
	}{
		{
			title: "normal add",
			resource: &Resource{
				PodName:      "testpod",
				PodNamespace: "testns",
				Node:         "testnode",
				ResourceKind: "testkind",
				Value:        1,
			},
			hasErr: false,
		},
		{
			title: "err empty Node name",
			resource: &Resource{
				PodName:      "testpod",
				PodNamespace: "testns",
				Node:         "",
				ResourceKind: "testkind",
				Value:        1,
			},
			hasErr: true,
		},
		{
			title:    "err empty resource",
			resource: nil,
			hasErr:   true,
		},
	}
	for _, test := range testCases {
		t.Logf("test title %s", test.title)
		testCache := NewResourceCache()
		err := testCache.UpdateResource(test.resource)
		if err != nil {
			if test.hasErr {
				continue
			}
			t.Fatalf("expect no error but get err %s", err.Error())
		}
		if test.hasErr {
			t.Fatalf("expect error but get no error")
		}
		storedRes := testCache.GetResource(test.resource.Key())
		if !reflect.DeepEqual(storedRes, test.resource) {
			t.Fatalf("expect resource %v, but get %v", test.resource, storedRes)
		}
	}
}

func getResources() []*Resource {
	return []*Resource{
		{
			PodName:      "testpod2",
			PodNamespace: "testns",
			Node:         "testnode",
			ResourceKind: "testkind",
			Value:        1,
		},
		{
			PodName:      "testpod1",
			PodNamespace: "testns",
			Node:         "testnode",
			ResourceKind: "testkind",
			Value:        1,
		},
	}
}

// TestResourceCacheDelete test delete function
func TestResourceCacheDelete(t *testing.T) {
	testCache := NewResourceCache()
	for _, res := range getResources() {
		testCache.UpdateResource(res)
	}
	err := testCache.DeleteResource("testpod1/testns")
	if err != nil {
		t.Fatalf("expect no err, but get err %s", err.Error())
	}
	res := testCache.GetResource("testpod1/testns")
	if res != nil {
		t.Fatalf("expect nil but get %v", res)
	}
}

// TestGetNodeResources test get node resources function
func TestGetNodeResources(t *testing.T) {
	tmpList := getResources()
	testCache := NewResourceCache()
	for _, res := range tmpList {
		testCache.UpdateResource(res)
	}
	resList := testCache.GetNodeResources("testnode")
	sort.Slice(tmpList, func(i, j int) bool {
		return tmpList[i].Key() < tmpList[j].Key()
	})

	if !reflect.DeepEqual(resList, tmpList) {
		for _, res := range resList {
			t.Logf("%v", res)
		}
		for _, res := range tmpList {
			t.Logf("%v", res)
		}
		t.Fatalf("node resources are not correct")
	}

	resList = testCache.GetNodeResources("emptynode")
	if len(resList) != 0 {
		t.Fatalf("expect no resources, but get %v", resList)
	}
}
