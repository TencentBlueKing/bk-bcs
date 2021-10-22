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
import os
import time
from unittest import mock

import pytest

from backend.container_service.clusters import models
from backend.container_service.clusters.tasks import ClusterOrNodeTaskPoller, TaskStatusResultHandler
from backend.packages.blue_krill.async_utils.poll_task import (
    CallbackResult,
    CallbackStatus,
    PollingMetadata,
    TaskPoller,
)
from backend.tests.bcs_mocks.fake_ops import FakeOPSClient

pytestmark = pytest.mark.django_db


@pytest.fixture(autouse=True)
def cluster_record():
    models.ClusterInstallLog.objects.create(
        id=1,
        project_id="project_id",
        cluster_id="cluster_id",
        task_id="task_id",
        token="access_token",
        operator="admin",
        params='{"cc_app_id": 1}',
        log='{}',
        is_finished=False,
        is_polling=True,
        status="RUNNING",
        oper_type="initialize",
    )


# 社区版不需要执行以下单元测试
@pytest.mark.skipif(not os.environ.get('IS_INTERNAL_MODE'), reason='skip internal only unittest')
class TestClusterOrNodeTaskPoller:
    @mock.patch(
        "backend.container_service.clusters.tasks.ops.OPSClient",
        new=FakeOPSClient,
    )
    @mock.patch(
        "backend.container_service.clusters.tasks.paas_auth.get_access_token",
        return_values={"access_token": "access_token"},
    )
    def test_task_query(self, get_access_token):
        record = models.ClusterInstallLog.objects.get(id=1)
        assert record.project_id == "project_id"

        started_at = time.time()
        metadata = PollingMetadata(retries=0, query_started_at=started_at, queried_count=0)
        poller = ClusterOrNodeTaskPoller(params={"model_type": "ClusterInstallLog", "pk": 1}, metadata=metadata)
        task_record = poller.get_task_record()
        assert task_record.project_id == "project_id"

        step_logs, status, _ = poller._get_task_result(record)
        assert step_logs == [{"state": "RUNNING", "name": "- Step one"}, {"state": "FAILURE", "name": "- Step two"}]
        assert status == "FAILURE"


class TestTaskStatusResultHandler:
    @mock.patch("backend.container_service.clusters.tasks.update_status", return_values=None)
    def test_callback(self, status):
        callback_result = CallbackResult(status=CallbackStatus.TIMEOUT.value)
        started_at = time.time()
        metadata = PollingMetadata(retries=0, query_started_at=started_at, queried_count=0)
        poller = ClusterOrNodeTaskPoller(params={"model_type": "ClusterInstallLog", "pk": 1}, metadata=metadata)
        TaskStatusResultHandler().handle(callback_result, poller)

        record = models.ClusterInstallLog.objects.get(id=1)
        assert record.status == models.CommonStatus.InitialFailed
