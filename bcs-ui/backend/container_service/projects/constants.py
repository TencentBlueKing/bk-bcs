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
from enum import Enum

from django.utils.translation import ugettext_lazy as _

from backend.utils.basic import ChoicesEnum

# 管理员标示
SUPER_ROLE = 'manager'


class Policy(Enum):
    # 项目管理
    PROJECT = 'modify:project:btn'
    # 集群管理
    CLUSTER = 'cluster:menu'
    # 节点管理
    NODE = 'node:menu'
    # 应用管理
    APP = 'app:menu'
    # 配置管理
    CONFIGURATION = 'configuration:menu'
    # 网络管理
    NETWORK = 'network:menu'
    # 资源管理
    RESOURCE = 'resource:menu'
    # 仓库管理
    REPO = 'repo:menu'
    # 仓库按钮
    REPO_MODIFY = 'modify:repo:btn'


class PolicyEffect(Enum):
    # 正常
    NORMAL = 0
    # 隐藏
    HIDDEN = 1
    # 按钮置灰
    DISABLED = 2


PolicyLabelOrdering = [_("容器服务"), _("仓库管理"), _("项目管理")]

PolicyOrdering = {
    'jfrog': [_("prod环境拉取"), _("prod环境推送"), _("test环境拉取"), _("test环境推送"), _("dev环境拉取"), _("dev环境推送")],
    'paas_backend': [_("集群管理"), _("节点管理"), _("应用管理"), _("网络管理"), _("仓库管理"), _("资源管理")],
    'apigw': [],
}


class StaffInfoStatus(ChoicesEnum):
    # 审批中，默认
    NORMAL = 0
    INCUMBENCY = 1
    RESIGN = 2
    TRIAL = 3
    WAITING_ENTRY = 8
    NOT_ENTRY = 9

    _choices_labels = (
        (NORMAL, _("正常")),
        (INCUMBENCY, _("在职")),  # 现在都返回显示正常
        (RESIGN, _("已离职")),
        (TRIAL, _("试用")),
        (WAITING_ENTRY, _("待入职")),
        (NOT_ENTRY, _("待入职")),
    )
