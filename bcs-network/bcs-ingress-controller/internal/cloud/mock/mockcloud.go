/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mock

import (
	v1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	cloud "github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/cloud"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockLoadBalance is a mock of LoadBalance interface
type MockLoadBalance struct {
	ctrl     *gomock.Controller
	recorder *MockLoadBalanceMockRecorder
}

// MockLoadBalanceMockRecorder is the mock recorder for MockLoadBalance
type MockLoadBalanceMockRecorder struct {
	mock *MockLoadBalance
}

// NewMockLoadBalance creates a new mock instance
func NewMockLoadBalance(ctrl *gomock.Controller) *MockLoadBalance {
	mock := &MockLoadBalance{ctrl: ctrl}
	mock.recorder = &MockLoadBalanceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLoadBalance) EXPECT() *MockLoadBalanceMockRecorder {
	return m.recorder
}

// DescribeLoadBalancer mocks base method
func (m *MockLoadBalance) DescribeLoadBalancer(region, lbID, name string) (*cloud.LoadBalanceObject, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeLoadBalancer", region, lbID, name)
	ret0, _ := ret[0].(*cloud.LoadBalanceObject)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeLoadBalancer indicates an expected call of DescribeLoadBalancer
func (mr *MockLoadBalanceMockRecorder) DescribeLoadBalancer(region, lbID, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeLoadBalancer", reflect.TypeOf((*MockLoadBalance)(nil).DescribeLoadBalancer), region, lbID, name)
}

// DescribeLoadBalancerWithNs mocks base method
func (m *MockLoadBalance) DescribeLoadBalancerWithNs(ns, region, lbID, name string) (*cloud.LoadBalanceObject, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeLoadBalancerWithNs", ns, region, lbID, name)
	ret0, _ := ret[0].(*cloud.LoadBalanceObject)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeLoadBalancerWithNs indicates an expected call of DescribeLoadBalancerWithNs
func (mr *MockLoadBalanceMockRecorder) DescribeLoadBalancerWithNs(ns, region, lbID, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeLoadBalancerWithNs", reflect.TypeOf((*MockLoadBalance)(nil).DescribeLoadBalancerWithNs), ns, region, lbID, name)
}

// IsNamespaced mocks base method
func (m *MockLoadBalance) IsNamespaced() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsNamespaced")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsNamespaced indicates an expected call of IsNamespaced
func (mr *MockLoadBalanceMockRecorder) IsNamespaced() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsNamespaced", reflect.TypeOf((*MockLoadBalance)(nil).IsNamespaced))
}

// EnsureListener mocks base method
func (m *MockLoadBalance) EnsureListener(region string, listener *v1.Listener) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureListener", region, listener)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureListener indicates an expected call of EnsureListener
func (mr *MockLoadBalanceMockRecorder) EnsureListener(region, listener interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureListener", reflect.TypeOf((*MockLoadBalance)(nil).EnsureListener), region, listener)
}

// DeleteListener mocks base method
func (m *MockLoadBalance) DeleteListener(region string, listener *v1.Listener) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteListener", region, listener)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteListener indicates an expected call of DeleteListener
func (mr *MockLoadBalanceMockRecorder) DeleteListener(region, listener interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteListener", reflect.TypeOf((*MockLoadBalance)(nil).DeleteListener), region, listener)
}

// EnsureMultiListeners mocks base method
func (m *MockLoadBalance) EnsureMultiListeners(region, lbID string, listeners []*v1.Listener) (map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureMultiListeners", region, lbID, listeners)
	ret0, _ := ret[0].(map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureMultiListeners indicates an expected call of EnsureMultiListeners
func (mr *MockLoadBalanceMockRecorder) EnsureMultiListeners(region, lbID, listeners interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureMultiListeners", reflect.TypeOf((*MockLoadBalance)(nil).EnsureMultiListeners), region, lbID, listeners)
}

// DeleteMultiListeners mocks base method
func (m *MockLoadBalance) DeleteMultiListeners(region, lbID string, listeners []*v1.Listener) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteMultiListeners", region, lbID, listeners)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteMultiListeners indicates an expected call of DeleteMultiListeners
func (mr *MockLoadBalanceMockRecorder) DeleteMultiListeners(region, lbID, listeners interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMultiListeners", reflect.TypeOf((*MockLoadBalance)(nil).DeleteMultiListeners), region, lbID, listeners)
}

// EnsureSegmentListener mocks base method
func (m *MockLoadBalance) EnsureSegmentListener(region string, listener *v1.Listener) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureSegmentListener", region, listener)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureSegmentListener indicates an expected call of EnsureSegmentListener
func (mr *MockLoadBalanceMockRecorder) EnsureSegmentListener(region, listener interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureSegmentListener", reflect.TypeOf((*MockLoadBalance)(nil).EnsureSegmentListener), region, listener)
}

// EnsureMultiSegmentListeners mocks base method
func (m *MockLoadBalance) EnsureMultiSegmentListeners(region, lbID string, listeners []*v1.Listener) (map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureMultiSegmentListeners", region, lbID, listeners)
	ret0, _ := ret[0].(map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureMultiSegmentListeners indicates an expected call of EnsureMultiSegmentListeners
func (mr *MockLoadBalanceMockRecorder) EnsureMultiSegmentListeners(region, lbID, listeners interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureMultiSegmentListeners", reflect.TypeOf((*MockLoadBalance)(nil).EnsureMultiSegmentListeners), region, lbID, listeners)
}

// DeleteSegmentListener mocks base method
func (m *MockLoadBalance) DeleteSegmentListener(region string, listener *v1.Listener) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSegmentListener", region, listener)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSegmentListener indicates an expected call of DeleteSegmentListener
func (mr *MockLoadBalanceMockRecorder) DeleteSegmentListener(region, listener interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSegmentListener", reflect.TypeOf((*MockLoadBalance)(nil).DeleteSegmentListener), region, listener)
}

// MockValidater is a mock of Validater interface
type MockValidater struct {
	ctrl     *gomock.Controller
	recorder *MockValidaterMockRecorder
}

// MockValidaterMockRecorder is the mock recorder for MockValidater
type MockValidaterMockRecorder struct {
	mock *MockValidater
}

// NewMockValidater creates a new mock instance
func NewMockValidater(ctrl *gomock.Controller) *MockValidater {
	mock := &MockValidater{ctrl: ctrl}
	mock.recorder = &MockValidaterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockValidater) EXPECT() *MockValidaterMockRecorder {
	return m.recorder
}

// IsIngressValid mocks base method
func (m *MockValidater) IsIngressValid(ingress *v1.Ingress) (bool, string) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsIngressValid", ingress)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(string)
	return ret0, ret1
}

// IsIngressValid indicates an expected call of IsIngressValid
func (mr *MockValidaterMockRecorder) IsIngressValid(ingress interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsIngressValid", reflect.TypeOf((*MockValidater)(nil).IsIngressValid), ingress)
}

// CheckNoConflictsInIngress mocks base method
func (m *MockValidater) CheckNoConflictsInIngress(ingress *v1.Ingress) (bool, string) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckNoConflictsInIngress", ingress)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(string)
	return ret0, ret1
}

// CheckNoConflictsInIngress indicates an expected call of CheckNoConflictsInIngress
func (mr *MockValidaterMockRecorder) CheckNoConflictsInIngress(ingress interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckNoConflictsInIngress", reflect.TypeOf((*MockValidater)(nil).CheckNoConflictsInIngress), ingress)
}
