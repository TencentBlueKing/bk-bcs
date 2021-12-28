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
from abc import ABC, abstractmethod
from typing import Dict, List, Optional, Type

import attr
from django.conf import settings

from .client import IAMClient
from .exceptions import AttrValidationError, PermissionDeniedError
from .request import ActionResourcesRequest, IAMResource, ResourceRequest

logger = logging.getLogger(__name__)


@attr.dataclass
class ResCreatorAction:
    """用于新建关联属性授权接口"""

    creator: str
    project_id: str
    resource_type: str

    def __attrs_post_init__(self):
        self.system = settings.BK_IAM_SYSTEM_ID

    def to_data(self) -> Dict:
        return {'creator': self.creator, 'system': self.system, 'type': self.resource_type}


@attr.dataclass
class PermCtx:
    """
    权限参数上下文
    note: 由于 force_raise 默认值的原因，其子类属性必须设置默认值
    """

    username: str
    force_raise: bool = False  # 如果为 True, 表示不做权限校验，直接以无权限方式抛出异常

    def validate_resource_id(self):
        """校验资源实例 ID. 如果校验不过，抛出 AttrValidationError 异常"""
        if not self.resource_id:
            raise AttrValidationError('missing valid resource_id')

    @property
    def resource_id(self) -> str:
        return ''

    def validate(self):
        if not self.username:
            raise AttrValidationError('username must not be empty')


class Permission(ABC, IAMClient):
    """
    对接 IAM 的权限基类
    """

    resource_type: str = ''
    resource_request_cls: Type[ResourceRequest] = ResourceRequest
    parent_res_perm: Optional['Permission'] = None  # 父级资源的权限类对象

    def can_action(self, perm_ctx: PermCtx, action_id: str, raise_exception: bool, use_cache: bool = False) -> bool:
        """
        校验用户的 action_id 权限

        :param perm_ctx: 权限校验的上下文
        :param action_id: 资源操作 ID
        :param raise_exception: 无权限时，是否抛出异常
        :param use_cache: 是否使用本地缓存 (缓存时间 1 min) 校验权限。用于非敏感操作鉴权，比如 view 操作
        """
        if perm_ctx.force_raise:
            self._raise_permission_denied_error(perm_ctx, action_id)

        is_allowed = self._can_action(perm_ctx, action_id, use_cache)

        if raise_exception and not is_allowed:
            self._raise_permission_denied_error(perm_ctx, action_id)

        return is_allowed

    def can_multi_actions(self, perm_ctx: PermCtx, action_ids: List[str], raise_exception: bool) -> bool:
        """
        校验同类型单个资源的多个 action 权限

        :param perm_ctx: 权限校验的上下文
        :param action_ids: 资源 action_id 列表
        :param raise_exception: 无权限时，是否抛出异常
        :return: 只有 action_id 都有权限时，才返回 True; 否则返回 False 或者抛出异常
        """
        if not perm_ctx.resource_id:
            raise ValueError("perm_ctx.resource_id must not be empty")

        # perms 结构如 {'project_view': True, 'project_edit': False}
        if perm_ctx.force_raise:
            perms = {action_id: False for action_id in action_ids}
        else:
            perms = self.resource_inst_multi_actions_allowed(
                perm_ctx.username, action_ids, res_request=self.make_res_request(perm_ctx.resource_id, perm_ctx)
            )

        return self._can_multi_actions(perm_ctx, perms, raise_exception)

    def grant_resource_creator_actions(self, username: str, creator_action: ResCreatorAction):
        """
        用于创建资源时，注册用户对该资源的关联操作权限.
        note: 具体的关联操作见权限模型的 resource_creator_actions 字段
        """
        return self.iam._client.grant_resource_creator_actions(None, username, creator_action.to_data())

    def make_res_request(self, res_id: str, perm_ctx: PermCtx) -> ResourceRequest:
        """创建当前资源 request"""
        return self.resource_request_cls(res_id)

    def has_parent_resource(self) -> bool:
        return self.parent_res_perm is not None

    @abstractmethod
    def get_parent_chain(self, perm_ctx: PermCtx) -> List[IAMResource]:
        """从 ctx 中获取 parent_chain"""

    def _can_action(self, perm_ctx: PermCtx, action_id: str, use_cache: bool = False) -> bool:
        res_id = self.get_resource_id(perm_ctx)

        if res_id:  # 与当前资源实例相关
            res_request = self.make_res_request(res_id, perm_ctx)
            return self.resource_inst_allowed(perm_ctx.username, action_id, res_request, use_cache)

        # 与当前资源实例无关, 并且无关联上级资源, 按资源实例无关处理
        if not self.has_parent_resource():
            return self.resource_type_allowed(perm_ctx.username, action_id, use_cache)

        # 有关联上级资源
        res_request = self.parent_res_perm.make_res_request(
            res_id=self.parent_res_perm.get_resource_id(perm_ctx), perm_ctx=perm_ctx
        )
        return self.resource_inst_allowed(perm_ctx.username, action_id, res_request, use_cache)

    def _can_multi_actions(self, perm_ctx: PermCtx, perms: Dict[str, bool], raise_exception: bool) -> bool:
        messages = []
        action_request_list = []

        for action_id, is_allowed in perms.items():
            if is_allowed:
                continue

            try:
                self._raise_permission_denied_error(perm_ctx, action_id)
            except PermissionDeniedError as e:
                messages.append(e.message)
                action_request_list.extend(e.action_request_list)

        if not messages:
            return True

        if not raise_exception:
            return False

        raise PermissionDeniedError(
            message=';'.join(messages), username=perm_ctx.username, action_request_list=action_request_list
        )

    def _raise_permission_denied_error(self, perm_ctx: PermCtx, action_id: str):
        """抛出 PermissionDeniedError 异常, 其中 username 和 action_request_list 可用于生成权限申请跳转链接"""
        res_id = self.get_resource_id(perm_ctx)
        resources = None
        resource_type = self.resource_type
        parent_chain = None

        if res_id:
            resources = [res_id]
            parent_chain = self.get_parent_chain(perm_ctx)
        elif self.has_parent_resource():
            resource_type = self.parent_res_perm.resource_type
            resources = [self.parent_res_perm.get_resource_id(perm_ctx)]
            parent_chain = self.parent_res_perm.get_parent_chain(perm_ctx)

        raise PermissionDeniedError(
            f"no {action_id} permission",
            username=perm_ctx.username,
            action_request_list=[ActionResourcesRequest(action_id, resource_type, resources, parent_chain)],
        )

    @abstractmethod
    def get_resource_id(self, perm_ctx: PermCtx) -> Optional[str]:
        """从 ctx 中获取当前资源对应的 id"""
