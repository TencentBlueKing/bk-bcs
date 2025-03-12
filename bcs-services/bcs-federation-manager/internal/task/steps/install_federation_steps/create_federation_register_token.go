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

// Package steps include all steps for federation manager
package steps

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// CreateRegisterToken step name for create cluster
	CreateRegisterTokenName = fedsteps.StepNames{
		Alias: "create federation register token which is used for register to federation cluster",
		Name:  "CREATE_REGISTERTOKEN",
	}
)

// NewCreateRegisterToken sum step
func NewCreateRegisterToken() *CreateRegisterToken {
	return &CreateRegisterToken{}
}

// CreateRegisterToken sum step
type CreateRegisterToken struct{}

// Alias step name
func (s CreateRegisterToken) Alias() string {
	return CreateRegisterTokenName.Alias
}

// GetName step name
func (s CreateRegisterToken) GetName() string {
	return CreateRegisterTokenName.Name
}

// DoWork for worker exec task
func (s CreateRegisterToken) DoWork(t *types.Task) error {
	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	clusterId, ok := t.GetCommonParams(fedsteps.ClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.ClusterIdKey)
	}

	// check if bootstrap secret already exist
	obj, err := cluster.GetClusterClient().GetBootstrapSecret(&cluster.ResourceGetOptions{
		ClusterId: clusterId,
		Namespace: fedsteps.BootstrapTokenNamespace})
	if err == nil || obj != nil {
		return nil
	}

	tokenid, err := fedsteps.GenerateRandomStr(6)
	if err != nil {
		return err
	}
	token, err := fedsteps.GenerateRandomStr(16)
	if err != nil {
		return err
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: fedsteps.BootstrapTokenNamespace,
			Name:      "bootstrap-token-" + tokenid,
		},
		Type: corev1.SecretTypeBootstrapToken,
		StringData: map[string]string{
			fedsteps.BootstrapTokenIdKey:     tokenid,
			fedsteps.BootstrapTokenSecretKey: token,
			"description":                    "The bootstrap token used by clusternet cluster registration.",
			"usage-bootstrap-authentication": "true",
			"usage-bootstrap-signing":        "true",
			"auth-extra-groups":              "system:bootstrappers:clusternet:register-cluster-token",
			"expiration":                     time.Now().AddDate(20, 0, 0).Format(time.RFC3339),
		},
	}

	// create secret, if secret already exist, ignore it
	if err := cluster.GetClusterClient().CreateSecret(secret, &cluster.ResourceCreateOptions{
		ClusterId: clusterId,
		Namespace: fedsteps.BootstrapTokenNamespace,
	}); err != nil {
		return err
	}

	blog.Infof("taskId: %s, taskType: %s, taskName: %s result: %v\n", t.GetTaskID(), t.GetTaskType(), step.GetName(), fedsteps.Success)
	return nil
}

// BuildStep build step
func (s CreateRegisterToken) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
