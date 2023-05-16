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
import random
from typing import Dict

import pytest
from django.utils.crypto import get_random_string

from backend.container_service.observability.log_collect.constants import LogSourceType, SupportedWorkload
from backend.container_service.observability.log_collect.models import LogCollectMetadata

from .fake_confs import list_confs


@pytest.fixture
def bk_biz_id():
    return random.randint(1, 10000)


@pytest.fixture
def config_id():
    return random.randint(1, 10000)


@pytest.fixture
def log_collect_result(config_id):
    return {
        'rule_id': config_id,
        'std_index_set_id': random.randint(1, 1000),
        'file_index_set_id': random.randint(1, 1000),
    }


@pytest.fixture
def allcontainers_create_params(project_id, cluster_id, namespace, bk_biz_id) -> Dict[str, Dict]:
    collector_config_name = f'namespace_{namespace}_all_log'
    req_params = {
        'log_source_type': LogSourceType.ALL_CONTAINERS.value,
        'namespace': namespace,
        'config_name': collector_config_name,
        'bk_biz_id': bk_biz_id,
        'add_pod_label': True,
        'extra_labels': {'env': 'test', 'app': 'nginx'},
        'base': {'enable_stdout': True, 'log_paths': ['/data/1.log', '/data/2.log']},
    }

    config_params = {
        'bk_biz_id': bk_biz_id,
        'project_id': project_id,
        'collector_config_name': collector_config_name,
        'collector_config_name_en': collector_config_name,
        'custom_type': 'log',
        'description': '',
        'bcs_cluster_id': cluster_id,
        'add_pod_label': req_params['add_pod_label'],
        'extra_labels': [{'key': 'env', 'value': 'test'}, {'key': 'app', 'value': 'nginx'}],
        'config': [
            {
                'data_encoding': 'UTF-8',
                'namespaces': [namespace],
                'paths': req_params['base']['log_paths'],
                'enable_stdout': req_params['base']['enable_stdout'],
            }
        ],
    }

    return {'req_params': req_params, 'config_params': config_params}


@pytest.fixture
def selectedcontainers_create_params(project_id, cluster_id, namespace, bk_biz_id) -> Dict[str, Dict]:
    collector_config_name = f'{SupportedWorkload.Deployment}_nginx_deployment_log'.lower()
    req_params = {
        'log_source_type': LogSourceType.SELECTED_CONTAINERS.value,
        'namespace': namespace,
        'config_name': collector_config_name,
        'bk_biz_id': bk_biz_id,
        'add_pod_label': True,
        'workload': {
            'name': 'nginx-deployment',
            'kind': SupportedWorkload.Deployment.value,
            'container_confs': [
                {'name': 'nginx', 'enable_stdout': False, 'log_paths': ['/log/access.log']},
                {'name': 'notify', 'enable_stdout': True},
            ],
        },
    }

    config_params = {
        'bk_biz_id': bk_biz_id,
        'project_id': project_id,
        'collector_config_name': collector_config_name,
        'collector_config_name_en': collector_config_name,
        'custom_type': 'log',
        'description': '',
        'bcs_cluster_id': cluster_id,
        'add_pod_label': req_params['add_pod_label'],
        'extra_labels': [],
        'config': [
            {
                'data_encoding': 'UTF-8',
                'namespaces': [namespace],
                'container': {
                    'workload_type': SupportedWorkload.Deployment.value,
                    'workload_name': 'nginx-deployment',
                    'container_name': 'nginx',
                },
                'enable_stdout': False,
                'paths': ['/log/access.log'],
            },
            {
                'data_encoding': 'UTF-8',
                'namespaces': [namespace],
                'container': {
                    'workload_type': SupportedWorkload.Deployment.value,
                    'workload_name': 'nginx-deployment',
                    'container_name': 'notify',
                },
                'enable_stdout': True,
            },
        ],
    }

    return {'req_params': req_params, 'config_params': config_params}


@pytest.fixture
def selectedlabels_create_params(project_id, cluster_id, namespace, bk_biz_id) -> Dict[str, Dict]:
    config_name = 'app_nginx_log'
    req_params = {
        'log_source_type': LogSourceType.SELECTED_LABELS.value,
        'namespace': namespace,
        'bk_biz_id': bk_biz_id,
        'config_name': config_name,
        'add_pod_label': True,
        'selector': {'enable_stdout': False, 'log_paths': ['/log/access.log'], 'match_labels': {'app': 'nginx'}},
    }

    config_params = {
        'bk_biz_id': bk_biz_id,
        'project_id': project_id,
        'collector_config_name': config_name,
        'collector_config_name_en': config_name,
        'custom_type': 'log',
        'description': '',
        'bcs_cluster_id': cluster_id,
        'add_pod_label': req_params['add_pod_label'],
        'extra_labels': [],
        'config': [
            {
                'data_encoding': 'UTF-8',
                'namespaces': [namespace],
                'enable_stdout': False,
                'label_selector': {'match_labels': [{'key': 'app', 'value': 'nginx'}]},
                'paths': ['/log/access.log'],
            }
        ],
    }

    return {'req_params': req_params, 'config_params': config_params}


@pytest.fixture
def log_collect_meta_data(project_id, cluster_id, namespace, config_id):
    return LogCollectMetadata.objects.create(
        project_id=project_id,
        cluster_id=cluster_id,
        log_source_type=LogSourceType.ALL_CONTAINERS.value,
        config_id=config_id,
        config_name=f'namespace_{namespace}_all_log',
    )


@pytest.fixture
def log_collect_meta_data_qset(project_id, cluster_id, namespace):
    LogCollectMetadata.objects.create(
        project_id=project_id,
        cluster_id=cluster_id,
        log_source_type=LogSourceType.ALL_CONTAINERS.value,
        namespace='',
        config_id=random.randint(1, 10000),
        config_name='cluster_scoped_all_log',
    )
    LogCollectMetadata.objects.create(
        project_id=project_id,
        cluster_id=cluster_id,
        log_source_type=LogSourceType.SELECTED_CONTAINERS.value,
        namespace=namespace,
        config_id=random.randint(1, 10000),
        config_name=f'deployment_nginx_log',
    )
    LogCollectMetadata.objects.create(
        project_id=project_id,
        cluster_id=cluster_id,
        log_source_type=LogSourceType.SELECTED_LABELS.value,
        namespace=namespace,
        config_id=random.randint(1, 10000),
        config_name=f'label_selector_{get_random_string(6)}_log'.lower(),
    )
    return LogCollectMetadata.objects.filter(project_id=project_id, cluster_id=cluster_id)


@pytest.fixture
def collect_conf_list(log_collect_meta_data_qset):
    return list_confs(log_collect_meta_data_qset)


@pytest.fixture
def allcontainers_update_params(project_id, cluster_id, namespace, bk_biz_id) -> Dict[str, Dict]:
    collector_config_name = f'namespace_{namespace}_all_log'

    req_params = {
        'log_source_type': LogSourceType.ALL_CONTAINERS.value,
        'namespace': namespace,
        'config_name': collector_config_name,
        'bk_biz_id': bk_biz_id,
        'add_pod_label': True,
        'extra_labels': {'env': 'test', 'app': 'nginx'},
        'base': {'enable_stdout': True, 'log_paths': ['/data/1.log', '/data/2.log']},
    }

    config_params = {
        'bk_biz_id': bk_biz_id,
        'project_id': project_id,
        'collector_config_name': collector_config_name,
        'description': '',
        'bcs_cluster_id': cluster_id,
        'add_pod_label': req_params['add_pod_label'],
        'extra_labels': [{'key': 'env', 'value': 'test'}, {'key': 'app', 'value': 'nginx'}],
        'config': [
            {
                'data_encoding': 'UTF-8',
                'namespaces': [namespace],
                'paths': req_params['base']['log_paths'],
                'enable_stdout': req_params['base']['enable_stdout'],
            }
        ],
    }

    return {'req_params': req_params, 'config_params': config_params}
