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
import abc
from copy import deepcopy
from dataclasses import dataclass
from typing import Any, Dict, List, Optional, Tuple, Union

import arrow
from django.utils import timezone
from kubernetes.dynamic.resource import ResourceInstance

from backend.resources.utils.common import calculate_age
from backend.utils.basic import normalize_datetime


class ResourceFormatter(abc.ABC):
    """资源格式化抽象类"""

    @abc.abstractmethod
    def format_list(self, resources: Union[ResourceInstance, List[Dict], None]) -> List[Dict]:
        raise NotImplementedError

    @abc.abstractmethod
    def format(self, resource: Optional[ResourceInstance]) -> Dict:
        raise NotImplementedError

    @abc.abstractmethod
    def format_dict(self, resource_dict: Dict) -> Dict:
        raise NotImplementedError


class ResourceDefaultFormatter(ResourceFormatter):
    """格式化 Kubernetes 资源为通用资源格式"""

    def format_list(self, resources: Union[ResourceInstance, List[Dict], None]) -> List[Dict]:
        if isinstance(resources, (list, tuple)):
            return [self.format_dict(res) for res in resources]
        if resources is None:
            return []
        # Type: ResourceInstance with multiple results returned by DynamicClient
        return [self.format_dict(res) for res in resources.to_dict()['items']]

    def format(self, resource: Optional[ResourceInstance]) -> Dict:
        if resource is None:
            return {}
        return self.format_dict(resource.to_dict())

    def format_common_dict(self, resource_dict: Dict) -> Dict:
        metadata = deepcopy(resource_dict['metadata'])
        self.set_metadata_null_values(metadata)

        create_time, update_time = self.parse_create_update_time(metadata)
        return {
            'age': calculate_age(metadata.get('creationTimestamp', '')),
            'createTime': create_time,
            'updateTime': update_time,
        }

    def format_dict(self, resource_dict: Dict) -> Dict:
        resource_copy = deepcopy(resource_dict)
        metadata = resource_copy['metadata']
        self.set_metadata_null_values(metadata)

        # Get create_time and update_time
        create_time, update_time = self.parse_create_update_time(metadata)
        return {
            "data": resource_copy,
            "clusterId": self.get_cluster_id(metadata),
            "resourceType": resource_copy['kind'],
            "resourceName": metadata['name'],
            "namespace": metadata.get('namespace', ''),
            "createTime": create_time,
            "updateTime": update_time,
        }

    def set_metadata_null_values(self, metadata: Dict):
        """设置 metadata 字段里的空值"""
        metadata['annotations'] = metadata.get('annotations') or {}
        metadata['labels'] = metadata.get('labels') or {}

    def parse_create_time(self, metadata: Dict) -> str:
        """获取 metadata 中的 create_time"""
        create_time = metadata.get("creationTimestamp", "")
        if create_time:
            # create_time format: '2019-12-16T09:10:59Z'
            d_time = arrow.get(create_time).datetime
            create_time = timezone.localtime(d_time).strftime("%Y-%m-%d %H:%M:%S")
        return create_time

    def parse_create_update_time(self, metadata: Dict) -> Tuple:
        """获取 metadata 中的 create_time, update_time"""
        create_time = self.parse_create_time(metadata)
        update_time = metadata['annotations'].get("io.tencent.paas.updateTime") or create_time
        if update_time:
            update_time = normalize_datetime(update_time)
        return create_time, update_time

    def get_cluster_id(self, metadata: Dict) -> str:
        """获取集群 ID"""
        labels = metadata.get("labels", {})
        return labels.get("io.tencent.bcs.clusterid") or ""
