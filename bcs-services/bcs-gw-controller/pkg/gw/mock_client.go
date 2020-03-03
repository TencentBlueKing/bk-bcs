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

package gw

// MockClient mock gw client
type MockClient struct {
	svcs []*Service
}

// Run implements interface
func (mc *MockClient) Run() error {
	return nil
}

// Update implements interface
func (mc *MockClient) Update(svcs []*Service) error {
	mc.svcs = svcs
	return nil
}

// Delete implements interface
func (mc *MockClient) Delete(svcs []*Service) error {
	for _, svcToDel := range svcs {
		for index, svc := range mc.svcs {
			if svc.Key() == svcToDel.Key() {
				mc.svcs = append(mc.svcs[index:], mc.svcs[:index+1]...)
				break
			}
		}
	}

	return nil
}

// List list services
func (mc *MockClient) List() ([]*Service, error) {
	return mc.svcs, nil
}
