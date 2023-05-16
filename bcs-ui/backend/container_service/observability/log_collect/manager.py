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
import logging
from typing import Dict, List

from django.db import IntegrityError

from backend.components.base import ComponentAuth
from backend.components.databus.collector import LogCollectorClient

from .builder import get_builder_class
from .models import LogCollectMetadata, LogIndexSet
from .serializers import CollectConfSLZ

logger = logging.getLogger(__name__)


class CollectConfManager:
    """日志采集配置管理器"""

    def __init__(self, access_token: str, project_id: str, cluster_id: str):
        self.project_id = project_id
        self.cluster_id = cluster_id
        self.client = LogCollectorClient(auth=ComponentAuth(access_token))

    def get_config(self, config_id: int) -> Dict:
        """查询某个具体的采集规则详情"""
        meta_data = LogCollectMetadata.objects.get(
            config_id=config_id, project_id=self.project_id, cluster_id=self.cluster_id
        )

        data = self.client.list_collect_configs(cluster_id=self.cluster_id)
        rule_configs = {conf['rule_id']: conf for conf in data}

        serializer = CollectConfSLZ(meta_data, context={'rule_configs': rule_configs})

        return serializer.data

    def create_config(self, username: str, config: Dict) -> int:
        """创建日志采集规则"""
        builder = get_builder_class(config['log_source_type']).from_dict(config)

        kwargs = {
            'project_id': self.project_id,
            'cluster_id': self.cluster_id,
            'log_source_type': builder.log_source_type,
            'config_name': builder.config_name,
            'namespace': builder.namespace,
        }
        if LogCollectMetadata.objects.filter(**kwargs).exists():
            raise IntegrityError(
                f'collect config create failed: '
                f'config(cluster_id: {self.cluster_id}, namespace: {builder.namespace}, '
                f'log_source_type: {builder.log_source_type}, config_name: {builder.config_name}) '
                f'is already exists!'
            )

        data = self.client.create_collect_config(config=builder.build_create_config())

        LogCollectMetadata.objects.create(config_id=data['rule_id'], creator=username, updator=username, **kwargs)
        LogIndexSet.safe_create(project_id=self.project_id, bk_biz_id=builder.bk_biz_id, **data)

        return data['rule_id']

    def update_config(self, username: str, config_id: int, config: Dict) -> int:
        """更新日志采集规则"""
        meta_data = LogCollectMetadata.objects.get(
            config_id=config_id, project_id=self.project_id, cluster_id=self.cluster_id
        )
        builder = get_builder_class(meta_data.log_source_type).from_dict(config)
        data = self.client.update_collect_config(config_id=config_id, config=builder.build_update_config())

        meta_data.updator = username
        meta_data.save(update_fields=['updator', 'updated'])

        LogIndexSet.safe_create(project_id=self.project_id, bk_biz_id=builder.bk_biz_id, **data)

        return config_id

    def list_configs(self) -> List[Dict]:
        """列举日志采集规则"""
        data = self.client.list_collect_configs(cluster_id=self.cluster_id)

        rule_configs = {conf['rule_id']: conf for conf in data}
        meta_data_qset = LogCollectMetadata.objects.filter(project_id=self.project_id, cluster_id=self.cluster_id)
        serializer = CollectConfSLZ(meta_data_qset, many=True, context={'rule_configs': rule_configs})

        return serializer.data

    def delete_config(self, username: str, config_id: int) -> int:
        """删除日志采集规则"""
        meta_data = LogCollectMetadata.objects.get(
            config_id=config_id, project_id=self.project_id, cluster_id=self.cluster_id
        )
        # TODO 增加删除审计, 先记录日志
        logger.error(
            '%s delete log collect conf(config_id=%s, config_name=%s) from %s',
            username,
            config_id,
            meta_data.config_name,
            self.cluster_id,
        )

        self.client.delete_collect_config(config_id=config_id)
        meta_data.delete()

        return config_id
