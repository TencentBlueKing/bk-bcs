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

package formdata

import (
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/util"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// DeployComplex 单元测试用 Deployment 表单数据(全量)
var DeployComplex = model.Deploy{
	Metadata: model.Metadata{
		APIVersion: "apps/v1",
		Kind:       resCsts.Deploy,
		Name:       "deploy-complex-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
			{"label-key-2", "label-val-2"},
		},
		Annotations: []model.Annotation{
			{"anno-key-1", "anno-val-1"},
			{"anno-key-2", "anno-val-2"},
		},
	},
	Spec: model.DeploySpec{
		Replicas: model.DeployReplicas{
			Cnt:                  "2",
			UpdateStrategy:       resCsts.DefaultUpdateStrategy,
			MaxSurge:             0,
			MSUnit:               util.UnitCnt,
			MaxUnavailable:       20,
			MUAUnit:              util.UnitPercent,
			MinReadySecs:         0,
			ProgressDeadlineSecs: 600,
		},
		NodeSelect: nodeSelect,
		Affinity:   affinity,
		Toleration: toleration,
		Networking: networking,
		Security:   security,
		Other:      specOther,
	},
	ContainerGroup: model.ContainerGroup{
		InitContainers: initContainers,
		Containers:     containers,
	},
	Volume: volume,
}

// DeploySimple 单元测试用 Deployment 表单数据(最简单版本)
var DeploySimple = model.Deploy{
	Metadata: model.Metadata{
		APIVersion: "apps/v1",
		Kind:       resCsts.Deploy,
		Name:       "deploy-simple-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
		},
	},
	Spec: model.DeploySpec{
		Replicas: model.DeployReplicas{
			Cnt:            "2",
			UpdateStrategy: resCsts.DefaultUpdateStrategy,
			MaxSurge:       1,
			MSUnit:         util.UnitCnt,
		},
	},
	ContainerGroup: model.ContainerGroup{
		Containers: []model.Container{
			{
				Basic: model.ContainerBasic{
					Name:       "busybox",
					Image:      "busybox:latest",
					PullPolicy: "IfNotPresent",
				},
			},
		},
	},
}

// STSComplex 单元测试用 StatefulSet 表单数据(全量)
var STSComplex = model.STS{
	Metadata: model.Metadata{
		APIVersion: "apps/v1",
		Kind:       resCsts.STS,
		Name:       "sts-complex-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
			{"label-key-2", "label-val-2"},
		},
		Annotations: []model.Annotation{
			{"anno-key-1", "anno-val-1"},
			{"anno-key-2", "anno-val-2"},
		},
	},
	Spec: model.STSSpec{
		Replicas: model.STSReplicas{
			SVCName:        "svc-complex-y3xk1r9vg9",
			Cnt:            "2",
			UpdateStrategy: resCsts.DefaultUpdateStrategy,
			PodManPolicy:   "OrderedReady",
			Partition:      3,
		},
		VolumeClaimTmpl: model.STSVolumeClaimTmpl{
			Claims: []model.VolumeClaim{
				{
					PVCName:     "pvc-complex-k42wnpaqn7",
					ClaimType:   resCsts.PVCTypeCreateBySC,
					PVName:      "",
					SCName:      "standard",
					StorageSize: 1,
					AccessModes: []string{"ReadOnlyMany", "ReadWriteOnce"},
				},
			},
		},
		NodeSelect: nodeSelect,
		Affinity:   affinity,
		Toleration: toleration,
		Networking: networking,
		Security:   security,
		Other:      specOther,
	},
	ContainerGroup: model.ContainerGroup{
		InitContainers: initContainers,
		Containers:     containers,
	},
	Volume: volume,
}

// DSComplex 单元测试用 DaemonSet 表单数据(全量)
var DSComplex = model.DS{
	Metadata: model.Metadata{
		APIVersion: "apps/v1",
		Kind:       resCsts.DS,
		Name:       "ds-complex-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
			{"label-key-2", "label-val-2"},
		},
		Annotations: []model.Annotation{
			{"anno-key-1", "anno-val-1"},
			{"anno-key-2", "anno-val-2"},
		},
	},
	Spec: model.DSSpec{
		Replicas: model.DSReplicas{
			UpdateStrategy: resCsts.DefaultUpdateStrategy,
			MaxUnavailable: 20,
			MUAUnit:        util.UnitPercent,
			MinReadySecs:   0,
		},
		NodeSelect: nodeSelect,
		Affinity:   affinity,
		Toleration: toleration,
		Networking: networking,
		Security:   security,
		Other:      specOther,
	},
	ContainerGroup: model.ContainerGroup{
		InitContainers: initContainers,
		Containers:     containers,
	},
	Volume: volume,
}

// CJComplex 单元测试用 CronJob 表单数据(全量)
var CJComplex = model.CJ{
	Metadata: model.Metadata{
		APIVersion: "batch/v1",
		Kind:       resCsts.CJ,
		Name:       "cj-complex-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
			{"label-key-2", "label-val-2"},
		},
		Annotations: []model.Annotation{
			{"anno-key-1", "anno-val-1"},
			{"anno-key-2", "anno-val-2"},
		},
	},
	Spec: model.CJSpec{
		JobManage: model.CJJobManage{
			Schedule:                   "0 3 * * *",
			ConcurrencyPolicy:          "Forbid",
			Suspend:                    false,
			Completions:                5,
			Parallelism:                2,
			BackoffLimit:               1,
			ActiveDDLSecs:              600,
			SuccessfulJobsHistoryLimit: 2,
			FailedJobsHistoryLimit:     1,
			StartingDDLSecs:            300,
		},
		NodeSelect: nodeSelect,
		Affinity:   affinity,
		Toleration: toleration,
		Networking: networking,
		Security:   security,
		Other: model.SpecOther{
			RestartPolicy:              "OnFailure",
			TerminationGracePeriodSecs: 30,
			ImagePullSecrets: []string{
				"default-token-1",
			},
			SAName: "default-1",
		},
	},
	ContainerGroup: model.ContainerGroup{
		InitContainers: initContainers,
		Containers:     containers,
	},
	Volume: volume,
}

// JobComplex 单元测试用 Job 表单数据(全量)
var JobComplex = model.Job{
	Metadata: model.Metadata{
		APIVersion: "batch/v1",
		Kind:       resCsts.Job,
		Name:       "job-complex-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
			{"label-key-2", "label-val-2"},
		},
		Annotations: []model.Annotation{
			{"anno-key-1", "anno-val-1"},
			{"anno-key-2", "anno-val-2"},
		},
	},
	Spec: model.JobSpec{
		JobManage: model.JobManage{
			Completions:   5,
			Parallelism:   2,
			BackoffLimit:  1,
			ActiveDDLSecs: 600,
		},
		NodeSelect: nodeSelect,
		Affinity:   affinity,
		Toleration: toleration,
		Networking: networking,
		Security:   security,
		Other: model.SpecOther{
			RestartPolicy:              "Never",
			TerminationGracePeriodSecs: 30,
			ImagePullSecrets: []string{
				"default-token-2",
			},
			SAName: "default-2",
		},
	},
	ContainerGroup: model.ContainerGroup{
		InitContainers: initContainers,
		Containers:     containers,
	},
	Volume: volume,
}

// PodComplex 单元测试用 Pod 表单数据(全量)
var PodComplex = model.Po{
	Metadata: model.Metadata{
		APIVersion: "v1",
		Kind:       resCsts.Po,
		Name:       "pod-complex-" + strings.ToLower(stringx.Rand(10, "")),
		Namespace:  envs.TestNamespace,
		Labels: []model.Label{
			{"label-key-1", "label-val-1"},
			{"label-key-2", "label-val-2"},
		},
		Annotations: []model.Annotation{
			{"anno-key-1", "anno-val-1"},
			{"anno-key-2", "anno-val-2"},
		},
	},
	Spec: model.PoSpec{
		NodeSelect: nodeSelect,
		Affinity:   affinity,
		Toleration: toleration,
		Networking: networking,
		Security:   security,
		Other:      specOther,
	},
	ContainerGroup: model.ContainerGroup{
		InitContainers: initContainers,
		Containers:     containers,
	},
	Volume: volume,
}

var initContainers = []model.Container{
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
		Envs: model.ContainerEnvs{
			Vars: []model.EnvVar{
				{
					Type:  resCsts.EnvVarTypeKeyVal,
					Name:  "ENV_KEY",
					Value: "envValue",
				},
			},
		},
		Resource: model.ContainerRes{
			Requests: model.ResRequirement{
				CPU:    100,
				Memory: 128,
			},
			Limits: model.ResRequirement{
				CPU:    200,
				Memory: 256,
			},
		},
		Mount: model.ContainerMount{
			Volumes: []model.MountVolume{
				{
					Name:      "emptydir",
					MountPath: "/data",
					SubPath:   "cr-init.log",
					ReadOnly:  true,
				},
			},
		},
	},
}

var containers = []model.Container{
	{
		Basic: model.ContainerBasic{
			Name:       "nginx",
			Image:      "nginx:latest",
			PullPolicy: "IfNotPresent",
		},
		Command: model.ContainerCommand{
			WorkingDir: "/data/dev",
			Stdin:      false,
			StdinOnce:  true,
			Tty:        false,
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
					Type:  resCsts.EnvVarTypeKeyVal,
					Name:  "ENV_KEY",
					Value: "envValue",
				},
				{
					Type:  resCsts.EnvVarTypePodField,
					Name:  "MY_POD_NAMESPACE",
					Value: "metadata.namespace",
				},
				{
					Type:   resCsts.EnvVarTypeResource,
					Name:   "MY_CPU_REQUEST",
					Source: "busybox",
					Value:  "requests.cpu",
				},
				{
					Type:   resCsts.EnvVarTypeCMKey,
					Name:   "CM_T_CA_CRT",
					Source: "kube-user-ca.crt",
					Value:  "ca.crt",
				},
				{
					Type:   resCsts.EnvVarTypeSecretKey,
					Name:   "SECRET_T_CA_CRT",
					Source: "default-token-12345",
					Value:  "ca.crt",
				},
				{
					Type:   resCsts.EnvVarTypeCM,
					Name:   "CM_T_",
					Source: "kube-user-ca.crt",
				},
				{
					Type:   resCsts.EnvVarTypeSecret,
					Name:   "SECRET_T_",
					Source: "default-token-12345",
				},
			},
		},
		Healthz: model.ContainerHealthz{
			ReadinessProbe: model.Probe{
				Enabled:          true,
				PeriodSecs:       10,
				InitialDelaySecs: 0,
				TimeoutSecs:      3,
				SuccessThreshold: 1,
				FailureThreshold: 3,
				Type:             resCsts.ProbeTypeTCPSocket,
				Port:             80,
			},
			LivenessProbe: model.Probe{
				Enabled:          true,
				PeriodSecs:       10,
				InitialDelaySecs: 0,
				TimeoutSecs:      3,
				SuccessThreshold: 1,
				FailureThreshold: 3,
				Type:             resCsts.ProbeTypeExec,
				Command:          []string{"echo hello"},
			},
		},
		Resource: model.ContainerRes{
			Requests: model.ResRequirement{
				CPU:    100,
				Memory: 128,
				Extra: []model.ResExtra{
					{
						Key:   "tencent.com/fgpu",
						Value: "1",
					},
					{
						Key:   "tke.cloud.tencent.com/eip",
						Value: "1",
					},
				},
			},
			Limits: model.ResRequirement{
				CPU:    500,
				Memory: 1024,
				Extra: []model.ResExtra{
					{
						Key:   "tencent.com/fgpu",
						Value: "1",
					},
					{
						Key:   "tke.cloud.tencent.com/eip",
						Value: "1",
					},
				},
			},
		},
		Security: model.SecurityCtx{
			Privileged:               true,
			AllowPrivilegeEscalation: true,
			RunAsUser:                1111,
			RunAsGroup:               2222,
			ProcMount:                "Default",
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
}

var volume = model.WorkloadVolume{
	PVC: []model.PVCVolume{
		{
			Name:     "pvc",
			PVCName:  "pvc-123456",
			ReadOnly: false,
		},
	},
	HostPath: []model.HostPathVolume{
		{
			Name: "hostpath",
			Path: "/tmp/hostP.log",
			Type: "FileOrCreate",
		},
	},
	ConfigMap: []model.CMVolume{
		{
			Name:        "cm",
			DefaultMode: "420",
			CMName:      "kube-root-ca.crt",
			Items: []model.KeyToPath{
				{
					Key:  "ca.crt",
					Path: "ca.crt",
				},
			},
		},
	},
	Secret: []model.SecretVolume{
		{
			Name:        "secret",
			DefaultMode: "420",
			SecretName:  "ssh-auth-test",
			Items:       []model.KeyToPath{},
		},
	},
	EmptyDir: []model.EmptyDirVolume{
		{
			Name: "emptydir",
		},
	},
	NFS: []model.NFSVolume{
		{
			Name:     "nfs",
			Path:     "/data",
			Server:   "1.1.1.1",
			ReadOnly: false,
		},
	},
}

var nodeSelect = model.NodeSelect{
	Type:     resCsts.NodeSelectTypeSpecificNode,
	NodeName: "vm-123",
	Selector: []model.NodeSelector{
		{Key: "kubernetes.io/arch", Value: "amd64"},
	},
}

var affinity = model.Affinity{
	NodeAffinity: []model.NodeAffinity{
		{
			Priority: resCsts.AffinityPriorityRequired,
			Selector: model.NodeAffinitySelector{
				Expressions: []model.ExpSelector{
					{Key: "testKey", Op: "In", Values: "testValue1"},
				},
				Fields: []model.FieldSelector{
					{Key: "metadata.name", Op: "In", Values: "test-name"},
				},
			},
		},
		{
			Priority: resCsts.AffinityPriorityPreferred,
			Weight:   10,
			Selector: model.NodeAffinitySelector{
				Expressions: []model.ExpSelector{
					{Key: "testKey", Op: "In", Values: "testVal1,testVal2,testVal3"},
				},
				Fields: []model.FieldSelector{
					{Key: "metadata.name", Op: "In", Values: "test-name1"},
				},
			},
		},
	},
	PodAffinity: []model.PodAffinity{
		{
			Type:     resCsts.AffinityTypeAffinity,
			Priority: resCsts.AffinityPriorityPreferred,
			Namespaces: []string{
				"kube-system",
				"default",
			},
			Weight:      30,
			TopologyKey: "topoKeyTest1",
			Selector: model.PodAffinitySelector{
				Expressions: []model.ExpSelector{
					{Key: "testKey", Op: "Exists", Values: ""},
				},
				Labels: []model.LabelSelector{
					{Key: "labelKey", Value: "labelVal"},
				},
			},
		},
		{
			Type:     resCsts.AffinityTypeAffinity,
			Priority: resCsts.AffinityPriorityRequired,
			Namespaces: []string{
				"kube-node-lease",
				"default",
			},
			TopologyKey: "topoKeyTest0",
			Selector: model.PodAffinitySelector{
				Expressions: []model.ExpSelector{
					{Key: "testKey0", Op: "In", Values: "testVal0,testVal1"},
				},
				Labels: []model.LabelSelector{
					{Key: "labelKey1", Value: "labelVal1"},
				},
			},
		},
		{
			Type:     resCsts.AffinityTypeAntiAffinity,
			Priority: resCsts.AffinityPriorityPreferred,
			Namespaces: []string{
				"default",
				"kube-system",
			},
			Weight:      50,
			TopologyKey: "topoKeyTest2",
			Selector: model.PodAffinitySelector{
				Expressions: []model.ExpSelector{
					{Key: "testKey2", Op: "In", Values: "testVal2,testVal2"},
				},
				Labels: []model.LabelSelector{
					{Key: "testKey3", Value: "testVal3"},
				},
			},
		},
		{
			Type:     resCsts.AffinityTypeAntiAffinity,
			Priority: resCsts.AffinityPriorityRequired,
			Namespaces: []string{
				"default",
			},
			TopologyKey: "topoKeyTest3",
			Selector: model.PodAffinitySelector{
				Expressions: []model.ExpSelector{
					{Key: "testKey3", Op: "In", Values: "testVal3,testVal4"},
				},
				Labels: []model.LabelSelector{
					{Key: "testKey4", Value: "testVal4"},
				},
			},
		},
	},
}

var toleration = model.Toleration{
	Rules: []model.TolerationRule{
		{
			Key:    "testTolKey1",
			Op:     "Exists",
			Effect: "PreferNoSchedule",
		},
		{
			Key:            "testTolKey2",
			Op:             "Equal",
			Effect:         "NoExecute",
			Value:          "tolVal",
			TolerationSecs: 120,
		},
	},
}

var networking = model.Networking{
	DNSPolicy:             "ClusterFirst",
	HostIPC:               true,
	HostNetwork:           false,
	HostPID:               false,
	ShareProcessNamespace: false,
	Hostname:              "vm-12345",
	Subdomain:             "blueking",
	NameServers: []string{
		"1.1.1.1",
		"2.2.2.2",
	},
	Searches: []string{
		"3.3.3.3",
	},
	DNSResolverOpts: []model.DNSResolverOpt{
		{Name: "testName", Value: "testValue"},
	},
	HostAliases: []model.HostAlias{
		{IP: "5.5.5.5", Alias: "vm-1,vm-2"},
	},
}

var security = model.PodSecurityCtx{
	RunAsUser:    1111,
	RunAsNonRoot: true,
	RunAsGroup:   2222,
	FSGroup:      3333,
	SELinuxOpt: model.SELinuxOpt{
		Level: "4444",
		Role:  "5555",
		Type:  "6666",
		User:  "7777",
	},
}

var specOther = model.SpecOther{
	RestartPolicy:              "Always",
	TerminationGracePeriodSecs: 30,
	ImagePullSecrets: []string{
		"default-token-1",
		"default-token-2",
	},
	SAName: "default",
}
