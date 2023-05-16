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
from typing import Dict, List

from backend.container_service.observability.log_collect.constants import LogSourceType, SupportedWorkload


def list_confs(collect_meta_data_qset) -> List[Dict]:
    allcontainers_data = collect_meta_data_qset.get(log_source_type=LogSourceType.ALL_CONTAINERS.value)
    selectedcontainers_data = collect_meta_data_qset.get(log_source_type=LogSourceType.SELECTED_CONTAINERS.value)
    selectelabels_data = collect_meta_data_qset.get(log_source_type=LogSourceType.SELECTED_LABELS.value)
    return [
        {
            'rule_id': allcontainers_data.config_id,
            'add_pod_label': True,
            'container_config': [{'enable_stdout': True}],
        },
        {
            'rule_id': selectedcontainers_data.config_id,
            'add_pod_label': True,
            'extra_labels': [{'key': 'app', 'value': 'nginx'}],
            'container_config': [
                {
                    'data_encoding': 'UTF-8',
                    'container': {
                        'workload_type': SupportedWorkload.Deployment.value,
                        'workload_name': 'nginx-deployment',
                        'container_name': 'nginx',
                    },
                    'enable_stdout': False,
                    'params': {'paths': ['/log/access.log']},
                },
                {
                    'data_encoding': 'UTF-8',
                    'container': {
                        'workload_type': SupportedWorkload.Deployment.value,
                        'workload_name': 'nginx-deployment',
                        'container_name': 'notify',
                    },
                    'enable_stdout': True,
                },
            ],
        },
        {
            'rule_id': selectelabels_data.config_id,
            'add_pod_label': False,
            'container_config': [
                {
                    'enable_stdout': True,
                    'label_selector': {'match_labels': [{'key': 'app', 'value': 'nginx'}]},
                }
            ],
        },
    ]
