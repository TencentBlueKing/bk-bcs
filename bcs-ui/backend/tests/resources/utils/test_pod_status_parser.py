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
import pytest

from backend.resources.constants import SimplePodStatus
from backend.resources.workloads.pod.utils import PodStatusParser
from backend.tests.resources.utils.contents.pod_configs import (
    CompletedStatusPodConfig,
    CreateContainerErrorStatusPodConfig,
    FailedStatusPodConfig,
    PendingStatusPodConfig,
    RunningStatusPodConfig,
    SucceededStatusPodConfig,
    TerminatingStatusPodConfig,
    UnknownStatusPodConfig,
)


@pytest.mark.parametrize(
    'config, expected_status',
    [
        (FailedStatusPodConfig, SimplePodStatus.PodFailed.value),
        (SucceededStatusPodConfig, SimplePodStatus.PodSucceeded.value),
        (RunningStatusPodConfig, SimplePodStatus.PodRunning.value),
        (PendingStatusPodConfig, SimplePodStatus.PodPending.value),
        (TerminatingStatusPodConfig, SimplePodStatus.Terminating.value),
        (UnknownStatusPodConfig, SimplePodStatus.PodUnknown.value),
        (CompletedStatusPodConfig, SimplePodStatus.Completed.value),
        (CreateContainerErrorStatusPodConfig, 'CreateContainerError'),
    ],
)
def test_pod_status_parser(config, expected_status):
    """测试 Pod 状态解析逻辑"""
    actual_status = PodStatusParser(config).parse()
    assert actual_status == expected_status
