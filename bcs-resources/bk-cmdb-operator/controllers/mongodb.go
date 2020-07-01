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

package controllers

import (
	"fmt"

	bkcmdbv1 "github.com/Tencent/bk-bcs/bcs-resources/bk-cmdb-operator/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	defaultMongoDatabase = "cmdb"
	defaultMongoUsername = "cc"
	defaultMongoPassword = "cc"
	defaultMongoPort     = 27017
)

var mongoDatabase, mongoUsername, mongoPassword string
var mongoPort int32

// reconcileMongoDb reconciles bk-cmdb mongodb
func (r *BkcmdbReconciler) reconcileMongoDb(instance *bkcmdbv1.Bkcmdb) error {
	secret := makeMongoSecret(instance)
	if err := controllerutil.SetControllerReference(instance, secret, r.Scheme); err != nil {
		return fmt.Errorf("failed to set mongodb secret owner reference: %s", err.Error())
	}
	err := r.Client.CreateOrUpdateSecret(secret)
	if err != nil {
		return fmt.Errorf("failed to create or update mongodb secret: %s", err.Error())
	}

	deploy := makeMongoDeploy(instance)
	if err := controllerutil.SetControllerReference(instance, deploy, r.Scheme); err != nil {
		return fmt.Errorf("failed to set mongodb deploy owner reference: %s", err.Error())
	}

	err = r.Client.CreateOrUpdateDeploy(deploy)
	if err != nil {
		return fmt.Errorf("failed to create or update mongodb deploy: %s", err.Error())
	}

	service := makeMongoService(instance)
	if err := controllerutil.SetControllerReference(instance, service, r.Scheme); err != nil {
		return fmt.Errorf("failed to set mongodb service owner reference: %s", err.Error())
	}

	err = r.Client.CreateOrUpdateService(service)
	if err != nil {
		return fmt.Errorf("failed to create or update mongodb service: %s", err.Error())
	}

	return nil
}

// makeMongoDeploy builds mongodb deployment object
func makeMongoDeploy(z *bkcmdbv1.Bkcmdb) *appsv1.Deployment {
	replicas := int32(1)
	fsGroup := int64(1001)

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-mongodb",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "mongodb",
				"release": z.GetName(),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     "mongodb",
					"release": z.GetName(),
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":     "mongodb",
						"release": z.GetName(),
					},
				},
				Spec: v1.PodSpec{
					SecurityContext: &v1.PodSecurityContext{
						FSGroup: &fsGroup,
					},
					InitContainers: makeMongoInitContainer(),
					Containers:     makeMongoContainers(z),
					Volumes: []v1.Volume{
						{
							Name: "data",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
}

// makeMongoInitContainer builds mongodb init-container object
func makeMongoInitContainer() []v1.Container {
	return []v1.Container{
		{
			Name:            "volume-mount-hack",
			Image:           "busybox:latest",
			ImagePullPolicy: "IfNotPresent",
			Command:         []string{"sh", "-c", "chown -R 1001:1001 /bitnami/"},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "data",
					MountPath: "/bitnami/",
				},
			},
		},
	}
}

// makeMongoContainers build mongodb containers object
func makeMongoContainers(z *bkcmdbv1.Bkcmdb) []v1.Container {
	runAsUser := int64(1001)
	runAsNonRoot := true

	mongoUsername = defaultMongoUsername
	mongoDatabase = defaultMongoDatabase

	return []v1.Container{
		{
			Name:            z.GetName() + "-mongodb",
			Image:           "docker.io/bitnami/mongodb:4.0.6",
			ImagePullPolicy: "IfNotPresent",
			SecurityContext: &v1.SecurityContext{
				RunAsUser:    &runAsUser,
				RunAsNonRoot: &runAsNonRoot,
			},
			Env: []v1.EnvVar{
				{
					Name: "MONGODB_ROOT_PASSWORD",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: z.GetName() + "-mongodb",
							},
							Key: "mongodb-root-password",
						},
					},
				},
				{
					Name:  "MONGODB_USERNAME",
					Value: mongoUsername,
				},
				{
					Name: "MONGODB_PASSWORD",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: z.GetName() + "-mongodb",
							},
							Key: "mongodb-password",
						},
					},
				},
				{
					Name:  "MONGODB_SYSTEM_LOG_VERBOSITY",
					Value: "0",
				},
				{
					Name:  "MONGODB_DISABLE_SYSTEM_LOG",
					Value: "no",
				},
				{
					Name:  "MONGODB_DATABASE",
					Value: mongoDatabase,
				},
				{
					Name:  "MONGODB_ENABLE_IPV6",
					Value: "yes",
				},
			},
			Ports: []v1.ContainerPort{
				{
					Name:          "mongodb",
					ContainerPort: 27017,
				},
			},
			LivenessProbe: &v1.Probe{
				Handler: v1.Handler{
					Exec: &v1.ExecAction{
						Command: []string{"mongo", "--eval", "db.adminCommand('ping')"},
					},
				},
				InitialDelaySeconds: 30,
				TimeoutSeconds:      5,
				PeriodSeconds:       10,
				SuccessThreshold:    1,
				FailureThreshold:    6,
			},
			ReadinessProbe: &v1.Probe{
				Handler: v1.Handler{
					Exec: &v1.ExecAction{
						Command: []string{"mongo", "--eval", "db.adminCommand('ping')"},
					},
				},
				InitialDelaySeconds: 5,
				TimeoutSeconds:      5,
				PeriodSeconds:       10,
				SuccessThreshold:    1,
				FailureThreshold:    6,
			},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "data",
					MountPath: "/bitnami/mongodb",
				},
			},
		},
	}
}

// makeMongoSecret builds mongodb secret object
func makeMongoSecret(z *bkcmdbv1.Bkcmdb) *v1.Secret {
	mongoPassword = defaultMongoPassword
	//base64Ps := base64.StdEncoding.EncodeToString([]byte(mongoPassword))
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-mongodb",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "mongodb",
				"release": z.GetName(),
			},
		},
		Type: "Opaque",
		Data: map[string][]byte{
			"mongodb-root-password": []byte(mongoPassword),
			"mongodb-password":      []byte(mongoPassword),
		},
	}
}

// makeMongoService build mongodb service object
func makeMongoService(z *bkcmdbv1.Bkcmdb) *v1.Service {
	mongoPort = defaultMongoPort
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-mongodb",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "mongodb",
				"release": z.GetName(),
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Name: "mongodb",
					Port: mongoPort,
					TargetPort: intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "mongodb",
					},
				},
			},
			Selector: map[string]string{
				"app":     "mongodb",
				"release": z.GetName(),
			},
		},
	}
}
