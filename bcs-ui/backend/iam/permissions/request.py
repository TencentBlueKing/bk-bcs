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
from abc import ABC
from collections import namedtuple
from typing import Dict, List, Optional, Union

from django.conf import settings
from iam import Resource
from iam.apply import models


class ResourceRequest(ABC):
    resource_type: str = ''
    attr: Optional[Dict] = None

    def __init__(self, res: Union[List[str], str], **attr_kwargs):
        """
        :param res: 单个资源 ID 或资源 ID 列表
        :param attr_kwargs: 用于替换 attr 中可能需要 format 的值
        """
        self.res = [str(res_id) for res_id in res] if isinstance(res, list) else str(res)
        self.attr_kwargs = dict(**attr_kwargs)
        self._validate_attr_kwargs()

    def make_resources(self) -> List[Resource]:
        if isinstance(self.res, str):
            return [Resource(settings.APP_ID, self.resource_type, self.res, self._make_attribute(self.res))]

        return [
            Resource(settings.APP_ID, self.resource_type, res_id, self._make_attribute(res_id)) for res_id in self.res
        ]

    def _validate_attr_kwargs(self):
        """如果校验不通过，抛出 AttrValidateError 异常"""
        return

    def _make_attribute(self, res_id: str) -> Dict:
        return {}


IAMResource = namedtuple('IAMResource', 'resource_type resource_id')


class ActionResourcesRequest:
    """
    操作资源请求.
    note: resources 是由资源 ID 构成的列表. 为 None 时，表示资源无关.
    资源实例相关时，resources 表示的资源必须具有相同的父实例。以命名空间为例，它们必须是同项目同集群下
    """

    def __init__(
        self,
        action_id: str,
        resource_type: Optional[str] = None,
        resources: Optional[List[str]] = None,
        parent_chain: List[IAMResource] = None,
    ):
        """
        :param action_id: 操作 ID
        :param resource_type: 资源类型
        :param resources: 资源 ID 列表
        :param parent_chain: 按照父类层级排序(父->子) [(resource_type, resource_id), ]
        """
        self.action_id = action_id
        self.resource_type = resource_type
        self.resources = resources
        self.parent_chain = parent_chain

    def to_action(self) -> Union[models.ActionWithResources, models.ActionWithoutResources]:
        # 资源实例相关
        if self.resources:
            parent_chain_node = self._to_parent_chain_node()
            instances = [
                models.ResourceInstance(parent_chain_node + [models.ResourceNode(self.resource_type, res_id, res_id)])
                for res_id in self.resources
            ]
            related_resource_type = models.RelatedResourceType(settings.APP_ID, self.resource_type, instances)
            return models.ActionWithResources(self.action_id, [related_resource_type])

        # 资源实例无关
        return models.ActionWithoutResources(self.action_id)

    def _to_parent_chain_node(self) -> List[models.ResourceNode]:
        if self.parent_chain:
            return [models.ResourceNode(p.resource_type, p.resource_id, p.resource_id) for p in self.parent_chain]
        return []
