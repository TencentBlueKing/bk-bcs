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

import mock
import pytest
from rest_framework.response import Response
from rest_framework.test import APIRequestFactory, force_authenticate
from rest_framework.validators import ValidationError

from backend.bcs_web.audit_log.audit.auditors import HelmAuditor
from backend.bcs_web.audit_log.audit.context import AuditContext
from backend.bcs_web.audit_log.audit.decorators import log_audit, log_audit_on_view
from backend.bcs_web.audit_log.constants import ActivityStatus, ActivityType, ResourceType
from backend.bcs_web.audit_log.models import UserActivityLog
from backend.templatesets.legacy_apps.configuration.auditor import TemplatesetAuditor
from backend.tests.testing_utils.mocks.viewsets import FakeSystemViewSet

pytestmark = pytest.mark.django_db

factory = APIRequestFactory()


class TemplatesetsViewSet(FakeSystemViewSet):
    @log_audit_on_view(TemplatesetAuditor, activity_type=ActivityType.Retrieve)
    def list(self, request, project_id):
        return Response()

    @log_audit_on_view(TemplatesetAuditor, activity_type=ActivityType.Add)
    def create(self, request, project_id):
        request.audit_ctx.update_fields(resource='nginx')
        raise ValidationError('invalid manifest')

    @log_audit_on_view(TemplatesetAuditor, activity_type=ActivityType.Delete, ignore_exceptions=(ValidationError,))
    def delete(self, request, project_id):
        raise ValidationError('test')


@log_audit(HelmAuditor, activity_type=ActivityType.Add)
def install_chart(audit_ctx: AuditContext):
    audit_ctx.update_fields(
        description=f'test {ActivityType.Add} {ResourceType.HelmApp}',
        extra={'chart': 'http://example.chart.com/nginx/nginx1.12.tgz'},
    )


class HelmViewSet(FakeSystemViewSet):
    def create(self, request, project_id):
        install_chart(AuditContext(user=request.user.username, project_id=project_id))
        return Response()

    def upgrade(self, request, project_id):
        self._upgrade(request, project_id)
        return Response()

    @log_audit(HelmAuditor, activity_type=ActivityType.Modify)
    def _upgrade(self, request, project_id):
        self.audit_ctx.update_fields(
            project_id=project_id,
            user=request.user.username,
            description=f'test {ActivityType.Modify} {ResourceType.HelmApp}',
            extra={'chart': 'http://example.chart.com/nginx/nginx1.12.tgz'},
        )


class TestAuditDecorator:
    def test_log_audit_on_view_succeed(self, bk_user, project_id):
        t_view = TemplatesetsViewSet.as_view({'get': 'list'})
        request = factory.get('/')
        force_authenticate(request, bk_user)
        t_view(request, project_id=project_id)

        activity_log = UserActivityLog.objects.get(
            project_id=project_id, user=bk_user.username, activity_type=ActivityType.Retrieve
        )
        assert activity_log.activity_status == ActivityStatus.Succeed
        assert activity_log.description == f'查询 模板集 成功'

    def test_log_audit_on_view_failed(self, bk_user, project_id):
        t_view = TemplatesetsViewSet.as_view({'post': 'create'})
        request = factory.post('/', data={'version': '1.6.0'})
        force_authenticate(request, bk_user)
        t_view(request, project_id=project_id)

        activity_log = UserActivityLog.objects.get(
            project_id=project_id, user=bk_user.username, activity_type=ActivityType.Add
        )
        assert activity_log.activity_status == ActivityStatus.Failed
        assert json.loads(activity_log.extra)['version'] == '1.6.0'
        assert activity_log.description == f"创建 模板集 nginx 失败: {ValidationError('invalid manifest')}"

    def test_log_audit_ignore_exceptions(self, bk_user, project_id):
        t_view = TemplatesetsViewSet.as_view({'delete': 'delete'})
        request_post = factory.delete('/')
        force_authenticate(request_post, bk_user)
        try:
            t_view(request_post, project_id=project_id)
        except Exception:
            pass

        assert (
            UserActivityLog.objects.filter(
                project_id=project_id, user=bk_user.username, activity_type=ActivityType.Delete
            ).count()
            == 0
        )

    def test_log_audit_for_func(self, bk_user, project_id):
        h_view = HelmViewSet.as_view({'post': 'create'})
        request = factory.post('/')
        force_authenticate(request, bk_user)
        h_view(request, project_id=project_id)

        activity_log = UserActivityLog.objects.get(
            project_id=project_id, user=bk_user.username, resource_type=ResourceType.HelmApp
        )
        assert activity_log.activity_type == ActivityType.Add
        assert (
            activity_log.description == f'test {ActivityType.Add} {ResourceType.HelmApp} '
            f'{ActivityStatus.get_choice_label(ActivityStatus.Succeed)}'
        )
        assert json.loads(activity_log.extra)['chart'] == 'http://example.chart.com/nginx/nginx1.12.tgz'

    def test_log_audit_for_method(self, bk_user, project_id):
        h_view = HelmViewSet.as_view({'put': 'upgrade'})
        request = factory.put('/')
        force_authenticate(request, bk_user)
        h_view(request, project_id=project_id)

        activity_log = UserActivityLog.objects.get(
            project_id=project_id, user=bk_user.username, resource_type=ResourceType.HelmApp
        )
        assert activity_log.activity_type == ActivityType.Modify
        assert (
            activity_log.description == f'test {ActivityType.Modify} {ResourceType.HelmApp} '
            f'{ActivityStatus.get_choice_label(ActivityStatus.Succeed)}'
        )
        assert json.loads(activity_log.extra)['chart'] == 'http://example.chart.com/nginx/nginx1.12.tgz'
