# -*- coding: utf-8 -*-
"""
Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community
Edition) available.
Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://opensource.org/licenses/MIT

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
"""
from backend.packages.blue_krill.data_types.enum import EnumField, StructuredEnum
from backend.utils.basic import ChoicesEnum

# cronjob 不在 preferred resource 中，需要指定 api_version
DEFAULT_CRON_JOB_API_VERSION = 'batch/v1beta1'

# HPA 需要指定 api_version
DEFAULT_HPA_API_VERSION = 'autoscaling/v2beta2'

# 至多展示的 HPA 指标数量
HPA_METRIC_MAX_DISPLAY_NUM = 3


class WorkloadTypes(ChoicesEnum):
    Deployment = "Deployment"
    ReplicaSet = "ReplicaSet"
    StatefulSet = "StatefulSet"
    DaemonSet = "DaemonSet"
    Job = "Job"
    GameStatefulSet = "GameStatefulSet"
    GameDeployment = "GameDeployment"

    _choices_labels = (
        (Deployment, "Deployment"),
        (ReplicaSet, "ReplicaSet"),
        (StatefulSet, "StatefulSet"),
        (DaemonSet, "DaemonSet"),
        (Job, "Job"),
        (GameStatefulSet, "GameStatefulSet"),
        (GameDeployment, "GameDeployment"),
    )


class K8sResourceKind(ChoicesEnum):
    # workload
    Deployment = "Deployment"
    ReplicaSet = "ReplicaSet"
    StatefulSet = "StatefulSet"
    DaemonSet = "DaemonSet"
    CronJob = "CronJob"
    Job = "Job"
    Pod = "Pod"
    # network
    Ingress = "Ingress"
    Service = "Service"
    Endpoints = "Endpoints"
    # configuration
    ConfigMap = "ConfigMap"
    Secret = "Secret"
    # storage
    PersistentVolume = "PersistentVolume"
    PersistentVolumeClaim = "PersistentVolumeClaim"
    StorageClass = "StorageClass"
    # rbac
    ServiceAccount = "ServiceAccount"
    # CustomResource
    CustomResourceDefinition = "CustomResourceDefinition"
    CustomObject = "CustomObject"
    # hpa
    HorizontalPodAutoscaler = "HorizontalPodAutoscaler"
    # other
    Event = "Event"
    Namespace = "Namespace"
    Node = "Node"

    _choices_labels = (
        # workload
        (Deployment, "Deployment"),
        (ReplicaSet, "ReplicaSet"),
        (StatefulSet, "StatefulSet"),
        (DaemonSet, "DaemonSet"),
        (CronJob, "CronJob"),
        (Job, "Job"),
        (Pod, "Pod"),
        # network
        (Endpoints, "Endpoints"),
        (Ingress, "Ingress"),
        (Service, "service"),
        # configuration
        (ConfigMap, "ConfigMap"),
        (Secret, "Secret"),
        # storage
        (PersistentVolume, "PersistentVolume"),
        (PersistentVolumeClaim, "PersistentVolumeClaim"),
        (StorageClass, "StorageClass"),
        # rbac
        (ServiceAccount, "ServiceAccount"),
        # CustomResource
        (CustomResourceDefinition, "CustomResourceDefinition"),
        (CustomObject, "CustomObject"),
        # hpa
        (HorizontalPodAutoscaler, "HorizontalPodAutoscaler"),
        # other
        (Event, "Event"),
        (Namespace, "Namespace"),
        (Node, "Node"),
    )


class K8sServiceTypes(ChoicesEnum):
    ClusterIP = "ClusterIP"
    NodePort = "NodePort"
    LoadBalancer = "LoadBalancer"

    _choices_labels = ((ClusterIP, "ClusterIP"), (NodePort, "NodePort"), (LoadBalancer, "LoadBalancer"))


class PatchType(ChoicesEnum):
    JSON_PATCH_JSON = "application/json-patch+json"
    MERGE_PATCH_JSON = "application/merge-patch+json"
    STRATEGIC_MERGE_PATCH_JSON = "application/strategic-merge-patch+json"
    APPLY_PATCH_YAML = "application/apply-patch+yaml"

    _choices_labels = (
        (JSON_PATCH_JSON, "application/json-patch+json"),
        (MERGE_PATCH_JSON, "application/merge-patch+json"),
        (STRATEGIC_MERGE_PATCH_JSON, "application/strategic-merge-patch+json"),
        (APPLY_PATCH_YAML, "application/apply-patch+yaml"),
    )


class PodConditionType(ChoicesEnum):
    """k8s PodConditionType"""

    PodScheduled = 'PodScheduled'
    PodReady = 'Ready'
    PodInitialized = 'Initialized'
    PodReasonUnschedulable = 'Unschedulable'
    ContainersReady = 'ContainersReady'


class PodPhase(ChoicesEnum):
    """k8s PodPhase"""

    PodPending = 'Pending'
    PodRunning = 'Running'
    PodSucceeded = 'Succeeded'
    PodFailed = 'Failed'
    PodUnknown = 'Unknown'


class SimplePodStatus(ChoicesEnum):
    """
    用于页面展示的简单 Pod 状态
    在 k8s PodPhase 基础上细分了状态
    """

    # 原始 PodPhase 状态
    PodPending = 'Pending'
    PodRunning = 'Running'
    PodSucceeded = 'Succeeded'
    PodFailed = 'Failed'
    PodUnknown = 'Unknown'
    # 细分状态
    NotReady = 'NotReady'
    Terminating = 'Terminating'
    Completed = 'Completed'


class ConditionStatus(ChoicesEnum):
    """k8s ConditionStatus"""

    ConditionTrue = 'True'
    ConditionFalse = 'False'
    ConditionUnknown = 'Unknown'


class PersistentVolumeAccessMode(str, StructuredEnum):
    """k8s PersistentVolumeAccessMode"""

    ReadWriteOnce = EnumField('ReadWriteOnce', label='RWO')
    ReadOnlyMany = EnumField('ReadOnlyMany', label='ROX')
    ReadWriteMany = EnumField('ReadWriteMany', label='RWX')

    @property
    def shortname(self):
        """k8s 官方缩写"""
        return self.get_choice_label(self.value)


class MetricSourceType(str, StructuredEnum):
    """k8s MetricSourceType"""

    Object = EnumField('Object')
    Pods = EnumField('Pods')
    Resource = EnumField('Resource')
    External = EnumField('External')
    ContainerResource = EnumField('ContainerResource')


class NodeConditionStatus(str, StructuredEnum):
    """节点状态"""

    Ready = EnumField("Ready", label="正常状态")
    NotReady = EnumField("NotReady", label="非正常状态")
    Unknown = EnumField("Unknown", label="未知状态")


class NodeConditionType(str, StructuredEnum):
    """节点状态类型
    ref: node condition types
    """

    Ready = EnumField("Ready", label="kubelet is healthy and ready to accept pods")
    MemoryPressure = EnumField(
        "MemoryPressure", label="kubelet is under pressure due to insufficient available memory"
    )
    DiskPressure = EnumField("DiskPressure", label="kubelet is under pressure due to insufficient available disk")
    PIDPressure = EnumField("PIDPressure", label="kubelet is under pressure due to insufficient available PID")
    NetworkUnavailable = EnumField("NetworkUnavailable", label="network for the node is not correctly configured")


class ResourceScope(str, StructuredEnum):
    """ 资源维度 命名空间/集群 """

    Namespaced = 'Namespaced'
    Cluster = 'Cluster'


# 设置 bcs cluster id 缓存时间为7天
BCS_CLUSTER_EXPIRATION_TIME = 3600 * 24 * 7

# 集群维度的资源（K8S原生）
NATIVE_CLUSTER_SCOPE_RES_KINDS = [
    K8sResourceKind.Namespace.value,
    K8sResourceKind.PersistentVolume.value,
    K8sResourceKind.StorageClass.value,
    K8sResourceKind.CustomResourceDefinition.value,
]
