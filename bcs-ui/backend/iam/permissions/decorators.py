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
from typing import Dict, List, Type

import wrapt
from django.utils.module_loading import import_string
from rest_framework.exceptions import ValidationError

from backend.utils.basic import str2bool
from backend.utils.response import PermsResponse

from .client import IAMClient
from .exceptions import AttrValidationError, PermissionDeniedError
from .perm import PermCtx
from .perm import Permission as PermPermission
from .request import ResourceRequest

logger = logging.getLogger(__name__)


def can_skip_related_perms(method_name: str) -> bool:
    if method_name == 'can_view':
        return True
    return False


class RelatedPermission(metaclass=ABCMeta):
    """
    用于资源 Permission 类的方法装饰, 目的是支持 related_actions 的权限校验

    note: 如果被装饰的方法/函数名符合 can_skip_related_perms 中的规则(如 can_view)，并且校验有权限，
    则跳过 related_actions 的权限校验，目的是加速资源查看类型鉴权

    related_project_perm 和 related_cluster_perm 装饰器的用法:

    class ClusterPermission(Permission):

        resource_type: str = 'cluster'
        resource_request_cls: Type[ResourceRequest] = ClusterRequest

        @related_project_perm(method_name='can_view')
        def can_view(self, perm_ctx: ClusterPermCtx, raise_exception: bool = True, just_raise: bool = False) -> bool:
            return self.can_action(perm_ctx, ClusterAction.VIEW, raise_exception, just_raise)

        @related_cluster_perm(method_name='can_view')
        def can_manage(self, perm_ctx: ClusterPermCtx, raise_exception: bool = True, just_raise: bool = False) -> bool:
            return self.can_action(perm_ctx, ClusterAction.MANAGE, raise_exception, just_raise)

    """

    module_name: str  # 资源模块名 如 cluster, project

    def __init__(self, method_name: str):
        """
        :param method_name: 权限类的 can_{action} 方法名，用于校验用户是否具有对应的操作权限
        """
        self.method_name = method_name

    def _gen_perm_obj(self) -> PermPermission:
        """获取权限类实例，如 project.ProjectPermission"""
        p_module_name = __name__.rsplit('.', 1)[0]
        return import_string(
            f'{p_module_name}.resources.{self.module_name}.{self.module_name.capitalize()}Permission'
        )()

    @wrapt.decorator
    def __call__(self, wrapped, instance, args, kwargs):
        self.perm_obj = self._gen_perm_obj()

        perm_ctx = self._convert_perm_ctx(instance, args, kwargs)

        try:
            is_allowed = wrapped(*args, **kwargs)
        except PermissionDeniedError as e:
            # 按照权限中心的建议，无论关联资源操作是否有权限，统一按照无权限返回，目的是生成最终的 apply_url
            perm_ctx.force_raise = True
            try:
                getattr(self.perm_obj, self.method_name)(perm_ctx)
            except PermissionDeniedError as err:
                raise PermissionDeniedError(
                    f'{e.message}; {err.message}',
                    username=perm_ctx.username,
                    action_request_list=e.action_request_list + err.action_request_list,
                )
        else:
            # 无权限，并且没有抛出 PermissionDeniedError, 说明 raise_exception = False
            if not is_allowed:
                return is_allowed

            # 如果是查看类操作有权限，不再继续校验 related_actions 权限
            if can_skip_related_perms(wrapped.__name__):
                return is_allowed

            # 继续校验 related_actions 权限
            logger.debug(f'continue to verify {self.method_name} {self.module_name} permission...')
            raise_exception = kwargs.get('raise_exception', True)
            return getattr(self.perm_obj, self.method_name)(perm_ctx, raise_exception=raise_exception)

    @abstractmethod
    def _convert_perm_ctx(self, instance, args, kwargs) -> PermCtx:
        """将被装饰的方法中的 perm_ctx 转换成 perm_obj.method_name 需要的 perm_ctx"""

    @property
    def action_id(self) -> str:
        return f'{self.perm_obj.resource_type}_{self.method_name[4:]}'


class Permission:
    """鉴权装饰器基类，用于装饰函数或者方法"""

    module_name: str  # 资源模块名 如 cluster, project

    def __init__(self, method_name: str):
        """
        :param method_name: 权限类的 can_{action} 方法名，用于校验用户是否具有对应的操作权限
        """
        self.method_name = method_name

    def _gen_perm_obj(self) -> PermPermission:
        """获取权限类实例，如 project.ProjectPermission"""
        p_module_name = __name__.rsplit('.', 1)[0]
        return import_string(
            f'{p_module_name}.resources.{self.module_name}.{self.module_name.capitalize()}Permission'
        )()

    @wrapt.decorator
    def __call__(self, wrapped, instance, args, kwargs):

        self.perm_obj = self._gen_perm_obj()

        if len(args) <= 0:
            raise TypeError('missing PermCtx instance argument')
        if not isinstance(args[0], PermCtx):
            raise TypeError('missing ProjectPermCtx instance argument')

        getattr(self.perm_obj, self.method_name)(args[0])

        return wrapped(*args, **kwargs)


class response_perms:
    """
    view 装饰器, 向 web_annotations 中注入 perms 数据

    note: 只支持处理同一类资源
    """

    def __init__(
        self,
        action_ids: List[str],
        res_request_cls: Type[ResourceRequest],
        resource_id_key: str = 'id',
        force_add: bool = False,
    ):
        """
        :param action_ids: 权限 action_id 列表
        :param res_request_cls: 对应资源的 ResourceRequest 类, 如 ClusterRequest
        :param resource_id_key: 示例, 如果 resource_data = [{'cluster_id': 'BCS-K8S-40000'}],
                                那么 resource_id_key 设置为 cluster_id，其中 BCS-K8S-40000 是注册到权限中心的集群 ID
        :param force_add: 是否强制添加权限数据到 web_annotations 中。如果为 True, 则忽略请求中的 with_perms 参数, 主动添加权限数据
        """
        self.action_ids = action_ids
        self.res_request_cls = res_request_cls
        self.resource_id_key = resource_id_key
        self.force_add = force_add

    @wrapt.decorator
    def __call__(self, wrapped, instance, args, kwargs):
        resp = wrapped(*args, **kwargs)
        if not isinstance(resp, PermsResponse):
            raise TypeError('response_perms decorator only support PermsResponse')

        if not resp.resource_data:
            return resp

        request = args[0]
        with_perms = True
        if not self.force_add:  # 根据前端请求，决定是否返回权限数据
            with_perms = str2bool(request.query_params.get('with_perms', True))

        if not with_perms:
            return resp

        perms = self._calc_perms(request, resp)

        annots = getattr(resp, 'web_annotations', None) or {}
        resp.web_annotations = {"perms": perms, **annots}

        return resp

    def _calc_perms(self, request, resp: PermsResponse) -> Dict[str, Dict[str, bool]]:
        if isinstance(resp.resource_data, list):
            res = [item.get(self.resource_id_key) for item in resp.resource_data]
        else:
            res = resp.resource_data.get(self.resource_id_key)

        try:
            iam_path_attrs = {'project_id': request.project.project_id}
        except Exception as e:
            logger.error('create iam_path_attrs failed: %s', e)
            iam_path_attrs = {}

        iam_path_attrs.update(resp.iam_path_attrs)
        try:
            client = IAMClient()
            return client.batch_resource_multi_actions_allowed(
                request.user.username,
                self.action_ids,
                self.res_request_cls(res, **iam_path_attrs),
            )
        except AttrValidationError as e:
            raise ValidationError(e)
