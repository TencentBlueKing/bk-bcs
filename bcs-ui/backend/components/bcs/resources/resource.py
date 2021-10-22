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
import re

import arrow
from django.utils import timezone
from kubernetes import client

from backend.utils.basic import getitems, normalize_datetime


class Resource:
    def __init__(self, api_client, version=None):
        self.version = version
        self.api_client = api_client

    @property
    def api_class(self):
        return self.get_api_class(self.api_client)

    @property
    def api_instance(self):
        return getattr(client, self.api_class)(self.api_client)

    def render_resource(self, resource_type, resource, resource_name, namespace):
        if not getitems(resource, ["metadata", "annotations"], {}):
            resource["metadata"]["annotations"] = {}
        if not getitems(resource, ["metadata", "labels"], {}):
            resource["metadata"]["labels"] = {}
        create_time = getitems(resource, ["metadata", "creationTimestamp"], "")
        if create_time:
            # create_time format: '2019-12-16T09:10:59Z'
            d_time = arrow.get(create_time).datetime
            create_time = timezone.localtime(d_time).strftime("%Y-%m-%d %H:%M:%S")
        annotations = getitems(resource, ["metadata", "annotations"], {})
        update_time = annotations.get("io.tencent.paas.updateTime") or create_time
        if update_time:
            update_time = normalize_datetime(update_time)
        labels = getitems(resource, ["metadata", "labels"], {})
        return {
            "data": resource,
            "clusterId": labels.get("io.tencent.bcs.clusterid") or "",
            "resourceName": resource_name,
            "resourceType": resource_type,
            "createTime": create_time,
            "updateTime": update_time,
            "namespace": namespace,
        }

    def render_resource_for_preload_content(self, resource_type, resource, resource_name, namespace):
        # 兼容逻辑，防止view层出错
        if not resource.metadata.annotations:
            resource.metadata.annotations = {}
        if not resource.metadata.labels:
            resource.metadata.labels = {}
        create_time = resource.metadata.creation_timestamp or ""
        if create_time:
            create_time = timezone.localtime(create_time).strftime("%Y-%m-%d %H:%M:%S")
        annotations = resource.metadata.annotations
        update_time = annotations.get("io.tencent.paas.updateTime") or create_time
        if update_time:
            update_time = normalize_datetime(update_time)
        labels = resource.metadata.labels
        return {
            "data": resource.to_dict(),
            "clusterId": labels.get("io.tencent.bcs.clusterid") or "",
            "resourceName": resource_name,
            "resourceType": resource_type,
            "createTime": create_time,
            "updateTime": update_time,
            "namespace": namespace,
        }


class FilterResourceData:
    def format_name_and_ns_list(self, params):
        name = params.get("name") or []
        if isinstance(name, str):
            name = re.findall(r"[^,;]+", name)
        namespace = params.get("namespace") or []
        if isinstance(namespace, str):
            namespace = re.findall(r"[^,;]+", namespace)
        return name, namespace

    def _get_resource(self, params, name, namespace, resource, resource_type):
        name_list, namespace_list = self.format_name_and_ns_list(params)
        resource = self.render_resource(resource_type, resource, name, namespace)
        # 过滤参数包含name和namespace
        if name_list and namespace_list:
            if name in name_list and namespace in namespace_list:
                return resource
            return
        # 过滤参数只包含name或namespace
        if name_list or namespace_list:
            if name in name_list or namespace in namespace_list:
                return resource
            return
        return resource

    def filter_data(self, resource_type, resp_data, params):
        data_list = []
        for resource in resp_data.get("items") or []:
            name = getitems(resource, ["metadata", "name"], "")
            namespace = getitems(resource, ["metadata", "namespace"], "")
            # filter data
            resource = self._get_resource(params, name, namespace, resource, resource_type)
            if resource:
                data_list.append(resource)

        return data_list


class BaseMixins:
    def compose_api_class(self, group_version):
        """通过group version 组装对应分组的api class"""
        return f"{''.join([info.capitalize() for info in group_version.split('/')])}Api"


class APPAPIClassMixins(BaseMixins):
    """
    支持资源类型: Deployment/DaemonSet/StatefulSet/ResplicaSet
    """

    def get_api_class(self, api_client):
        resp = client.AppsApi(api_client).get_api_group()
        group_version = resp.preferred_version.group_version
        return self.compose_api_class(group_version)


class CoreAPIClassMixins(BaseMixins):
    """
    支持资源类型: ConfigMap/Endpoints/Event/Namespace/Node/Pod/PersistentVolume/secret等
    """

    def get_api_class(self, api_client):
        resp = client.CoreApi(api_client).get_api_versions()
        version = resp.versions[0]
        return f"Core{version.capitalize()}Api"


class ExtensionsAPIClassMixins(BaseMixins):
    """
    支持资源类型: Ingress
    """

    def get_api_class(self, api_client):
        resp = client.ExtensionsApi(api_client).get_api_group()
        group_version = resp.preferred_version.group_version
        return self.compose_api_class(group_version)


class BatchAPIClassMixins(BaseMixins):
    """
    支持资源类型: Job
    """

    def get_api_class(self, api_client):
        resp = client.BatchApi(api_client).get_api_group()
        group_version = resp.preferred_version.group_version
        return self.compose_api_class(group_version)


class StorageAPIClassMixins(BaseMixins):
    """
    支持资源类型: StorageClass
    """

    def get_api_class(self, api_client):
        resp = client.StorageApi(api_client).get_api_group()
        version = resp.preferred_version.version
        return self.compose_api_class(f"Storage/{version}")
