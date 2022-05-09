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

from attr import dataclass

from backend.resources.constants import ConditionStatus, PodConditionType, SimplePodStatus
from backend.utils.basic import getitems


@dataclass
class PodStatusParser:
    """
    Pod 状态解析器，解析逻辑参考
    kubernetes/dashboard getPodStatus
    https://github.com/kubernetes/dashboard/blob/master/src/app/backend/resource/pod/common.go#L40
    """

    pod: Dict
    initializing: bool = False
    tol_status: str = SimplePodStatus.PodUnknown.value

    def parse(self) -> str:
        """获取 Pod 总状态"""
        # 1. 默认使用 Pod.Status.Phase
        self.tol_status = getitems(self.pod, 'status.phase')
        # 2. 若有具体的 Pod.Status.Reason 则使用
        if getitems(self.pod, 'status.reason'):
            self.tol_status = getitems(self.pod, 'status.reason')

        # 3. 根据 Pod 容器状态更新状态
        self._update_status_by_init_container_statuses()
        if not self.initializing:
            self._update_status_by_container_statuses()

        # 4. 根据 Pod.Metadata.DeletionTimestamp 更新状态
        if getitems(self.pod, 'metadata.deletionTimestamp'):
            if getitems(self.pod, 'status.reason') == 'NodeLost':
                self.tol_status = SimplePodStatus.PodUnknown.value
            else:
                self.tol_status = SimplePodStatus.Terminating.value

        # 5. 若状态未初始化或在转移中丢失，则标记为未知状态
        if not self.tol_status:
            self.tol_status = SimplePodStatus.PodUnknown.value

        return self.tol_status

    def _update_status_by_init_container_statuses(self):
        """根据 pod.Status.InitContainerStatuses 更新 总状态"""
        for idx, container in enumerate(getitems(self.pod, 'status.initContainerStatuses', [])):
            # 检查每个容器的 state.terminated.exitCode 判断是否初始化完成
            term_exit_code = getitems(container, 'state.terminated.exitCode')
            if term_exit_code == 0:
                continue

            # 只要有不是 term_exit_code 为 0 的，说明正在初始化流程中
            self.initializing = True
            if getitems(container, 'state.terminated'):
                # terminated 状态优先级 reason > signal > exit_code
                term_reason = getitems(container, 'state.terminated.reason')
                term_signal = getitems(container, 'state.terminated.signal')
                if term_reason:
                    self.tol_status = f'Init: {term_reason}'
                elif term_signal:
                    self.tol_status = f'Init: Signal {term_signal}'
                else:
                    self.tol_status = f"Init: ExitCode {term_exit_code}"
            else:
                waiting_reason = getitems(container, 'state.waiting.reason')
                if waiting_reason and waiting_reason != 'PodInitializing':
                    self.tol_status = f"Init: {waiting_reason}"
                else:
                    self.tol_status = f"Init: {idx}/{len(getitems(self.pod, 'spec.initContainers', []))}"
            break

    def _has_pod_ready_condition(self, conditions: List[Dict]) -> bool:
        """检查 pod condition Ready 状态"""
        for c in conditions:
            if (
                c.get('type') == PodConditionType.PodReady.value
                and c.get('status') == ConditionStatus.ConditionTrue.value
            ):
                return True
        return False

    def _update_status_by_container_statuses(self):
        """根据 pod.Status.ContainerStatuses 更新 总状态"""
        hasRunning = False
        for container in reversed(getitems(self.pod, 'status.containerStatuses', [])):
            waiting_reason = getitems(container, 'state.waiting.reason')
            if waiting_reason:
                self.tol_status = waiting_reason

            elif getitems(container, 'state.terminated'):
                # terminated 状态优先级 reason > signal > exit_code
                term_reason = getitems(container, 'state.terminated.reason')
                term_signal = getitems(container, 'state.terminated.signal')
                if term_reason:
                    self.tol_status = term_reason
                elif term_signal:
                    self.tol_status = f"Signal: {term_signal}"
                else:
                    self.tol_status = f"ExitCode: {getitems(container, 'state.terminated.exitCode')}"

            elif container.get('ready') and getitems(container, 'state.running'):
                hasRunning = True

        if self.tol_status == SimplePodStatus.Completed.value and hasRunning:
            if self._has_pod_ready_condition(getitems(self.pod, 'status.conditions', [])):
                self.tol_status = SimplePodStatus.PodRunning.value
            else:
                self.tol_status = SimplePodStatus.NotReady.value
