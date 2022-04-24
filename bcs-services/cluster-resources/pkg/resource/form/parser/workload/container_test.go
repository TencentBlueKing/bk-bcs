/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package workload

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

var lightManifest4ContainerGroupTest = map[string]interface{}{
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
	"spec": map[string]interface{}{
		"template": map[string]interface{}{
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"image":           "busybox:latest",
						"imagePullPolicy": "IfNotPresent",
						"name":            "busybox",
						"workingDir":      "/data/dev",
						"stdin":           false,
						"stdinOnce":       true,
						"tty":             false,
						"command": []interface{}{
							"/bin/bash",
							"-c",
						},
						"args": []interface{}{
							"echo hello",
						},
						"ports": []interface{}{
							map[string]interface{}{
								"name":          "tcp",
								"protocol":      "TCP",
								"containerPort": int64(80),
								"hostPort":      int64(80),
							},
						},
						"readinessProbe": map[string]interface{}{
							"periodSeconds":    int64(10),
							"timeoutSeconds":   int64(3),
							"successThreshold": int64(1),
							"failureThreshold": int64(3),
							"tcpSocket": map[string]interface{}{
								"port": int64(80),
							},
						},
						"livenessProbe": map[string]interface{}{
							"periodSeconds":    int64(10),
							"timeoutSeconds":   int64(3),
							"successThreshold": int64(1),
							"failureThreshold": int64(3),
							"exec": map[string]interface{}{
								"command": []interface{}{
									"echo hello",
								},
							},
						},
						"resources": map[string]interface{}{
							"requests": map[string]interface{}{
								"memory": "128Mi",
								"cpu":    "100m",
							},
							"limits": map[string]interface{}{
								"memory": "1Gi",
								"cpu":    "0.5",
							},
						},
						"env": []interface{}{
							map[string]interface{}{
								"name":  "ENV_KEY",
								"value": "envValue",
							},
							map[string]interface{}{
								"name": "MY_POD_NAMESPACE",
								"valueFrom": map[string]interface{}{
									"fieldRef": map[string]interface{}{
										"apiVersion": "v1",
										"fieldPath":  "metadata.namespace",
									},
								},
							},
							map[string]interface{}{
								"name": "MY_CPU_REQUEST",
								"valueFrom": map[string]interface{}{
									"resourceFieldRef": map[string]interface{}{
										"containerName": "busybox",
										"divisor":       int64(0),
										"resource":      "requests.cpu",
									},
								},
							},
							map[string]interface{}{
								"name": "CM_T_CA_CRT",
								"valueFrom": map[string]interface{}{
									"configMapKeyRef": map[string]interface{}{
										"name": "kube-user-ca.crt",
										"key":  "ca.crt",
									},
								},
							},
							map[string]interface{}{
								"name": "SECRET_T_CA_CRT",
								"valueFrom": map[string]interface{}{
									"secretKeyRef": map[string]interface{}{
										"name": "default-token-12345",
										"key":  "ca.crt",
									},
								},
							},
						},
						"envFrom": []interface{}{
							map[string]interface{}{
								"prefix": "CM_T_",
								"configMapRef": map[string]interface{}{
									"name": "kube-user-ca.crt",
								},
							},
							map[string]interface{}{
								"prefix": "SECRET_T_",
								"secretRef": map[string]interface{}{
									"name": "default-token-12345",
								},
							},
						},
						"volumeMounts": []interface{}{
							map[string]interface{}{
								"name":      "emptydir",
								"mountPath": "/data",
								"subPath":   "cr.log",
								"readOnly":  true,
							},
						},
						"securityContext": map[string]interface{}{
							"privileged":               true,
							"allowPrivilegeEscalation": true,
							"runAsUser":                int64(1111),
							"runAsGroup":               int64(2222),
							"procMount":                "3333",
							"capabilities": map[string]interface{}{
								"add": []interface{}{
									"AUDIT_CONTROL",
									"AUDIT_WRITE",
								},
								"drop": []interface{}{
									"BLOCK_SUSPEND",
									"CHOWN",
								},
							},
							"seLinuxOptions": map[string]interface{}{
								"level": "111",
								"role":  "222",
								"type":  "333",
								"user":  "444",
							},
						},
						"terminationMessagePath":   "/dev/termination-log",
						"terminationMessagePolicy": "File",
					},
				},
			},
		},
	},
}

var exceptedContainerGroup = model.ContainerGroup{
	Containers: []model.Container{
		{
			Basic: model.ContainerBasic{
				Name:       "busybox",
				Image:      "busybox:latest",
				PullPolicy: "IfNotPresent",
			},
			Command: model.ContainerCommand{
				WorkingDir: "/data/dev",
				Stdin:      false,
				StdinOnce:  true,
				Tty:        false,
				Command:    []string{"/bin/bash", "-c"},
				Args:       []string{"echo hello"},
			},
			Service: model.ContainerService{
				Ports: []model.ContainerPort{
					{
						Name:          "tcp",
						Protocol:      "TCP",
						ContainerPort: 80,
						HostPort:      80,
					},
				},
			},
			Envs: model.ContainerEnvs{
				Vars: []model.EnvVar{
					{
						Type:  EnvVarTypeKeyVal,
						Name:  "ENV_KEY",
						Value: "envValue",
					},
					{
						Type:  EnvVarTypePodField,
						Name:  "MY_POD_NAMESPACE",
						Value: "metadata.namespace",
					},
					{
						Type:   EnvVarTypeResource,
						Name:   "MY_CPU_REQUEST",
						Source: "busybox",
						Value:  "requests.cpu",
					},
					{
						Type:   EnvVarTypeCMKey,
						Name:   "CM_T_CA_CRT",
						Source: "kube-user-ca.crt",
						Value:  "ca.crt",
					},
					{
						Type:   EnvVarTypeSecretKey,
						Name:   "SECRET_T_CA_CRT",
						Source: "default-token-12345",
						Value:  "ca.crt",
					},
					{
						Type:   EnvVarTypeCM,
						Name:   "CM_T_",
						Source: "kube-user-ca.crt",
					},
					{
						Type:   EnvVarTypeSecret,
						Name:   "SECRET_T_",
						Source: "default-token-12345",
					},
				},
			},
			Healthz: model.ContainerHealthz{
				ReadinessProbe: model.Probe{
					PeriodSecs:       10,
					InitialDelaySecs: 0,
					TimeoutSecs:      3,
					SuccessThreshold: 1,
					FailureThreshold: 3,
					Type:             ProbeTypeTCPSocket,
					Port:             80,
				},
				LivenessProbe: model.Probe{
					PeriodSecs:       10,
					InitialDelaySecs: 0,
					TimeoutSecs:      3,
					SuccessThreshold: 1,
					FailureThreshold: 3,
					Type:             ProbeTypeExec,
					Command:          []string{"echo hello"},
				},
			},
			Resource: model.ContainerRes{
				Requests: model.ResRequirement{
					CPU:    100,
					Memory: 128,
				},
				Limits: model.ResRequirement{
					CPU:    500,
					Memory: 1024,
				},
			},
			Security: model.SecurityCtx{
				Privileged:               true,
				AllowPrivilegeEscalation: true,
				RunAsUser:                1111,
				RunAsGroup:               2222,
				ProcMount:                "3333",
				Capabilities: model.Capabilities{
					Add: []string{
						"AUDIT_CONTROL",
						"AUDIT_WRITE",
					},
					Drop: []string{
						"BLOCK_SUSPEND",
						"CHOWN",
					},
				},
				SELinuxOpt: model.SELinuxOpt{
					Level: "111",
					Role:  "222",
					Type:  "333",
					User:  "444",
				},
			},
			Mount: model.ContainerMount{
				Volumes: []model.MountVolume{
					{
						Name:      "emptydir",
						MountPath: "/data",
						SubPath:   "cr.log",
						ReadOnly:  true,
					},
				},
			},
		},
	},
}

func TestParseContainerGroup(t *testing.T) {
	actualContainerGroup := model.ContainerGroup{}
	ParseContainerGroup(lightManifest4ContainerGroupTest, &actualContainerGroup)
	assert.Equal(t, exceptedContainerGroup, actualContainerGroup)
}
