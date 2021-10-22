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

logger = logging.getLogger(__name__)


def enabled_hpa_feature(cluster_id_list: list) -> bool:
    """HPA按集群做白名单控制"""
    return True


def enabled_sync_namespace(project_id: str) -> bool:
    """是否允许非导航【命名空间】页面创建的命名空间数据"""
    return True


def enabled_force_sync_chart_repo(project_id: str) -> bool:
    """是否允许强制同步仓库数据"""
    return False


def enable_helm_v3(cluster_id: str) -> bool:
    """是否允许集群使用helm3功能"""
    return True


def enable_incremental_sync_chart_repo(project_id: str) -> bool:
    """是否开启增量同步仓库数据"""
    return False


try:
    from .whitelist_ext import *  # noqa
except ImportError as e:
    logger.debug('Load extension failed: %s', e)
