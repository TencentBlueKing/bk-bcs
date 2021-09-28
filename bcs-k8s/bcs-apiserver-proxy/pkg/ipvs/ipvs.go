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
	"net"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-apiserver-proxy/pkg/ipvs/exec"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-apiserver-proxy/pkg/ipvs/version"
)

// Interface is an injectable interface for running ipvs commands.  Implementations must be goroutine-safe.
type Interface interface {
	// Flush clears all virtual servers in system. return occurred error immediately.
	Flush() error
	// AddVirtualServer creates the specified virtual server.
	AddVirtualServer(*VirtualServer) error
	// UpdateVirtualServer updates an already existing virtual server.  If the virtual server does not exist, return error.
	UpdateVirtualServer(*VirtualServer) error
	// DeleteVirtualServer deletes the specified virtual server.  If the virtual server does not exist, return error.
	DeleteVirtualServer(*VirtualServer) error
	// Given a partial virtual server, IsVirtualServerAvailable will return the specified virtual server information in the system.
	GetVirtualServer(*VirtualServer) (*VirtualServer, error)
	// GetVirtualServers lists all virtual servers in the system.
	GetVirtualServers() ([]*VirtualServer, error)
	// CreateRealServer creates the specified real server for the specified virtual server.
	AddRealServer(*VirtualServer, *RealServer) error
	// GetRealServers returns all real servers for the specified virtual server.
	GetRealServers(*VirtualServer) ([]*RealServer, error)
	// DeleteRealServer deletes the specified real server from the specified virtual server.
	DeleteRealServer(*VirtualServer, *RealServer) error
	// UpdateRealServer updates the specified real server from the specified virtual server.
	UpdateRealServer(*VirtualServer, *RealServer) error
}

// VirtualServer is an user-oriented definition of an IPVS virtual server in its entirety.
type VirtualServer struct {
	Address   net.IP
	Protocol  string
	Port      uint32
	Scheduler string
	Flags     ServiceFlags
	Timeout   uint32
}

// ServiceFlags is used to specify session affinity, ip hash etc.
type ServiceFlags uint32

const (
	// FlagPersistent specify IPVS service session affinity
	FlagPersistent = 0x1
	// FlagHashed specify IPVS service hash flag
	FlagHashed = 0x2
	// IPVSProxyMode is match set up cluster with ipvs proxy model
	IPVSProxyMode = "ipvs"
)

// IPVS required kernel modules.
const (
	// KernelModuleIPVS is the kernel module "ip_vs"
	KernelModuleIPVS string = "ip_vs"
	// KernelModuleIPVSRR is the kernel module "ip_vs_rr"
	KernelModuleIPVSRR string = "ip_vs_rr"
	// KernelModuleIPVSWRR is the kernel module "ip_vs_wrr"
	KernelModuleIPVSWRR string = "ip_vs_wrr"
	// KernelModuleIPVSSH is the kernel module "ip_vs_sh"
	KernelModuleIPVSSH string = "ip_vs_sh"
	// KernelModuleNfConntrackIPV4 is the module "nf_conntrack_ipv4"
	KernelModuleNfConntrackIPV4 string = "nf_conntrack_ipv4"
	// KernelModuleNfConntrack is the kernel module "nf_conntrack"
	KernelModuleNfConntrack string = "nf_conntrack"
)

// Equal check the equality of virtual server.
// We don't use struct == since it doesn't work because of slice.
func (svc *VirtualServer) Equal(other *VirtualServer) bool {
	return svc.Address.Equal(other.Address) &&
		svc.Protocol == other.Protocol &&
		svc.Port == other.Port &&
		svc.Scheduler == other.Scheduler &&
		svc.Flags == other.Flags &&
		svc.Timeout == other.Timeout
}

// String return virtual server addr:port
func (svc *VirtualServer) String() string {
	return net.JoinHostPort(svc.Address.String(), strconv.Itoa(int(svc.Port))) + "/" + svc.Protocol
}

// RealServer is an user-oriented definition of an IPVS real server in its entirety.
type RealServer struct {
	Address      net.IP
	Port         uint32
	Weight       int
	ActiveConn   int
	InactiveConn int
}

// String return real server addr:port
func (rs *RealServer) String() string {
	return net.JoinHostPort(rs.Address.String(), strconv.Itoa(int(rs.Port)))
}

// Equal check the equality of real server.
// We don't use struct == since it doesn't work because of slice.
func (rs *RealServer) Equal(other *RealServer) bool {
	return rs.Address.Equal(other.Address) &&
		rs.Port == other.Port
}

// GetKernelVersionAndIPVSMods returns the linux kernel version and the required ipvs modules
func GetKernelVersionAndIPVSMods(executor exec.Interface) (kernelVersion string, ipvsModules []string, err error) {
	kernelVersionFile := "/proc/sys/kernel/osrelease"
	out, err := executor.Command("cut", "-f1", "-d", " ", kernelVersionFile).CombinedOutput()
	if err != nil {
		return "", nil, fmt.Errorf("error getting os release kernel version: %v(%s)", err, out)
	}
	kernelVersion = strings.TrimSpace(string(out))
	// parse kernel version
	ver1, err := version.ParseGeneric(kernelVersion)
	if err != nil {
		return kernelVersion, nil, fmt.Errorf("error parsing kernel version: %v(%s)", err, kernelVersion)
	}
	// "nf_conntrack_ipv4" has been removed since v4.19
	// see https://github.com/torvalds/linux/commit/a0ae2562c6c4b2721d9fddba63b7286c13517d9f
	ver2, _ := version.ParseGeneric("4.19")
	// get required ipvs modules
	if ver1.LessThan(ver2) {
		ipvsModules = append(ipvsModules, KernelModuleIPVS, KernelModuleIPVSRR, KernelModuleIPVSWRR, KernelModuleIPVSSH, KernelModuleNfConntrackIPV4)
	} else {
		ipvsModules = append(ipvsModules, KernelModuleIPVS, KernelModuleIPVSRR, KernelModuleIPVSWRR, KernelModuleIPVSSH, KernelModuleNfConntrack)
	}

	return kernelVersion, ipvsModules, nil
}
