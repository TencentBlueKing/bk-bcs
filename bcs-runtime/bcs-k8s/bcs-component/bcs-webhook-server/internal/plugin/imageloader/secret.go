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

package imageloader

import (
	"context"
	"fmt"
	"reflect"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
)

func injectImagePullSecrets(lister corev1lister.SecretLister, k8sClient kubernetes.Interface,
	namespace string, job *batchv1.Job, template corev1.PodTemplateSpec) error {
	for _, s := range template.Spec.ImagePullSecrets {
		systemSecretName, err := handleImagePullSecret(lister, k8sClient, namespace, s.Name)
		if err != nil {
			klog.Infof("imageloader handle secret failed: %v", err)
			return err
		}
		job.Spec.Template.Spec.ImagePullSecrets = append(job.Spec.Template.Spec.ImagePullSecrets,
			corev1.LocalObjectReference{Name: systemSecretName})
	}
	return nil
}

func handleImagePullSecret(lister corev1lister.SecretLister, k8sClient kubernetes.Interface,
	namespace, secretName string) (string, error) {
	originSecret, err := lister.Secrets(namespace).Get(secretName)
	if err != nil {
		return "", fmt.Errorf("failed to get origin secret %s/%s: %v", namespace, secretName, err)
	}

	// create if not exsit
	systemSecretName := generateSystemSecretName(namespace, secretName)
	systemSecret, err := lister.Secrets(pluginName).Get(systemSecretName)
	if errors.IsNotFound(err) {
		systemSecret = generateSystemSecret(systemSecretName, originSecret.Data)
		_, createErr := k8sClient.CoreV1().Secrets(pluginName).Create(context.Background(),
			systemSecret, metav1.CreateOptions{})
		if createErr != nil {
			return "", fmt.Errorf("failed to create system secret %s/%s: %v",
				pluginName, systemSecretName, createErr)
		}
		return systemSecretName, nil
	}

	if err != nil {
		return "", fmt.Errorf("failed to get system secret %s: %v", systemSecretName, err)
	}

	// update if different
	if !reflect.DeepEqual(originSecret.Data, systemSecret.Data) {
		systemSecret = generateSystemSecret(systemSecretName, originSecret.Data)
		_, updateErr := k8sClient.CoreV1().Secrets(pluginName).Update(context.Background(),
			systemSecret, metav1.UpdateOptions{})
		if updateErr != nil {
			return "", fmt.Errorf("failed to update system secret %s/%s: %v",
				pluginName, systemSecretName, updateErr)
		}
		return systemSecretName, nil
	}

	return systemSecretName, nil
}

func generateSystemSecretName(namespace, secretName string) string {
	name := fmt.Sprintf("%s-%s", namespace, secretName)
	if len(name) >= 64 {
		name = name[:64]
	}
	return name
}

func generateSystemSecret(name string, data map[string][]byte) *corev1.Secret {
	systemSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: pluginName,
			Name:      name,
		},
		Data: data,
		Type: corev1.SecretTypeDockerConfigJson,
	}
	return systemSecret
}
