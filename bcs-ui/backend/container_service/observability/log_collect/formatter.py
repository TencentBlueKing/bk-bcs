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
from typing import Any, Dict, List

from backend.utils.basic import getitems

from .constants import LogSourceType


def format_allcontainers_conf(container_config: List[Dict]) -> Dict[str, Any]:
    # 采集源为所有容器时, container_config 只会有一个配置
    conf = container_config[0]
    return {'base': {'log_paths': getitems(conf, 'params.paths', []), 'enable_stdout': conf['enable_stdout']}}


def format_selectedcontainers_conf(container_config: List[Dict]) -> Dict[str, Any]:
    container_confs = [
        {
            'name': getitems(config, 'container.container_name'),
            'log_paths': getitems(config, 'params.paths', []),
            'enable_stdout': config['enable_stdout'],
        }
        for config in container_config
    ]
    conf = container_config[0]
    workload = {
        'name': getitems(conf, 'container.workload_name'),
        'kind': getitems(conf, 'container.workload_type'),
        'container_confs': container_confs,
    }

    return {'workload': workload}


def format_selectedlabels_conf(container_config: List[Dict]) -> Dict[str, Any]:
    # 采集源为选定标签时, container_config 只会有一个配置
    conf = container_config[0]

    match_labels = getitems(conf, 'label_selector.match_labels', [])
    match_labels = {label['key']: label['value'] for label in match_labels}

    return {
        'selector': {
            'match_labels': match_labels,
            'match_expressions': getitems(conf, 'label_selector.match_expressions', []),
            'log_paths': getitems(conf, 'params.paths', []),
            'enable_stdout': conf['enable_stdout'],
        }
    }


FORMAT_FUNC = {
    LogSourceType.ALL_CONTAINERS.value: format_allcontainers_conf,
    LogSourceType.SELECTED_CONTAINERS.value: format_selectedcontainers_conf,
    LogSourceType.SELECTED_LABELS.value: format_selectedlabels_conf,
}


def format(log_source_type: str, container_config: List[Dict]) -> Dict[str, Any]:
    """清除一些无用的数据, 同时规整结构"""
    return FORMAT_FUNC[log_source_type](container_config)
