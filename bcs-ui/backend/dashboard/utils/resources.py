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
from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.custom_object import CustomResourceDefinition


def get_crd_scope(crd_name: str, project_id: str, cluster_id: str, access_token: str) -> str:
    """
    获取 CRD 资源维度
    NOTE 不可直接使用 ctx_cluster，每次请求对象不同，缓存不生效

    :param crd_name: CRD 名称
    :param project_id: 项目 ID
    :param cluster_id: 集群 ID
    :param access_token: 用户 token
    :return: Namespaced / Cluster
    """
    ctx_cluster = CtxCluster.create(id=cluster_id, project_id=project_id, token=access_token)
    return CustomResourceDefinition(ctx_cluster).get(crd_name)['scope']
