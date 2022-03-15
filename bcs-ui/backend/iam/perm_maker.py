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

from django.utils.module_loading import import_string

from .permissions.exceptions import AttrValidationError
from .permissions.perm import PermCtx, Permission

logger = logging.getLogger(__name__)

PermClsNamePrefix = {
    'project': 'Project',
    'cluster': 'Cluster',
    'cluster_scoped': 'ClusterScoped',
    'namespace': 'Namespace',
    'namespace_scoped': 'NamespaceScoped',
    'templateset': 'Templateset',
}


def make_perm_ctx(action_id: str, username: str = 'anonymous', **ctx_kwargs) -> PermCtx:
    """根据 action_id，生成对应的 PermCtx"""
    p_module_name = __name__.rsplit('.', 1)[0]
    perm_ctx_cls = import_string(f'{p_module_name}.permissions.resources.{_get_cls_name_suffix(action_id)}PermCtx')

    try:
        perm_ctx = perm_ctx_cls(username=username, **ctx_kwargs)
    except TypeError as e:
        logger.error(e)
        raise AttrValidationError("perm ctx got an unexpected init argument")

    return perm_ctx


def make_res_permission(action_id: str) -> Permission:
    """根据 action_id，生成对应的 Permission"""
    p_module_name = __name__.rsplit('.', 1)[0]
    perm_cls = import_string(f'{p_module_name}.permissions.resources.{_get_cls_name_suffix(action_id)}Permission')
    return perm_cls()


def _get_cls_name_suffix(action_id: str) -> str:
    return PermClsNamePrefix[action_id.rsplit('_', 1)[0]]
