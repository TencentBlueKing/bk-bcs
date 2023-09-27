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

// Package option include controller options from command line and env
package option

import (
	"fmt"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/pkg/errors"
	"path"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
)

var (
	// GlobalGitopsOpt gitops opt
	GlobalGitopsOpt *store.Options

	// consulScheme consul scheme
	consulScheme *tfexec.BackendConfigOption
	// consulAddress consul address
	consulAddress *tfexec.BackendConfigOption
	// consulPathPrefix consul path prefix(bcs-terraform-controller)
	consulPathPrefix string
	// vault private ca dir
	vaultCaPath string
)

// ControllerOption options for controller
type ControllerOption struct {
	// Address for server
	Address string

	// PodIPs contains ipv4 and ipv6 address get from status.podIPs
	PodIPs []string

	// Port port for server
	Port int

	// MetricPort port for metric server
	MetricPort int

	// LogConfig for blog
	conf.LogConfig

	// KubernetesQPS the qps of k8s client request
	KubernetesQPS int

	// KubernetesBurst the burst of k8s client request
	KubernetesBurst int

	// ConsulScheme consul协议(http)
	ConsulScheme string
	// ConsulAddress consul地址
	ConsulAddress string
	// ConsulPath consul前缀
	ConsulPath string

	// GitopsHost gitops host
	GitopsHost string
	// GitopsUsername gitops username
	GitopsUsername string
	// GitopsPassword gitops password
	GitopsPassword string

	// vault私有证书路径，通过secret挂载到pod中
	VaultCaPath string
}

// CheckControllerOption ControllerOption参数校验
func CheckControllerOption(o *ControllerOption) error {
	if len(o.ConsulScheme) == 0 {
		return errors.New("controller start param consul_scheme is nil")
	}
	if len(o.ConsulAddress) == 0 {
		return errors.New("controller start param consul_address is nil")
	}
	if len(o.ConsulPath) == 0 {
		return errors.New("controller start param consul_path is nil")
	}
	if len(o.GitopsHost) == 0 {
		return errors.New("controller start param gitops_host is nil")
	}
	if len(o.GitopsUsername) == 0 {
		return errors.New("controller start param gitops_username is nil")
	}
	if len(o.GitopsPassword) == 0 {
		return errors.New("controller start param gitops_password is nil")
	}

	GlobalGitopsOpt = &store.Options{
		Service: o.GitopsHost,
		User:    o.GitopsUsername,
		Pass:    o.GitopsPassword,
	}
	consulPathPrefix = o.ConsulPath
	consulScheme = tfexec.BackendConfig(fmt.Sprintf("scheme=%s", o.ConsulScheme))
	consulAddress = tfexec.BackendConfig(fmt.Sprintf("address=%s", o.ConsulAddress))
	vaultCaPath = o.VaultCaPath

	return nil
}

// GetConsulScheme return consul scheme
func GetConsulScheme() *tfexec.BackendConfigOption {
	return consulScheme
}

// GetConsulAddress return consul address
func GetConsulAddress() *tfexec.BackendConfigOption {
	return consulAddress
}

// GetConsulPath return consul path
func GetConsulPath(namespace, name string) *tfexec.BackendConfigOption {
	return tfexec.BackendConfig(fmt.Sprintf("path=%s", path.Join(consulPathPrefix, namespace, name)))
}

// GetVaultCaPath return vault ca path
func GetVaultCaPath() string {
	return vaultCaPath
}
