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
from typing import Dict, List, Optional, Union

from django.utils.translation import ugettext_lazy as _
from kubernetes.client import CoreV1Api
from kubernetes.stream import stream

from backend.dashboard.exceptions import ResourceNotExist
from backend.dashboard.workloads.constants import VOLUME_RESOURCE_NAME_KEY_MAP
from backend.resources.constants import K8sResourceKind
from backend.resources.resource import ResourceClient, ResourceList
from backend.resources.utils.format import ResourceDefaultFormatter
from backend.resources.workloads.job import Job
from backend.resources.workloads.pod.formatter import PodFormatter
from backend.resources.workloads.replicaset import ReplicaSet
from backend.utils.basic import getitems
from backend.utils.string import decapitalize


class Pod(ResourceClient):
    """Pod 相关资源 Client"""

    kind = K8sResourceKind.Pod.value
    formatter = PodFormatter()

    def list(
        self,
        is_format: bool = True,
        formatter: Optional[ResourceDefaultFormatter] = None,
        owner_kind: Optional[str] = None,
        owner_name: Optional[str] = None,
        **kwargs
    ) -> Union[ResourceList, Dict]:
        """
        查询 Pod 列表

        :param is_format: 是否使用 PodFormatter 进行格式化
        :param formatter: 指定的特殊格式化器
        :param owner_kind: 上级资源类型
        :param owner_name: 上级资源名称
        :return: Pod 列表信息
        """
        resp = super().list(is_format, formatter, **kwargs)
        if not (kwargs.get('namespace') and owner_kind and owner_name):
            return resp

        # 找到当前指定资源关联的 Pod 的 owner reference 信息
        pod_owner_references = self._get_pod_owner_references(
            kwargs['namespace'], owner_kind, owner_name, kwargs.get('label_selector')
        )

        # NOTE: Pod 类型 list 若需要支持根据 owner_references 过滤，
        # 结果不是返回 ResourceList 而是 Dict，需要上层进行兼容
        # 原因是：ResourceInstance 对象属性不支持赋值
        resp = resp.data.to_dict()
        resp['items'] = self._filter_by_owner_references(resp['items'], pod_owner_references)
        return resp

    def fetch_manifest(self, namespace: str, pod_name: str) -> Dict:
        """
        获取指定 Pod 配置信息

        :param ctx_cluster: 集群 Context
        :param namespace: 命名空间
        :param pod_name: Pod 名称
        :return: Pod 配置信息
        """
        pod = self.get(namespace=namespace, name=pod_name, is_format=False)
        if not pod:
            raise ResourceNotExist(_('Pod {}/{} 不存在').format(namespace, pod_name))
        return pod.data.to_dict()

    def filter_related_resources(self, client: ResourceClient, namespace: str, pod_name: str) -> Dict:
        """
        过滤与 Pod 相关联的资源

        :param client: 关联资源 client
        :param namespace: 命名空间
        :param pod_name: Pod 名称
        :return: 关联资源列表
        """
        pod_manifest = self.fetch_manifest(namespace, pod_name)
        # Pod 配置中资源类型为驼峰式，需要将 ResourceKind 首字母小写
        resource_kind, resource_name_key = decapitalize(client.kind), VOLUME_RESOURCE_NAME_KEY_MAP[client.kind]
        # 获取与指定 Pod 相关联的 某种资源 的资源名称列表
        resource_name_list = [
            volume[resource_kind][resource_name_key]
            for volume in getitems(pod_manifest, 'spec.volumes', [])
            if resource_kind in volume
        ]
        # 查询指定命名空间下该资源的所有实例，并根据名称列表过滤
        resources = client.list(namespace=namespace, is_format=False).data.to_dict()
        resources['items'] = [item for item in resources['items'] if item['metadata']['name'] in resource_name_list]
        # 组装展示用扩展信息等
        manifest_ext = {item['metadata']['uid']: client.formatter.format_dict(item) for item in resources['items']}
        return {'manifest': resources, 'manifest_ext': manifest_ext}

    def exec_command(self, namespace: str, pod_name: str, container_name: str, command: List):
        """
        在指定 Pod 容器中执行命令

        :param namespace: 命名空间
        :param pod_name: Pod 名称
        :param container_name: 容器名称
        :param command: 待执行指令，argv array 格式
        :return: 指令执行结果（stdout，stderr）
        """
        api_instance = CoreV1Api(self.dynamic_client.client)
        return stream(
            api_instance.connect_get_namespaced_pod_exec,
            name=pod_name,
            namespace=namespace,
            command=command,
            container=container_name,
            stderr=True,
            stdin=False,
            stdout=True,
            tty=False,
        )

    def _get_pod_owner_references(
        self, namespace: str, owner_kind: str, owner_name: str, label_selector: str
    ) -> List[Dict]:
        """
        非直接关联 Pod 的资源，找到下层直接关联的子资源

        :param namespace: 命名空间
        :param owner_kind: 所属资源类型
        :param owner_name: 所属资源名称列表
        :param label_selector: 标签选择器
        """
        sub_owner_references = [{'kind': owner_kind, 'name': owner_name}]
        if owner_kind not in [K8sResourceKind.Deployment.value, K8sResourceKind.CronJob.value]:
            return sub_owner_references

        # Deployment/CronJob 不直接关联 Pod，而是通过 ReplicaSet/Job 间接关联，需要向下钻取 Pod 的 owner reference 信息
        SubResClient = {
            K8sResourceKind.Deployment.value: ReplicaSet,
            K8sResourceKind.CronJob.value: Job,
        }[owner_kind]
        sub_res = (
            SubResClient(self.ctx_cluster)
            .list(namespace=namespace, label_selector=label_selector, is_format=False)
            .data.to_dict()
        )
        owner_names = [
            getitems(sr, 'metadata.name')
            for sr in self._filter_by_owner_references(sub_res['items'], sub_owner_references)
        ]
        return [{'kind': SubResClient.kind, 'name': n} for n in owner_names]

    def _filter_by_owner_references(self, sub_res_items: List, owner_references: List) -> List:
        """
        根据 owner_references 过滤关联的子资源

        :param sub_res_items: 子资源列表
        :param owner_references: 所属资源信息
        :return: 根据 owner_references 过滤后的资源列表
        """
        ret = []
        for sub_res in sub_res_items:
            if 'ownerReferences' not in sub_res['metadata']:
                continue
            for owner_ref in sub_res['metadata']['ownerReferences']:
                for ref in owner_references:
                    if owner_ref['kind'] == ref['kind'] and owner_ref['name'] == ref['name']:
                        ret.append(sub_res)
                        break
        return ret
