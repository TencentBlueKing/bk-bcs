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
import functools
from typing import List, Tuple

from backend.resources.utils.kube_client import CoreDynamicClient
from backend.utils.async_run import async_run
from backend.utils.basic import getitems

from .. import models


class AppReleaseManager:
    """
    向上层提供模板集 release 的 CRUD 操作，职责如下
    - 管理 AppRelease 表，完成 release 数据的 CRUD
    - 通过 ReleaseResourceManager 管理 release 中的 resource 变更
    - 一个 release 至少包含一个 resource, 一个 resource 对应一个 ReleaseResourceManager
    """

    def __init__(self, dynamic_client: CoreDynamicClient):
        self.dynamic_client = dynamic_client

    def update_or_create(self, operator: str, release_data: models.AppReleaseData) -> Tuple[models.AppRelease, bool]:
        app_release, created = models.AppRelease.objects.update_or_create(
            name=release_data.name,
            cluster_id=release_data.cluster_id,
            namespace=release_data.namespace,
            defaults={
                "template_id": release_data.template_id,
                "project_id": release_data.project_id,
                "creator": operator,
                "updator": operator,
            },
        )

        try:
            self._deploy(operator, app_release.id, release_data.resource_list)
        except Exception as e:
            app_release.update_status(models.ReleaseStatus.FAILED.value, str(e))
        else:
            app_release.update_status(models.ReleaseStatus.DEPLOYED.value)

        return app_release, created

    def _deploy(self, operator: str, app_release_id: int, resource_list: List[models.ResourceData]):
        res_mgr = ReleaseResourceManager(self.dynamic_client, app_release_id)
        tasks = [functools.partial(res_mgr.update_or_create, operator, resource) for resource in resource_list]
        async_run(tasks)

    def delete(self, operator: str, app_release_id: int):
        self._delete(operator, app_release_id)
        models.AppRelease.objects.filter(id=app_release_id).delete()

    def _delete(self, operator: str, app_release_id: int):
        res_mgr = ReleaseResourceManager(self.dynamic_client, app_release_id)
        tasks = [
            functools.partial(res_mgr.delete, operator, resource_inst.id)
            for resource_inst in models.ResourceInstance.objects.filter(app_release_id=app_release_id)
        ]
        async_run(tasks)


class ReleaseResourceManager:
    """
    提供对 AppRelease 中 resource 的 CRUD 操作
    - 管理 ResourceInstance 表，完成 resource 数据的 CRUD
    - 通过与集群的实际对接，管理 resource 的运行状态
    """

    def __init__(self, dynamic_client: CoreDynamicClient, app_release_id: int):
        self.dynamic_client = dynamic_client
        self.app_release_id = app_release_id

    def update_or_create(self, operator: str, resource: models.ResourceData) -> Tuple[models.ResourceInstance, bool]:
        api = self._get_api(kind=resource.kind, api_version=getitems(resource.manifest, 'apiVersion'))
        api.update_or_create(
            body=resource.manifest,
            name=resource.name,
            namespace=getitems(resource.manifest, 'metadata.namespace'),
            update_method="replace",
        )

        # https://youngminz.netlify.app/posts/get-or-create-deadlock
        # 并发条件下update_or_create 触发 deadlock found when trying to get lock; try restarting transaction
        kwargs = {'app_release_id': self.app_release_id, 'kind': resource.kind, 'name': resource.name}
        defaults = {
            'manifest': resource.manifest,
            'version': resource.version,
            'revision': resource.revision,
            'updator': operator,
            'creator': operator,
        }
        try:
            resource_inst = models.ResourceInstance.objects.get(**kwargs)
        except models.ResourceInstance.DoesNotExist:
            kwargs.update(defaults)
            return models.ResourceInstance.objects.create(**kwargs), True
        else:
            resource_inst.__dict__.update(defaults)
            resource_inst.save()
            return resource_inst, False

    def edit(self, operator: str, resource: models.ResourceData):
        """
        直接edit线上资源(类似kubectl edit)，不做manifest和version等信息的保存
        """
        api = self._get_api(kind=resource.kind, api_version=getitems(resource.manifest, 'apiVersion'))
        api.update_or_create(
            body=resource.manifest,
            name=resource.name,
            namespace=getitems(resource.manifest, 'metadata.namespace'),
            update_method="replace",
        )
        models.ResourceInstance.objects.filter(
            app_release_id=self.app_release_id, name=resource.name, namespace=resource.namespace, kind=resource.kind
        ).update(edited=True, updator=operator)

    def delete(self, operator: str, resource_inst_id: int):
        resource_inst = models.ResourceInstance.objects.get(id=resource_inst_id)
        api = self._get_api(resource_inst.kind, getitems(resource_inst.manifest, 'apiVersion'))
        api.delete_ignore_nonexistent(
            name=resource_inst.name, namespace=getitems(resource_inst.manifest, 'metadata.namespace')
        )
        return resource_inst.delete()

    def _get_api(self, kind, api_version):
        if api_version:
            return self.dynamic_client.resources.get(kind=kind, api_version=api_version)
        return self.dynamic_client.get_preferred_resource(kind=kind)
