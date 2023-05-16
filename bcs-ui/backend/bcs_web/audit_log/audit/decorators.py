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
import logging
from abc import ABCMeta, abstractmethod
from typing import Optional, Tuple, Type

import wrapt
from rest_framework.exceptions import ValidationError

from backend.metrics import Result, counter_inc
from backend.packages.blue_krill.web.std_error import APIError

from .auditors import Auditor
from .context import AuditContext

logger = logging.getLogger(__name__)


class BaseLogAudit(metaclass=ABCMeta):
    """
    带参数的审计装饰器(抽象基类)

    :param auditor_cls: 执行审计记录的类, 默认为 Auditor
    :param activity_type: 操作类型，默认为''。可在 audit_ctx 中覆盖
    :param auto_audit: 是否记录审计，默认为记录
    :param ignore_exceptions: 忽略审计的异常类列表。忽略父类后，子类会一并忽略
    """

    # 记录原始错误信息的异常列表
    err_msg_exceptions = (APIError, ValidationError)

    def __init__(
        self,
        auditor_cls: Type[Auditor] = type(Auditor),
        activity_type: str = '',
        auto_audit: bool = True,
        use_raw_audit: bool = False,
        ignore_exceptions: Optional[Tuple[Type[Exception]]] = None,
    ):

        self.auditor_cls = auditor_cls
        self.activity_type = activity_type
        self.auto_audit = auto_audit
        self.use_raw_audit = use_raw_audit
        self.ignore_exceptions = ignore_exceptions or tuple()

    @wrapt.decorator
    def __call__(self, wrapped, instance, args, kwargs):
        audit_ctx = self._pre_audit_ctx(instance, *args, **kwargs)
        err_msg = ''

        auto_audit = self.auto_audit

        try:
            ret = wrapped(*args, **kwargs)
            return ret
        except self.ignore_exceptions:
            # 如果是 ignore_exceptions 中的异常，不做审计记录
            auto_audit = False
            raise
        except self.err_msg_exceptions as e:
            err_msg = str(e)
            raise
        except Exception as e:
            # 屏蔽非预期的异常信息
            logger.error("log audit failed: %s" % e)
            err_msg = "unknown error"
            raise
        finally:
            if auto_audit:
                if self.use_raw_audit:
                    self._save_raw_audit(audit_ctx)
                else:
                    audit_ctx = self._post_audit_ctx(audit_ctx, *args, **kwargs)
                    self._save_audit(audit_ctx, err_msg)

    @abstractmethod
    def _pre_audit_ctx(self, instance, *args, **kwargs) -> AuditContext:
        """前置获取初始 audit_ctx"""

    def _post_audit_ctx(self, audit_ctx: AuditContext, *args, **kwargs) -> AuditContext:
        """后置更新 audit_ctx"""
        return audit_ctx

    def _save_audit(self, audit_ctx: AuditContext, err_msg: str):
        """审计内容入库: 如果 err_msg 有错误信息，则审计状态标记为 failed; 否则 succeed"""
        auditor = self.auditor_cls(audit_ctx)
        if err_msg:
            auditor.log_failed(err_msg)
        else:
            auditor.log_succeed()

    def _save_raw_audit(self, audit_ctx: AuditContext):
        """保存原始的审计信息"""
        self.auditor_cls(audit_ctx).log_raw()


class log_audit_on_view(BaseLogAudit):
    """
    用于 view 的操作审计装饰器

    使用示例:
    class TemplatesetsViewSet(SystemViewSet):

        @log_audit_on_view(TemplatesetsAuditor, activity_type='create')
        def create(self, request, project_id):
            request.audit_ctx.update_fields(resource='nginx')
            return Response()
    """

    def _pre_audit_ctx(self, instance, *args, **kwargs) -> AuditContext:
        request = args[0]
        if hasattr(request, 'audit_ctx'):
            request.audit_ctx.update_fields(activity_type=self.activity_type)
        else:
            request.audit_ctx = AuditContext(
                user=request.user.username,
                project_id=self._get_project_id(request, **kwargs),
                activity_type=self.activity_type,
            )
        return request.audit_ctx

    def _post_audit_ctx(self, audit_ctx: AuditContext, *args, **kwargs) -> AuditContext:
        """根据请求参数，生成默认的extra"""
        if audit_ctx.extra:
            return audit_ctx

        # TODO 优化默认 extra 的构成
        request = args[0]
        extra = dict(**kwargs)
        if hasattr(request, 'data'):
            if isinstance(request.data, dict):
                extra.update(request.data)
            elif isinstance(request.data, str):
                extra['request_body'] = request.data

        audit_ctx.extra = extra
        return audit_ctx

    def _get_project_id(self, request, **kwargs) -> str:
        if hasattr(request, 'project'):
            return request.project.project_id
        return kwargs.get('project_id', '')


class log_audit(BaseLogAudit):
    """
    用于一般类实例方法或函数的操作审计装饰器。使用规则:
    - 对于类实例方法，第二个位置参数是 AuditContext 实例或者 self.audit_ctx (如果没有会创建)
    - 对于普通方法，通常需要第一个位置参数是 AuditContext 实例

    使用示例:

    @log_audit(HelmAuditor, activity_type='install')
    def install_chart(audit_ctx: AuditContext):
        audit_ctx.update_fields(
            description='test install helm', extra={'chart': 'http://example.chart.com/nginx/nginx1.12.tgz'}
        )
    """

    def _pre_audit_ctx(self, instance, *args, **kwargs) -> AuditContext:

        if instance is None:  # 被装饰者为类/函数/静态方法
            # 只处理函数/静态方法的情况
            if len(args) <= 0:
                raise TypeError('missing AuditContext instance argument')
            # 第一个参数是 AuditContext 类型
            if isinstance(args[0], AuditContext):
                audit_ctx = args[0]
            else:
                raise TypeError('missing AuditContext instance argument')
        else:  # 被装饰者为 classmethod 类方法 或者普通类方法
            audit_ctx = self._get_or_create_audit_ctx(instance)

        if not audit_ctx.activity_type:
            audit_ctx.activity_type = self.activity_type
        return audit_ctx

    def _get_or_create_audit_ctx(self, instance) -> AuditContext:
        if hasattr(instance, 'audit_ctx') and isinstance(instance.audit_ctx, AuditContext):
            return instance.audit_ctx

        audit_ctx = AuditContext()
        instance.audit_ctx = audit_ctx
        return audit_ctx
