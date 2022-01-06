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
from typing import Any, Dict, List, NewType, Optional, Tuple, Union

from django.conf import settings

from backend.components import base as comp_base
from backend.components import ops, paas_auth, paas_cc
from backend.container_service.clusters import models
from backend.packages.blue_krill.async_utils import poll_task

TASK_FAILED_STATUS_LIST = ["FAILURE", "REVOKED", "FAILED"]
TASK_SUCCESS_STATUS_LIST = ["SUCCESS", "FINISHED"]

ModelLogRecord = NewType("ModelLogRecord", Union[models.ClusterInstallLog, models.NodeUpdateLog])

logger = logging.getLogger(__name__)


class ClusterOrNodeTaskPoller(poll_task.TaskPoller):
    """轮训集群及节点添加、删除任务的状态"""

    default_retry_delay_seconds = getattr(settings, "POLLING_INTERVAL_SECONDS", 10)
    overall_timeout_seconds = getattr(settings, "POLLING_TIMEOUT_SECONDS", 3600)

    def query(self) -> poll_task.PollingResult:
        # 获取任务记录
        record = self.get_task_record()
        # 解析任务
        step_logs, status, tke_cluster_id = self._get_task_result(record)
        # 更新记录字段
        self._update_record_fields(record, status, step_logs, tke_cluster_id)
        polling_status = poll_task.PollingStatus.DOING.value
        if record.is_finished:
            polling_status = poll_task.PollingStatus.DONE.value

        return poll_task.PollingResult(status=polling_status)

    def get_task_record(self) -> Optional[ModelLogRecord]:
        """获取task记录"""
        params = self.params
        # 任务类型: cluster/node
        model_type = params["model_type"]
        # 任务记录的ID
        task_record_id = params["pk"]
        task_model = models.log_factory(model_type)
        if not task_model:
            logger.error(f'not found {model_type} task')
            return

        # 获取记录
        try:
            record = task_model.objects.get(pk=task_record_id)
        except task_model.DoesNotExist:
            logger.error(f'not found task: {task_record_id}')
            return
        # 判断任务是否结束
        if record.is_finished:
            logger.info(f'record: {task_record_id} has been finished')
            return record
        return record

    def _parse_steps(self, data: Dict) -> List:
        step_logs = data.get("steps") or []
        logs, failed_logs = [], []
        # 获取返回的任务日志
        for log in step_logs:
            # 渲染步骤名称，用于前端展示
            step_status = log.get("status")
            step_name_status = {"state": step_status, "name": "- %s" % (log.get("name", "").capitalize())}
            # NOTE: 兼容先前，并行时，返回的日志，可能正常和错误交叉，为方便前端展示，需要处理为前面步骤为成功，后面步骤为失败
            if step_status in TASK_FAILED_STATUS_LIST:
                failed_logs.append(step_name_status)
            else:
                logs.append(step_name_status)
        # 合并子步骤
        logs.extend(failed_logs)
        return logs

    def _get_task_result(self, record: ModelLogRecord) -> Tuple[List, str, str]:
        """解析任务状态
        兼容先前的逻辑, 返回日志格式{"state": "任务总状态", "node_tasks": "子步骤logs"}
        """
        token = paas_auth.get_access_token()
        task_result = get_task_result(token["access_token"], record)
        # 获取状态及步骤
        data = task_result.get("data") or {}
        return self._parse_steps(data), data.get("status", ""), data.get("extra_cluster_id", "")

    def _transform_task_status(self, status: str, record: ModelLogRecord):
        """转换任务状态，用以前端展示"""
        # 如果任务失败，则需要根据操作类型转换状态
        # 如果任务成功，则转换为normal状态
        # 否则，状态为任务初始化时的running状态
        if status in TASK_FAILED_STATUS_LIST:
            record.status = transform_failed_status_by_op_type(record.oper_type)
        elif status in TASK_SUCCESS_STATUS_LIST:
            record.status = models.CommonStatus.Normal

    def _is_task_finished(self, status: str) -> bool:
        """判断任务是否结束"""
        if status in TASK_FAILED_STATUS_LIST or status in TASK_SUCCESS_STATUS_LIST:
            return True
        return False

    def _update_record_fields(
        self,
        record: ModelLogRecord,
        status: str,
        step_logs: List,
        tke_cluster_id: str = "",
    ):
        # 转换任务状态
        self._transform_task_status(status, record)
        # 处于失败或成功状态时，更新任务结束标志位
        if self._is_task_finished(status):
            record.is_finished, record.is_polling = True, False
        if step_logs:
            record.log = json.dumps({"state": status, "node_tasks": step_logs})
        # 添加tke集群的cluster_id
        log_params = record.log_params
        # 如果配置不存在，则不作调整
        if "config" in log_params:
            log_params["config"].update({"extra_cluster_id": tke_cluster_id})
            record.params = json.dumps(log_params)
        record.save(update_fields=["status", "is_finished", "is_polling", "params", "log", "update_at"])


class TaskStatusResultHandler(poll_task.CallbackHandler):
    """处理超时状态，更新db记录及bcs cc中状态"""

    def handle(self, result: poll_task.CallbackResult, poller: poll_task.TaskPoller):
        record = poller.get_task_record()
        if result.status == poll_task.CallbackStatus.TIMEOUT.value:
            record.set_finish_polling_status(
                finish_flag=True, polling_flag=False, status=transform_failed_status_by_op_type(record.oper_type)
            )
        # 更新bcs cc中集群或节点状态
        update_status(poller.params["model_type"], record)


def get_task_result(access_token: str, record: ModelLogRecord) -> Dict[str, Any]:
    params = json.loads(record.params)
    return ops.get_task_result(
        access_token, record.project_id, record.task_id, params.get("cc_app_id"), params.get("username")
    )


class StatusUpdater:
    def __init__(self, record: ModelLogRecord):
        self.record = record

    @property
    def _bcs_cc_client(self) -> comp_base.BkApiClient:
        token = paas_auth.get_access_token()
        access_token = token["access_token"]
        return paas_cc.PaaSCCClient(comp_base.ComponentAuth(access_token=access_token))

    @property
    def _bcs_cc_record_status(self) -> str:
        """获取存储在bcs cc的记录的状态"""
        if self.record.oper_type not in [models.ClusterOperType.ClusterRemove, models.NodeOperType.NodeRemove]:
            return self.record.status
        # 针对删除操作的处理
        if self.record.status == models.CommonStatus.Normal:
            return models.CommonStatus.Removed
        return self.record.status

    def update_status(self):
        """更新状态"""
        pass


class ClusterStatusUpdater(StatusUpdater):
    """集群任务状态的操作"""

    def update_status(self):
        record = self.record
        bcs_cc_client = self._bcs_cc_client
        project_id, cluster_id = record.project_id, record.cluster_id
        # 更新集群状态
        params = json.loads(record.params)
        status = self._bcs_cc_record_status
        req_data = {"status": status, "extra_cluster_id": params.get("config", {}).get("extra_cluster_id")}
        bcs_cc_client.update_cluster(project_id, cluster_id, req_data)
        logger.info(f"Update cluster[{cluster_id}] success")
        # 如果当前为删除操作，并且检测到状态处于已移除，则需要调用bcs cc接口删除存储的集群记录
        if (record.oper_type == models.ClusterOperType.ClusterRemove) and (status == models.CommonStatus.Removed):
            # 调用接口，删除集群
            bcs_cc_client.delete_cluster(project_id, cluster_id)
            logger.info(f"Delete cluster[{cluster_id}] success")


class NodeStatusUpdater(StatusUpdater):
    """节点任务状态的操作"""

    def update_status(self):
        record = self.record
        params = json.loads(record.params)
        ip_list = list(params.get("node_info", {}).keys())
        # 更新节点状态
        self._bcs_cc_client.update_node_list(
            record.project_id,
            record.cluster_id,
            [paas_cc.UpdateNodesData(inner_ip=ip, status=self._bcs_cc_record_status) for ip in ip_list],
        )
        logger.info(f'Update node[{json.dumps(ip_list)}] status success')


def update_status(model_type: str, record: ModelLogRecord):
    updater = ClusterStatusUpdater if model_type == "ClusterInstallLog" else NodeStatusUpdater
    updater(record).update_status()


def transform_failed_status_by_op_type(op_type: str) -> str:
    """处理异常时，针对不同操作的状态"""
    if op_type in [models.ClusterOperType.ClusterRemove, models.NodeOperType.NodeRemove]:
        status = models.CommonStatus.RemoveFailed
    elif op_type in [models.ClusterOperType.ClusterUpgrade, models.ClusterOperType.ClusterReupgrade]:
        status = models.ClusterStatus.UpgradeFailed
    else:
        status = models.CommonStatus.InitialFailed
    return status


try:
    from .tasks_ext import get_task_result  # noqa
except ImportError as e:
    logger.debug("Load extension failed: %s", e)
