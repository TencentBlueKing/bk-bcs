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
from django.utils.translation import ugettext_lazy as _

from backend.packages.blue_krill.data_types.enum import EnumField, StructuredEnum


class ActivityType(str, StructuredEnum):
    """ 操作类型 """

    Add = EnumField('add', _('创建'))
    Modify = EnumField('modify', _('更新'))
    Rollback = EnumField('rollback', _('回滚'))
    Delete = EnumField('delete', _('删除'))
    Begin = EnumField('begin', _('开始'))
    End = EnumField('end', _('结束'))
    Start = EnumField('start', _('启动'))
    Pause = EnumField('pause', _('暂停'))
    CarryOn = EnumField('carryon', _('继续'))
    Stop = EnumField('stop', _('停止'))
    Restart = EnumField('restart', _('重启'))
    Retrieve = EnumField('retrieve', _('查询'))
    Reschedule = EnumField('reschedule', _('重新调度'))


class ActivityStatus(str, StructuredEnum):
    """ 操作状态 """

    Completed = EnumField('completed', _('完成'))
    Error = EnumField('error', _('错误'))
    Succeed = EnumField('succeed', _('成功'))
    Failed = EnumField('failed', _('失败'))


class ResourceType(str, StructuredEnum):
    """ 资源类型 """

    Project = EnumField('project', _('项目'))
    Cluster = EnumField('cluster', _('集群'))
    Node = EnumField('node', _('节点'))
    Namespace = EnumField('namespace', _('命名空间'))
    Template = EnumField('template', _('模板集'))
    Variable = EnumField('variable', _('变量'))
    Instance = EnumField('instance', _('应用'))
    LB = EnumField('lb', 'LoadBalancer')
    Metric = EnumField('metric', 'Metric')
    WebConsole = EnumField('web_console', 'WebConsole')
    HelmApp = EnumField('helm_app', 'Helm')
    HPA = EnumField('hpa', 'HPA')
    # workloads
    Deployment = EnumField('deployment', 'Deployment')
    DaemonSet = EnumField('daemonset', 'DaemonSet')
    StatefulSet = EnumField('statefulset', 'StatefulSet')
    CronJob = EnumField('cronjob', 'CronJob')
    Job = EnumField('job', 'Job')
    Pod = EnumField('pod', 'Pod')
    # networks
    Ingress = EnumField('ingress', 'Ingress')
    Service = EnumField('service', 'Service')
    Endpoints = EnumField('endpoints', 'Endpoints')
    # configurations
    ConfigMap = EnumField('configmap', 'Configmap')
    Secret = EnumField('secret', 'Secret')
    # storages
    PersistentVolume = EnumField('persistentvolume', 'PersistentVolume')
    PersistentVolumeClaim = EnumField('persistentvolumeclaim', 'PersistentVolumeClaim')
    StorageClass = EnumField('storageclass', 'StorageClass')
    # rbac
    ServiceAccount = EnumField('serviceaccount', 'ServiceAccount')
    # CustomResource
    CRD = EnumField('crd', _('自定义资源定义'))
    CustomObject = EnumField('customobject', _('自定义对象'))
    ResourceQuota = EnumField('resourcequota', 'resourcequota')


ResourceTypes = dict(ResourceType.get_choices())

MetaMap = {
    'activity_type': dict(ActivityType.get_choices()),
    'activity_status': dict(ActivityStatus.get_choices()),
    'resource_type': ResourceTypes,
}
