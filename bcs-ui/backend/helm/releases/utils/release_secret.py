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
from typing import Dict, List

from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.configs import secret
from backend.utils.basic import getitems

from .formatter import ReleaseSecretFormatter

logger = logging.getLogger(__name__)


def list_namespaced_releases(ctx_cluster: CtxCluster, namespace: str) -> List[Dict]:
    """查询namespace下的release
    NOTE: 为防止后续helm release对应的secret名称规则(sh.helm.release.v1.名称.v版本)变动，不直接根据secret名称进行过滤
    """
    client = secret.Secret(ctx_cluster)
    # 查询指定命名空间下的secrets
    return client.list(formatter=ReleaseSecretFormatter(), label_selector="owner=helm", namespace=namespace)


def get_release_detail(ctx_cluster: CtxCluster, namespace: str, release_name: str) -> Dict:
    """获取release详情"""
    release_list = list_namespaced_releases(ctx_cluster, namespace)
    release_list = [release for release in release_list if release.get("name") == release_name]
    if not release_list:
        logger.error(
            "not found release: [cluster_id: %s, namespace: %s, name: %s]", ctx_cluster.id, namespace, release_name
        )
        return {}
    # 通过release中的version对比，过滤到最新的 release data
    # NOTE: helm存储到secret中的release数据，每变动一次，增加一个secret，对应的revision就会增加一个，也就是最大的revision为当前release的存储数据
    return max(release_list, key=lambda item: item["version"])


def get_release_notes(ctx_cluster: CtxCluster, namespace: str, release_name: str) -> str:
    """查询release的notes"""
    release_detail = get_release_detail(ctx_cluster, namespace, release_name)
    return getitems(release_detail, "info.notes", "")
