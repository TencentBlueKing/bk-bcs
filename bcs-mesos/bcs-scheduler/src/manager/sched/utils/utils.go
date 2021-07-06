/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"fmt"
	"reflect"
	"strings"

	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	types "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

// will ignore label key begin with "io.tencent"
func isLabelsChanged(newLabels, oldLabels map[string]string) error {
	if len(newLabels) != len(oldLabels) {
		return fmt.Errorf("length of labels changed")
	}
	for key, value := range oldLabels {
		if strings.HasPrefix(key, "io.tencent.") {
			continue
		}
		newValue, ok := newLabels[key]
		if !ok {
			return fmt.Errorf("key %s 's value missing", key)
		}
		if newValue != value {
			return fmt.Errorf("key %s 's value changed", key)
		}
	}
	return nil
}

func isDockerEnvChanged(oldEnv, newEnv map[string]string) error {
	if len(oldEnv) != len(newEnv) {
		return fmt.Errorf("Env cannot be changed")
	}
	for k, v := range oldEnv {
		newV, ok := newEnv[k]
		if !ok || newV != v {
			return fmt.Errorf("Env cannot be changed")
		}
	}
	return nil
}

func isDockerChanged(oldDocker, newDocker *types.Docker) error {
	if oldDocker.Hostname != newDocker.Hostname {
		return fmt.Errorf("Hostname cannot be changed")
	}
	if oldDocker.Image != newDocker.Image {
		return fmt.Errorf("Image cannot be changed")
	}
	if oldDocker.ImagePullUser != newDocker.ImagePullUser {
		return fmt.Errorf("ImagePullUser cannot be changed")
	}
	if oldDocker.ImagePullPasswd != newDocker.ImagePullPasswd {
		return fmt.Errorf("ImagePullPasswd cannot be changed")
	}
	if oldDocker.Network != newDocker.Network {
		return fmt.Errorf("Network cannot be changed")
	}
	if oldDocker.NetworkType != newDocker.NetworkType {
		return fmt.Errorf("NetworkType cannot be changed")
	}
	if oldDocker.Command != newDocker.Command {
		return fmt.Errorf("Command cannot be changed")
	}
	if len(oldDocker.Arguments) != len(newDocker.Arguments) {
		return fmt.Errorf("Arguments cannot be changed")
	}
	for i, oldArg := range oldDocker.Arguments {
		newArg := newDocker.Arguments[i]
		if newArg != oldArg {
			return fmt.Errorf("Arguments cannot be changed")
		}
	}
	if len(oldDocker.Parameters) != len(newDocker.Parameters) {
		return fmt.Errorf("Parameters cannot be changed")
	}
	for i, oldParam := range oldDocker.Parameters {
		newParam := newDocker.Parameters[i]
		if !reflect.DeepEqual(newParam, oldParam) {
			return fmt.Errorf("Parameters cannot be changed")
		}
	}
	if len(oldDocker.PortMappings) != len(newDocker.PortMappings) {
		return fmt.Errorf("PortMappings cannot be changed")
	}
	for i, oldPort := range oldDocker.PortMappings {
		newPort := newDocker.PortMappings[i]
		if !reflect.DeepEqual(newPort, oldPort) {
			return fmt.Errorf("PortMappings cannot be changed")
		}
	}
	if oldDocker.Privileged != newDocker.Privileged {
		return fmt.Errorf("Privileged cannot be changed")
	}
	if err := isDockerEnvChanged(oldDocker.Env, newDocker.Env); err != nil {
		return err
	}

	return nil
}

func isContainerVolumeChanged(oldVolumes, newVolumes []*types.Volume) error {
	if len(oldVolumes) != len(newVolumes) {
		return fmt.Errorf("volumes of cannot be changed")
	}
	for j, oldVolume := range oldVolumes {
		if !reflect.DeepEqual(oldVolume, newVolumes[j]) {
			return fmt.Errorf("volumes %d cannot be changed", j)
		}
	}
	return nil
}

func isContainerConfigMapsChanged(oldConfigMaps, newConfigMaps []commtypes.ConfigMap) error {
	if len(oldConfigMaps) != len(newConfigMaps) {
		return fmt.Errorf("configmaps cannot be changed")
	}
	for j, oldConfigMap := range oldConfigMaps {
		if oldConfigMap.Name != newConfigMaps[j].Name || len(oldConfigMap.Items) != len(newConfigMaps[j].Items) {
			return fmt.Errorf("configmaps %d cannot be changed", j)
		}
		if len(oldConfigMap.Items) == 0 && len(newConfigMaps[j].Items) == 0 {
			return nil
		}
		if !reflect.DeepEqual(oldConfigMap, newConfigMaps[j]) {
			return fmt.Errorf("configmaps %d cannot be changed", j)
		}
	}
	return nil
}

func isContainerSecretChanged(oldSecrets, newSecrets []commtypes.Secret) error {
	if len(newSecrets) != len(oldSecrets) {
		return fmt.Errorf("secret cannot be empty")
	}
	for j, oldSecret := range oldSecrets {
		if oldSecret.SecretName != newSecrets[j].SecretName ||
			len(oldSecret.Items) != len(newSecrets[j].Items) {
			return fmt.Errorf("secret cannot be empty")
		}
		if len(oldSecret.Items) == 0 && len(newSecrets[j].Items) == 0 {
			return nil
		}
		if !reflect.DeepEqual(oldSecret, newSecrets[j]) {
			return fmt.Errorf("secrets %d cannot be changed", j)
		}
	}
	return nil
}

func isContainerChanged(oldContainer, newContainer *types.Container) error {
	if oldContainer.Type != newContainer.Type {
		return fmt.Errorf("container type cannot be changed")
	}
	if err := isDockerChanged(oldContainer.Docker, newContainer.Docker); err != nil {
		return fmt.Errorf("docker parameter cannot be changed, err %s", err.Error())
	}
	if err := isContainerVolumeChanged(oldContainer.Volumes, newContainer.Volumes); err != nil {
		return fmt.Errorf("Volumes of cannot be changed")
	}
	if err := isContainerConfigMapsChanged(oldContainer.ConfigMaps, newContainer.ConfigMaps); err != nil {
		return fmt.Errorf("Configmaps cannot be changed")
	}
	if err := isContainerSecretChanged(oldContainer.Secrets, newContainer.Secrets); err != nil {
		return fmt.Errorf("Secrets cannot be changed")
	}
	if len(oldContainer.HealthChecks) != len(newContainer.HealthChecks) {
		return fmt.Errorf("health checks cannot be changed")
	}
	for j, oldHealth := range oldContainer.HealthChecks {
		newHealth := oldContainer.HealthChecks[j]
		if !reflect.DeepEqual(oldHealth, newHealth) {
			return fmt.Errorf("health checks %d cannot be changed", j)
		}
	}
	if !reflect.DeepEqual(oldContainer.NetLimit, newContainer.NetLimit) {
		return fmt.Errorf("netlimit cannot be changed")
	}

	if oldContainer.Resources == nil || newContainer.Resources == nil {
		return fmt.Errorf("resources cannot be empty")
	}
	if oldContainer.LimitResoures == nil || newContainer.LimitResoures == nil {
		return fmt.Errorf("limit resources cannot be empty")
	}
	return nil
}

// IsOnlyResourceIncreased to check if there is only resources increased
func IsOnlyResourceIncreased(old, new *types.Version) error {
	if old == nil || new == nil {
		return fmt.Errorf("version cannot be empty")
	}
	if old.PodObjectMeta.Name != new.PodObjectMeta.Name {
		return fmt.Errorf("name of pod object meta cannot be changed")
	}
	if old.PodObjectMeta.NameSpace != new.PodObjectMeta.NameSpace {
		return fmt.Errorf("namespace of pod object meta cannot be changed")
	}
	if err := isLabelsChanged(old.PodObjectMeta.Labels, new.PodObjectMeta.Labels); err != nil {
		return err
	}
	if old.Instances != new.Instances {
		return fmt.Errorf("instances cannot be changed")
	}
	if old.RunAs != new.RunAs {
		return fmt.Errorf("namespace cannot be changed")
	}
	if err := isLabelsChanged(old.Labels, new.Labels); err != nil {
		return fmt.Errorf("labels cannot be changed")
	}
	if !reflect.DeepEqual(old.KillPolicy, new.KillPolicy) {
		return fmt.Errorf("kill policy cannot be changed")
	}
	if !reflect.DeepEqual(old.RestartPolicy, new.RestartPolicy) {
		return fmt.Errorf("restartPolicy cannot be changed")
	}
	if !reflect.DeepEqual(old.Constraints, new.Constraints) {
		return fmt.Errorf("constraints cannot be changed")
	}
	if old.Kind != new.Kind {
		return fmt.Errorf("kind cannot be changed")
	}
	if len(old.Container) != len(new.Container) {
		return fmt.Errorf("container number cannot be changed")
	}

	for i := range old.Container {
		oldContainer := old.Container[i]
		newContainer := new.Container[i]
		if err := isContainerChanged(oldContainer, newContainer); err != nil {
			return fmt.Errorf("container %d: %s", i, err.Error())
		}
	}
	return nil
}
