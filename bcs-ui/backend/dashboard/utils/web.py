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
from typing import Dict

from django.utils.translation import ugettext_lazy as _

from backend.dashboard.permissions import validate_cluster_perm


def gen_base_web_annotations(request, project_id: str, cluster_id: str) -> Dict:
    """ 生成资源视图相关的页面控制信息，用于控制按钮展示等 """
    has_cluster_perm = validate_cluster_perm(request, project_id, cluster_id, raise_exception=False)
    # 目前 创建 / 删除 / 更新 按钮权限 & 提示信息相同
    tip = _('当前用户没有操作集群 {} 的权限').format(cluster_id) if not has_cluster_perm else ''
    btn_perm = {'clickable': has_cluster_perm, 'tip': tip}
    return {
        'perms': {
            'page': {
                'create_btn': btn_perm,
                'update_btn': btn_perm,
                'delete_btn': btn_perm,
                'reschedule_pod_btn': btn_perm,
                'web_console_btn': btn_perm,
            }
        }
    }
