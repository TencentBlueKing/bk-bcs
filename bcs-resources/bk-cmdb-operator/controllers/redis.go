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
	defaultRedisPassword = "cmdb"
	defaultRedisPort     = 6379
)

var redisPassword string
var redisPort int32

// reconcileRedis reconcile bk-cmdb redis
func (r *BkcmdbReconciler) reconcileRedis(instance *bkcmdbv1.Bkcmdb) error {
	secret := makeRedisSecret(instance)
	if err := controllerutil.SetControllerReference(instance, secret, r.Scheme); err != nil {
		return fmt.Errorf("failed to set redis secret owner reference: %s", err.Error())
	}
	err := r.Client.CreateOrUpdateSecret(secret)
	if err != nil {
		return fmt.Errorf("failed to create or update redis secret: %s", err.Error())
	}

	cm := makeRedisCm(instance)
	if err := controllerutil.SetControllerReference(instance, cm, r.Scheme); err != nil {
		return fmt.Errorf("failed to set redis configmap owner reference: %s", err.Error())
	}
	err = r.Client.CreateOrUpdateCm(cm)
	if err != nil {
		return fmt.Errorf("failed to create or update redis configmap: %s", err.Error())
	}

	healthCm := makeRedisHealthCm(instance)
	if err := controllerutil.SetControllerReference(instance, healthCm, r.Scheme); err != nil {
		return fmt.Errorf("failed to set redis configmap owner reference: %s", err.Error())
	}
	err = r.Client.CreateOrUpdateCm(healthCm)
	if err != nil {
		return fmt.Errorf("failed to create or update redis configmap: %s", err.Error())
	}

	sa := makeRedisSa(instance)
	if err := controllerutil.SetControllerReference(instance, sa, r.Scheme); err != nil {
		return fmt.Errorf("failed to set redis serviceaccount owner reference: %s", err.Error())
	}
	err = r.Client.CreateOrUpdateSa(sa)
	if err != nil {
		return fmt.Errorf("failed to create or update redis serviceaccount: %s", err.Error())
	}

	deploy := makeRedisDeploy(instance)
	if err := controllerutil.SetControllerReference(instance, deploy, r.Scheme); err != nil {
		return fmt.Errorf("failed to set redis deployment owner reference: %s", err.Error())
	}
	err = r.Client.CreateOrUpdateDeploy(deploy)
	if err != nil {
		return fmt.Errorf("failed to create or update redis deployment: %s", err.Error())
	}

	sts := makeRedisSts(instance)
	if err := controllerutil.SetControllerReference(instance, sts, r.Scheme); err != nil {
		return fmt.Errorf("failed to set redis statefulset owner reference: %s", err.Error())
	}
	err = r.Client.CreateOrUpdateSts(sts)
	if err != nil {
		return fmt.Errorf("failed to create or update redis statefulset: %s", err.Error())
	}

	masterSvc := makeRedisMasterService(instance)
	if err := controllerutil.SetControllerReference(instance, masterSvc, r.Scheme); err != nil {
		return fmt.Errorf("failed to set redis service owner reference: %s", err.Error())
	}
	err = r.Client.CreateOrUpdateService(masterSvc)
	if err != nil {
		return fmt.Errorf("failed to create or update redis service: %s", err.Error())
	}

	slaveSvc := makeRedisSlaveService(instance)
	if err := controllerutil.SetControllerReference(instance, slaveSvc, r.Scheme); err != nil {
		return fmt.Errorf("failed to set redis service owner reference: %s", err.Error())
	}
	err = r.Client.CreateOrUpdateService(slaveSvc)
	if err != nil {
		return fmt.Errorf("failed to create or update redis service: %s", err.Error())
	}

	return nil
}

// makeRedisSecret build redis secret object
func makeRedisSecret(z *bkcmdbv1.Bkcmdb) *v1.Secret {
	redisPassword = defaultRedisPassword
	//base64Ps := base64.StdEncoding.EncodeToString([]byte(redisPassword))
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-redis",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "redis",
				"release": z.GetName(),
			},
		},
		Type: "Opaque",
		Data: map[string][]byte{
			"redis-password": []byte(redisPassword),
		},
	}
}

// makeRedisCm builds redis configmap
func makeRedisCm(z *bkcmdbv1.Bkcmdb) *v1.ConfigMap {
	redisConfContent := `# User-supplied configuration:
# maxmemory-policy volatile-lru`

	masterConfContent := `dir /data
rename-command FLUSHDB ""
rename-command FLUSHALL ""`

	replicaConfContent := `dir /data
rename-command FLUSHDB ""
rename-command FLUSHALL ""`

	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-redis",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "redis",
				"release": z.GetName(),
			},
		},
		Data: map[string]string{
			"redis.conf":   redisConfContent,
			"master.conf":  masterConfContent,
			"replica.conf": replicaConfContent,
		},
	}
}

// makeRedisHealthCm build redis health configmap
func makeRedisHealthCm(z *bkcmdbv1.Bkcmdb) *v1.ConfigMap {
	pingLocalContent := `response=$(
  redis-cli \
    -a $REDIS_PASSWORD \
    -h localhost \
    -p $REDIS_PORT \
    ping
)
if [ "$response" != "PONG" ]; then
  echo "$response"
  exit 1
fi`

	pingMasterConfContent := `response=$(
  redis-cli \
  -a $REDIS_MASTER_PASSWORD \
  -h $REDIS_MASTER_HOST \
  -p $REDIS_MASTER_PORT_NUMBER \
  ping
)
if [ "$response" != "PONG" ]; then
  echo "$response"
  exit 1
fi`

	pingLocalMasterContent := `script_dir="$(dirname "$0")"
exit_status=0
"$script_dir/ping_local.sh" || exit_status=$?
"$script_dir/ping_master.sh" || exit_status=$?
exit $exit_status`

	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-redis-health",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "redis",
				"release": z.GetName(),
			},
		},
		Data: map[string]string{
			"ping_local.sh":            pingLocalContent,
			"ping_master.sh":           pingMasterConfContent,
			"ping_local_and_master.sh": pingLocalMasterContent,
		},
	}
}

// makeRedisSa build redis serviceaccount object
func makeRedisSa(z *bkcmdbv1.Bkcmdb) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-redis",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "redis",
				"release": z.GetName(),
			},
		},
	}
}

// makeRedisDeploy build redis deployment object
func makeRedisDeploy(z *bkcmdbv1.Bkcmdb) *appsv1.Deployment {
	replicas := int32(1)
	fsGroup := int64(1001)
	runAsUser := int64(1001)
	defaultMode := int32(0755)

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-redis-slave",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "redis",
				"release": z.GetName(),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     "redis",
					"release": z.GetName(),
					"role":    "slave",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":     "redis",
						"release": z.GetName(),
						"role":    "slave",
					},
				},
				Spec: v1.PodSpec{
					SecurityContext: &v1.PodSecurityContext{
						FSGroup:   &fsGroup,
						RunAsUser: &runAsUser,
					},
					ServiceAccountName: z.GetName() + "-redis",
					Containers:         makeRedisSlaveContainers(z),
					Volumes: []v1.Volume{
						{
							Name: "health",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: z.GetName() + "-redis-health",
									},
									DefaultMode: &defaultMode,
								},
							},
						},
						{
							Name: "config",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: z.GetName() + "-redis",
									},
								},
							},
						},
						{
							Name: "redis-data",
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

// makeRedisSlaveContainers build redis-slave containers object
func makeRedisSlaveContainers(z *bkcmdbv1.Bkcmdb) []v1.Container {
	return []v1.Container{
		{
			Name:            z.GetName() + "-redis",
			Image:           "docker.io/bitnami/redis:4.0.12",
			ImagePullPolicy: "IfNotPresent",
			Command:         []string{"/run.sh"},
			Args: []string{"--port", "$(REDIS_PORT)", "--slaveof", "$(REDIS_MASTER_HOST)", "$(REDIS_MASTER_PORT_NUMBER)",
				"--requirepass", "$(REDIS_PASSWORD)", "--masterauth", "$(REDIS_MASTER_PASSWORD)", "--include",
				"/opt/bitnami/redis/etc/redis.conf", "--include", "/opt/bitnami/redis/etc/replica.conf"},
			Env: []v1.EnvVar{
				{
					Name:  "REDIS_REPLICATION_MODE",
					Value: "slave",
				},
				{
					Name:  "REDIS_MASTER_HOST",
					Value: z.GetName() + "-redis-master",
				},
				{
					Name:  "REDIS_PORT",
					Value: "6379",
				},
				{
					Name:  "REDIS_MASTER_PORT_NUMBER",
					Value: "6379",
				},
				{
					Name: "REDIS_PASSWORD",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: z.GetName() + "-redis",
							},
							Key: "redis-password",
						},
					},
				},
				{
					Name: "REDIS_MASTER_PASSWORD",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: z.GetName() + "-redis",
							},
							Key: "redis-password",
						},
					},
				},
			},
			Ports: []v1.ContainerPort{
				{
					Name:          "redis",
					ContainerPort: 6379,
				},
			},
			LivenessProbe: &v1.Probe{
				Handler: v1.Handler{
					Exec: &v1.ExecAction{
						Command: []string{"sh", "-c", "/health/ping_local_and_master.sh"},
					},
				},
				InitialDelaySeconds: 5,
				TimeoutSeconds:      5,
				PeriodSeconds:       5,
				SuccessThreshold:    1,
				FailureThreshold:    5,
			},
			ReadinessProbe: &v1.Probe{
				Handler: v1.Handler{
					Exec: &v1.ExecAction{
						Command: []string{"sh", "-c", "/health/ping_local_and_master.sh"},
					},
				},
				InitialDelaySeconds: 5,
				TimeoutSeconds:      5,
				PeriodSeconds:       5,
				SuccessThreshold:    1,
				FailureThreshold:    5,
			},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "health",
					MountPath: "/health",
				},
				{
					Name:      "redis-data",
					MountPath: "/data",
				},
				{
					Name:      "config",
					MountPath: "/opt/bitnami/redis/etc",
				},
			},
		},
	}
}

// makeRedisSts build redis-master statefulset object
func makeRedisSts(z *bkcmdbv1.Bkcmdb) *appsv1.StatefulSet {
	fsGroup := int64(1001)
	runAsUser := int64(1001)
	defaultMode := int32(0755)

	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-redis-master",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "redis",
				"release": z.GetName(),
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     "redis",
					"release": z.GetName(),
					"role":    "master",
				},
			},
			ServiceName: z.GetName() + "-redis-master",
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":     "redis",
						"release": z.GetName(),
						"role":    "master",
					},
				},
				Spec: v1.PodSpec{
					SecurityContext: &v1.PodSecurityContext{
						RunAsUser: &runAsUser,
						FSGroup:   &fsGroup,
					},
					ServiceAccountName: z.GetName() + "-redis",
					Containers:         makeRedisMasterContainers(z),
					Volumes: []v1.Volume{
						{
							Name: "health",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: z.GetName() + "-redis-health",
									},
									DefaultMode: &defaultMode,
								},
							},
						},
						{
							Name: "config",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: z.GetName() + "-redis",
									},
								},
							},
						},
						{
							Name: "redis-data",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
		},
	}
}

// makeRedisMasterContainers builds redis-master containers object
func makeRedisMasterContainers(z *bkcmdbv1.Bkcmdb) []v1.Container {
	return []v1.Container{
		{
			Name:            z.GetName() + "-redis",
			Image:           "docker.io/bitnami/redis:4.0.12",
			ImagePullPolicy: "IfNotPresent",
			Command:         []string{"/run.sh"},
			Args: []string{"--port", "$(REDIS_PORT)", "--requirepass", "$(REDIS_PASSWORD)", "--include",
				"/opt/bitnami/redis/etc/redis.conf", "--include", "/opt/bitnami/redis/etc/master.conf"},
			Env: []v1.EnvVar{
				{
					Name:  "REDIS_REPLICATION_MODE",
					Value: "master",
				},
				{
					Name:  "REDIS_PORT",
					Value: "6379",
				},
				{
					Name: "REDIS_PASSWORD",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: z.GetName() + "-redis",
							},
							Key: "redis-password",
						},
					},
				},
			},
			Ports: []v1.ContainerPort{
				{
					Name:          "redis",
					ContainerPort: 6379,
				},
			},
			LivenessProbe: &v1.Probe{
				Handler: v1.Handler{
					Exec: &v1.ExecAction{
						Command: []string{"sh", "-c", "/health/ping_local.sh"},
					},
				},
				InitialDelaySeconds: 5,
				TimeoutSeconds:      5,
				PeriodSeconds:       5,
				SuccessThreshold:    1,
				FailureThreshold:    5,
			},
			ReadinessProbe: &v1.Probe{
				Handler: v1.Handler{
					Exec: &v1.ExecAction{
						Command: []string{"sh", "-c", "/health/ping_local.sh"},
					},
				},
				InitialDelaySeconds: 5,
				TimeoutSeconds:      5,
				PeriodSeconds:       5,
				SuccessThreshold:    1,
				FailureThreshold:    5,
			},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "health",
					MountPath: "/health",
				},
				{
					Name:      "redis-data",
					MountPath: "/data",
				},
				{
					Name:      "config",
					MountPath: "/opt/bitnami/redis/etc",
				},
			},
		},
	}
}

// makeRedisMasterService build redis-master service object
func makeRedisMasterService(z *bkcmdbv1.Bkcmdb) *v1.Service {
	redisPort = defaultRedisPort
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-redis-master",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "redis",
				"release": z.GetName(),
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Name: "redis",
					Port: redisPort,
					TargetPort: intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "redis",
					},
				},
			},
			Selector: map[string]string{
				"app":     "redis",
				"release": z.GetName(),
				"role":    "master",
			},
		},
	}
}

// makeRedisSlaveService build redis-slave service object
func makeRedisSlaveService(z *bkcmdbv1.Bkcmdb) *v1.Service {
	redisPort = defaultRedisPort
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-redis-slave",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "redis",
				"release": z.GetName(),
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Name: "redis",
					Port: redisPort,
					TargetPort: intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "redis",
					},
				},
			},
			Selector: map[string]string{
				"app":     "redis",
				"release": z.GetName(),
				"role":    "slave",
			},
		},
	}
}
