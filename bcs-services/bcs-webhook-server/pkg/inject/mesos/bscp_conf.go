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

package mesos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"bk-bcs/bcs-common/common/blog"
	commtypes "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-services/bcs-webhook-server/pkg/inject/common/bscp"
)

// BscpInject implements MesosInject
type BscpInject struct {
	//template containers
	temContainers []commtypes.Container
}

// NewBscpInject new BscpInject object
func NewBscpInject() *BscpInject {
	return &BscpInject{}
}

// InitTemplate load template from file
func (bi *BscpInject) InitTemplate(templatePath string) error {
	by, err := ioutil.ReadFile(templatePath)
	if err != nil {
		blog.Errorf("bscp load template file %s failed, err %s", templatePath, err.Error())
		return fmt.Errorf("bscp load template file %s failed, err %s", templatePath, err.Error())
	}

	err = json.Unmarshal(by, &bi.temContainers)
	if err != nil {
		//template format err, then exit
		blog.Errorf("bscp Unmarshal template file %s error %s", templatePath, err.Error())
		return fmt.Errorf("bscp Unmarshal template file %s error %s", templatePath, err.Error())
	}

	if len(bi.temContainers) == 0 {
		blog.Errorf("bscp init template %s failed, No template information found", templatePath)
		return fmt.Errorf("bscp init template %s failed, No template information found", templatePath)
	}
	return nil
}

// checkAnnotations check if deployment/application has bscp inject annotations
func checkAnnotations(typeMeta commtypes.TypeMeta, objMeta commtypes.ObjectMeta) bool {
	// check annotation for SideCar injection
	annotation := objMeta.GetAnnotations()
	if annotation == nil {
		blog.Infof("MetaType %s: %s/%s Annotation is empty, skip bscp-SideCar injection",
			typeMeta.Kind, objMeta.GetNamespace(), objMeta.GetName())
		return false
	}
	v, ok := annotation[bscp.AnnotationKey]
	if !ok {
		// check annotation, if Deployment or Application do not mark ,
		// bscp mesos injector do nothing
		blog.Infof("MetaType %s: %s/%s Annotation find no specified annotation, skip bscp-SideCar injection",
			typeMeta.Kind, objMeta.GetNamespace(), objMeta.GetName())
		return false
	}
	if v != bscp.AnnotationValue {
		// check annotation, if Deployment or Application do not mark ,
		// bscp mesos injector do nothing
		blog.Warnf("MetaType %s: %s/%s Do not need SideCar injection, skip bscp-SideCar injection",
			typeMeta.Kind, objMeta.GetNamespace(), objMeta.GetName())
		return false
	}
	return true
}

// InjectApplicationContent inject SideCar into mesos application
func (bi *BscpInject) InjectApplicationContent(application *commtypes.ReplicaController) (*commtypes.ReplicaController, error) {
	// if get no bscp inject annotations, just return original content
	if !checkAnnotations(application.TypeMeta, application.ObjectMeta) {
		return application, nil
	}
	name := fmt.Sprintf("%s/%s", application.GetNamespace(), application.GetName())

	// retrieve ENV for SideCar setup
	// when retrieve ENV failed, return error
	retrievedContainers, envMap, envErr := bi.retrieveEnvFromContainer(
		application.ReplicaControllerSpec.Template.PodSpec.Containers, name)
	if envErr != nil {
		blog.Errorf("bscp retrieve specified Environment for App %s failed, %s", name, envErr.Error())
		return nil, fmt.Errorf("bscp retrieve specified Environment for App %s failed, %s", name, envErr.Error())
	}

	//  append envMap to template Container
	containers := bi.injectEnvToContainer(bi.temContainers, envMap)
	containers = append(containers, retrievedContainers...)
	application.ReplicaControllerSpec.Template.PodSpec.Containers = containers
	blog.Infof("bscp inject bscp-mesos-SideCar for Application %s successfully", name)
	return application, nil
}

// InjectDeployContent inject SideCar into mesos deployment
func (bi *BscpInject) InjectDeployContent(deploy *commtypes.BcsDeployment) (*commtypes.BcsDeployment, error) {
	// if get no bscp inject annotations, just return original content
	if !checkAnnotations(deploy.TypeMeta, deploy.ObjectMeta) {
		return deploy, nil
	}
	name := fmt.Sprintf("%s/%s", deploy.GetNamespace(), deploy.GetName())

	// retrieve ENV for SideCar setup
	// when retrieve ENV failed, return error
	retrievedContainers, envMap, envErr := bi.retrieveEnvFromContainer(
		deploy.Spec.Template.PodSpec.Containers, name)
	if envErr != nil {
		blog.Errorf("bscp retrieve specified Environment for Deployment %s failed, %s", name, envErr.Error())
		return nil, fmt.Errorf("bscp retrieve specified Environment for Deployment %s failed, %s", name, envErr.Error())
	}

	containers := bi.injectEnvToContainer(bi.temContainers, envMap)
	containers = append(containers, retrievedContainers...)
	deploy.Spec.Template.PodSpec.Containers = containers
	blog.Infof("bscp inject bscp-mesos-SideCar for Deployment %s successfully", name)
	return deploy, nil
}

func (bi *BscpInject) retrieveEnvFromContainer(containers []commtypes.Container, name string) ([]commtypes.Container, map[string]string, error) {
	envMap := make(map[string]string)
	var retContainers []commtypes.Container
	for _, c := range containers {
		for _, env := range c.Env {
			if strings.Contains(env.Name, bscp.SideCarPrefix) {
				envMap[env.Name] = env.Value
				blog.Infof("Injection for %s [%s=%s]", name, env.Name, env.Value)
			}
			//check specified directory for share within pod
			if env.Name == bscp.SideCarCfgPath {
				v := commtypes.VolumeUnit{
					Name: bscp.SideCarVolumeName,
					Volume: commtypes.Volume{
						MountPath: env.Value,
						ReadOnly:  false,
					},
				}
				blog.Infof("Injection for shared directory: %v", v)
				c.Volumes = append(c.Volumes, v)
			}
		}
		retContainers = append(retContainers, c)
	}
	cfgPath, ok := envMap[bscp.SideCarCfgPath]
	if !ok {
		return nil, nil, fmt.Errorf("bscp SideCar environment lost BSCP_BCSSIDECAR_APPCFG_PATH")
	}
	if len(cfgPath) == 0 {
		return nil, nil, fmt.Errorf("bscp SideCar environment BSCP_BCSSIDECAR_APPCFG_PATH is empty")
	}
	if modValue, ok := envMap[bscp.SideCarMod]; !ok {
		// for single app
		if !bscp.ValidateEnvs(envMap) {
			return nil, nil, fmt.Errorf("bscp SideCar envs are invalid")
		}
	} else {
		// for multiple app, inject BSCP_BCSSIDECAR_APPCFG_PATH into every app path
		// if BSCP_BCSSideCar_APPINFO_MOD's value is invlaid, cannot do inject;
		modValue, err := bscp.AddPathIntoAppInfoMode(modValue, cfgPath)
		if err != nil {
			return nil, nil, fmt.Errorf("env %s:%s is invalid", bscp.SideCarMod, modValue)
		}
		envMap[bscp.SideCarMod] = modValue
	}
	return retContainers, envMap, nil
}

func (bi *BscpInject) injectEnvToContainer(tempContainers []commtypes.Container, envs map[string]string) []commtypes.Container {
	var injectContainers []commtypes.Container
	for _, container := range tempContainers {
		//inject environments
		for key, value := range envs {
			env := commtypes.EnvVar{
				Name:  key,
				Value: value,
			}
			container.Env = append(container.Env, env)
			//check specified directory for share within pod
			if key == bscp.SideCarCfgPath {
				v := commtypes.VolumeUnit{
					Name: bscp.SideCarVolumeName,
					Volume: commtypes.Volume{
						MountPath: value,
						ReadOnly:  false,
					},
				}
				container.Volumes = append(container.Volumes, v)
			}
		}
		injectContainers = append(injectContainers, container)
	}

	return injectContainers
}
