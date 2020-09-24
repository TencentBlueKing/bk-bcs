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

package k8s

import (
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/options"
	v1 "github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v1"
	listers "github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/client/listers/bk-bcs/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// DbPrivConfInject implements K8sInject
type DbPrivConfInject struct {
	BcsDbPrivConfigLister listers.BcsDbPrivConfigLister
	Injects               options.InjectOptions
	DbPrivSecret          *corev1.Secret
}

// NewDbPrivConfInject create DbPrivConfInject object
func NewDbPrivConfInject(bcsDbPrivConfLister listers.BcsDbPrivConfigLister, injects options.InjectOptions, dbPrivSecret *corev1.Secret) K8sInject { // nolint
	k8sInject := &DbPrivConfInject{
		BcsDbPrivConfigLister: bcsDbPrivConfLister,
		Injects:               injects,
		DbPrivSecret:          dbPrivSecret,
	}

	return k8sInject
}

// InjectContent inject db privilege init-container
func (dbPrivConf *DbPrivConfInject) InjectContent(pod *corev1.Pod) ([]PatchOperation, error) {
	var patch []PatchOperation

	bcsDbPrivConfs, err := dbPrivConf.BcsDbPrivConfigLister.BcsDbPrivConfigs(pod.Namespace).List(labels.Everything())
	if err != nil {
		blog.Errorf("list BcsDbPrivConfig error %s", err.Error())
		return nil, err
	}

	var matchedBdpcs []*v1.BcsDbPrivConfig
	for _, d := range bcsDbPrivConfs {

		labelSelector := &metav1.LabelSelector{
			MatchLabels: d.Spec.PodSelector,
		}
		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return nil, fmt.Errorf("invalid label selector: %s", err.Error())
		}
		if selector.Matches(labels.Set(pod.Labels)) {
			matchedBdpcs = append(matchedBdpcs, d)
		}
	}
	if len(matchedBdpcs) > 0 {
		initContainers := pod.Spec.InitContainers
		for i, matched := range matchedBdpcs {
			initContainers = append(initContainers, dbPrivConf.generateInitContainer(i, matched))
		}
		patch = append(patch, PatchOperation{
			Op:    "replace",
			Path:  "/spec/initContainers",
			Value: initContainers,
		})
	}

	return patch, nil
}

// generateInitContainer generate an init-container with BcsDbPrivConfig
func (dbPrivConf *DbPrivConfInject) generateInitContainer(i int, matched *v1.BcsDbPrivConfig) corev1.Container {
	var fieldPath, callType string

	if dbPrivConf.Injects.DbPriv.NetworkType == "overlay" {
		fieldPath = "status.hostIP"
	} else if dbPrivConf.Injects.DbPriv.NetworkType == "underlay" {
		fieldPath = "status.podIP"
	}

	if matched.Spec.DbType == "mysql" {
		callType = "mysql_ignoreCC"
	} else if matched.Spec.DbType == "spider" {
		callType = "spider_ignoreCC"
	}

	initContainer := corev1.Container{
		Name:  "db-privilege" + "-" + strconv.Itoa(i),
		Image: dbPrivConf.Injects.DbPriv.InitContainerImage,
		Env: []corev1.EnvVar{
			{
				Name: "io_tencent_bcs_privilege_ip",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: fieldPath,
					},
				},
			},
			{
				Name:  "io_tencent_bcs_esb_url",
				Value: dbPrivConf.Injects.DbPriv.EsbUrl,
			},
			{
				Name:  "io_tencent_bcs_app_code",
				Value: string(dbPrivConf.DbPrivSecret.Data["sdk-appCode"][:]),
			},
			{
				Name:  "io_tencent_bcs_app_secret",
				Value: string(dbPrivConf.DbPrivSecret.Data["sdk-appSecret"]),
			},
			{
				Name:  "io_tencent_bcs_app_operator",
				Value: string(dbPrivConf.DbPrivSecret.Data["sdk-operator"]),
			},
			{
				Name:  "io_tencent_bcs_db_privilege_app_name",
				Value: matched.Spec.AppName,
			},
			{
				Name:  "io_tencent_bcs_db_privilege_target",
				Value: matched.Spec.TargetDb,
			},
			{
				Name:  "io_tencent_bcs_db_privilege_db_name",
				Value: matched.Spec.DbName,
			},
			{
				Name:  "io_tencent_bcs_db_privilege_call_user",
				Value: matched.Spec.CallUser,
			},
			{
				Name:  "io_tencent_bcs_db_privilege_db_type",
				Value: callType,
			},
		},
	}

	return initContainer
}
