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
	"path"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	secretop "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/pkg/secret"
)

var (
	// consulScheme consul scheme
	consulScheme *tfexec.BackendConfigOption
	// consulAddress consul address
	consulAddress *tfexec.BackendConfigOption
	// consulPathPrefix consul path prefix(bcs-terraform-controller)
	consulPathPrefix string
)

// ControllerOption options for controller
type ControllerOption struct {
	conf.FileConfig
	conf.LogConfig

	Address    string `json:"address" value:"0.0.0.0" usage:"local address"`
	Port       int    `json:"port" value:"8080" usage:"http port"`
	MetricPort int    `json:"metricPort" value:"8081" usage:"metric port"`
	GRPCPort   int    `json:"grpcPort" value:"8082" usage:"grpc port"`

	EnableLeaderElection bool `json:"enableLeaderElection" value:"true" usage:"whether enable leader election"`
	KubernetesQPS        int  `json:"kubernetesQPS" value:"50" usage:"kubernetes qps"`
	KubernetesBurst      int  `json:"kubernetesBurst" value:"100" usage:"kubernetes burst"`

	ConsulScheme  string `json:"consulScheme" value:"" usage:"the scheme of consul"`
	ConsulAddress string `json:"consulAddress" value:"" usage:"the address of consul"`
	ConsulPath    string `json:"consulPath" value:"" usage:"the path of consul"`

	SecretType      string `json:"secretType" value:"" usage:"the type of secret"`
	SecretEndpoints string `json:"secretEndpoints" value:"" usage:"the endpoints of secret"`
	SecretToken     string `json:"secretToken" value:"" usage:"the token of secret"`
	SecretCA        string `json:"secretCA" value:"" usage:"the ca path of secret"`

	WorkerQueue     int    `json:"workerQueue" value:"" usage:"the queue number of worker"`
	WorkerNamespace string `json:"workerNamespace" value:"" usage:"the namespace of worker"`
	WorkerName      string `json:"workerName" value:"" usage:"the name of worker statefulset"`

	ControllerGRPCAddress string `json:"controllerGRPCAddress" value:"" usage:"the grpc address of controller manager"`
	ArgoAdminNamespace    string `json:"argoAdminNamespace" value:"" usage:"the admin namespace of argo"`
	IsWorker              bool   `json:"isWorker" value:"false" usage:"whether is worker"`
}

var (
	globalOption  *ControllerOption
	secretManager secret.SecretManagerWithVersion
)

// Parse the ControllerOperation from config file
func Parse() error {
	globalOption = &ControllerOption{}
	conf.Parse(globalOption)
	consulPathPrefix = globalOption.ConsulPath
	consulScheme = tfexec.BackendConfig(fmt.Sprintf("scheme=%s", globalOption.ConsulScheme))
	consulAddress = tfexec.BackendConfig(fmt.Sprintf("address=%s", globalOption.ConsulAddress))
	secretManager = secret.NewSecretManager(&secretop.Options{
		Secret: secretop.SecretOptions{
			CA:        globalOption.SecretCA,
			Type:      globalOption.SecretType,
			Endpoints: globalOption.SecretEndpoints,
			Token:     globalOption.SecretToken,
		},
	})
	if err := secretManager.Init(); err != nil {
		return errors.Wrapf(err, "secret manager init failed")
	}
	return nil
}

// GlobalOption return the global option
func GlobalOption() *ControllerOption {
	return globalOption
}

// GetSecretManager return the SecretManager instance
func GetSecretManager() secret.SecretManagerWithVersion {
	return secretManager
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
func GetConsulPath(namespace, name, uid string) *tfexec.BackendConfigOption {
	return tfexec.BackendConfig(fmt.Sprintf("path=%s", path.Join(consulPathPrefix, namespace, name, uid)))
}
