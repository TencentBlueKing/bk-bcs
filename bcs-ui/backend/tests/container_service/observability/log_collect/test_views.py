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
from unittest import mock

import pytest
from rest_framework.test import APIRequestFactory, force_authenticate

from backend.components.databus.collector import LogCollectorClient
from backend.container_service.observability.log_collect import views
from backend.container_service.observability.log_collect.constants import LogSourceType
from backend.container_service.observability.log_collect.models import LogCollectMetadata, LogIndexSet

factory = APIRequestFactory()

pytestmark = pytest.mark.django_db


class TestLogCollectViewSet:
    def test_create_allcontainers_config(
        self, bk_user, project_id, cluster_id, log_collect_result, allcontainers_create_params
    ):
        self._run_test_create_config(bk_user, project_id, cluster_id, log_collect_result, allcontainers_create_params)

    def test_create_selectedcontainers_config(
        self, bk_user, project_id, cluster_id, log_collect_result, selectedcontainers_create_params
    ):
        self._run_test_create_config(
            bk_user, project_id, cluster_id, log_collect_result, selectedcontainers_create_params
        )

    def test_create_selectedlabels_config(
        self, bk_user, project_id, cluster_id, log_collect_result, selectedlabels_create_params
    ):
        self._run_test_create_config(bk_user, project_id, cluster_id, log_collect_result, selectedlabels_create_params)

    def _run_test_create_config(self, bk_user, project_id, cluster_id, log_collect_result, create_params):
        t_view = views.LogCollectViewSet.as_view({'post': 'create'})
        request = factory.post(
            f'/api/log_collect/projects/{project_id}/clusters/{cluster_id}/configs/',
            create_params['req_params'],
        )
        force_authenticate(request, bk_user)

        with mock.patch.object(
            LogCollectorClient,
            'create_collect_config',
            return_value=log_collect_result,
        ) as create_collect_config:
            t_view(request, project_id=project_id, cluster_id=cluster_id)

            create_collect_config.assert_called_with(config=create_params['config_params'])
            assert LogCollectMetadata.objects.filter(config_id=log_collect_result['rule_id']).count() == 1
            assert (
                LogIndexSet.objects.filter(
                    project_id=project_id,
                    std_index_set_id=log_collect_result['std_index_set_id'],
                    file_index_set_id=log_collect_result['file_index_set_id'],
                ).count()
                == 1
            )

    def test_update_allcontainers_config(
        self,
        bk_user,
        project_id,
        cluster_id,
        config_id,
        log_collect_result,
        log_collect_meta_data,
        allcontainers_update_params,
    ):
        t_view = views.LogCollectViewSet.as_view({'put': 'update'})
        request = factory.put(
            f'/api/log_collect/projects/{project_id}/clusters/{cluster_id}/configs/{config_id}/',
            allcontainers_update_params['req_params'],
        )
        force_authenticate(request, bk_user)

        with mock.patch.object(
            LogCollectorClient,
            'update_collect_config',
            return_value=log_collect_result,
        ) as update_collect_config:
            t_view(request, project_id=project_id, cluster_id=cluster_id, pk=config_id)

            update_collect_config.assert_called_with(
                config_id=config_id, config=allcontainers_update_params['config_params']
            )
            assert (
                LogCollectMetadata.objects.get(
                    config_id=config_id, project_id=project_id, cluster_id=cluster_id
                ).updator
                == bk_user.username
            )

    def test_delete_config(
        self, bk_user, project_id, cluster_id, config_id, log_collect_result, log_collect_meta_data
    ):
        t_view = views.LogCollectViewSet.as_view({'delete': 'destroy'})
        request = factory.delete(f'/api/log_collect/projects/{project_id}/clusters/{cluster_id}/configs/{config_id}/')
        force_authenticate(request, bk_user)

        with mock.patch.object(
            LogCollectorClient,
            'delete_collect_config',
            return_value=log_collect_result,
        ) as delete_collect_config:
            t_view(request, project_id=project_id, cluster_id=cluster_id, pk=config_id)

            delete_collect_config.assert_called_with(config_id=config_id)
            assert (
                LogCollectMetadata.objects.filter(
                    config_id=config_id, project_id=project_id, cluster_id=cluster_id
                ).exists()
                is False
            )

    def test_list_configs(self, bk_user, project_id, cluster_id, namespace, collect_conf_list):
        t_view = views.LogCollectViewSet.as_view({'get': 'list'})
        request = factory.get(f'/api/log_collect/projects/{project_id}/clusters/{cluster_id}/configs/')
        force_authenticate(request, bk_user)

        with mock.patch.object(
            LogCollectorClient,
            'list_collect_configs',
            return_value=collect_conf_list,
        ):
            resp = t_view(request, project_id=project_id, cluster_id=cluster_id)
            configs = resp.data

            selectedlabels_conf = configs[0]
            selectedcontainers_conf = configs[1]
            allcontainers_conf = configs[2]

            assert selectedlabels_conf['log_source_type'] == LogSourceType.SELECTED_LABELS
            assert selectedlabels_conf['namespace'] == namespace
            selector = selectedlabels_conf['selector']
            assert selector['match_labels'] == {'app': 'nginx'}

            assert selectedcontainers_conf['log_source_type'] == LogSourceType.SELECTED_CONTAINERS
            assert selectedcontainers_conf['add_pod_label']
            assert selectedcontainers_conf['extra_labels'] == {'app': 'nginx'}
            workload = selectedcontainers_conf['workload']
            assert workload['name'] == 'nginx-deployment'
            assert workload['container_confs'][0]['log_paths'] == ['/log/access.log']
            assert workload['container_confs'][1]['name'] == 'notify'

            assert allcontainers_conf['log_source_type'] == LogSourceType.ALL_CONTAINERS
            assert allcontainers_conf['add_pod_label']
            assert allcontainers_conf['base']['enable_stdout']
