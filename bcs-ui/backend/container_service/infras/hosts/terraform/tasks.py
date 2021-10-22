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
import json
import logging
from typing import Dict

from backend.container_service.infras.hosts.terraform.engines.sops import get_task_state_and_steps
from backend.packages.blue_krill.async_utils.poll_task import (
    CallbackHandler,
    CallbackResult,
    CallbackStatus,
    PollingResult,
    PollingStatus,
    TaskPoller,
)
from backend.utils.error_codes import APIError

from .constants import TaskStatus
from .models import HostApplyTaskLog

logger = logging.getLogger(__name__)


class ApplyHostStatusPoller(TaskPoller):
    """申请主机任务启动后，轮训申请主机资源的状态"""

    default_retry_delay_seconds = 20
    overall_timeout_seconds = 3600 * 24 * 2  # 超时时间设置为2天

    def query(self) -> PollingResult:
        params = self.params
        state_and_steps = self.query_state_and_steps(params["task_id"])

        status = PollingStatus.DOING.value
        if state_and_steps["state"] in [
            TaskStatus.FAILED.value,
            TaskStatus.REVOKED.value,
            TaskStatus.FINISHED.value,
        ]:
            status = PollingStatus.DONE.value

        return PollingResult(status=status, data=state_and_steps)

    def query_state_and_steps(self, task_id: str) -> Dict:
        try:
            state_and_steps = get_task_state_and_steps(task_id=task_id)
        except APIError as e:
            logger.error("request sops task status error, %s", e)
            return {}
        return state_and_steps


class ApplyHostStatusResultHandler(CallbackHandler):
    """处理最终状态，更新db记录"""

    def handle(self, result: CallbackResult, poller: TaskPoller):
        poll_data = result.data
        if result.status != CallbackStatus.NORMAL.value:
            status = TaskStatus.FAILED.value
        else:
            status = poll_data["state"]

        self.update_task_log(log_id=poller.params["log_id"], status=status, logs=poll_data.get("steps"))

    def update_task_log(self, log_id: int, status: str, logs: Dict):
        try:
            log = HostApplyTaskLog.objects.get(id=log_id)
        except HostApplyTaskLog.DoesNotExist:
            logger.error("HostApplyTaskLog not found record: %s", log_id)
            return
        # 更新任务记录状态及log参数
        log.status = status
        log.is_finished = True
        log.logs = logs
        log.save(update_fields=["status", "is_finished", "logs", "updated"])
