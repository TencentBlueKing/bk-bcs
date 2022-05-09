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

package mock

import (
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/providers"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	"github.com/stretchr/testify/mock"
)

//MockProvider mock provider for test
type MockProvider struct {
	mock.Mock
}

// Run mock provider run
func (p *MockProvider) Run(run *hookv1alpha1.HookRun, metric hookv1alpha1.Metric) hookv1alpha1.Measurement {
	args := p.Called(run, metric)
	return args.Get(0).(hookv1alpha1.Measurement)
}

// Resume mock provider resume
func (p *MockProvider) Resume(run *hookv1alpha1.HookRun, metric hookv1alpha1.Metric, measurement hookv1alpha1.Measurement) hookv1alpha1.Measurement {
	args := p.Called(run, metric, measurement)
	return args.Get(0).(hookv1alpha1.Measurement)
}

// Terminate mock provider terminate
func (p *MockProvider) Terminate(run *hookv1alpha1.HookRun, metric hookv1alpha1.Metric, measurement hookv1alpha1.Measurement) hookv1alpha1.Measurement {
	args := p.Called(run, metric, measurement)
	return args.Get(0).(hookv1alpha1.Measurement)
}

// GarbageCollect mock provider garbage collector
func (p *MockProvider) GarbageCollect(run *hookv1alpha1.HookRun, metric hookv1alpha1.Metric, i int) error {
	args := p.Called(run, metric, i)
	return args.Error(0)
}

var _ providers.Provider = &MockProvider{}
