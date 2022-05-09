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

from django.utils.translation import ugettext_lazy as _
from rest_framework.exceptions import ValidationError
from rest_framework.response import Response

from backend.bcs_web.audit_log.client import ContextActivityLogClient
from backend.bcs_web.viewsets import SystemViewSet
from backend.container_service.infras.hosts.terraform.engines.sops import HostData, create_and_start_host_application
from backend.container_service.projects.cmdb import is_biz_maintainer
from backend.utils.exceptions import PermissionDeniedError

from .constants import SCR_URL, TaskStatus
from .models import HostApplyTaskLog
from .serializers import ApplyHostDataSLZ, TaskLogSLZ
from .tasks import ApplyHostStatusPoller, ApplyHostStatusResultHandler

logger = logging.getLogger(__name__)


class ApplyHostViewSet(SystemViewSet):
    def verify_task_exist(self, project_id):
        if HostApplyTaskLog.objects.filter(project_id=project_id, is_finished=False).exists():
            raise ValidationError(_("项目下正在申请机器资源，请等待任务结束后，再确认是否继续申请资源!"))

    def can_apply_host(self, biz_id: int, username: str):
        # 仅运维角色人员才能申请主机
        if not is_biz_maintainer(biz_id, username):
            raise PermissionDeniedError(_("用户【{}】不是业务的运维角色，请联系业务运维申请机器!").format(username), "")

    def apply_host(self, request, project_id):
        # NOTE: 项目下如果有在申请的任务，必须等任务结束后，才能再次申请
        self.verify_task_exist(project_id)
        username = request.user.username
        self.can_apply_host(int(request.project.cc_app_id), username)
        slz = ApplyHostDataSLZ(data=request.data)
        slz.is_valid(raise_exception=True)
        data = slz.validated_data
        # 组装申请主机需要的信息
        host_data = HostData.from_dict(data)

        with ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type="instance",
            resource="apply host",
            extra=data,
            description=_("申请主机"),
        ).log_add():
            task_id, task_url = create_and_start_host_application(request.project.cc_app_id, username, host_data)

        # 存储数据，用于后续的查询
        log_record = HostApplyTaskLog.objects.create(
            project_id=project_id,
            task_id=task_id,
            task_url=task_url,
            status=TaskStatus.RUNNING.value,
            operator=username,
            params=data,
            logs={"申请主机": {"state": "RUNNING"}},
        )
        # 启动轮询任务
        ApplyHostStatusPoller.start({"log_id": log_record.id, "task_id": task_id}, ApplyHostStatusResultHandler)
        return Response()

    def get_task_log(self, request, project_id):
        """查询log，用于展示"""
        # 获取最新的
        task_log = HostApplyTaskLog.objects.filter(project_id=project_id).last()
        if not task_log:
            return Response()
        # 组装返回数据
        data = TaskLogSLZ(task_log).data
        data["scr_url"] = SCR_URL
        # TODO: 后续是否把IP直接提示在否个地方
        return Response(data)


try:
    from .views_ext import CVMTypeViewSet, DiskTypeViewSet, ZoneViewSet
except ImportError as e:
    logger.debug("Load extension failed: %s", e)
