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

package serviced

import (
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
)

const (
	// defaultKeepAliveInterval service key lease keep alive interval.
	defaultKeepAliveInterval = 5 * time.Second
	// defaultSyncMasterInterval sync master interval.
	defaultSyncMasterInterval = 10 * time.Second
	// defaultGrantLeaseTTL etcd lease ttl.
	defaultGrantLeaseTTL = 10
	// defaultErrSleepTime is exec failed need to wait time.
	defaultErrSleepTime = time.Second
)

// ServiceOption defines a service related options.
type ServiceOption struct {
	Name cc.Name
	IP   string
	Port uint
	// Uid is a service's unique identity.
	Uid string
}

// Validate the service option
func (so ServiceOption) Validate() error {
	if len(so.Name) == 0 {
		return errors.New("service name is empty")
	}

	if len(so.IP) == 0 || so.IP == "0.0.0.0" || so.IP == "::" {
		return errors.New("invalid service ip")
	}

	if so.Port == 0 {
		return errors.New("invalid service port")
	}

	if len(so.Uid) == 0 {
		return errors.New("invalid service uid")
	}

	return nil
}

// DiscoveryOption defines all the service discovery
// related options.
type DiscoveryOption struct {
	Name cc.Name
}

// GrpcServiceDiscoveryName grpc dial service discovery target name, protocol rule: Scheme:///ServiceDiscoveryName.
func GrpcServiceDiscoveryName(serviceName cc.Name) string {
	return "etcd:///" + ServiceDiscoveryName(serviceName)
}

// ServiceDiscoveryName return the service's register path in etcd.
func ServiceDiscoveryName(serviceName cc.Name) string {
	return fmt.Sprintf("/bk-bscp/services/%s", serviceName)
}

// key return service's register key in etcd.
// e.g: /bk-bscp/services/data-service/0fa709f2-8e35-11ec-83f6-acde48001122
func key(path, uid string) string {
	return fmt.Sprintf("%s/%s", path, uid)
}
