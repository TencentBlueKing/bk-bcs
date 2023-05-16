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
from typing import Dict, List

import attr

from backend.utils.url_slug import NAMESPACE_REGEX

from .constants import LogSourceType

NAMESPACE_PATTERN = re.compile(NAMESPACE_REGEX)


@attr.s(kw_only=True)
class ConfBuilder:
    log_source_type = attr.ib(validator=attr.validators.instance_of(str))
    bk_biz_id = attr.ib(validator=attr.validators.instance_of(int))
    project_id = attr.ib(validator=attr.validators.instance_of(str))
    cluster_id = attr.ib(validator=attr.validators.instance_of(str))
    add_pod_label = attr.ib(validator=attr.validators.instance_of(bool))
    extra_labels = attr.ib(validator=attr.validators.instance_of(dict))
    config_name = attr.ib(validator=attr.validators.instance_of(str))

    @classmethod
    def from_dict(cls, init_data: Dict) -> 'ConfBuilder':
        field_names = [f.name for f in attr.fields(cls)]
        return cls(**{k: v for k, v in init_data.items() if k in field_names})

    def build_create_config(self) -> Dict:
        """生成日志采集配置参数, 用于创建采集规则"""
        return {
            'bk_biz_id': self.bk_biz_id,
            'project_id': self.project_id,
            'collector_config_name': self.config_name,
            'collector_config_name_en': self.config_name,
            'custom_type': 'log',
            'description': '',
            'bcs_cluster_id': self.cluster_id,
            'add_pod_label': self.add_pod_label,
            'extra_labels': [{'key': k, 'value': v} for k, v in self.extra_labels.items()],
            'config': self._build_config(),
        }

    def build_update_config(self) -> Dict:
        """生成日志采集配置参数, 用于更新采集规则"""
        return {
            'bk_biz_id': self.bk_biz_id,
            'project_id': self.project_id,
            'collector_config_name': self.config_name,
            'description': '',
            'bcs_cluster_id': self.cluster_id,
            'add_pod_label': self.add_pod_label,
            'extra_labels': [{'key': k, 'value': v} for k, v in self.extra_labels.items()],
            'config': self._build_config(),
        }

    def _build_config(self) -> List[Dict]:
        """子类需要覆盖此方法"""
        return []


@attr.s(kw_only=True)
class AllContainersConfBuilder(ConfBuilder):
    """日志采集配置生成器(日志源是所有容器)

    :param: namespace: 命名空间. 为空串时表示所有命名空间
    :param: base: 基础容器配置. 格式如 {'enable_stdout': True, 'log_paths': []}
    """

    base = attr.ib(validator=attr.validators.instance_of(dict))
    namespace = attr.ib(validator=attr.validators.instance_of(str), default='')

    @namespace.validator
    def validate_ns(self, attribute, value):
        if value and not NAMESPACE_PATTERN.match(value):
            raise ValueError(f'the namespace {value} is invalid')

    def _build_config(self) -> List[Dict]:
        config = {'data_encoding': 'UTF-8', 'enable_stdout': self.base['enable_stdout']}

        if self.namespace:
            config['namespaces'] = [self.namespace]

        log_paths = self.base.get('log_paths')
        if log_paths:
            config['paths'] = log_paths

        return [config]


@attr.s(kw_only=True)
class SelectedContainersConfBuilder(ConfBuilder):
    """日志采集配置生成器(日志源是指定容器)

    :param: namespace: 命名空间. 不能为空
    :param: workload: 工作负载日志采集配置. 格式如 {'name': 'nginx', 'kind': 'Deployment',
        'container_confs': [{'name': 'nginx', 'enable_stdout': True, 'log_paths': []}]}
    """

    namespace = attr.ib(validator=attr.validators.matches_re(NAMESPACE_REGEX))
    workload = attr.ib(validator=attr.validators.instance_of(dict))

    def _build_config(self) -> List[Dict]:
        configs = []

        for container in self.workload['container_confs']:
            conf = {
                'data_encoding': 'UTF-8',
                'namespaces': [self.namespace],
                'container': {
                    'workload_type': self.workload['kind'],
                    'workload_name': self.workload['name'],
                    'container_name': container['name'],
                },
                'enable_stdout': container['enable_stdout'],
            }

            log_paths = container.get('log_paths')
            if log_paths:
                conf['paths'] = log_paths

            configs.append(conf)

        return configs


@attr.s(kw_only=True)
class SelectedLabelsConfBuilder(ConfBuilder):
    """日志采集配置生成器(日志源是指定标签, 匹配到标签的容器)

    :param: namespace: 命名空间. 不能为空
    :param: selector: 选择器采集配置. 格式如 {'enable_stdout': True, 'log_paths': [],
        'match_labels': {'app': 'nginx'}, 'match_expressions': [{'key': '', 'operator': '', 'values': ''}]}.
        其中 match_labels 和 match_expressions 至少一个有效
    """

    namespace = attr.ib(validator=attr.validators.matches_re(NAMESPACE_REGEX))
    selector = attr.ib(validator=attr.validators.instance_of(dict))

    def _build_config(self) -> List[Dict]:
        label_selector = {}

        match_labels = self.selector.get('match_labels')
        if match_labels:
            label_selector['match_labels'] = [{'key': k, 'value': v} for k, v in match_labels.items()]

        match_expressions = self.selector.get('match_expressions')
        if match_expressions:
            label_selector['match_expressions'] = match_expressions

        config = {
            'data_encoding': 'UTF-8',
            'namespaces': [self.namespace],
            'enable_stdout': self.selector['enable_stdout'],
            'label_selector': label_selector,
        }

        log_paths = self.selector.get('log_paths')
        if log_paths:
            config['paths'] = log_paths

        return [config]


CONF_BUILDER_CLASSES = {
    LogSourceType.ALL_CONTAINERS.value: AllContainersConfBuilder,
    LogSourceType.SELECTED_CONTAINERS.value: SelectedContainersConfBuilder,
    LogSourceType.SELECTED_LABELS.value: SelectedLabelsConfBuilder,
}


def get_builder_class(log_source_type: str):
    return CONF_BUILDER_CLASSES[log_source_type]
