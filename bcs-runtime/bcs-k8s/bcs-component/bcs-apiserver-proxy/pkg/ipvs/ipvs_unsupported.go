//go:build !linux
// +build !linux

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

package ipvs

import (
	"fmt"
)

// New returns a dummy Interface for unsupported platform.
func New() Interface {
	return &runner{}
}

type runner struct {
}

// Flush xxx
func (runner *runner) Flush() error {
	return fmt.Errorf("IPVS not supported for this platform")
}

// AddVirtualServer xxx
func (runner *runner) AddVirtualServer(*VirtualServer) error {
	return fmt.Errorf("IPVS not supported for this platform")
}

// UpdateVirtualServer xxx
func (runner *runner) UpdateVirtualServer(*VirtualServer) error {
	return fmt.Errorf("IPVS not supported for this platform")
}

// DeleteVirtualServer xxx
func (runner *runner) DeleteVirtualServer(*VirtualServer) error {
	return fmt.Errorf("IPVS not supported for this platform")
}

// GetVirtualServer xxx
func (runner *runner) GetVirtualServer(*VirtualServer) (*VirtualServer, error) {
	return nil, fmt.Errorf("IPVS not supported for this platform")
}

// GetVirtualServers xxx
func (runner *runner) GetVirtualServers() ([]*VirtualServer, error) {
	return nil, fmt.Errorf("IPVS not supported for this platform")
}

// AddRealServer xxx
func (runner *runner) AddRealServer(*VirtualServer, *RealServer) error {
	return fmt.Errorf("IPVS not supported for this platform")
}

// GetRealServers xxx
func (runner *runner) GetRealServers(*VirtualServer) ([]*RealServer, error) {
	return nil, fmt.Errorf("IPVS not supported for this platform")
}

// DeleteRealServer xxx
func (runner *runner) DeleteRealServer(*VirtualServer, *RealServer) error {
	return fmt.Errorf("IPVS not supported for this platform")
}

// UpdateRealServer xxx
func (runner *runner) UpdateRealServer(*VirtualServer, *RealServer) error {
	return fmt.Errorf("IPVS not supported for this platform")
}

var _ = Interface(&runner{})
